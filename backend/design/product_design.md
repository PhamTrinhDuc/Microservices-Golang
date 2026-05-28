# Thiết Kế Chi Tiết Luồng CRUD Sản Phẩm (Product Catalog Design)
*Tham chiếu mô hình Catalog của Thế Giới Di Động (TGDD)*

Tài liệu này mô tả chi tiết thiết kế hệ thống luồng **CRUD** (Create, Read, Update, Delete) cho thực thể **Sản phẩm (Product)** sau khi đối chiếu và tối ưu theo mô hình thực tế của Thế Giới Di Động (ví dụ trang [iPhone 17 thường](https://www.thegioididong.com/dtdd/iphone-17)).

---

## 1. Phân Tích Mô Hình Catalog Thực Tế (Thế Giới Di Động)

Khi phân tích trang sản phẩm của Thế Giới Di Động, ta thấy hành vi hệ thống được tổ chức như sau:

```
[Dòng sản phẩm gốc: iPhone 17 thường]
   |
   +---> [Sản phẩm bán 1: iPhone 17 128GB] (URL: /dtdd/iphone-17-128gb) -> Specs: 128GB ROM
   |        |
   |        +---> [Biến thể màu 1: Xanh Lá] (Giá: A)
   |        +---> [Biến thể màu 2: Xanh Lam] (Giá: B)
   |
   +---> [Sản phẩm bán 2: iPhone 17 256GB] (URL: /dtdd/iphone-17) -> Specs: 256GB ROM
   |        |
   |        +---> [Biến thể màu 1: Xanh Lá] (Giá: C)
   |        +---> [Biến thể màu 2: Xanh Lam] (Giá: D)
   |
   +---> [Sản phẩm bán 3: iPhone 17 512GB] (URL: /dtdd/iphone-17-512gb) -> Specs: 512GB ROM
```

### Cách TGDD phân cấp và tổ chức:
1. **Sản phẩm bán chính (Product Entity)**: 
   - Mỗi cấu hình phần cứng khác nhau (ví dụ: `128GB`, `256GB`, `512GB`) được coi là **một trang sản phẩm độc lập (bản ghi Product riêng biệt)**.
   - Mỗi sản phẩm này có:
     - URL (Slug) riêng biệt (ví dụ: `/iphone-17` cho bản 256GB và `/iphone-17-512gb` cho bản 512GB).
     - ID riêng, tiêu đề SEO riêng, mô tả đặc điểm nổi bật riêng.
     - Bộ thông số kỹ thuật (Specs) riêng (ví dụ bản 128GB thì ROM hiển thị là 128GB).
   - Trên giao diện của trang này, các nút chuyển đổi dung lượng (128GB, 256GB...) thực chất là **đường link chuyển trang (Redirect)** sang sản phẩm tương ứng.
2. **Thuộc tính tùy chọn (Product Options)**:
   - Trong cùng một trang sản phẩm (ví dụ: `iPhone 17 256GB`), hệ thống cho phép chọn giữa các **Màu sắc** khác nhau (Xanh Lá Xô Thơm, Xanh Lam Khói, Tím Oải Hương, Trắng).
   - Khi click chọn Màu sắc, **URL không thay đổi**, nhưng ảnh sản phẩm, giá bán, và khuyến mãi sẽ thay đổi tương ứng.
3. **Biến thể cụ thể (Product Variant)**:
   - Các lựa chọn Màu sắc chính là các **Biến thể (Variant)** của sản phẩm bán chính.
   - Đây là thực thể có SKU cụ thể và quản lý tồn kho (`product_inventory`), cũng như liên kết trực tiếp với Đơn hàng (`order_details`).

---

## 2. Ánh Xạ Vào Cấu Trúc Database Hiện Tại

Thiết kế database trong dự án của bạn **hoàn toàn tương thích** với mô hình chuẩn này:

* **Bảng `product`**: Lưu trữ Sản phẩm bán chính (Ví dụ: `id` = `"iphone-17-256gb"`, `name` = `"iPhone 17 256GB"`, `slug` = `"iphone-17"`).
* **Bảng `product_spec`**: Lưu thông số kỹ thuật riêng của sản phẩm bán đó (Ví dụ: ROM = 256GB).
* **Bảng `product_option_type` và `product_option_value`**: Lưu trữ thuộc tính thay đổi trực tiếp trên trang. Đối với mô hình TGDD, đây là nhóm `"Màu sắc"` (Đỏ, Xanh...).
* **Bảng `product_variant`**: Lưu trữ từng phiên bản Màu cụ thể (Ví dụ: `"iPhone 17 256GB - Xanh Lá"`, SKU = `"IP17-256G-GREEN"`, Giá = `24.990.000đ`).
* **Bảng `product_variant_option`**: Liên kết biến thể với giá trị màu.

---

## 3. Sơ đồ các Lớp Thiết Kế (UML Class Diagram - Toàn bộ CRUD)

Dưới đây là thiết kế chi tiết cho cấu trúc Class trong Layer Clean Architecture để thực hiện các nghiệp vụ CRUD cho Product:

```mermaid
classDiagram
    direction BT
    
    %% DTO structures %%
    class CreateProductRequest {
        +string ID
        +int CategoryID
        +int BrandID
        +string Name
        +string Slug
        +ProductSpecDTO[] Specs
        +OptionDTO[] Options
        +VariantDTO[] Variants
    }
    
    class UpdateProductRequest {
        +int CategoryID
        +int BrandID
        +string Name
        +string Slug
        +ProductSpecDTO[] Specs
    }
    
    class ProductSearchQuery {
        +int* CategoryID
        +int* BrandID
        +string Query
        +int Page
        +int Limit
        +string Sort
    }
    
    class CreateProductInput {
        +Product Product
        +ProductSpec[] Specs
        +ProductOptionType[] OptionTypes
        +ProductVariant[] Variants
    }

    %% Interfaces and controllers %%
    class CatalogController {
        -CatalogUsecase usecase
        -Validate validator
        +CreateProduct(ctx: *gin.Context) void
        +SearchProducts(ctx: *gin.Context) void
        +GetProductDetails(ctx: *gin.Context) void
        +UpdateProduct(ctx: *gin.Context) void
        +DeleteProduct(ctx: *gin.Context) void
    }

    class CatalogUsecase {
        <<interface>>
        +CreateProduct(ctx: Context, req: *CreateProductRequest) (*Product, error)
        +SearchProducts(ctx: Context, q: *ProductSearchQuery) (*ProductSearchResult, error)
        +GetProductDetails(ctx: Context, id: string) (*ProductDetailsResponse, error)
        +UpdateProduct(ctx: Context, id: string, req: *UpdateProductRequest) (*Product, error)
        +DeleteProduct(ctx: Context, id: string) error
    }

    class CatalogRepository {
        <<interface>>
        +CreateProduct(ctx: Context, input: *CreateProductInput) (*Product, error)
        +SearchProducts(ctx: Context, q: *ProductSearchQuery) (*ProductSearchResult, error)
        +GetProductDetails(ctx: Context, id: string) (*ProductDetailsResponse, error)
        +UpdateProduct(ctx: Context, prod: *Product, specs: []*ProductSpec) (*Product, error)
        +DeleteProduct(ctx: Context, id: string) error
        +GetCategoryByID(ctx: Context, id: int) (*Category, error)
        +GetBrandByID(ctx: Context, id: int) (*Brand, error)
    }

    CatalogController --> CatalogUsecase : calls
    CatalogUsecase --> CatalogRepository : calls
    CatalogController ..> CreateProductRequest : binds JSON
    CatalogController ..> UpdateProductRequest : binds JSON
    CatalogController ..> ProductSearchQuery : binds Query Params
```

---

## 4. Thiết Kế Luồng Xử Lý Chi Tiết (Sequence Diagrams)

### 4.1 Luồng C-CREATE (Thêm sản phẩm đồng bộ)

Admin gửi 1 request gộp chứa thông tin chung, thông số kĩ thuật, các thuộc tính màu và các variant tương ứng.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as CatalogController<br>(controller/catalog.go)
    participant U as CatalogUsecase<br>(usecase/catalog.go)
    participant R as CatalogRepository<br>(repository/catalog.go)
    participant DB as PostgreSQL Database

    Admin->>C: POST /api/v1/admin/products<br>[JSON Payload]
    activate C
    Note over C: Binds JSON & Runs validator.Struct()

    C->>U: CreateProduct(ctx, &req)
    activate U
    Note over U: 1. Kiểm tra CategoryID & BrandID tồn tại<br>2. Tự sinh Slug nếu trống<br>3. Kiểm tra trùng lặp Slug<br>4. Ánh xạ dữ liệu sang domain.CreateProductInput

    U->>R: CreateProduct(ctx, &input)
    activate R
    
    R->>DB: Begin(ctx) (Bắt đầu Transaction)
    activate DB
    R->>DB: 1. INSERT INTO product (...)
    DB-->>R: Trả về thành công
    
    loop Lưu Specs
        R->>DB: 2. INSERT INTO product_spec (...)
    end
    
    loop Lưu Option Types & Values
        R->>DB: 3. INSERT INTO product_option_type (...) RETURNING id
        DB-->>R: Nhận OptionTypeID
        loop Lưu Values của Option
            R->>DB: 4. INSERT INTO product_option_value (...) RETURNING id
            DB-->>R: Nhận OptionValueID
            Note over R: Map "OptionName:Value" -> DB ID
        end
    end

    loop Lưu Variants & Variant Options
        R->>DB: 5. INSERT INTO product_variant (...) RETURNING id
        DB-->>R: Nhận VariantID
        loop Lưu Option của Variant
            Note over R: Lấy ID từ Map
            R->>DB: 6. INSERT INTO product_variant_option (variant_id, option_value_id)
        end
    end

    R->>DB: Commit(ctx)
    DB-->>R: Transaction Committed
    deactivate DB
    
    R-->>U: Trả về *domain.Product
    deactivate R
    U-->>C: Trả về *domain.Product, nil
    deactivate U
    C-->>Admin: HTTP 201 Created
    deactivate C
```

---

### 4.2 Luồng R-READ (Lọc & Tìm kiếm sản phẩm + Chi tiết sản phẩm)

#### 4.2.1 Luồng Tìm kiếm & Phân trang (`GET /products`)

```mermaid
sequenceDiagram
    autonumber
    actor Customer as Khách hàng
    participant C as CatalogController
    participant U as CatalogUsecase
    participant R as CatalogRepository
    participant DB as PostgreSQL Database

    Customer->>C: GET /api/v1/products?q=iphone&page=1&limit=10
    activate C
    Note over C: Binds Query Params to domain.ProductSearchQuery

    C->>U: SearchProducts(ctx, &query)
    activate U
    U->>R: SearchProducts(ctx, &query)
    activate R

    Note over R: Xây dựng Dynamic SQL Query phòng chống SQL Injection
    R->>DB: 1. SELECT COUNT(*) WHERE is_active=true AND is_deleted=false ... (Count)
    activate DB
    DB-->>R: Trả về tổng số dòng (TotalCount)
    
    R->>DB: 2. SELECT ... LIMIT $1 OFFSET $2 (Get Paginated Rows)
    DB-->>R: Trả về danh sách Product rows
    deactivate DB

    R-->>U: Trả về *domain.ProductSearchResult
    deactivate R
    U-->>C: Trả về *domain.ProductSearchResult, nil
    deactivate U
    C-->>Customer: HTTP 200 OK<br>{ data: ProductSearchResult }
    deactivate C
```

#### 4.2.2 Luồng Lấy Chi tiết Sản phẩm (`GET /products/:id`)

Nhằm hiển thị trang chi tiết tương tự TGDD, API chi tiết cần cung cấp đầy đủ thông tin: thông số kĩ thuật, các tuỳ chọn màu, các biến thể màu cụ thể và danh sách sản phẩm cùng nhóm (Ví dụ: iPhone 17 128GB, iPhone 17 512GB) để khách chuyển dung lượng.

```mermaid
sequenceDiagram
    autonumber
    actor Customer as Khách hàng
    participant C as CatalogController
    participant U as CatalogUsecase
    participant R as CatalogRepository
    participant DB as PostgreSQL Database

    Customer->>C: GET /api/v1/products/iphone-17-256gb
    activate C
    C->>U: GetProductDetails(ctx, idOrSlug)
    activate U
    U->>R: GetProductDetails(ctx, idOrSlug)
    activate R

    R->>DB: 1. SELECT product + join brand & category (Lấy thông tin chung)
    activate DB
    DB-->>R: Product Row
    
    R->>DB: 2. SELECT FROM product_spec (Lấy thông số kĩ thuật)
    DB-->>R: Specs Rows
    
    R->>DB: 3. SELECT option_type + values (Lấy nhóm màu sắc)
    DB-->>R: Option Types & Values Rows
    
    R->>DB: 4. SELECT FROM product_variant (Lấy các biến thể màu)
    DB-->>R: Variant Rows

    R->>DB: 5. SELECT FROM product_variant_option + value (Lấy liên kết variant và màu)
    DB-->>R: Variant-Option Mapping Rows
    
    R->>DB: 6. SELECT FROM product_image (Lấy ảnh)
    DB-->>R: Image Rows
    deactivate DB

    Note over R: Kết hợp các mảng dữ liệu vào Struct ProductDetailsResponse

    R-->>U: Trả về *domain.ProductDetailsResponse
    deactivate R
    
    Note over U: Lấy thêm các sản phẩm cùng nhóm để vẽ nút chuyển dung lượng:<br>Tìm các sản phẩm cùng Category, Brand và có tên tương tự (ILIKE "iPhone 17%")
    U->>R: SearchProducts(ctx, &siblingQuery)
    activate R
    R-->>U: Sibling products list
    deactivate R

    U-->>C: Trả về ProductDetailsResponse (kèm sibling products)
    deactivate U
    C-->>Customer: HTTP 200 OK
    deactivate C
```

---

### 4.3 Luồng U-UPDATE (Cập nhật sản phẩm & Specs)

Admin thực hiện chỉnh sửa thông tin sản phẩm gốc và cập nhật lại danh sách thông số kĩ thuật (Specs).

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as CatalogController
    participant U as CatalogUsecase
    participant R as CatalogRepository
    participant DB as PostgreSQL Database

    Admin->>C: PUT /api/v1/admin/products/iphone-17-256gb<br>[JSON Payload]
    activate C
    Note over C: Binds JSON to domain.UpdateProductRequest

    C->>U: UpdateProduct(ctx, id, &req)
    activate U
    Note over U: 1. Kiểm tra sản phẩm tồn tại trong hệ thống<br>2. Kiểm tra CategoryID & BrandID mới tồn tại<br>3. Ánh xạ sang Domain Product & Specs

    U->>R: UpdateProduct(ctx, &prod, specs)
    activate R
    
    R->>DB: Begin(ctx) (Bắt đầu Transaction)
    activate DB
    
    R->>DB: 1. UPDATE product SET name=$1, category_id=$2, ... WHERE id=$3
    DB-->>R: Thành công
    
    R->>DB: 2. DELETE FROM product_spec WHERE product_id = $1 (Xóa Specs cũ)
    DB-->>R: Thành công
    
    loop Lưu Specs mới
        R->>DB: 3. INSERT INTO product_spec (...)
    end
    DB-->>R: Thành công
    
    R->>DB: Commit(ctx)
    DB-->>R: Transaction Committed
    deactivate DB

    R-->>U: Trả về *domain.Product
    deactivate R
    U-->>C: Trả về *domain.Product, nil
    deactivate U
    C-->>Admin: HTTP 200 OK
    deactivate C
```

---

### 4.4 Luồng D-DELETE (Xóa mềm Sản phẩm & Biến thể)

Admin yêu cầu xóa sản phẩm. Hệ thống tự động xóa mềm sản phẩm và toàn bộ các biến thể thuộc sản phẩm đó để bảo toàn lịch sử đơn hàng.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin CMS
    participant C as CatalogController
    participant U as CatalogUsecase
    participant R as CatalogRepository
    participant DB as PostgreSQL Database

    Admin->>C: DELETE /api/v1/admin/products/iphone-17-256gb
    activate C
    C->>U: DeleteProduct(ctx, id)
    activate U
    Note over U: Kiểm tra sản phẩm có tồn tại không

    U->>R: DeleteProduct(ctx, id)
    activate R
    
    R->>DB: Begin(ctx) (Bắt đầu Transaction)
    activate DB
    
    R->>DB: 1. UPDATE product SET is_deleted=true, updated_at=NOW() WHERE id=$1
    DB-->>R: Thành công
    
    R->>DB: 2. UPDATE product_variant SET is_deleted=true WHERE product_id=$1
    DB-->>R: Thành công
    
    R->>DB: Commit(ctx)
    DB-->>R: Transaction Committed
    deactivate DB

    R-->>U: Trả về nil (Không lỗi)
    deactivate R
    U-->>C: Trả về nil
    deactivate U
    C-->>Admin: HTTP 204 No Content
    deactivate C
```

---

## 6. Mô Tả Chi Tiết Phương Thức cho Các Lớp (Tất cả luồng)

### 6.1 Lớp API Controller (`controller/catalog.go`)
* **CreateProduct**: Phục vụ luồng C. Parse JSON payload của `CreateProductRequest`, chạy validator và chuyển xuống Usecase.
* **SearchProducts**: Phục vụ luồng R. Đọc query params (`q`, `category_id`, `brand_id`, `page`, `limit`, `sort`), ép kiểu và gọi Usecase.
* **GetProductDetails**: Phục vụ luồng R. Lấy `id` từ Param Path (Ví dụ: `iphone-17-256gb`), gọi Usecase và trả về payload chi tiết.
* **UpdateProduct**: Phục vụ luồng U. Lấy `id` từ Param Path, parse JSON payload `UpdateProductRequest`, gọi Usecase.
* **DeleteProduct**: Phục vụ luồng D. Lấy `id` từ Param Path, gọi Usecase, trả về Status `StatusNoContent` (204).

---

### 6.2 Lớp Nghiệp Vụ (`usecase/catalog.go`)

* **SearchProducts**:
  - Giao tiếp với Repository để lấy danh sách sản phẩm phân trang.
* **GetProductDetails**:
  - Gọi `repo.GetProductDetails(ctx, id)`.
  - Để hỗ trợ tính năng chuyển dung lượng giống TGDD, Usecase tiến hành tìm kiếm các sản phẩm cùng nhóm: Gọi `repo.SearchProducts(ctx, &siblingQuery)` với điều kiện tên tương tự và cùng Category/Brand. Danh sách này được đính kèm vào response trả về cho Client.
* **UpdateProduct**:
  - Lấy thông tin sản phẩm hiện tại để đảm bảo sự tồn tại.
  - Kiểm tra `category_id` và `brand_id` mới có tồn tại trong hệ thống hay không.
  - Kiểm tra tính độc bản của Slug mới (nếu slug thay đổi).
  - Ánh xạ sang domain struct và gọi `repo.UpdateProduct`.
* **DeleteProduct**:
  - Gọi `repo.DeleteProduct(ctx, id)`.

---

### 6.3 Lớp Truy Xuất Cơ Sở Dữ Liệu (`repository/catalog.go`)

Mọi câu lệnh truy xuất đều kế thừa `context.Context` để tránh treo kết nối khi client ngắt kết nối.

* **SearchProducts**:
  - Thực hiện build SQL động dựa trên số lượng tham số truyền vào:
    ```sql
    SELECT COUNT(*) FROM product WHERE is_deleted=false AND is_active=true [AND category_id = $1 ...]
    ```
    và câu truy vấn phân trang:
    ```sql
    SELECT id, category_id, brand_id, name, slug, img_thumb, ... 
    FROM product 
    WHERE is_deleted=false AND is_active=true [AND category_id = $1 ...] 
    ORDER BY created_at DESC LIMIT $limit OFFSET $offset
    ```
* **GetProductDetails**:
  - Thực hiện 6 câu truy vấn riêng biệt (hoặc join chọn lọc) trong connection pool để gom dữ liệu của sản phẩm, specs, options, variants, variant-options, và images.
  - Phân tích và nhóm dữ liệu trong Go: gom các `product_option_value` vào đúng `product_option_type` dựa trên ID nhóm, và gắn mảng lựa chọn thuộc tính tương ứng vào từng `product_variant`.
* **UpdateProduct (Transaction)**:
  - Bắt đầu transaction.
  - Chạy `UPDATE product SET ... WHERE id = $id`.
  - Chạy `DELETE FROM product_spec WHERE product_id = $id`.
  - Duyệt qua mảng Specs mới và chạy `INSERT INTO product_spec (product_id, "group", key, value, unit, sort_order) VALUES (...)`.
  - Commit transaction.
* **DeleteProduct (Transaction)**:
  - Bắt đầu transaction.
  - Chạy `UPDATE product SET is_deleted = true, updated_at = NOW() WHERE id = $id`.
  - Chạy `UPDATE product_variant SET is_deleted = true WHERE product_id = $id`.
  - Commit transaction.
