package domain

import (
	"context"
	"time"
)

// Models

type Store struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Hotline   *string   `json:"hotline" db:"hotline"`
	District  string    `json:"district" db:"district"`
	Province  string    `json:"province" db:"province"`
	Ward      string    `json:"ward" db:"ward"`
	Road      *string   `json:"road" db:"road"`
	Email     *string   `json:"email" db:"email"`
	Lat       *float64  `json:"lat" db:"lat"`
	Lng       *float64  `json:"lng" db:"lng"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Supplier struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Address   *string `json:"address" db:"address"`
	Phone     *string `json:"phone" db:"phone"`
	IsDeleted bool    `json:"is_deleted" db:"is_deleted"`
}

type ProductInventory struct {
	VariantID   int       `json:"variant_id" db:"variant_id"`
	StoreID     int       `json:"store_id" db:"store_id"`
	Quantity    int       `json:"quantity" db:"quantity"`
	Reserved    int       `json:"reserved" db:"reserved"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`

	// Joined properties
	VariantName string    `json:"variant_name,omitempty" db:"variant_name"`
	ProductName string    `json:"product_name,omitempty" db:"product_name"`
	SKU         string    `json:"sku,omitempty" db:"sku"`
}

type ImportInvoice struct {
	ID         int       `json:"id" db:"id"`
	SupplierID int       `json:"supplier_id" db:"supplier_id"`
	StoreID    int       `json:"store_id" db:"store_id"`
	CreatedBy  int       `json:"created_by" db:"created_by"`
	TotalItems int       `json:"total_items" db:"total_items"`
	Note       *string   `json:"note" db:"note"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type ImportInvoiceDetail struct {
	ID          int     `json:"id" db:"id"`
	InvoiceID   int     `json:"invoice_id" db:"invoice_id"`
	VariantID   int     `json:"variant_id" db:"variant_id"`
	Quantity    int     `json:"quantity" db:"quantity"`
	StockBefore int     `json:"stock_before" db:"stock_before"`
	PriceImport float64 `json:"price_import" db:"price_import"`
}

type InventoryLog struct {
	ID        int       `json:"id" db:"id"`
	VariantID int       `json:"variant_id" db:"variant_id"`
	StoreID    int       `json:"store_id" db:"store_id"`
	ChangeQty int       `json:"change_qty" db:"change_qty"`
	QtyAfter  int       `json:"qty_after" db:"qty_after"`
	Reason    string    `json:"reason" db:"reason"`
	RefID     *string   `json:"ref_id" db:"ref_id"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Request DTOs

type CreateStoreRequest struct {
	Name     string   `json:"name" validate:"required"`
	Hotline  *string  `json:"hotline"`
	District string   `json:"district" validate:"required"`
	Province string   `json:"province" validate:"required"`
	Ward     string   `json:"ward" validate:"required"`
	Road     *string  `json:"road"`
	Email    *string  `json:"email" validate:"omitempty,email"`
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
}

type UpdateStoreRequest struct {
	Name     string   `json:"name" validate:"required"`
	Hotline  *string  `json:"hotline"`
	District string   `json:"district" validate:"required"`
	Province string   `json:"province" validate:"required"`
	Ward     string   `json:"ward" validate:"required"`
	Road     *string  `json:"road"`
	Email    *string  `json:"email" validate:"omitempty,email"`
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
	IsActive bool     `json:"is_active"`
}

type CreateSupplierRequest struct {
	Name    string  `json:"name" validate:"required"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

type UpdateSupplierRequest struct {
	Name    string  `json:"name" validate:"required"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

type ImportItemDTO struct {
	VariantID   int     `json:"variant_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	PriceImport float64 `json:"price_import" validate:"required,gt=0"`
}

type ImportGoodsRequest struct {
	SupplierID int             `json:"supplier_id" validate:"required"`
	StoreID    int             `json:"store_id" validate:"required"`
	Note       *string         `json:"note"`
	Items      []ImportItemDTO `json:"items" validate:"required,min=1,dive"`
}

type AdjustItemDTO struct {
	VariantID   int `json:"variant_id" validate:"required"`
	NewQuantity int `json:"new_quantity" validate:"required,gte=0"`
}

type AdjustInventoryRequest struct {
	Adjustments []AdjustItemDTO `json:"adjustments" validate:"required,min=1,dive"`
}

type InventoryLogsQuery struct {
	StoreID   *int   `form:"store_id"`
	VariantID *int   `form:"variant_id"`
	Reason    string `form:"reason"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

// Response DTOs

type LowStockAlertResponse struct {
	VariantID         int     `json:"variant_id"`
	SKU               string  `json:"sku"`
	VariantName       string  `json:"variant_name"`
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name"`
	StoreID           int     `json:"store_id"`
	StoreName         string  `json:"store_name"`
	Quantity          int     `json:"quantity"`
	LowStockThreshold int     `json:"low_stock_threshold"`
}

type InventoryLogResponse struct {
	ID          int       `json:"id"`
	VariantID   int       `json:"variant_id"`
	SKU         string    `json:"sku"`
	VariantName string    `json:"variant_name"`
	StoreID     int       `json:"store_id"`
	StoreName   string    `json:"store_name"`
	ChangeQty   int       `json:"change_qty"`
	QtyAfter    int       `json:"qty_after"`
	Reason      string    `json:"reason"`
	RefID       *string   `json:"ref_id"`
	CreatedBy   int       `json:"created_by"`
	CreatorName string    `json:"creator_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type InventoryLogsResult struct {
	Logs       []*InventoryLogResponse `json:"logs"`
	TotalCount int                     `json:"total_count"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
}

type ImportInvoiceResponse struct {
	ID           int       `json:"id"`
	SupplierID   int       `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	StoreID      int       `json:"store_id"`
	StoreName    string    `json:"store_name"`
	CreatedBy    int       `json:"created_by"`
	CreatorName  string    `json:"creator_name"`
	TotalItems   int       `json:"total_items"`
	Note         *string   `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type ImportInvoiceDetailResponse struct {
	ID          int     `json:"id"`
	VariantID   int     `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	StockBefore int     `json:"stock_before"`
	PriceImport float64 `json:"price_import"`
}

type ImportInvoiceDetailsResponse struct {
	Invoice *ImportInvoiceResponse         `json:"invoice"`
	Details []*ImportInvoiceDetailResponse `json:"details"`
}

// Store Inventory interfaces

type InventoryRepository interface {
	CreateStore(ctx context.Context, store *Store) (*Store, error)
	ListStores(ctx context.Context, province string, district string) ([]*Store, error)
	GetStoreByID(ctx context.Context, id int) (*Store, error)
	UpdateStore(ctx context.Context, store *Store) (*Store, error)
	DeactivateStore(ctx context.Context, id int) error

	CreateSupplier(ctx context.Context, supplier *Supplier) (*Supplier, error)
	ListSuppliers(ctx context.Context) ([]*Supplier, error)
	GetSupplierByID(ctx context.Context, id int) (*Supplier, error)
	UpdateSupplier(ctx context.Context, supplier *Supplier) (*Supplier, error)
	DeleteSupplier(ctx context.Context, id int) error

	CreateImportInvoice(ctx context.Context, creatorID int, invoice *ImportInvoice, details []*ImportInvoiceDetail) (*ImportInvoice, error)
	ListImportInvoices(ctx context.Context, storeID *int, page, limit int) ([]*ImportInvoiceResponse, int, error)
	GetImportInvoiceDetails(ctx context.Context, invoiceID int) (*ImportInvoiceDetailsResponse, error)

	AdjustInventory(ctx context.Context, storeID int, creatorID int, adjustments []*AdjustItemDTO) error
	ListStoreInventory(ctx context.Context, storeID int) ([]*ProductInventory, error)
	GetLowStockAlerts(ctx context.Context, storeID *int) ([]*LowStockAlertResponse, error)
	GetInventoryLogs(ctx context.Context, query *InventoryLogsQuery) (*InventoryLogsResult, error)
}

type InventoryUsecase interface {
	CreateStore(ctx context.Context, req *CreateStoreRequest) (*Store, error)
	ListStores(ctx context.Context, province string, district string) ([]*Store, error)
	GetStoreByID(ctx context.Context, id int) (*Store, error)
	UpdateStore(ctx context.Context, id int, req *UpdateStoreRequest) (*Store, error)
	DeactivateStore(ctx context.Context, id int) error

	CreateSupplier(ctx context.Context, req *CreateSupplierRequest) (*Supplier, error)
	ListSuppliers(ctx context.Context) ([]*Supplier, error)
	UpdateSupplier(ctx context.Context, id int, req *UpdateSupplierRequest) (*Supplier, error)
	DeleteSupplier(ctx context.Context, id int) error

	ImportGoods(ctx context.Context, creatorID int, req *ImportGoodsRequest) (*ImportInvoice, error)
	ListImportInvoices(ctx context.Context, storeID *int, page, limit int) ([]*ImportInvoiceResponse, int, error)
	GetImportInvoiceDetails(ctx context.Context, invoiceID int) (*ImportInvoiceDetailsResponse, error)

	AdjustInventory(ctx context.Context, storeID int, creatorID int, req *AdjustInventoryRequest) error
	ListStoreInventory(ctx context.Context, storeID int) ([]*ProductInventory, error)
	GetLowStockAlerts(ctx context.Context, storeID *int) ([]*LowStockAlertResponse, error)
	GetInventoryLogs(ctx context.Context, query *InventoryLogsQuery) (*InventoryLogsResult, error)
}
