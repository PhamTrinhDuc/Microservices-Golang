import { useEffect, useState } from 'react'
import { productAPI } from '../services/productAPI'
import type { Product } from '../types'

export const useFeaturedProducts = (limit = 24) => {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoading(true)
        const res = await productAPI.getProducts({ limit })
        setProducts(res.data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch featured products')
      } finally {
        setLoading(false)
      }
    }

    void fetchProducts()
  }, [limit])

  return {
    products,
    loading,
    error,
  }
}
