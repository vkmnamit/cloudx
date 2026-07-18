package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/vkmnamit/cloudx/internal/auth"
)

func Setup(db *pgx.Conn) *gin.Engine {

	r := gin.Default()

	api := r.Group("/api")

	auth.RegisterRoutes(api, db)

	return r
}
