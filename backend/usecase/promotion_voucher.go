package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/domain"
)

type PromotionVoucherUsecase struct {
	repo domain.PromotionVoucherRepository
}

func NewPromotionVoucherUsecase(repo domain.PromotionVoucherRepository) *PromotionVoucherUsecase {
	return &PromotionVoucherUsecase{repo: repo}
}

// --- Promotions CRUD Usecases ---

func (uc *PromotionVoucherUsecase) CreatePromotion(ctx context.Context, req *domain.CreatePromotionRequest) (*domain.Promotion, error) {
	if req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("start date must be before or equal to end date")
	}

	// 1. Verify product exists
	exists, err := uc.repo.VerifyProductExists(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrProductNotFound
	}

	// 2. Verify variant exists if provided
	if req.VariantID != nil {
		exists, err := uc.repo.VerifyVariantExists(ctx, req.ProductID, *req.VariantID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrVariantNotFound
		}
	}

	promo := &domain.Promotion{
		ProductID:     req.ProductID,
		VariantID:     req.VariantID,
		Name:          req.Name,
		Description:   req.Description,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		IsActive:      req.IsActive,
	}

	return uc.repo.CreatePromotion(ctx, promo)
}

func (uc *PromotionVoucherUsecase) ListPromotions(ctx context.Context) ([]*domain.Promotion, error) {
	return uc.repo.ListPromotions(ctx)
}

func (uc *PromotionVoucherUsecase) GetPromotionByID(ctx context.Context, id int) (*domain.Promotion, error) {
	return uc.repo.GetPromotionByID(ctx, id)
}

func (uc *PromotionVoucherUsecase) UpdatePromotion(ctx context.Context, id int, req *domain.UpdatePromotionRequest) (*domain.Promotion, error) {
	if req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("start date must be before or equal to end date")
	}

	promo, err := uc.repo.GetPromotionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	promo.Name = req.Name
	promo.Description = req.Description
	promo.DiscountType = req.DiscountType
	promo.DiscountValue = req.DiscountValue
	promo.StartDate = req.StartDate
	promo.EndDate = req.EndDate
	promo.IsActive = req.IsActive

	return uc.repo.UpdatePromotion(ctx, promo)
}

func (uc *PromotionVoucherUsecase) DeletePromotion(ctx context.Context, id int) error {
	return uc.repo.DeletePromotion(ctx, id)
}

// --- Vouchers CRUD Usecases ---

func (uc *PromotionVoucherUsecase) CreateVoucher(ctx context.Context, req *domain.CreateVoucherRequest) (*domain.Voucher, error) {
	if req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("start date must be before or equal to end date")
	}

	// Check if code is already taken
	existing, err := uc.repo.GetVoucherByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("voucher code already exists")
	}

	v := &domain.Voucher{
		Code:              req.Code,
		Name:              req.Name,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		DiscountTarget:    req.DiscountTarget,
		MinOrderValue:     req.MinOrderValue,
		MaxDiscountAmount: req.MaxDiscountAmount,
		MaxUsageTotal:     req.MaxUsageTotal,
		MaxUsagePerUser:   req.MaxUsagePerUser,
	}

	return uc.repo.CreateVoucher(ctx, v)
}

func (uc *PromotionVoucherUsecase) ListVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	return uc.repo.ListVouchers(ctx)
}

func (uc *PromotionVoucherUsecase) GetVoucherByID(ctx context.Context, id int) (*domain.Voucher, error) {
	return uc.repo.GetVoucherByID(ctx, id)
}

func (uc *PromotionVoucherUsecase) UpdateVoucher(ctx context.Context, id int, req *domain.UpdateVoucherRequest) (*domain.Voucher, error) {
	if req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("start date must be before or equal to end date")
	}

	v, err := uc.repo.GetVoucherByID(ctx, id)
	if err != nil {
		return nil, err
	}

	v.Name = req.Name
	v.StartDate = req.StartDate
	v.EndDate = req.EndDate
	v.DiscountType = req.DiscountType
	v.DiscountValue = req.DiscountValue
	v.DiscountTarget = req.DiscountTarget
	v.MinOrderValue = req.MinOrderValue
	v.MaxDiscountAmount = req.MaxDiscountAmount
	v.MaxUsageTotal = req.MaxUsageTotal
	v.MaxUsagePerUser = req.MaxUsagePerUser

	return uc.repo.UpdateVoucher(ctx, v)
}

func (uc *PromotionVoucherUsecase) DeleteVoucher(ctx context.Context, id int) error {
	return uc.repo.DeleteVoucher(ctx, id)
}

func (uc *PromotionVoucherUsecase) ListActiveVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	return uc.repo.ListActiveVouchers(ctx)
}

// --- Voucher Apply & Usage Logic ---

func (uc *PromotionVoucherUsecase) ApplyVoucher(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
	v, err := uc.repo.GetVoucherByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	// 1. Timeframe check
	now := time.Now()
	if now.Before(v.StartDate) || now.After(v.EndDate) {
		return nil, domain.ErrVoucherExpired
	}

	// 2. Global usage limit check
	if v.MaxUsageTotal != nil && v.UsedCount >= *v.MaxUsageTotal {
		return nil, domain.ErrVoucherLimitReached
	}

	// 3. User usage limit check
	userUsages, err := uc.repo.CountUserVoucherUsages(ctx, v.ID, userID)
	if err != nil {
		return nil, err
	}
	if userUsages >= v.MaxUsagePerUser {
		return nil, domain.ErrVoucherUserLimitReached
	}

	// 4. Minimum order value check
	if req.OrderAmount < v.MinOrderValue {
		return nil, domain.ErrVoucherMinAmountNotMet
	}

	// 5. Calculate discount amount
	var discount float64
	if v.DiscountType == "percentage" {
		discount = req.OrderAmount * (v.DiscountValue / 100.0)
		if v.MaxDiscountAmount != nil && discount > *v.MaxDiscountAmount {
			discount = *v.MaxDiscountAmount
		}
	} else if v.DiscountType == "fixed" {
		discount = v.DiscountValue
		if discount > req.OrderAmount {
			discount = req.OrderAmount // cannot exceed order amount
		}
	}

	return &domain.ApplyVoucherResponse{
		Valid:          true,
		DiscountAmount: discount,
		VoucherID:      v.ID,
	}, nil
}

func (uc *PromotionVoucherUsecase) UseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	return uc.repo.UseVoucher(ctx, userID, voucherID, orderID)
}

func (uc *PromotionVoucherUsecase) ReleaseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	return uc.repo.ReleaseVoucher(ctx, userID, voucherID, orderID)
}
