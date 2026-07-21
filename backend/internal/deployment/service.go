package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vkmnamit/cloudx/internal/analyzer"
	"github.com/vkmnamit/cloudx/internal/docker"
	"github.com/vkmnamit/cloudx/internal/environment"
	"github.com/vkmnamit/cloudx/internal/git"
	"github.com/vkmnamit/cloudx/internal/logs"
	"github.com/vkmnamit/cloudx/internal/runtime"
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
	repo        Repository
	git         git.GitService
	analyzer    analyzer.AnalyzerService
	docker      docker.DockerService
	runtime     runtime.RuntimeService
	logSvc      logs.LogService
	envSvc      environment.EnvironmentService
}

func NewService(
	repo Repository,
	g git.GitService,
	a analyzer.AnalyzerService,
	d docker.DockerService,
	r runtime.RuntimeService,
	l logs.LogService,
	e environment.EnvironmentService,
) Service {
	return &service{
		repo:     repo,
		git:      g,
		analyzer: a,
		docker:   d,
		runtime:  r,
		logSvc:   l,
		envSvc:   e,
	}
}

func (s *service) Deploy(ctx context.Context, userID string, req DeployRequest) (*Deployment, error) {
	now := time.Now()
	dep := &Deployment{
		ID:          uuid.New().String(),
		UserID:      userID,
		ProjectName: req.ProjectName,
		Repository:  req.RepoURL,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// ── Step 1: Save Pending — history is always preserved ──────────
	if err := s.repo.Create(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, "Deployment created with status: pending")

	// ── Step 2: Save env vars if provided ───────────────────────────
	if len(req.EnvVars) > 0 {
		if err := s.envSvc.SetVars(ctx, dep.ID, req.EnvVars); err != nil {
			_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelError, "Failed to save env vars: "+err.Error())
			_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
			return dep, fmt.Errorf("failed to save env vars: %w", err)
		}
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, fmt.Sprintf("Saved %d environment variable(s)", len(req.EnvVars)))
	}

	// ── Step 3: Clone repository ─────────────────────────────────────
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, "Cloning repository: "+req.RepoURL)

	projectPath, err := s.git.Clone(req.RepoURL, dep.ID)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelError, "Clone failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("clone failed: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseClone, logs.LevelInfo, "Repository cloned to: "+projectPath)

	// ── Step 4: Analyze ─────────────────────────────────────────────
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Analyzing project...")
	analysis, err := s.analyzer.Analyze(projectPath)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Analysis failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("analysis failed: %w", err)
	}

	dep.Language = analysis.Language
	dep.Framework = analysis.Framework
	dep.ContainerPort = analysis.ExposedPort
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo,
		fmt.Sprintf("Detected: %s / %s (port %d)", analysis.Language, analysis.Framework, analysis.ExposedPort))

	// ── Step 5: Generate Dockerfile ──────────────────────────────────
	if err := s.docker.GenerateDockerfile(analysis, projectPath); err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Dockerfile generation failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("dockerfile generation failed: %w", err)
	}
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Dockerfile ready")

	// ── Step 6: Build Image ──────────────────────────────────────────
	_ = s.repo.UpdateStatus(ctx, dep.ID, StatusBuilding)
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Building Docker image...")

	imageName := fmt.Sprintf("cloudx-%s", dep.ID)
	buildWriter := s.logSvc.NewWriter(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo)

	if err := s.docker.BuildImage(projectPath, imageName, buildWriter); err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelError, "Docker build failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("docker build failed: %w", err)
	}
	dep.ImageName = imageName
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseBuild, logs.LevelInfo, "Image built: "+imageName)

	// ── Step 7: Fetch env args and run container ─────────────────────
	envArgs, err := s.envSvc.ToDockerArgs(ctx, dep.ID)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelError, "Failed to load env args: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("failed to load env args: %w", err)
	}

	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo, "Starting container...")
	runtimeWriter := s.logSvc.NewWriter(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo)

	containerID, hostPort, err := s.runtime.RunContainer(imageName, envArgs, runtimeWriter)
	if err != nil {
		_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelError, "Container run failed: "+err.Error())
		_ = s.repo.UpdateStatus(ctx, dep.ID, StatusFailed)
		return dep, fmt.Errorf("container run failed: %w", err)
	}

	dep.ContainerID = containerID
	dep.HostPort = hostPort
	dep.URL = fmt.Sprintf("http://localhost:%d", hostPort)
	dep.Status = StatusRunning
	dep.UpdatedAt = time.Now()

	_ = s.repo.UpdateStatus(ctx, dep.ID, StatusRunning)
	_ = s.logSvc.Append(ctx, dep.ID, logs.PhaseRuntime, logs.LevelInfo,
		fmt.Sprintf("Container running at %s (container: %s)", dep.URL, containerID))

	return dep, nil
}

func (s *service) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListDeployments(ctx context.Context, userID string) ([]*Deployment, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) DeleteDeployment(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
