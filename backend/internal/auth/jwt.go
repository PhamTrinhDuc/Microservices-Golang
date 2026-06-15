package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	ContextKeyUserID ContextKey = "user_id"
	ContextKeyEmail  ContextKey = "email"
	ContextKeyRole   ContextKey = "role"
)

// JWTClaims represents JWT token claims.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}



// ExtractTokenFromHeader extracts JWT token from Authorization header.
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", fmt.Errorf("invalid authorization header format, expected: Bearer <token>")
	}

	return parts[1], nil
}

var (
	oidcOnce     sync.Once
	oidcVerifier *oidc.IDTokenVerifier
	oidcErr      error
)

func getOIDCVerifier() (*oidc.IDTokenVerifier, error) {
	issuer := os.Getenv("KEYCLOAK_ISSUER_URL")
	if issuer == "" {
		return nil, fmt.Errorf("KEYCLOAK_ISSUER_URL is not set")
	}

	oidcOnce.Do(func() {
		ctx := context.Background()
		var provider *oidc.Provider
		provider, oidcErr = oidc.NewProvider(ctx, issuer)
		if oidcErr == nil {
			oidcVerifier = provider.Verifier(&oidc.Config{
				SkipClientIDCheck: true,
			})
		}
	})

	if oidcErr != nil {
		oidcOnce = sync.Once{} // Allow retry if Keycloak was down
		return nil, oidcErr
	}

	return oidcVerifier, nil
}

// ValidateToken validates a JWT token string using Keycloak OIDC.
func ValidateToken(tokenString string) (*JWTClaims, error) {
	verifier, err := getOIDCVerifier()
	if err != nil {
		return nil, fmt.Errorf("OIDC verifier not initialized: %w", err)
	}

	ctx := context.Background()
	idToken, verifyErr := verifier.Verify(ctx, tokenString)
	if verifyErr != nil {
		return nil, fmt.Errorf("invalid token: %w", verifyErr)
	}

	var kcClaims struct {
		Sub         string   `json:"sub"`
		Email       string   `json:"email"`
		Name        string   `json:"name"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}
	if err := idToken.Claims(&kcClaims); err == nil {
		role := "customer"
		for _, r := range kcClaims.RealmAccess.Roles {
			if r == "admin" {
				role = "admin"
				break
			}
		}

		return &JWTClaims{
			UserID: kcClaims.Sub,
			Email:  kcClaims.Email,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   kcClaims.Sub,
				ExpiresAt: jwt.NewNumericDate(idToken.Expiry),
				IssuedAt:  jwt.NewNumericDate(idToken.IssuedAt),
			},
		}, nil
	}

	return nil, fmt.Errorf("failed to parse Keycloak token claims")
}

// GenerateToken is disabled. Tokens must be obtained directly from Keycloak.
func GenerateToken(userID string, email, role string) (string, error) {
	return "", fmt.Errorf("local token generation is disabled, use Keycloak to obtain tokens")
}
