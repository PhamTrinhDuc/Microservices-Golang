package controller

import (
	"errors"
	"strconv"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type InventoryController struct {
	usecase  domain.InventoryUsecase
	validate *validator.Validate
}

func NewInventoryController(usecase domain.InventoryUsecase) *InventoryController {
	return &InventoryController{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// --- Store CRUD ---

func (ctl *InventoryController) CreateStore(ctx *gin.Context) {
	var req domain.CreateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.CreateStore(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *InventoryController) ListStores(ctx *gin.Context) {
	province := ctx.Query("province")
	district := ctx.Query("district")

	res, err := ctl.usecase.ListStores(ctx.Request.Context(), province, district)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) GetStoreByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid store id")
		return
	}

	res, err := ctl.usecase.GetStoreByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) UpdateStore(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid store id")
		return
	}

	var req domain.UpdateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.UpdateStore(ctx.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) DeactivateStore(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid store id")
		return
	}

	err = ctl.usecase.DeactivateStore(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Supplier CRUD ---

func (ctl *InventoryController) CreateSupplier(ctx *gin.Context) {
	var req domain.CreateSupplierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.CreateSupplier(ctx.Request.Context(), &req)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *InventoryController) ListSuppliers(ctx *gin.Context) {
	res, err := ctl.usecase.ListSuppliers(ctx.Request.Context())
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) UpdateSupplier(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid supplier id")
		return
	}

	var req domain.UpdateSupplierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.UpdateSupplier(ctx.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) DeleteSupplier(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid supplier id")
		return
	}

	err = ctl.usecase.DeleteSupplier(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondNoContent(ctx)
}

// --- Inventory Operations ---

func (ctl *InventoryController) ImportGoods(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	var req domain.ImportGoodsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.ImportGoods(ctx.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) ||
			errors.Is(err, domain.ErrSupplierNotFound) ||
			errors.Is(err, domain.ErrVariantNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondCreated(ctx, res)
}

func (ctl *InventoryController) ListImportInvoices(ctx *gin.Context) {
	var storeID *int
	if storeIDStr := ctx.Query("store_id"); storeIDStr != "" {
		sID, err := strconv.Atoi(storeIDStr)
		if err == nil {
			storeID = &sID
		}
	}

	page := utils.GetQueryInt(ctx, "page", 1)
	limit := utils.GetQueryInt(ctx, "limit", 10)

	invoices, total, err := ctl.usecase.ListImportInvoices(ctx.Request.Context(), storeID, page, limit)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{
		"invoices":    invoices,
		"total_count": total,
		"page":        page,
		"limit":       limit,
	})
}

func (ctl *InventoryController) GetImportInvoiceDetails(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid invoice id")
		return
	}

	res, err := ctl.usecase.GetImportInvoiceDetails(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrImportInvoiceNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) AdjustInventory(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(ctx, "unauthorized context missing user_id")
		return
	}
	userID := userIDVal.(int)

	storeIDParam := ctx.Param("id")
	storeID, err := strconv.Atoi(storeIDParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid store id")
		return
	}

	var req domain.AdjustInventoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	if err := ctl.validate.Struct(&req); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	err = ctl.usecase.AdjustInventory(ctx.Request.Context(), storeID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) || errors.Is(err, domain.ErrVariantNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{"message": "inventory adjusted successfully"})
}

func (ctl *InventoryController) ListStoreInventory(ctx *gin.Context) {
	storeIDParam := ctx.Param("id")
	storeID, err := strconv.Atoi(storeIDParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid store id")
		return
	}

	res, err := ctl.usecase.ListStoreInventory(ctx.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) GetLowStockAlerts(ctx *gin.Context) {
	var storeID *int
	if storeIDStr := ctx.Query("store_id"); storeIDStr != "" {
		sID, err := strconv.Atoi(storeIDStr)
		if err == nil {
			storeID = &sID
		}
	}

	res, err := ctl.usecase.GetLowStockAlerts(ctx.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) GetInventoryLogs(ctx *gin.Context) {
	var q domain.InventoryLogsQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		utils.RespondBadRequest(ctx, err.Error())
		return
	}

	res, err := ctl.usecase.GetInventoryLogs(ctx.Request.Context(), &q)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, res)
}

func (ctl *InventoryController) ConfirmImportInvoice(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid invoice id")
		return
	}

	err = ctl.usecase.ConfirmImportInvoice(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrImportInvoiceNotFound) {
			utils.RespondNotFound(ctx, err.Error())
			return
		}
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, gin.H{"message": "import invoice confirmed and published successfully"})
}

func (ctl *InventoryController) GetLastImportPrice(ctx *gin.Context) {
	idParam := ctx.Param("id")
	variantID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid variant id")
		return
	}

	price, err := ctl.usecase.GetLastImportPrice(ctx.Request.Context(), variantID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, price)
}
