package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/controller"
	"backend/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockCatalogUsecase struct {
	CreateCategoryFunc    func(ctx context.Context, c *domain.Category) (*domain.Category, error)
	ListCategoryTreeFunc  func(ctx context.Context) ([]*domain.CategoryNode, error)
	CreateBrandFunc       func(ctx context.Context, b *domain.Brand) (*domain.Brand, error)
	ListBrandsFunc        func(ctx context.Context) ([]*domain.Brand, error)
	CreateProductFunc     func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	SearchProductsFunc    func(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*domain.Product, int64, error)
	GetProductDetailsFunc func(ctx context.Context, id string) (*domain.ProductDetail, error)
	AddOptionValuesFunc   func(ctx context.Context, optionTypeID int, values []domain.ProductOptionValue) ([]domain.ProductOptionValue, error)
	GenerateVariantFunc   func(ctx context.Context, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error)
}

func (m *mockCatalogUsecase) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	return m.CreateCategoryFunc(ctx, c)
}

func (m *mockCatalogUsecase) ListCategoryTree(ctx context.Context) ([]*domain.CategoryNode, error) {
	return m.ListCategoryTreeFunc(ctx)
}

func (m *mockCatalogUsecase) CreateBrand(ctx context.Context, b *domain.Brand) (*domain.Brand, error) {
	return m.CreateBrandFunc(ctx, b)
}

func (m *mockCatalogUsecase) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return m.ListBrandsFunc(ctx)
}

func (m *mockCatalogUsecase) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	return m.CreateProductFunc(ctx, req)
}

func (m *mockCatalogUsecase) SearchProducts(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
	return m.SearchProductsFunc(ctx, query, categoryID, brandID, specsQuery, page, limit)
}

func (m *mockCatalogUsecase) GetProductDetails(ctx context.Context, id string) (*domain.ProductDetail, error) {
	return m.GetProductDetailsFunc(ctx, id)
}

func (m *mockCatalogUsecase) AddOptionValues(ctx context.Context, optionTypeID int, values []domain.ProductOptionValue) ([]domain.ProductOptionValue, error) {
	return m.AddOptionValuesFunc(ctx, optionTypeID, values)
}

func (m *mockCatalogUsecase) GenerateVariant(ctx context.Context, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error) {
	return m.GenerateVariantFunc(ctx, req)
}

func TestCatalogController_ListCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success list",
			setupMock: func(m *mockCatalogUsecase) {
				m.ListCategoryTreeFunc = func(ctx context.Context) ([]*domain.CategoryNode, error) {
					return []*domain.CategoryNode{
						{Category: &domain.Category{ID: 1, Name: "Tech"}},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error list",
			setupMock: func(m *mockCatalogUsecase) {
				m.ListCategoryTreeFunc = func(ctx context.Context) ([]*domain.CategoryNode, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.GET("/categories", ctl.ListCategories)

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_ListBrands(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success list",
			setupMock: func(m *mockCatalogUsecase) {
				m.ListBrandsFunc = func(ctx context.Context) ([]*domain.Brand, error) {
					return []*domain.Brand{
						{ID: 1, Name: "Samsung"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error list",
			setupMock: func(m *mockCatalogUsecase) {
				m.ListBrandsFunc = func(ctx context.Context) ([]*domain.Brand, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.GET("/brands", ctl.ListBrands)

			req := httptest.NewRequest(http.MethodGet, "/brands", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_SearchProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		url            string
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success empty query",
			url:  "/products",
			setupMock: func(m *mockCatalogUsecase) {
				m.SearchProductsFunc = func(ctx context.Context, query string, categoryID *int, brandID *int, specs map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
					return []*domain.Product{{ID: "1", Name: "Laptop"}}, 1, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success complete query",
			url:  "/products?q=laptop&category_id=2&brand_id=3&page=1&limit=5&specs=%7B%22ram%22%3A%2216GB%22%7D", // specs={"ram":"16GB"} URL encoded
			setupMock: func(m *mockCatalogUsecase) {
				m.SearchProductsFunc = func(ctx context.Context, query string, categoryID *int, brandID *int, specs map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
					if query == "laptop" && categoryID != nil && *categoryID == 2 && brandID != nil && *brandID == 3 && specs["ram"] == "16GB" && page == 1 && limit == 5 {
						return []*domain.Product{{ID: "1", Name: "Laptop"}}, 1, nil
					}
					return nil, 0, errors.New("mismatched params")
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error search",
			url:  "/products",
			setupMock: func(m *mockCatalogUsecase) {
				m.SearchProductsFunc = func(ctx context.Context, query string, categoryID *int, brandID *int, specs map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
					return nil, 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.GET("/products", ctl.SearchProducts)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_GetProductDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productIDParam string
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name:           "success details",
			productIDParam: "laptop-123",
			setupMock: func(m *mockCatalogUsecase) {
				m.GetProductDetailsFunc = func(ctx context.Context, id string) (*domain.ProductDetail, error) {
					return &domain.ProductDetail{
						Product: &domain.Product{ID: id, Name: "Super Laptop"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "product not found",
			productIDParam: "unknown",
			setupMock: func(m *mockCatalogUsecase) {
				m.GetProductDetailsFunc = func(ctx context.Context, id string) (*domain.ProductDetail, error) {
					return nil, errors.New("product not found")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.GET("/products/:id", ctl.GetProductDetails)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productIDParam, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_CreateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success created",
			body: domain.Category{Name: "Tech"},
			setupMock: func(m *mockCatalogUsecase) {
				m.CreateCategoryFunc = func(ctx context.Context, c *domain.Category) (*domain.Category, error) {
					c.ID = 1
					return c, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "validation error - empty name",
			body:           domain.Category{Name: ""},
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "usecase internal error",
			body: domain.Category{Name: "Tech"},
			setupMock: func(m *mockCatalogUsecase) {
				m.CreateCategoryFunc = func(ctx context.Context, c *domain.Category) (*domain.Category, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.POST("/admin/categories", ctl.CreateCategory)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/categories", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_CreateBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success created",
			body: domain.Brand{Name: "Apple"},
			setupMock: func(m *mockCatalogUsecase) {
				m.CreateBrandFunc = func(ctx context.Context, b *domain.Brand) (*domain.Brand, error) {
					b.ID = 1
					return b, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "validation error - empty name",
			body:           domain.Brand{Name: ""},
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.POST("/admin/brands", ctl.CreateBrand)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/brands", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success created",
			body: domain.CreateProductRequest{
				Product: &domain.Product{
					ID:         "macbook-pro",
					CategoryID: 1,
					BrandID:    1,
					Name:       "MacBook Pro",
				},
			},
			setupMock: func(m *mockCatalogUsecase) {
				m.CreateProductFunc = func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
					return req.Product, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty product",
			body: domain.CreateProductRequest{
				Product: nil,
			},
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.POST("/admin/products", ctl.CreateProduct)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/products", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_AddOptionValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name: "success added",
			body: domain.CreateOptionValueRequest{
				OptionTypeID: 1,
				Values: []domain.ProductOptionValue{
					{Value: "Space Gray"},
				},
			},
			setupMock: func(m *mockCatalogUsecase) {
				m.AddOptionValuesFunc = func(ctx context.Context, optionTypeID int, values []domain.ProductOptionValue) ([]domain.ProductOptionValue, error) {
					return values, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty values list",
			body: domain.CreateOptionValueRequest{
				OptionTypeID: 1,
				Values:       nil,
			},
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.POST("/admin/option-values", ctl.AddOptionValues)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/option-values", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestCatalogController_GenerateVariant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productIDParam string
		body           interface{}
		setupMock      func(m *mockCatalogUsecase)
		expectedStatus int
	}{
		{
			name:           "success generated",
			productIDParam: "laptop-123",
			body: domain.GenerateVariantRequest{
				Variant: &domain.ProductVariant{
					Name:  "MacBook Pro 16",
					SKU:   "MBP16-SG",
					Price: 1999.99,
				},
				OptionValueIDs: []int{1, 2},
			},
			setupMock: func(m *mockCatalogUsecase) {
				m.GenerateVariantFunc = func(ctx context.Context, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error) {
					return req.Variant, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			productIDParam: "laptop-123",
			body:           "invalid-json",
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "validation error - empty SKU",
			productIDParam: "laptop-123",
			body: domain.GenerateVariantRequest{
				Variant: &domain.ProductVariant{
					Name:  "MacBook Pro 16",
					SKU:   "", // required
					Price: 1999.99,
				},
				OptionValueIDs: []int{1, 2},
			},
			setupMock:      func(m *mockCatalogUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockCatalogUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewCatalogController(mockUC)
			r := gin.New()
			r.POST("/admin/products/:id/variants", ctl.GenerateVariant)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/products/"+tt.productIDParam+"/variants", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}
