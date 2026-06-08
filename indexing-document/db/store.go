// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package db manages all interactions with PostgreSQL + pgvector.
// It handles connection pooling, schema migrations on startup,
// and provides typed methods for documents and chunks.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgxpool.Pool and exposes all database operations.
// Use NewStore to create an instance.
type Store struct {
	pool       *pgxpool.Pool
	dimensions int
}

// NewStore creates a connection pool using the given DSN and runs
// all pending migrations. dimensions is the vector size produced by the
// configured embedding model (e.g. 768 for nomic-embed-text).
// Returns an error if the connection or migrations fail.
func NewStore(ctx context.Context, dsn string, dimensions int) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	s := &Store{pool: pool, dimensions: dimensions}
	if err := s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

// Close releases all connections in the pool. Should be deferred after NewStore.
func (s *Store) Close() {
	s.pool.Close()
}

// migrate applies the schema idempotently. It enables pgvector, creates the
// policies and policy_chunks tables if they don't exist, and adds the vector index.
// Safe to run on every startup.
func (s *Store) migrate(ctx context.Context) error {
	// Guard: if the policy_chunks table already exists and the embedding column has
	// explicit dimensions, verify they match the configured value.
	var colType string
	err := s.pool.QueryRow(ctx, `
		SELECT pg_catalog.format_type(a.atttypid, a.atttypmod)
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = 'policy_chunks'
		  AND n.nspname = 'public'
		  AND a.attname = 'embedding'
		  AND a.attnum > 0
		  AND NOT a.attisdropped
	`).Scan(&colType)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Table does not exist yet — nothing to validate.
	case err != nil:
		return fmt.Errorf("checking existing embedding dimensions: %w", err)
	default:
		// colType is e.g. "vector(768)" or "vector" (no dimensions).
		// Only validate when dimensions are explicitly set.
		var current int
		if n, _ := fmt.Sscanf(colType, "vector(%d)", &current); n == 1 && current != s.dimensions {
			return fmt.Errorf(
				"embedding dimension mismatch: database has vector(%d) but config says %d — "+
					"change embeddings.dimensions back to %d or drop the policy_chunks table and re-ingest",
				current, s.dimensions, current,
			)
		}
	}

	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,

		`CREATE TABLE IF NOT EXISTS policies (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title       TEXT        NOT NULL,
			slug        TEXT        NOT NULL UNIQUE,
			content     TEXT        NOT NULL,
			category    TEXT        NOT NULL,
			is_active   BOOLEAN     NOT NULL DEFAULT true,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS policy_chunks (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			policy_id     UUID        NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
			chunk_index   INTEGER     NOT NULL,
			content       TEXT        NOT NULL,
			embedding     vector(%d),
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, s.dimensions),

		// Index for semantic search using HNSW
		`CREATE INDEX IF NOT EXISTS policy_chunks_embedding_idx
			ON policy_chunks USING hnsw (embedding vector_cosine_ops)`,
	}

	for _, stmt := range statements {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("executing migration statement: %w\nSQL: %s", err, stmt)
		}
	}

	return nil
}