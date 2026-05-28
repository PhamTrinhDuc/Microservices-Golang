package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// GoogleClaims represents claims returned from Google TokenInfo API
type GoogleClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Aud           string `json:"aud"`
	Exp           string `json:"exp"`
}

// VerifyGoogleIDToken validates the Google ID token and returns its claims
func VerifyGoogleIDToken(idToken string) (*GoogleClaims, error) {
	if idToken == "" {
		return nil, errors.New("google ID token is empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, fmt.Errorf("failed to reach google token verification endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google id token verification returned status: %d", resp.StatusCode)
	}

	var claims GoogleClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse google token response: %w", err)
	}

	return &claims, nil
}
