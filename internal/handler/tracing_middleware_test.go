package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTracingMiddleware_SetsRequestIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TracingMiddleware())
	router.GET("/health", HealthHandler)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestTracingMiddleware_PropagatesExistingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TracingMiddleware())
	router.GET("/health", HealthHandler)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "existing-id-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "existing-id-123" {
		t.Errorf("X-Request-ID = %q, want %q", got, "existing-id-123")
	}
}
