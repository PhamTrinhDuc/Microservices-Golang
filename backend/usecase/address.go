package usecase

import (
	"context"
	"errors"

	"backend/domain"
)

type AddressUsecase struct {
	repo domain.AddressRepository
}

func NewAddressUsecase(repo domain.AddressRepository) *AddressUsecase {
	return &AddressUsecase{repo: repo}
}

func (uc *AddressUsecase) Create(ctx context.Context, userID int, req *domain.CreateAddressRequest) (*domain.Address, error) {
	if req == nil {
		return nil, errors.New("address payload cannot be nil")
	}

	a := &domain.Address{
		UserID:        userID,
		FullName:      req.FullName,
		Phone:         req.Phone,
		District:      req.District,
		Province:      req.Province,
		Ward:          req.Ward,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	}

	return uc.repo.Create(ctx, a)
}

func (uc *AddressUsecase) List(ctx context.Context, userID int) ([]*domain.Address, error) {
	return uc.repo.ListByUserID(ctx, userID)
}

func (uc *AddressUsecase) SetDefault(ctx context.Context, userID, addressID int) error {
	a, err := uc.repo.GetByID(ctx, addressID)
	if err != nil {
		return err
	}

	if a.UserID != userID {
		return domain.ErrUnauthorized
	}

	return uc.repo.SetDefault(ctx, userID, addressID)
}

func (uc *AddressUsecase) Delete(ctx context.Context, userID, addressID int) error {
	a, err := uc.repo.GetByID(ctx, addressID)
	if err != nil {
		return err
	}

	if a.UserID != userID {
		return domain.ErrUnauthorized
	}

	return uc.repo.Delete(ctx, addressID)
}

func (uc *AddressUsecase) Update(ctx context.Context, userID, addressID int, req *domain.CreateAddressRequest) (*domain.Address, error) {
	if req == nil {
		return nil, errors.New("address payload cannot be nil")
	}

	a, err := uc.repo.GetByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if a.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	a.FullName = req.FullName
	a.Phone = req.Phone
	a.Province = req.Province
	a.District = req.District
	a.Ward = req.Ward
	a.DetailAddress = req.DetailAddress
	a.IsDefault = req.IsDefault

	updated, err := uc.repo.Update(ctx, a)
	if err != nil {
		return nil, err
	}

	if req.IsDefault {
		err = uc.repo.SetDefault(ctx, userID, addressID)
		if err != nil {
			return nil, err
		}
		updated.IsDefault = true
	}

	return updated, nil
}
