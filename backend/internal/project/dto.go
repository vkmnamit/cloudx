package project

type CreateProjectRequest struct {
	OrganizationID string  `json:"organization_id" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Slug           string  `json:"slug" binding:"required"`
	Description    *string `json:"description"`
	Visibility     string  `json:"visibility"`
}

type UpdateProjectRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	Visibility  string  `json:"visibility"`
}
