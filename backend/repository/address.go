package repository

import (
	"context"
	"fmt"

	"backend/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AddressRepository struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) Create(ctx context.Context, a *domain.Address) (*domain.Address, error) {
	if a.IsDefault {
		tx, err := r.db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		// Reset other defaults
		resetQuery := `UPDATE address SET is_default = false WHERE user_id = $1 AND is_deleted = false`
		_, err = tx.Exec(ctx, resetQuery, a.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to reset default addresses: %w", err)
		}

		insertQuery := `
			INSERT INTO address (user_id, full_name, phone, district, province, ward, detail_address, is_default, is_deleted)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false)
			RETURNING id`
		err = tx.QueryRow(ctx, insertQuery, a.UserID, a.FullName, a.Phone, a.District, a.Province, a.Ward, a.DetailAddress, a.IsDefault).Scan(&a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create default address: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return a, nil
	}

	query := `
		INSERT INTO address (user_id, full_name, phone, district, province, ward, detail_address, is_default, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		a.UserID, a.FullName, a.Phone, a.District, a.Province, a.Ward, a.DetailAddress, a.IsDefault,
	).Scan(&a.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return a, nil
}

func (r *AddressRepository) GetByID(ctx context.Context, id int) (*domain.Address, error) {
	a := &domain.Address{}
	query := `
		SELECT id, user_id, full_name, phone, district, province, ward, detail_address, is_default, is_deleted
		FROM address
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.FullName, &a.Phone, &a.District, &a.Province, &a.Ward, &a.DetailAddress, &a.IsDefault, &a.IsDeleted,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	return a, nil
}

func (r *AddressRepository) ListByUserID(ctx context.Context, userID int) ([]*domain.Address, error) {
	query := `
		SELECT id, user_id, full_name, phone, district, province, ward, detail_address, is_default, is_deleted
		FROM address
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY is_default DESC, id DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]*domain.Address, 0)
	for rows.Next() {
		a := &domain.Address{}
		err := rows.Scan(
			&a.ID, &a.UserID, &a.FullName, &a.Phone, &a.District, &a.Province, &a.Ward, &a.DetailAddress, &a.IsDefault, &a.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		addresses = append(addresses, a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return addresses, nil
}

func (r *AddressRepository) Update(ctx context.Context, a *domain.Address) (*domain.Address, error) {
	query := `
		UPDATE address
		SET full_name = $1, phone = $2, district = $3, province = $4, ward = $5, detail_address = $6, is_default = $7
		WHERE id = $8 AND is_deleted = false`

	tag, err := r.db.Exec(ctx, query,
		a.FullName, a.Phone, a.District, a.Province, a.Ward, a.DetailAddress, a.IsDefault, a.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrAddressNotFound
	}

	return a, nil
}

func (r *AddressRepository) SetDefault(ctx context.Context, userID, addressID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Reset default flag
	resetQuery := `UPDATE address SET is_default = false WHERE user_id = $1 AND is_deleted = false`
	_, err = tx.Exec(ctx, resetQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to reset default flags: %w", err)
	}

	// Set default flag
	setQuery := `UPDATE address SET is_default = true WHERE id = $1 AND user_id = $2 AND is_deleted = false`
	tag, err := tx.Exec(ctx, setQuery, addressID, userID)
	if err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrAddressNotFound
	}

	return tx.Commit(ctx)
}

func (r *AddressRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE address SET is_deleted = true, is_default = false WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}
