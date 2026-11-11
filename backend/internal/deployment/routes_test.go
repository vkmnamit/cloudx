package deployment

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vkmnamit/cloudx/internal/auth"
)

func TestDeploymentsRouteWithoutTrailingSlash(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router.Group("/api"), &Handler{}, auth.NewJWTService("test-secret"))

	request := httptest.NewRequest(http.MethodPost, "/api/deployments", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected the deployment route to reach auth middleware with status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
