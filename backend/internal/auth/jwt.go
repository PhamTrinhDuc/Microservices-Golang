package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// EnsureKeysExist creates RSA keys if they don't exist.
func EnsureKeysExist(dir string) error {
	privPath := filepath.Join(dir, "private_key.pem")
	pubPath := filepath.Join(dir, "public_key.pem")

	// If keys already exist, do nothing
	if _, err := os.Stat(privPath); err == nil {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save Private Key
	privFile, err := os.Create(privPath)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privFile, privBlock); err != nil {
		return err
	}

	// Save Public Key
	pubFile, err := os.Create(pubPath)
	if err != nil {
		return err
	}
	defer pubFile.Close()

	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	if err := pem.Encode(pubFile, pubBlock); err != nil {
		return err
	}

	return nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header.
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format, expected: Bearer <token>")
	}

	return parts[1], nil
}

// ValidateToken validates a JWT token string using RSA or HMAC fallback.
func ValidateToken(tokenString string) (*JWTClaims, error) {
	keysDir := os.Getenv("DEMO_KEYS_DIR")
	if keysDir == "" {
		keysDir = "tmp/demo-keys"
	}
	pubKeyPath := filepath.Join(keysDir, "public_key.pem")

	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		data, readErr := os.ReadFile(pubKeyPath)

		// RSA branch: if public key file exists and is readable
		if readErr == nil {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			block, _ := pem.Decode(data)
			if block == nil {
				return nil, fmt.Errorf("failed to decode PEM block from public key file")
			}

			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key: %w", err)
			}

			rsaPub, ok := pub.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("key is not an RSA public key")
			}

			return rsaPub, nil
		}

		// HMAC branch: fallback to HMAC if keyfile does not exist
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "your-secret-key-change-this-in-production"
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	return claims, nil
}

// GenerateTokenWithPrivateKey signs a token using the RSA private key (used for testing or admin creation).
func GenerateTokenWithPrivateKey(userID, email, role string) (string, error) {
	keysDir := os.Getenv("DEMO_KEYS_DIR")
	if keysDir == "" {
		keysDir = "tmp/demo-keys"
	}
	privKeyPath := filepath.Join(keysDir, "private_key.pem")

	privKeyData, err := os.ReadFile(privKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privKeyData)
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privKey)
}

// GenerateToken signs a token using RS256 if key exists, otherwise HS256.
func GenerateToken(userID string, email, role string) (string, error) {
	keysDir := os.Getenv("DEMO_KEYS_DIR")
	if keysDir == "" {
		keysDir = "tmp/demo-keys"
	}
	privKeyPath := filepath.Join(keysDir, "private_key.pem")

	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	privKeyData, err := os.ReadFile(privKeyPath)
	if err == nil {
		block, _ := pem.Decode(privKeyData)
		if block != nil {
			privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				return token.SignedString(privKey)
			}
		}
	}

	// Fallback to HMAC
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-in-production"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
