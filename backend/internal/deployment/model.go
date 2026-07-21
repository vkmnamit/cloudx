package deployment

import "time"

// Deployment represents a deployment record in the database.
type Deployment struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ProjectName   string    `json:"project_name"`
	Repository    string    `json:"repository"`
	Framework     string    `json:"framework"`
	Language      string    `json:"language"`
	ImageName     string    `json:"image_name"`
	ContainerID   string    `json:"container_id"`
	HostPort      int       `json:"host_port"`
	ContainerPort int       `json:"container_port"`
	Status        string    `json:"status"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
