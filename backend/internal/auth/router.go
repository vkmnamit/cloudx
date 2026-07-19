package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/vkmnamit/cloudx/configs"
)

func RegisterRoutes(rg *gin.RouterGroup, db *pgx.Conn) {

	repository := NewRepository(db)

	jwtService := NewJWTService(configs.AppConfig.JWTSecret)

	service := NewService(repository, jwtService)

	handler := NewHandler(service)

	rg.POST("/signup", handler.Signup)
	rg.POST("/login", handler.Login)

	// Future
	// rg.POST("/refresh", handler.Refresh)
}
