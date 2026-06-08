package controller

import (
	"net/http"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PolicyController struct {
	usecase  domain.PolicyUsecase
	validate *validator.Validate
}

func NewPolicyController(usecase domain.PolicyUsecase) *PolicyController {
	return &PolicyController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// Create handles POST /admin/policies
// @Summary Create a new policy
// @Description Creates a new policy and synchronizes its vector chunks in the database for the chatbot RAG
// @Tags Admin Policies
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param policy body domain.CreatePolicyRequest true "Policy object"
// @Success 201 {object} domain.Policy
// @Failure 400 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/admin/policies [post]
func (ctl *PolicyController) Create(ctx *gin.Context) {
	var req domain.CreatePolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}
	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.Create(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondCreated(ctx, res)
}

// Update handles PUT /admin/policies/:id
// @Summary Update an existing policy
// @Description Updates a policy and re-synchronizes its vector chunks in the database for the chatbot RAG
// @Tags Admin Policies
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Policy ID (UUID)"
// @Param policy body domain.UpdatePolicyRequest true "Policy object"
// @Success 200 {object} domain.Policy
// @Failure 400 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/admin/policies/{id} [put]
func (ctl *PolicyController) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid policy ID format (must be UUID)")
		return
	}

	var req domain.UpdatePolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}
	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.Update(ctx.Request.Context(), id, &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

// Delete handles DELETE /admin/policies/:id
// @Summary Delete a policy
// @Description Deletes a policy and all its vector chunks
// @Tags Admin Policies
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Policy ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/admin/policies/{id} [delete]
func (ctl *PolicyController) Delete(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid policy ID format (must be UUID)")
		return
	}

	err = ctl.usecase.Delete(ctx.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondNoContent(ctx)
}

// GetBySlug handles GET /policies/:slug
// @Summary Get policy details by slug
// @Description Retrieves full policy details by its unique slug (for web display)
// @Tags Public Policies
// @Produce json
// @Param slug path string true "Policy slug"
// @Success 200 {object} domain.Policy
// @Failure 404 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/policies/{slug} [get]
func (ctl *PolicyController) GetBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if slug == "" {
		utils.RespondBadRequest(ctx, "slug is required")
		return
	}

	res, err := ctl.usecase.GetBySlug(ctx.Request.Context(), slug)
	if err != nil {
		utils.RespondNotFound(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

// List handles GET /policies
// @Summary List policies
// @Description Retrieves a list of policies, optionally filtered by category
// @Tags Public Policies
// @Produce json
// @Param category query string false "Filter by category (e.g. refund, shipping)"
// @Param limit query int false "Limit count"
// @Param offset query int false "Offset count"
// @Success 200 {object} map[string]interface{} "List of policies and total count"
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/policies [get]
func (ctl *PolicyController) List(ctx *gin.Context) {
	category := ctx.Query("category")
	limit := utils.GetQueryInt(ctx, "limit", 10)
	offset := utils.GetQueryInt(ctx, "offset", 0)

	policies, total, err := ctl.usecase.List(ctx.Request.Context(), category, limit, offset)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{
		"policies": policies,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// ChatbotSearch handles POST /policies/chat
// @Summary Chatbot vector RAG search
// @Description Semantic search query over policy text fragments using vector embeddings
// @Tags Public Policies
// @Accept json
// @Produce json
// @Param query body domain.ChatQueryRequest true "Chat query object"
// @Success 200 {array} domain.PolicySearchResult "Top closest policy text segments"
// @Failure 400 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /api/v1/policies/chat [post]
func (ctl *PolicyController) ChatbotSearch(ctx *gin.Context) {
	var req domain.ChatQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}
	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 3
	}

	res, err := ctl.usecase.Search(ctx.Request.Context(), req.Query, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.RespondOK(ctx, res)
}
