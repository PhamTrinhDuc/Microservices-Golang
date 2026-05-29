import { useEffect, useState, useTransition } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import SearchBar from '../components/SearchBar'
import ProductCard from '../components/ProductCard'
import { productAPI } from '../services/productAPI'
import type { Product } from '../types'

const SearchPage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [, startTransition] = useTransition()
  const searchQuery = searchParams.get('q') || ''
  const activeSort = searchParams.get('sort') || 'newest'
  const activePage = parseInt(searchParams.get('page') || '1')

  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 12,
    total: 0,
    total_pages: 1,
  })

  useEffect(() => {
    const fetchSearchResults = async () => {
      if (!searchQuery.trim()) {
        setProducts([])
        return
      }

      try {
        setLoading(true)
        const params: any = {
          q: searchQuery,
          page: activePage,
          limit: 12,
          sort: activeSort,
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
        console.error('Failed to load search results', err)
      } finally {
        setLoading(false)
      }
    }

    void fetchSearchResults()
  }, [searchQuery, activeSort, activePage])

  const handleSearchSubmit = (newQuery: string) => {
    startTransition(() => {
      setSearchParams({ q: newQuery })
    })
  }

  const updateParam = (key: string, value: string | null) => {
    startTransition(() => {
      setSearchParams(prev => {
        const next = new URLSearchParams(prev)
        if (value === null) {
          next.delete(key)
        } else {
          next.set(key, value)
        }
        if (key !== 'page') {
          next.delete('page')
        }
        return next
      })
    })
  }

  return (
    <div className="bg-neutral-50 min-h-screen py-8 font-sans">
      <div className="mx-auto max-w-7xl px-4 space-y-8">
        
        {/* Breadcrumbs */}
        <nav className="flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">
            Trang chủ
          </Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Tìm kiếm</span>
        </nav>

        {/* Large Premium Search Bar container */}
        <div className="max-w-2xl mx-auto text-center space-y-4">
          <h1 className="text-2xl font-black text-neutral-900 tracking-tight uppercase">Tìm kiếm sản phẩm</h1>
          <SearchBar onSearch={handleSearchSubmit} initialValue={searchQuery} />
        </div>

        {/* Results Info and Sort Option */}
        {searchQuery && (
          <div className="border-t border-neutral-200 pt-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h2 className="text-sm font-bold text-neutral-800 uppercase tracking-wide">
                  Kết quả tìm kiếm cho: <span className="text-neutral-500 font-semibold">"{searchQuery}"</span>
                </h2>
                <p className="text-[11px] text-neutral-400 mt-0.5">Tìm thấy {pagination.total} sản phẩm phù hợp</p>
              </div>

              <div className="flex items-center gap-3">
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

            {/* Results display */}
            <div className="mt-8">
              {loading ? (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                  {[...Array(4)].map((_, i) => (
                    <div key={i} className="aspect-[3/4] bg-neutral-200 animate-pulse rounded-lg"></div>
                  ))}
                </div>
              ) : (
                <>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                    {products.map((product) => (
                      <ProductCard key={product.id} product={product} />
                    ))}
                  </div>

                  {products.length === 0 && (
                    <div className="border border-neutral-250 border-dashed rounded-lg p-16 text-center bg-white max-w-lg mx-auto">
                      <span className="text-4xl">🔍</span>
                      <h3 className="text-sm font-bold text-neutral-800 uppercase tracking-wider mt-4">Không tìm thấy sản phẩm</h3>
                      <p className="text-xs text-neutral-450 mt-1 leading-relaxed">
                        Thử điều chỉnh từ khóa tìm kiếm của bạn bằng cách dùng các từ phổ thông hoặc kiểm tra lại chính tả xem sao.
                      </p>
                    </div>
                  )}

                  {/* Pagination */}
                  {pagination.total_pages > 1 && (
                    <div className="flex items-center justify-center gap-2 pt-10">
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
        )}

        {/* Empty state when query is blank */}
        {!searchQuery && (
          <div className="text-center py-20 text-neutral-450 space-y-1">
            <span className="text-4xl block mb-4">🛒</span>
            <p className="text-xs font-bold text-neutral-600 uppercase tracking-wide">Nhập từ khóa để bắt đầu tìm kiếm</p>
            <p className="text-[10px] text-neutral-400">Tìm kiếm các sản phẩm hot từ Apple, Samsung, Sony...</p>
          </div>
        )}

      </div>
    </div>
  )
}

export default SearchPage
