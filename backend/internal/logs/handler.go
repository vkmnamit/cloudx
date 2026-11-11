package logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler exposes log endpoints.
type Handler struct {
	service LogService
	db      *pgxpool.Pool
}

func NewHandler(service LogService, db *pgxpool.Pool) *Handler {
	return &Handler{service: service, db: db}
}

// GetLogs handles GET /deployments/:id/logs
func (h *Handler) GetLogs(c *gin.Context) {
	deploymentID := c.Param("id")
	userID, _ := c.Get("user_id")
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM deployments WHERE id=$1 AND deployed_by=$2)`, deploymentID, userID).Scan(&owned)
	if err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	entries, err := h.service.GetLogs(c.Request.Context(), deploymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}
