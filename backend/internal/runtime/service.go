package runtime

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RuntimeService manages running containers.
// It is the only service that knows how to start, stop, and inspect containers.
type RuntimeService interface {
	// RunContainer starts a container from an image, injecting envArgs (e.g. ["-e", "KEY=VAL"]).
	// stdout/stderr from docker run are written to logWriter.
	RunContainer(imageName string, envArgs []string, logWriter io.Writer) (containerID string, hostPort int, err error)
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	InspectContainer(containerID string) (string, error)
}

type runtimeService struct{}

func NewRuntimeService() RuntimeService {
	return &runtimeService{}
}

// RunContainer starts a container with -d -P plus any env var flags.
func (s *runtimeService) RunContainer(imageName string, envArgs []string, logWriter io.Writer) (string, int, error) {
	// Build args: docker run -d -P [-e KEY=VAL ...] imageName
	args := []string{"run", "-d", "-P"}
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

	// Output: "8080/tcp -> 0.0.0.0:32768"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("no port mappings found")
	}

	parts := strings.Split(lines[0], ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("unexpected port mapping format: %s", lines[0])
	}

	port, err := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1]))
	if err != nil {
		return 0, fmt.Errorf("could not parse port: %w", err)
	}

	return port, nil
}

func (s *runtimeService) StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *runtimeService) DeleteContainer(containerID string) error {
	cmd := exec.Command("docker", "rm", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
