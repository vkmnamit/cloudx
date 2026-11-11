package organization

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
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	org, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

func (h *Handler) List(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	orgs, err := h.service.ListByOwner(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsOrganization(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	org, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsOrganization(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !h.ownsOrganization(c, id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

func (h *Handler) ownsOrganization(c *gin.Context, id string) bool {
	userID, ok := c.Get("user_id")
	if !ok {
		return false
	}
	var owned bool
	err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM organizations WHERE id=$1 AND owner_id=$2)`, id, userID).Scan(&owned)
	return err == nil && owned
}
