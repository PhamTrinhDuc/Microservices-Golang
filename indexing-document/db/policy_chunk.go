package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// PolicyChunk represents a text segment extracted from a policy, with its embedding.
type PolicyChunk struct {
	ID         uuid.UUID `json:"id"`
	PolicyID   uuid.UUID `json:"policy_id"`
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
}

// PolicySearchResult is a chunk augmented with similarity score and parent policy metadata
// returned by semantic search queries.
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

// InsertPolicyChunks inserts a batch of chunks with their embeddings into the database.
// All chunks belong to the same policy. The operation is performed in a single transaction for atomicity.
func (s *Store) InsertPolicyChunks(ctx context.Context, policyID uuid.UUID, chunks []PolicyChunk, embeddings [][]float32) error {
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks (%d) and embeddings (%d) length mismatch", len(chunks), len(embeddings))
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for i, chunk := range chunks {
		vec := pgvector.NewVector(embeddings[i])
		_, err := tx.Exec(ctx, `
			INSERT INTO policy_chunks (policy_id, chunk_index, content, embedding)
			VALUES ($1, $2, $3, $4)
		`, policyID, chunk.ChunkIndex, chunk.Content, vec)
		if err != nil {
			return fmt.Errorf("inserting policy chunk %d: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing chunk transaction: %w", err)
	}
	return nil
}

// DeleteChunksByPolicy removes all chunks belonging to a policy.
// Used before re-inserting on retry to prevent duplicates.
func (s *Store) DeleteChunksByPolicy(ctx context.Context, policyID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM policy_chunks WHERE policy_id = $1`, policyID)
	if err != nil {
		return fmt.Errorf("deleting chunks for policy %s: %w", policyID, err)
	}
	return nil
}

// SearchPolicies performs cosine similarity search over chunk embeddings.
// Returns the top `limit` most similar chunks joined with policy metadata.
func (s *Store) SearchPolicies(ctx context.Context, queryEmbedding []float32, limit int) ([]*PolicySearchResult, error) {
	vec := pgvector.NewVector(queryEmbedding)

	rows, err := s.pool.Query(ctx, `
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
		LIMIT $2
	`, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("executing search query: %w", err)
	}
	defer rows.Close()

	var results []*PolicySearchResult
	for rows.Next() {
		r := &PolicySearchResult{}
		if err := rows.Scan(
			&r.ChunkID, &r.PolicyID, &r.PolicyTitle, &r.PolicySlug,
			&r.PolicyCategory, &r.ChunkIndex, &r.Content, &r.Score,
		); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}
	return results, nil
}
