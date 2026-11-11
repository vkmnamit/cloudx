package project

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
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	var owned bool
	if err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM organizations WHERE id=$1 AND owner_id=$2)`, req.OrganizationID, userID).Scan(&owned); err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	project, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *Handler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projects, err := h.service.ListByOwner(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsProject(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	project, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsProject(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsProject(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted successfully"})
}

func (h *Handler) ownsProject(c *gin.Context, id string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.id=$1 AND o.owner_id=$2)`, id, userID).Scan(&owned)
	return err == nil && owned
}
