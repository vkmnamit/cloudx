package deployment

import (
	"context"
	"fmt"
	"os"

	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/vkmnamit/cloudx/internal/analyzer"
	"github.com/vkmnamit/cloudx/internal/docker"
	"github.com/vkmnamit/cloudx/internal/environment"
	"github.com/vkmnamit/cloudx/internal/git"
	"github.com/vkmnamit/cloudx/internal/logs"
	"github.com/vkmnamit/cloudx/internal/runtime"
	projectservice "github.com/vkmnamit/cloudx/internal/service"
)

// Service is the deployment orchestrator. It coordinates all other services.
// It never knows Docker commands or HTTP — it only coordinates.
type Service interface {
	Deploy(ctx context.Context, userID string, req DeployRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, id string) (*Deployment, error)
	ListDeployments(ctx context.Context, userID string) ([]*Deployment, error)
	DeleteDeployment(ctx context.Context, id string) error
}

type service struct {
	repo     Repository
	service  projectservice.Service
	git      git.GitService
	analyzer analyzer.AnalyzerService
	docker   docker.DockerService
	runtime  runtime.RuntimeService
	logSvc   logs.LogService
	envSvc   environment.EnvironmentService
}

func NewService(
	repo Repository,
	serviceSvc projectservice.Service,
	g git.GitService,
	a analyzer.AnalyzerService,
	d docker.DockerService,
	r runtime.RuntimeService,
	l logs.LogService,
	e environment.EnvironmentService,
) Service {
	return &service{
		repo:     repo,
		service:  serviceSvc,
		git:      g,
		analyzer: a,
		docker:   d,
		runtime:  r,
		logSvc:   l,
		envSvc:   e,
	}
}

func (s *service) Deploy(ctx context.Context, userID string, req DeployRequest) (*Deployment, error) {
	if req.ServiceID == "" {
		return nil, fmt.Errorf("service_id is required")
	}

	serviceRecord, err := s.service.GetByID(ctx, req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load service: %w", err)
	}
	if serviceRecord.RepositoryURL == "" {
		return nil, fmt.Errorf("service %s does not have a repository_url configured", serviceRecord.ID)
	}

	now := time.Now()
	imageTag := fmt.Sprintf("cloudx-%s", uuid.New().String())
	dep := &Deployment{
		ID:         uuid.New().String(),
		ServiceID:  serviceRecord.ID,
		DeployedBy: userID,
		ImageTag:   imageTag,
		Status:     StatusQueued,
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// ── Step 1: Save Pending — history is always preserved ──────────
	if err := s.repo.Create(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, "Deployment created with status: queued")

	// ── Step 2: Save env vars if provided ───────────────────────────
	if len(req.EnvVars) > 0 {
		if err := s.envSvc.SetVars(ctx, serviceRecord.ID, req.EnvVars); err != nil {
			_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelError, "Failed to save env vars: "+err.Error())
			_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
			return dep, fmt.Errorf("failed to save env vars: %w", err)
		}
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, fmt.Sprintf("Saved %d environment variable(s)", len(req.EnvVars)))
	}

	// ── Step 3: Clone repository ─────────────────────────────────────
	branch := serviceRecord.Branch
	if branch == "" {
		branch = "main"
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, fmt.Sprintf("Cloning repository: %s (%s)", serviceRecord.RepositoryURL, branch))

	projectPath, err := s.git.Clone(serviceRecord.RepositoryURL, branch, dep.ID)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelError, "Clone failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("clone failed: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, "Repository cloned to: "+projectPath)

	if commitSHA, commitMessage, metaErr := s.git.CommitMetadata(projectPath); metaErr == nil {
		dep.CommitSHA = commitSHA
		dep.CommitMessage = commitMessage
		if err := s.repo.UpdateMetadata(ctx, dep.ID, commitSHA, commitMessage); err != nil {
			_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelWarning, "Failed to persist commit metadata: "+err.Error())
		}
	} else {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelWarning, "Failed to read commit metadata: "+metaErr.Error())
	}

	buildContext := filepath.Join(projectPath, serviceRecord.BuildContext)
	dockerfilePath := filepath.Join(projectPath, serviceRecord.DockerfilePath)
	port := serviceRecord.Port
	if port == 0 {
		port = 8080
	}

	// ── Step 4: Analyze ─────────────────────────────────────────────
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Analyzing project...")
	analysis, err := s.analyzer.Analyze(buildContext)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Analysis failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("analysis failed: %w", err)
	}

	if analysis.ProjectRoot != "" && analysis.ProjectRoot != buildContext {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Found project in sub-directory: "+filepath.Base(analysis.ProjectRoot))
		buildContext = analysis.ProjectRoot
		
		// If the original dockerfile path was a relative default, update it to the new root
		if serviceRecord.DockerfilePath == "" || serviceRecord.DockerfilePath == "./Dockerfile" || serviceRecord.DockerfilePath == "Dockerfile" {
			dockerfilePath = filepath.Join(buildContext, "Dockerfile")
		}
	}

	dep.Status = StatusBuilding
	_ = s.repo.UpdateStatus(ctx, dep.ID, StatusBuilding)
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo,
		fmt.Sprintf("Detected: %s / %s (port %d)", analysis.Language, analysis.Framework, analysis.ExposedPort))

	// ── Step 5: Generate Dockerfile ──────────────────────────────────
	// Ensure the runtime port matches the detected application port
	if analysis.ExposedPort > 0 {
		port = analysis.ExposedPort
	}
	
	if err := s.docker.GenerateDockerfile(analysis, dockerfilePath, port); err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Dockerfile generation failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("dockerfile generation failed: %w", err)
	}

	fmt.Println("========== GENERATED DOCKERFILE ==========")

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		fmt.Println("Failed to read Dockerfile:", err)
	} else {
		fmt.Println(string(content))
	}

	fmt.Println("==========================================")

	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Dockerfile ready")

	// ── Step 6: Build Image ──────────────────────────────────────────
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Building Docker image...")

	buildWriter := s.logSvc.NewWriter(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo)

	if err := s.docker.BuildImage(buildContext, dockerfilePath, imageTag, buildWriter); err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Docker build failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("docker build failed: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Image built: "+imageTag)

	// ── Step 7: Fetch env args and run container ─────────────────────
	envArgs, err := s.envSvc.ToDockerArgs(ctx, serviceRecord.ID)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelError, "Failed to load env args: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("failed to load env args: %w", err)
	}

	dep.Status = StatusDeploying
	_ = s.repo.UpdateStatus(ctx, dep.ID, StatusDeploying)
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo, "Starting container...")
	runtimeWriter := s.logSvc.NewWriter(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo)

	fmt.Println("========== DEPLOY DEBUG ==========")
	fmt.Println("Dockerfile:", dockerfilePath)
	fmt.Println("Build Context:", buildContext)
	fmt.Println("Image Tag:", imageTag)
	fmt.Printf("Env Args: %#v\n", envArgs)
	fmt.Println("==================================")

	containerID, hostPort, err := s.runtime.RunContainer(imageTag, envArgs, runtimeWriter)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelError, "Container run failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("container run failed: %w", err)
	}

	// ── Step 8: Wait for Health ──────────────────────────────────────
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo, "Waiting for container to become healthy...")
	
	if err := s.runtime.WaitUntilHealthy(ctx, containerID, hostPort, 30*time.Second); err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelError, "Health check failed: "+err.Error())
		_ = s.runtime.StopContainer(containerID)
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("health check failed: %w", err)
	}
	
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo, "Health check passed")

	dep.ImageTag = imageTag
	dep.FinishedAt = ptrTime(time.Now())
	dep.Status = StatusRunning
	dep.UpdatedAt = time.Now()

	_ = s.repo.UpdateStatus(ctx, dep.ID, StatusRunning)
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo,
		fmt.Sprintf("Container running on host port %d (container: %s)", hostPort, containerID))

	return dep, nil
}

func (s *service) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListDeployments(ctx context.Context, userID string) ([]*Deployment, error) {
	return s.repo.ListByDeployedBy(ctx, userID)
}

func (s *service) DeleteDeployment(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
