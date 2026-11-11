package deployment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler handles HTTP requests and delegates to the Service.
// It never contains business logic.
type Handler struct {
	service Service
	db      *pgxpool.Pool
}

func NewHandler(service Service, db *pgxpool.Pool) *Handler {
	return &Handler{service: service, db: db}
}

// Deploy handles POST /deployments
func (h *Handler) Deploy(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}
	var ownsService bool
	if err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM services s JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE s.id=$1 AND o.owner_id=$2)`, req.ServiceID, userID).Scan(&ownsService); err != nil || !ownsService {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	dep, err := h.service.Deploy(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dep)
}

// GetDeployments handles GET /deployments
func (h *Handler) GetDeployments(c *gin.Context) {
	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	deps, err := h.service.ListDeployments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deps)
}

// GetDeployment handles GET /deployments/:id
func (h *Handler) GetDeployment(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsDeployment(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

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
	if !h.ownsDeployment(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	if err := h.service.DeleteDeployment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment deleted"})
}

func (h *Handler) ownsDeployment(c *gin.Context, id string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM deployments WHERE id=$1 AND deployed_by=$2)`, id, userID).Scan(&owned)
	return err == nil && owned
}
