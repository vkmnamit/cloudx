package projectenv

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service Service
	db      *pgxpool.Pool
}

func NewHandler(service Service, db *pgxpool.Pool) *Handler {
	return &Handler{service: service, db: db}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.id=$1 AND o.owner_id=$2)`, req.ProjectID, userID).Scan(&owned)
	if err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	env, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, env)
}

func (h *Handler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	envs, err := h.service.ListByOwner(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, envs)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsEnvironment(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}
	env, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}

	c.JSON(http.StatusOK, env)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsEnvironment(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}
	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, env)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsEnvironment(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "environment deleted successfully"})
}

func (h *Handler) ownsEnvironment(c *gin.Context, id string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM environments e JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE e.id=$1 AND o.owner_id=$2)`, id, userID).Scan(&owned)
	return err == nil && owned
}
