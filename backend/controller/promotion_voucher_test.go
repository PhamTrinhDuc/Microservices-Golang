package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/controller"
	"backend/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPromotionVoucherUsecase struct {
	CreatePromotionFunc      func(ctx context.Context, req *domain.CreatePromotionRequest) (*domain.Promotion, error)
	ListPromotionsFunc       func(ctx context.Context) ([]*domain.Promotion, error)
	GetPromotionByIDFunc     func(ctx context.Context, id int) (*domain.Promotion, error)
	UpdatePromotionFunc      func(ctx context.Context, id int, req *domain.UpdatePromotionRequest) (*domain.Promotion, error)
	DeletePromotionFunc      func(ctx context.Context, id int) error
	CreateVoucherFunc        func(ctx context.Context, req *domain.CreateVoucherRequest) (*domain.Voucher, error)
	ListVouchersFunc         func(ctx context.Context) ([]*domain.Voucher, error)
	GetVoucherByIDFunc       func(ctx context.Context, id int) (*domain.Voucher, error)
	UpdateVoucherFunc        func(ctx context.Context, id int, req *domain.UpdateVoucherRequest) (*domain.Voucher, error)
	DeleteVoucherFunc        func(ctx context.Context, id int) error
	ListActiveVouchersFunc   func(ctx context.Context) ([]*domain.Voucher, error)
	ApplyVoucherFunc         func(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error)
	UseVoucherFunc           func(ctx context.Context, userID int, voucherID int, orderID int) error
	ReleaseVoucherFunc       func(ctx context.Context, userID int, voucherID int, orderID int) error
}

func (m *mockPromotionVoucherUsecase) CreatePromotion(ctx context.Context, req *domain.CreatePromotionRequest) (*domain.Promotion, error) {
	return m.CreatePromotionFunc(ctx, req)
}

func (m *mockPromotionVoucherUsecase) ListPromotions(ctx context.Context) ([]*domain.Promotion, error) {
	return m.ListPromotionsFunc(ctx)
}

func (m *mockPromotionVoucherUsecase) GetPromotionByID(ctx context.Context, id int) (*domain.Promotion, error) {
	return m.GetPromotionByIDFunc(ctx, id)
}

func (m *mockPromotionVoucherUsecase) UpdatePromotion(ctx context.Context, id int, req *domain.UpdatePromotionRequest) (*domain.Promotion, error) {
	return m.UpdatePromotionFunc(ctx, id, req)
}

func (m *mockPromotionVoucherUsecase) DeletePromotion(ctx context.Context, id int) error {
	return m.DeletePromotionFunc(ctx, id)
}

func (m *mockPromotionVoucherUsecase) CreateVoucher(ctx context.Context, req *domain.CreateVoucherRequest) (*domain.Voucher, error) {
	return m.CreateVoucherFunc(ctx, req)
}

func (m *mockPromotionVoucherUsecase) ListVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	return m.ListVouchersFunc(ctx)
}

func (m *mockPromotionVoucherUsecase) GetVoucherByID(ctx context.Context, id int) (*domain.Voucher, error) {
	return m.GetVoucherByIDFunc(ctx, id)
}

func (m *mockPromotionVoucherUsecase) UpdateVoucher(ctx context.Context, id int, req *domain.UpdateVoucherRequest) (*domain.Voucher, error) {
	return m.UpdateVoucherFunc(ctx, id, req)
}

func (m *mockPromotionVoucherUsecase) DeleteVoucher(ctx context.Context, id int) error {
	return m.DeleteVoucherFunc(ctx, id)
}

func (m *mockPromotionVoucherUsecase) ListActiveVouchers(ctx context.Context) ([]*domain.Voucher, error) {
	return m.ListActiveVouchersFunc(ctx)
}

func (m *mockPromotionVoucherUsecase) ApplyVoucher(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
	return m.ApplyVoucherFunc(ctx, userID, req)
}

func (m *mockPromotionVoucherUsecase) UseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	return m.UseVoucherFunc(ctx, userID, voucherID, orderID)
}

func (m *mockPromotionVoucherUsecase) ReleaseVoucher(ctx context.Context, userID int, voucherID int, orderID int) error {
	return m.ReleaseVoucherFunc(ctx, userID, voucherID, orderID)
}

func TestPromotionVoucherController_CreatePromotion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success create promotion", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			CreatePromotionFunc: func(ctx context.Context, req *domain.CreatePromotionRequest) (*domain.Promotion, error) {
				return &domain.Promotion{
					ID:            1,
					ProductID:     req.ProductID,
					VariantID:     req.VariantID,
					Name:          req.Name,
					DiscountType:  req.DiscountType,
					DiscountValue: req.DiscountValue,
					StartDate:     req.StartDate,
					EndDate:       req.EndDate,
					IsActive:      true,
				}, nil
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/admin/promotions", ctl.CreatePromotion)

		now := time.Now()
		reqBody := domain.CreatePromotionRequest{
			ProductID:     "prod-123",
			Name:          "Summer Sale",
			DiscountType:  "percentage",
			DiscountValue: 15.0,
			StartDate:     now,
			EndDate:       now.Add(24 * time.Hour),
			IsActive:      true,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/promotions", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("product not found error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			CreatePromotionFunc: func(ctx context.Context, req *domain.CreatePromotionRequest) (*domain.Promotion, error) {
				return nil, domain.ErrProductNotFound
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/admin/promotions", ctl.CreatePromotion)

		now := time.Now()
		reqBody := domain.CreatePromotionRequest{
			ProductID:     "invalid-prod",
			Name:          "Summer Sale",
			DiscountType:  "percentage",
			DiscountValue: 15.0,
			StartDate:     now,
			EndDate:       now.Add(24 * time.Hour),
			IsActive:      true,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/promotions", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNotFound, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/admin/promotions", ctl.CreatePromotion)

		reqBody := domain.CreatePromotionRequest{
			ProductID:     "", // product_id is required
			Name:          "Summer Sale",
			DiscountType:  "percentage",
			DiscountValue: -5.0, // must be gt=0
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/promotions", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})
}

func TestPromotionVoucherController_ApplyVoucher(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success apply percentage discount", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			ApplyVoucherFunc: func(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
				return &domain.ApplyVoucherResponse{
					Valid:          true,
					DiscountAmount: 50000.0,
					VoucherID:      10,
				}, nil
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/vouchers/apply", func(c *gin.Context) {
			c.Set("user_id", 42) // Authenticated user ID in context
			ctl.ApplyVoucher(c)
		})

		reqBody := domain.ApplyVoucherRequest{
			Code:        "DISCOUNT50",
			OrderAmount: 200000.0,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/apply", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		is.Equal(true, data["valid"])
		is.Equal(50000.0, data["discount_amount"])
		is.Equal(float64(10), data["voucher_id"])
	})

	t.Run("expired voucher error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			ApplyVoucherFunc: func(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
				return nil, domain.ErrVoucherExpired
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/vouchers/apply", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.ApplyVoucher(c)
		})

		reqBody := domain.ApplyVoucherRequest{
			Code:        "EXPIREDVOUCHER",
			OrderAmount: 150000.0,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/apply", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})

	t.Run("global usage limit reached error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			ApplyVoucherFunc: func(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
				return nil, domain.ErrVoucherLimitReached
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/vouchers/apply", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.ApplyVoucher(c)
		})

		reqBody := domain.ApplyVoucherRequest{
			Code:        "LIMITEDVOUCHER",
			OrderAmount: 150000.0,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/apply", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})

	t.Run("minimum amount not met error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockPromotionVoucherUsecase{
			ApplyVoucherFunc: func(ctx context.Context, userID int, req *domain.ApplyVoucherRequest) (*domain.ApplyVoucherResponse, error) {
				return nil, domain.ErrVoucherMinAmountNotMet
			},
		}

		ctl := controller.NewPromotionVoucherController(mockUC)
		r := gin.New()
		r.POST("/api/v1/vouchers/apply", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.ApplyVoucher(c)
		})

		reqBody := domain.ApplyVoucherRequest{
			Code:        "MIN500K",
			OrderAmount: 100000.0, // Less than minimum amount of 500k
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/apply", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})
}
