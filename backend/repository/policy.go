package repository

import (
	"context"
	"errors"
	"fmt"

	"backend/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type PolicyRepository struct {
	db *pgxpool.Pool
}

func NewPolicyRepository(db *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func (r *PolicyRepository) Create(ctx context.Context, p *domain.Policy) (*domain.Policy, error) {
	query := `
		INSERT INTO policies (title, slug, content, category, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, p.Title, p.Slug, p.Content, p.Category, p.IsActive).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}
	return p, nil
}

func (r *PolicyRepository) Update(ctx context.Context, p *domain.Policy) (*domain.Policy, error) {
	query := `
		UPDATE policies
		SET title = $1, slug = $2, content = $3, category = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query, p.Title, p.Slug, p.Content, p.Category, p.IsActive, p.ID).
		Scan(&p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}
	return p, nil
}

func (r *PolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	p := &domain.Policy{}
	query := `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at 
		FROM policies WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy by ID: %w", err)
	}
	return p, nil
}

func (r *PolicyRepository) GetBySlug(ctx context.Context, slug string) (*domain.Policy, error) {
	p := &domain.Policy{}
	query := `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at 
		FROM policies WHERE slug = $1`

	err := r.db.QueryRow(ctx, query, slug).Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy by slug: %w", err)
	}
	return p, nil
}

func (r *PolicyRepository) List(ctx context.Context, category string, limit, offset int) ([]*domain.Policy, int, error) {
	countQuery := "SELECT COUNT(*) FROM policies WHERE 1=1"
	countArgs := []interface{}{}
	argIdx := 1

	if category != "" {
		countQuery += fmt.Sprintf(" AND category = $%d", argIdx)
		countArgs = append(countArgs, category)
		argIdx++
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count policies: %w", err)
	}

	query := `
		SELECT id, title, slug, content, category, is_active, created_at, updated_at 
		FROM policies WHERE 1=1`
	args := []interface{}{}
	argIdx = 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	policies := make([]*domain.Policy, 0)
	for rows.Next() {
		p := &domain.Policy{}
		err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, p)
	}

	return policies, total, nil
}

func (r *PolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM policies WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("policy not found")
	}
	return nil
}

func (r *PolicyRepository) SearchChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]*domain.PolicySearchResult, error) {
	vec := pgvector.NewVector(queryEmbedding)

	query := `
		SELECT
			c.id,
			c.policy_id,
			p.title,
			p.slug,
			p.category,
			c.chunk_index,
			c.content,
			1 - (c.embedding <=> $1) AS score
		FROM policy_chunks c
		JOIN policies p ON p.id = c.policy_id
		WHERE c.embedding IS NOT NULL AND p.is_active = true
		ORDER BY c.embedding <=> $1
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query semantic policy chunks: %w", err)
	}
	defer rows.Close()

	results := make([]*domain.PolicySearchResult, 0)
	for rows.Next() {
		res := &domain.PolicySearchResult{}
		err := rows.Scan(
			&res.ChunkID, &res.PolicyID, &res.PolicyTitle, &res.PolicySlug,
			&res.PolicyCategory, &res.ChunkIndex, &res.Content, &res.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during search results iteration: %w", err)
	}
	return results, nil
}
