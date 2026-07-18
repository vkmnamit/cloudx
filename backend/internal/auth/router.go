package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func RegisterRoutes(rg *gin.RouterGroup, db *pgx.Conn) {

	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	rg.POST("/signup", handler.Signup)

	// Future
	// rg.POST("/login", handler.Login)
	// rg.POST("/refresh", handler.Refresh)
}
