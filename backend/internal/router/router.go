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
	"github.com/vkmnamit/cloudx/internal/organization"
	"github.com/vkmnamit/cloudx/internal/project"
	"github.com/vkmnamit/cloudx/internal/projectenv"
	"github.com/vkmnamit/cloudx/internal/runtime"
	projectservice "github.com/vkmnamit/cloudx/internal/service"
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

	// ─── Organizations ────────────────────────────────────────────
	orgRepo := organization.NewRepository(db)
	orgSvc := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgSvc)

	protected.POST("/organizations", orgHandler.Create)
	protected.GET("/organizations", orgHandler.List)
	protected.GET("/organizations/:id", orgHandler.GetByID)
	protected.PUT("/organizations/:id", orgHandler.Update)
	protected.DELETE("/organizations/:id", orgHandler.Delete)

	// ─── Projects ─────────────────────────────────────────────────
	projRepo := project.NewRepository(db)
	projSvc := project.NewService(projRepo)
	projHandler := project.NewHandler(projSvc)

	protected.POST("/projects", projHandler.Create)
	protected.GET("/projects", projHandler.List)
	protected.GET("/projects/:id", projHandler.GetByID)
	protected.PUT("/projects/:id", projHandler.Update)
	protected.DELETE("/projects/:id", projHandler.Delete)

	// ─── Environments ──────────────────────────────────────────────
	envProjRepo := projectenv.NewRepository(db)
	envProjSvc := projectenv.NewService(envProjRepo)
	envProjHandler := projectenv.NewHandler(envProjSvc)

	protected.POST("/environments", envProjHandler.Create)
	protected.GET("/environments", envProjHandler.List)
	protected.GET("/environments/:id", envProjHandler.GetByID)
	protected.PUT("/environments/:id", envProjHandler.Update)
	protected.DELETE("/environments/:id", envProjHandler.Delete)

	// ─── Services ─────────────────────────────────────────────────
	serviceRepo := projectservice.NewRepository(db)
	serviceSvc := projectservice.NewService(serviceRepo)
	serviceHandler := projectservice.NewHandler(serviceSvc)

	protected.POST("/services", serviceHandler.Create)
	protected.GET("/services", serviceHandler.List)
	protected.GET("/services/:id", serviceHandler.GetByID)
	protected.PUT("/services/:id", serviceHandler.Update)
	protected.DELETE("/services/:id", serviceHandler.Delete)

	// ─── Logs ─────────────────────────────────────────────────────
	logRepo := logs.NewRepository(db)
	logSvc := logs.NewLogService(logRepo)
	logHandler := logs.NewHandler(logSvc)
	// GET /api/deployments/:id/logs
	protected.GET("/deployments/:id/logs", logHandler.GetLogs)

	// ─── Deployment Environment Variables ─────────────────────────
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

	deployService := deployment.NewService(deployRepo, serviceSvc, gitSvc, analyzerSvc, dockerSvc, runtimeSvc, logSvc, envSvc)
	deployHandler := deployment.NewHandler(deployService)
	deployment.RegisterRoutes(protected, deployHandler, jwtService)

	return r
}
