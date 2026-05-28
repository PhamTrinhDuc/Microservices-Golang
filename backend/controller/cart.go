package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CartController struct {
	usecase  domain.CartUsecase
	validate *validator.Validate
}

func NewCartController(usecase domain.CartUsecase) *CartController {
	return &CartController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

func (ctl *CartController) GetCart(ctx *gin.Context) {
	userID, sessionID := ctl.resolveUserAndSession(ctx)
	if userID == nil && sessionID == nil {
		utils.RespondBadRequest(ctx, "either user_id or X-Session-ID is required")
		return
	}

	res, err := ctl.usecase.GetCart(ctx.Request.Context(), userID, sessionID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *CartController) AddToCart(ctx *gin.Context) {
	userID, sessionID := ctl.resolveUserAndSession(ctx)

	var req domain.AddToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if userID == nil && sessionID == nil {
		if req.SessionID != nil && *req.SessionID != "" {
			sessionID = req.SessionID
		} else {
			utils.RespondBadRequest(ctx, "session_id is required for guest users")
			return
		}
	}

	res, err := ctl.usecase.AddToCart(ctx.Request.Context(), userID, sessionID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrVariantNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrInvalidQuantity) {
			utils.RespondBadRequest(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *CartController) UpdateItemQuantity(ctx *gin.Context) {
	idParam := ctx.Param("id")
	itemID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid cart item id")
		return
	}

	userID, sessionID := ctl.resolveUserAndSession(ctx)
	if userID == nil && sessionID == nil {
		utils.RespondBadRequest(ctx, "either user_id or X-Session-ID is required")
		return
	}

	var req domain.UpdateQuantityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.UpdateItemQuantity(ctx.Request.Context(), userID, sessionID, itemID, req.Quantity)
	if err != nil {
		if errors.Is(err, domain.ErrCartItemNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.RespondForbidden(ctx, "permission denied for this cart item")
			return
		}
		if errors.Is(err, domain.ErrInvalidQuantity) {
			utils.RespondBadRequest(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *CartController) RemoveItem(ctx *gin.Context) {
	idParam := ctx.Param("id")
	itemID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid cart item id")
		return
	}

	userID, sessionID := ctl.resolveUserAndSession(ctx)
	if userID == nil && sessionID == nil {
		utils.RespondBadRequest(ctx, "either user_id or X-Session-ID is required")
		return
	}

	err = ctl.usecase.RemoveItem(ctx.Request.Context(), userID, sessionID, itemID)
	if err != nil {
		if errors.Is(err, domain.ErrCartItemNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.RespondForbidden(ctx, "permission denied for this cart item")
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

func (ctl *CartController) ClearCart(ctx *gin.Context) {
	userID, sessionID := ctl.resolveUserAndSession(ctx)
	if userID == nil && sessionID == nil {
		utils.RespondBadRequest(ctx, "either user_id or X-Session-ID is required")
		return
	}

	err := ctl.usecase.ClearCart(ctx.Request.Context(), userID, sessionID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

func (ctl *CartController) MergeCart(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.MergeCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	err := ctl.usecase.MergeCart(ctx.Request.Context(), userID, req.SessionID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{"message": "cart merged successfully"})
}

func (ctl *CartController) resolveUserAndSession(ctx *gin.Context) (*int, *string) {
	var userID *int
	var sessionID *string

	if userIDVal, exists := ctx.Get("user_id"); exists {
		if id, ok := userIDVal.(int); ok {
			userID = &id
		}
	}

	if sID := ctx.GetHeader("X-Session-ID"); sID != "" {
		sessionID = &sID
	} else if sID := ctx.Query("session_id"); sID != "" {
		sessionID = &sID
	}

	return userID, sessionID
}
