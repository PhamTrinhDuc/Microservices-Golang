package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKeyType struct{}
var txCtxKey = txKeyType{}

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) getExecutor(ctx context.Context) DBExecutor {
	if tx, ok := ctx.Value(txCtxKey).(pgx.Tx); ok {
		return tx
	}
	return r.db
}

// WithTransaction executes the function fn inside a database transaction.
func (r *OrderRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		}
	}()

	txCtx := context.WithValue(ctx, txCtxKey, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// --- Order CRUD Operations ---

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	executor := r.getExecutor(ctx)
	query := `
		INSERT INTO orders (
			order_code, user_id, store_id, voucher_id, order_status_id, payment_status_id, shipping_status_id,
			total_amount, voucher_discount, shipping_price, payment_method, payment_code, payos_order_code, note,
			receiver_name, receiver_address, receiver_phone, sender_name, sender_address, sender_phone,
			shipping_provider, shipping_code, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := executor.QueryRow(ctx, query,
		order.OrderCode,
		order.UserID,
		order.StoreID,
		order.VoucherID,
		order.OrderStatusID,
		order.PaymentStatusID,
		order.ShippingStatusID,
		order.TotalAmount,
		order.VoucherDiscount,
		order.ShippingPrice,
		order.PaymentMethod,
		order.PaymentCode,
		order.PayosOrderCode,
		order.Note,
		order.ReceiverName,
		order.ReceiverAddress,
		order.ReceiverPhone,
		order.SenderName,
		order.SenderAddress,
		order.SenderPhone,
		order.ShippingProvider,
		order.ShippingCode,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) CreateOrderDetail(ctx context.Context, detail *domain.OrderDetail) (*domain.OrderDetail, error) {
	executor := r.getExecutor(ctx)
	query := `
		INSERT INTO order_details (order_id, variant_id, quantity, unit_price, total_cost)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := executor.QueryRow(ctx, query,
		detail.OrderID,
		detail.VariantID,
		detail.Quantity,
		detail.UnitPrice,
		detail.TotalCost,
	).Scan(&detail.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create order detail: %w", err)
	}
	return detail, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id int) (*domain.Order, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			id, order_code, user_id, store_id, voucher_id, order_status_id, payment_status_id, shipping_status_id,
			total_amount, voucher_discount, shipping_price, payment_method, payment_code, payos_order_code, note,
			receiver_name, receiver_address, receiver_phone, sender_name, sender_address, sender_phone,
			shipping_provider, shipping_code, created_at, updated_at
		FROM orders
		WHERE id = $1`

	o := &domain.Order{}
	err := executor.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.OrderCode, &o.UserID, &o.StoreID, &o.VoucherID, &o.OrderStatusID, &o.PaymentStatusID, &o.ShippingStatusID,
		&o.TotalAmount, &o.VoucherDiscount, &o.ShippingPrice, &o.PaymentMethod, &o.PaymentCode, &o.PayosOrderCode, &o.Note,
		&o.ReceiverName, &o.ReceiverAddress, &o.ReceiverPhone, &o.SenderName, &o.SenderAddress, &o.SenderPhone,
		&o.ShippingProvider, &o.ShippingCode, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) GetOrderByIDForUpdate(ctx context.Context, id int) (*domain.Order, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			id, order_code, user_id, store_id, voucher_id, order_status_id, payment_status_id, shipping_status_id,
			total_amount, voucher_discount, shipping_price, payment_method, payment_code, payos_order_code, note,
			receiver_name, receiver_address, receiver_phone, sender_name, sender_address, sender_phone,
			shipping_provider, shipping_code, created_at, updated_at
		FROM orders
		WHERE id = $1 FOR UPDATE`

	o := &domain.Order{}
	err := executor.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.OrderCode, &o.UserID, &o.StoreID, &o.VoucherID, &o.OrderStatusID, &o.PaymentStatusID, &o.ShippingStatusID,
		&o.TotalAmount, &o.VoucherDiscount, &o.ShippingPrice, &o.PaymentMethod, &o.PaymentCode, &o.PayosOrderCode, &o.Note,
		&o.ReceiverName, &o.ReceiverAddress, &o.ReceiverPhone, &o.SenderName, &o.SenderAddress, &o.SenderPhone,
		&o.ShippingProvider, &o.ShippingCode, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by id for update: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) GetOrderByPaymentRefForUpdate(ctx context.Context, payosOrderCode string) (*domain.Order, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			id, order_code, user_id, store_id, voucher_id, order_status_id, payment_status_id, shipping_status_id,
			total_amount, voucher_discount, shipping_price, payment_method, payment_code, payos_order_code, note,
			receiver_name, receiver_address, receiver_phone, sender_name, sender_address, sender_phone,
			shipping_provider, shipping_code, created_at, updated_at
		FROM orders
		WHERE payos_order_code = $1 FOR UPDATE`

	o := &domain.Order{}
	err := executor.QueryRow(ctx, query, payosOrderCode).Scan(
		&o.ID, &o.OrderCode, &o.UserID, &o.StoreID, &o.VoucherID, &o.OrderStatusID, &o.PaymentStatusID, &o.ShippingStatusID,
		&o.TotalAmount, &o.VoucherDiscount, &o.ShippingPrice, &o.PaymentMethod, &o.PaymentCode, &o.PayosOrderCode, &o.Note,
		&o.ReceiverName, &o.ReceiverAddress, &o.ReceiverPhone, &o.SenderName, &o.SenderAddress, &o.SenderPhone,
		&o.ShippingProvider, &o.ShippingCode, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by payment ref: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) GetOrderDetails(ctx context.Context, orderID int) ([]domain.OrderDetailResponse, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			od.id, od.variant_id, pv.name, pv.sku, od.quantity, od.unit_price, od.total_cost
		FROM order_details od
		JOIN product_variant pv ON od.variant_id = pv.id
		WHERE od.order_id = $1
		ORDER BY od.id ASC`

	rows, err := executor.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}
	defer rows.Close()

	var details []domain.OrderDetailResponse
	for rows.Next() {
		var d domain.OrderDetailResponse
		err := rows.Scan(&d.ID, &d.VariantID, &d.VariantName, &d.SKU, &d.Quantity, &d.UnitPrice, &d.TotalCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order detail response: %w", err)
		}
		details = append(details, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return details, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int) ([]*domain.Order, int, error) {
	executor := r.getExecutor(ctx)
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	var args []any
	argCount := 1

	if userID != nil {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *userID)
		argCount++
	}
	if storeID != nil {
		whereClause += fmt.Sprintf(" AND store_id = $%d", argCount)
		args = append(args, *storeID)
		argCount++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", whereClause)
	var totalCount int
	err := executor.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get paginated orders
	selectQuery := fmt.Sprintf(`
		SELECT 
			id, order_code, user_id, store_id, voucher_id, order_status_id, payment_status_id, shipping_status_id,
			total_amount, voucher_discount, shipping_price, payment_method, payment_code, payos_order_code, note,
			receiver_name, receiver_address, receiver_phone, sender_name, sender_address, sender_phone,
			shipping_provider, shipping_code, created_at, updated_at
		FROM orders
		%s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)
	rows, err := executor.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		err := rows.Scan(
			&o.ID, &o.OrderCode, &o.UserID, &o.StoreID, &o.VoucherID, &o.OrderStatusID, &o.PaymentStatusID, &o.ShippingStatusID,
			&o.TotalAmount, &o.VoucherDiscount, &o.ShippingPrice, &o.PaymentMethod, &o.PaymentCode, &o.PayosOrderCode, &o.Note,
			&o.ReceiverName, &o.ReceiverAddress, &o.ReceiverPhone, &o.SenderName, &o.SenderAddress, &o.SenderPhone,
			&o.ShippingProvider, &o.ShippingCode, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return orders, totalCount, nil
}

func (r *OrderRepository) UpdateOrderStatuses(ctx context.Context, id int, orderStatusID int, paymentStatusID int, shippingStatusID int) error {
	executor := r.getExecutor(ctx)
	query := `
		UPDATE orders 
		SET order_status_id = $1, payment_status_id = $2, shipping_status_id = $3, updated_at = NOW() 
		WHERE id = $4`
	_, err := executor.Exec(ctx, query, orderStatusID, paymentStatusID, shippingStatusID, id)
	if err != nil {
		return fmt.Errorf("failed to update order statuses: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateOrderShippingInfo(ctx context.Context, id int, provider string, code string) error {
	executor := r.getExecutor(ctx)
	query := `
		UPDATE orders 
		SET shipping_provider = $1, shipping_code = $2, updated_at = NOW() 
		WHERE id = $3`
	_, err := executor.Exec(ctx, query, provider, code, id)
	if err != nil {
		return fmt.Errorf("failed to update order shipping info: %w", err)
	}
	return nil
}

// --- Order Status History Operations ---

func (r *OrderRepository) CreateOrderStatusHistory(ctx context.Context, history *domain.OrderStatusHistory) error {
	executor := r.getExecutor(ctx)
	query := `
		INSERT INTO order_status_history (order_id, status_type, from_status, to_status, changed_by, note, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, changed_at`
	err := executor.QueryRow(ctx, query,
		history.OrderID,
		history.StatusType,
		history.FromStatus,
		history.ToStatus,
		history.ChangedBy,
		history.Note,
	).Scan(&history.ID, &history.ChangedAt)
	if err != nil {
		return fmt.Errorf("failed to create order status history: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetStatusIDByCode(ctx context.Context, statusType string, code string) (int, error) {
	executor := r.getExecutor(ctx)
	var tableName string
	switch statusType {
	case "order":
		tableName = "order_status"
	case "payment":
		tableName = "payment_status"
	case "shipping":
		tableName = "shipping_status"
	default:
		return 0, fmt.Errorf("invalid status type: %s", statusType)
	}

	query := fmt.Sprintf("SELECT id FROM %s WHERE code = $1", tableName)
	var id int
	err := executor.QueryRow(ctx, query, code).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get status id for %s with code %s: %w", statusType, code, err)
	}
	return id, nil
}

func (r *OrderRepository) GetStatusLabelByID(ctx context.Context, statusType string, id int) (string, error) {
	executor := r.getExecutor(ctx)
	var tableName string
	switch statusType {
	case "order":
		tableName = "order_status"
	case "payment":
		tableName = "payment_status"
	case "shipping":
		tableName = "shipping_status"
	default:
		return "", fmt.Errorf("invalid status type: %s", statusType)
	}

	query := fmt.Sprintf("SELECT label FROM %s WHERE id = $1", tableName)
	var label string
	err := executor.QueryRow(ctx, query, id).Scan(&label)
	if err != nil {
		return "", fmt.Errorf("failed to get status label for %s with id %d: %w", statusType, id, err)
	}
	return label, nil
}

// --- Inventory Reservation Operations ---

func (r *OrderRepository) CreateReservation(ctx context.Context, res *domain.InventoryReservation) error {
	executor := r.getExecutor(ctx)
	itemsJSON, err := json.Marshal(res.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal reservation items: %w", err)
	}

	query := `
		INSERT INTO inventory_reservations (id, user_id, store_id, items, status, payment_code, payos_order_code, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at`
	err = executor.QueryRow(ctx, query,
		res.ID,
		res.UserID,
		res.StoreID,
		itemsJSON,
		res.Status,
		res.PaymentCode,
		res.PayosOrderCode,
		res.ExpiresAt,
	).Scan(&res.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateReservationStatus(ctx context.Context, id string, status string) error {
	executor := r.getExecutor(ctx)
	query := `UPDATE inventory_reservations SET status = $1 WHERE id = $2`
	_, err := executor.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetReservationByOrderID(ctx context.Context, orderID int) (*domain.InventoryReservation, error) {
	executor := r.getExecutor(ctx)
	// Reservations are mapped by payos_order_code and payment_code which are linked on orders
	query := `
		SELECT ir.id, ir.user_id, ir.store_id, ir.items, ir.status, ir.payment_code, ir.payos_order_code, ir.expires_at, ir.created_at
		FROM inventory_reservations ir
		JOIN orders o ON ir.payos_order_code = o.payos_order_code
		WHERE o.id = $1`

	res := &domain.InventoryReservation{}
	var itemsBytes []byte
	err := executor.QueryRow(ctx, query, orderID).Scan(
		&res.ID, &res.UserID, &res.StoreID, &itemsBytes, &res.Status, &res.PaymentCode, &res.PayosOrderCode, &res.ExpiresAt, &res.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("reservation not found for order id %d: %w", orderID, err)
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	if err := json.Unmarshal(itemsBytes, &res.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservation items: %w", err)
	}
	return res, nil
}

func (r *OrderRepository) GetExpiredPendingReservations(ctx context.Context) ([]*domain.InventoryReservation, error) {
	executor := r.getExecutor(ctx)
	// SKIP LOCKED handles concurrency safely across instances
	query := `
		SELECT id, user_id, store_id, items, status, payment_code, payos_order_code, expires_at, created_at
		FROM inventory_reservations
		WHERE status = 'pending' AND expires_at < NOW()
		FOR UPDATE SKIP LOCKED`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.InventoryReservation
	for rows.Next() {
		res := &domain.InventoryReservation{}
		var itemsBytes []byte
		err := rows.Scan(
			&res.ID, &res.UserID, &res.StoreID, &itemsBytes, &res.Status, &res.PaymentCode, &res.PayosOrderCode, &res.ExpiresAt, &res.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired reservation: %w", err)
		}
		if err := json.Unmarshal(itemsBytes, &res.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}
		reservations = append(reservations, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reservations, nil
}

// --- Transactional domain logic helper methods ---

func (r *OrderRepository) LockStock(ctx context.Context, variantID int, storeID int) (quantity int, reserved int, err error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT quantity, reserved 
		FROM product_inventory 
		WHERE variant_id = $1 AND store_id = $2 
		FOR UPDATE`

	err = executor.QueryRow(ctx, query, variantID, storeID).Scan(&quantity, &reserved)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, domain.ErrVariantNotFound
		}
		return 0, 0, fmt.Errorf("failed to lock product inventory: %w", err)
	}
	return quantity, reserved, nil
}

func (r *OrderRepository) UpdateReservedStock(ctx context.Context, variantID int, storeID int, change int) error {
	executor := r.getExecutor(ctx)
	query := `
		UPDATE product_inventory 
		SET reserved = reserved + $1, last_updated = NOW() 
		WHERE variant_id = $2 AND store_id = $3`

	_, err := executor.Exec(ctx, query, change, variantID, storeID)
	if err != nil {
		return fmt.Errorf("failed to update reserved stock: %w", err)
	}
	return nil
}

func (r *OrderRepository) DeductStock(ctx context.Context, variantID int, storeID int, quantity int) (qtyAfter int, err error) {
	executor := r.getExecutor(ctx)
	query := `
		UPDATE product_inventory 
		SET quantity = quantity - $1, last_updated = NOW() 
		WHERE variant_id = $2 AND store_id = $3 
		RETURNING quantity`

	err = executor.QueryRow(ctx, query, quantity, variantID, storeID).Scan(&qtyAfter)
	if err != nil {
		return 0, fmt.Errorf("failed to deduct stock: %w", err)
	}
	return qtyAfter, nil
}

func (r *OrderRepository) AddInventoryLog(ctx context.Context, variantID int, storeID int, quantityChange int, quantityAfter int, reason string, referenceID string) error {
	executor := r.getExecutor(ctx)
	query := `
		INSERT INTO inventory_log (variant_id, store_id, change_qty, qty_after, reason, ref_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := executor.Exec(ctx, query, variantID, storeID, quantityChange, quantityAfter, reason, referenceID)
	if err != nil {
		return fmt.Errorf("failed to write inventory log: %w", err)
	}
	return nil
}

func (r *OrderRepository) LockVoucherByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			id, code, name, start_date, end_date, discount_type, discount_value, discount_target, 
			min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE code = $1 AND is_deleted = false
		FOR UPDATE`

	v := &domain.Voucher{}
	err := executor.QueryRow(ctx, query, code).Scan(
		&v.ID, &v.Code, &v.Name, &v.StartDate, &v.EndDate, &v.DiscountType, &v.DiscountValue, &v.DiscountTarget,
		&v.MinOrderValue, &v.MaxDiscountAmount, &v.MaxUsageTotal, &v.MaxUsagePerUser, &v.UsedCount, &v.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, fmt.Errorf("failed to lock voucher by code: %w", err)
	}
	return v, nil
}

func (r *OrderRepository) LockVoucherByID(ctx context.Context, voucherID int) (*domain.Voucher, error) {
	executor := r.getExecutor(ctx)
	query := `
		SELECT 
			id, code, name, start_date, end_date, discount_type, discount_value, discount_target, 
			min_order_value, max_discount_amount, max_usage_total, max_usage_per_user, used_count, is_deleted
		FROM vouchers
		WHERE id = $1 AND is_deleted = false
		FOR UPDATE`

	v := &domain.Voucher{}
	err := executor.QueryRow(ctx, query, voucherID).Scan(
		&v.ID, &v.Code, &v.Name, &v.StartDate, &v.EndDate, &v.DiscountType, &v.DiscountValue, &v.DiscountTarget,
		&v.MinOrderValue, &v.MaxDiscountAmount, &v.MaxUsageTotal, &v.MaxUsagePerUser, &v.UsedCount, &v.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, fmt.Errorf("failed to lock voucher by id: %w", err)
	}
	return v, nil
}

func (r *OrderRepository) IncrementVoucherUsedCount(ctx context.Context, voucherID int, amount int) error {
	executor := r.getExecutor(ctx)
	query := `UPDATE vouchers SET used_count = used_count + $1 WHERE id = $2`
	_, err := executor.Exec(ctx, query, amount, voucherID)
	if err != nil {
		return fmt.Errorf("failed to increment voucher used count: %w", err)
	}
	return nil
}

func (r *OrderRepository) RecordVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error {
	executor := r.getExecutor(ctx)
	query := `
		INSERT INTO voucher_usages (voucher_id, user_id, order_id, used_at)
		VALUES ($1, $2, $3, NOW())`
	_, err := executor.Exec(ctx, query, voucherID, userID, orderID)
	if err != nil {
		return fmt.Errorf("failed to record voucher usage: %w", err)
	}
	return nil
}

func (r *OrderRepository) DeleteVoucherUsage(ctx context.Context, voucherID int, userID int, orderID int) error {
	executor := r.getExecutor(ctx)
	query := `
		DELETE FROM voucher_usages 
		WHERE voucher_id = $1 AND user_id = $2 AND order_id = $3`
	_, err := executor.Exec(ctx, query, voucherID, userID, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete voucher usage: %w", err)
	}
	return nil
}
