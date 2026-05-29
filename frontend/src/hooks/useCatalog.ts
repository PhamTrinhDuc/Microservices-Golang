import { useEffect, useState } from 'react'
import { productAPI } from '../services/productAPI'
import type { Brand, Category } from '../types'

export const useCatalog = () => {
  const [categories, setCategories] = useState<Category[]>([])
  const [brands, setBrands] = useState<Brand[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchCatalog = async () => {
      try {
        setLoading(true)
        const [catsData, brandsData] = await Promise.all([
          productAPI.getCategories(),
          productAPI.getBrands(),
        ])
        setCategories(catsData)
        setBrands(brandsData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch catalog')
      } finally {
        setLoading(false)
      }
    }

    void fetchCatalog()
  }, [])

  return {
    categories,
    brands,
    loading,
    error,
  }
}
