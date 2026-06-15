package database

import (
	"context"
	"fmt"
	"mcp-server/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32 // tối đa X connect 1 lúc
	MinConns int32 // tối thiếu X connect 1 lúc
}

type PolicyChunk struct {
	ID         uuid.UUID `json:"id"`
	PolicyId   string    `json:"pylicy_id"`
	Chunkindex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"embedding,omitempty"`
	CreatedAt  time.Time `json:"creatd_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DB struct {
	pool *pgxpool.Pool
}

func (cfg *DBConfig) Validate() error {
	// 1. Kiểm tra các trường BẮT BUỘC phải có
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		return fmt.Errorf("host, user, and dbname are required")
	}

	// 2. Gán giá trị MẶC ĐỊNH cho các trường tùy chọn nếu chúng trống
	if cfg.Port <= 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.MaxConns <= 0 {
		cfg.MaxConns = 10
	}
	if cfg.MinConns < 0 {
		cfg.MinConns = 2
	}
	return nil
}

func NewDB(ctx context.Context, cfg DBConfig) (*DB, error) {
	// 0.Validate config
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate fields in config: %w", err)
	}

	// 1. Concate connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.MaxConns, cfg.MinConns,
	)

	// 2. parse config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection string: %w", err)
	}

	// 3. Config connection pool
	poolConfig.MaxConns = cfg.MaxConns             // override connString
	poolConfig.MinConns = cfg.MinConns             // override connString
	poolConfig.MaxConnLifetime = time.Hour         // thời gian sống tối đa của 1 kết nối
	poolConfig.MaxConnIdleTime = 30 * time.Minute  // thời gian tối đa của 1 kết nối không hoạt động
	poolConfig.HealthCheckPeriod = 1 * time.Minute // thời gian kiểm tra kết nối

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return nil
	}
	// 4. Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %w", err)
	}
	// 5. Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Failed ping to database %w", err)
	}
	// 6. Return DB{pool}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

// BeginTx starts a new transaction with tenant context
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// SearchDocuments searches documents using text search
func (db *DB) SearchDocuments(ctx context.Context, query string, limit int) ([]*PolicyChunk, error) {
	searchQuery := `
		SELECT id, policy_id, chunk_index, content, embedding, created_at, updated_at
		FROM policy_chunk 
		WHERE is_active = TRUE AND 
			to_tsvector('simple', f_unaccent(content)) @@ websearch_to_tsquery('simple', unaccent($1))
		ORDER BY ts_rank_cd(to_tsvector('simple', f_unaccent(content)), websearch_to_tsquery('simple', unaccent($1))) DESC
		LIMIT $2
	`
	rows, err := db.pool.Query(ctx, searchQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	var results []*PolicyChunk
	for rows.Next() {
		doc := &PolicyChunk{}
		err := rows.Scan(
			&doc.ID,
			&doc.PolicyId,
			&doc.Chunkindex,
			&doc.Content,
			&doc.Embedding,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		results = append(results, doc)
	}
	return results, nil
}

func (db *DB) GetDocument(ctx context.Context, docID uuid.UUID) (*PolicyChunk, error) {
	query := `
		SELECT id, policy_id, chunk_index, content, embedding, created_at, updated_at
		FROM policy_chunk
		WHERE id = $1 AND is_active = TRUE
	`
	doc := &PolicyChunk{}
	var dbEmbedding *pgvector.Vector
	err := db.pool.QueryRow(ctx, query, docID).Scan(
		&doc.ID, &doc.PolicyId, &doc.Chunkindex, &doc.Content, &doc.Embedding, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if dbEmbedding != nil {
		doc.Embedding = dbEmbedding.Slice()
	}

	return doc, nil
}

// ListDocuments lists all documents
func (db *DB) ListDocuments(ctx context.Context, limit int, offset int) ([]*PolicyChunk, error) {
	query := `
		SELECT id, policy_id, chunk_index, content, embedding, created_at, updated_at
		FROM policy_chunk 
		WHERE is_active = TRUE
		ORDER BY created_at DESC 
		LIMIT (CASE WHEN $1 > 0 THEN $1 ELSE NULL END)
		OFFSET (CASE WHEN $2 > 0 THEN $2 ELSE 0 END)
	`

	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var results []*PolicyChunk
	for rows.Next() {
		doc := &PolicyChunk{}
		err := rows.Scan(
			&doc.ID,
			&doc.PolicyId,
			&doc.Chunkindex,
			&doc.Content,
			&doc.Embedding,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to Scan documents: %w", err)
		}
		results = append(results, doc)
	}
	return results, nil
}

// UpdateDocument updates an existing document in knowledge_base
func (db *DB) UpdateDocument(ctx context.Context, doc *PolicyChunk) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE policy_chunk
		SET policy_id = $1, chunk_index = $2, content = $3, embedding = $4
		WHERE id = $5
		RETURNING updated_at
	`
	var embedding pgvector.Vector
	if doc.Embedding != nil {
		embedding = pgvector.NewVector(doc.Embedding)
	}

	err = tx.QueryRow(ctx, query, doc.PolicyId, doc.Chunkindex, doc.Content, embedding, doc.ID).Scan(&doc.UpdatedAt)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("document not found")
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// DeleteDocumentByID deletes a document by ID
func (db *DB) DeleteDocumentByID(ctx context.Context, docID uuid.UUID) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `DELETE FROM policy_chunk WHERE id = $1`
	result, err := tx.Exec(ctx, query, docID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}
	return tx.Commit(ctx)
}

func NewDBConfig() DBConfig {
	return DBConfig{
		Host:     utils.GetEnvString("DB_HOST", "localhost"),
		Port:     utils.GetEnvInt("DB_PORT", 5433),
		User:     utils.GetEnvString("DB_USER", "mcp_user"),
		Password: utils.GetEnvString("DB_PASSWORD", "mcp_password"),
		DBName:   utils.GetEnvString("DB_NAME", "salon_chain"),
		SSLMode:  utils.GetEnvString("DB_SSLMODE", "disable"),
		MaxConns: int32(utils.GetEnvInt("DB_MAX_CONNS", 10)),
		MinConns: int32(utils.GetEnvInt("DB_MIN_CONNS", 2)),
	}
}
