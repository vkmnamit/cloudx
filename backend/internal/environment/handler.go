package environment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Handler exposes env var endpoints.
type Handler struct {
	service EnvironmentService
	db      *pgx.Conn
}

func NewHandler(service EnvironmentService, db *pgx.Conn) *Handler {
	return &Handler{service: service, db: db}
}

// SetVars handles POST /deployments/:id/env
func (h *Handler) SetVars(c *gin.Context) {
	deploymentID := c.Param("id")
	if !h.ownsDeployment(c, deploymentID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

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
	if !h.ownsDeployment(c, deploymentID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

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

func (h *Handler) ownsDeployment(c *gin.Context, deploymentID string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var allowed bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM deployments d JOIN services s ON s.id=d.service_id JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE d.id=$1 AND (o.owner_id=$2 OR d.deployed_by=$2))`, deploymentID, userID).Scan(&allowed)
	return err == nil && allowed
}

func (h *Handler) SetServiceVars(c *gin.Context) {
	serviceID := c.Param("id")
	userID, _ := c.Get("user_id")
	var req SetEnvVarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Vars) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most 200 environment variables can be saved at once"})
		return
	}
	var allowed bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM services s JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE s.id=$1 AND o.owner_id=$2)`, serviceID, userID).Scan(&allowed)
	if err != nil || !allowed {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	if err := h.service.ReplaceVars(c.Request.Context(), serviceID, req.Vars); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service environment variables saved", "count": len(req.Vars)})
}

func (h *Handler) GetServiceVars(c *gin.Context) {
	serviceID := c.Param("id")
	userID, _ := c.Get("user_id")
	var allowed bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM services s JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE s.id=$1 AND o.owner_id=$2)`, serviceID, userID).Scan(&allowed)
	if err != nil || !allowed {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	vars, err := h.service.GetVars(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load environment variables"})
		return
	}
	for _, variable := range vars {
		if variable.IsSecret {
			variable.Value = ""
		}
	}
	c.JSON(http.StatusOK, vars)
}

func (h *Handler) DeleteServiceVar(c *gin.Context) {
	serviceID := c.Param("id")
	variableID := c.Param("varId")
	userID, _ := c.Get("user_id")
	result, err := h.db.Exec(c.Request.Context(), `DELETE FROM environment_variables ev USING services s, environments e, projects p, organizations o WHERE ev.id=$1 AND ev.service_id=$2 AND s.id=ev.service_id AND e.id=s.environment_id AND p.id=e.project_id AND o.id=p.organization_id AND o.owner_id=$3`, variableID, serviceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete environment variable"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment variable not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
