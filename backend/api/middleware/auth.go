package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
	db *pgxpool.Pool
}

func NewAuthMiddleware(db *pgxpool.Pool) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) getOrCreateUserID(ctx context.Context, claims *auth.JWTClaims) (int, string, error) {
	// 1. Try parsing claims.UserID as an integer (for backwards compatibility / local tests)
	if id, err := strconv.Atoi(claims.UserID); err == nil {
		return id, claims.Role, nil
	}

	// 2. Otherwise it is a Keycloak UUID. Find user by keycloak_id
	var userID int
	var role string
	err := m.db.QueryRow(ctx, "SELECT id, role FROM users WHERE keycloak_id = $1", claims.UserID).Scan(&userID, &role)
	if err == nil {
		return userID, role, nil
	}

	// 3. User not found by keycloak_id. Check by email to link existing user
	var existingID int
	var existingKeycloakID *string
	err = m.db.QueryRow(ctx, "SELECT id, keycloak_id FROM users WHERE email = $1", claims.Email).Scan(&existingID, &existingKeycloakID)
	if err == nil {
		// Found user by email. Link Keycloak ID if not set
		if existingKeycloakID == nil || *existingKeycloakID == "" {
			_, err = m.db.Exec(ctx, "UPDATE users SET keycloak_id = $1, updated_at = NOW() WHERE id = $2", claims.UserID, existingID)
			if err != nil {
				return 0, "", fmt.Errorf("failed to link keycloak id: %w", err)
			}
		}
		var updatedRole string
		_ = m.db.QueryRow(ctx, "SELECT role FROM users WHERE id = $1", existingID).Scan(&updatedRole)
		return existingID, updatedRole, nil
	}

	// 4. User not found by email. Create a new user record
	fullName := claims.Email
	if parts := strings.Split(claims.Email, "@"); len(parts) > 0 {
		fullName = parts[0]
	}

	err = m.db.QueryRow(ctx, 
		`INSERT INTO users (keycloak_id, full_name, email, role, is_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		 RETURNING id`,
		claims.UserID, fullName, claims.Email, claims.Role,
	).Scan(&userID)

	if err != nil {
		return 0, "", fmt.Errorf("failed to create synchronized user: %w", err)
	}

	return userID, claims.Role, nil
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

		// Get or create integer user_id for DB compatibility
		userID, userRole, err := m.getOrCreateUserID(ctx.Request.Context(), claims)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to synchronize user: " + err.Error()})
			ctx.Abort()
			return
		}

		// Add claims to context
		ctx.Set("user_id", userID)
		ctx.Set("email", claims.Email)
		ctx.Set(string(auth.ContextKeyRole), userRole)

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
					// Token is valid, parse/sync user_id and add to context
					userID, userRole, err := m.getOrCreateUserID(ctx.Request.Context(), claims)
					if err == nil {
						ctx.Set("user_id", userID)
						ctx.Set("email", claims.Email)
						ctx.Set(string(auth.ContextKeyRole), userRole)
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
