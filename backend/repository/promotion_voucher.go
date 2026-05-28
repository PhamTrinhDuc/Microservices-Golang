package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PromotionVoucherRepository struct {
	db *pgxpool.Pool
}

func NewPromotionVoucherRepository(db *pgxpool.Pool) *PromotionVoucherRepository {
	return &PromotionVoucherRepository{db: db}
}

// --- Promotions CRUD ---

func (r *PromotionVoucherRepository) CreatePromotion(ctx context.Context, p *domain.Promotion) (*domain.Promotion, error) {
	query := `
		INSERT INTO promotions (product_id, variant_id, name, description, discount_type, discount_value, start_date, end_date, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, false)
		RETURNING id`
	err := r.db.QueryRow(ctx, query,
		p.ProductID,
		p.VariantID,
		p.Name,
		p.Description,
		p.DiscountType,
		p.DiscountValue,
		p.StartDate,
		p.EndDate,
	).Scan(&p.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}
	p.IsActive = true
	p.IsDeleted = false
	return p, nil
}

func (r *PromotionVoucherRepository) ListPromotions(ctx context.Context) ([]*domain.Promotion, error) {
	query := `
		SELECT id, product_id, variant_id, name, description, discount_type, discount_value, start_date, end_date, is_active, is_deleted
		FROM promotions
		WHERE is_deleted = false
		ORDER BY id DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list promotions: %w", err)
	}
	defer rows.Close()

	var promotions []*domain.Promotion
	for rows.Next() {
		p := &domain.Promotion{}
		err := rows.Scan(
			&p.ID,
			&p.ProductID,
			&p.VariantID,
			&p.Name,
			&p.Description,
			&p.DiscountType,
			&p.DiscountValue,
			&p.StartDate,
			&p.EndDate,
			&p.IsActive,
			&p.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan promotion: %w", err)
		}
		promotions = append(promotions, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("promotions rows error: %w", err)
	}
	return promotions, nil
}

func (r *PromotionVoucherRepository) GetPromotionByID(ctx context.Context, id int) (*domain.Promotion, error) {
	query := `
		SELECT id, product_id, variant_id, name, description, discount_type, discount_value, start_date, end_date, is_active, is_deleted
		FROM promotions
		WHERE id = $1 AND is_deleted = false`
	p := &domain.Promotion{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.ProductID,
		&p.VariantID,
		&p.Name,
		&p.Description,
		&p.DiscountType,
		&p.DiscountValue,
		&p.StartDate,
		&p.EndDate,
		&p.IsActive,
		&p.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPromotionNotFound
		}
		return nil, fmt.Errorf("failed to get promotion by id: %w", err)
	}
	return p, nil
}

func (r *PromotionVoucherRepository) UpdatePromotion(ctx context.Context, p *domain.Promotion) (*domain.Promotion, error) {
	query := `
		UPDATE promotions
		SET name = $1, description = $2, discount_type = $3, discount_value = $4, start_date = $5, end_date = $6, is_active = $7
		WHERE id = $8 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query,
		p.Name,
		p.Description,
		p.DiscountType,
		p.DiscountValue,
		p.StartDate,
		p.EndDate,
		p.IsActive,
		p.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update promotion: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrPromotionNotFound
	}
	return p, nil
}

func (r *PromotionVoucherRepository) DeletePromotion(ctx context.Context, id int) error {
	query := `UPDATE promotions SET is_deleted = true WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft-delete promotion: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPromotionNotFound
	}
	return nil
}

// --- Vouchers CRUD ---

func (r *PromotionVoucherRepository) CreateVoucher(ctx context.Context, v *domain.Voucher) (*domain.Voucher, error) {
	query := `
		INSERT INTO vouchers (code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 0, false)
		RETURNING id`
	err := r.db.QueryRow(ctx, query,
		v.Code,
		v.Name,
		v.StartDate,
		v.EndDate,
		v.DiscountType,
		v.DiscountValue,
		v.DiscountTarget,
		v.MinOrderValue,
		v.MaxDiscountAmount,
		v.MaxUsageTotal,
		v.MaxUsagePerUser,
	).Scan(&v.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create voucher: %w", err)
	}
	v.UsedCount = 0
	v.IsDeleted = false
	return v, nil
}

func (r *PromotionVoucherRepository) ListVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	query := `
		SELECT id, code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE is_deleted = false
		ORDER BY id DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list vouchers: %w", err)
	}
	defer rows.Close()

	var vouchers []*domain.Voucher
	for rows.Next() {
		v := &domain.Voucher{}
		err := rows.Scan(
			&v.ID,
			&v.Code,
			&v.Name,
			&v.StartDate,
			&v.EndDate,
			&v.DiscountType,
			&v.DiscountValue,
			&v.DiscountTarget,
			&v.MinOrderValue,
			&v.MaxDiscountAmount,
			&v.MaxUsageTotal,
			&v.MaxUsagePerUser,
			&v.UsedCount,
			&v.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan voucher: %w", err)
		}
		vouchers = append(vouchers, v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("vouchers rows error: %w", err)
	}
	return vouchers, nil
}

func (r *PromotionVoucherRepository) GetVoucherByID(ctx context.Context, id int) (*domain.Voucher, error) {
	query := `
		SELECT id, code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE id = $1 AND is_deleted = false`
	v := &domain.Voucher{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.Code,
		&v.Name,
		&v.StartDate,
		&v.EndDate,
		&v.DiscountType,
		&v.DiscountValue,
		&v.DiscountTarget,
		&v.MinOrderValue,
		&v.MaxDiscountAmount,
		&v.MaxUsageTotal,
		&v.MaxUsagePerUser,
		&v.UsedCount,
		&v.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, fmt.Errorf("failed to get voucher by id: %w", err)
	}
	return v, nil
}

func (r *PromotionVoucherRepository) GetVoucherByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	query := `
		SELECT id, code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE code = $1 AND is_deleted = false`
	v := &domain.Voucher{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&v.ID,
		&v.Code,
		&v.Name,
		&v.StartDate,
		&v.EndDate,
		&v.DiscountType,
		&v.DiscountValue,
		&v.DiscountTarget,
		&v.MinOrderValue,
		&v.MaxDiscountAmount,
		&v.MaxUsageTotal,
		&v.MaxUsagePerUser,
		&v.UsedCount,
		&v.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, fmt.Errorf("failed to get voucher by code: %w", err)
	}
	return v, nil
}

func (r *PromotionVoucherRepository) UpdateVoucher(ctx context.Context, v *domain.Voucher) (*domain.Voucher, error) {
	query := `
		UPDATE vouchers
		SET name = $1, start_date = $2, end_date = $3, discount_type = $4, discount_value = $5, discount_target = $6, min_order_value = $7, max_discount_amount = $8, max_usage_total = $9, max_usage_per_user = $10
		WHERE id = $11 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query,
		v.Name,
		v.StartDate,
		v.EndDate,
		v.DiscountType,
		v.DiscountValue,
		v.DiscountTarget,
		v.MinOrderValue,
		v.MaxDiscountAmount,
		v.MaxUsageTotal,
		v.MaxUsagePerUser,
		v.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update voucher: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrVoucherNotFound
	}
	return v, nil
}

func (r *PromotionVoucherRepository) DeleteVoucher(ctx context.Context, id int) error {
	query := `UPDATE vouchers SET is_deleted = true WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft-delete voucher: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVoucherNotFound
	}
	return nil
}

func (r *PromotionVoucherRepository) ListActiveVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	query := `
		SELECT id, code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE is_deleted = false AND NOW() >= start_date AND NOW() <= end_date
		ORDER BY id DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active vouchers: %w", err)
	}
	defer rows.Close()

	var vouchers []*domain.Voucher
	for rows.Next() {
		v := &domain.Voucher{}
		err := rows.Scan(
			&v.ID,
			&v.Code,
			&v.Name,
			&v.StartDate,
			&v.EndDate,
			&v.DiscountType,
			&v.DiscountValue,
			&v.DiscountTarget,
			&v.MinOrderValue,
			&v.MaxDiscountAmount,
			&v.MaxUsageTotal,
			&v.MaxUsagePerUser,
			&v.UsedCount,
			&v.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active voucher: %w", err)
		}
		vouchers = append(vouchers, v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("active vouchers rows error: %w", err)
	}
	return vouchers, nil
}

func (r *PromotionVoucherRepository) CountUserVoucherUsages(ctx context.Context, voucherID int, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM voucher_usages WHERE voucher_id = $1 AND user_id = $2`
	var count int
	err := r.db.QueryRow(ctx, query, voucherID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user voucher usages: %w", err)
	}
	return count, nil
}

// --- Concurrency Protection / Row Locking Operations ---

func (r *PromotionVoucherRepository) UseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for using voucher: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. SELECT FOR UPDATE to lock the voucher row
	var v domain.Voucher
	querySelect := `
		SELECT id, code, name, start_date, end_date, discount_type, discount_value, discount_target, min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE id = $1 AND is_deleted = false
		FOR UPDATE`
	
	err = tx.QueryRow(ctx, querySelect, voucherID).Scan(
		&v.ID,
		&v.Code,
		&v.Name,
		&v.StartDate,
		&v.EndDate,
		&v.DiscountType,
		&v.DiscountValue,
		&v.DiscountTarget,
		&v.MinOrderValue,
		&v.MaxDiscountAmount,
		&v.MaxUsageTotal,
		&v.MaxUsagePerUser,
		&v.UsedCount,
		&v.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVoucherNotFound
		}
		return fmt.Errorf("failed to select voucher for update: %w", err)
	}

	// Validate timeframe
	now := time.Now()
	if now.Before(v.StartDate) || now.After(v.EndDate) {
		return domain.ErrVoucherExpired
	}

	// Validate global usage limit
	if v.MaxUsageTotal != nil && v.UsedCount >= *v.MaxUsageTotal {
		return domain.ErrVoucherLimitReached
	}

	// Validate user usage limit inside transaction lock
	var userUsages int
	queryUsages := `SELECT COUNT(*) FROM voucher_usages WHERE voucher_id = $1 AND user_id = $2`
	err = tx.QueryRow(ctx, queryUsages, voucherID, userID).Scan(&userUsages)
	if err != nil {
		return fmt.Errorf("failed to check user usages inside tx: %w", err)
	}
	if userUsages >= v.MaxUsagePerUser {
		return domain.ErrVoucherUserLimitReached
	}

	// 2. UPDATE vouchers set used_count = used_count + 1
	queryUpdate := `UPDATE vouchers SET used_count = used_count + 1 WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdate, voucherID)
	if err != nil {
		return fmt.Errorf("failed to update voucher used_count: %w", err)
	}

	// 3. INSERT INTO voucher_usages
	queryInsert := `INSERT INTO voucher_usages (voucher_id, user_id, order_id, used_at) VALUES ($1, $2, $3, NOW())`
	_, err = tx.Exec(ctx, queryInsert, voucherID, userID, orderID)
	if err != nil {
		return fmt.Errorf("failed to insert voucher usage record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit voucher usage transaction: %w", err)
	}

	return nil
}

func (r *PromotionVoucherRepository) ReleaseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for releasing voucher: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if the usage actually exists
	var usageExists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM voucher_usages WHERE voucher_id = $1 AND user_id = $2 AND order_id = $3)`
	err = tx.QueryRow(ctx, queryCheck, voucherID, userID, orderID).Scan(&usageExists)
	if err != nil {
		return fmt.Errorf("failed to check if voucher usage exists: %w", err)
	}

	if usageExists {
		// Delete the usage record
		queryDelete := `DELETE FROM voucher_usages WHERE voucher_id = $1 AND user_id = $2 AND order_id = $3`
		_, err = tx.Exec(ctx, queryDelete, voucherID, userID, orderID)
		if err != nil {
			return fmt.Errorf("failed to delete voucher usage record: %w", err)
		}

		// Decrement the used_count, using FOR UPDATE lock on the voucher row
		var usedCount int
		querySelect := `SELECT used_count FROM vouchers WHERE id = $1 FOR UPDATE`
		err = tx.QueryRow(ctx, querySelect, voucherID).Scan(&usedCount)
		if err != nil {
			return fmt.Errorf("failed to lock voucher row for decrement: %w", err)
		}

		if usedCount > 0 {
			queryUpdate := `UPDATE vouchers SET used_count = used_count - 1 WHERE id = $1`
			_, err = tx.Exec(ctx, queryUpdate, voucherID)
			if err != nil {
				return fmt.Errorf("failed to decrement voucher used_count: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit voucher release transaction: %w", err)
	}

	return nil
}

func (r *PromotionVoucherRepository) VerifyProductExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product WHERE id = $1 AND is_deleted = false)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *PromotionVoucherRepository) VerifyVariantExists(ctx context.Context, productID string, id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product_variant WHERE id = $1 AND product_id = $2 AND is_deleted = false)`
	err := r.db.QueryRow(ctx, query, id, productID).Scan(&exists)
	return exists, err
}
