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
  price: number
  price_base?: number | null
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
}

