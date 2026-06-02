package domain

import (
	"context"
	"time"
)

type User struct {
	ID         int       `json:"id" db:"id"`
	FullName   string    `json:"full_name" db:"full_name"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	Phone      *string   `json:"phone" db:"phone"`
	Gender     *string   `json:"gender" db:"gender"`
	DOB        *time.Time `json:"dob,omitempty" db:"dob"`
	Role       string    `json:"role" db:"role"`
	Avatar     *string   `json:"avatar,omitempty" db:"avatar"`
	IsLock     bool      `json:"is_lock" db:"is_lock"`
	IsVerified bool      `json:"is_verified" db:"is_verified"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type GoogleLoginRequest struct {
	Credential string `json:"credential" validate:"required"`
}

type UserRepository interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateLockStatus(ctx context.Context, id int, isLock bool) error
	List(ctx context.Context, page, limit int, query string) ([]*User, int, error)
}

type UserUsecase interface {
	Register(ctx context.Context, req *RegisterRequest) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, string, error)
	LoginOrRegisterWithGoogle(ctx context.Context, idToken string) (*User, string, error)
	GetByID(ctx context.Context, id int) (*User, error)
	LockUser(ctx context.Context, id int, isLock bool) error
	List(ctx context.Context, page, limit int, query string) ([]*User, int, error)
}
