# Thiết Kế Chi Tiết Luồng Store & Inventory (Store & Inventory Module Design)
*Bổ sung tính năng Nhật ký tồn kho (Inventory Log), Xem lại hoá đơn nhập, Cảnh báo hàng thấp và Lịch sử thay đổi*

Tài liệu này mô tả chi tiết thiết kế hệ thống luồng dữ liệu nghiệp vụ cho module **Cửa hàng (Store)** và **Tồn kho (Inventory)**, bổ sung cơ chế lưu vết thay đổi kho (Inventory Log) và 3 tính năng quản trị kho nâng cao.

---

## 1. Cấu Trúc Cơ Sở Dữ Liệu Bổ Sung (Schema Change)

Hệ thống bổ sung bảng `inventory_log` để lưu vết mọi thay đổi số lượng tồn kho (nhập, xuất, kiểm kho, điều chỉnh thủ công):

```sql
CREATE TABLE inventory_log (
    id SERIAL PRIMARY KEY,
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES store(id) ON DELETE CASCADE,
    change_qty INTEGER NOT NULL,      -- Số lượng thay đổi: dương = nhập, âm = xuất/bán
    qty_after INTEGER NOT NULL,       -- Số lượng tồn kho snapshot ngay sau khi thay đổi
    reason VARCHAR(100) NOT NULL,     -- Lý do: "import", "manual_adjust", "order_confirmed", "order_cancelled"
    ref_id VARCHAR(100),              -- ID tham chiếu: mã hóa đơn nhập hoặc mã đơn hàng tương ứng
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index để tối ưu truy vấn lịch sử
CREATE INDEX idx_inventory_log_store_var ON inventory_log(store_id, variant_id);
CREATE INDEX idx_inventory_log_created ON inventory_log(created_at DESC);
```

---

## 2. Sơ đồ Quan hệ Thực thể Cập nhật (Updated Entity Relationships)

```mermaid
erDiagram
    store ||--o{ product_inventory : "holds stock"
    product_variant ||--o{ product_inventory : "stocked in"
    suppliers ||--o{ import_invoices : "supplies"
    store ||--o{ import_invoices : "receives"
    users ||--o{ import_invoices : "created by"
    import_invoices ||--o{ import_invoice_details : "contains"
    product_variant ||--o{ import_invoice_details : "imported item"
    store ||--o{ inventory_log : "tracked in"
    product_variant ||--o{ inventory_log : "tracked item"
    users ||--o{ inventory_log : "logged by"
```

---

## 3. Các API Endpoints Bổ Sung

### 3.1 Cửa hàng & Nhà cung cấp (Stores & Suppliers CRUD)
* `GET /api/v1/stores` - Xem danh sách cửa hàng (Public).
* `POST /api/v1/admin/stores` - Thêm cửa hàng (Chỉ Admin).
* `PUT /api/v1/admin/stores/:id` - Sửa cửa hàng (Chỉ Admin).
* `DELETE /api/v1/admin/stores/:id` - Tạm dừng cửa hàng (`is_active = false`) (Chỉ Admin).
* `POST /api/v1/admin/suppliers` - Thêm nhà cung cấp (Chỉ Admin).
* `GET /api/v1/admin/suppliers` - Xem danh sách nhà cung cấp (Chỉ Admin).
* `PUT /api/v1/admin/suppliers/:id` - Cập nhật nhà cung cấp (Chỉ Admin).
* `DELETE /api/v1/admin/suppliers/:id` - Tạm dừng nhà cung cấp (`is_deleted = true`) (Chỉ Admin).

### 3.2 Tồn kho & Hoá đơn & Logs (Inventory & Logs)
* `GET /api/v1/admin/stores/:id/inventory` - Xem danh sách tồn kho cửa hàng (Chỉ Admin).
* `PUT /api/v1/admin/stores/:id/inventory` - Điều chỉnh tồn kho thủ công (Tự động ghi nhận `manual_adjust` vào `inventory_log`) (Chỉ Admin).
* `POST /api/v1/admin/inventory/import` - Tạo hóa đơn nhập kho & Tự động cộng kho & Ghi nhận `import` log (Chỉ Admin).
* `GET /api/v1/admin/inventory/imports` - Xem danh sách hoá đơn nhập hàng (Chỉ Admin).
* `GET /api/v1/admin/inventory/imports/:id` - Xem chi tiết 1 hoá đơn nhập (kèm danh sách chi tiết hàng nhập) (Chỉ Admin).
* `GET /api/v1/admin/inventory/low-stock` - **Cảnh báo hàng thấp**: Xem các variant có tồn kho `<= low_stock_threshold` của sản phẩm đó (Chỉ Admin).
* `GET /api/v1/admin/inventory/logs` - **Lịch sử thay đổi tồn kho**: Xem toàn bộ lịch sử biến động kho (Phân trang, lọc theo store, variant, lý do) (Chỉ Admin).

---

## 4. Thiết Kế Luồng Nghiệp Vụ & Sơ đồ Tuần tự (Sequence Diagrams)

### 4.1 Luồng Thêm Hoá Đơn Nhập Hàng & Lưu Log (Import Goods Flow with Log)

Khi có hàng nhập về, luồng xử lý sẽ ghi nhận vào bảng Hoá đơn, tăng tồn kho hiện tại và chèn log vào bảng `inventory_log` để kiểm soát lịch sử.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as InventoryController
    participant U as InventoryUsecase
    participant R as InventoryRepository
    participant DB as PostgreSQL Database

    Admin->>C: POST /api/v1/admin/inventory/import [Payload chi tiết nhập]
    activate C
    Note over C: Lấy Admin ID từ JWT token

    C->>U: ImportGoods(ctx, adminUserID, &req)
    activate U
    U->>R: CreateImportInvoice(ctx, adminUserID, invoice, details)
    activate R
    
    R->>DB: Begin(ctx) (Bắt đầu Transaction)
    activate DB
    
    R->>DB: 1. INSERT INTO import_invoices (...) RETURNING id
    DB-->>R: Nhận InvoiceID
    
    loop Duyệt từng mặt hàng nhập
        R->>DB: 2. SELECT quantity FROM product_inventory FOR UPDATE (Khoá dòng)
        DB-->>R: Trả về quantity hiện tại (stock_before)
        Note over R: Nếu chưa có tồn kho, mặc định = 0
        
        R->>DB: 3. INSERT INTO import_invoice_details (...) (Lưu chi tiết)
        
        Note over R: Tính qty_after = stock_before + import_qty
        
        R->>DB: 4. UPSERT INTO product_inventory (Cập nhật số lượng mới)
        
        R->>DB: 5. INSERT INTO inventory_log (variant_id, store_id, change_qty, qty_after, reason='import', ref_id=InvoiceID, created_by)
    end
    
    R->>DB: Commit(ctx)
    DB-->>R: Transaction Committed
    deactivate DB
    
    R-->>U: Trả về *domain.ImportInvoice
    deactivate R
    U-->>C: Trả về *domain.ImportInvoice, nil
    deactivate U
    C-->>Admin: HTTP 201 Created
    deactivate C
```

---

### 4.2 Luồng Điều chỉnh Tồn kho thủ công & Ghi Log (Manual Stock Adjust Flow)

CMS cung cấp giao diện cho Admin điều chỉnh lại tồn kho khi đối soát kiểm kho thực tế. Hệ thống so sánh số lượng mới và cũ để tính ra hiệu số `change_qty` và ghi nhận log `manual_adjust`.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as InventoryController
    participant U as InventoryUsecase
    participant R as InventoryRepository
    participant DB as PostgreSQL Database

    Admin->>C: PUT /api/v1/admin/stores/:id/inventory [Mảng Variant + Số lượng thực tế]
    activate C
    C->>U: AdjustInventory(ctx, storeID, adminUserID, &req)
    activate U
    U->>R: AdjustInventory(ctx, storeID, adminUserID, adjustments)
    activate R
    
    R->>DB: Begin(ctx) (Bắt đầu Transaction)
    activate DB
    
    loop Duyệt từng variant cần điều chỉnh
        R->>DB: 1. SELECT quantity FROM product_inventory WHERE variant_id=$1 AND store_id=$2 FOR UPDATE
        DB-->>R: Số lượng cũ (old_qty)
        
        Note over R: Tính hiệu số change_qty = new_qty - old_qty
        
        R->>DB: 2. UPSERT INTO product_inventory (Set quantity = new_qty)
        
        R->>DB: 3. INSERT INTO inventory_log (change_qty, qty_after = new_qty, reason='manual_adjust', created_by)
    end
    
    R->>DB: Commit(ctx)
    DB-->>R: Transaction Committed
    deactivate DB
    
    R-->>U: Thành công
    deactivate R
    U-->>C: Thành công
    deactivate U
    C-->>Admin: HTTP 200 OK (Cập nhật hoàn tất)
    deactivate C
```

---

### 4.3 Luồng Cảnh Báo Hàng Thấp (Low Stock Alert Flow)

Hệ thống so sánh số lượng còn lại trong kho của từng phiên bản tại từng cửa hàng (`product_inventory.quantity`) với ngưỡng cảnh báo hàng thấp (`product.low_stock_threshold`) để gửi cảnh báo.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as InventoryController
    participant U as InventoryUsecase
    participant R as InventoryRepository
    participant DB as PostgreSQL Database

    Admin->>C: GET /api/v1/admin/inventory/low-stock?store_id=1
    activate C
    C->>U: GetLowStockAlerts(ctx, storeIDOpt)
    activate U
    U->>R: GetLowStockAlerts(ctx, storeIDOpt)
    activate R
    Note over R: SELECT JOIN product, variant, store WHERE quantity <= low_stock_threshold
    R->>DB: SELECT pi.variant_id, pi.quantity, p.low_stock_threshold, pv.name as variant_name...
    activate DB
    DB-->>R: Danh sách các dòng hàng sắp hết
    deactivate DB
    R-->>U: Trả về []*domain.LowStockAlertResponse
    deactivate R
    U-->>C: Trả về []*domain.LowStockAlertResponse
    deactivate U
    C-->>Admin: HTTP 200 OK
    deactivate C
```

---

### 4.4 Luồng Xem Lịch sử Thay đổi Kho (Inventory Logs History Flow)

Cho phép quản trị viên xem chi tiết quá trình biến động kho, hỗ trợ phân trang và tìm kiếm theo cửa hàng hoặc biến thể sản phẩm.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as InventoryController
    participant U as InventoryUsecase
    participant R as InventoryRepository
    participant DB as PostgreSQL Database

    Admin->>C: GET /api/v1/admin/inventory/logs?store_id=1&page=1&limit=20
    activate C
    C->>U: GetInventoryLogs(ctx, &filterQuery)
    activate U
    U->>R: GetInventoryLogs(ctx, &filterQuery)
    activate R
    Note over R: SELECT JOIN store, variant, users ORDER BY created_at DESC
    R->>DB: SELECT il.id, il.change_qty, il.qty_after, pv.name, u.full_name... LIMIT $1 OFFSET $2
    activate DB
    DB-->>R: Danh sách logs + TotalCount
    deactivate DB
    R-->>U: Trả về *domain.InventoryLogsResult
    deactivate R
    U-->>C: Trả về *domain.InventoryLogsResult
    deactivate U
    C-->>Admin: HTTP 200 OK
    deactivate C
```

---

## 5. Đặc Tả Lớp & Cấu Trúc Dữ Liệu Cập Nhật (UML Class Diagram)

```mermaid
classDiagram
    direction BT

    class Store {
        +int ID
        +string Name
        +string Hotline
        +string District
        +string Province
        +string Ward
        +string Road
        +string Email
        +float64* Lat
        +float64* Lng
        +bool IsActive
    }

    class InventoryLog {
        +int ID
        +int VariantID
        +int StoreID
        +int ChangeQty
        +int QtyAfter
        +string Reason
        +string RefID
        +int CreatedBy
        +Time CreatedAt
    }

    class ImportInvoiceRequest {
        +int SupplierID
        +int StoreID
        +string Note
        +ImportItemDTO[] Items
    }

    class ManualAdjustRequest {
        +AdjustItemDTO[] Adjustments
    }

    class AdjustItemDTO {
        +int VariantID
        +int NewQuantity
    }

    class InventoryLogsQuery {
        +int* StoreID
        +int* VariantID
        +string Reason
        +int Page
        +int Limit
    }

    class Supplier {
        +int ID
        +string Name
        +string* Address
        +string* Phone
        +bool IsDeleted
    }

    class InventoryController {
        -InventoryUsecase usecase
        +CreateStore(ctx: *gin.Context) void
        +ListStores(ctx: *gin.Context) void
        +CreateSupplier(ctx: *gin.Context) void
        +ListSuppliers(ctx: *gin.Context) void
        +UpdateSupplier(ctx: *gin.Context) void
        +DeleteSupplier(ctx: *gin.Context) void
        +ImportGoods(ctx: *gin.Context) void
        +AdjustInventory(ctx: *gin.Context) void
        +ListStoreInventory(ctx: *gin.Context) void
        +GetLowStockAlerts(ctx: *gin.Context) void
        +GetInventoryLogs(ctx: *gin.Context) void
    }

    class InventoryUsecase {
        <<interface>>
        +CreateStore(ctx: Context, s: *Store) (*Store, error)
        +ListStores(ctx: Context, province: string, district: string) ([]*Store, error)
        +CreateSupplier(ctx: Context, req: *CreateSupplierRequest) (*Supplier, error)
        +ListSuppliers(ctx: Context) ([]*Supplier, error)
        +UpdateSupplier(ctx: Context, id: int, req: *UpdateSupplierRequest) (*Supplier, error)
        +DeleteSupplier(ctx: Context, id: int) error
        +ImportGoods(ctx: Context, creatorID: int, req: *ImportInvoiceRequest) (*ImportInvoice, error)
        +AdjustInventory(ctx: Context, storeID: int, creatorID: int, req: *ManualAdjustRequest) error
        +GetLowStockAlerts(ctx: Context, storeID *int) ([]*LowStockAlertResponse, error)
        +GetInventoryLogs(ctx: Context, q: *InventoryLogsQuery) (*InventoryLogsResult, error)
    }

    class InventoryRepository {
        <<interface>>
        +CreateStore(ctx: Context, s: *Store) (*Store, error)
        +ListStores(ctx: Context, province: string, district: string) ([]*Store, error)
        +GetStoreByID(ctx: Context, id: int) (*Store, error)
        +CreateSupplier(ctx: Context, s: *Supplier) (*Supplier, error)
        +ListSuppliers(ctx: Context) ([]*Supplier, error)
        +GetSupplierByID(ctx: Context, id: int) (*Supplier, error)
        +UpdateSupplier(ctx: Context, s: *Supplier) (*Supplier, error)
        +DeleteSupplier(ctx: Context, id: int) error
        +CreateImportInvoice(ctx: Context, creatorID: int, invoice: *ImportInvoice, details: []*ImportInvoiceDetail) (*ImportInvoice, error)
        +AdjustInventory(ctx: Context, storeID: int, creatorID: int, adjustments: []*AdjustItemDTO) error
        +GetLowStockAlerts(ctx: Context, storeID *int) ([]*LowStockAlertResponse, error)
        +GetInventoryLogs(ctx: Context, q: *InventoryLogsQuery) (*InventoryLogsResult, error)
    }

    InventoryController --> InventoryUsecase : calls
    InventoryUsecase --> InventoryRepository : calls
    InventoryController ..> ImportInvoiceRequest : binds JSON
    InventoryController ..> ManualAdjustRequest : binds JSON
    InventoryController ..> InventoryLogsQuery : binds Query Params
```
