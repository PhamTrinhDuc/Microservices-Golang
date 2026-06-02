package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"backend/domain"
	"backend/internal/payment"
	"backend/usecase"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	useCase     domain.OrderUsecase
	payosClient usecase.PayOSInterface
}

func NewOrderController(useCase domain.OrderUsecase, payosClient usecase.PayOSInterface) *OrderController {
	return &OrderController{
		useCase:     useCase,
		payosClient: payosClient,
	}
}

// Checkout creates a new order from current user cart
// @Summary Customer Checkout
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body domain.CheckoutOrderRequest true "Checkout Request Details"
// @Success 201 {object} domain.OrderResponse
// @Router /api/v1/orders/checkout [post]
func (oc *OrderController) Checkout(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	var req domain.CheckoutOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := oc.useCase.CheckoutOrder(c.Request.Context(), userID, &req)
	if err != nil {
		if err == domain.ErrEmptyCart || err == domain.ErrInsufficientStock || err == domain.ErrInvalidPaymentMethod || err == domain.ErrAddressNotFound || err == domain.ErrInvalidAddress {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrVoucherExpired || err == domain.ErrVoucherLimitReached || err == domain.ErrVoucherUserLimitReached || err == domain.ErrVoucherMinAmountNotMet {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// ListMyOrders shows personal paginated order history
// @Summary List customer orders
// @Tags Orders
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} gin.H
// @Router /api/v1/orders [get]
func (oc *OrderController) ListMyOrders(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	res, total, err := oc.useCase.ListOrders(c.Request.Context(), &userID, nil, page, limit, "", nil, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetMyOrderDetails retrieves specific customer order details
// @Summary View order details
// @Tags Orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Order ID"
// @Success 200 {object} domain.OrderResponse
// @Router /api/v1/orders/{id} [get]
func (oc *OrderController) GetMyOrderDetails(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	res, err := oc.useCase.GetOrderDetails(c.Request.Context(), orderID, &userID)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// CancelMyOrder cancels order if it is in pending/unpaid state
// @Summary Customer cancel order
// @Tags Orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Order ID"
// @Success 200 {object} gin.H
// @Router /api/v1/orders/{id}/cancel [post]
func (oc *OrderController) CancelMyOrder(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	err = oc.useCase.CancelOrder(c.Request.Context(), orderID, userID, false, "Khách hàng tự hủy đơn hàng")
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrOrderCannotBeCancelled {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}

// ConfirmPaymentWebhook handles PayOS public webhook notifications
// @Summary Webhook confirm payment from PayOS
// @Tags Webhook
// @Accept json
// @Produce json
// @Param request body payment.PayOSWebhookPayload true "Webhook payload"
// @Success 200 {object} gin.H
// @Router /api/v1/payments/webhook [post]
func (oc *OrderController) ConfirmPaymentWebhook(c *gin.Context) {
	var payload payment.PayOSWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Checksum Signature from PayOS Sandbox
	if !oc.payosClient.VerifyWebhookSignature(payload) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	status, _ := payload.Data["status"].(string)
	if status != "PAID" {
		// Stop PayOS retries but log that it is ignored
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "message": "status is not PAID"})
		return
	}

	orderCodeVal := payload.Data["orderCode"]
	var payosOrderCode string
	switch v := orderCodeVal.(type) {
	case string:
		payosOrderCode = v
	case float64:
		payosOrderCode = strconv.FormatInt(int64(v), 10)
	case int64:
		payosOrderCode = strconv.FormatInt(v, 10)
	default:
		payosOrderCode = fmt.Sprintf("%v", v)
	}

	paymentCode, _ := payload.Data["paymentLinkId"].(string)

	err := oc.useCase.ConfirmPayment(c.Request.Context(), payosOrderCode, paymentCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// AdminListOrders lists all system orders for admin management
// @Summary Admin list all orders
// @Tags Admin Orders
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param store_id query int false "Filter by store"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/orders [get]
func (oc *OrderController) AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	storeIDStr := c.Query("store_id")
	query := c.Query("q")
	orderStatusStr := c.Query("order_status")
	paymentStatusStr := c.Query("payment_status")
	shippingStatusStr := c.Query("shipping_status")

	var storeID *int
	if storeIDStr != "" {
		if id, err := strconv.Atoi(storeIDStr); err == nil {
			storeID = &id
		}
	}

	var orderStatus *string
	if orderStatusStr != "" {
		orderStatus = &orderStatusStr
	}
	var paymentStatus *string
	if paymentStatusStr != "" {
		paymentStatus = &paymentStatusStr
	}
	var shippingStatus *string
	if shippingStatusStr != "" {
		shippingStatus = &shippingStatusStr
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	res, total, err := oc.useCase.ListOrders(c.Request.Context(), nil, storeID, page, limit, query, orderStatus, paymentStatus, shippingStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// AdminUpdateStatus updates status values for an order
// @Summary Admin update order status
// @Tags Admin Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Order ID"
// @Param request body domain.UpdateOrderStatusRequest true "Update Request Payload"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/orders/{id}/status [put]
func (oc *OrderController) AdminUpdateStatus(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req domain.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = oc.useCase.UpdateOrderStatus(c.Request.Context(), orderID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}

// AdminGetOrderDetails retrieves specific order details for admin (without userID restriction)
// @Summary Admin view order details
// @Tags Admin Orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Order ID"
// @Success 200 {object} domain.OrderResponse
// @Router /api/v1/admin/orders/{id} [get]
func (oc *OrderController) AdminGetOrderDetails(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	res, err := oc.useCase.GetOrderDetails(c.Request.Context(), orderID, nil)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
