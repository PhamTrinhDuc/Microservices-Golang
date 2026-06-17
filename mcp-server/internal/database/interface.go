package database

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for database operations
// This interface enables testing with mocks
type Store interface {
	// InsertDocument(ctx context.Context, doc *PolicyChunk) error
	// InsertDocuments(ctx context.Context, docs []*PolicyChunk) error
	SearchDocuments(ctx context.Context, query string, limit int) ([]*PolicyChunk, error)
	GetDocument(ctx context.Context, docID uuid.UUID) (*PolicyChunk, error)
	ListDocuments(ctx context.Context, limit int, offset int) ([]*PolicyChunk, error)
	UpdateDocument(ctx context.Context, doc *PolicyChunk) error
	DeleteDocumentByID(ctx context.Context, docID uuid.UUID) error
}
