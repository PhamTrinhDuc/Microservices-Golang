// Package server exposes the HTTP API for the indexing-document service using Gin.
package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"indexing/db"
	"indexing/embedder"
	"indexing/ingestion"
)

// Handler holds all dependencies.
type Handler struct {
	store    *db.Store
	sync     *ingestion.SyncManager
	embedder *embedder.Embedder
	log      *slog.Logger
}

func NewHandler(store *db.Store, sync *ingestion.SyncManager, emb *embedder.Embedder, log *slog.Logger) *Handler {
	return &Handler{store: store, sync: sync, embedder: emb, log: log}
}

// NewRouter registers all policy routes and returns the Gin engine.
//
// POST   /policies              → create
// GET    /policies              → list  (?category=&page=&limit=)
// GET    /policies/:id          → get by UUID
// GET    /policies/slug/:slug   → get by slug
// PUT    /policies/:id          → update
// DELETE /policies/:id          → delete
// POST   /policies/:id/sync     → re-embed chunks
// POST   /policies/search       → semantic search
func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	p := r.Group("/policies")
	{
		p.POST("", h.createPolicy)
		p.GET("", h.listPolicies)
		p.POST("/search", h.searchPolicies)
		p.GET("/slug/:slug", h.getPolicyBySlug)
		p.GET("/:id", h.getPolicy)
		p.PUT("/:id", h.updatePolicy)
		p.DELETE("/:id", h.deletePolicy)
		p.POST("/:id/sync", h.syncPolicy)
	}

	return r
}

// ─── Request types ────────────────────────────────────────────────────────────

type createPolicyReq struct {
	Title    string `json:"title"    binding:"required"`
	Slug     string `json:"slug"     binding:"required"`
	Content  string `json:"content"  binding:"required"`
	Category string `json:"category" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type updatePolicyReq struct {
	Title    string `json:"title"    binding:"required"`
	Slug     string `json:"slug"     binding:"required"`
	Content  string `json:"content"  binding:"required"`
	Category string `json:"category" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type searchReq struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// POST /policies
func (h *Handler) createPolicy(c *gin.Context) {
	var req createPolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy, err := h.store.CreatePolicy(c.Request.Context(), req.Title, req.Slug, req.Content, req.Category, isActive)
	if err != nil {
		h.log.Error("create policy", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create policy"})
		return
	}

	if err := h.sync.SyncPolicyChunks(c.Request.Context(), policy.ID, policy.Content); err != nil {
		h.log.Warn("chunk sync failed after create", "policy_id", policy.ID, "err", err)
	}

	c.JSON(http.StatusCreated, gin.H{"data": policy})
}

// GET /policies
func (h *Handler) listPolicies(c *gin.Context) {
	category := c.Query("category")
	page := max(1, intQuery(c, "page", 1))
	limit := clamp(intQuery(c, "limit", 10), 1, 100)
	offset := (page - 1) * limit

	result, err := h.store.ListPolicies(c.Request.Context(), category, limit, offset)
	if err != nil {
		h.log.Error("list policies", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list policies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"items":  result.Policies,
		"total":  result.Total,
		"page":   page,
		"limit":  limit,
		"offset": offset,
	}})
}

// GET /policies/:id
func (h *Handler) getPolicy(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}

	policy, err := h.store.GetPolicy(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		h.log.Error("get policy", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": policy})
}

// GET /policies/slug/:slug
func (h *Handler) getPolicyBySlug(c *gin.Context) {
	slug := c.Param("slug")

	policy, err := h.store.GetPolicyBySlug(c.Request.Context(), slug)
	if err != nil {
		h.log.Error("get policy by slug", "slug", slug, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get policy"})
		return
	}
	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": policy})
}

// PUT /policies/:id
func (h *Handler) updatePolicy(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}

	var req updatePolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy, err := h.store.UpdatePolicy(c.Request.Context(), id, req.Title, req.Slug, req.Content, req.Category, isActive)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		h.log.Error("update policy", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update policy"})
		return
	}

	if err := h.sync.SyncPolicyChunks(c.Request.Context(), policy.ID, policy.Content); err != nil {
		h.log.Warn("chunk sync failed after update", "policy_id", policy.ID, "err", err)
	}

	c.JSON(http.StatusOK, gin.H{"data": policy})
}

// DELETE /policies/:id
func (h *Handler) deletePolicy(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}

	if err := h.store.DeletePolicy(c.Request.Context(), id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		h.log.Error("delete policy", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "policy deleted successfully"}})
}

// POST /policies/:id/sync
func (h *Handler) syncPolicy(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}

	policy, err := h.store.GetPolicy(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		h.log.Error("get policy for sync", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get policy"})
		return
	}

	if err := h.sync.SyncPolicyChunks(c.Request.Context(), policy.ID, policy.Content); err != nil {
		h.log.Error("sync policy chunks", "policy_id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "chunk sync failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"message":   "chunks synchronized successfully",
		"policy_id": id.String(),
	}})
}

// POST /policies/search
func (h *Handler) searchPolicies(c *gin.Context) {
	var req searchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Limit <= 0 || req.Limit > 20 {
		req.Limit = 5
	}

	embedding, err := h.embedder.Embed(c.Request.Context(), req.Query)
	if err != nil {
		h.log.Error("embed search query", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate query embedding"})
		return
	}

	results, err := h.store.SearchPolicies(c.Request.Context(), embedding, req.Limit)
	if err != nil {
		h.log.Error("search policies", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"query":   req.Query,
		"limit":   req.Limit,
		"results": results,
	}})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func parseUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return uuid.UUID{}, false
	}
	return id, true
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func intQuery(c *gin.Context, key string, def int) int {
	v := 0
	if _, err := fmt.Sscanf(c.Query(key), "%d", &v); err != nil || v <= 0 {
		return def
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
