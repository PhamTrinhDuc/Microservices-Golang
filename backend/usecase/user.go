package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"backend/domain"
	"backend/internal/auth"
	"backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo              domain.UserRepository
	verifyGoogleToken func(idToken string) (*auth.GoogleClaims, error)
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo:              repo,
		verifyGoogleToken: auth.VerifyGoogleIDToken,
	}
}

func (uc *UserUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	if req == nil {
		return nil, errors.New("registration payload cannot be nil")
	}

	// Validate email format
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Email duplicate check
	_, err := uc.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, domain.ErrEmailTaken
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	// Encrypt password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &domain.User{
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   string(hashedPassword),
		Role:       "customer", // default role
		IsLock:     false,
		IsVerified: false,
	}

	return uc.repo.Create(ctx, u)
}

func (uc *UserUsecase) Authenticate(ctx context.Context, email, password string) (*domain.User, string, error) {
	u, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidPassword
		}
		return nil, "", fmt.Errorf("failed to query user: %w", err)
	}

	if u.IsLock {
		return nil, "", domain.ErrLocked
	}

	// Compare bcrypt hashes
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return nil, "", domain.ErrInvalidPassword
	}

	// Generate access token
	tokenStr, err := auth.GenerateToken(strconv.Itoa(u.ID), u.Email, u.Role)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign credentials: %w", err)
	}

	return u, tokenStr, nil
}

func (uc *UserUsecase) LoginOrRegisterWithGoogle(ctx context.Context, idToken string) (*domain.User, string, error) {
	claims, err := uc.verifyGoogleToken(idToken)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", domain.ErrGoogleLoginFailed, err)
	}

	u, err := uc.repo.GetByEmail(ctx, claims.Email)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", fmt.Errorf("failed to query user by email: %w", err)
		}

		// User does not exist, automatically register them
		fullName := claims.Name
		if fullName == "" {
			fullName = claims.Email
		}

		var avatar *string
		if claims.Picture != "" {
			avatar = &claims.Picture
		}

		newUser := &domain.User{
			FullName:   fullName,
			Email:      claims.Email,
			Password:   "google_oauth", // Simple placeholder password
			Avatar:     avatar,
			Role:       "customer",
			IsLock:     false,
			IsVerified: true, // trusted verified via Google
		}

		u, err = uc.repo.Create(ctx, newUser)
		if err != nil {
			return nil, "", fmt.Errorf("failed to automatically create Google user: %w", err)
		}
	} else {
		// User exists, verify that the account is not locked
		if u.IsLock {
			return nil, "", domain.ErrLocked
		}
	}

	// Generate token for either registered or logged in user
	tokenStr, err := auth.GenerateToken(strconv.Itoa(u.ID), u.Email, u.Role)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign Google session credentials: %w", err)
	}

	return u, tokenStr, nil
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) LockUser(ctx context.Context, id int, isLock bool) error {
	return uc.repo.UpdateLockStatus(ctx, id, isLock)
}

func (uc *UserUsecase) List(ctx context.Context, page, limit int, query string) ([]*domain.User, int, error) {
	return uc.repo.List(ctx, page, limit, query)
}
