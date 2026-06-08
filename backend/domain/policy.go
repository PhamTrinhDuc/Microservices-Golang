package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Policy struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Slug      string    `json:"slug" db:"slug"`
	Content   string    `json:"content" db:"content"`
	Category  string    `json:"category" db:"category"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePolicyRequest struct {
	Title    string `json:"title" validate:"required"`
	Slug     string `json:"slug" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Category string `json:"category" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type UpdatePolicyRequest struct {
	Title    string `json:"title" validate:"required"`
	Slug     string `json:"slug" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Category string `json:"category" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type PolicySearchResult struct {
	ChunkID        uuid.UUID `json:"chunk_id"`
	PolicyID       uuid.UUID `json:"policy_id"`
	PolicyTitle    string    `json:"policy_title"`
	PolicySlug     string    `json:"policy_slug"`
	PolicyCategory string    `json:"policy_category"`
	ChunkIndex     int       `json:"chunk_index"`
	Content        string    `json:"content"`
	Score          float64   `json:"score"`
}

type ChatQueryRequest struct {
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit"`
}

type PolicyRepository interface {
	Create(ctx context.Context, p *Policy) (*Policy, error)
	Update(ctx context.Context, p *Policy) (*Policy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	GetBySlug(ctx context.Context, slug string) (*Policy, error)
	List(ctx context.Context, category string, limit, offset int) ([]*Policy, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SearchChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]*PolicySearchResult, error)
}

type PolicyUsecase interface {
	Create(ctx context.Context, req *CreatePolicyRequest) (*Policy, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*Policy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	GetBySlug(ctx context.Context, slug string) (*Policy, error)
	List(ctx context.Context, category string, limit, offset int) ([]*Policy, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, query string, limit int) ([]*PolicySearchResult, error)
}
