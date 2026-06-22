package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"single-agent/internal/auth"
)

func TestNewFixedWindowLimiter(t *testing.T) {
	// Limiter setup
	limiter := NewFixedWindowLimiter(nil, 100, time.Minute)

	assert.NotNil(t, limiter)
	assert.Equal(t, 100, limiter.limit)
	assert.Equal(t, time.Minute, limiter.window)
}

func TestFixedWindowLimiter_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewFixedWindowLimiter(nil, 100, time.Minute)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(auth.ContextKeyUserID), "user123")
		c.Next()
	})
	r.Use(limiter.Handler())

	r.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get(string(auth.ContextKeyUserID))
		assert.True(t, exists)
		assert.Equal(t, "user123", userID)
		c.String(http.StatusOK, "success")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
