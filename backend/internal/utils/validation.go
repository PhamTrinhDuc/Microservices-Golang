package utils

import (
	"backend/domain"
	"regexp"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if email matches standard structure
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return domain.ErrInvalidEmail
	}
	return nil
}

// ValidatePasswordStrength ensures password has at least 8 characters,
// one uppercase, one lowercase, one number, and one special character.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsNumber(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return domain.ErrWeakPassword
	}

	return nil
}
