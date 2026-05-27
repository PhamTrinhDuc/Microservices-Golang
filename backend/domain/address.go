package domain

import "context"

type Address struct {
	ID            int    `json:"id" db:"id"`
	UserID        int    `json:"user_id" db:"user_id"`
	FullName      string `json:"full_name" db:"full_name"`
	Phone         string `json:"phone" db:"phone"`
	District      string `json:"district" db:"district"`
	Province      string `json:"province" db:"province"`
	Ward          string `json:"ward" db:"ward"`
	DetailAddress string `json:"detail_address" db:"detail_address"`
	IsDefault     bool   `json:"is_default" db:"is_default"`
	IsDeleted     bool   `json:"is_deleted" db:"is_deleted"`
}

type CreateAddressRequest struct {
	FullName      string `json:"full_name" validate:"required"`
	Phone         string `json:"phone" validate:"required"`
	District      string `json:"district" validate:"required"`
	Province      string `json:"province" validate:"required"`
	Ward          string `json:"ward" validate:"required"`
	DetailAddress string `json:"detail_address" validate:"required"`
	IsDefault     bool   `json:"is_default"`
}

type AddressRepository interface {
	Create(ctx context.Context, a *Address) (*Address, error)
	GetByID(ctx context.Context, id int) (*Address, error)
	ListByUserID(ctx context.Context, userID int) ([]*Address, error)
	Update(ctx context.Context, a *Address) (*Address, error)
	SetDefault(ctx context.Context, userID, addressID int) error
	Delete(ctx context.Context, id int) error
}

type AddressUsecase interface {
	Create(ctx context.Context, userID int, req *CreateAddressRequest) (*Address, error)
	List(ctx context.Context, userID int) ([]*Address, error)
	SetDefault(ctx context.Context, userID, addressID int) error
	Delete(ctx context.Context, userID, addressID int) error
}
