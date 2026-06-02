package repository

import (
	"context"
	"errors"
	"fmt"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BannerRepository struct {
	db *pgxpool.Pool
}

func NewBannerRepository(db *pgxpool.Pool) *BannerRepository {
	return &BannerRepository{db: db}
}

func (r *BannerRepository) Create(ctx context.Context, b *domain.Banner) (*domain.Banner, error) {
	query := `
		INSERT INTO banners (title, subtitle, description, image_url, tag, link_url, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, b.Title, b.Subtitle, b.Description, b.ImageURL, b.Tag, b.LinkURL, b.SortOrder, b.IsActive).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create banner: %w", err)
	}
	return b, nil
}

func (r *BannerRepository) GetByID(ctx context.Context, id int) (*domain.Banner, error) {
	b := &domain.Banner{}
	query := `SELECT id, title, subtitle, description, image_url, tag, link_url, sort_order, is_active, created_at, updated_at FROM banners WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.Title, &b.Subtitle, &b.Description, &b.ImageURL, &b.Tag, &b.LinkURL, &b.SortOrder, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("banner not found")
		}
		return nil, fmt.Errorf("failed to get banner by ID: %w", err)
	}
	return b, nil
}

func (r *BannerRepository) List(ctx context.Context, onlyActive bool) ([]*domain.Banner, error) {
	query := `SELECT id, title, subtitle, description, image_url, tag, link_url, sort_order, is_active, created_at, updated_at FROM banners`
	if onlyActive {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY sort_order ASC, id DESC"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query banners: %w", err)
	}
	defer rows.Close()

	banners := make([]*domain.Banner, 0)
	for rows.Next() {
		b := &domain.Banner{}
		err := rows.Scan(&b.ID, &b.Title, &b.Subtitle, &b.Description, &b.ImageURL, &b.Tag, &b.LinkURL, &b.SortOrder, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan banner: %w", err)
		}
		banners = append(banners, b)
	}
	return banners, nil
}

func (r *BannerRepository) Update(ctx context.Context, b *domain.Banner) (*domain.Banner, error) {
	query := `
		UPDATE banners
		SET title = $1, subtitle = $2, description = $3, image_url = $4, tag = $5, link_url = $6, sort_order = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query, b.Title, b.Subtitle, b.Description, b.ImageURL, b.Tag, b.LinkURL, b.SortOrder, b.IsActive, b.ID).Scan(&b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update banner: %w", err)
	}
	return b, nil
}

func (r *BannerRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM banners WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete banner: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("banner not found")
	}
	return nil
}
