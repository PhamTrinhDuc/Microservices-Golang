package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	useCase domain.UserUsecase
}

func NewUserController(useCase domain.UserUsecase) *UserController {
	return &UserController{
		useCase: useCase,
	}
}

// GetMe handles GET /profile.
func (uc *UserController) GetMe(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}

	userID := userIDVal.(int)
	user, err := uc.useCase.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, user)
}

// UpdateProfile handles PUT /profile.
func (uc *UserController) UpdateProfile(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	user, err := uc.useCase.UpdateProfile(ctx.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, user)
}

// UpdatePassword handles PUT /profile/password.
func (uc *UserController) UpdatePassword(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	err := uc.useCase.UpdatePassword(ctx.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrInvalidPassword) {
			utils.RespondBadRequest(ctx, "Mật khẩu cũ không chính xác")
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{"message": "Đổi mật khẩu thành công"})
}

type lockRequest struct {
	IsLock bool `json:"is_lock"`
}

// LockUser handles PUT /admin/users/:id/lock.
func (uc *UserController) LockUser(ctx *gin.Context) {
	idParam := ctx.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid user id param")
		return
	}

	var req lockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	err = uc.useCase.LockUser(ctx.Request.Context(), userID, req.IsLock)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	action := "unlocked"
	if req.IsLock {
		action = "locked"
	}
	utils.RespondOK(ctx, gin.H{"message": "User successfully " + action})
}

// ListUsers handles GET /admin/users.
func (uc *UserController) ListUsers(ctx *gin.Context) {
	page := utils.GetQueryInt(ctx, "page", 1)
	limit := utils.GetQueryInt(ctx, "limit", 10)
	query := ctx.Query("q")

	users, total, err := uc.useCase.List(ctx.Request.Context(), page, limit, query)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{
		"users":       users,
		"total_count": total,
		"page":        page,
		"limit":       limit,
	})
}
