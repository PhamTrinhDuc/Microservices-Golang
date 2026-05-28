# Thiết Kế Chi Tiết Luồng Giỏ Hàng (Cart Module Design)

Tài liệu này mô tả chi tiết thiết kế hệ thống luồng dữ liệu nghiệp vụ cho module **Giỏ hàng (Cart)**, hỗ trợ cả người dùng vãng lai (Guest - sử dụng `session_id`) và người dùng đã đăng nhập (Authenticated User - sử dụng `user_id`), kèm cơ chế đồng bộ (merge) giỏ hàng khi đăng nhập.

---

## 1. Cấu Trúc Cơ Sở Dữ Liệu (Schema)

Bảng `cart_items` đã được định nghĩa sẵn trong cơ sở dữ liệu:

```sql
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(100),
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Chỉ mục tối ưu hóa truy vấn
CREATE INDEX idx_cart_items_user ON cart_items(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_cart_items_session ON cart_items(session_id) WHERE session_id IS NOT NULL;
```

> [!NOTE]
> Bảng `cart_items` có `user_id` và `session_id` đều nullable. 
> - Nếu người dùng chưa đăng nhập: `session_id` được điền, `user_id` = NULL.
> - Nếu người dùng đã đăng nhập: `user_id` được điền, `session_id` = NULL.

---

## 2. Các API Endpoints

Mọi endpoint hỗ trợ cả Guest (qua header `X-Session-ID` hoặc query param `session_id`) và User đã đăng nhập (thông qua JWT Token).

* `GET /api/v1/cart` - Xem giỏ hàng hiện tại (kèm thông tin chi tiết Variant, tên Product, hình ảnh, giá bán, base price).
* `POST /api/v1/cart` - Thêm sản phẩm (variant) vào giỏ hàng.
  * *Request Body*: `{"variant_id": 12, "quantity": 2, "session_id": "optional-guest-uuid"}`
* `PUT /api/v1/cart/items/:id` - Cập nhật số lượng của một mục trong giỏ hàng.
  * *Request Body*: `{"quantity": 5}`
* `DELETE /api/v1/cart/items/:id` - Xóa một mục khỏi giỏ hàng.
* `DELETE /api/v1/cart` - Xóa sạch toàn bộ giỏ hàng.
* `POST /api/v1/cart/merge` - Trộn/Đồng bộ giỏ hàng vãng lai vào tài khoản sau khi đăng nhập (Chỉ yêu cầu Authenticated).
  * *Request Body*: `{"session_id": "guest-uuid"}`

---

## 3. Thiết Kế Luồng Nghiệp Vụ & Sơ đồ Tuần tự (Sequence Diagrams)

### 3.1 Luồng Thêm Vào Giỏ Hàng (Add To Cart Flow)

Khi thêm sản phẩm vào giỏ hàng, hệ thống kiểm tra sự tồn tại của Variant. Nếu sản phẩm đã có trong giỏ hàng của user/session đó, hệ thống sẽ cộng dồn số lượng. Ngược lại, tạo mới một dòng.

```mermaid
sequenceDiagram
    autonumber
    actor Client as Client App
    participant C as CartController
    participant U as CartUsecase
    participant R as CartRepository
    participant DB as PostgreSQL

    Client->>C: POST /api/v1/cart [Payload: variant_id, quantity]
    activate C
    Note over C: Lấy UserID từ JWT (nếu có)<br/>hoặc X-Session-ID từ Header (nếu chưa đăng nhập)

    C->>U: AddToCart(ctx, userIDOpt, sessionIDOpt, req)
    activate U
    
    U->>R: GetVariantByID(ctx, req.variant_id) (Gọi chéo Catalog)
    R-->>U: Trả về Variant (Nếu không tồn tại -> báo lỗi)
    
    U->>R: FindCartItem(ctx, userIDOpt, sessionIDOpt, req.variant_id)
    activate R
    R->>DB: SELECT * FROM cart_items WHERE ...
    DB-->>R: Trả về CartItem hoặc pgx.ErrNoRows
    deactivate R
    
    alt CartItem đã tồn tại
        Note over U: Tính tổng số lượng mới = quantity_old + quantity_added
        U->>R: UpdateCartItemQuantity(ctx, item.ID, new_qty)
        R->>DB: UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2
    else CartItem chưa tồn tại
        U->>R: CreateCartItem(ctx, &CartItem)
        R->>DB: INSERT INTO cart_items (user_id, session_id, variant_id, quantity) VALUES (...)
    end
    
    U-->>C: Trả về CartItem thành công
    deactivate U
    C-->>Client: HTTP 200 OK / 201 Created
    deactivate C
```

---

### 3.2 Luồng Đồng bộ Giỏ hàng (Merge Cart Flow)

Sau khi đăng nhập thành công, Client sẽ gọi API `/api/v1/cart/merge` gửi kèm `session_id` cũ. Hệ thống sẽ gộp tất cả các sản phẩm trong giỏ hàng guest sang giỏ hàng của user.

```mermaid
sequenceDiagram
    autonumber
    actor Client as Client App
    participant C as CartController
    participant U as CartUsecase
    participant R as CartRepository
    participant DB as PostgreSQL

    Client->>C: POST /api/v1/cart/merge [session_id: "guest-uuid"] (JWT Token)
    activate C
    Note over C: Yêu cầu bắt buộc Authenticated để lấy UserID

    C->>U: MergeCart(ctx, userID, req.session_id)
    activate U
    
    R->>DB: Begin Transaction
    
    U->>R: ListCartItems(ctx, nil, req.session_id) (Lấy giỏ hàng Guest)
    R-->>U: Danh sách guest items
    
    U->>R: ListCartItems(ctx, userID, nil) (Lấy giỏ hàng User)
    R-->>U: Danh sách user items
    
    loop Duyệt qua từng Guest Item
        alt Trùng Variant với User Item nào đó
            Note over U: Tính tổng qty = guest_qty + user_qty
            U->>R: UpdateCartItemQuantity(ctx, user_item.id, total_qty)
            U->>R: DeleteCartItem(ctx, guest_item.id)
        else Không trùng Variant
            Note over U: Chuyển đổi quyền sở hữu
            U->>R: LinkGuestItemToUser(ctx, guest_item.id, userID)
            R->>DB: UPDATE cart_items SET user_id = $1, session_id = NULL WHERE id = $2
        end
    end
    
    R->>DB: Commit Transaction
    U-->>C: Đồng bộ hoàn tất
    deactivate U
    C-->>Client: HTTP 200 OK (Gộp giỏ hàng thành công)
    deactivate C
```

---

## 4. Đặc Tả Lớp & Cấu Trúc Dữ Liệu (UML Class Diagram)

```mermaid
classDiagram
    direction BT

    class CartItem {
        +int ID
        +int* UserID
        +string* SessionID
        +int VariantID
        +int Quantity
        +Time CreatedAt
        +Time UpdatedAt
    }

    class CartItemResponse {
        +int ID
        +int VariantID
        +string VariantName
        +string SKU
        +float64 Price
        +float64* PriceBase
        +string ProductID
        +string ProductName
        +string* ImageURL
        +int Quantity
    }

    class AddToCartRequest {
        +int VariantID
        +int Quantity
        +string* SessionID
    }

    class UpdateQuantityRequest {
        +int Quantity
    }

    class MergeCartRequest {
        +string SessionID
    }

    class CartController {
        -CartUsecase usecase
        +GetCart(ctx: *gin.Context) void
        +AddToCart(ctx: *gin.Context) void
        +UpdateItemQuantity(ctx: *gin.Context) void
        +RemoveItem(ctx: *gin.Context) void
        +ClearCart(ctx: *gin.Context) void
        +MergeCart(ctx: *gin.Context) void
    }

    class CartUsecase {
        <<interface>>
        +GetCart(ctx: Context, userID *int, sessionID *string) ([]*CartItemResponse, error)
        +AddToCart(ctx: Context, userID *int, sessionID *string, req: *AddToCartRequest) (*CartItem, error)
        +UpdateItemQuantity(ctx: Context, userID *int, sessionID *string, itemID: int, quantity: int) (*CartItem, error)
        +RemoveItem(ctx: Context, userID *int, sessionID *string, itemID: int) error
        +ClearCart(ctx: Context, userID *int, sessionID *string) error
        +MergeCart(ctx: Context, userID int, sessionID string) error
    }

    class CartRepository {
        <<interface>>
        +ListCartItems(ctx: Context, userID *int, sessionID *string) ([]*CartItem, error)
        +FindCartItem(ctx: Context, userID *int, sessionID *string, variantID int) (*CartItem, error)
        +GetCartItemByID(ctx: Context, id int) (*CartItem, error)
        +CreateCartItem(ctx: Context, item: *CartItem) (*CartItem, error)
        +UpdateCartItemQuantity(ctx: Context, id int, quantity int) (*CartItem, error)
        +LinkGuestItemToUser(ctx: Context, id int, userID int) error
        +DeleteCartItem(ctx: Context, id int) error
        +ClearCart(ctx: Context, userID *int, sessionID *string) error
        +GetCartDetails(ctx: Context, userID *int, sessionID *string) ([]*CartItemResponse, error)
    }

    CartController --> CartUsecase : calls
    CartUsecase --> CartRepository : calls
```
