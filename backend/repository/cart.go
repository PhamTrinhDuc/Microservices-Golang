package repository

import (
	"context"
	"fmt"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) ListCartItems(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItem, error) {
	var query string
	var arg interface{}

	if userID != nil {
		query = `SELECT id, user_id, session_id, variant_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = $1 ORDER BY id ASC`
		arg = *userID
	} else if sessionID != nil {
		query = `SELECT id, user_id, session_id, variant_id, quantity, created_at, updated_at FROM cart_items WHERE session_id = $1 ORDER BY id ASC`
		arg = *sessionID
	} else {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}

	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.CartItem, 0)
	for rows.Next() {
		item := &domain.CartItem{}
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.SessionID,
			&item.VariantID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cart items rows error: %w", err)
	}

	return items, nil
}

func (r *CartRepository) FindCartItem(ctx context.Context, userID *int, sessionID *string, variantID int) (*domain.CartItem, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `SELECT id, user_id, session_id, variant_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = $1 AND variant_id = $2`
		args = []interface{}{*userID, variantID}
	} else if sessionID != nil {
		query = `SELECT id, user_id, session_id, variant_id, quantity, created_at, updated_at FROM cart_items WHERE session_id = $1 AND variant_id = $2`
		args = []interface{}{*sessionID, variantID}
	} else {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}

	item := &domain.CartItem{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&item.ID,
		&item.UserID,
		&item.SessionID,
		&item.VariantID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Return nil, nil when not found to simplify check logic in usecase
		}
		return nil, fmt.Errorf("failed to find cart item: %w", err)
	}

	return item, nil
}

func (r *CartRepository) GetCartItemByID(ctx context.Context, id int) (*domain.CartItem, error) {
	item := &domain.CartItem{}
	query := `SELECT id, user_id, session_id, variant_id, quantity, created_at, updated_at FROM cart_items WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.UserID,
		&item.SessionID,
		&item.VariantID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to get cart item by ID: %w", err)
	}

	return item, nil
}

func (r *CartRepository) CreateCartItem(ctx context.Context, item *domain.CartItem) (*domain.CartItem, error) {
	query := `
		INSERT INTO cart_items (user_id, session_id, variant_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		item.UserID,
		item.SessionID,
		item.VariantID,
		item.Quantity,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create cart item: %w", err)
	}

	return item, nil
}

func (r *CartRepository) UpdateCartItemQuantity(ctx context.Context, id int, quantity int) (*domain.CartItem, error) {
	item := &domain.CartItem{}
	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, session_id, variant_id, quantity, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, quantity, id).Scan(
		&item.ID,
		&item.UserID,
		&item.SessionID,
		&item.VariantID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	return item, nil
}

func (r *CartRepository) LinkGuestItemToUser(ctx context.Context, id int, userID int) error {
	query := `UPDATE cart_items SET user_id = $1, session_id = NULL, updated_at = NOW() WHERE id = $2`

	tag, err := r.db.Exec(ctx, query, userID, id)
	if err != nil {
		return fmt.Errorf("failed to link guest cart item to user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *CartRepository) DeleteCartItem(ctx context.Context, id int) error {
	query := `DELETE FROM cart_items WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, userID *int, sessionID *string) error {
	var query string
	var arg interface{}

	if userID != nil {
		query = `DELETE FROM cart_items WHERE user_id = $1`
		arg = *userID
	} else if sessionID != nil {
		query = `DELETE FROM cart_items WHERE session_id = $1`
		arg = *sessionID
	} else {
		return fmt.Errorf("either user_id or session_id must be provided")
	}

	_, err := r.db.Exec(ctx, query, arg)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

func (r *CartRepository) GetCartDetails(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error) {
	var query string
	var arg interface{}

	if userID != nil {
		query = `
			SELECT ci.id, ci.variant_id, pv.name as variant_name, pv.sku, pv.price, pv.price_base,
			       p.id as product_id, p.name as product_name, img.url as image_url, ci.quantity
			FROM cart_items ci
			JOIN product_variant pv ON ci.variant_id = pv.id
			JOIN product p ON pv.product_id = p.id
			LEFT JOIN LATERAL (
				SELECT url FROM product_image
				WHERE product_id = p.id AND (variant_id = pv.id OR variant_id IS NULL)
				ORDER BY is_thumbnail DESC, sort_order ASC, id ASC
				LIMIT 1
			) img ON true
			WHERE ci.user_id = $1
			ORDER BY ci.id ASC`
		arg = *userID
	} else if sessionID != nil {
		query = `
			SELECT ci.id, ci.variant_id, pv.name as variant_name, pv.sku, pv.price, pv.price_base,
			       p.id as product_id, p.name as product_name, img.url as image_url, ci.quantity
			FROM cart_items ci
			JOIN product_variant pv ON ci.variant_id = pv.id
			JOIN product p ON pv.product_id = p.id
			LEFT JOIN LATERAL (
				SELECT url FROM product_image
				WHERE product_id = p.id AND (variant_id = pv.id OR variant_id IS NULL)
				ORDER BY is_thumbnail DESC, sort_order ASC, id ASC
				LIMIT 1
			) img ON true
			WHERE ci.session_id = $1
			ORDER BY ci.id ASC`
		arg = *sessionID
	} else {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}

	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart details: %w", err)
	}
	defer rows.Close()

	details := make([]*domain.CartItemResponse, 0)
	for rows.Next() {
		d := &domain.CartItemResponse{}
		err := rows.Scan(
			&d.ID,
			&d.VariantID,
			&d.VariantName,
			&d.SKU,
			&d.Price,
			&d.PriceBase,
			&d.ProductID,
			&d.ProductName,
			&d.ImageURL,
			&d.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart details row: %w", err)
		}
		details = append(details, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cart details rows error: %w", err)
	}

	return details, nil
}

func (r *CartRepository) VerifyVariantExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM product_variant pv
			JOIN product p ON pv.product_id = p.id
			WHERE pv.id = $1 AND pv.is_active = true AND pv.is_deleted = false
			  AND p.is_active = true AND p.is_deleted = false
		)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}
