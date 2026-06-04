package domain

import (
	"context"
	"time"
)

// WishlistItem maps to database table wishlist_items
type WishlistItem struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	VariantID int       `json:"variant_id" db:"variant_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WishlistItemResponse represents a wishlisted variant with full detail for displaying
type WishlistItemResponse struct {
	ID            int      `json:"id"`
	VariantID     int      `json:"variant_id"`
	VariantName   string   `json:"variant_name"`
	SKU           string   `json:"sku"`
	SellPrice     float64  `json:"sell_price"`
	ComparePrice  *float64 `json:"compare_price"`
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	ImageURL      *string  `json:"image_url"`
	Stock         int      `json:"stock"`
	DiscountPrice *float64 `json:"discount_price"`
	Rating        float64  `json:"rating"`
}

// AddToWishlistRequest is the JSON payload for adding a variant to wishlist
type AddToWishlistRequest struct {
	VariantID int `json:"variant_id" validate:"required"`
}

// WishlistRepository defines the database interactions
type WishlistRepository interface {
	Create(ctx context.Context, userID int, variantID int) (*WishlistItem, error)
	Delete(ctx context.Context, userID int, variantID int) error
	GetByUserID(ctx context.Context, userID int) ([]*WishlistItemResponse, error)
	IsWishlisted(ctx context.Context, userID int, variantID int) (bool, error)
}

// WishlistUsecase defines the business logic methods
type WishlistUsecase interface {
	AddToWishlist(ctx context.Context, userID int, variantID int) (*WishlistItem, error)
	RemoveFromWishlist(ctx context.Context, userID int, variantID int) error
	GetWishlist(ctx context.Context, userID int) ([]*WishlistItemResponse, error)
}
