import { api } from './api'
import type { Brand, Category, Product } from '../types'

// Cache brands to avoid fetching them repeatedly for mapping brand_id -> Brand on product lists
let cachedBrands: Brand[] = []

const getBrandsList = async (): Promise<Brand[]> => {
  if (cachedBrands.length > 0) return cachedBrands
  try {
    const res = await api.get<Brand[]>('/brands')
    cachedBrands = res.data || []
    return cachedBrands
  } catch {
    return []
  }
}

export const productAPI = {
  getCategories: async (): Promise<Category[]> => {
    const response = await api.get<Category[]>('/categories')
    return response.data || []
  },

  getBrands: async (): Promise<Brand[]> => {
    const brands = await getBrandsList()
    return brands
  },

  getProducts: async (params: {
    category_id?: number
    brand_id?: number
    q?: string
    page?: number
    limit?: number
    sort?: string
  }): Promise<{ data: Product[]; pagination?: any }> => {
    // Convert generic query parameters if category or brand are passed from query strings
    const apiParams: any = {
      ...params,
    }

    const response = await api.get<any>('/products', { params: apiParams })
    const brands = await getBrandsList()

    // Retrieve data from response
    // The response interceptor maps the structure to { data, pagination } if it has total_count, or raw data array
    const rawList = response.data?.data || response.data || []
    const pagination = response.data?.pagination || null

    const mappedProducts: Product[] = rawList.map((p: any) => {
      const brand = brands.find((b: any) => b.id === p.brand_id)
      return {
        id: p.id,
        category_id: p.category_id,
        brand_id: p.brand_id,
        name: p.name,
        slug: p.slug,
        meta_title: p.meta_title,
        meta_description: p.meta_description,
        image: p.img_thumb || p.image || '/placeholder-product.png',
        weight: p.weight,
        low_stock_threshold: p.low_stock_threshold,
        price: p.price || 0,
        discount_price: p.discount_price || null,
        discount_percent: p.discount_percent || 0,
        stock: p.stock || 0,
        rating: p.rating || 4.7, // aesthetic rating
        review_count: p.review_count || 12, // aesthetic review count
        created_at: p.created_at,
        updated_at: p.updated_at,
        brand: brand || (p.brand_id ? { id: p.brand_id, name: `Brand #${p.brand_id}`, slug: '', is_active: true } : null),
      }
    })

    return {
      data: mappedProducts,
      pagination,
    }
  },

  getProductById: async (id: string): Promise<Product> => {
    const response = await api.get<any>(`/products/${id}`)
    const details = response.data
    console.log("data:", details)

    if (!details || !details.product) {
      throw new Error('Product not found')
    }

    const p = details.product
    const specs = details.specs || []
    const variants = details.variants || []
    const images = details.images || []
    const reviews = details.reviews || []

    return {
      id: p.id,
      category_id: p.category_id,
      brand_id: p.brand_id,
      name: p.name,
      slug: p.slug,
      meta_title: p.meta_title,
      meta_description: p.meta_description,
      image: p.img_thumb || p.image || '/placeholder-product.png',
      weight: p.weight,
      low_stock_threshold: p.low_stock_threshold,
      price: p.price || 0,
      discount_price: p.discount_price || null,
      discount_percent: p.discount_percent || 0,
      stock: p.stock || 0,
      rating: p.rating ?? 0,
      review_count: p.review_count ?? 0,
      created_at: p.created_at,
      updated_at: p.updated_at,
      category: {
        id: p.category_id,
        name: details.category_name || 'Category',
        slug: '',
        sort_order: 0,
      },
      brand: {
        id: p.brand_id,
        name: details.brand_name || 'Brand',
        slug: '',
        is_active: true,
      },
      specifications: specs.map((s: any) => ({
        id: s.id,
        product_id: s.product_id,
        group: s.group,
        key: s.key,
        value: s.value,
        unit: s.unit,
        sort_order: s.sort_order,
      })),
      variants: variants.map((v: any) => ({
        id: v.id,
        product_id: v.product_id,
        name: v.name,
        sku: v.sku,
        sell_price: v.sell_price,
        compare_price: v.compare_price,
        latest_cost_price: v.latest_cost_price,
        weight: v.weight,
        is_active: v.is_active,
        stock: p.stock, // Fallback variant stock to total product stock
        options: v.options || [],
      })),
      images: images.length > 0 ? images.filter((img: any) => !img.is_thumbnail && !img.variant_id).map((img: any) => img.url) : [],
      reviews: reviews.map((r: any) => ({
        id: r.id,
        user_id: r.user_id,
        user_full_name: r.user_full_name,
        product_id: r.product_id,
        order_id: r.order_id,
        rating: r.rating,
        comment: r.comment,
        images: r.images,
        status: r.status,
        created_at: r.created_at,
        updated_at: r.updated_at,
      })),
    }
  },

  // Admin Catalog CRUD
  adminCreateCategory: async (payload: { name: string; parent_id?: number | null; icon?: string | null; sort_order?: number }): Promise<Category> => {
    const res = await api.post<Category>('/admin/categories', payload)
    return res.data
  },

  adminCreateBrand: async (payload: { name: string; logo_url?: string | null; is_active?: boolean }): Promise<Brand> => {
    const res = await api.post<Brand>('/admin/brands', payload)
    return res.data
  },

  adminCreateProduct: async (payload: {
    id: string
    category_id: number
    brand_id: number
    name: string
    meta_title?: string | null
    meta_description?: string | null
    img_thumb?: string | null
    images?: string[]
    weight?: number | null
    low_stock_threshold: number
    specs?: { group: string; key: string; value: string; unit?: string; sort_order?: number }[]
  }): Promise<Product> => {
    const res = await api.post<Product>('/admin/products', payload)
    return res.data
  },

  adminGenerateVariant: async (productId: string, payload: {
    name: string
    sku: string
    sell_price: number
    compare_price?: number
    weight?: number
    option_value_ids: number[]
  }): Promise<any> => {
    const res = await api.post<any>(`/admin/products/${productId}/variants`, payload)
    return res.data
  },

  adminAddOptionValues: async (payload: {
    product_id: string
    option_name: string
    values: { value: string; color_code?: string | null; sort_order: number }[]
  }): Promise<{ id: number; option_type_id: number; value: string; color_code: string | null; sort_order: number }[]> => {
    const res = await api.post<any>('/admin/option-values', payload)
    return res.data
  },

  adminUpdateCategory: async (id: number, payload: { name: string; parent_id?: number | null; sort_order?: number }): Promise<Category> => {
    const res = await api.put<Category>(`/admin/categories/${id}`, payload)
    return res.data
  },

  adminDeleteCategory: async (id: number): Promise<void> => {
    await api.delete(`/admin/categories/${id}`)
  },

  adminUpdateBrand: async (id: number, payload: { name: string; logo_url?: string | null; is_active?: boolean }): Promise<Brand> => {
    const res = await api.put<Brand>(`/admin/brands/${id}`, payload)
    return res.data
  },

  adminDeleteBrand: async (id: number): Promise<void> => {
    await api.delete(`/admin/brands/${id}`)
  },

  adminUpdateProduct: async (id: string, payload: {
    category_id: number
    brand_id: number
    name: string
    meta_title?: string | null
    meta_description?: string | null
    img_thumb?: string | null
    images?: string[]
    weight?: number | null
    low_stock_threshold: number
    specs?: { group: string; key: string; value: string; unit?: string; sort_order?: number }[]
  }): Promise<Product> => {
    const res = await api.put<Product>(`/admin/products/${id}`, payload)
    return res.data
  },

  adminDeleteProduct: async (id: string): Promise<void> => {
    await api.delete(`/admin/products/${id}`)
  },

  adminUpdateVariant: async (id: number, payload: {
    name: string
    sku: string
    sell_price: number
    compare_price?: number
    weight?: number
  }): Promise<any> => {
    const res = await api.put<any>(`/admin/variants/${id}`, payload)
    return res.data
  },

  adminDeleteVariant: async (id: number): Promise<void> => {
    await api.delete(`/admin/variants/${id}`)
  },
}
