package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"single-agent/internal/auth"
)

func TestAuthMiddleware_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing Token", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()
		r.Use(middleware.Handler())

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()
		r.Use(middleware.Handler())

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_OptionalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing Token - Proceeds Without Claims", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()
		r.Use(middleware.OptionalHandler())

		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get(string(auth.ContextKeyUserID))
			assert.False(t, exists)
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Has Required Role", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()

		// Mock role in context
		r.Use(func(c *gin.Context) {
			c.Set(string(auth.ContextKeyRole), "admin")
			c.Next()
		})
		r.Use(middleware.RequireRole("admin"))

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Insufficient Role", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()

		// Mock role in context
		r.Use(func(c *gin.Context) {
			c.Set(string(auth.ContextKeyRole), "user")
			c.Next()
		})
		r.Use(middleware.RequireRole("admin"))

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Missing Role", func(t *testing.T) {
		r := gin.New()
		middleware := NewAuthMiddleware()

		r.Use(middleware.RequireRole("admin"))

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
