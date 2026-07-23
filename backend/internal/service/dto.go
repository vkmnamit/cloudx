package service

type CreateServiceRequest struct {
	EnvironmentID  string `json:"environment_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Runtime        string `json:"runtime" binding:"required"`
	RepositoryURL  string `json:"repository_url"`
	Branch         string `json:"branch"`
	DockerfilePath string `json:"dockerfile_path"`
	BuildContext   string `json:"build_context"`
	Port           int    `json:"port"`
	HealthCheck    string `json:"health_check"`
	CPULimit       string `json:"cpu_limit"`
	MemoryLimit    string `json:"memory_limit"`
	Replicas       int    `json:"replicas"`
	AutoDeploy     *bool  `json:"auto_deploy"`
}

type UpdateServiceRequest struct {
	Name           string `json:"name"`
	Runtime        string `json:"runtime"`
	RepositoryURL  string `json:"repository_url"`
	Branch         string `json:"branch"`
	DockerfilePath string `json:"dockerfile_path"`
	BuildContext   string `json:"build_context"`
	Port           int    `json:"port"`
	HealthCheck    string `json:"health_check"`
	CPULimit       string `json:"cpu_limit"`
	MemoryLimit    string `json:"memory_limit"`
	Replicas       int    `json:"replicas"`
	AutoDeploy     *bool  `json:"auto_deploy"`
}
