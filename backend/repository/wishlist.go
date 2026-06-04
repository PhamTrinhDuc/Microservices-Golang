package repository

import (
	"context"
	"fmt"

	"backend/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WishlistRepository struct {
	db *pgxpool.Pool
}

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{db: db}
}

func (r *WishlistRepository) Create(ctx context.Context, userID int, variantID int) (*domain.WishlistItem, error) {
	item := &domain.WishlistItem{
		UserID:    userID,
		VariantID: variantID,
	}
	query := `
		INSERT INTO wishlist_items (user_id, variant_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, variant_id) DO UPDATE SET created_at = NOW()
		RETURNING id, created_at`
	
	err := r.db.QueryRow(ctx, query, userID, variantID).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wishlist item: %w", err)
	}
	return item, nil
}

func (r *WishlistRepository) Delete(ctx context.Context, userID int, variantID int) error {
	query := `DELETE FROM wishlist_items WHERE user_id = $1 AND variant_id = $2`
	_, err := r.db.Exec(ctx, query, userID, variantID)
	if err != nil {
		return fmt.Errorf("failed to delete wishlist item: %w", err)
	}
	return nil
}

func (r *WishlistRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.WishlistItemResponse, error) {
	query := `
		SELECT 
			w.id, w.variant_id, pv.name as variant_name, pv.sku, pv.sell_price, pv.compare_price,
			p.id as product_id, p.name as product_name, img.url as image_url,
			COALESCE(inv.total_qty, 0) as stock,
			COALESCE(avg_rev.rating, 0.0) as rating
		FROM wishlist_items w
		JOIN product_variant pv ON w.variant_id = pv.id
		JOIN product p ON pv.product_id = p.id
		LEFT JOIN LATERAL (
			SELECT url FROM product_image
			WHERE product_id = p.id AND (variant_id = pv.id OR variant_id IS NULL)
			ORDER BY is_thumbnail DESC, sort_order ASC, id ASC
			LIMIT 1
		) img ON true
		LEFT JOIN (
			SELECT variant_id, SUM(quantity - reserved) as total_qty
			FROM product_inventory
			GROUP BY variant_id
		) inv ON pv.id = inv.variant_id
		LEFT JOIN (
			SELECT product_id, AVG(rating) as rating
			FROM reviews
			GROUP BY product_id
		) avg_rev ON p.id = avg_rev.product_id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist items: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.WishlistItemResponse, 0)
	for rows.Next() {
		item := &domain.WishlistItemResponse{}
		err := rows.Scan(
			&item.ID,
			&item.VariantID,
			&item.VariantName,
			&item.SKU,
			&item.SellPrice,
			&item.ComparePrice,
			&item.ProductID,
			&item.ProductName,
			&item.ImageURL,
			&item.Stock,
			&item.Rating,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wishlist item row: %w", err)
		}
		
		// Set discount price if it has a discount
		if item.ComparePrice != nil && *item.ComparePrice > item.SellPrice {
			item.DiscountPrice = &item.SellPrice
		}
		
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("wishlist items rows error: %w", err)
	}

	return items, nil
}

func (r *WishlistRepository) IsWishlisted(ctx context.Context, userID int, variantID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE user_id = $1 AND variant_id = $2)`
	err := r.db.QueryRow(ctx, query, userID, variantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check wishlist item existence: %w", err)
	}
	return exists, nil
}
