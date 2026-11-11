package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	service Service
	db      *pgx.Conn
}

func NewHandler(service Service, db *pgx.Conn) *Handler {
	return &Handler{service: service, db: db}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM environments e JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE e.id=$1 AND o.owner_id=$2)`, req.EnvironmentID, userID).Scan(&owned)
	if err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}

	record, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *Handler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	records, err := h.service.ListByOwner(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsService(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	record, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsService(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsService(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service deleted successfully"})
}

func (h *Handler) ownsService(c *gin.Context, id string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM services s JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE s.id=$1 AND o.owner_id=$2)`, id, userID).Scan(&owned)
	return err == nil && owned
}
