package usecase_test

import (
	"context"
	"testing"
	"time"

	"backend/domain"
	"backend/internal/payment"
	"backend/usecase"

	"github.com/stretchr/testify/assert"
)

type mockOrderRepository struct {
	CreateOrderFunc                   func(ctx context.Context, order *domain.Order) (*domain.Order, error)
	CreateOrderDetailFunc             func(ctx context.Context, detail *domain.OrderDetail) (*domain.OrderDetail, error)
	GetOrderByIDFunc                  func(ctx context.Context, id int) (*domain.Order, error)
	GetOrderByIDForUpdateFunc         func(ctx context.Context, id int) (*domain.Order, error)
	GetOrderByPaymentRefForUpdateFunc func(ctx context.Context, payosOrderCode string) (*domain.Order, error)
	GetOrderDetailsFunc               func(ctx context.Context, orderID int) ([]domain.OrderDetailResponse, error)
	ListOrdersFunc                    func(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*domain.Order, int, error)
	UpdateOrderStatusesFunc           func(ctx context.Context, id int, orderStatusID int, paymentStatusID int, shippingStatusID int) error
	UpdateOrderShippingInfoFunc       func(ctx context.Context, id int, provider string, code string) error
	CreateOrderStatusHistoryFunc       func(ctx context.Context, history *domain.OrderStatusHistory) error
	GetStatusIDByCodeFunc             func(ctx context.Context, statusType string, code string) (int, error)
	GetStatusLabelByIDFunc             func(ctx context.Context, statusType string, id int) (string, error)
	CreateReservationFunc             func(ctx context.Context, res *domain.InventoryReservation) error
	UpdateReservationStatusFunc       func(ctx context.Context, id string, status string) error
	GetReservationByOrderIDFunc       func(ctx context.Context, orderID int) (*domain.InventoryReservation, error)
	GetExpiredPendingReservationsFunc func(ctx context.Context) ([]*domain.InventoryReservation, error)
	WithTransactionFunc               func(ctx context.Context, fn func(ctx context.Context) error) error
	LockStockFunc                     func(ctx context.Context, variantID int, storeID int) (quantity int, reserved int, err error)
	UpdateReservedStockFunc           func(ctx context.Context, variantID int, storeID int, change int) error
	DeductStockFunc                   func(ctx context.Context, variantID int, storeID int, quantity int) (qtyAfter int, err error)
	AddInventoryLogFunc               func(ctx context.Context, variantID int, storeID int, quantityChange int, quantityAfter int, reason string, referenceID string) error
	LockVoucherByCodeFunc             func(ctx context.Context, code string) (*domain.Voucher, error)
	IncrementVoucherUsedCountFunc     func(ctx context.Context, voucherID int, amount int) error
	RecordVoucherUsageFunc             func(ctx context.Context, voucherID int, userID int, orderID int) error
	DeleteVoucherUsageFunc             func(ctx context.Context, voucherID int, userID int, orderID int) error
	LockVoucherByIDFunc               func(ctx context.Context, voucherID int) (*domain.Voucher, error)
	CountUserVoucherUsagesFunc        func(ctx context.Context, voucherID int, userID int) (int, error)
}

func (m *mockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return m.CreateOrderFunc(ctx, order)
}
func (m *mockOrderRepository) CreateOrderDetail(ctx context.Context, detail *domain.OrderDetail) (*domain.OrderDetail, error) {
	return m.CreateOrderDetailFunc(ctx, detail)
}
func (m *mockOrderRepository) GetOrderByID(ctx context.Context, id int) (*domain.Order, error) {
	return m.GetOrderByIDFunc(ctx, id)
}
func (m *mockOrderRepository) GetOrderByIDForUpdate(ctx context.Context, id int) (*domain.Order, error) {
	return m.GetOrderByIDForUpdateFunc(ctx, id)
}
func (m *mockOrderRepository) GetOrderByPaymentRefForUpdate(ctx context.Context, payosOrderCode string) (*domain.Order, error) {
	return m.GetOrderByPaymentRefForUpdateFunc(ctx, payosOrderCode)
}
func (m *mockOrderRepository) GetOrderDetails(ctx context.Context, orderID int) ([]domain.OrderDetailResponse, error) {
	return m.GetOrderDetailsFunc(ctx, orderID)
}
func (m *mockOrderRepository) ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*domain.Order, int, error) {
	return m.ListOrdersFunc(ctx, userID, storeID, page, limit)
}
func (m *mockOrderRepository) UpdateOrderStatuses(ctx context.Context, id int, orderStatusID int, paymentStatusID int, shippingStatusID int) error {
	return m.UpdateOrderStatusesFunc(ctx, id, orderStatusID, paymentStatusID, shippingStatusID)
}
func (m *mockOrderRepository) UpdateOrderShippingInfo(ctx context.Context, id int, provider string, code string) error {
	return m.UpdateOrderShippingInfoFunc(ctx, id, provider, code)
}
func (m *mockOrderRepository) CreateOrderStatusHistory(ctx context.Context, history *domain.OrderStatusHistory) error {
	return m.CreateOrderStatusHistoryFunc(ctx, history)
}
func (m *mockOrderRepository) GetStatusIDByCode(ctx context.Context, statusType string, code string) (int, error) {
	return m.GetStatusIDByCodeFunc(ctx, statusType, code)
}
func (m *mockOrderRepository) GetStatusLabelByID(ctx context.Context, statusType string, id int) (string, error) {
	return m.GetStatusLabelByIDFunc(ctx, statusType, id)
}
func (m *mockOrderRepository) CreateReservation(ctx context.Context, res *domain.InventoryReservation) error {
	return m.CreateReservationFunc(ctx, res)
}
func (m *mockOrderRepository) UpdateReservationStatus(ctx context.Context, id string, status string) error {
	return m.UpdateReservationStatusFunc(ctx, id, status)
}
func (m *mockOrderRepository) GetReservationByOrderID(ctx context.Context, orderID int) (*domain.InventoryReservation, error) {
	return m.GetReservationByOrderIDFunc(ctx, orderID)
}
func (m *mockOrderRepository) GetExpiredPendingReservations(ctx context.Context) ([]*domain.InventoryReservation, error) {
	return m.GetExpiredPendingReservationsFunc(ctx)
}
func (m *mockOrderRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.WithTransactionFunc != nil {
		return m.WithTransactionFunc(ctx, fn)
	}
	return fn(ctx) // execute without tx by default
}
func (m *mockOrderRepository) LockStock(ctx context.Context, variantID int, storeID int) (quantity int, reserved int, err error) {
	return m.LockStockFunc(ctx, variantID, storeID)
}
func (m *mockOrderRepository) UpdateReservedStock(ctx context.Context, variantID int, storeID int, change int) error {
	return m.UpdateReservedStockFunc(ctx, variantID, storeID, change)
}
func (m *mockOrderRepository) DeductStock(ctx context.Context, variantID int, storeID int, quantity int) (qtyAfter int, err error) {
	return m.DeductStockFunc(ctx, variantID, storeID, quantity)
}
func (m *mockOrderRepository) AddInventoryLog(ctx context.Context, variantID int, storeID int, quantityChange int, quantityAfter int, reason string, referenceID string) error {
	return m.AddInventoryLogFunc(ctx, variantID, storeID, quantityChange, quantityAfter, reason, referenceID)
}
func (m *mockOrderRepository) LockVoucherByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	return m.LockVoucherByCodeFunc(ctx, code)
}
func (m *mockOrderRepository) IncrementVoucherUsedCount(ctx context.Context, voucherID int, amount int) error {
	return m.IncrementVoucherUsedCountFunc(ctx, voucherID, amount)
}
func (m *mockOrderRepository) RecordVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error {
	return m.RecordVoucherUsageFunc(ctx, voucherID, userID, orderID)
}
func (m *mockOrderRepository) DeleteVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error {
	return m.DeleteVoucherUsageFunc(ctx, voucherID, userID, orderID)
}
func (m *mockOrderRepository) LockVoucherByID(ctx context.Context, voucherID int) (*domain.Voucher, error) {
	return m.LockVoucherByIDFunc(ctx, voucherID)
}
func (m *mockOrderRepository) CountUserVoucherUsages(ctx context.Context, voucherID int, userID int) (int, error) {
	return m.CountUserVoucherUsagesFunc(ctx, voucherID, userID)
}

type mockCartRepository struct {
	ListCartItemsFunc          func(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItem, error)
	FindCartItemFunc           func(ctx context.Context, userID *int, sessionID *string, variantID int) (*domain.CartItem, error)
	GetCartItemByIDFunc        func(ctx context.Context, id int) (*domain.CartItem, error)
	CreateCartItemFunc         func(ctx context.Context, item *domain.CartItem) (*domain.CartItem, error)
	UpdateCartItemQuantityFunc func(ctx context.Context, id int, quantity int) (*domain.CartItem, error)
	LinkGuestItemToUserFunc    func(ctx context.Context, id int, userID int) error
	DeleteCartItemFunc         func(ctx context.Context, id int) error
	ClearCartFunc              func(ctx context.Context, userID *int, sessionID *string) error
	GetCartDetailsFunc         func(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error)
	VerifyVariantExistsFunc    func(ctx context.Context, id int) (bool, error)
}

func (m *mockCartRepository) ListCartItems(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItem, error) {
	return m.ListCartItemsFunc(ctx, userID, sessionID)
}
func (m *mockCartRepository) FindCartItem(ctx context.Context, userID *int, sessionID *string, variantID int) (*domain.CartItem, error) {
	return m.FindCartItemFunc(ctx, userID, sessionID, variantID)
}
func (m *mockCartRepository) GetCartItemByID(ctx context.Context, id int) (*domain.CartItem, error) {
	return m.GetCartItemByIDFunc(ctx, id)
}
func (m *mockCartRepository) CreateCartItem(ctx context.Context, item *domain.CartItem) (*domain.CartItem, error) {
	return m.CreateCartItemFunc(ctx, item)
}
func (m *mockCartRepository) UpdateCartItemQuantity(ctx context.Context, id int, quantity int) (*domain.CartItem, error) {
	return m.UpdateCartItemQuantityFunc(ctx, id, quantity)
}
func (m *mockCartRepository) LinkGuestItemToUser(ctx context.Context, id int, userID int) error {
	return m.LinkGuestItemToUserFunc(ctx, id, userID)
}
func (m *mockCartRepository) DeleteCartItem(ctx context.Context, id int) error {
	return m.DeleteCartItemFunc(ctx, id)
}
func (m *mockCartRepository) ClearCart(ctx context.Context, userID *int, sessionID *string) error {
	return m.ClearCartFunc(ctx, userID, sessionID)
}
func (m *mockCartRepository) GetCartDetails(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error) {
	return m.GetCartDetailsFunc(ctx, userID, sessionID)
}
func (m *mockCartRepository) VerifyVariantExists(ctx context.Context, id int) (bool, error) {
	return m.VerifyVariantExistsFunc(ctx, id)
}

type mockAddressRepository struct {
	CreateFunc       func(ctx context.Context, a *domain.Address) (*domain.Address, error)
	GetByIDFunc      func(ctx context.Context, id int) (*domain.Address, error)
	ListByUserIDFunc func(ctx context.Context, userID int) ([]*domain.Address, error)
	UpdateFunc       func(ctx context.Context, a *domain.Address) (*domain.Address, error)
	SetDefaultFunc   func(ctx context.Context, userID, addressID int) error
	DeleteFunc       func(ctx context.Context, id int) error
}

func (m *mockAddressRepository) Create(ctx context.Context, a *domain.Address) (*domain.Address, error) {
	return m.CreateFunc(ctx, a)
}
func (m *mockAddressRepository) GetByID(ctx context.Context, id int) (*domain.Address, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockAddressRepository) ListByUserID(ctx context.Context, userID int) ([]*domain.Address, error) {
	return m.ListByUserIDFunc(ctx, userID)
}
func (m *mockAddressRepository) Update(ctx context.Context, a *domain.Address) (*domain.Address, error) {
	return m.UpdateFunc(ctx, a)
}
func (m *mockAddressRepository) SetDefault(ctx context.Context, userID, addressID int) error {
	return m.SetDefaultFunc(ctx, userID, addressID)
}
func (m *mockAddressRepository) Delete(ctx context.Context, id int) error {
	return m.DeleteFunc(ctx, id)
}

type mockPayOSClient struct {
	CreatePaymentLinkFunc      func(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error)
	VerifyWebhookSignatureFunc func(payload payment.PayOSWebhookPayload) bool
}

func (m *mockPayOSClient) CreatePaymentLink(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error) {
	return m.CreatePaymentLinkFunc(orderCode, amount, description, returnURL, cancelURL)
}
func (m *mockPayOSClient) VerifyWebhookSignature(payload payment.PayOSWebhookPayload) bool {
	return m.VerifyWebhookSignatureFunc(payload)
}

type mockGHNClient struct {
	CalculateShippingFeeFunc func(toDistrictID int, toWardCode string, weight float64) (float64, error)
	CreateShippingOrderFunc  func(order *domain.Order, items []domain.OrderDetailResponse) (string, error)
}

func (m *mockGHNClient) CalculateShippingFee(toDistrictID int, toWardCode string, weight float64) (float64, error) {
	return m.CalculateShippingFeeFunc(toDistrictID, toWardCode, weight)
}
func (m *mockGHNClient) CreateShippingOrder(order *domain.Order, items []domain.OrderDetailResponse) (string, error) {
	return m.CreateShippingOrderFunc(order, items)
}

func TestOrderUsecase_CheckoutOrder_COD_Success(t *testing.T) {
	is := assert.New(t)

	userID := 42
	storeID := 1
	qty := 2
	price := 50000.0

	mockCart := &mockCartRepository{
		GetCartDetailsFunc: func(ctx context.Context, u *int, s *string) ([]*domain.CartItemResponse, error) {
			return []*domain.CartItemResponse{
				{ID: 1, VariantID: 10, VariantName: "Red Shirt", SKU: "RED-SHIRT-M", Price: price, Quantity: qty},
			}, nil
		},
		ClearCartFunc: func(ctx context.Context, u *int, s *string) error {
			return nil
		},
	}

	receiverName := "Trinh Duc"
	receiverAddress := "123 Le Loi, District 1, HCMC"
	receiverPhone := "0987654321"

	mockOrder := &mockOrderRepository{
		LockStockFunc: func(ctx context.Context, variantID, sID int) (int, int, error) {
			is.Equal(10, variantID)
			is.Equal(storeID, sID)
			return 10, 0, nil // 10 in stock, 0 reserved
		},
		DeductStockFunc: func(ctx context.Context, variantID, sID, quantity int) (int, error) {
			is.Equal(10, variantID)
			is.Equal(storeID, sID)
			is.Equal(qty, quantity)
			return 8, nil // 8 in stock after deduction
		},
		AddInventoryLogFunc: func(ctx context.Context, variantID, sID, change, after int, reason, ref string) error {
			is.Equal(10, variantID)
			is.Equal(storeID, sID)
			is.Equal(-qty, change)
			is.Equal(8, after)
			is.Equal("order_confirmed", reason)
			return nil
		},
		GetStatusIDByCodeFunc: func(ctx context.Context, statusType, code string) (int, error) {
			return 100, nil
		},
		CreateOrderFunc: func(ctx context.Context, o *domain.Order) (*domain.Order, error) {
			o.ID = 50
			return o, nil
		},
		CreateOrderDetailFunc: func(ctx context.Context, d *domain.OrderDetail) (*domain.OrderDetail, error) {
			d.ID = 501
			return d, nil
		},
		CreateOrderStatusHistoryFunc: func(ctx context.Context, h *domain.OrderStatusHistory) error {
			return nil
		},
		GetStatusLabelByIDFunc: func(ctx context.Context, statusType string, id int) (string, error) {
			return "MockStatus", nil
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, mockCart, &mockAddressRepository{}, &mockPayOSClient{}, &mockGHNClient{})

	req := &domain.CheckoutOrderRequest{
		StoreID:         storeID,
		PaymentMethod:   "cod",
		ReceiverName:    &receiverName,
		ReceiverAddress: &receiverAddress,
		ReceiverPhone:   &receiverPhone,
	}

	resp, err := uc.CheckoutOrder(context.Background(), userID, req)
	is.NoError(err)
	is.NotNil(resp)
	is.Equal(50, resp.Order.ID)
	is.Equal("cod", resp.Order.PaymentMethod)
	is.Equal(130000.0, resp.Order.TotalAmount) // subtotal 100k + shipping 30k
	is.Nil(resp.CheckoutURL)
}

func TestOrderUsecase_CheckoutOrder_PayOS_Success(t *testing.T) {
	is := assert.New(t)

	userID := 42
	storeID := 1
	qty := 2
	price := 50000.0

	mockCart := &mockCartRepository{
		GetCartDetailsFunc: func(ctx context.Context, u *int, s *string) ([]*domain.CartItemResponse, error) {
			return []*domain.CartItemResponse{
				{ID: 1, VariantID: 10, VariantName: "Red Shirt", SKU: "RED-SHIRT-M", Price: price, Quantity: qty},
			}, nil
		},
		ClearCartFunc: func(ctx context.Context, u *int, s *string) error {
			return nil
		},
	}

	receiverName := "Trinh Duc"
	receiverAddress := "123 Le Loi, District 1, HCMC"
	receiverPhone := "0987654321"

	mockOrder := &mockOrderRepository{
		LockStockFunc: func(ctx context.Context, variantID, sID int) (int, int, error) {
			return 10, 0, nil
		},
		UpdateReservedStockFunc: func(ctx context.Context, variantID, sID, change int) error {
			is.Equal(10, variantID)
			is.Equal(storeID, sID)
			is.Equal(qty, change)
			return nil
		},
		GetStatusIDByCodeFunc: func(ctx context.Context, statusType, code string) (int, error) {
			return 100, nil
		},
		CreateOrderFunc: func(ctx context.Context, o *domain.Order) (*domain.Order, error) {
			o.ID = 50
			return o, nil
		},
		CreateOrderDetailFunc: func(ctx context.Context, d *domain.OrderDetail) (*domain.OrderDetail, error) {
			d.ID = 501
			return d, nil
		},
		CreateOrderStatusHistoryFunc: func(ctx context.Context, h *domain.OrderStatusHistory) error {
			return nil
		},
		GetStatusLabelByIDFunc: func(ctx context.Context, statusType string, id int) (string, error) {
			return "MockStatus", nil
		},
		CreateReservationFunc: func(ctx context.Context, res *domain.InventoryReservation) error {
			is.Equal("RES-LINK123", res.ID)
			is.Equal("pending", res.Status)
			return nil
		},
	}

	mockPay := &mockPayOSClient{
		CreatePaymentLinkFunc: func(orderCode int64, amount float64, description, returnURL, cancelURL string) (string, string, error) {
			return "https://checkout.payos.vn/pay/link123", "LINK123", nil
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, mockCart, &mockAddressRepository{}, mockPay, &mockGHNClient{})

	req := &domain.CheckoutOrderRequest{
		StoreID:         storeID,
		PaymentMethod:   "payos",
		ReceiverName:    &receiverName,
		ReceiverAddress: &receiverAddress,
		ReceiverPhone:   &receiverPhone,
	}

	resp, err := uc.CheckoutOrder(context.Background(), userID, req)
	is.NoError(err)
	is.NotNil(resp)
	is.Equal(50, resp.Order.ID)
	is.Equal("payos", resp.Order.PaymentMethod)
	is.Equal("https://checkout.payos.vn/pay/link123", *resp.CheckoutURL)
}

func TestOrderUsecase_CheckoutOrder_VoucherUserLimit(t *testing.T) {
	is := assert.New(t)

	userID := 42
	storeID := 1

	mockCart := &mockCartRepository{
		GetCartDetailsFunc: func(ctx context.Context, u *int, s *string) ([]*domain.CartItemResponse, error) {
			return []*domain.CartItemResponse{
				{ID: 1, VariantID: 10, VariantName: "Red Shirt", SKU: "RED-SHIRT-M", Price: 100000.0, Quantity: 1},
			}, nil
		},
	}

	voucherCode := "ONCEONLY"
	mockOrder := &mockOrderRepository{
		LockStockFunc: func(ctx context.Context, variantID, sID int) (int, int, error) {
			return 10, 0, nil
		},
		LockVoucherByCodeFunc: func(ctx context.Context, code string) (*domain.Voucher, error) {
			is.Equal(voucherCode, code)
			return &domain.Voucher{
				ID:              99,
				Code:            voucherCode,
				Name:            "Once Only Voucher",
				StartDate:       time.Now().Add(-1 * time.Hour),
				EndDate:         time.Now().Add(1 * time.Hour),
				DiscountType:    "flat",
				DiscountValue:   20000.0,
				MinOrderValue:   50000.0,
				MaxUsagePerUser: 1,
				UsedCount:       5,
			}, nil
		},
		CountUserVoucherUsagesFunc: func(ctx context.Context, voucherID, uID int) (int, error) {
			is.Equal(99, voucherID)
			is.Equal(userID, uID)
			return 1, nil // User has used this voucher once before
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, mockCart, &mockAddressRepository{}, &mockPayOSClient{}, &mockGHNClient{})

	receiverName := "Trinh Duc"
	receiverAddress := "123 Le Loi, District 1, HCMC"
	receiverPhone := "0987654321"

	req := &domain.CheckoutOrderRequest{
		StoreID:         storeID,
		PaymentMethod:   "cod",
		ReceiverName:    &receiverName,
		ReceiverAddress: &receiverAddress,
		ReceiverPhone:   &receiverPhone,
		VoucherCode:     &voucherCode,
	}

	resp, err := uc.CheckoutOrder(context.Background(), userID, req)
	is.Error(err)
	is.Nil(resp)
	is.Equal(domain.ErrVoucherUserLimitReached, err)
}

func TestOrderUsecase_ConfirmPayment_Success(t *testing.T) {
	is := assert.New(t)

	payosOrderCode := "123456789"
	paymentCode := "PAY-REF-99"

	mockOrder := &mockOrderRepository{
		GetOrderByPaymentRefForUpdateFunc: func(ctx context.Context, ref string) (*domain.Order, error) {
			is.Equal(payosOrderCode, ref)
			return &domain.Order{
				ID:              50,
				OrderCode:       "ORD-50",
				PaymentStatusID: 101, // unpaid
				StoreID:         1,
			}, nil
		},
		GetStatusIDByCodeFunc: func(ctx context.Context, statusType, code string) (int, error) {
			if statusType == "payment" && code == "paid" {
				return 200, nil // paidStatusID
			}
			return 100, nil // other status ID
		},
		GetReservationByOrderIDFunc: func(ctx context.Context, orderID int) (*domain.InventoryReservation, error) {
			is.Equal(50, orderID)
			return &domain.InventoryReservation{
				ID:      "RES-PAY-REF-99",
				StoreID: 1,
				Items: []domain.InventoryReservationItem{
					{VariantID: 10, Quantity: 2},
				},
			}, nil
		},
		UpdateReservationStatusFunc: func(ctx context.Context, resID, status string) error {
			is.Equal("RES-PAY-REF-99", resID)
			is.Equal("completed", status)
			return nil
		},
		LockStockFunc: func(ctx context.Context, variantID, storeID int) (int, int, error) {
			return 10, 2, nil
		},
		UpdateReservedStockFunc: func(ctx context.Context, variantID, storeID, change int) error {
			is.Equal(-2, change)
			return nil
		},
		DeductStockFunc: func(ctx context.Context, variantID, storeID, quantity int) (int, error) {
			is.Equal(2, quantity)
			return 8, nil
		},
		AddInventoryLogFunc: func(ctx context.Context, variantID, storeID, change, after int, reason, ref string) error {
			return nil
		},
		UpdateOrderStatusesFunc: func(ctx context.Context, orderID, orderStatus, payStatus, shipStatus int) error {
			is.Equal(50, orderID)
			is.Equal(200, payStatus) // paid status
			return nil
		},
		CreateOrderStatusHistoryFunc: func(ctx context.Context, h *domain.OrderStatusHistory) error {
			return nil
		},
		GetOrderDetailsFunc: func(ctx context.Context, orderID int) ([]domain.OrderDetailResponse, error) {
			return []domain.OrderDetailResponse{
				{VariantID: 10, Quantity: 2},
			}, nil
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, &mockCartRepository{}, &mockAddressRepository{}, &mockPayOSClient{}, &mockGHNClient{})

	err := uc.ConfirmPayment(context.Background(), payosOrderCode, paymentCode)
	is.NoError(err)
}

func TestOrderUsecase_ConfirmPayment_Idempotency(t *testing.T) {
	is := assert.New(t)

	payosOrderCode := "123456789"
	paymentCode := "PAY-REF-99"

	mockOrder := &mockOrderRepository{
		GetOrderByPaymentRefForUpdateFunc: func(ctx context.Context, ref string) (*domain.Order, error) {
			return &domain.Order{
				ID:              50,
				OrderCode:       "ORD-50",
				PaymentStatusID: 200, // already paid
			}, nil
		},
		GetStatusIDByCodeFunc: func(ctx context.Context, statusType, code string) (int, error) {
			if statusType == "payment" && code == "paid" {
				return 200, nil // paidStatusID
			}
			return 100, nil
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, &mockCartRepository{}, &mockAddressRepository{}, &mockPayOSClient{}, &mockGHNClient{})

	err := uc.ConfirmPayment(context.Background(), payosOrderCode, paymentCode)
	is.NoError(err) // Should immediately exit with nil error and not call repository updates again
}

func TestOrderUsecase_CancelExpiredReservations(t *testing.T) {
	is := assert.New(t)

	payosOrderCodeStr := "123456"
	mockOrder := &mockOrderRepository{
		GetExpiredPendingReservationsFunc: func(ctx context.Context) ([]*domain.InventoryReservation, error) {
			return []*domain.InventoryReservation{
				{
					ID:             "RES-EXPIRED",
					StoreID:        1,
					Status:         "pending",
					PayosOrderCode: &payosOrderCodeStr,
					Items: []domain.InventoryReservationItem{
						{VariantID: 10, Quantity: 2},
					},
				},
			}, nil
		},
		UpdateReservationStatusFunc: func(ctx context.Context, id, status string) error {
			is.Equal("RES-EXPIRED", id)
			is.Equal("expired", status)
			return nil
		},
		GetOrderByPaymentRefForUpdateFunc: func(ctx context.Context, ref string) (*domain.Order, error) {
			is.Equal("123456", ref)
			return &domain.Order{
				ID:        50,
				VoucherID: nil,
				UserID:    42,
			}, nil
		},
		GetStatusIDByCodeFunc: func(ctx context.Context, statusType, code string) (int, error) {
			return 999, nil
		},
		UpdateReservedStockFunc: func(ctx context.Context, variantID, storeID, change int) error {
			is.Equal(10, variantID)
			is.Equal(-2, change)
			return nil
		},
		UpdateOrderStatusesFunc: func(ctx context.Context, orderID, orderStatus, payStatus, shipStatus int) error {
			is.Equal(50, orderID)
			is.Equal(999, orderStatus) // cancelled
			return nil
		},
		CreateOrderStatusHistoryFunc: func(ctx context.Context, h *domain.OrderStatusHistory) error {
			return nil
		},
	}

	uc := usecase.NewOrderUsecase(mockOrder, &mockCartRepository{}, &mockAddressRepository{}, &mockPayOSClient{}, &mockGHNClient{})

	err := uc.CancelExpiredReservations(context.Background())
	is.NoError(err)
}
