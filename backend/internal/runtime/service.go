package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// RuntimeService manages running containers.
type RuntimeService interface {
	RunContainer(imageName string, envArgs []string, logWriter io.Writer, containerPort int) (containerID string, hostPort int, err error)
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	InspectContainer(containerID string) (string, error)
	WaitUntilHealthy(ctx context.Context, containerID string, hostPort int, timeout time.Duration) error
}

type runtimeService struct{}

func NewRuntimeService() RuntimeService {
	return &runtimeService{}
}

func (s *runtimeService) RunContainer(imageName string, envArgs []string, logWriter io.Writer, containerPort int) (string, int, error) {
	network := os.Getenv("CLOUDX_RUNTIME_NETWORK")
	if network == "" {
		network = "bridge"
	}
	args := []string{"run", "-d", "--network", network, "-p", fmt.Sprintf("0:%d", containerPort)}
	args = append(args, envArgs...)
	args = append(args, imageName)

	cmd := exec.Command("docker", args...)
	if logWriter != nil {
		cmd.Stderr = logWriter
	} else {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("docker run failed: %w", err)
	}

	containerID := strings.TrimSpace(string(out))
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	hostPort, err := s.getHostPort(containerID)
	if err != nil {
		return containerID, 0, fmt.Errorf("container started but could not get host port: %w", err)
	}

	return containerID, hostPort, nil
}

func (s *runtimeService) getHostPort(containerID string) (int, error) {
	cmd := exec.Command("docker", "port", containerID)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var mappings map[string][]struct {
		HostPort string `json:"HostPort"`
	}
	if err := json.Unmarshal(out, &mappings); err != nil {
		return 0, fmt.Errorf("parse Docker port mappings: %w", err)
	}
	for _, bindings := range mappings {
		if len(bindings) == 0 {
			continue
		}
		port, err := strconv.Atoi(bindings[0].HostPort)
		if err != nil {
			return 0, fmt.Errorf("could not parse host port: %w", err)
		}
		return port, nil
	}
	return 0, fmt.Errorf("no port mappings found")
}

func (s *runtimeService) StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	return cmd.Run()
}

func (s *runtimeService) DeleteContainer(containerID string) error {
	cmd := exec.Command("docker", "rm", "-f", containerID)
	return cmd.Run()
}

func (s *runtimeService) InspectContainer(containerID string) (string, error) {
	cmd := exec.Command("docker", "inspect", containerID)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker inspect failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// WaitUntilHealthy waits for the container to become healthy.
// It checks TCP connectivity first. If it connects, it checks HTTP if applicable,
// but accepts TCP connection as healthy for non-HTTP services or if the endpoint returns anything (even 404).
func (s *runtimeService) WaitUntilHealthy(ctx context.Context, containerID string, hostPort int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	targets := []string{}
	targets = append(targets, fmt.Sprintf("127.0.0.1:%d", hostPort))
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for container to become healthy")
		case <-ticker.C:
			// Check if container is still running
			out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerID).Output()
			if err != nil || strings.TrimSpace(string(out)) != "true" {
				return fmt.Errorf("container stopped unexpectedly")
			}

			// Try TCP dial
			for _, target := range targets {
				conn, err := net.DialTimeout("tcp", target, 500*time.Millisecond)
				if err == nil {
					conn.Close()

					// Optional: Try HTTP GET / to see if it responds, but we don't strictly require 200 OK.
					return nil
				}
			}
		}
	}
}
