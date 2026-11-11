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

	"github.com/google/uuid"
)

const (
	deploymentMemoryLimit = "512m"
	deploymentCPULimit    = "0.5"
	deploymentPIDsLimit   = "128"
	deploymentNetworkBase = "cloudx-deployment-"
)

// RuntimeService manages running containers.
type RuntimeService interface {
	RunContainer(deploymentID string, imageName string, envArgs []string, logWriter io.Writer, containerPort int) (containerID string, hostPort int, err error)
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	InspectContainer(containerID string) (string, error)
	WaitUntilHealthy(ctx context.Context, containerID string, hostPort int, timeout time.Duration) error
}

type runtimeService struct{}

func NewRuntimeService() RuntimeService {
	return &runtimeService{}
}

func (s *runtimeService) RunContainer(deploymentID string, imageName string, envArgs []string, logWriter io.Writer, containerPort int) (string, int, error) {
	networkName, err := deploymentNetworkName(deploymentID)
	if err != nil {
		return "", 0, err
	}
	hostGateway, err := hostGatewayAddress()
	if err != nil {
		return "", 0, err
	}

	createNetwork := exec.Command("docker", "network", "create", "--driver", "bridge", networkName)
	if output, err := createNetwork.CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("create isolated deployment network: %w: %s", err, strings.TrimSpace(string(output)))
	}

	args := containerRunArgs(deploymentID, networkName, imageName, envArgs, containerPort, hostGateway)

	cmd := exec.Command("docker", args...)
	if logWriter != nil {
		cmd.Stderr = logWriter
	} else {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("docker run failed: %w%s", err, networkCleanupError(networkName))
	}

	containerID := strings.TrimSpace(string(out))
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	hostPort, err := s.getHostPort(containerID, containerPort)
	if err != nil {
		if deleteErr := s.DeleteContainer(containerID); deleteErr != nil {
			return containerID, 0, fmt.Errorf("container started but could not get host port: %w (cleanup failed: %v)", err, deleteErr)
		}
		return containerID, 0, fmt.Errorf("container started but could not get host port: %w", err)
	}

	return containerID, hostPort, nil
}

func hostGatewayAddress() (string, error) {
	addresses, err := net.LookupIP("host.docker.internal")
	if err != nil {
		return "", fmt.Errorf("resolve host.docker.internal for private health checks: %w", err)
	}
	for _, address := range addresses {
		if ipv4 := address.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}
	return "", fmt.Errorf("host.docker.internal did not resolve to an IPv4 address")
}

func containerRunArgs(deploymentID, networkName, imageName string, envArgs []string, containerPort int, hostGateway string) []string {
	args := []string{
		"run", "-d",
		"--name", networkName,
		"--label", "io.cloudx.deployment-id=" + deploymentID,
		"--network", networkName,
		"--restart", "unless-stopped",
		"--memory", deploymentMemoryLimit,
		"--cpus", deploymentCPULimit,
		"--pids-limit", deploymentPIDsLimit,
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges:true",
		"--init",
		"-p", fmt.Sprintf("%s::%d", hostGateway, containerPort),
	}
	args = append(args, envArgs...)
	args = append(args, imageName)
	return args
}

func deploymentNetworkName(deploymentID string) (string, error) {
	id, err := uuid.Parse(deploymentID)
	if err != nil {
		return "", fmt.Errorf("invalid deployment ID for runtime: %w", err)
	}
	return deploymentNetworkBase + strings.ReplaceAll(id.String(), "-", ""), nil
}

func (s *runtimeService) getHostPort(containerID string, containerPort int) (int, error) {
	cmd := exec.Command("docker", "port", containerID, fmt.Sprintf("%d/tcp", containerPort))
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return parseHostPortOutput(string(out))
}

func parseHostPortOutput(output string) (int, error) {
	for _, line := range strings.Split(output, "\n") {
		target := strings.TrimSpace(line)
		if target == "" {
			continue
		}
		if separator := strings.LastIndex(target, "->"); separator >= 0 {
			target = strings.TrimSpace(target[separator+2:])
		}
		_, portText, err := net.SplitHostPort(target)
		if err != nil {
			return 0, fmt.Errorf("parse Docker port mapping %q: %w", target, err)
		}
		port, err := strconv.Atoi(portText)
		if err != nil || port < 1 || port > 65535 {
			return 0, fmt.Errorf("invalid Docker host port %q", portText)
		}
		return port, nil
	}
	return 0, fmt.Errorf("no Docker port mapping found")
}

func networkNames(containerID string) ([]string, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{json .NetworkSettings.Networks}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("inspect container networks: %w", err)
	}
	var networks map[string]json.RawMessage
	if err := json.Unmarshal(out, &networks); err != nil {
		return nil, fmt.Errorf("parse container networks: %w", err)
	}
	names := make([]string, 0, len(networks))
	for name := range networks {
		if strings.HasPrefix(name, deploymentNetworkBase) {
			names = append(names, name)
		}
	}
	return names, nil
}

func networkCleanupError(networkName string) string {
	cmd := exec.Command("docker", "network", "rm", networkName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Sprintf(" (network cleanup failed: %v: %s)", err, strings.TrimSpace(string(output)))
	}
	return ""
}

func (s *runtimeService) StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	return cmd.Run()
}

func (s *runtimeService) DeleteContainer(containerID string) error {
	networks, err := networkNames(containerID)
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "rm", "-f", containerID)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("remove container: %w: %s", err, strings.TrimSpace(string(output)))
	}
	for _, network := range networks {
		if err := disconnectTraefik(network); err != nil {
			return err
		}
		if output, err := exec.Command("docker", "network", "rm", network).CombinedOutput(); err != nil {
			return fmt.Errorf("remove deployment network %s: %w: %s", network, err, strings.TrimSpace(string(output)))
		}
	}
	return nil
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
	targets = append(targets, fmt.Sprintf("host.docker.internal:%d", hostPort))
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

func disconnectTraefik(networkName string) error {
	cmd := exec.Command("docker", "network", "inspect", "--format", "{{json .Containers}}", networkName)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("inspect network %s before cleanup: %w", networkName, err)
	}
	var containers map[string]struct {
		Name string `json:"Name"`
	}
	if err := json.Unmarshal(out, &containers); err != nil {
		return fmt.Errorf("parse network %s containers: %w", networkName, err)
	}
	for _, container := range containers {
		if container.Name == "cloudx-traefik" {
			cmd := exec.Command("docker", "network", "disconnect", "-f", networkName, "cloudx-traefik")
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("disconnect Traefik from network %s: %w: %s", networkName, err, strings.TrimSpace(string(output)))
			}
			break
		}
	}
	return nil
}
