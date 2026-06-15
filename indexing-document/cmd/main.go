// Binary cmd/main.go khởi động HTTP API server cho indexing-document.
//
// Env vars:
//
//	SERVER_PORT        (default: 8081)
//	DATABASE_URL       (required)
//	EMBEDDING_BASE_URL (default: http://localhost:11434)
//	EMBEDDING_API_KEY  (default: ollama)
//	EMBEDDING_MODEL    (default: nomic-embed-text)
//	EMBEDDING_DIM      (default: 768)
//	CHUNK_SIZE         (default: 512)
//	CHUNK_OVERLAP      (default: 64)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	utils "indexing"
	"indexing/chunker"
	"indexing/db"
	"indexing/embedder"
	"indexing/ingestion"
	"indexing/server"
	// "github.com/rs/zerolog/log"
)

func main() {
	// log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					src := a.Value.Any().(*slog.Source)
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line))
				}
				return a
			},
		}),
	)
	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Config
	port := utils.GetEnvString("SERVER_PORT", "8081")
	dsn := utils.GetEnvString("DATABASE_URL", "postgres://jiyuu_user:jiyuu_password@localhost:5433/ecommerce_db")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	embBaseURL := utils.GetEnvString("EMBEDDING_BASE_URL", "http://localhost:11434")
	embAPIKey := utils.GetEnvString("EMBEDDING_API_KEY", "ollama")
	embModel := utils.GetEnvString("EMBEDDING_MODEL", "qwen3-embedding:0.6b")
	embDim := utils.GetEnvInt("EMBEDDING_DIM", 1024)
	chunkSize := utils.GetEnvInt("CHUNK_SIZE", 512)
	chunkOverlap := utils.GetEnvInt("CHUNK_OVERLAP", 64)

	// Store
	store, err := db.NewStore(ctx, dsn, embDim)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer store.Close()
	log.Info("database connected")

	// Embedder
	emb := embedder.New(embBaseURL, embAPIKey, embModel)
	if err := emb.Ping(ctx); err != nil {
		return fmt.Errorf("embedder ping: %w", err)
	}
	log.Info("embedder ready", "model", embModel)

	// Sync manager
	chk := chunker.New(chunkSize, chunkOverlap)
	sync := ingestion.NewSyncManager(store, emb, chk, log)

	// HTTP server
	h := server.NewHandler(store, sync, emb, log)
	engine := server.NewRouter(h)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case <-ctx.Done():
		log.Info("shutting down…")
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	}
}
