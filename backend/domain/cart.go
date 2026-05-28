package domain

import (
	"context"
	"time"
)

// Models

type CartItem struct {
	ID        int       `json:"id" db:"id"`
	UserID    *int      `json:"user_id" db:"user_id"`
	SessionID *string   `json:"session_id" db:"session_id"`
	VariantID int       `json:"variant_id" db:"variant_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CartItemResponse struct {
	ID          int      `json:"id"`
	VariantID   int      `json:"variant_id"`
	VariantName string   `json:"variant_name"`
	SKU         string   `json:"sku"`
	Price       float64  `json:"price"`
	PriceBase   *float64 `json:"price_base"`
	ProductID   string   `json:"product_id"`
	ProductName string   `json:"product_name"`
	ImageURL    *string  `json:"image_url"`
	Quantity    int      `json:"quantity"`
}

// Request DTOs

type AddToCartRequest struct {
	VariantID int     `json:"variant_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	SessionID *string `json:"session_id"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type MergeCartRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

// Interfaces

type CartRepository interface {
	ListCartItems(ctx context.Context, userID *int, sessionID *string) ([]*CartItem, error)
	FindCartItem(ctx context.Context, userID *int, sessionID *string, variantID int) (*CartItem, error)
	GetCartItemByID(ctx context.Context, id int) (*CartItem, error)
	CreateCartItem(ctx context.Context, item *CartItem) (*CartItem, error)
	UpdateCartItemQuantity(ctx context.Context, id int, quantity int) (*CartItem, error)
	LinkGuestItemToUser(ctx context.Context, id int, userID int) error
	DeleteCartItem(ctx context.Context, id int) error
	ClearCart(ctx context.Context, userID *int, sessionID *string) error
	GetCartDetails(ctx context.Context, userID *int, sessionID *string) ([]*CartItemResponse, error)
	VerifyVariantExists(ctx context.Context, id int) (bool, error)
}

type CartUsecase interface {
	GetCart(ctx context.Context, userID *int, sessionID *string) ([]*CartItemResponse, error)
	AddToCart(ctx context.Context, userID *int, sessionID *string, req *AddToCartRequest) (*CartItem, error)
	UpdateItemQuantity(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*CartItem, error)
	RemoveItem(ctx context.Context, userID *int, sessionID *string, itemID int) error
	ClearCart(ctx context.Context, userID *int, sessionID *string) error
	MergeCart(ctx context.Context, userID int, sessionID string) error
}
