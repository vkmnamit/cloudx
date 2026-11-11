package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var hostnameLabelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

type RouteManager interface {
	URLForService(serviceID, serviceName string) (string, error)
	Publish(serviceID, serviceName, deploymentID string, containerPort int) (string, error)
	Remove(serviceID string) error
}

type traefikService struct {
	dynamicDir string
	appDomain  string
}

func NewTraefikService(dynamicDir, appDomain string) RouteManager {
	return &traefikService{
		dynamicDir: dynamicDir,
		appDomain:  strings.TrimSuffix(strings.ToLower(strings.TrimSpace(appDomain)), "."),
	}
}

func (s *traefikService) URLForService(serviceID, serviceName string) (string, error) {
	if s.dynamicDir == "" && s.appDomain == "" {
		return "", nil
	}
	if s.dynamicDir == "" || s.appDomain == "" {
		return "", fmt.Errorf("both CLOUDX_TRAEFIK_DYNAMIC_DIR and CLOUDX_APP_DOMAIN are required to publish deployments")
	}
	serviceUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return "", fmt.Errorf("invalid service ID for route: %w", err)
	}
	if err := validateDomain(s.appDomain); err != nil {
		return "", err
	}

	host, err := serviceHostname(serviceName, serviceUUID.String(), s.appDomain)
	if err != nil {
		return "", err
	}
	return "https://" + host, nil
}

func (s *traefikService) Publish(serviceID, serviceName, deploymentID string, containerPort int) (string, error) {
	if s.dynamicDir == "" && s.appDomain == "" {
		return "", nil
	}
	serviceUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return "", fmt.Errorf("invalid service ID for route: %w", err)
	}
	deploymentUUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return "", fmt.Errorf("invalid deployment ID for route: %w", err)
	}
	if containerPort < 1 || containerPort > 65535 {
		return "", fmt.Errorf("invalid container port %d for route", containerPort)
	}
	publicURL, err := s.URLForService(serviceID, serviceName)
	if err != nil {
		return "", err
	}
	host := strings.TrimPrefix(publicURL, "https://")
	routerName := "cloudx-" + strings.ReplaceAll(serviceUUID.String(), "-", "")
	containerName := deploymentContainerName(deploymentUUID)
	networkName := containerName

	connect := exec.Command("docker", "network", "connect", networkName, "cloudx-traefik")
	if output, err := connect.CombinedOutput(); err != nil {
		return "", fmt.Errorf("connect Traefik to deployment network: %w: %s", err, strings.TrimSpace(string(output)))
	}

	config := routeConfiguration(routerName, host, containerName, containerPort)
	if err := s.writeRoute(routerName, []byte(config)); err != nil {
		if disconnectErr := disconnectTraefik(networkName); disconnectErr != nil {
			return "", fmt.Errorf("write Traefik route: %w (network cleanup failed: %v)", err, disconnectErr)
		}
		return "", fmt.Errorf("write Traefik route: %w", err)
	}
	return publicURL, nil
}

func routeConfiguration(routerName, host, containerName string, containerPort int) string {
	return fmt.Sprintf(`http:
  routers:
    %s:
      rule: 'Host(%q)'
      entryPoints:
        - websecure
      service: %s
      tls:
        certResolver: le
  services:
    %s:
      loadBalancer:
        servers:
          - url: "http://%s:%d"
`, routerName, host, routerName, routerName, containerName, containerPort)
}

func (s *traefikService) Remove(serviceID string) error {
	if s.dynamicDir == "" {
		return nil
	}
	serviceUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return fmt.Errorf("invalid service ID for route removal: %w", err)
	}
	routePath := filepath.Join(s.dynamicDir, "cloudx-"+strings.ReplaceAll(serviceUUID.String(), "-", "")+".yml")
	if err := os.Remove(routePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove Traefik route: %w", err)
	}
	return nil
}

func (s *traefikService) writeRoute(name string, config []byte) error {
	if err := os.MkdirAll(s.dynamicDir, 0o755); err != nil {
		return fmt.Errorf("create Traefik dynamic configuration directory: %w", err)
	}
	file, err := os.CreateTemp(s.dynamicDir, name+"-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary Traefik route: %w", err)
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if _, err := file.Write(config); err != nil {
		_ = file.Close()
		return fmt.Errorf("write Traefik route: %w", err)
	}
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("set Traefik route permissions: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close Traefik route: %w", err)
	}
	routePath := filepath.Join(s.dynamicDir, name+".yml")
	if err := os.Rename(tempPath, routePath); err != nil {
		return fmt.Errorf("activate Traefik route: %w", err)
	}
	return nil
}

func validateDomain(domain string) error {
	if domain == "" || len(domain) > 253 {
		return fmt.Errorf("CLOUDX_APP_DOMAIN must be a valid DNS domain")
	}
	for _, label := range strings.Split(domain, ".") {
		if !hostnameLabelPattern.MatchString(label) {
			return fmt.Errorf("CLOUDX_APP_DOMAIN contains an invalid DNS label %q", label)
		}
	}
	return nil
}

func serviceHostname(name, serviceID, domain string) (string, error) {
	if err := validateDomain(domain); err != nil {
		return "", err
	}
	id := strings.ReplaceAll(serviceID, "-", "")
	if len(id) < 12 {
		return "", fmt.Errorf("service ID is too short to create a unique hostname")
	}
	var slug strings.Builder
	lastHyphen := false
	for _, char := range strings.ToLower(name) {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			slug.WriteRune(char)
			lastHyphen = false
		} else if slug.Len() > 0 && !lastHyphen {
			slug.WriteByte('-')
			lastHyphen = true
		}
	}
	label := strings.Trim(slug.String(), "-")
	if label == "" {
		label = "service"
	}
	if len(label) > 50 {
		label = strings.TrimRight(label[:50], "-")
	}
	host := label + "-" + id[:12] + "." + domain
	if err := validateDomain(host); err != nil {
		return "", err
	}
	return host, nil
}

func deploymentContainerName(id uuid.UUID) string {
	return "cloudx-deployment-" + strings.ReplaceAll(id.String(), "-", "")
}

func disconnectTraefik(networkName string) error {
	output, err := exec.Command("docker", "network", "disconnect", "-f", networkName, "cloudx-traefik").CombinedOutput()
	if err == nil || strings.Contains(string(output), "is not connected to network") {
		return nil
	}
	return fmt.Errorf("disconnect Traefik from %s: %w: %s", networkName, err, strings.TrimSpace(string(output)))
}
