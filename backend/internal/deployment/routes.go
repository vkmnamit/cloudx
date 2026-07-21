package deployment

import (
	"github.com/gin-gonic/gin"
	"github.com/vkmnamit/cloudx/internal/auth"
)

func RegisterRoutes(router *gin.RouterGroup, h *Handler, jwtService *auth.JWTService) {
	deployments := router.Group("/deployments")
	deployments.Use(auth.AuthMiddleware(jwtService))

	deployments.POST("/", h.Deploy)
	deployments.GET("/", h.GetDeployments)
	deployments.GET("/:id", h.GetDeployment)
	deployments.DELETE("/:id", h.DeleteDeployment)
}
