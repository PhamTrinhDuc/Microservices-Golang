package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AddressController struct {
	useCase  domain.AddressUsecase
	validate *validator.Validate
}

func NewAddressController(useCase domain.AddressUsecase) *AddressController {
	return &AddressController{
		useCase:  useCase,
		validate: validator.New(),
	}
}

// Create handles POST /addresses.
func (ac *AddressController) Create(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.CreateAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ac.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := ac.useCase.Create(ctx.Request.Context(), userID, &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// List handles GET /addresses.
func (ac *AddressController) List(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	addresses, err := ac.useCase.List(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, addresses)
}

// SetDefault handles PUT /addresses/:id/default.
func (ac *AddressController) SetDefault(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	idParam := ctx.Param("id")
	addressID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid address id param")
		return
	}

	err = ac.useCase.SetDefault(ctx.Request.Context(), userID, addressID)
	if err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.RespondForbidden(ctx, "permission denied for this address")
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{"message": "Address set as default successfully"})
}

// Delete handles DELETE /addresses/:id.
func (ac *AddressController) Delete(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	idParam := ctx.Param("id")
	addressID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid address id param")
		return
	}

	err = ac.useCase.Delete(ctx.Request.Context(), userID, addressID)
	if err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.RespondForbidden(ctx, "permission denied for this address")
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

// Update handles PUT /addresses/:id.
func (ac *AddressController) Update(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	idParam := ctx.Param("id")
	addressID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid address id param")
		return
	}

	var req domain.CreateAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ac.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := ac.useCase.Update(ctx.Request.Context(), userID, addressID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.RespondForbidden(ctx, "permission denied for this address")
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, result)
}
