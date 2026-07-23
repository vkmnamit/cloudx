package deployment

import "time"

// Deployment represents a deployment record in the database.
type Deployment struct {
	ID            string     `json:"id"`
	ServiceID     string     `json:"service_id"`
	CommitSHA     string     `json:"commit_sha,omitempty"`
	CommitMessage string     `json:"commit_message,omitempty"`
	ImageTag      string     `json:"image_tag,omitempty"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	DeployedBy    string     `json:"deployed_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
