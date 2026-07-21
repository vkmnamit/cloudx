package logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes log endpoints.
type Handler struct {
	service LogService
}

func NewHandler(service LogService) *Handler {
	return &Handler{service: service}
}

// GetLogs handles GET /deployments/:id/logs
func (h *Handler) GetLogs(c *gin.Context) {
	deploymentID := c.Param("id")

	entries, err := h.service.GetLogs(c.Request.Context(), deploymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}
