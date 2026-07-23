package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vkmnamit/cloudx/internal/analyzer"
	"github.com/vkmnamit/cloudx/internal/templates"
)

// DockerService generates Dockerfiles and builds images for deployments.
type DockerService interface {
	GenerateDockerfile(analysis *analyzer.Analysis, dockerfilePath string, port int) error
	BuildImage(buildContext string, dockerfilePath string, imageTag string, logWriter io.Writer) error
}

type dockerService struct{}

func NewDockerService() DockerService {
	return &dockerService{}
}

func (s *dockerService) GenerateDockerfile(analysis *analyzer.Analysis, dockerfilePath string, port int) error {
	// If a Dockerfile already exists, respect it.
	if analysis != nil && analysis.DockerfileExists {
		if _, err := os.Stat(dockerfilePath); err == nil {
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		return fmt.Errorf("failed to create dockerfile directory: %w", err)
	}

	if port > 0 && analysis != nil {
		analysis.ExposedPort = port
	} else if port <= 0 && analysis != nil && analysis.ExposedPort <= 0 {
		analysis.ExposedPort = 8080
	} else if port <= 0 && analysis == nil {
		return fmt.Errorf("port must be > 0 or analysis must provide a port")
	}

	content, err := templates.RenderDockerfile(analysis)
	if err != nil {
		return fmt.Errorf("failed to render dockerfile: %w", err)
	}

	if err := os.WriteFile(dockerfilePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write dockerfile: %w", err)
	}

	return nil
}

func (s *dockerService) BuildImage(buildContext string, dockerfilePath string, imageTag string, logWriter io.Writer) error {
	args := []string{"build", "-t", imageTag, "-f", dockerfilePath, buildContext}
	cmd := exec.Command("docker", args...)
	if logWriter != nil {
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	return nil
}
