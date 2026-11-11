package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	db *pgxpool.Pool,
	jwtService *JWTService,
) {

	repository := NewRepository(db)

	service := NewService(repository, jwtService)

	handler := NewHandler(service)

	rg.POST("/signup", handler.Signup)
	rg.POST("/login", handler.Login)
}
