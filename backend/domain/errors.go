package domain

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already in use")
	ErrInvalidPassword = errors.New("invalid credentials")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrAddressNotFound = errors.New("address not found")
	ErrLocked          = errors.New("account is locked")
)
