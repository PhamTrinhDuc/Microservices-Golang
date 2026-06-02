package controller

import (
	"strconv"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BannerController struct {
	usecase  domain.BannerUsecase
	validate *validator.Validate
}

func NewBannerController(usecase domain.BannerUsecase) *BannerController {
	return &BannerController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

func (ctl *BannerController) Create(ctx *gin.Context) {
	var req domain.CreateBannerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}
	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.Create(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondCreated(ctx, res)
}

func (ctl *BannerController) ListPublic(ctx *gin.Context) {
	res, err := ctl.usecase.List(ctx.Request.Context(), true)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *BannerController) ListAdmin(ctx *gin.Context) {
	res, err := ctl.usecase.List(ctx.Request.Context(), false)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *BannerController) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid banner id")
		return
	}

	var req domain.UpdateBannerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}
	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.Update(ctx.Request.Context(), id, &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondOK(ctx, res)
}

func (ctl *BannerController) Delete(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid banner id")
		return
	}

	err = ctl.usecase.Delete(ctx.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}
	utils.RespondNoContent(ctx)
}
