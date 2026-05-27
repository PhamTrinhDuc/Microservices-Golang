package controller

import (
	"errors"
	"net/http"
	"strconv"

	"backend/domain"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserController struct {
	useCase  domain.UserUsecase
	validate *validator.Validate
}

func NewUserController(useCase domain.UserUsecase) *UserController {
	return &UserController{
		useCase:  useCase,
		validate: validator.New(),
	}
}

// Register handles POST /auth/register.
func (uc *UserController) Register(ctx *gin.Context) {
	var req domain.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := uc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	result, err := uc.useCase.Register(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			ctx.JSON(http.StatusConflict, utils.HTTPResponse{Error: err.Error()})
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, result)
}

// Login handles POST /auth/login.
func (uc *UserController) Login(ctx *gin.Context) {
	var req domain.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := uc.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	user, token, err := uc.useCase.Authenticate(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPassword) {
			utils.RespondUnauthorized(ctx, err.Error())
			return
		}
		if errors.Is(err, domain.ErrLocked) {
			utils.RespondForbidden(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, domain.LoginResponse{
		User:  user,
		Token: token,
	})
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
