package controller

import (
	"encoding/json"
	"strconv"

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

// ListCategories handles GET /categories.
func (cc *CatalogController) ListCategories(ctx *gin.Context) {
	tree, err := cc.useCase.ListCategoryTree(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, tree)
}

// ListBrands handles GET /brands.
func (cc *CatalogController) ListBrands(ctx *gin.Context) {
	brands, err := cc.useCase.ListBrands(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, brands)
}

// SearchProducts handles GET /products.
func (cc *CatalogController) SearchProducts(ctx *gin.Context) {
	query := ctx.Query("q")
	page := utils.GetQueryInt(ctx, "page", 1)
	limit := utils.GetQueryInt(ctx, "limit", 10)

	var categoryID *int
	if catStr := ctx.Query("category_id"); catStr != "" {
		if catID, err := strconv.Atoi(catStr); err == nil {
			categoryID = &catID
		}
	}

	var brandID *int
	if brandStr := ctx.Query("brand_id"); brandStr != "" {
		if brID, err := strconv.Atoi(brandStr); err == nil {
			brandID = &brID
		}
	}

	// Parse optional specs JSON string filter (e.g. ?specs={"ram":"8GB"})
	specsMap := make(map[string]interface{})
	if specsStr := ctx.Query("specs"); specsStr != "" {
		_ = json.Unmarshal([]byte(specsStr), &specsMap)
	}

	products, total, err := cc.useCase.SearchProducts(ctx.Request.Context(), query, categoryID, brandID, specsMap, page, limit)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	// Structured pagination output matching helper response format
	utils.RespondOK(ctx, gin.H{
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetProductDetails handles GET /products/:id.
func (cc *CatalogController) GetProductDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		utils.RespondBadRequest(ctx, "product id is required")
		return
	}

	details, err := cc.useCase.GetProductDetails(ctx.Request.Context(), id)
	if err != nil {
		utils.RespondNotFound(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, details)
}

// CreateCategory handles POST /admin/categories.
func (cc *CatalogController) CreateCategory(ctx *gin.Context) {
	var category domain.Category
	if err := ctx.ShouldBindJSON(&category); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&category); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.CreateCategory(ctx.Request.Context(), &category)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// CreateBrand handles POST /admin/brands.
func (cc *CatalogController) CreateBrand(ctx *gin.Context) {
	var brand domain.Brand
	if err := ctx.ShouldBindJSON(&brand); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&brand); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.CreateBrand(ctx.Request.Context(), &brand)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// CreateProduct handles POST /admin/products.
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
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// AddOptionValues handles POST /admin/option-values.
func (cc *CatalogController) AddOptionValues(ctx *gin.Context) {
	var req domain.CreateOptionValueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.AddOptionValues(ctx.Request.Context(), req.OptionTypeID, req.Values)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// GenerateVariant handles POST /admin/products/:id/variants.
func (cc *CatalogController) GenerateVariant(ctx *gin.Context) {
	id := ctx.Param("id")
	var req domain.GenerateVariantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if req.Variant != nil {
		req.Variant.ProductID = id
	}

	if err := cc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := cc.useCase.GenerateVariant(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}
