package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/controller"
	"backend/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockCatalogUsecase struct {
	CreateCategoryFunc func(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.Category, error)
	ListCategoriesFunc func(ctx context.Context) ([]*domain.Category, error)
	UpdateCategoryFunc func(ctx context.Context, id int, req *domain.UpdateCategoryRequest) (*domain.Category, error)
	DeleteCategoryFunc func(ctx context.Context, id int) error

	CreateBrandFunc func(ctx context.Context, req *domain.CreateBrandRequest) (*domain.Brand, error)
	ListBrandsFunc  func(ctx context.Context) ([]*domain.Brand, error)
	UpdateBrandFunc func(ctx context.Context, id int, req *domain.UpdateBrandRequest) (*domain.Brand, error)
	DeleteBrandFunc func(ctx context.Context, id int) error

	CreateProductFunc     func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	SearchProductsFunc    func(ctx context.Context, query *domain.ProductSearchQuery) (*domain.ProductSearchResult, error)
	GetProductDetailsFunc func(ctx context.Context, id string) (*domain.ProductDetailsResponse, error)
	UpdateProductFunc     func(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error)
	DeleteProductFunc     func(ctx context.Context, id string) error

	AddOptionValuesFunc func(ctx context.Context, req *domain.AddOptionValuesRequest) ([]*domain.ProductOptionValue, error)
	GenerateVariantFunc func(ctx context.Context, productID string, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error)
	UpdateVariantFunc   func(ctx context.Context, id int, req *domain.UpdateVariantRequest) (*domain.ProductVariant, error)
	DeleteVariantFunc   func(ctx context.Context, id int) error
}

func (m *mockCatalogUsecase) CreateCategory(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.Category, error) {
	return m.CreateCategoryFunc(ctx, req)
}

func (m *mockCatalogUsecase) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return m.ListCategoriesFunc(ctx)
}

func (m *mockCatalogUsecase) UpdateCategory(ctx context.Context, id int, req *domain.UpdateCategoryRequest) (*domain.Category, error) {
	return m.UpdateCategoryFunc(ctx, id, req)
}

func (m *mockCatalogUsecase) DeleteCategory(ctx context.Context, id int) error {
	return m.DeleteCategoryFunc(ctx, id)
}

func (m *mockCatalogUsecase) CreateBrand(ctx context.Context, req *domain.CreateBrandRequest) (*domain.Brand, error) {
	return m.CreateBrandFunc(ctx, req)
}

func (m *mockCatalogUsecase) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return m.ListBrandsFunc(ctx)
}

func (m *mockCatalogUsecase) UpdateBrand(ctx context.Context, id int, req *domain.UpdateBrandRequest) (*domain.Brand, error) {
	return m.UpdateBrandFunc(ctx, id, req)
}

func (m *mockCatalogUsecase) DeleteBrand(ctx context.Context, id int) error {
	return m.DeleteBrandFunc(ctx, id)
}

func (m *mockCatalogUsecase) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	return m.CreateProductFunc(ctx, req)
}

func (m *mockCatalogUsecase) SearchProducts(ctx context.Context, query *domain.ProductSearchQuery) (*domain.ProductSearchResult, error) {
	return m.SearchProductsFunc(ctx, query)
}

func (m *mockCatalogUsecase) GetProductDetails(ctx context.Context, id string) (*domain.ProductDetailsResponse, error) {
	return m.GetProductDetailsFunc(ctx, id)
}

func (m *mockCatalogUsecase) UpdateProduct(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error) {
	return m.UpdateProductFunc(ctx, id, req)
}

func (m *mockCatalogUsecase) DeleteProduct(ctx context.Context, id string) error {
	return m.DeleteProductFunc(ctx, id)
}

func (m *mockCatalogUsecase) AddOptionValues(ctx context.Context, req *domain.AddOptionValuesRequest) ([]*domain.ProductOptionValue, error) {
	return m.AddOptionValuesFunc(ctx, req)
}

func (m *mockCatalogUsecase) GenerateVariant(ctx context.Context, productID string, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error) {
	return m.GenerateVariantFunc(ctx, productID, req)
}

func (m *mockCatalogUsecase) UpdateVariant(ctx context.Context, id int, req *domain.UpdateVariantRequest) (*domain.ProductVariant, error) {
	return m.UpdateVariantFunc(ctx, id, req)
}

func (m *mockCatalogUsecase) DeleteVariant(ctx context.Context, id int) error {
	return m.DeleteVariantFunc(ctx, id)
}

func TestCatalogController_CreateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success created", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCatalogUsecase{
			CreateCategoryFunc: func(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.Category, error) {
				return &domain.Category{ID: 1, Name: req.Name, Slug: "dieu-hoa"}, nil
			},
		}

		ctl := controller.NewCatalogController(mockUC)
		r := gin.New()
		r.POST("/admin/categories", ctl.CreateCategory)

		reqBody := domain.CreateCategoryRequest{Name: "Điều Hòa"}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/categories", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCatalogUsecase{}

		ctl := controller.NewCatalogController(mockUC)
		r := gin.New()
		r.POST("/admin/categories", ctl.CreateCategory)

		reqBody := domain.CreateCategoryRequest{Name: ""} // Name is required
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/categories", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})
}

func TestCatalogController_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success created", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCatalogUsecase{
			CreateProductFunc: func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
				return &domain.Product{ID: req.ID, Name: req.Name}, nil
			},
		}

		ctl := controller.NewCatalogController(mockUC)
		r := gin.New()
		r.POST("/admin/products", ctl.CreateProduct)

		reqBody := domain.CreateProductRequest{
			ID:         "iphone-17",
			CategoryID: 1,
			BrandID:    1,
			Name:       "iPhone 17",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/products", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("duplicate slug conflict", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCatalogUsecase{
			CreateProductFunc: func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
				return nil, domain.ErrDuplicateSlug
			},
		}

		ctl := controller.NewCatalogController(mockUC)
		r := gin.New()
		r.POST("/admin/products", ctl.CreateProduct)

		reqBody := domain.CreateProductRequest{
			ID:         "iphone-17",
			CategoryID: 1,
			BrandID:    1,
			Name:       "iPhone 17",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/products", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusConflict, w.Code)
	})
}
