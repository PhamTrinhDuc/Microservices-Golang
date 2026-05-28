package domain

import (
	"context"
	"time"
)

// Models

type Promotion struct {
	ID            int       `json:"id" db:"id"`
	ProductID     string    `json:"product_id" db:"product_id"`
	VariantID     *int      `json:"variant_id" db:"variant_id"`
	Name          string    `json:"name" db:"name"`
	Description   *string   `json:"description" db:"description"`
	DiscountType  string    `json:"discount_type" db:"discount_type"` // "percentage", "fixed"
	DiscountValue float64   `json:"discount_value" db:"discount_value"`
	StartDate     time.Time `json:"start_date" db:"start_date"`
	EndDate       time.Time `json:"end_date" db:"end_date"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	IsDeleted     bool      `json:"is_deleted" db:"is_deleted"`
}

type Voucher struct {
	ID                int        `json:"id" db:"id"`
	Code              string     `json:"code" db:"code"`
	Name              string     `json:"name" db:"name"`
	StartDate         time.Time  `json:"start_date" db:"start_date"`
	EndDate           time.Time  `json:"end_date" db:"end_date"`
	DiscountType      string     `json:"discount_type" db:"discount_type"` // "percentage", "fixed"
	DiscountValue     float64    `json:"discount_value" db:"discount_value"`
	DiscountTarget    string     `json:"discount_target" db:"discount_target"` // "order", "shipping"
	MinOrderValue     float64    `json:"min_order_value" db:"min_order_value"`
	MaxDiscountAmount *float64   `json:"max_discount_amount" db:"max_discount_amount"`
	MaxUsageTotal     *int       `json:"max_usage_total" db:"max_usage_total"`
	MaxUsagePerUser   int        `json:"max_usage_per_user" db:"max_usage_per_user"`
	UsedCount         int        `json:"used_count" db:"used_count"`
	IsDeleted         bool       `json:"is_deleted" db:"is_deleted"`
}

// Request and Response DTOs

type CreatePromotionRequest struct {
	ProductID     string    `json:"product_id" validate:"required"`
	VariantID     *int      `json:"variant_id"`
	Name          string    `json:"name" validate:"required"`
	Description   *string   `json:"description"`
	DiscountType  string    `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue float64   `json:"discount_value" validate:"required,gt=0"`
	StartDate     time.Time `json:"start_date" validate:"required"`
	EndDate       time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	IsActive      bool      `json:"is_active"`
}

type UpdatePromotionRequest struct {
	Name          string    `json:"name" validate:"required"`
	Description   *string   `json:"description"`
	DiscountType  string    `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue float64   `json:"discount_value" validate:"required,gt=0"`
	StartDate     time.Time `json:"start_date" validate:"required"`
	EndDate       time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	IsActive      bool      `json:"is_active"`
}

type CreateVoucherRequest struct {
	Code              string     `json:"code" validate:"required"`
	Name              string     `json:"name" validate:"required"`
	StartDate         time.Time  `json:"start_date" validate:"required"`
	EndDate           time.Time  `json:"end_date" validate:"required,gtfield=StartDate"`
	DiscountType      string     `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue     float64    `json:"discount_value" validate:"required,gt=0"`
	DiscountTarget    string     `json:"discount_target" validate:"required,oneof=order shipping"`
	MinOrderValue     float64    `json:"min_order_value" validate:"gte=0"`
	MaxDiscountAmount *float64   `json:"max_discount_amount" validate:"omitempty,gt=0"`
	MaxUsageTotal     *int       `json:"max_usage_total" validate:"omitempty,gt=0"`
	MaxUsagePerUser   int        `json:"max_usage_per_user" validate:"required,gt=0"`
}

type UpdateVoucherRequest struct {
	Name              string     `json:"name" validate:"required"`
	StartDate         time.Time  `json:"start_date" validate:"required"`
	EndDate           time.Time  `json:"end_date" validate:"required,gtfield=StartDate"`
	DiscountType      string     `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue     float64    `json:"discount_value" validate:"required,gt=0"`
	DiscountTarget    string     `json:"discount_target" validate:"required,oneof=order shipping"`
	MinOrderValue     float64    `json:"min_order_value" validate:"gte=0"`
	MaxDiscountAmount *float64   `json:"max_discount_amount" validate:"omitempty,gt=0"`
	MaxUsageTotal     *int       `json:"max_usage_total" validate:"omitempty,gt=0"`
	MaxUsagePerUser   int        `json:"max_usage_per_user" validate:"required,gt=0"`
}

type ApplyVoucherRequest struct {
	Code        string  `json:"code" validate:"required"`
	OrderAmount float64 `json:"order_amount" validate:"required,gt=0"`
}

type ApplyVoucherResponse struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discount_amount"`
	VoucherID      int     `json:"voucher_id"`
}

// Interfaces

type PromotionVoucherRepository interface {
	CreatePromotion(ctx context.Context, p *Promotion) (*Promotion, error)
	ListPromotions(ctx context.Context) ([]*Promotion, error)
	GetPromotionByID(ctx context.Context, id int) (*Promotion, error)
	UpdatePromotion(ctx context.Context, p *Promotion) (*Promotion, error)
	DeletePromotion(ctx context.Context, id int) error

	CreateVoucher(ctx context.Context, v *Voucher) (*Voucher, error)
	ListVouchers(ctx context.Context) ([]*Voucher, error)
	GetVoucherByID(ctx context.Context, id int) (*Voucher, error)
	GetVoucherByCode(ctx context.Context, code string) (*Voucher, error)
	UpdateVoucher(ctx context.Context, v *Voucher) (*Voucher, error)
	DeleteVoucher(ctx context.Context, id int) error
	ListActiveVouchers(ctx context.Context) ([]*Voucher, error)

	CountUserVoucherUsages(ctx context.Context, voucherID int, userID int) (int, error)
	UseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error
	ReleaseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error

	VerifyProductExists(ctx context.Context, id string) (bool, error)
	VerifyVariantExists(ctx context.Context, productID string, id int) (bool, error)
}

type PromotionVoucherUsecase interface {
	CreatePromotion(ctx context.Context, req *CreatePromotionRequest) (*Promotion, error)
	ListPromotions(ctx context.Context) ([]*Promotion, error)
	GetPromotionByID(ctx context.Context, id int) (*Promotion, error)
	UpdatePromotion(ctx context.Context, id int, req *UpdatePromotionRequest) (*Promotion, error)
	DeletePromotion(ctx context.Context, id int) error

	CreateVoucher(ctx context.Context, req *CreateVoucherRequest) (*Voucher, error)
	ListVouchers(ctx context.Context) ([]*Voucher, error)
	GetVoucherByID(ctx context.Context, id int) (*Voucher, error)
	UpdateVoucher(ctx context.Context, id int, req *UpdateVoucherRequest) (*Voucher, error)
	DeleteVoucher(ctx context.Context, id int) error
	ListActiveVouchers(ctx context.Context) ([]*Voucher, error)

	ApplyVoucher(ctx context.Context, userID int, req *ApplyVoucherRequest) (*ApplyVoucherResponse, error)
	UseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error
	ReleaseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error
}
