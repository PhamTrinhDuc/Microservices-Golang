package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
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

func (uc *UserUsecase) UpdateProfile(ctx context.Context, id int, req *domain.UpdateProfileRequest) (*domain.User, error) {
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u.FullName = req.FullName
	u.Phone = req.Phone
	u.Gender = req.Gender

	if req.DOB != nil && *req.DOB != "" {
		parsedTime, err := time.Parse("2006-01-02", *req.DOB)
		if err != nil {
			return nil, fmt.Errorf("invalid dob format, must be YYYY-MM-DD: %w", err)
		}
		u.DOB = &parsedTime
	} else {
		u.DOB = nil
	}

	if req.Avatar != nil {
		u.Avatar = req.Avatar
	}

	err = uc.repo.Update(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *UserUsecase) UpdatePassword(ctx context.Context, id int, req *domain.UpdatePasswordRequest) error {
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if u.Password != "google_oauth" {
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPassword))
		if err != nil {
			return domain.ErrInvalidPassword
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	u.Password = string(hashedPassword)

	return uc.repo.Update(ctx, u)
}
