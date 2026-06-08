package ingestion

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"indexing/chunker"
	"indexing/db"
	"indexing/embedder"
)

// SyncManager orchestrates splitting policy content into chunks,
// generating embeddings, and storing them in the database.
type SyncManager struct {
	store    *db.Store
	embedder *embedder.Embedder
	chunker  *chunker.Chunker
	log      *slog.Logger
}

// NewSyncManager creates a new SyncManager.
func NewSyncManager(store *db.Store, emb *embedder.Embedder, chk *chunker.Chunker, log *slog.Logger) *SyncManager {
	return &SyncManager{
		store:    store,
		embedder: emb,
		chunker:  chk,
		log:      log,
	}
}

// SyncPolicyChunks splits policy content, embeds each segment, and stores the chunks.
// It deletes any previously inserted chunks for this policy before inserting new ones.
func (sm *SyncManager) SyncPolicyChunks(ctx context.Context, policyID uuid.UUID, content string) error {
	chunks := sm.chunker.Split(content)
	if len(chunks) == 0 {
		// If content is empty/blank, clean up any existing chunks.
		if err := sm.store.DeleteChunksByPolicy(ctx, policyID); err != nil {
			return fmt.Errorf("deleting existing chunks for empty policy content: %w", err)
		}
		return nil
	}

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	embeddings, err := sm.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("generating embeddings for policy chunks: %w", err)
	}

	// Delete old chunks first (to prevent duplicate chunks on update)
	if err := sm.store.DeleteChunksByPolicy(ctx, policyID); err != nil {
		return fmt.Errorf("deleting previous chunks for policy %s: %w", policyID, err)
	}

	dbChunks := make([]db.PolicyChunk, len(chunks))
	for i, c := range chunks {
		dbChunks[i] = db.PolicyChunk{
			PolicyID:   policyID,
			ChunkIndex: c.Index,
			Content:    c.Content,
		}
	}

	if err := sm.store.InsertPolicyChunks(ctx, policyID, dbChunks, embeddings); err != nil {
		return fmt.Errorf("storing policy chunks: %w", err)
	}

	sm.log.Info("policy chunks synchronized successfully",
		"policy_id", policyID,
		"chunk_count", len(dbChunks),
	)

	return nil
}
