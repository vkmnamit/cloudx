package deployment

import "github.com/vkmnamit/cloudx/internal/environment"

// DeployRequest is the payload sent by the client to trigger a deployment.
type DeployRequest struct {
	ServiceID string `json:"service_id"`

	EnvVars []environment.EnvVarInput `json:"env_vars"`
}

// DeployResponse is what we return after a deployment is created.
type DeployResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
