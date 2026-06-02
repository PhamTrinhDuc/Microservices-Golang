package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/controller"
	"backend/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockInventoryUsecase struct {
	CreateStoreFunc             func(ctx context.Context, req *domain.CreateStoreRequest) (*domain.Store, error)
	ListStoresFunc              func(ctx context.Context, province string, district string) ([]*domain.Store, error)
	GetStoreByIDFunc            func(ctx context.Context, id int) (*domain.Store, error)
	UpdateStoreFunc             func(ctx context.Context, id int, req *domain.UpdateStoreRequest) (*domain.Store, error)
	DeactivateStoreFunc         func(ctx context.Context, id int) error
	CreateSupplierFunc          func(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.Supplier, error)
	ListSuppliersFunc           func(ctx context.Context) ([]*domain.Supplier, error)
	UpdateSupplierFunc          func(ctx context.Context, id int, req *domain.UpdateSupplierRequest) (*domain.Supplier, error)
	DeleteSupplierFunc          func(ctx context.Context, id int) error
	ImportGoodsFunc             func(ctx context.Context, creatorID int, req *domain.ImportGoodsRequest) (*domain.ImportInvoice, error)
	ListImportInvoicesFunc      func(ctx context.Context, storeID *int, page, limit int) ([]*domain.ImportInvoiceResponse, int, error)
	GetImportInvoiceDetailsFunc func(ctx context.Context, invoiceID int) (*domain.ImportInvoiceDetailsResponse, error)
	AdjustInventoryFunc         func(ctx context.Context, storeID int, creatorID int, req *domain.AdjustInventoryRequest) error
	ListStoreInventoryFunc      func(ctx context.Context, storeID int) ([]*domain.ProductInventory, error)
	GetLowStockAlertsFunc       func(ctx context.Context, storeID *int) ([]*domain.LowStockAlertResponse, error)
	GetInventoryLogsFunc        func(ctx context.Context, query *domain.InventoryLogsQuery) (*domain.InventoryLogsResult, error)
}

func (m *mockInventoryUsecase) CreateStore(ctx context.Context, req *domain.CreateStoreRequest) (*domain.Store, error) {
	return m.CreateStoreFunc(ctx, req)
}

func (m *mockInventoryUsecase) ListStores(ctx context.Context, province string, district string) ([]*domain.Store, error) {
	return m.ListStoresFunc(ctx, province, district)
}

func (m *mockInventoryUsecase) GetStoreByID(ctx context.Context, id int) (*domain.Store, error) {
	return m.GetStoreByIDFunc(ctx, id)
}

func (m *mockInventoryUsecase) UpdateStore(ctx context.Context, id int, req *domain.UpdateStoreRequest) (*domain.Store, error) {
	return m.UpdateStoreFunc(ctx, id, req)
}

func (m *mockInventoryUsecase) DeactivateStore(ctx context.Context, id int) error {
	return m.DeactivateStoreFunc(ctx, id)
}

func (m *mockInventoryUsecase) CreateSupplier(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.Supplier, error) {
	return m.CreateSupplierFunc(ctx, req)
}

func (m *mockInventoryUsecase) ListSuppliers(ctx context.Context) ([]*domain.Supplier, error) {
	return m.ListSuppliersFunc(ctx)
}

func (m *mockInventoryUsecase) UpdateSupplier(ctx context.Context, id int, req *domain.UpdateSupplierRequest) (*domain.Supplier, error) {
	return m.UpdateSupplierFunc(ctx, id, req)
}

func (m *mockInventoryUsecase) DeleteSupplier(ctx context.Context, id int) error {
	return m.DeleteSupplierFunc(ctx, id)
}

func (m *mockInventoryUsecase) ImportGoods(ctx context.Context, creatorID int, req *domain.ImportGoodsRequest) (*domain.ImportInvoice, error) {
	return m.ImportGoodsFunc(ctx, creatorID, req)
}

func (m *mockInventoryUsecase) ListImportInvoices(ctx context.Context, storeID *int, page, limit int) ([]*domain.ImportInvoiceResponse, int, error) {
	return m.ListImportInvoicesFunc(ctx, storeID, page, limit)
}

func (m *mockInventoryUsecase) GetImportInvoiceDetails(ctx context.Context, invoiceID int) (*domain.ImportInvoiceDetailsResponse, error) {
	return m.GetImportInvoiceDetailsFunc(ctx, invoiceID)
}

func (m *mockInventoryUsecase) AdjustInventory(ctx context.Context, storeID int, creatorID int, req *domain.AdjustInventoryRequest) error {
	return m.AdjustInventoryFunc(ctx, storeID, creatorID, req)
}

func (m *mockInventoryUsecase) ListStoreInventory(ctx context.Context, storeID int) ([]*domain.ProductInventory, error) {
	return m.ListStoreInventoryFunc(ctx, storeID)
}

func (m *mockInventoryUsecase) GetLowStockAlerts(ctx context.Context, storeID *int) ([]*domain.LowStockAlertResponse, error) {
	return m.GetLowStockAlertsFunc(ctx, storeID)
}

func (m *mockInventoryUsecase) GetInventoryLogs(ctx context.Context, query *domain.InventoryLogsQuery) (*domain.InventoryLogsResult, error) {
	return m.GetInventoryLogsFunc(ctx, query)
}

func TestInventoryController_CreateStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success created", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			CreateStoreFunc: func(ctx context.Context, req *domain.CreateStoreRequest) (*domain.Store, error) {
				return &domain.Store{
					ID:       1,
					Name:     req.Name,
					Province: req.Province,
					District: req.District,
					Ward:     req.Ward,
				}, nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.POST("/admin/stores", ctl.CreateStore)

		reqBody := domain.CreateStoreRequest{
			Name:     "Cửa hàng Quận 1",
			Province: "Hồ Chí Minh",
			District: "Quận 1",
			Ward:     "Bến Nghé",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/stores", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)
	})

	t.Run("validation error missing fields", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.POST("/admin/stores", ctl.CreateStore)

		reqBody := domain.CreateStoreRequest{
			Name: "", // missing required field Name
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/stores", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusBadRequest, w.Code)
	})
}

func TestInventoryController_ImportGoods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success created", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			ImportGoodsFunc: func(ctx context.Context, creatorID int, req *domain.ImportGoodsRequest) (*domain.ImportInvoice, error) {
				return &domain.ImportInvoice{
					ID:         100,
					SupplierID: req.SupplierID,
					StoreID:    req.StoreID,
					CreatedBy:  creatorID,
					TotalItems: 10,
				}, nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.POST("/admin/inventory/import", func(c *gin.Context) {
			c.Set("user_id", 1) // Mock logged in user ID
			ctl.ImportGoods(c)
		})

		reqBody := domain.ImportGoodsRequest{
			SupplierID: 1,
			StoreID:    2,
			Items: []domain.ImportItemDTO{
				{VariantID: 10, Quantity: 10, PriceImport: 1500000},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/inventory/import", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusCreated, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		is.Equal(float64(100), data["id"])
		is.Equal(float64(1), data["created_by"])
	})

	t.Run("unauthorized user context missing", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.POST("/admin/inventory/import", ctl.ImportGoods)

		reqBody := domain.ImportGoodsRequest{
			SupplierID: 1,
			StoreID:    2,
			Items: []domain.ImportItemDTO{
				{VariantID: 10, Quantity: 10, PriceImport: 1500000},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/inventory/import", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusUnauthorized, w.Code)
	})

	t.Run("not found entities", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			ImportGoodsFunc: func(ctx context.Context, creatorID int, req *domain.ImportGoodsRequest) (*domain.ImportInvoice, error) {
				return nil, domain.ErrStoreNotFound
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.POST("/admin/inventory/import", func(c *gin.Context) {
			c.Set("user_id", 1)
			ctl.ImportGoods(c)
		})

		reqBody := domain.ImportGoodsRequest{
			SupplierID: 1,
			StoreID:    999, // store not found
			Items: []domain.ImportItemDTO{
				{VariantID: 10, Quantity: 10, PriceImport: 1500000},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/admin/inventory/import", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNotFound, w.Code)
	})
}

func TestInventoryController_AdjustInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success adjust", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			AdjustInventoryFunc: func(ctx context.Context, storeID int, creatorID int, req *domain.AdjustInventoryRequest) error {
				return nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.PUT("/admin/stores/:id/inventory", func(c *gin.Context) {
			c.Set("user_id", 1)
			ctl.AdjustInventory(c)
		})

		reqBody := domain.AdjustInventoryRequest{
			Adjustments: []domain.AdjustItemDTO{
				{VariantID: 5, NewQuantity: 20},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/admin/stores/1/inventory", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)
	})
}

func TestInventoryController_GetLowStockAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success list alerts", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			GetLowStockAlertsFunc: func(ctx context.Context, storeID *int) ([]*domain.LowStockAlertResponse, error) {
				return []*domain.LowStockAlertResponse{
					{
						VariantID:         12,
						SKU:               "IP17-BLU",
						VariantName:       "Xanh Dương",
						ProductID:         "iphone-17",
						ProductName:       "iPhone 17",
						StoreID:           1,
						StoreName:         "Store 1",
						Quantity:          2,
						LowStockThreshold: 5,
					},
				}, nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.GET("/admin/inventory/low-stock", ctl.GetLowStockAlerts)

		req := httptest.NewRequest(http.MethodGet, "/admin/inventory/low-stock?store_id=1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		is.Len(data, 1)

		alert := data[0].(map[string]interface{})
		is.Equal(float64(12), alert["variant_id"])
		is.Equal("IP17-BLU", alert["sku"])
		is.Equal(float64(2), alert["quantity"])
		is.Equal(float64(5), alert["low_stock_threshold"])
	})
}

func TestInventoryController_GetInventoryLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success fetch history logs", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			GetInventoryLogsFunc: func(ctx context.Context, q *domain.InventoryLogsQuery) (*domain.InventoryLogsResult, error) {
				return &domain.InventoryLogsResult{
					Logs: []*domain.InventoryLogResponse{
						{
							ID:          1,
							VariantID:   12,
							SKU:         "IP17-BLU",
							VariantName: "Xanh Dương",
							StoreID:     1,
							StoreName:   "Store 1",
							ChangeQty:   5,
							QtyAfter:    10,
							Reason:      "import",
							CreatorName: "Admin",
						},
					},
					TotalCount: 1,
					Page:       q.Page,
					Limit:      q.Limit,
				}, nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.GET("/admin/inventory/logs", ctl.GetInventoryLogs)

		req := httptest.NewRequest(http.MethodGet, "/admin/inventory/logs?store_id=1&reason=import&page=1&limit=10", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		is.Equal(float64(1), data["total_count"])

		logs := data["logs"].([]interface{})
		is.Len(logs, 1)

		log := logs[0].(map[string]interface{})
		is.Equal(float64(5), log["change_qty"])
		is.Equal("import", log["reason"])
		is.Equal("Store 1", log["store_name"])
	})
}

func TestInventoryController_SupplierCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update supplier success", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			UpdateSupplierFunc: func(ctx context.Context, id int, req *domain.UpdateSupplierRequest) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:           id,
					Name:         req.Name,
					Address:      req.Address,
					ContactPhone: req.ContactPhone,
				}, nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.PUT("/admin/suppliers/:id", ctl.UpdateSupplier)

		addr := "Hà Nội"
		phone := "0987654321"
		reqBody := domain.UpdateSupplierRequest{
			Name:         "Nhà cung cấp mới",
			Address:      &addr,
			ContactPhone: &phone,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/admin/suppliers/1", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusOK, w.Code)
	})

	t.Run("delete supplier success", func(t *testing.T) {
		is := assert.New(t)
		mockUC := &mockInventoryUsecase{
			DeleteSupplierFunc: func(ctx context.Context, id int) error {
				return nil
			},
		}

		ctl := controller.NewInventoryController(mockUC)
		r := gin.New()
		r.DELETE("/admin/suppliers/:id", ctl.DeleteSupplier)

		req := httptest.NewRequest(http.MethodDelete, "/admin/suppliers/1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		is.Equal(http.StatusNoContent, w.Code)
	})
}
