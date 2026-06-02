package domain

import (
	"context"
	"time"
)

type Banner struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Subtitle    string    `json:"subtitle" db:"subtitle"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Tag         string    `json:"tag" db:"tag"`
	LinkURL     string    `json:"link_url" db:"link_url"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateBannerRequest struct {
	Title       string `json:"title" validate:"required"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url" validate:"required,url"`
	Tag         string `json:"tag"`
	LinkURL     string `json:"link_url"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

type UpdateBannerRequest struct {
	Title       string `json:"title" validate:"required"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url" validate:"required,url"`
	Tag         string `json:"tag"`
	LinkURL     string `json:"link_url"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

type BannerRepository interface {
	Create(ctx context.Context, b *Banner) (*Banner, error)
	GetByID(ctx context.Context, id int) (*Banner, error)
	List(ctx context.Context, onlyActive bool) ([]*Banner, error)
	Update(ctx context.Context, b *Banner) (*Banner, error)
	Delete(ctx context.Context, id int) error
}

type BannerUsecase interface {
	Create(ctx context.Context, req *CreateBannerRequest) (*Banner, error)
	GetByID(ctx context.Context, id int) (*Banner, error)
	List(ctx context.Context, onlyActive bool) ([]*Banner, error)
	Update(ctx context.Context, id int, req *UpdateBannerRequest) (*Banner, error)
	Delete(ctx context.Context, id int) error
}
