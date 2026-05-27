package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"backend/domain"
)

type CatalogUsecase struct {
	repo domain.CatalogRepository
}

func NewCatalogUsecase(repo domain.CatalogRepository) *CatalogUsecase {
	return &CatalogUsecase{repo: repo}
}

// slugify formats a string into a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)
	// Replace non-alphanumeric chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")
	// Replace whitespace with hyphens
	s = strings.Join(strings.Fields(s), "-")
	return s
}

// ===================== Category operations =====================

func (uc *CatalogUsecase) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	if c.Name == "" {
		return nil, errors.New("category name is required")
	}

	if c.Slug == "" {
		c.Slug = slugify(c.Name)
	}

	// Validate parent if provided
	if c.ParentID != nil {
		_, err := uc.repo.GetCategoryByID(ctx, *c.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
	}

	return uc.repo.CreateCategory(ctx, c)
}

func (uc *CatalogUsecase) ListCategoryTree(ctx context.Context) ([]*domain.CategoryNode, error) {
	all, err := uc.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	// Create mapping nodes
	nodesMap := make(map[int]*domain.CategoryNode)
	for _, c := range all {
		nodesMap[c.ID] = &domain.CategoryNode{
			Category: c,
			Children: make([]*domain.CategoryNode, 0),
		}
	}

	rootNodes := make([]*domain.CategoryNode, 0)
	for _, c := range all {
		node := nodesMap[c.ID]
		if c.ParentID == nil {
			rootNodes = append(rootNodes, node)
		} else {
			parent, exists := nodesMap[*c.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not active/deleted, treat as root
				rootNodes = append(rootNodes, node)
			}
		}
	}

	return rootNodes, nil
}

// ===================== Brand operations =====================

func (uc *CatalogUsecase) CreateBrand(ctx context.Context, b *domain.Brand) (*domain.Brand, error) {
	if b.Name == "" {
		return nil, errors.New("brand name is required")
	}

	if b.Slug == "" {
		b.Slug = slugify(b.Name)
	}

	return uc.repo.CreateBrand(ctx, b)
}

func (uc *CatalogUsecase) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return uc.repo.ListBrands(ctx)
}

// ===================== Product operations =====================

func (uc *CatalogUsecase) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	p := req.Product
	if p == nil {
		return nil, errors.New("product metadata cannot be nil")
	}

	if p.ID == "" {
		return nil, errors.New("product ID is required")
	}

	if p.Slug == "" {
		p.Slug = slugify(p.Name)
	}

	// 1. Verify Category exists
	_, err := uc.repo.GetCategoryByID(ctx, p.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	// 2. Verify Brand exists
	_, err = uc.repo.GetBrandByID(ctx, p.BrandID)
	if err != nil {
		return nil, fmt.Errorf("invalid brand: %w", err)
	}

	// 3. Create Product with Specs & Options inside Repository transaction
	return uc.repo.CreateProduct(ctx, p, req.Specs, req.OptionTypes)
}

func (uc *CatalogUsecase) SearchProducts(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
	// Defaults fallback
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return uc.repo.ListProducts(ctx, query, categoryID, brandID, specsQuery, page, limit)
}

func (uc *CatalogUsecase) GetProductDetails(ctx context.Context, id string) (*domain.ProductDetail, error) {
	p, err := uc.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	brand, _ := uc.repo.GetBrandByID(ctx, p.BrandID)
	cat, _ := uc.repo.GetCategoryByID(ctx, p.CategoryID)
	specs, _ := uc.repo.GetSpecs(ctx, p.ID)
	optionTypes, _ := uc.repo.GetOptionTypes(ctx, p.ID)
	variants, _ := uc.repo.GetVariants(ctx, p.ID)
	images, _ := uc.repo.GetImages(ctx, p.ID)

	return &domain.ProductDetail{
		Product:     p,
		Brand:       brand,
		Category:    cat,
		Specs:       specs,
		OptionTypes: optionTypes,
		Variants:    variants,
		Images:      images,
	}, nil
}

// ===================== Option & Variant operations =====================

func (uc *CatalogUsecase) AddOptionValues(ctx context.Context, optionTypeID int, values []domain.ProductOptionValue) ([]domain.ProductOptionValue, error) {
	if len(values) == 0 {
		return nil, errors.New("values list cannot be empty")
	}

	return uc.repo.CreateOptionValues(ctx, optionTypeID, values)
}

func (uc *CatalogUsecase) GenerateVariant(ctx context.Context, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error) {
	v := req.Variant
	if v == nil {
		return nil, errors.New("variant cannot be nil")
	}

	if v.ProductID == "" || v.SKU == "" {
		return nil, errors.New("product_id and sku are required")
	}

	// Verify parent product exists
	_, err := uc.repo.GetProductByID(ctx, v.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent product: %w", err)
	}

	return uc.repo.CreateVariant(ctx, v, req.OptionValueIDs, req.Images)
}
