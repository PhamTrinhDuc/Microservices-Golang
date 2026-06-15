// https://www.tigerdata.com/blog/you-dont-need-elasticsearch-bm25-is-now-in-postgres
package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// VectorSearchResult holds params for vector search
type VectorSearchResult struct {
	Document PolicyChunk
	Score    float64
}

// FullTextSearchResult holds params for full text search
type FullTextSearchResult struct {
	Document  PolicyChunk
	BM25Score float64
}

// HybridSearchParams holds parameters for hybrid search
type HybridSearchParams struct {
	Query        string
	BranchID     uuid.UUID
	Embedding    []float32
	Limit        int
	BM25Weight   float64 // Weight for lexical search (0.0 to 1.0)
	VectorWeight float64 // Weight for semantic search (0.0 to 1.0)
	MinBM25Score float64 // Minimum BM25 score threshold
	MinVectorSim float64 // Minimum vector similarity threshold
}

// HybridSearchResult represents a result from hybrid search
type HybridSearchResult struct {
	Document      PolicyChunk
	BM25Score     float64
	VectorScore   float64
	CombinedScore float64
}

// VectorSearch performs similarity search using pgvector
func (db *DB) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*VectorSearchResult, error) {
	queryString := `
		SELECT id, branch_id, title, content, metadata, embedding, created_at, updated_at,
		1 - (embedding <=>$1::vector) AS similarity_score
		FROM knowledge_base 
		WHERE embedding IS NOT NULL AND is_active = TRUE
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx, queryString, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector search: %w", err)
	}
	defer rows.Close()

	var results []*VectorSearchResult
	for rows.Next() {
		doc := &PolicyChunk{}
		var score float64
		var dbEmbedding pgvector.Vector

		err := rows.Scan(
			&doc.ID,
			&doc.PolicyId,
			&doc.Chunkindex,
			&doc.Content,
			&doc.Embedding,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&score,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan documents: %w", err)
		}
		doc.Embedding = dbEmbedding.Slice()
		results = append(results, &VectorSearchResult{
			Document: *doc,
			Score:    score,
		})
	}
	return results, nil
}

// FullTextSearch performs full-text search using pg_textsearch (ParadEDB)
func (db *DB) FullTextSearch(ctx context.Context, query string, limit int) ([]*FullTextSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	queryString := `
		SELECT
			id, branch_id, title, content, metadata, created_at, updated_at,
			paradedb.score(id) AS score
		FROM knowledge_base
		WHERE is_active = TRUE
		  AND id @@@ paradedb.parse($1)
		ORDER BY score DESC
		LIMIT $2
	`

	rows, err := db.pool.Query(ctx, queryString, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform full text search: %w", err)
	}
	defer rows.Close()

	var results []*FullTextSearchResult
	for rows.Next() {
		doc := &PolicyChunk{}
		var score float64

		err := rows.Scan(
			&doc.ID,
			&doc.PolicyId,
			&doc.Chunkindex,
			&doc.Content,
			&doc.Embedding,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fulltext search result: %w", err)
		}

		results = append(results, &FullTextSearchResult{
			Document:  *doc,
			BM25Score: score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// HybridSearch implements a Reciprocal Rank Fusion (RRF) approach for combining results
func (db *DB) HybridSearch(ctx context.Context, params HybridSearchParams) ([]HybridSearchResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Default weights nếu không được cung cấp
	if params.BM25Weight == 0 && params.VectorWeight == 0 {
		params.BM25Weight = 1.0
		params.VectorWeight = 1.0
	}

	query := `
		WITH 
		-- 1. BM25 Search using pg_textsearch
		bm25_results AS (
			SELECT 
				id, branch_id, title, content, metadata, embedding, created_at, updated_at,
				ROW_NUMBER() OVER (ORDER BY paradedb.score(id) DESC) AS bm25_rank
			FROM knowledge_base
			WHERE is_active = TRUE
			  	AND (branch_id = $2 OR branch_id IS NULL OR $2 IS NULL)
			   	AND id @@@ paradedb.parse($1)
			ORDER BY paradedb.score(id) DESC
			LIMIT 50
		),

		-- 2. Vector Search
		vector_results AS (
			SELECT 
				id, branch_id, title, content, metadata, embedding, created_at, updated_at,
				ROW_NUMBER() OVER (ORDER BY embedding <=> $3) AS vector_rank
			FROM knowledge_base
			WHERE is_active = TRUE 
			  AND embedding IS NOT NULL
			  AND (branch_id = $2 OR branch_id IS NULL OR $2 IS NULL)
			ORDER BY embedding <=> $3
			LIMIT 50
		),

		-- 3. Fusion với Reciprocal Rank Fusion
		fused AS (
			SELECT 
				COALESCE(b.id, v.id) AS id,
				COALESCE(b.branch_id, v.branch_id) AS branch_id,
				COALESCE(b.title, v.title) AS title,
				COALESCE(b.content, v.content) AS content,
				COALESCE(b.metadata, v.metadata) AS metadata,
				COALESCE(b.embedding, v.embedding) AS embedding,
				COALESCE(b.created_at, v.created_at) AS created_at,
				COALESCE(b.updated_at, v.updated_at) AS updated_at,
				
				COALESCE(1.0 / (60 + b.bm25_rank), 0) * $4 +
				COALESCE(1.0 / (60 + v.vector_rank), 0) * $5 AS combined_score,

				COALESCE(b.bm25_rank, 999) AS bm25_rank,
				COALESCE(v.vector_rank, 999) AS vector_rank
			FROM bm25_results b
			FULL OUTER JOIN vector_results v ON b.id = v.id
		)

		SELECT 
			id, branch_id, title, content, metadata, embedding,
			created_at, updated_at,
			(1.0 / (60 + bm25_rank)) AS bm25_score,     -- approximate score
			(1.0 / (60 + vector_rank)) AS vector_score,
			combined_score
		FROM fused
		ORDER BY combined_score DESC
		LIMIT $6
	`

	var embedding interface{}
	if params.Embedding != nil {
		embedding = pgvector.NewVector(params.Embedding)
	} else {
		embedding = nil
	}

	rows, err := db.pool.Query(ctx, query,
		params.Query,        // $1: text query cho BM25
		params.BranchID,     // $2: branch filter (UUID hoặc nil)
		embedding,           // $3: embedding
		params.BM25Weight,   // $4: trọng số BM25
		params.VectorWeight, // $5: trọng số Vector
		params.Limit,        // $6
	)
	if err != nil {
		return nil, fmt.Errorf("failed to perform hybrid search: %w", err)
	}
	defer rows.Close()

	var results []HybridSearchResult
	for rows.Next() {
		var doc PolicyChunk
		var bm25Score, vectorScore, combinedScore float64
		var dbEmbedding *pgvector.Vector

		err := rows.Scan(
			&doc.ID,
			&doc.PolicyId,
			&doc.Chunkindex,
			&doc.Content,
			&dbEmbedding,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&bm25Score,
			&vectorScore,
			&combinedScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hybrid search result: %w", err)
		}

		if dbEmbedding != nil {
			doc.Embedding = dbEmbedding.Slice()
		}

		results = append(results, HybridSearchResult{
			Document:      doc,
			BM25Score:     bm25Score,
			VectorScore:   vectorScore,
			CombinedScore: combinedScore,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(results) <= 0 {
		fmt.Println("WARNING: Not found documents!")
	}

	return results, nil
}
