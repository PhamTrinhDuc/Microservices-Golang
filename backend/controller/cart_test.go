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

type mockCartUsecase struct {
	GetCartFunc            func(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error)
	AddToCartFunc          func(ctx context.Context, userID *int, sessionID *string, req *domain.AddToCartRequest) (*domain.CartItem, error)
	UpdateItemQuantityFunc func(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error)
	RemoveItemFunc         func(ctx context.Context, userID *int, sessionID *string, itemID int) error
	ClearCartFunc          func(ctx context.Context, userID *int, sessionID *string) error
	MergeCartFunc          func(ctx context.Context, userID int, sessionID string) error
}

func (m *mockCartUsecase) GetCart(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error) {
	return m.GetCartFunc(ctx, userID, sessionID)
}

func (m *mockCartUsecase) AddToCart(ctx context.Context, userID *int, sessionID *string, req *domain.AddToCartRequest) (*domain.CartItem, error) {
	return m.AddToCartFunc(ctx, userID, sessionID, req)
}

func (m *mockCartUsecase) UpdateItemQuantity(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error) {
	return m.UpdateItemQuantityFunc(ctx, userID, sessionID, itemID, quantity)
}

func (m *mockCartUsecase) RemoveItem(ctx context.Context, userID *int, sessionID *string, itemID int) error {
	return m.RemoveItemFunc(ctx, userID, sessionID, itemID)
}

func (m *mockCartUsecase) ClearCart(ctx context.Context, userID *int, sessionID *string) error {
	return m.ClearCartFunc(ctx, userID, sessionID)
}

func (m *mockCartUsecase) MergeCart(ctx context.Context, userID int, sessionID string) error {
	return m.MergeCartFunc(ctx, userID, sessionID)
}

func TestCartController_AddToCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success guest add", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			AddToCartFunc: func(ctx context.Context, userID *int, sessionID *string, req *domain.AddToCartRequest) (*domain.CartItem, error) {
				guestID := "guest-uuid"
				return &domain.CartItem{
					ID:        1,
					SessionID: &guestID,
					VariantID: req.VariantID,
					Quantity:  req.Quantity,
				}, nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.POST("/api/v1/cart", ctl.AddToCart)

		reqBody := domain.AddToCartRequest{
			VariantID: 10,
			Quantity:  2,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("success user authenticated add", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			AddToCartFunc: func(ctx context.Context, userID *int, sessionID *string, req *domain.AddToCartRequest) (*domain.CartItem, error) {
				uID := 5
				return &domain.CartItem{
					ID:        2,
					UserID:    &uID,
					VariantID: req.VariantID,
					Quantity:  req.Quantity,
				}, nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.POST("/api/v1/cart", func(c *gin.Context) {
			c.Set("user_id", 5) // Authenticated user ID in context
			ctl.AddToCart(c)
		})

		reqBody := domain.AddToCartRequest{
			VariantID: 15,
			Quantity:  1,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("validation error invalid quantity", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.POST("/api/v1/cart", ctl.AddToCart)

		reqBody := domain.AddToCartRequest{
			VariantID: 10,
			Quantity:  0, // Invalid quantity, must be gt=0
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})

	t.Run("missing identity info", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.POST("/api/v1/cart", ctl.AddToCart)

		reqBody := domain.AddToCartRequest{
			VariantID: 10,
			Quantity:  3,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		// X-Session-ID header and query session_id are missing
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})
}

func TestCartController_GetCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success view guest cart", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			GetCartFunc: func(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error) {
				url := "http://image.url"
				return []*domain.CartItemResponse{
					{
						ID:          1,
						VariantID:   12,
						VariantName: "Blue",
						SKU:         "V-BLUE",
						SellPrice:   200000,
						ProductID:   "prod-1",
						ProductName: "Product 1",
						ImageURL:    &url,
						Quantity:    2,
					},
				}, nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.GET("/api/v1/cart", ctl.GetCart)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		is.Len(data, 1)

		item := data[0].(map[string]interface{})
		is.Equal(float64(12), item["variant_id"])
		is.Equal("V-BLUE", item["sku"])
		is.Equal(float64(2), item["quantity"])
	})
}

func TestCartController_UpdateItemQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success update quantity", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			UpdateItemQuantityFunc: func(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error) {
				return &domain.CartItem{
					ID:        itemID,
					VariantID: 10,
					Quantity:  quantity,
				}, nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.PUT("/api/v1/cart/items/:id", ctl.UpdateItemQuantity)

		reqBody := domain.UpdateQuantityRequest{Quantity: 5}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cart/items/1", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)
	})

	t.Run("cart item not found", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			UpdateItemQuantityFunc: func(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error) {
				return nil, domain.ErrCartItemNotFound
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.PUT("/api/v1/cart/items/:id", ctl.UpdateItemQuantity)

		reqBody := domain.UpdateQuantityRequest{Quantity: 5}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cart/items/99", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNotFound, w.Code)
	})

	t.Run("permission denied forbidden", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			UpdateItemQuantityFunc: func(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error) {
				return nil, domain.ErrUnauthorized
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.PUT("/api/v1/cart/items/:id", ctl.UpdateItemQuantity)

		reqBody := domain.UpdateQuantityRequest{Quantity: 5}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cart/items/1", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusForbidden, w.Code)
	})
}

func TestCartController_RemoveItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success delete item", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			RemoveItemFunc: func(ctx context.Context, userID *int, sessionID *string, itemID int) error {
				return nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.DELETE("/api/v1/cart/items/:id", ctl.RemoveItem)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/cart/items/1", nil)
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNoContent, w.Code)
	})
}

func TestCartController_ClearCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success clear cart", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			ClearCartFunc: func(ctx context.Context, userID *int, sessionID *string) error {
				return nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.DELETE("/api/v1/cart", ctl.ClearCart)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/cart", nil)
		req.Header.Set("X-Session-ID", "guest-uuid")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNoContent, w.Code)
	})
}

func TestCartController_MergeCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success merge cart", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockCartUsecase{
			MergeCartFunc: func(ctx context.Context, userID int, sessionID string) error {
				return nil
			},
		}

		ctl := controller.NewCartController(mockUC)
		r := gin.New()
		r.POST("/api/v1/cart/merge", func(c *gin.Context) {
			c.Set("user_id", 5) // Authenticated user ID
			ctl.MergeCart(c)
		})

		reqBody := domain.MergeCartRequest{SessionID: "guest-uuid"}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/merge", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)
	})
}
