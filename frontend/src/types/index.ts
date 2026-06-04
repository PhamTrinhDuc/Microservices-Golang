export interface User {
  id: number
  email: string
  full_name: string
  phone?: string
  gender?: string
  dob?: string
  role: string
  avatar?: string
  is_lock: boolean
  is_verified: boolean
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  full_name: string
  email: string
  password: string
}

export interface GoogleLoginRequest {
  credential: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface Brand {
  id: number
  name: string
  slug: string
  logo?: string
  logo_url?: string
  is_active: boolean
}

export interface Category {
  id: number
  name: string
  parent_id?: number | null
  icon?: string
  slug: string
  sort_order: number
}

export interface ProductSpec {
  id: number
  product_id: string
  group: string
  key: string
  value: string
  unit?: string
  sort_order: number
}

export interface ProductVariant {
  id: number
  product_id: string
  name: string
  sku: string
  sell_price: number
  compare_price?: number | null
  latest_cost_price: number
  discount_price?: number | null
  weight?: number | null
  is_active: boolean
  stock: number
  options?: any[]
}

export interface ProductImage {
  id: number
  product_id: string
  variant_id?: number | null
  url: string
  alt_text?: string | null
  sort_order: number
  is_thumbnail: boolean
}

export interface Product {
  id: string
  category_id: number
  brand_id: number
  name: string
  slug: string
  description?: string | null
  meta_title?: string | null
  meta_description?: string | null
  image: string
  weight?: number | null
  low_stock_threshold: number
  price: number
  discount_price: number | null
  discount_percent: number
  stock: number
  rating: number
  review_count: number
  created_at: string
  updated_at: string
  
  category?: Category | null
  brand?: Brand | null
  specifications?: ProductSpec[]
  variants?: ProductVariant[]
  images?: string[]
  reviews?: Review[]
  specs_jsonb?: any
}

export interface Review {
  id: number
  user_id: number
  user_full_name: string
  product_id: string
  order_id: number
  rating: number
  comment?: string | null
  images?: any
  status: string
  created_at: string
  updated_at: string
}

export interface CartItem {
  id: number
  variant_id: number
  variant_name: string
  sku: string
  sell_price: number
  compare_price: number
  product_id: string
  product_name: string
  image_url?: string
  quantity: number
}

export interface AddToCartRequest {
  variant_id: number
  quantity: number
  session_id?: string
}

// ============================================================
// ADMIN & ORDER TYPES
// ============================================================

export interface Address {
  id: number
  user_id: number
  full_name: string
  phone: string
  district: string
  province: string
  ward: string
  detail_address: string
  is_default: boolean
  is_deleted?: boolean
}

export interface CreateAddressRequest {
  full_name: string
  phone: string
  district: string
  province: string
  ward: string
  detail_address: string
  is_default?: boolean
}

export interface CheckoutOrderRequest {
  store_id: number
  address_id?: number | null
  manual_receiver?: {
    name: string
    phone: string
    address: string
  } | null
  shipping_provider: 'ghn' | 'ghtk'
  payment_method: 'payos' | 'cod' | 'bank_transfer'
  voucher_code?: string | null
  note?: string | null
}

export interface Order {
  id: number
  order_code: string
  user_id: number
  store_id: number
  voucher_id?: number | null
  order_status_id: number
  payment_status_id: number
  shipping_status_id: number
  total_amount: number
  voucher_discount: number
  shipping_price: number
  payment_method: string
  payment_code?: string | null
  payos_order_code?: string | null
  note?: string | null
  receiver_name: string
  receiver_address: string
  receiver_phone: string
  shipping_provider?: string | null
  shipping_code?: string | null
  created_at: string
  updated_at: string
}

export interface OrderItem {
  id: number
  order_id: number
  variant_id: number
  variant_name: string
  quantity: number
  unit_price: number
  total_cost: number
}

export interface OrderStatusHistory {
  id: number
  order_id: number
  status_type: string // "order", "payment", "shipping"
  from_status?: string
  to_status: string
  changed_by?: number
  changed_by_name?: string
  note?: string
  changed_at: string
}

export interface OrderResponse {
  order: Order
  items: OrderItem[]
  order_status_label: string
  payment_status_label: string
  shipping_status_label: string
  checkout_url?: string | null
  history?: OrderStatusHistory[]
}

export interface UpdateOrderStatusRequest {
  order_status_code?: string
  payment_status_code?: string
  shipping_status_code?: string
  shipping_provider?: string
  shipping_code?: string
  note?: string
}

export interface Voucher {
  id: number
  code: string
  name: string
  start_date: string
  end_date: string
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  discount_target: 'order' | 'shipping'
  min_order_value: number
  max_discount_amount: number | null
  max_usage_total: number | null
  max_usage_per_user: number
  used_count: number
  is_deleted: boolean
}

export interface ApplyVoucherRequest {
  code: string
  order_amount: number
}

export interface ApplyVoucherResponse {
  valid: boolean
  voucher_id: number
  discount_amount: number
}

export interface CreateVoucherRequest {
  code: string
  name: string
  start_date: string
  end_date: string
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  discount_target?: 'order' | 'shipping'
  min_order_value: number
  max_discount_amount?: number | null
  max_usage_total?: number | null
  max_usage_per_user?: number
}

export interface UpdateVoucherRequest extends Partial<CreateVoucherRequest> {}

export interface Promotion {
  id: number
  product_id: string
  variant_id?: number | null
  name: string
  description?: string | null
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  start_date: string
  end_date: string
  is_active: boolean
  is_deleted: boolean
  
  product_name?: string
  variant_name?: string
}

export interface CreatePromotionRequest {
  product_id: string
  variant_id?: number | null
  name: string
  description?: string | null
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  start_date: string
  end_date: string
}

export interface UpdatePromotionRequest {
  name: string
  description?: string | null
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  start_date: string
  end_date: string
  is_active: boolean
}

export interface Store {
  id: number
  name: string
  hotline?: string | null
  district: string
  province: string
  ward: string
  road?: string | null
  email?: string | null
  lat?: number | null
  lng?: number | null
  is_active: boolean
  created_at?: string
  updated_at?: string
}

export interface Supplier {
  id: number
  name: string
  address?: string | null
  phone?: string | null
  email?: string | null
  contact_name?: string | null
  contact_phone?: string | null
  is_deleted: boolean
  total_imports?: number
  last_imported_at?: string | null
  total_import_value?: number
}

export interface ProductInventory {
  variant_id: number
  store_id: number
  quantity: number
  reserved: number
  last_updated: string
  
  // joined properties
  product_name?: string
  variant_name?: string
  sku?: string
  price?: number
}

export interface ImportInvoiceResponse {
  id: number
  supplier_id: number
  supplier_name: string
  store_id: number
  store_name: string
  created_by: number
  creator_name: string
  total_items: number
  note?: string | null
  status: string
  created_at: string
}

export interface ImportInvoiceDetail {
  id: number
  invoice_id: number
  variant_id: number
  variant_name: string
  sku: string
  quantity: number
  stock_before: number
  price_import: number
}

export interface ImportInvoiceDetailsResponse {
  invoice: ImportInvoiceResponse
  details: ImportInvoiceDetail[]
}

export interface LowStockAlertResponse {
  variant_id: number
  product_id: string
  product_name: string
  variant_name: string
  sku: string
  quantity: number
  low_stock_threshold: number
}

export interface InventoryLogResponse {
  id: number
  variant_id: number
  variant_name: string
  sku: string
  store_id: number
  store_name: string
  change_qty: number
  qty_after: number
  reason: string
  ref_id?: string | null
  created_by?: number | null
  creator_name?: string | null
  created_at: string
}

export interface Banner {
  id: number
  title: string
  subtitle?: string | null
  description?: string | null
  image_url: string
  tag?: string | null
  link_url?: string | null
  sort_order: number
  is_active: boolean
  created_at?: string
  updated_at?: string
}export interface WishlistItemResponse {
  id: number
  variant_id: number
  variant_name: string
  sku: string
  sell_price: number
  compare_price?: number | null
  product_id: string
  product_name: string
  image_url?: string | null
  stock: number
  discount_price?: number | null
  rating: number
}

export interface InventorySummary {
  total_sku: number
  total_quantity: number
  low_stock_count: number
  out_of_stock_count: number
}

export interface TopCategoryReport {
  category_name: string
  sold_qty: number
  revenue: number
}

export interface SalesOverTime {
  date: string
  revenue: number
  orders_count: number
}

export interface TopProductReport {
  product_name: string
  sold_qty: number
  revenue: number
}

export interface StoreReport {
  store_id: number
  store_name: string
  orders_count: number
  revenue: number
}

export interface StatusDistribution {
  status_label: string
  count: number
}

export interface AnalyticsSummary {
  total_sales: number
  total_orders: number
  average_order_value: number
  items_sold: number
  prev_total_sales: number
  prev_total_orders: number
  prev_average_order_value: number
  sales_growth: number
  orders_growth: number
  aov_growth: number
  inventory: InventorySummary
  sales_over_time: SalesOverTime[]
  top_products: TopProductReport[]
  top_categories: TopCategoryReport[]
  store_sales: StoreReport[]
  status_distribution: StatusDistribution[]
}

