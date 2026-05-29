import { useEffect, useState, useTransition } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import { useCatalog } from '../hooks/useCatalog'
import { productAPI } from '../services/productAPI'
import type { Product } from '../types'

const BrowsePage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { categories, brands, loading: catalogLoading } = useCatalog()
  const [, startTransition] = useTransition()

  // Extract query params
  const categoryIdParam = searchParams.get('category')
  const brandIdParam = searchParams.get('brand')
  const searchQuery = searchParams.get('q') || ''
  const activeSort = searchParams.get('sort') || 'newest'
  const activePage = parseInt(searchParams.get('page') || '1')

  const [products, setProducts] = useState<Product[]>([])
  const [loadingProducts, setLoadingProducts] = useState(true)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 12,
    total: 0,
    total_pages: 1,
  })

  // Fetch products when params change
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoadingProducts(true)
        const params: any = {
          page: activePage,
          limit: 12,
          sort: activeSort,
        }

        if (categoryIdParam) {
          params.category_id = parseInt(categoryIdParam)
        }
        if (brandIdParam) {
          params.brand_id = parseInt(brandIdParam)
        }
        if (searchQuery) {
          params.q = searchQuery
        }

        const res = await productAPI.getProducts(params)
        setProducts(res.data)
        if (res.pagination) {
          setPagination(res.pagination)
        } else {
          setPagination({
            page: activePage,
            limit: 12,
            total: res.data.length,
            total_pages: 1,
          })
        }
      } catch (err) {
        console.error('Failed to load products', err)
      } finally {
        setLoadingProducts(false)
      }
    }

    void fetchProducts()
  }, [categoryIdParam, brandIdParam, searchQuery, activeSort, activePage])

  const updateParam = (key: string, value: string | null) => {
    startTransition(() => {
      setSearchParams(prev => {
        const next = new URLSearchParams(prev)
        if (value === null) {
          next.delete(key)
        } else {
          next.set(key, value)
        }
        // Reset to page 1 when filter changes
        if (key !== 'page') {
          next.delete('page')
        }
        return next
      })
    })
  }

  const clearAllFilters = () => {
    startTransition(() => {
      setSearchParams(new URLSearchParams())
    })
  }

  const selectedCategory = categories.find(c => String(c.id) === categoryIdParam)
  const selectedBrand = brands.find(b => String(b.id) === brandIdParam)

  return (
    <div className="bg-neutral-50 min-h-screen py-8 font-sans">
      <div className="mx-auto max-w-7xl px-4">
        {/* Breadcrumbs */}
        <nav className="mb-6 flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">
            Trang chủ
          </Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Tất cả sản phẩm</span>
        </nav>

        {/* Header Title with query results */}
        <div className="mb-8 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-black text-neutral-900 tracking-tight uppercase">
              {selectedCategory ? selectedCategory.name : 'Khám phá sản phẩm'}
            </h1>
            <p className="text-xs text-neutral-450 mt-1">
              Hiển thị {pagination.total} sản phẩm phù hợp
            </p>
          </div>

          {/* Sort selection dropdown */}
          <div className="flex items-center gap-3 self-end md:self-auto">
            <span className="text-xs font-bold text-neutral-500 uppercase tracking-wider">Sắp xếp:</span>
            <select
              value={activeSort}
              onChange={(e) => updateParam('sort', e.target.value)}
              className="bg-white border border-neutral-250 text-xs font-bold px-3 py-2 rounded focus:outline-none focus:border-black cursor-pointer shadow-sm"
            >
              <option value="newest">Mới nhất</option>
              <option value="popular">Bán chạy nhất</option>
              <option value="price_asc">Giá: Thấp đến Cao</option>
              <option value="price_desc">Giá: Cao đến Thấp</option>
            </select>
          </div>
        </div>

        {/* Layout Grid: Sidebar + Product Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          
          {/* Left Column: Sidebar Filters */}
          <div className="space-y-6">
            
            {/* Active Filters list */}
            {(categoryIdParam || brandIdParam || searchQuery) && (
              <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-3.5 shadow-sm">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-black text-neutral-800 uppercase tracking-wider">Đang lọc</span>
                  <button
                    onClick={clearAllFilters}
                    className="text-[10px] font-bold text-red-500 hover:text-red-650 transition-colors uppercase"
                  >
                    Xóa tất cả
                  </button>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {searchQuery && (
                    <span className="inline-flex items-center gap-1 bg-neutral-100 text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded border border-neutral-200">
                      Từ khóa: {searchQuery}
                      <button onClick={() => updateParam('q', null)} className="hover:text-black text-xs">✕</button>
                    </span>
                  )}
                  {selectedCategory && (
                    <span className="inline-flex items-center gap-1 bg-neutral-100 text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded border border-neutral-200">
                      {selectedCategory.name}
                      <button onClick={() => updateParam('category', null)} className="hover:text-black text-xs">✕</button>
                    </span>
                  )}
                  {selectedBrand && (
                    <span className="inline-flex items-center gap-1 bg-neutral-100 text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded border border-neutral-200">
                      {selectedBrand.name}
                      <button onClick={() => updateParam('brand', null)} className="hover:text-black text-xs">✕</button>
                    </span>
                  )}
                </div>
              </div>
            )}

            {/* Categories filter box */}
            <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-4 shadow-sm">
              <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">Danh Mục</h3>
              {catalogLoading ? (
                <div className="space-y-2">
                  {[...Array(5)].map((_, i) => (
                    <div key={i} className="h-4 w-3/4 bg-neutral-100 animate-pulse rounded"></div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <button
                    onClick={() => updateParam('category', null)}
                    className={`text-left text-xs py-1.5 font-bold transition-colors ${!categoryIdParam ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                  >
                    Tất cả danh mục
                  </button>
                  {categories.map((c) => (
                    <button
                      key={c.id}
                      onClick={() => updateParam('category', String(c.id))}
                      className={`text-left text-xs py-1.5 font-bold transition-colors ${categoryIdParam === String(c.id) ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                    >
                      {c.name}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Brands filter box */}
            <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-4 shadow-sm">
              <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">Thương Hiệu</h3>
              {catalogLoading ? (
                <div className="space-y-2">
                  {[...Array(5)].map((_, i) => (
                    <div key={i} className="h-4 w-3/4 bg-neutral-100 animate-pulse rounded"></div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <button
                    onClick={() => updateParam('brand', null)}
                    className={`text-left text-xs py-1.5 font-bold transition-colors ${!brandIdParam ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                  >
                    Tất cả thương hiệu
                  </button>
                  {brands.map((b) => (
                    <button
                      key={b.id}
                      onClick={() => updateParam('brand', String(b.id))}
                      className={`text-left text-xs py-1.5 font-bold transition-colors ${brandIdParam === String(b.id) ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                    >
                      {b.name}
                    </button>
                  ))}
                </div>
              )}
            </div>

          </div>

          {/* Right Column: Products Grid */}
          <div className="lg:col-span-3 space-y-8">
            
            {/* Loading state or products display */}
            {loadingProducts ? (
              <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="aspect-[3/4] bg-neutral-200 animate-pulse rounded-lg"></div>
                ))}
              </div>
            ) : (
              <>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
                  {products.map((product) => (
                    <ProductCard key={product.id} product={product} />
                  ))}
                </div>

                {products.length === 0 && (
                  <div className="border border-neutral-250 border-dashed rounded-lg p-16 text-center bg-white">
                    <span className="text-4xl">🔍</span>
                    <h3 className="text-sm font-bold text-neutral-800 uppercase tracking-wider mt-4">Không tìm thấy sản phẩm</h3>
                    <p className="text-xs text-neutral-450 mt-1 max-w-sm mx-auto leading-relaxed">
                      Rất tiếc, chúng tôi không tìm thấy kết quả phù hợp với các bộ lọc hiện tại của bạn. Thử thay đổi hoặc xóa các tiêu chí xem sao.
                    </p>
                    <button
                      onClick={clearAllFilters}
                      className="mt-6 bg-black text-white text-xs font-black uppercase px-6 py-2.5 rounded hover:bg-neutral-850 transition-colors"
                    >
                      Xóa bộ lọc
                    </button>
                  </div>
                )}

                {/* Pagination */}
                {pagination.total_pages > 1 && (
                  <div className="flex items-center justify-center gap-2 pt-6">
                    <button
                      disabled={activePage <= 1}
                      onClick={() => updateParam('page', String(activePage - 1))}
                      className="border border-neutral-250 px-3.5 py-2 rounded text-xs font-bold text-neutral-600 hover:border-black disabled:opacity-40 disabled:hover:border-neutral-250 transition-all bg-white"
                    >
                      Trước
                    </button>
                    {[...Array(pagination.total_pages)].map((_, idx) => {
                      const pageNum = idx + 1
                      return (
                        <button
                          key={pageNum}
                          onClick={() => updateParam('page', String(pageNum))}
                          className={`w-9 h-9 rounded text-xs font-bold transition-all border ${
                            activePage === pageNum
                              ? 'bg-black text-white border-black'
                              : 'bg-white text-neutral-650 border-neutral-250 hover:border-black'
                          }`}
                        >
                          {pageNum}
                        </button>
                      )
                    })}
                    <button
                      disabled={activePage >= pagination.total_pages}
                      onClick={() => updateParam('page', String(activePage + 1))}
                      className="border border-neutral-250 px-3.5 py-2 rounded text-xs font-bold text-neutral-600 hover:border-black disabled:opacity-40 disabled:hover:border-neutral-250 transition-all bg-white"
                    >
                      Sau
                    </button>
                  </div>
                )}
              </>
            )}

          </div>
        </div>
      </div>
    </div>
  )
}

export default BrowsePage
