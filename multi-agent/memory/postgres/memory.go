package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"multi-agent/memory/types"
	"multi-agent/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

var (
	embeddingDim = 1024
)

// EmbeddingModel is an interface for generating embeddings from text.
type EmbeddingModel interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Dimension() int
}

// PostgresMemoryService implements memory.Service using PostgreSQL with pgvector.
type PostgresMemoryService struct {
	pool           *pgxpool.Pool
	embeddingModel EmbeddingModel
	embeddingDim   int
	topKBM25       int
	topKVector     int
	topKHybrid     int
	weightBM25     float64
	weightVector   float64
}

// PostgresMemoryServiceConfig holds configuration for PostgresMemoryService.
type PostgresMemoryServiceConfig struct {
	// ConnString is the PostgreSQL connection string
	// e.g., "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	ConnString string
	// EmbeddingModel is used to generate embeddings for semantic search (optional)
	EmbeddingModel EmbeddingModel
	topKBM25       int
	topKVector     int
	topKHybrid     int
	weightBM25     float64
	weightVector   float64
}

func GetConfigPGMem() PostgresMemoryServiceConfig {
	return PostgresMemoryServiceConfig{
		ConnString: utils.GetEnvString("POSTGRES_URL", "postgres://jiyuu_user:jiyuu_password@localhost:5433/ecommerce_db"),
		EmbeddingModel: NewOpenAICompatibleEmbedding(
			OpenAICompatibleEmbeddingConfig{
				// BaseURL:   utils.GetEnvString("BASE_URL_OLLAMA", "http://localhost:11434/v1"),
				// Model:     utils.GetEnvString("EMBEDDING_MODEL", "qwen3-embedding:0.6b"),
				BaseURL:   utils.GetEnvString("OPENAI_BASE_URL", "https://api.openai.com/v1"),
				Model:     utils.GetEnvString("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large"),
				Dimension: utils.GetEnvInt("OPENAI_EMBEDDING_DIM", embeddingDim),
				APIKey:    utils.GetEnvString("OPENAI_API_KEY", ""),
			},
		),
		topKBM25:     50,
		topKVector:   50,
		topKHybrid:   5,
		weightBM25:   0.5,
		weightVector: 0.5,
	}
}

func NewPostgresMemoryService(ctx context.Context, cfg PostgresMemoryServiceConfig) (*PostgresMemoryService, error) {
	// 1. parse config
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection string: %w", err)
	}

	// 2. Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %w", err)
	}
	// 3. Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Failed ping to database %w", err)
	}
	// 4. Init embedding dimension
	embeddingDim := 0
	if cfg.EmbeddingModel != nil {
		embeddingDim = cfg.EmbeddingModel.Dimension()
		// If dimension is not preset, probe the model to auto-detect it
		if embeddingDim == 0 {
			embedding, err := cfg.EmbeddingModel.Embed(ctx, "dimension probe")
			if err != nil {
				return nil, fmt.Errorf("failed to probe embedding dimension: %w", err)
			}
			embeddingDim = len(embedding)
		}
	}

	// 5. Return DB{pool}
	return &PostgresMemoryService{
		pool:           pool,
		embeddingDim:   embeddingDim,
		embeddingModel: cfg.EmbeddingModel,
		topKBM25:       cfg.topKBM25,
		topKVector:     cfg.topKVector,
		topKHybrid:     cfg.topKHybrid,
		weightBM25:     cfg.weightBM25,
		weightVector:   cfg.weightVector,
	}, nil
}

// AddSessionToMemory extracts memory entries from a session and stores them.
func (s *PostgresMemoryService) AddSessionToMemory(ctx context.Context, sess session.Session) error {
	events := sess.Events()
	if events == nil || events.Len() == 0 {
		log.Printf("No events found in session to add memory %s", sess.ID())
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var query string

	if s.embeddingModel != nil {
		query = `
			INSERT INTO memory_entries (app_name, user_id, session_id, event_id, author, content, content_text, embedding, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (app_name, user_id, session_id, event_id) DO UPDATE 
			SET content = EXCLUDED.content, content_text = EXCLUDED.content_text, embedding = EXCLUDED.embedding
		`
	} else {
		query = `
			INSERT INTO memory_entries (app_name, user_id, session_id, event_id, author, content, content_text, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (app_name, user_id, session_id, event_id) DO UPDATE 
			SET content = EXCLUDED.content, content_text = EXCLUDED.content_text
		`
	}

	for event := range events.All() {
		if event.Content == nil || len(event.Content.Parts) == 0 {
			continue
		}
		text := extractTextFromContent(event.Content)
		if text == "" {
			continue
		}

		contenJson, err := json.Marshal(event.Content)
		if err != nil {
			continue
		}

		timestamp := event.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		eventID := event.ID
		if eventID == "" {
			eventID = fmt.Sprintf("%s-%d", event.InvocationID, timestamp.UnixNano())
		}

		var embedding interface{}
		if s.embeddingModel != nil {
			embededContent, err := s.embeddingModel.Embed(ctx, text)
			if err != nil {
				log.Printf("Failed to generate embedding: %v", err)
				continue
			}
			embedding = pgvector.NewVector(embededContent)

			_, err = tx.Exec(
				ctx,
				query,
				sess.AppName(),
				sess.UserID(),
				sess.ID(),
				eventID,
				event.Author,
				contenJson,
				text,
				embedding,
				timestamp,
			)
			if err != nil {
				log.Printf("Failed to insert memory entry: %v", err)
				return fmt.Errorf("Failed to insert memory entry: %v", err)
			}
		} else {
			_, err = tx.Exec(
				ctx,
				query,
				sess.AppName(),
				sess.UserID(),
				sess.ID(),
				eventID,
				event.Author,
				contenJson,
				text,
				timestamp,
			)
			if err != nil {
				log.Printf("Failed to insert memory entry: %v", err)
				return fmt.Errorf("Failed to insert memory entry: %v", err)
			}
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresMemoryService) SearchMemory(ctx context.Context, req *memory.SearchRequest) (*memory.SearchResponse, error) {
	if s.embeddingModel != nil {
		embedding, err := s.embeddingModel.Embed(ctx, req.Query)
		if err != nil {
			log.Printf("Failed to generate embedding: %v", err)
			// Nếu embed lỗi, fallback sang text search
			memories, err := s.SearchByText(ctx, req, s.topKBM25)
			if err != nil {
				return nil, err
			}
			return &memory.SearchResponse{Memories: memories}, nil
		}
		memories, err := s.HybridSearch(ctx, req, embedding, s.topKBM25, s.topKVector, s.topKHybrid, s.weightBM25, s.weightVector)
		if err != nil {
			return nil, fmt.Errorf("failed to search memory: %w", err)
		}
		return &memory.SearchResponse{Memories: memories}, nil
	}

	// Nếu không có model embedding, mặc định dùng text search (BM25)
	memories, err := s.SearchByText(ctx, req, s.topKBM25)
	if err != nil {
		return nil, fmt.Errorf("failed to search memory by text: %w", err)
	}

	return &memory.SearchResponse{Memories: memories}, nil
}

// searchByText performs full-text search using PostgreSQL's tsvector.
func (s *PostgresMemoryService) SearchByText(ctx context.Context, req *memory.SearchRequest, topk int) ([]memory.Entry, error) {
	query := `
		SELECT content, author, timestamp 
		FROM memory_entries 
		WHERE app_name = $1 AND user_id = $2
		AND id @@@ paradedb.parse($3)
		ORDER BY paradedb.score(id) DESC, timestamp DESC
		LIMIT $4
	`

	rows, err := s.pool.Query(ctx, query, req.AppName, req.UserID, req.Query, topk)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

// searchByVector performs semantic similarity search.
func (s *PostgresMemoryService) SearchByVector(ctx context.Context, req *memory.SearchRequest, embedding []float32, topK int) ([]memory.Entry, error) {
	query := `
		SELECT 
			content, author, timestamp 
		FROM memory_entries 
		WHERE 
			app_name = $1 AND user_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $3::vector
        LIMIT $4
	`

	rows, err := s.pool.Query(ctx, query, req.AppName, req.UserID, pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

func (s *PostgresMemoryService) SearchWithID(ctx context.Context, req *memory.SearchRequest) ([]types.EntryWithID, error) {
	query := `
		SELECT id, content, author, timestamp 
		FROM memory_entries 
		WHERE app_name = $1 AND user_id = $2
		ORDER BY id DESC
	`

	rows, err := s.pool.Query(ctx, query, req.AppName, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}
	defer rows.Close()

	var entries []types.EntryWithID
	for rows.Next() {
		var entry types.EntryWithID
		if err := rows.Scan(&entry.ID, &entry.Content, &entry.Author, &entry.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan memory entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *PostgresMemoryService) HybridSearch(ctx context.Context, req *memory.SearchRequest, embedding []float32, topKBm25 int, topKVector int, topkFinal int, weightBM25 float64, weightVector float64) ([]memory.Entry, error) {
	query := `
		WITH 
			bm25_results AS (
				SELECT id, ROW_NUMBER() OVER (ORDER BY paradedb.score(id) DESC) AS bm25_rank 
				FROM memory_entries 
				WHERE app_name=$1 AND user_id=$2 
					AND id @@@ paradedb.parse($3) 
				LIMIT $5
			),
			vector_results AS (
				SELECT id, ROW_NUMBER() OVER (ORDER BY embedding <=> $4::vector) AS vector_rank 
				FROM memory_entries 
				WHERE app_name=$1 AND user_id=$2 AND embedding IS NOT NULL 
				ORDER BY embedding <=> $4::vector
				LIMIT $6
			)
		SELECT m.content, m.author, m.timestamp
		FROM memory_entries m
		JOIN (
			SELECT COALESCE(b.id, v.id) AS id,
				(COALESCE(1.0 / (60 + b.bm25_rank), 0) * $7 + 
					COALESCE(1.0 / (60 + v.vector_rank), 0) * $8) AS combined_score
			FROM bm25_results b
			FULL OUTER JOIN vector_results v ON b.id = v.id
		) fused ON m.id = fused.id
		ORDER BY fused.combined_score DESC, m.timestamp DESC
		LIMIT $9
	`

	rows, err := s.pool.Query(ctx, query,
		req.AppName,                   // $1
		req.UserID,                    // $2
		req.Query,                     // $3
		pgvector.NewVector(embedding), // $4
		topKBm25,                      // $5
		topKVector,                    // $6
		weightBM25,                    // $7
		weightVector,                  // $8
		topkFinal,                     // $9
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

func (s *PostgresMemoryService) UpdateMemory(ctx context.Context, appName string, userID string, entryID int, newContent string) error {
	if newContent == "" {
		return fmt.Errorf("content cannot be empty")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	content := &genai.Content{
		Parts: []*genai.Part{{Text: newContent}},
		Role:  "assistant",
	}

	contentJson, err := json.Marshal(content)
	if err != nil {
		log.Printf("failed to marshal content: %v\n", err)
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	if s.embeddingModel != nil {
		var embedding interface{}
		embededContent, err := s.embeddingModel.Embed(ctx, newContent)
		if err != nil {
			log.Printf("failed to generate embedding: %v\n", err)
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		embedding = pgvector.NewVector(embededContent)

		query := `
			UPDATE memory_entries
			SET content = $1, content_text = $2, embedding = $3, timestamp = NOW()
			WHERE app_name = $4 AND id = $5 AND user_id = $6
		`
		_, err = tx.Exec(ctx, query, contentJson, newContent, embedding, appName, entryID, userID)
		if err != nil {
			log.Printf("failed to update memory: %v\n", err)
			return fmt.Errorf("failed to update memory: %w", err)
		}
	} else {
		query := `
			UPDATE memory_entries
			SET content = $1, content_text = $2, timestamp = NOW()
			WHERE app_name = $3 AND id = $4 AND user_id = $5
		`
		_, err = tx.Exec(ctx, query, contentJson, newContent, appName, entryID, userID)
		if err != nil {
			log.Printf("failed to update memory: %v\n", err)
			return fmt.Errorf("failed to update memory: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresMemoryService) DeleteMemory(ctx context.Context, appName string, userID string, entryID int) error {
	query := `
		DELETE FROM memory_entries 
		WHERE app_name = $1 AND id = $2 AND user_id = $3
	`
	result, err := s.pool.Exec(ctx, query, appName, entryID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no memory found with id %d for user %s in app %s", entryID, userID, appName)
	}
	return nil
}

func scanMemories(rows pgx.Rows) ([]memory.Entry, error) {
	var entries []memory.Entry
	for rows.Next() {
		var contentJson []byte
		var author string
		var timestamp time.Time

		if err := rows.Scan(&contentJson, &author, &timestamp); err != nil {
			log.Println("failed to scan row for vector search: %w", err)
			continue
		}

		var content genai.Content
		if err := json.Unmarshal(contentJson, &content); err != nil {
			log.Println("failed to unmarshal content: %w", err)
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}
		entry := memory.Entry{
			Author:    author,
			Content:   &content,
			Timestamp: timestamp,
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Close closes the database connection.
func (s *PostgresMemoryService) Close() {
	s.pool.Close()
}

func extractTextFromContent(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var parts []string
	for _, part := range content.Parts {
		parts = append(parts, part.Text)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

var _ memory.Service = (*PostgresMemoryService)(nil)
var _ types.ExtendedMemoryService = (*PostgresMemoryService)(nil)
