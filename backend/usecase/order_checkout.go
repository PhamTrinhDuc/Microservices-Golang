package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"backend/domain"
	"backend/internal/payment"
)

type PayOSInterface interface {
	CreatePaymentLink(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error)
	VerifyWebhookSignature(payload payment.PayOSWebhookPayload) bool
}

type GHNInterface interface {
	CalculateShippingFee(toDistrictID int, toWardCode string, weight float64) (float64, error)
	CreateShippingOrder(order *domain.Order, items []domain.OrderDetailResponse) (string, error)
}

type OrderUsecase struct {
	orderRepo   domain.OrderRepository
	cartRepo    domain.CartRepository
	addressRepo domain.AddressRepository
	payosClient PayOSInterface
	ghnClient   GHNInterface
}

func NewOrderUsecase(
	orderRepo domain.OrderRepository,
	cartRepo domain.CartRepository,
	addressRepo domain.AddressRepository,
	payosClient PayOSInterface,
	ghnClient GHNInterface,
) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		addressRepo: addressRepo,
		payosClient: payosClient,
		ghnClient:   ghnClient,
	}
}

func (u *OrderUsecase) CheckoutOrder(ctx context.Context, userID int, req *domain.CheckoutOrderRequest) (*domain.OrderResponse, error) {
	// 1. Basic validation
	if req.PaymentMethod != "cod" && req.PaymentMethod != "bank_transfer" && req.PaymentMethod != "payos" {
		return nil, domain.ErrInvalidPaymentMethod
	}

	var checkoutURL *string
	var orderResponse *domain.OrderResponse

	// Run checkout inside database transaction block
	err := u.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// 2. Fetch cart items
		cartItems, err := u.cartRepo.GetCartDetails(txCtx, &userID, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch cart details: %w", err)
		}
		if len(cartItems) == 0 {
			return domain.ErrEmptyCart
		}

		// 3. Resolve shipping address
		var receiverName, receiverAddress, receiverPhone string
		if req.AddressID != nil {
			addr, err := u.addressRepo.GetByID(txCtx, *req.AddressID)
			if err != nil {
				return err
			}
			if addr.UserID != userID {
				return domain.ErrAddressNotFound
			}
			receiverName = addr.FullName
			receiverPhone = addr.Phone
			receiverAddress = fmt.Sprintf("%s, %s, %s, %s", addr.DetailAddress, addr.Ward, addr.District, addr.Province)
		} else {
			if req.ReceiverName == nil || req.ReceiverAddress == nil || req.ReceiverPhone == nil ||
				*req.ReceiverName == "" || *req.ReceiverAddress == "" || *req.ReceiverPhone == "" {
				return domain.ErrInvalidAddress
			}
			receiverName = *req.ReceiverName
			receiverAddress = *req.ReceiverAddress
			receiverPhone = *req.ReceiverPhone
		}

		// 4. Calculate subtotal & check stock
		var subtotal float64
		for _, item := range cartItems {
			subtotal += item.SellPrice * float64(item.Quantity)

			// Lock stock and check availability
			quantity, reserved, err := u.orderRepo.LockStock(txCtx, item.VariantID, req.StoreID)
			if err != nil {
				return err
			}
			if quantity-reserved < item.Quantity {
				return domain.ErrInsufficientStock
			}
		}

		// 5. Calculate shipping fee (dynamic or fallback flat rate)
		shippingFee := 30000.0
		if subtotal >= 500000.0 {
			shippingFee = 0.0
		} else if req.AddressID != nil {
			// Optional: Try to parse district and ward from address for GHN calculation
			addr, _ := u.addressRepo.GetByID(txCtx, *req.AddressID)
			if addr != nil {
				distID, _ := strconv.Atoi(addr.District)
				if distID > 0 {
					dynFee, err := u.ghnClient.CalculateShippingFee(distID, addr.Ward, 500*float64(len(cartItems)))
					if err == nil && dynFee > 0 {
						shippingFee = dynFee
					}
				}
			}
		}

		// 6. Handle promotion/voucher lock & calculation
		var voucherID *int
		var discountAmount float64
		if req.VoucherCode != nil && *req.VoucherCode != "" {
			v, err := u.orderRepo.LockVoucherByCode(txCtx, *req.VoucherCode)
			if err != nil {
				return err
			}

			// Validate timeframe
			now := time.Now()
			if now.Before(v.StartDate) || now.After(v.EndDate) {
				return domain.ErrVoucherExpired
			}

			// Validate limits
			if v.MaxUsageTotal != nil && v.UsedCount >= *v.MaxUsageTotal {
				return domain.ErrVoucherLimitReached
			}

			// Validate user usage limit
			userUsages, err := u.orderRepo.CountUserVoucherUsages(txCtx, v.ID, userID)
			if err != nil {
				return err
			}
			if userUsages >= v.MaxUsagePerUser {
				return domain.ErrVoucherUserLimitReached
			}

			// Minimum order subtotal check
			if subtotal < v.MinOrderValue {
				return domain.ErrVoucherMinAmountNotMet
			}

			// Calculate discount
			if v.DiscountType == "percentage" {
				discountAmount = subtotal * (v.DiscountValue / 100.0)
				if v.MaxDiscountAmount != nil && discountAmount > *v.MaxDiscountAmount {
					discountAmount = *v.MaxDiscountAmount
				}
			} else {
				discountAmount = v.DiscountValue
			}

			if discountAmount > subtotal {
				discountAmount = subtotal
			}

			voucherID = &v.ID
		}

		totalAmount := subtotal + shippingFee - discountAmount
		if totalAmount < 0 {
			totalAmount = 0
		}

		// Get initial status IDs
		pendingStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "order", "pending")
		if err != nil {
			return err
		}
		unpaidStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "payment", "unpaid")
		if err != nil {
			return err
		}
		notShippedStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "shipping", "not_shipped")
		if err != nil {
			return err
		}

		// Generate unique codes
		payosOrderCode := int64(time.Now().UnixMilli()*100 + rand.Int63n(100))
		payosOrderCodeStr := strconv.FormatInt(payosOrderCode, 10)
		orderCode := fmt.Sprintf("ORD-%d", payosOrderCode)

		var payosLinkID *string
		var payosOrderCodePtr *string

		// 7. Inventory reservation / deduction and PayOS call
		if req.PaymentMethod == "payos" || req.PaymentMethod == "bank_transfer" {
			// Online flow: increment reserved stock
			for _, item := range cartItems {
				err := u.orderRepo.UpdateReservedStock(txCtx, item.VariantID, req.StoreID, item.Quantity)
				if err != nil {
					return err
				}
			}

			// Call PayOS outside DB, but inside the transaction block so we roll back if it fails
			desc := fmt.Sprintf("Thanh toan don hang %s", orderCode)
			retURL := "http://localhost:3000/payment/success"
			canURL := "http://localhost:3000/payment/cancel"
			if val := txCtx.Value("PAYOS_RETURN_URL"); val != nil {
				retURL = val.(string)
			}
			if val := txCtx.Value("PAYOS_CANCEL_URL"); val != nil {
				canURL = val.(string)
			}

			linkURL, linkID, err := u.payosClient.CreatePaymentLink(payosOrderCode, totalAmount, desc, retURL, canURL)
			if err != nil {
				return fmt.Errorf("payos link creation failed: %w", err)
			}

			checkoutURL = &linkURL
			payosLinkID = &linkID
			payosOrderCodePtr = &payosOrderCodeStr

			// Create reservation record
			resItems := make([]domain.InventoryReservationItem, len(cartItems))
			for i, item := range cartItems {
				resItems[i] = domain.InventoryReservationItem{
					VariantID: item.VariantID,
					Quantity:  item.Quantity,
				}
			}

			reservation := &domain.InventoryReservation{
				ID:             fmt.Sprintf("RES-%s", linkID),
				UserID:         userID,
				StoreID:        req.StoreID,
				Items:          resItems,
				Status:         "pending",
				PaymentCode:    payosLinkID,
				PayosOrderCode: payosOrderCodePtr,
				ExpiresAt:      time.Now().Add(15 * time.Minute),
			}
			err = u.orderRepo.CreateReservation(txCtx, reservation)
			if err != nil {
				return err
			}
		} else {
			// COD flow: deduct actual stock
			for _, item := range cartItems {
				qtyAfter, err := u.orderRepo.DeductStock(txCtx, item.VariantID, req.StoreID, item.Quantity)
				if err != nil {
					return err
				}
				err = u.orderRepo.AddInventoryLog(txCtx, item.VariantID, req.StoreID, -item.Quantity, qtyAfter, "order_confirmed", orderCode)
				if err != nil {
					return err
				}
			}
		}

		// 8. Create Order
		order := &domain.Order{
			OrderCode:        orderCode,
			UserID:           userID,
			StoreID:          req.StoreID,
			VoucherID:        voucherID,
			OrderStatusID:    pendingStatusID,
			PaymentStatusID:  unpaidStatusID,
			ShippingStatusID: notShippedStatusID,
			TotalAmount:      totalAmount,
			VoucherDiscount:  discountAmount,
			ShippingPrice:    shippingFee,
			PaymentMethod:    req.PaymentMethod,
			PaymentCode:      payosLinkID,
			PayosOrderCode:   payosOrderCodePtr,
			Note:             req.Note,
			ReceiverName:     receiverName,
			ReceiverAddress:  receiverAddress,
			ReceiverPhone:    receiverPhone,
			ShippingProvider: req.ShippingProvider,
		}

		order, err = u.orderRepo.CreateOrder(txCtx, order)
		if err != nil {
			return err
		}

		// Create Details
		var items []domain.OrderDetailResponse
		for _, item := range cartItems {
			detail := &domain.OrderDetail{
				OrderID:   order.ID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
				UnitPrice: item.SellPrice,
				TotalCost: item.SellPrice * float64(item.Quantity),
			}
			_, err = u.orderRepo.CreateOrderDetail(txCtx, detail)
			if err != nil {
				return err
			}
			items = append(items, domain.OrderDetailResponse{
				ID:          detail.ID,
				VariantID:   item.VariantID,
				VariantName: item.VariantName,
				SKU:         item.SKU,
				Quantity:    item.Quantity,
				UnitPrice:   item.SellPrice,
				TotalCost:   detail.TotalCost,
			})
		}

		// Write initial status history
		histOrder := &domain.OrderStatusHistory{
			OrderID:    order.ID,
			StatusType: "order",
			ToStatus:   "pending",
			ChangedBy:  &userID,
		}
		err = u.orderRepo.CreateOrderStatusHistory(txCtx, histOrder)
		if err != nil {
			return err
		}

		histPayment := &domain.OrderStatusHistory{
			OrderID:    order.ID,
			StatusType: "payment",
			ToStatus:   "unpaid",
			ChangedBy:  &userID,
		}
		err = u.orderRepo.CreateOrderStatusHistory(txCtx, histPayment)
		if err != nil {
			return err
		}

		// If voucher applied, record usage
		if voucherID != nil {
			err = u.orderRepo.IncrementVoucherUsedCount(txCtx, *voucherID, 1)
			if err != nil {
				return err
			}
			err = u.orderRepo.RecordVoucherUsage(txCtx, *voucherID, userID, order.ID)
			if err != nil {
				return err
			}
		}

		// 9. Clear cart
		err = u.cartRepo.ClearCart(txCtx, &userID, nil)
		if err != nil {
			return err
		}

		// Get Status Labels
		orderStatusLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "order", pendingStatusID)
		paymentStatusLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "payment", unpaidStatusID)
		shippingStatusLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "shipping", notShippedStatusID)

		orderResponse = &domain.OrderResponse{
			Order:               *order,
			Items:               items,
			OrderStatusLabel:    orderStatusLabel,
			PaymentStatusLabel:  paymentStatusLabel,
			ShippingStatusLabel: shippingStatusLabel,
			CheckoutURL:         checkoutURL,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return orderResponse, nil
}

func (u *OrderUsecase) ConfirmPayment(ctx context.Context, payosOrderCode string, paymentCode string) error {
	var orderToShip *domain.Order
	var detailsToShip []domain.OrderDetailResponse

	err := u.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Fetch order
		order, err := u.orderRepo.GetOrderByPaymentRefForUpdate(txCtx, payosOrderCode)
		if err != nil {
			return err
		}

		// 2. Fetch paid/unpaid status IDs
		paidStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "payment", "paid")
		if err != nil {
			return err
		}
		confirmedStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "order", "confirmed")
		if err != nil {
			return err
		}

		// Idempotency check: Already marked as paid
		if order.PaymentStatusID == paidStatusID {
			return nil // success (bypass duplicate webhook processing)
		}

		// 3. Fetch inventory reservation
		res, err := u.orderRepo.GetReservationByOrderID(txCtx, order.ID)
		if err != nil {
			return err
		}

		// 4. Update reservation status to completed
		err = u.orderRepo.UpdateReservationStatus(txCtx, res.ID, "completed")
		if err != nil {
			return err
		}

		// 5. Complete reservation & deduct actual stock
		for _, item := range res.Items {
			// Lock stock
			_, _, err := u.orderRepo.LockStock(txCtx, item.VariantID, res.StoreID)
			if err != nil {
				return err
			}
			// Decrement reserved stock
			err = u.orderRepo.UpdateReservedStock(txCtx, item.VariantID, res.StoreID, -item.Quantity)
			if err != nil {
				return err
			}
			// Deduct actual stock
			qtyAfter, err := u.orderRepo.DeductStock(txCtx, item.VariantID, res.StoreID, item.Quantity)
			if err != nil {
				return err
			}
			// Write inventory log
			err = u.orderRepo.AddInventoryLog(txCtx, item.VariantID, res.StoreID, -item.Quantity, qtyAfter, "order_confirmed", order.OrderCode)
			if err != nil {
				return err
			}
		}

		// 6. Update order status to confirmed & paid
		err = u.orderRepo.UpdateOrderStatuses(txCtx, order.ID, confirmedStatusID, paidStatusID, order.ShippingStatusID)
		if err != nil {
			return err
		}

		// 7. Write history
		histOrder := &domain.OrderStatusHistory{
			OrderID:    order.ID,
			StatusType: "order",
			FromStatus: pointerToString("pending"),
			ToStatus:   "confirmed",
			Note:       pointerToString("Webhook thanh toán thành công"),
		}
		err = u.orderRepo.CreateOrderStatusHistory(txCtx, histOrder)
		if err != nil {
			return err
		}

		histPayment := &domain.OrderStatusHistory{
			OrderID:    order.ID,
			StatusType: "payment",
			FromStatus: pointerToString("unpaid"),
			ToStatus:   "paid",
			Note:       pointerToString("Đã thanh toán"),
		}
		err = u.orderRepo.CreateOrderStatusHistory(txCtx, histPayment)
		if err != nil {
			return err
		}

		orderToShip = order
		details, err := u.orderRepo.GetOrderDetails(txCtx, order.ID)
		if err == nil {
			detailsToShip = details
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 8. Trigger shipping order to GHN Sandbox (Post-commit background trigger)
	if orderToShip != nil && orderToShip.ShippingProvider != nil && *orderToShip.ShippingProvider == "ghn" && len(detailsToShip) > 0 {
		shippingCode, err := u.ghnClient.CreateShippingOrder(orderToShip, detailsToShip)
		if err == nil && shippingCode != "" {
			_ = u.orderRepo.UpdateOrderShippingInfo(ctx, orderToShip.ID, "ghn", shippingCode)
		}
	}

	return nil
}

func (u *OrderUsecase) CancelExpiredReservations(ctx context.Context) error {
	reservations, err := u.orderRepo.GetExpiredPendingReservations(ctx)
	if err != nil {
		return err
	}

	for _, res := range reservations {
		// Isolate cancellation process for each expired reservation in a separate transaction
		_ = u.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
			// 1. Set reservation to expired
			err = u.orderRepo.UpdateReservationStatus(txCtx, res.ID, "expired")
			if err != nil {
				return err
			}

			// 2. Fetch order associated
			var order *domain.Order
			if res.PayosOrderCode != nil {
				order, err = u.orderRepo.GetOrderByPaymentRefForUpdate(txCtx, *res.PayosOrderCode)
				if err != nil {
					return err
				}
			}

			if order != nil {
				// Get status IDs
				cancelledStatusID, _ := u.orderRepo.GetStatusIDByCode(txCtx, "order", "cancelled")
				
				// Release reserved stock hold
				for _, item := range res.Items {
					_ = u.orderRepo.UpdateReservedStock(txCtx, item.VariantID, res.StoreID, -item.Quantity)
				}

				// Release voucher if applied
				if order.VoucherID != nil {
					_ = u.orderRepo.IncrementVoucherUsedCount(txCtx, *order.VoucherID, -1)
					_ = u.orderRepo.DeleteVoucherUsage(txCtx, *order.VoucherID, order.UserID, order.ID)
				}

				// Cancel order
				_ = u.orderRepo.UpdateOrderStatuses(txCtx, order.ID, cancelledStatusID, order.PaymentStatusID, order.ShippingStatusID)

				// Log history
				hist := &domain.OrderStatusHistory{
					OrderID:    order.ID,
					StatusType: "order",
					FromStatus: pointerToString("pending"),
					ToStatus:   "cancelled",
					Note:       pointerToString("Hủy đơn hàng do hết hạn thanh toán online"),
				}
				_ = u.orderRepo.CreateOrderStatusHistory(txCtx, hist)
			}

			return nil
		})
	}

	return nil
}

func (u *OrderUsecase) ListOrders(ctx context.Context, userID *int, storeID *int, page int, limit int, query string, orderStatus, paymentStatus, shippingStatus *string) ([]*domain.OrderResponse, int, error) {
	orders, total, err := u.orderRepo.ListOrders(ctx, userID, storeID, page, limit, query, orderStatus, paymentStatus, shippingStatus)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*domain.OrderResponse, len(orders))
	for i, o := range orders {
		details, _ := u.orderRepo.GetOrderDetails(ctx, o.ID)
		orderLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "order", o.OrderStatusID)
		payLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "payment", o.PaymentStatusID)
		shipLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "shipping", o.ShippingStatusID)

		responses[i] = &domain.OrderResponse{
			Order:               *o,
			Items:               details,
			OrderStatusLabel:    orderLabel,
			PaymentStatusLabel:  payLabel,
			ShippingStatusLabel: shipLabel,
		}
	}

	return responses, total, nil
}

func (u *OrderUsecase) GetOrderDetails(ctx context.Context, orderID int, userID *int) (*domain.OrderResponse, error) {
	order, err := u.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if userID != nil && order.UserID != *userID {
		return nil, domain.ErrOrderNotFound
	}

	details, err := u.orderRepo.GetOrderDetails(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	orderLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "order", order.OrderStatusID)
	payLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "payment", order.PaymentStatusID)
	shipLabel, _ := u.orderRepo.GetStatusLabelByID(ctx, "shipping", order.ShippingStatusID)
	history, _ := u.orderRepo.GetOrderStatusHistory(ctx, order.ID)

	return &domain.OrderResponse{
		Order:               *order,
		Items:               details,
		OrderStatusLabel:    orderLabel,
		PaymentStatusLabel:  payLabel,
		ShippingStatusLabel: shipLabel,
		History:             history,
	}, nil
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, orderID int, actorUserID int, isAdmin bool, note string) error {
	return u.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Fetch order
		order, err := u.orderRepo.GetOrderByIDForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}

		if !isAdmin && order.UserID != actorUserID {
			return domain.ErrOrderNotFound
		}

		// 2. Fetch pending status ID
		pendingStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "order", "pending")
		if err != nil {
			return err
		}
		cancelledStatusID, err := u.orderRepo.GetStatusIDByCode(txCtx, "order", "cancelled")
		if err != nil {
			return err
		}

		// Check if order is cancellable
		if order.OrderStatusID != pendingStatusID {
			return domain.ErrOrderCannotBeCancelled
		}

		// 3. Revert stock hold
		details, err := u.orderRepo.GetOrderDetails(txCtx, order.ID)
		if err != nil {
			return err
		}

		paidStatusID, _ := u.orderRepo.GetStatusIDByCode(txCtx, "payment", "paid")

		if order.PaymentMethod == "cod" || order.PaymentStatusID == paidStatusID {
			// Stock was already deducted, so we add it back
			for _, item := range details {
				qtyAfter, err := u.orderRepo.DeductStock(txCtx, item.VariantID, order.StoreID, -item.Quantity)
				if err != nil {
					return err
				}
				_ = u.orderRepo.AddInventoryLog(txCtx, item.VariantID, order.StoreID, item.Quantity, qtyAfter, "order_cancelled", order.OrderCode)
			}
		} else {
			// Stock is reserved, release reserved stock hold
			for _, item := range details {
				_ = u.orderRepo.UpdateReservedStock(txCtx, item.VariantID, order.StoreID, -item.Quantity)
			}
			// Mark reservation as expired/completed so it doesn't process again
			res, err := u.orderRepo.GetReservationByOrderID(txCtx, order.ID)
			if err == nil {
				_ = u.orderRepo.UpdateReservationStatus(txCtx, res.ID, "expired")
			}
		}

		// 4. Release voucher slots
		if order.VoucherID != nil {
			err = u.orderRepo.IncrementVoucherUsedCount(txCtx, *order.VoucherID, -1)
			if err != nil {
				return err
			}
			err = u.orderRepo.DeleteVoucherUsage(txCtx, *order.VoucherID, order.UserID, order.ID)
			if err != nil {
				return err
			}
		}

		// 5. Update status to cancelled
		err = u.orderRepo.UpdateOrderStatuses(txCtx, order.ID, cancelledStatusID, order.PaymentStatusID, order.ShippingStatusID)
		if err != nil {
			return err
		}

		// 6. Log history
		var oldStatusStr string
		oldStatusStr = "pending"

		hist := &domain.OrderStatusHistory{
			OrderID:    order.ID,
			StatusType: "order",
			FromStatus: &oldStatusStr,
			ToStatus:   "cancelled",
			ChangedBy:  &actorUserID,
			Note:       &note,
		}
		err = u.orderRepo.CreateOrderStatusHistory(txCtx, hist)
		return err
	})
}

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID int, actorUserID int, req *domain.UpdateOrderStatusRequest) error {
	var orderToShip *domain.Order
	var detailsToShip []domain.OrderDetailResponse

	err := u.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		order, err := u.orderRepo.GetOrderByIDForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}

		oldOrderLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "order", order.OrderStatusID)
		oldPayLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "payment", order.PaymentStatusID)
		oldShipLabel, _ := u.orderRepo.GetStatusLabelByID(txCtx, "shipping", order.ShippingStatusID)

		orderStatusID := order.OrderStatusID
		paymentStatusID := order.PaymentStatusID
		shippingStatusID := order.ShippingStatusID

		// Update order status if requested
		if req.OrderStatusCode != nil {
			id, err := u.orderRepo.GetStatusIDByCode(txCtx, "order", *req.OrderStatusCode)
			if err != nil {
				return err
			}
			orderStatusID = id

			hist := &domain.OrderStatusHistory{
				OrderID:    order.ID,
				StatusType: "order",
				FromStatus: &oldOrderLabel,
				ToStatus:   *req.OrderStatusCode,
				ChangedBy:  &actorUserID,
				Note:       req.Note,
			}
			_ = u.orderRepo.CreateOrderStatusHistory(txCtx, hist)
		}

		// Update payment status if requested
		if req.PaymentStatusCode != nil {
			id, err := u.orderRepo.GetStatusIDByCode(txCtx, "payment", *req.PaymentStatusCode)
			if err != nil {
				return err
			}
			paymentStatusID = id

			hist := &domain.OrderStatusHistory{
				OrderID:    order.ID,
				StatusType: "payment",
				FromStatus: &oldPayLabel,
				ToStatus:   *req.PaymentStatusCode,
				ChangedBy:  &actorUserID,
				Note:       req.Note,
			}
			_ = u.orderRepo.CreateOrderStatusHistory(txCtx, hist)
		}

		// Update shipping status if requested
		if req.ShippingStatusCode != nil {
			id, err := u.orderRepo.GetStatusIDByCode(txCtx, "shipping", *req.ShippingStatusCode)
			if err != nil {
				return err
			}
			shippingStatusID = id

			hist := &domain.OrderStatusHistory{
				OrderID:    order.ID,
				StatusType: "shipping",
				FromStatus: &oldShipLabel,
				ToStatus:   *req.ShippingStatusCode,
				ChangedBy:  &actorUserID,
				Note:       req.Note,
			}
			_ = u.orderRepo.CreateOrderStatusHistory(txCtx, hist)
		}

		// Update DB
		err = u.orderRepo.UpdateOrderStatuses(txCtx, order.ID, orderStatusID, paymentStatusID, shippingStatusID)
		if err != nil {
			return err
		}

		// Update shipping details if present
		if req.ShippingProvider != nil && req.ShippingCode != nil {
			err = u.orderRepo.UpdateOrderShippingInfo(txCtx, order.ID, *req.ShippingProvider, *req.ShippingCode)
			if err != nil {
				return err
			}
		}

		// Trigger Point for COD order creator: when admin updates status to confirmed/processing, we automatically trigger GHN order creation
		confirmedID, _ := u.orderRepo.GetStatusIDByCode(txCtx, "order", "confirmed")
		processingID, _ := u.orderRepo.GetStatusIDByCode(txCtx, "order", "processing")
		if order.PaymentMethod == "cod" && (orderStatusID == confirmedID || orderStatusID == processingID) && order.ShippingCode == nil {
			orderToShip = order
			details, err := u.orderRepo.GetOrderDetails(txCtx, order.ID)
			if err == nil {
				detailsToShip = details
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Trigger shipping order outside transaction block
	if orderToShip != nil && orderToShip.ShippingProvider != nil && *orderToShip.ShippingProvider == "ghn" && len(detailsToShip) > 0 {
		shippingCode, err := u.ghnClient.CreateShippingOrder(orderToShip, detailsToShip)
		if err == nil && shippingCode != "" {
			_ = u.orderRepo.UpdateOrderShippingInfo(ctx, orderToShip.ID, "ghn", shippingCode)
		}
	}

	return nil
}

func pointerToString(s string) *string {
	return &s
}
