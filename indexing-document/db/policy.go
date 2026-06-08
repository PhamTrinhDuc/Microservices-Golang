package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Policy represents a row in the policies table.
type Policy struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicyList wraps a page of policies with total count for pagination.
type PolicyList struct {
	Policies []*Policy `json:"policies"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// CreatePolicy inserts a new policy into the database.
func (s *Store) CreatePolicy(ctx context.Context, title, slug, content, category string, isActive bool) (*Policy, error) {
	p := &Policy{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO policies (title, slug, content, category, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, slug, content, category, is_active, created_at, updated_at
	`, title, slug, content, category, isActive).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating policy: %w", err)
	}
	return p, nil
}

// UpdatePolicy updates an existing policy in the database.
func (s *Store) UpdatePolicy(ctx context.Context, id uuid.UUID, title, slug, content, category string, isActive bool) (*Policy, error) {
	p := &Policy{}
	err := s.pool.QueryRow(ctx, `
		UPDATE policies
		SET title = $1, slug = $2, content = $3, category = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, title, slug, content, category, is_active, created_at, updated_at
	`, title, slug, content, category, isActive, id).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("updating policy: %w", err)
	}
	return p, nil
}

// GetPolicy retrieves a single policy by its ID.
func (s *Store) GetPolicy(ctx context.Context, id uuid.UUID) (*Policy, error) {
	p := &Policy{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at
		FROM policies
		WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("getting policy %s: %w", id, err)
	}
	return p, nil
}

// GetPolicyBySlug retrieves a single policy by its unique slug.
func (s *Store) GetPolicyBySlug(ctx context.Context, slug string) (*Policy, error) {
	p := &Policy{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at
		FROM policies
		WHERE slug = $1
	`, slug).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting policy by slug %s: %w", slug, err)
	}
	return p, nil
}

// ListPolicies returns a paginated page of policies, optionally filtered by category.
func (s *Store) ListPolicies(ctx context.Context, category string, limit, offset int) (*PolicyList, error) {
	if limit <= 0 {
		limit = 10
	}

	countQuery := "SELECT COUNT(*) FROM policies"
	countArgs := []any{}
	if category != "" {
		countQuery += " WHERE category = $1"
		countArgs = append(countArgs, category)
	}

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting policies: %w", err)
	}

	dataQuery := `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at
		FROM policies
	`
	dataArgs := []any{}
	if category != "" {
		dataQuery += " WHERE category = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		dataArgs = append(dataArgs, category, limit, offset)
	} else {
		dataQuery += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		dataArgs = append(dataArgs, limit, offset)
	}

	rows, err := s.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("listing policies: %w", err)
	}
	defer rows.Close()

	policies := make([]*Policy, 0)
	for rows.Next() {
		p := &Policy{}
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning policy row: %w", err)
		}
		policies = append(policies, p)
	}

	return &PolicyList{
		Policies: policies,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// DeletePolicy deletes a policy from the database. All dependent chunks are
// deleted automatically by PostgreSQL due to the ON DELETE CASCADE constraint.
func (s *Store) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM policies WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting policy %s: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy %s not found", id)
	}
	return nil
}
