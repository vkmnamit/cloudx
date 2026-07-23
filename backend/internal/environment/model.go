package environment

import "time"

// EnvVar represents a single environment variable bound to a deployment.
type EnvVar struct {
	ID        string
	ServiceID string

	Key   string
	Value string

	IsSecret bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetEnvVarsRequest is sent by the client to set multiple env vars at once.
type SetEnvVarsRequest struct {
	Vars []EnvVarInput `json:"vars" binding:"required"`
}

// EnvVarInput is a single key/value pair inside a SetEnvVarsRequest.
type EnvVarInput struct {
	Key      string `json:"key"      binding:"required"`
	Value    string `json:"value"    binding:"required"`
	IsSecret bool   `json:"is_secret"`
}
