package middleware

import (
	"net/http"
	"strconv"

	"backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Handler validates JWT token and adds claims to context
func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			ctx.Abort()
			return
		}

		// Parse user_id string to integer for DB compatibility
		userID, err := strconv.Atoi(claims.UserID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Malformed claims: user_id is not integer"})
			ctx.Abort()
			return
		}

		// Add claims to context
		ctx.Set("user_id", userID)
		ctx.Set("email", claims.Email)
		ctx.Set(string(auth.ContextKeyRole), claims.Role)

		ctx.Next()
	}
}

// OptionalHandler validates token if present, but doesn't block if missing
func (m *AuthMiddleware) OptionalHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" {
			// Try to extract and validate token
			tokenString, err := auth.ExtractTokenFromHeader(authHeader)
			if err == nil {
				// Try to validate token
				claims, err := auth.ValidateToken(tokenString)
				if err == nil {
					// Token is valid, parse user_id and add to context
					userID, err := strconv.Atoi(claims.UserID)
					if err == nil {
						ctx.Set("user_id", userID)
						ctx.Set("email", claims.Email)
						ctx.Set(string(auth.ContextKeyRole), claims.Role)
					}
				}
			}
		}
		ctx.Next()
	}
}

// RequireRole checks if user has required role
func (m *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		roleVal, exists := ctx.Get(string(auth.ContextKeyRole))
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			ctx.Abort()
			return
		}

		userRole := roleVal.(string)
		hasRole := false
		for _, role := range requiredRoles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
