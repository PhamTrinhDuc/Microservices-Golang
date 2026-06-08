package usecase

import (
	"context"
	"fmt"
	"log"

	"backend/domain"

	"github.com/google/uuid"
	"indexing/embedder"
	"indexing/ingestion"
)

type PolicyUsecase struct {
	repo        domain.PolicyRepository
	embedder    *embedder.Embedder
	syncManager *ingestion.SyncManager
}

func NewPolicyUsecase(repo domain.PolicyRepository, emb *embedder.Embedder, sync *ingestion.SyncManager) *PolicyUsecase {
	return &PolicyUsecase{
		repo:        repo,
		embedder:    emb,
		syncManager: sync,
	}
}

func (u *PolicyUsecase) Create(ctx context.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error) {
	p := &domain.Policy{
		Title:    req.Title,
		Slug:     req.Slug,
		Content:  req.Content,
		Category: req.Category,
		IsActive: *req.IsActive,
	}

	created, err := u.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Synchronize chunks and embeddings in database
	err = u.syncManager.SyncPolicyChunks(ctx, created.ID, created.Content)
	if err != nil {
		// Log the error but don't fail the policy creation since the database record is saved successfully.
		// Admin can trigger a retry by editing/updating the policy.
		log.Printf("Warning: failed to synchronize policy chunks to RAG for ID %s: %v", created.ID, err)
	}

	return created, nil
}

func (u *PolicyUsecase) Update(ctx context.Context, id uuid.UUID, req *domain.UpdatePolicyRequest) (*domain.Policy, error) {
	p, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Title = req.Title
	p.Slug = req.Slug
	p.Content = req.Content
	p.Category = req.Category
	p.IsActive = *req.IsActive

	updated, err := u.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	// Update chunks and embeddings in database
	err = u.syncManager.SyncPolicyChunks(ctx, updated.ID, updated.Content)
	if err != nil {
		log.Printf("Warning: failed to update policy chunks to RAG for ID %s: %v", updated.ID, err)
	}

	return updated, nil
}

func (u *PolicyUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *PolicyUsecase) GetBySlug(ctx context.Context, slug string) (*domain.Policy, error) {
	return u.repo.GetBySlug(ctx, slug)
}

func (u *PolicyUsecase) List(ctx context.Context, category string, limit, offset int) ([]*domain.Policy, int, error) {
	return u.repo.List(ctx, category, limit, offset)
}

func (u *PolicyUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *PolicyUsecase) Search(ctx context.Context, query string, limit int) ([]*domain.PolicySearchResult, error) {
	if limit <= 0 {
		limit = 3
	}

	// Convert textual search query into vector representation
	queryEmbedding, err := u.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed search query: %w", err)
	}

	// Search closest vector chunks in postgres
	return u.repo.SearchChunks(ctx, queryEmbedding, limit)
}
