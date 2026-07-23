package organization

type CreateOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
	Plan string `json:"plan"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}
