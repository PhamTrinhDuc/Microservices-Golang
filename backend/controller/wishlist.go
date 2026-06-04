package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WishlistController struct {
	usecase  domain.WishlistUsecase
	validate *validator.Validate
}

func NewWishlistController(usecase domain.WishlistUsecase) *WishlistController {
	return &WishlistController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

func (ctl *WishlistController) GetWishlist(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	items, err := ctl.usecase.GetWishlist(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, items)
}

func (ctl *WishlistController) AddToWishlist(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.AddToWishlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	item, err := ctl.usecase.AddToWishlist(ctx.Request.Context(), userID, req.VariantID)
	if err != nil {
		if errors.Is(err, domain.ErrVariantNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, item)
}

func (ctl *WishlistController) RemoveFromWishlist(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	variantIDParam := ctx.Param("variant_id")
	variantID, err := strconv.Atoi(variantIDParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid variant id")
		return
	}

	err = ctl.usecase.RemoveFromWishlist(ctx.Request.Context(), userID, variantID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}
