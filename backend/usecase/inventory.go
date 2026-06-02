package usecase

import (
	"context"
	"fmt"

	"backend/domain"
)

type InventoryUsecase struct {
	repo domain.InventoryRepository
}

func NewInventoryUsecase(repo domain.InventoryRepository) *InventoryUsecase {
	return &InventoryUsecase{repo: repo}
}

// --- Store CRUD ---

func (uc *InventoryUsecase) CreateStore(ctx context.Context, req *domain.CreateStoreRequest) (*domain.Store, error) {
	if req.Name == "" || req.District == "" || req.Province == "" || req.Ward == "" {
		return nil, fmt.Errorf("name, district, province, and ward are required")
	}

	store := &domain.Store{
		Name:     req.Name,
		Hotline:  req.Hotline,
		District: req.District,
		Province: req.Province,
		Ward:     req.Ward,
		Road:     req.Road,
		Email:    req.Email,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}

	return uc.repo.CreateStore(ctx, store)
}

func (uc *InventoryUsecase) ListStores(ctx context.Context, province string, district string) ([]*domain.Store, error) {
	return uc.repo.ListStores(ctx, province, district)
}

func (uc *InventoryUsecase) GetStoreByID(ctx context.Context, id int) (*domain.Store, error) {
	return uc.repo.GetStoreByID(ctx, id)
}

func (uc *InventoryUsecase) UpdateStore(ctx context.Context, id int, req *domain.UpdateStoreRequest) (*domain.Store, error) {
	if req.Name == "" || req.District == "" || req.Province == "" || req.Ward == "" {
		return nil, fmt.Errorf("name, district, province, and ward are required")
	}

	// Verify store exists
	store, err := uc.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}

	store.Name = req.Name
	store.Hotline = req.Hotline
	store.District = req.District
	store.Province = req.Province
	store.Ward = req.Ward
	store.Road = req.Road
	store.Email = req.Email
	store.Lat = req.Lat
	store.Lng = req.Lng
	store.IsActive = req.IsActive

	return uc.repo.UpdateStore(ctx, store)
}

func (uc *InventoryUsecase) DeactivateStore(ctx context.Context, id int) error {
	return uc.repo.DeactivateStore(ctx, id)
}

// --- Supplier CRUD ---

func (uc *InventoryUsecase) CreateSupplier(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.Supplier, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("supplier name is required")
	}

	supplier := &domain.Supplier{
		Name:         req.Name,
		Address:      req.Address,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
	}

	return uc.repo.CreateSupplier(ctx, supplier)
}

func (uc *InventoryUsecase) ListSuppliers(ctx context.Context) ([]*domain.Supplier, error) {
	return uc.repo.ListSuppliers(ctx)
}

func (uc *InventoryUsecase) UpdateSupplier(ctx context.Context, id int, req *domain.UpdateSupplierRequest) (*domain.Supplier, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("supplier name is required")
	}

	// Verify supplier exists
	supplier, err := uc.repo.GetSupplierByID(ctx, id)
	if err != nil {
		return nil, err
	}

	supplier.Name = req.Name
	supplier.Address = req.Address
	supplier.ContactName = req.ContactName
	supplier.ContactPhone = req.ContactPhone
	supplier.ContactEmail = req.ContactEmail
	if req.IsDeleted != nil {
		supplier.IsDeleted = *req.IsDeleted
	}

	return uc.repo.UpdateSupplier(ctx, supplier)
}

func (uc *InventoryUsecase) DeleteSupplier(ctx context.Context, id int) error {
	return uc.repo.DeleteSupplier(ctx, id)
}

// --- Inventory Transaction Operations ---

func (uc *InventoryUsecase) ImportGoods(ctx context.Context, creatorID int, req *domain.ImportGoodsRequest) (*domain.ImportInvoice, error) {
	// 1. Verify store exists
	_, err := uc.repo.GetStoreByID(ctx, req.StoreID)
	if err != nil {
		return nil, err
	}

	// 2. Verify supplier exists
	_, err = uc.repo.GetSupplierByID(ctx, req.SupplierID)
	if err != nil {
		return nil, err
	}

	// 3. Map details and compute totals
	totalItems := 0
	details := make([]*domain.ImportInvoiceDetail, len(req.Items))
	for i, item := range req.Items {
		totalItems += item.Quantity
		details[i] = &domain.ImportInvoiceDetail{
			VariantID:   item.VariantID,
			Quantity:    item.Quantity,
			PriceImport: item.PriceImport,
		}
	}

	invoice := &domain.ImportInvoice{
		SupplierID: req.SupplierID,
		StoreID:    req.StoreID,
		CreatedBy:  creatorID,
		TotalItems: totalItems,
		Note:       req.Note,
	}

	return uc.repo.CreateImportInvoice(ctx, creatorID, invoice, details)
}

func (uc *InventoryUsecase) ListImportInvoices(ctx context.Context, storeID *int, page, limit int) ([]*domain.ImportInvoiceResponse, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	return uc.repo.ListImportInvoices(ctx, storeID, page, limit)
}

func (uc *InventoryUsecase) GetImportInvoiceDetails(ctx context.Context, invoiceID int) (*domain.ImportInvoiceDetailsResponse, error) {
	return uc.repo.GetImportInvoiceDetails(ctx, invoiceID)
}

func (uc *InventoryUsecase) AdjustInventory(ctx context.Context, storeID int, creatorID int, req *domain.AdjustInventoryRequest) error {
	// 1. Verify store exists
	_, err := uc.repo.GetStoreByID(ctx, storeID)
	if err != nil {
		return err
	}

	adjustments := make([]*domain.AdjustItemDTO, len(req.Adjustments))
	for i, adj := range req.Adjustments {
		adjustments[i] = &domain.AdjustItemDTO{
			VariantID:   adj.VariantID,
			NewQuantity: adj.NewQuantity,
		}
	}

	return uc.repo.AdjustInventory(ctx, storeID, creatorID, adjustments)
}

func (uc *InventoryUsecase) ListStoreInventory(ctx context.Context, storeID int) ([]*domain.ProductInventory, error) {
	// Verify store exists
	_, err := uc.repo.GetStoreByID(ctx, storeID)
	if err != nil {
		return nil, err
	}

	return uc.repo.ListStoreInventory(ctx, storeID)
}

func (uc *InventoryUsecase) GetLowStockAlerts(ctx context.Context, storeID *int) ([]*domain.LowStockAlertResponse, error) {
	if storeID != nil {
		// If store ID is provided, verify it exists
		_, err := uc.repo.GetStoreByID(ctx, *storeID)
		if err != nil {
			return nil, err
		}
	}

	return uc.repo.GetLowStockAlerts(ctx, storeID)
}

func (uc *InventoryUsecase) GetInventoryLogs(ctx context.Context, query *domain.InventoryLogsQuery) (*domain.InventoryLogsResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	if query.StoreID != nil {
		// Verify store exists
		_, err := uc.repo.GetStoreByID(ctx, *query.StoreID)
		if err != nil {
			return nil, err
		}
	}

	return uc.repo.GetInventoryLogs(ctx, query)
}
