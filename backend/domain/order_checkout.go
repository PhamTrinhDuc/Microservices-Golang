package domain

import (
	"context"
	"time"
)

// Models

type Order struct {
	ID               int        `json:"id" db:"id"`
	OrderCode        string     `json:"order_code" db:"order_code"`
	UserID           int        `json:"user_id" db:"user_id"`
	StoreID          int        `json:"store_id" db:"store_id"`
	VoucherID        *int       `json:"voucher_id,omitempty" db:"voucher_id"`
	OrderStatusID    int        `json:"order_status_id" db:"order_status_id"`
	PaymentStatusID  int        `json:"payment_status_id" db:"payment_status_id"`
	ShippingStatusID int        `json:"shipping_status_id" db:"shipping_status_id"`
	TotalAmount      float64    `json:"total_amount" db:"total_amount"`
	VoucherDiscount  float64    `json:"voucher_discount" db:"voucher_discount"`
	ShippingPrice    float64    `json:"shipping_price" db:"shipping_price"`
	PaymentMethod    string     `json:"payment_method" db:"payment_method"`
	PaymentCode      *string    `json:"payment_code,omitempty" db:"payment_code"`
	PayosOrderCode   *string    `json:"payos_order_code,omitempty" db:"payos_order_code"`
	Note             *string    `json:"note,omitempty" db:"note"`
	ReceiverName     string     `json:"receiver_name" db:"receiver_name"`
	ReceiverAddress  string     `json:"receiver_address" db:"receiver_address"`
	ReceiverPhone    string     `json:"receiver_phone" db:"receiver_phone"`
	SenderName       *string    `json:"sender_name,omitempty" db:"sender_name"`
	SenderAddress    *string    `json:"sender_address,omitempty" db:"sender_address"`
	SenderPhone      *string    `json:"sender_phone,omitempty" db:"sender_phone"`
	ShippingProvider *string    `json:"shipping_provider,omitempty" db:"shipping_provider"`
	ShippingCode     *string    `json:"shipping_code,omitempty" db:"shipping_code"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type OrderDetail struct {
	ID        int     `json:"id" db:"id"`
	OrderID   int     `json:"order_id" db:"order_id"`
	VariantID int     `json:"variant_id" db:"variant_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	UnitPrice float64 `json:"unit_price" db:"unit_price"`
	TotalCost float64 `json:"total_cost" db:"total_cost"`
}

type OrderStatusHistory struct {
	ID         int       `json:"id" db:"id"`
	OrderID    int       `json:"order_id" db:"order_id"`
	StatusType string    `json:"status_type" db:"status_type"` // "order", "payment", "shipping"
	FromStatus *string   `json:"from_status,omitempty" db:"from_status"`
	ToStatus   string    `json:"to_status" db:"to_status"`
	ChangedBy  *int      `json:"changed_by,omitempty" db:"changed_by"`
	Note       *string   `json:"note,omitempty" db:"note"`
	ChangedAt  time.Time `json:"changed_at" db:"changed_at"`
}

type InventoryReservationItem struct {
	VariantID int `json:"variant_id"`
	Quantity  int `json:"quantity"`
}

type InventoryReservation struct {
	ID             string                     `json:"id" db:"id"`
	UserID         int                        `json:"user_id" db:"user_id"`
	StoreID        int                        `json:"store_id" db:"store_id"`
	Items          []InventoryReservationItem `json:"items" db:"items"`
	Status         string                     `json:"status" db:"status"` // "pending", "completed", "expired"
	PaymentCode    *string                    `json:"payment_code,omitempty" db:"payment_code"`
	PayosOrderCode *string                    `json:"payos_order_code,omitempty" db:"payos_order_code"`
	ExpiresAt      time.Time                  `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time                  `json:"created_at" db:"created_at"`
}

// Request and Response DTOs

type CheckoutOrderRequest struct {
	StoreID          int     `json:"store_id" validate:"required"`
	AddressID        *int    `json:"address_id"`
	ReceiverName     *string `json:"receiver_name"`
	ReceiverAddress  *string `json:"receiver_address"`
	ReceiverPhone    *string `json:"receiver_phone"`
	VoucherCode      *string `json:"voucher_code"`
	PaymentMethod    string  `json:"payment_method" validate:"required,oneof=cod bank_transfer payos"`
	ShippingProvider *string `json:"shipping_provider"`
	Note             *string `json:"note"`
}

type UpdateOrderStatusRequest struct {
	OrderStatusCode   *string `json:"order_status_code"`
	PaymentStatusCode *string `json:"payment_status_code"`
	ShippingStatusCode *string `json:"shipping_status_code"`
	ShippingProvider  *string `json:"shipping_provider"`
	ShippingCode      *string `json:"shipping_code"`
	Note              *string `json:"note"`
}

type OrderDetailResponse struct {
	ID          int     `json:"id"`
	VariantID   int     `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalCost   float64 `json:"total_cost"`
}

type OrderResponse struct {
	Order              Order                 `json:"order"`
	Items              []OrderDetailResponse `json:"items"`
	OrderStatusLabel   string                `json:"order_status_label"`
	PaymentStatusLabel string                `json:"payment_status_label"`
	ShippingStatusLabel string               `json:"shipping_status_label"`
	CheckoutURL        *string               `json:"checkout_url,omitempty"`
}

// Interfaces

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CreateOrderDetail(ctx context.Context, detail *OrderDetail) (*OrderDetail, error)
	GetOrderByID(ctx context.Context, id int) (*Order, error)
	GetOrderByIDForUpdate(ctx context.Context, id int) (*Order, error)
	GetOrderByPaymentRefForUpdate(ctx context.Context, payosOrderCode string) (*Order, error)
	GetOrderDetails(ctx context.Context, orderID int) ([]OrderDetailResponse, error)
	ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*Order, int, error)
	UpdateOrderStatuses(ctx context.Context, id int, orderStatusID int, paymentStatusID int, shippingStatusID int) error
	UpdateOrderShippingInfo(ctx context.Context, id int, provider string, code string) error
	
	CreateOrderStatusHistory(ctx context.Context, history *OrderStatusHistory) error
	GetStatusIDByCode(ctx context.Context, statusType string, code string) (int, error)
	GetStatusLabelByID(ctx context.Context, statusType string, id int) (string, error)

	CreateReservation(ctx context.Context, res *InventoryReservation) error
	UpdateReservationStatus(ctx context.Context, id string, status string) error
	GetReservationByOrderID(ctx context.Context, orderID int) (*InventoryReservation, error)
	GetExpiredPendingReservations(ctx context.Context) ([]*InventoryReservation, error)

	// Context Transaction management helper for UseCase
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	
	// Transactional domain logic helper methods inside OrderRepository
	LockStock(ctx context.Context, variantID int, storeID int) (quantity int, reserved int, err error)
	UpdateReservedStock(ctx context.Context, variantID int, storeID int, change int) error
	DeductStock(ctx context.Context, variantID int, storeID int, quantity int) (qtyAfter int, err error)
	AddInventoryLog(ctx context.Context, variantID int, storeID int, quantityChange int, quantityAfter int, reason string, referenceID string) error
	LockVoucherByCode(ctx context.Context, code string) (*Voucher, error)
	IncrementVoucherUsedCount(ctx context.Context, voucherID int, amount int) error
	RecordVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error
	DeleteVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error
	LockVoucherByID(ctx context.Context, voucherID int) (*Voucher, error)
}

type OrderUsecase interface {
	CheckoutOrder(ctx context.Context, userID int, req *CheckoutOrderRequest) (*OrderResponse, error)
	ConfirmPayment(ctx context.Context, payosOrderCode string, paymentCode string) error
	CancelExpiredReservations(ctx context.Context) error
	ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*OrderResponse, int, error)
	GetOrderDetails(ctx context.Context, orderID int, userID *int) (*OrderResponse, error)
	CancelOrder(ctx context.Context, orderID int, actorUserID int, isAdmin bool, note string) error
	UpdateOrderStatus(ctx context.Context, orderID int, actorUserID int, req *UpdateOrderStatusRequest) error
}
