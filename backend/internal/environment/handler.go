package environment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes env var endpoints.
type Handler struct {
	service EnvironmentService
}

func NewHandler(service EnvironmentService) *Handler {
	return &Handler{service: service}
}

// SetVars handles POST /deployments/:id/env
func (h *Handler) SetVars(c *gin.Context) {
	deploymentID := c.Param("id")

	var req SetEnvVarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SetVars(c.Request.Context(), deploymentID, req.Vars); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "environment variables saved"})
}

// GetVars handles GET /deployments/:id/env
// Secret values are redacted before being returned to the client.
func (h *Handler) GetVars(c *gin.Context) {
	deploymentID := c.Param("id")

	vars, err := h.service.GetVars(c.Request.Context(), deploymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redact secret values so they never leave the server in plaintext
	for _, v := range vars {
		if v.IsSecret {
			v.Value = "***"
		}
	}

	c.JSON(http.StatusOK, vars)
}
