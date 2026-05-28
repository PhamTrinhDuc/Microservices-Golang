package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PromotionVoucherController struct {
	usecase  domain.PromotionVoucherUsecase
	validate *validator.Validate
}

func NewPromotionVoucherController(usecase domain.PromotionVoucherUsecase) *PromotionVoucherController {
	return &PromotionVoucherController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// --- Promotions Admin Handlers ---

func (ctl *PromotionVoucherController) CreatePromotion(ctx *gin.Context) {
	var req domain.CreatePromotionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.CreatePromotion(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) || errors.Is(err, domain.ErrVariantNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *PromotionVoucherController) ListPromotions(ctx *gin.Context) {
	res, err := ctl.usecase.ListPromotions(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) GetPromotionByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid promotion id")
		return
	}

	res, err := ctl.usecase.GetPromotionByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPromotionNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) UpdatePromotion(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid promotion id")
		return
	}

	var req domain.UpdatePromotionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.UpdatePromotion(ctx.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrPromotionNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) DeletePromotion(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid promotion id")
		return
	}

	err = ctl.usecase.DeletePromotion(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPromotionNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Vouchers Admin Handlers ---

func (ctl *PromotionVoucherController) CreateVoucher(ctx *gin.Context) {
	var req domain.CreateVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.CreateVoucher(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *PromotionVoucherController) ListVouchers(ctx *gin.Context) {
	res, err := ctl.usecase.ListVouchers(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) GetVoucherByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid voucher id")
		return
	}

	res, err := ctl.usecase.GetVoucherByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVoucherNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) UpdateVoucher(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid voucher id")
		return
	}

	var req domain.UpdateVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.UpdateVoucher(ctx.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrVoucherNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) DeleteVoucher(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid voucher id")
		return
	}

	err = ctl.usecase.DeleteVoucher(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVoucherNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Public Vouchers Handlers ---

func (ctl *PromotionVoucherController) ListActiveVouchers(ctx *gin.Context) {
	res, err := ctl.usecase.ListActiveVouchers(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *PromotionVoucherController) ApplyVoucher(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.ApplyVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.ApplyVoucher(ctx.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrVoucherNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrVoucherExpired) ||
			errors.Is(err, domain.ErrVoucherLimitReached) ||
			errors.Is(err, domain.ErrVoucherUserLimitReached) ||
			errors.Is(err, domain.ErrVoucherMinAmountNotMet) {
			utils.RespondBadRequest(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}
