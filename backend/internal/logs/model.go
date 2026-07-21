package logs

import "time"

// LogPhase identifies which phase of deployment a log entry belongs to.
type LogPhase string

const (
	PhaseClone   LogPhase = "clone"
	PhaseBuild   LogPhase = "build"
	PhaseRuntime LogPhase = "runtime"
)

// LogLevel mirrors the postgres log_level enum.
type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARNING"
	LevelError   LogLevel = "ERROR"
	LevelDebug   LogLevel = "DEBUG"
)

// DeploymentLog is a single log line for a deployment.
type DeploymentLog struct {
	ID           string    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Phase        LogPhase  `json:"phase"`
	Level        LogLevel  `json:"level"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}
