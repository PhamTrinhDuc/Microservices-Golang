package auth_test

import (
	"os"
	"testing"

	"backend/internal/auth"
)

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	// Set test environment variable
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
	os.Setenv("DEMO_KEYS_DIR", "tmp/test-keys")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DEMO_KEYS_DIR")
		os.RemoveAll("tmp/test-keys")
	}()

	userID := "42"
	email := "test@example.com"
	role := "customer"

	// 1. Generate Token
	tokenStr, err := auth.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected token to be non-empty string")
	}

	// 2. Validate Token
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("expected no error validating token, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID to be %q, got %q", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("expected Email to be %q, got %q", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("expected Role to be %q, got %q", role, claims.Role)
	}
}
