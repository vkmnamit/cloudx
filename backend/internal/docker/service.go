package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vkmnamit/cloudx/internal/analyzer"
)

// DockerService is responsible for generating Dockerfiles and building images.
// It does NOT run containers — that's RuntimeService's job.
type DockerService interface {
	GenerateDockerfile(analysis *analyzer.Analysis, destPath string) error
	// BuildImage builds a Docker image. Build stdout/stderr is written to logWriter
	// so every line is captured in deployment_logs.
	BuildImage(projectPath string, imageName string, logWriter io.Writer) error
}

type dockerService struct{}

func NewDockerService() DockerService {
	return &dockerService{}
}

// GenerateDockerfile writes a language-specific Dockerfile to destPath.
// If a Dockerfile already exists, it is a no-op.
func (s *dockerService) GenerateDockerfile(analysis *analyzer.Analysis, destPath string) error {
	if analysis.DockerfileExists {
		return nil
	}

	var content string
	switch analysis.Language {
	case "Node.js":
		content = fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY . .
RUN %s
EXPOSE %d
CMD ["%s"]
`, analysis.BuildCommand, analysis.ExposedPort, analysis.StartCommand)

	case "Go":
		content = fmt.Sprintf(`FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN %s

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE %d
CMD ["%s"]
`, analysis.BuildCommand, analysis.ExposedPort, analysis.StartCommand)

	case "Python":
		content = fmt.Sprintf(`FROM python:3.11-slim
WORKDIR /app
COPY . .
RUN %s
EXPOSE %d
CMD ["%s"]
`, analysis.BuildCommand, analysis.ExposedPort, analysis.StartCommand)

	case "Rust":
		content = fmt.Sprintf(`FROM rust:1.75-slim AS builder
WORKDIR /app
COPY . .
RUN %s

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/target/release/app .
EXPOSE %d
CMD ["%s"]
`, analysis.BuildCommand, analysis.ExposedPort, analysis.StartCommand)

	default:
		return fmt.Errorf("unsupported language for dockerfile generation: %s", analysis.Language)
	}

	return os.WriteFile(filepath.Join(destPath, "Dockerfile"), []byte(content), 0644)
}

// BuildImage runs `docker build` using os/exec.
// logWriter (from LogService.NewWriter) receives every line of build output.
// Note: In a production system, this would use the Docker Engine API instead of shelling out.
func (s *dockerService) BuildImage(projectPath string, imageName string, logWriter io.Writer) error {
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = projectPath

	if logWriter != nil {
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
