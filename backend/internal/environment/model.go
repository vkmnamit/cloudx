package environment

import "time"

// EnvVar represents a single environment variable bound to a deployment.
type EnvVar struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetEnvVarsRequest is sent by the client to set multiple env vars at once.
type SetEnvVarsRequest struct {
	Vars []EnvVarInput `json:"vars" binding:"required"`
}

// EnvVarInput is a single key/value pair inside a SetEnvVarsRequest.
type EnvVarInput struct {
	Key      string `json:"key"      binding:"required"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}
