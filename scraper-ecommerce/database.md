// ─── USER & ADDRESS ───────────────────────────────────────────

Table Users {
  id integer [primary key]
  full_name varchar
  email varchar
  password varchar
  phone varchar
  gender varchar
  dob date
  role varchar // "admin", "staff", "customer" — giữ varchar, đủ đơn giản
  avatar varchar
  is_lock bool
  is_verified bool
  created_at timestamp
  updated_at timestamp
}

Table Address {
  id serial [primary key]
  user_id integer [ref: > Users.id]
  full_name varchar
  phone varchar
  district varchar
  province varchar
  ward varchar
  detail_address varchar
  is_default bool
  is_deleted bool
}

// ─── STORE ────────────────────────────────────────────────────

Table Store {
  id serial [primary key]
  name varchar
  hotline varchar
  district varchar
  province varchar
  ward varchar
  road varchar
  email varchar
  lat numeric  // tọa độ — dùng cho tính năng "tìm cửa hàng gần tôi" (khuyên dùng PostGIS geography(Point, 4326))
  lng numeric
  is_active bool
  created_at timestamp
  updated_at timestamp
}

// ─── BRAND ────────────────────────────────────────────────────

Table Brand {  
  id serial [primary key]
  name varchar
  slug varchar [unique]
  logo_url varchar
  is_active bool [default: true]
  is_deleted bool [default: false]
}

// ─── PRODUCT CATALOG ──────────────────────────────────────────

Table Category {
  id serial [primary key]
  name varchar
  parent_id integer [ref: > Category.id] // null = danh mục gốc (Điện thoại, Laptop...)
  icon varchar
  slug varchar
  sort_order integer [default: 0] // thứ tự hiển thị trên menu
  is_deleted bool [default: false]
}

// ─── PRODUCT ──────────────────────────────────────────────────

Table Product {
  id varchar [primary key]
  category_id integer [ref: > Category.id]
  brand_id integer [ref: > Brand.id]
  name varchar
  slug varchar [unique]
  meta_title varchar
  meta_description text
  img_thumb varchar     // ảnh đại diện chính, dùng trên listing page
  weight numeric        // gram — fallback nếu variant không có weight riêng
  low_stock_threshold integer [default: 5] // cảnh báo sắp hết hàng
  specs_jsonb jsonb     // snapshot thông số kỹ thuật dạng jsonb để query/filter nhanh bằng GIN index
  is_active bool [default: true]
  is_deleted bool [default: false]
  created_at timestamp
  updated_at timestamp
}

// Thông số kỹ thuật dạng EAV — mỗi dòng là 1 thông số
// Ưu điểm: mỗi category có thông số khác nhau, không cần thêm cột
// Vd: { group: "Bộ nhớ", key: "RAM", value: "8", unit: "GB" }
Table ProductSpec {
  id serial [primary key]
  product_id varchar [ref: > Product.id]
  group varchar      // nhóm hiển thị: "Bộ nhớ", "Màn hình", "Pin"
  key varchar        // "RAM", "Dung lượng pin", "Kích thước màn hình"
  value varchar      // "8", "4685", "6.7"
  unit varchar       // "GB", "mAh", "inch" — nullable nếu không có đơn vị
  sort_order integer [default: 0]
}

// Ảnh sản phẩm tách riêng để quản lý thứ tự và gắn theo variant
// variant_id null = ảnh chung của product, không gắn với màu/phiên bản cụ thể
Table ProductImage {
  id serial [primary key]
  product_id varchar [ref: > Product.id]
  variant_id integer [ref: > ProductVariant.id]
  url varchar
  alt_text varchar
  sort_order integer [default: 0]
  is_thumbnail bool [default: false]
}

// ─── VARIANT ──────────────────────────────────────────────────

// Định nghĩa các "trục" lựa chọn của product
// Vd: iPhone 15 Pro Max có 2 trục: "Màu sắc" và "Dung lượng"
Table ProductOptionType {
  id serial [primary key]
  product_id varchar [ref: > Product.id]
  name varchar       // "Màu sắc", "Dung lượng", "RAM", "Ổ cứng"
  sort_order integer [default: 0]
}

// Các giá trị cụ thể của từng trục
// Vd: Màu sắc → ["Titan Tự Nhiên", "Titan Đen"], Dung lượng → ["256GB", "1TB"]
Table ProductOptionValue {
  id serial [primary key]
  option_type_id integer [ref: > ProductOptionType.id]
  value varchar       // "Titan Tự Nhiên", "256GB"
  color_code varchar  // nullable — chỉ dùng khi option type là màu sắc
  sort_order integer [default: 0]
}

// Mỗi variant = 1 tổ hợp option value có giá và tồn kho riêng
// Vd: "Titan Đen / 512GB" là 1 variant, "Titan Trắng / 256GB" là variant khác
Table ProductVariant {
  id serial [primary key]
  product_id varchar [ref: > Product.id]
  name varchar       // tên đầy đủ tổ hợp, vd: "Titan Đen / 512GB"
  sku varchar [unique]
  price numeric      // giá bán hiện tại (VND) — đồng nhất numeric
  price_base numeric // giá gốc — dùng để tính % giảm giá và hiển thị gạch ngang
  weight numeric     // nullable — nếu null thì fallback về Product.weight
  is_active bool [default: true]
  is_deleted bool [default: false]
}

// Bảng nối: variant này gồm những option value nào
Table ProductVariantOption {
  id serial [primary key]
  variant_id integer [ref: > ProductVariant.id]
  option_value_id integer [ref: > ProductOptionValue.id]
}

// ─── INVENTORY ────────────────────────────────────────────────

// Tách tồn kho ra khỏi ProductVariant vì:
// 1. Tồn kho thay đổi liên tục → tách ra tránh lock row variant
// 2. Mỗi store có tồn kho riêng
// Primary key là (variant_id, store_id) — mỗi cặp là duy nhất
Table ProductInventory {
  variant_id integer [ref: > ProductVariant.id]
  store_id integer [ref: > Store.id]
  quantity integer   // số lượng thực tế trong kho
  reserved integer   // đang bị giữ bởi InventoryReservations, chưa trừ hẳn
  last_updated timestamp

  indexes {
    (variant_id, store_id) [pk]
  }
}

// Giữ hàng tạm thời trong lúc user đang thanh toán — tránh oversell
// Khi thanh toán thành công → trừ quantity, xóa reservation
// Khi hết hạn expires_at hoặc thất bại → release reserved, không trừ quantity
Table InventoryReservations {
  id varchar [primary key]
  user_id integer [ref: > Users.id]
  store_id integer [ref: > Store.id] // giữ hàng tại store nào
  items jsonb        // [{ variant_id, quantity }]
  status varchar     // "pending", "confirmed", "expired", "cancelled"
  payment_code varchar
  payos_order_code varchar
  expires_at timestamp
  created_at timestamp
}

Table Suppliers {
  id serial [primary key]
  name varchar
  address varchar
  phone varchar
}

Table ImportInvoices {
  id serial [primary key]
  supplier_id integer [ref: > Suppliers.id]
  store_id integer [ref: > Store.id]
  created_by integer [ref: > Users.id]
  total_items integer
  note text
  created_at timestamp
}

Table ImportInvoiceDetails {
  id serial [primary key]
  invoice_id integer [ref: > ImportInvoices.id]
  variant_id integer [ref: > ProductVariant.id]
  quantity integer
  stock_before integer
  price_import numeric
}

// ─── CART ─────────────────────────────────────────────────────

Table CartItems {
  id serial [primary key]
  user_id integer [ref: > Users.id] // nullable — link với user khi đã đăng nhập
  session_id varchar                 // dùng cho guest cart (khách vãng lai)
  variant_id integer [ref: > ProductVariant.id]
  quantity integer
  created_at timestamp
  updated_at timestamp
}

// ─── PROMOTION & VOUCHER ──────────────────────────────────────

// Giảm giá tự động theo sản phẩm — hiển thị giá gạch ngang trên listing
// Khác Vouchers: không cần user nhập mã, áp dụng trực tiếp theo thời gian
// variant_id nullable — null = áp dụng cho toàn bộ variant của product đó
Table Promotions {
  id serial [primary key]
  product_id varchar [ref: > Product.id]
  variant_id integer [ref: > ProductVariant.id]
  name varchar
  description varchar
  discount_type varchar    // "percent" hoặc "fixed"
  discount_value numeric   // % giảm giá hoặc số tiền giảm cố định
  start_date timestamp
  end_date timestamp
  is_active bool [default: true]
  is_deleted bool [default: false]
}

// Mã giảm giá user tự nhập ở checkout — khác Promotions ở chỗ có kiểm soát
// số lần dùng, giá trị đơn tối thiểu, và giới hạn theo từng user
Table Vouchers {
  id serial [primary key]
  code varchar [unique]
  name varchar
  start_date timestamp
  end_date timestamp
  discount_type varchar    // "percent" hoặc "fixed"
  discount_value numeric
  discount_target varchar  // "order", "shipping"
  min_order_value numeric  // đơn tối thiểu để áp dụng
  max_discount_amount numeric // trần giảm — dùng khi discount_type = "percent"
  max_usage_total integer  // tổng số lần dùng toàn hệ thống
  max_usage_per_user integer
  used_count integer [default: 0]
  is_deleted bool [default: false]
}

Table VoucherUsages {
  id serial [primary key]
  voucher_id integer [ref: > Vouchers.id]
  user_id integer [ref: > Users.id]
  order_id integer [ref: > Orders.id]
  used_at timestamp
}

// ─── ORDER ────────────────────────────────────────────────────

// Lookup table cho trạng thái đơn hàng
// Ưu điểm: thêm/sửa status mới không cần đụng code, chỉ insert DB
// Trạng thái đơn hàng (Order State: Chờ xử lý, Đang xử lý, Hoàn thành, Đã hủy)
Table OrderStatus {
  id serial [primary key]
  code varchar    // "pending", "processing", "completed", "cancelled"
  label varchar   // "Chờ xác nhận", "Đang xử lý", "Hoàn thành", "Đã hủy"
  sort_order integer
}

// Trạng thái thanh toán (Chưa thanh toán, Đã thanh toán, Đã hoàn tiền...)
Table PaymentStatus {
  id serial [primary key]
  code varchar    // "unpaid", "paid", "refunded", "partially_refunded"
  label varchar   // "Chưa thanh toán", "Đã thanh toán", "Đã hoàn tiền", "Hoàn tiền một phần"
  sort_order integer
}

// Trạng thái giao hàng (Chưa giao, Đang giao, Đã giao, Giao thất bại, Chờ trả hàng...)
Table ShippingStatus {
  id serial [primary key]
  code varchar    // "unshipped", "shipping", "shipped", "failed", "returned"
  label varchar   // "Chưa giao", "Đang giao", "Đã giao", "Giao thất bại", "Chờ chuyển hoàn"
  sort_order integer
}

Table Orders {
  id serial [primary key]
  order_code varchar [unique] // Mã đơn hàng hiển thị trực quan (VD: DH20260526-0001)
  user_id integer [ref: > Users.id]
  store_id integer [ref: > Store.id]  // store fulfill đơn hàng này
  voucher_id integer [ref: > Vouchers.id]
  order_status_id integer [ref: > OrderStatus.id]
  payment_status_id integer [ref: > PaymentStatus.id]
  shipping_status_id integer [ref: > ShippingStatus.id]
  total_amount numeric
  voucher_discount numeric
  shipping_price numeric  // phí ship tính cho cả đơn, không phải từng item
  payment_method varchar
  note text
  receiver_name varchar
  receiver_address text
  receiver_phone varchar
  sender_name varchar
  sender_address text
  sender_phone varchar
  created_at timestamp
  updated_at timestamp
}

Table OrderDetails {
  id serial [primary key]
  order_id integer [ref: > Orders.id]
  variant_id integer [ref: > ProductVariant.id]
  quantity integer
  unit_price numeric  // giá tại thời điểm đặt hàng — snapshot, không ref sang variant
  total_cost numeric
}

// Audit trail cho trạng thái đơn hàng (theo dõi cả 3 loại trạng thái)
Table OrderStatusHistory {
  id serial [primary key]
  order_id integer [ref: > Orders.id]
  order_status_id integer [ref: > OrderStatus.id]
  payment_status_id integer [ref: > PaymentStatus.id]
  shipping_status_id integer [ref: > ShippingStatus.id]
  changed_by integer [ref: > Users.id]
  note text
  changed_at timestamp
}

// ─── REVIEWS ──────────────────────────────────────────────────

// images giữ dạng jsonb — ảnh review do user upload, mục đích khác ProductImage
// Không cần sort hay moderation từng ảnh → không cần tách bảng riêng
Table Reviews {
  id serial [primary key]
  user_id integer [ref: > Users.id]
  product_id varchar [ref: > Product.id]
  order_id integer [ref: > Orders.id] // đảm bảo chỉ review sau khi mua
  rating integer  // 1–5
  comment text
  images jsonb    // ["url1", "url2"] — ảnh thực tế do user chụp
  status varchar  // "pending", "approved", "rejected"
  created_at timestamp
  updated_at timestamp
}