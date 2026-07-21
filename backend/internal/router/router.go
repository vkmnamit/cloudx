package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/vkmnamit/cloudx/configs"
	"github.com/vkmnamit/cloudx/internal/analyzer"
	"github.com/vkmnamit/cloudx/internal/auth"
	"github.com/vkmnamit/cloudx/internal/deployment"
	"github.com/vkmnamit/cloudx/internal/docker"
	"github.com/vkmnamit/cloudx/internal/environment"
	"github.com/vkmnamit/cloudx/internal/git"
	"github.com/vkmnamit/cloudx/internal/logs"
	"github.com/vkmnamit/cloudx/internal/runtime"
)

func Setup(db *pgx.Conn) *gin.Engine {

	r := gin.Default()
	api := r.Group("/api")

	// ─── Shared Services ──────────────────────────────────────────
	jwtService := auth.NewJWTService(configs.AppConfig.JWTSecret)

	// ─── Public Routes ────────────────────────────────────────────
	auth.RegisterRoutes(api, db, jwtService)

	// ─── Protected Routes ─────────────────────────────────────────
	protected := api.Group("/")
	protected.Use(auth.AuthMiddleware(jwtService))

	protected.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "email": email})
	})

	// ─── Logs ─────────────────────────────────────────────────────
	logRepo := logs.NewRepository(db)
	logSvc := logs.NewLogService(logRepo)
	logHandler := logs.NewHandler(logSvc)
	// GET /api/deployments/:id/logs
	protected.GET("/deployments/:id/logs", logHandler.GetLogs)

	// ─── Environment ──────────────────────────────────────────────
	envRepo := environment.NewRepository(db)
	envSvc := environment.NewEnvironmentService(envRepo)
	envHandler := environment.NewHandler(envSvc)
	// POST /api/deployments/:id/env
	protected.POST("/deployments/:id/env", envHandler.SetVars)
	// GET  /api/deployments/:id/env
	protected.GET("/deployments/:id/env", envHandler.GetVars)

	// ─── Deployment (orchestrator) ────────────────────────────────
	deployRepo := deployment.NewRepository(db)
	gitSvc := git.NewGitService("/tmp/cloudx/repos")
	analyzerSvc := analyzer.NewAnalyzerService()
	dockerSvc := docker.NewDockerService()
	runtimeSvc := runtime.NewRuntimeService()

	deployService := deployment.NewService(deployRepo, gitSvc, analyzerSvc, dockerSvc, runtimeSvc, logSvc, envSvc)
	deployHandler := deployment.NewHandler(deployService)
	deployment.RegisterRoutes(protected, deployHandler, jwtService)

	return r
}
