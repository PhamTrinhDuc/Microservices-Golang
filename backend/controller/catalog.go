package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CatalogController struct {
	useCase  domain.CatalogUsecase
	validate *validator.Validate
}

func NewCatalogController(useCase domain.CatalogUsecase) *CatalogController {
	return &CatalogController{
		useCase:  useCase,
		validate: validator.New(),
	}
}

// handleError maps domain errors to appropriate HTTP responses
func (cc *CatalogController) handleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	if errors.Is(err, domain.ErrCategoryNotFound) ||
		errors.Is(err, domain.ErrBrandNotFound) ||
		errors.Is(err, domain.ErrProductNotFound) ||
		errors.Is(err, domain.ErrVariantNotFound) ||
		errors.Is(err, domain.ErrProductOptionNotFound) ||
		errors.Is(err, domain.ErrOptionValueNotFound) {
		utils.RespondNotFound(ctx, err.Error())
		return
	}

	if errors.Is(err, domain.ErrDuplicateSlug) {
		ctx.JSON(http.StatusConflict, utils.HTTPResponse{Error: err.Error()})
		return
	}

	// Bad request for user input mistakes
	if strings.Contains(err.Error(), "does not exist") ||
		strings.Contains(err.Error(), "cannot be its own parent") ||
		strings.Contains(err.Error(), "already in use") {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	utils.RespondInternalError(ctx, err.Error())
}

// --- Category ---

// CreateCategory handles POST /admin/categories
func (cc *CatalogController) CreateCategory(ctx *gin.Context) {
	var req domain.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.CreateCategory(ctx.Request.Context(), &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondCreated(ctx, result)
}

// ListCategories handles GET /categories
func (cc *CatalogController) ListCategories(ctx *gin.Context) {
	result, err := cc.useCase.ListCategories(ctx.Request.Context())
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// UpdateCategory handles PUT /admin/categories/:id
func (cc *CatalogController) UpdateCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid category id param")
		return
	}

	var req domain.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.UpdateCategory(ctx.Request.Context(), id, &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// DeleteCategory handles DELETE /admin/categories/:id
func (cc *CatalogController) DeleteCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid category id param")
		return
	}

	err = cc.useCase.DeleteCategory(ctx.Request.Context(), id)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Brand ---

// CreateBrand handles POST /admin/brands
func (cc *CatalogController) CreateBrand(ctx *gin.Context) {
	var req domain.CreateBrandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.CreateBrand(ctx.Request.Context(), &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondCreated(ctx, result)
}

// ListBrands handles GET /brands
func (cc *CatalogController) ListBrands(ctx *gin.Context) {
	result, err := cc.useCase.ListBrands(ctx.Request.Context())
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// UpdateBrand handles PUT /admin/brands/:id
func (cc *CatalogController) UpdateBrand(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid brand id param")
		return
	}

	var req domain.UpdateBrandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.UpdateBrand(ctx.Request.Context(), id, &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// DeleteBrand handles DELETE /admin/brands/:id
func (cc *CatalogController) DeleteBrand(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid brand id param")
		return
	}

	err = cc.useCase.DeleteBrand(ctx.Request.Context(), id)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Product ---

// CreateProduct handles POST /admin/products
func (cc *CatalogController) CreateProduct(ctx *gin.Context) {
	var req domain.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.CreateProduct(ctx.Request.Context(), &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondCreated(ctx, result)
}

// SearchProducts handles GET /products
func (cc *CatalogController) SearchProducts(ctx *gin.Context) {
	var query domain.ProductSearchQuery

	// Gin binds query params to struct
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	// Fallback/Default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	result, err := cc.useCase.SearchProducts(ctx.Request.Context(), &query)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// GetProductDetails handles GET /products/:id
func (cc *CatalogController) GetProductDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		utils.RespondBadRequest(ctx, "invalid product id or slug param")
		return
	}

	result, err := cc.useCase.GetProductDetails(ctx.Request.Context(), id)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// UpdateProduct handles PUT /admin/products/:id
func (cc *CatalogController) UpdateProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		utils.RespondBadRequest(ctx, "invalid product id param")
		return
	}

	var req domain.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.UpdateProduct(ctx.Request.Context(), id, &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// DeleteProduct handles DELETE /admin/products/:id
func (cc *CatalogController) DeleteProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		utils.RespondBadRequest(ctx, "invalid product id param")
		return
	}

	err := cc.useCase.DeleteProduct(ctx.Request.Context(), id)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Option Types & Values ---

// AddOptionValues handles POST /admin/option-values
func (cc *CatalogController) AddOptionValues(ctx *gin.Context) {
	var req domain.AddOptionValuesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.AddOptionValues(ctx.Request.Context(), &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondCreated(ctx, result)
}

// --- Variant ---

// GenerateVariant handles POST /admin/products/:id/variants
func (cc *CatalogController) GenerateVariant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		utils.RespondBadRequest(ctx, "invalid product id param")
		return
	}

	var req domain.GenerateVariantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.GenerateVariant(ctx.Request.Context(), id, &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondCreated(ctx, result)
}

// UpdateVariant handles PUT /admin/variants/:id
func (cc *CatalogController) UpdateVariant(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid variant id param")
		return
	}

	var req domain.UpdateVariantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.UpdateVariant(ctx.Request.Context(), id, &req)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondOK(ctx, result)
}

// DeleteVariant handles DELETE /admin/variants/:id
func (cc *CatalogController) DeleteVariant(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid variant id param")
		return
	}

	err = cc.useCase.DeleteVariant(ctx.Request.Context(), id)
	if err != nil {
		cc.handleError(ctx, err)
		return
	}

	utils.RespondNoContent(ctx)
}

