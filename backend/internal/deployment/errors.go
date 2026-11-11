package deployment

// Status constants represent the lifecycle of a deployment.
const (
	StatusQueued     = "QUEUED"
	StatusBuilding   = "BUILDING"
	StatusDeploying  = "DEPLOYING"
	StatusRunning    = "RUNNING"
	StatusFailed     = "FAILED"
	StatusCancelled  = "CANCELLED"
	StatusSuperseded = "SUPERSEDED"
)
