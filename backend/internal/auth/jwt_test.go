package auth_test

import (
	"testing"

	"backend/internal/auth"
)

func TestJWTTokenGenerationDisabled(t *testing.T) {
	_, err := auth.GenerateToken("42", "test@example.com", "customer")
	if err == nil {
		t.Fatal("expected error since local token generation is disabled, got nil")
	}
	if err.Error() != "local token generation is disabled, use Keycloak to obtain tokens" {
		t.Errorf("unexpected error message: %v", err)
	}
}
