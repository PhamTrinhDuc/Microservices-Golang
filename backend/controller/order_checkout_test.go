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
	"backend/internal/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockOrderUsecase struct {
	CheckoutOrderFunc             func(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error)
	ConfirmPaymentFunc            func(ctx context.Context, payosOrderCode string, paymentCode string) error
	CancelExpiredReservationsFunc func(ctx context.Context) error
	ListOrdersFunc                func(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*domain.OrderResponse, int, error)
	GetOrderDetailsFunc           func(ctx context.Context, orderID int, userID *int) (*domain.OrderResponse, error)
	CancelOrderFunc               func(ctx context.Context, orderID int, actorUserID int, isAdmin bool, note string) error
	UpdateOrderStatusFunc         func(ctx context.Context, orderID int, actorUserID int, req *domain.UpdateOrderStatusRequest) error
}

func (m *mockOrderUsecase) CheckoutOrder(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error) {
	return m.CheckoutOrderFunc(ctx, userID, req)
}

func (m *mockOrderUsecase) ConfirmPayment(ctx context.Context, payosOrderCode string, paymentCode string) error {
	return m.ConfirmPaymentFunc(ctx, payosOrderCode, paymentCode)
}

func (m *mockOrderUsecase) CancelExpiredReservations(ctx context.Context) error {
	return m.CancelExpiredReservationsFunc(ctx)
}

func (m *mockOrderUsecase) ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*domain.OrderResponse, int, error) {
	return m.ListOrdersFunc(ctx, userID, storeID, page, limit)
}

func (m *mockOrderUsecase) GetOrderDetails(ctx context.Context, orderID int, userID *int) (*domain.OrderResponse, error) {
	return m.GetOrderDetailsFunc(ctx, orderID, userID)
}

func (m *mockOrderUsecase) CancelOrder(ctx context.Context, orderID int, actorUserID int, isAdmin bool, note string) error {
	return m.CancelOrderFunc(ctx, orderID, actorUserID, isAdmin, note)
}

func (m *mockOrderUsecase) UpdateOrderStatus(ctx context.Context, orderID int, actorUserID int, req *domain.UpdateOrderStatusRequest) error {
	return m.UpdateOrderStatusFunc(ctx, orderID, actorUserID, req)
}

type mockPayOSInterface struct {
	CreatePaymentLinkFunc      func(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error)
	VerifyWebhookSignatureFunc func(payload payment.PayOSWebhookPayload) bool
}

func (m *mockPayOSInterface) CreatePaymentLink(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error) {
	return m.CreatePaymentLinkFunc(orderCode, amount, description, returnURL, cancelURL)
}

func (m *mockPayOSInterface) VerifyWebhookSignature(payload payment.PayOSWebhookPayload) bool {
	return m.VerifyWebhookSignatureFunc(payload)
}

func TestOrderController_Checkout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success checkout", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			CheckoutOrderFunc: func(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error) {
				checkoutURL := "https://checkout.payos.vn/pay/123"
				return &domain.OrderResponse{
					Order: domain.Order{
						ID:            1,
						OrderCode:     "ORD-12345",
						UserID:        userID,
						StoreID:       req.StoreID,
						TotalAmount:   150000.0,
						PaymentMethod: req.PaymentMethod,
					},
					Items: []domain.OrderDetailResponse{
						{ID: 1, VariantID: 1, Quantity: 2, UnitPrice: 75000.0, TotalCost: 150000.0},
					},
					OrderStatusLabel:    "Chờ thanh toán",
					PaymentStatusLabel:  "Chưa thanh toán",
					ShippingStatusLabel: "Chưa giao hàng",
					CheckoutURL:         &checkoutURL,
				}, nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/checkout", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.Checkout(c)
		})

		reqBody := domain.CheckoutOrderRequest{
			StoreID:       1,
			PaymentMethod: "payos",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)

		var resp domain.OrderResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		is.NoError(err)
		is.Equal("ORD-12345", resp.Order.OrderCode)
		is.Equal(150000.0, resp.Order.TotalAmount)
		is.Equal("https://checkout.payos.vn/pay/123", *resp.CheckoutURL)
	})

	t.Run("unauthorized", func(t *testing.T) {
		is := assert.New(t)
		ctl := controller.NewOrderController(&mockOrderUsecase{}, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/checkout", ctl.Checkout)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusUnauthorized, w.Code)
	})

	t.Run("bad request body", func(t *testing.T) {
		is := assert.New(t)
		ctl := controller.NewOrderController(&mockOrderUsecase{}, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/checkout", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.Checkout(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader([]byte("{invalid-json}")))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error - domain error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			CheckoutOrderFunc: func(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error) {
				return nil, domain.ErrInsufficientStock
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/checkout", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.Checkout(c)
		})

		reqBody := domain.CheckoutOrderRequest{StoreID: 1, PaymentMethod: "payos"}
		jsonBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusBadRequest, w.Code)
		is.Contains(w.Body.String(), domain.ErrInsufficientStock.Error())
	})

	t.Run("usecase error - internal error", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			CheckoutOrderFunc: func(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error) {
				return nil, errors.New("db error")
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/checkout", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.Checkout(c)
		})

		reqBody := domain.CheckoutOrderRequest{StoreID: 1, PaymentMethod: "payos"}
		jsonBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusInternalServerError, w.Code)
	})
}

func TestOrderController_ListMyOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success list my orders", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			ListOrdersFunc: func(ctx context.Context, userID *int, storeID *int, page, limit int) ([]*domain.OrderResponse, int, error) {
				return []*domain.OrderResponse{
					{
						Order: domain.Order{ID: 1, OrderCode: "ORD-1", UserID: *userID},
					},
				}, 1, nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.GET("/api/v1/orders", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.ListMyOrders(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?page=2&limit=5", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		is.Equal(float64(1), resp["total"])
		is.Equal(float64(2), resp["page"])
		is.Equal(float64(5), resp["limit"])
	})
}

func TestOrderController_GetMyOrderDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success get details", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			GetOrderDetailsFunc: func(ctx context.Context, orderID int, userID *int) (*domain.OrderResponse, error) {
				return &domain.OrderResponse{
					Order: domain.Order{ID: orderID, OrderCode: "ORD-99", UserID: *userID},
				}, nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.GET("/api/v1/orders/:id", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.GetMyOrderDetails(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/99", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)

		var resp domain.OrderResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		is.Equal(99, resp.Order.ID)
	})

	t.Run("order not found", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			GetOrderDetailsFunc: func(ctx context.Context, orderID int, userID *int) (*domain.OrderResponse, error) {
				return nil, domain.ErrOrderNotFound
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.GET("/api/v1/orders/:id", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.GetMyOrderDetails(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/99", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusNotFound, w.Code)
	})
}

func TestOrderController_CancelMyOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success cancel", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			CancelOrderFunc: func(ctx context.Context, orderID, actorUserID int, isAdmin bool, note string) error {
				is.Equal(99, orderID)
				is.Equal(42, actorUserID)
				is.False(isAdmin)
				return nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.POST("/api/v1/orders/:id/cancel", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.CancelMyOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/99/cancel", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)
		is.Contains(w.Body.String(), "order cancelled successfully")
	})
}

func TestOrderController_ConfirmPaymentWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success webhook confirm PAID", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			ConfirmPaymentFunc: func(ctx context.Context, payosOrderCode string, paymentCode string) error {
				is.Equal("123456", payosOrderCode)
				is.Equal("PAY-REF-99", paymentCode)
				return nil
			},
		}

		mockPay := &mockPayOSInterface{
			VerifyWebhookSignatureFunc: func(payload payment.PayOSWebhookPayload) bool {
				return true
			},
		}

		ctl := controller.NewOrderController(mockUC, mockPay)
		r := gin.New()
		r.POST("/api/v1/payments/webhook", ctl.ConfirmPaymentWebhook)

		payload := payment.PayOSWebhookPayload{
			Code: "00",
			Desc: "Success",
			Data: map[string]interface{}{
				"status":        "PAID",
				"orderCode":     float64(123456),
				"paymentLinkId": "PAY-REF-99",
			},
			Signature: "valid_signature",
		}
		jsonBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)
		is.Contains(w.Body.String(), `"status":"success"`)
	})

	t.Run("invalid webhook signature", func(t *testing.T) {
		is := assert.New(t)
		mockPay := &mockPayOSInterface{
			VerifyWebhookSignatureFunc: func(payload payment.PayOSWebhookPayload) bool {
				return false
			},
		}

		ctl := controller.NewOrderController(&mockOrderUsecase{}, mockPay)
		r := gin.New()
		r.POST("/api/v1/payments/webhook", ctl.ConfirmPaymentWebhook)

		payload := payment.PayOSWebhookPayload{
			Code:      "00",
			Signature: "invalid_sig",
		}
		jsonBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusUnauthorized, w.Code)
	})

	t.Run("webhook status not PAID", func(t *testing.T) {
		is := assert.New(t)
		mockPay := &mockPayOSInterface{
			VerifyWebhookSignatureFunc: func(payload payment.PayOSWebhookPayload) bool {
				return true
			},
		}

		ctl := controller.NewOrderController(&mockOrderUsecase{}, mockPay)
		r := gin.New()
		r.POST("/api/v1/payments/webhook", ctl.ConfirmPaymentWebhook)

		payload := payment.PayOSWebhookPayload{
			Code: "00",
			Data: map[string]interface{}{
				"status": "PENDING",
			},
			Signature: "valid_signature",
		}
		jsonBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)
		is.Contains(w.Body.String(), `"status":"ignored"`)
	})
}

func TestOrderController_AdminListOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success admin list", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			ListOrdersFunc: func(ctx context.Context, userID *int, storeID *int, page, limit int) ([]*domain.OrderResponse, int, error) {
				is.Nil(userID)
				is.Equal(5, *storeID)
				return []*domain.OrderResponse{
					{Order: domain.Order{ID: 1, StoreID: 5}},
				}, 1, nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.GET("/api/v1/admin/orders", ctl.AdminListOrders)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/orders?store_id=5", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)
	})
}

func TestOrderController_AdminUpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success admin update", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockOrderUsecase{
			UpdateOrderStatusFunc: func(ctx context.Context, orderID, actorUserID int, req *domain.UpdateOrderStatusRequest) error {
				is.Equal(99, orderID)
				is.Equal(42, actorUserID)
				is.Equal("confirmed", *req.OrderStatusCode)
				return nil
			},
		}

		ctl := controller.NewOrderController(mockUC, &mockPayOSInterface{})
		r := gin.New()
		r.PUT("/api/v1/admin/orders/:id/status", func(c *gin.Context) {
			c.Set("user_id", 42)
			ctl.AdminUpdateStatus(c)
		})

		status := "confirmed"
		reqBody := domain.UpdateOrderStatusRequest{
			OrderStatusCode: &status,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/orders/99/status", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		is.Equal(http.StatusOK, w.Code)
	})
}
