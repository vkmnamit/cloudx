package projectenv

type CreateEnvironmentRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type UpdateEnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
}
