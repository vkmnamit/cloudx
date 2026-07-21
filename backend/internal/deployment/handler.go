package deployment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests and delegates to the Service.
// It never contains business logic.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Deploy handles POST /deployments
func (h *Handler) Deploy(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	dep, err := h.service.Deploy(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dep)
}

// GetDeployments handles GET /deployments
func (h *Handler) GetDeployments(c *gin.Context) {
	userID, _ := c.Get("user_id")

	deps, err := h.service.ListDeployments(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deps)
}

// GetDeployment handles GET /deployments/:id
func (h *Handler) GetDeployment(c *gin.Context) {
	id := c.Param("id")

	dep, err := h.service.GetDeployment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	c.JSON(http.StatusOK, dep)
}

// DeleteDeployment handles DELETE /deployments/:id
func (h *Handler) DeleteDeployment(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteDeployment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment deleted"})
}
