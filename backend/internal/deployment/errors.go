package deployment

// Status constants represent the lifecycle of a deployment.
const (
	StatusPending  = "pending"
	StatusBuilding = "building"
	StatusRunning  = "running"
	StatusFailed   = "failed"
	StatusStopped  = "stopped"
)
