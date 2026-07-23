package service

import "time"

// ServiceRecord represents a deployable service from the services table.
type ServiceRecord struct {
	ID             string    `json:"id"`
	EnvironmentID  string    `json:"environment_id"`
	Name           string    `json:"name"`
	Runtime        string    `json:"runtime"`
	RepositoryURL  string    `json:"repository_url"`
	Branch         string    `json:"branch"`
	DockerfilePath string    `json:"dockerfile_path"`
	BuildContext   string    `json:"build_context"`
	Port           int       `json:"port"`
	HealthCheck    string    `json:"health_check,omitempty"`
	CPULimit       string    `json:"cpu_limit,omitempty"`
	MemoryLimit    string    `json:"memory_limit,omitempty"`
	Replicas       int       `json:"replicas"`
	AutoDeploy     bool      `json:"auto_deploy"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
