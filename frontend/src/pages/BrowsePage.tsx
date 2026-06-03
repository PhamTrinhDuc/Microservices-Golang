import { useEffect, useState, useTransition, useMemo } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import SearchSkeleton from '../components/SearchSkeleton'
import { useCatalog } from '../hooks/useCatalog'
import { productAPI } from '../services/productAPI'
import { useCart } from '../hooks/useCart'
import type { Product } from '../types'

const BrowsePage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { categories, brands, loading: catalogLoading } = useCatalog()
  const [, startTransition] = useTransition()

  const [categorySearch, setCategorySearch] = useState('')
  const [brandSearch, setBrandSearch] = useState('')

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

  // Layout & Client Side Filtering State
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [inStockOnly, setInStockOnly] = useState(false)
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 100000000])
  const [selectedSpecs, setSelectedSpecs] = useState<{ [specKey: string]: string[] }>({})
  
  const { addToCart } = useCart()
  const [loadingCartMap, setLoadingCartMap] = useState<{ [key: string]: boolean }>({})

  // Fetch products when params change
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoadingProducts(true)
        const params: any = {
          page: activePage,
          limit: 24, // Fetch more for richer client-side filters
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
            limit: 24,
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
    // Reset client filters on route query updates
    setInStockOnly(false)
    setSelectedSpecs({})
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
    setInStockOnly(false)
    setSelectedSpecs({})
  }

  const selectedCategory = categories.find(c => String(c.id) === categoryIdParam)
  const selectedBrand = brands.find(b => String(b.id) === brandIdParam)

  const filteredCategories = useMemo(() => {
    return categories.filter(c => c.name.toLowerCase().includes(categorySearch.toLowerCase()))
  }, [categories, categorySearch])

  const filteredBrands = useMemo(() => {
    return brands.filter(b => b.name.toLowerCase().includes(brandSearch.toLowerCase()))
  }, [brands, brandSearch])

  // Extract specs-based dynamic filter options from loaded products
  const filterSpecsList = useMemo(() => {
    if (!products.length) return []
    const specsMap: { [key: string]: Set<string> } = {}

    products.forEach((product) => {
      let specs: any = {}
      if (product.specs_jsonb) {
        try {
          specs = typeof product.specs_jsonb === 'string'
            ? JSON.parse(product.specs_jsonb)
            : product.specs_jsonb
        } catch (e) {
          // ignore
        }
      }

      Object.keys(specs).forEach((groupName) => {
        const groupData = specs[groupName]
        if (typeof groupData === 'object' && groupData !== null) {
          Object.keys(groupData).forEach((specKey) => {
            const specInfo = groupData[specKey]
            if (specInfo && typeof specInfo === 'object') {
              const rawVal = specInfo.raw_value || specInfo.value
              if (rawVal) {
                const formattedVal = String(rawVal).trim()
                if (formattedVal && formattedVal !== '-' && formattedVal !== 'Không') {
                  if (!specsMap[specKey]) {
                    specsMap[specKey] = new Set()
                  }
                  specsMap[specKey].add(formattedVal)
                }
              }
            }
          })
        }
      })
    })

    const importantKeys = [
      'RAM',
      'Dung lượng lưu trữ',
      'Bộ nhớ trong',
      'Dung lượng pin',
      'Kích thước màn hình',
      'Độ phân giải',
      'Chipset',
      'CPU',
      'Card đồ họa',
      'Loại tai nghe',
      'Kháng nước',
      'Màu sắc'
    ]

    return Object.keys(specsMap)
      .filter((key) => specsMap[key].size > 1)
      .map((key) => ({
        key,
        options: Array.from(specsMap[key]).slice(0, 8),
      }))
      .sort((a, b) => {
        const aIdx = importantKeys.indexOf(a.key)
        const bIdx = importantKeys.indexOf(b.key)
        if (aIdx !== -1 && bIdx !== -1) return aIdx - bIdx
        if (aIdx !== -1) return -1
        if (bIdx !== -1) return 1
        return a.key.localeCompare(b.key)
      })
  }, [products])

  // Get boundary prices from results
  const priceLimits = useMemo(() => {
    if (!products.length) return { min: 0, max: 100000000 }
    const prices = products.map((p) => p.discount_price || p.price || 0)
    return {
      min: Math.min(...prices),
      max: Math.max(...prices),
    }
  }, [products])

  // Reset range state when boundaries change
  useEffect(() => {
    setPriceRange([priceLimits.min, priceLimits.max])
  }, [priceLimits])

  // Dynamic Specs handler
  const handleToggleSpecOption = (specKey: string, optionVal: string) => {
    setSelectedSpecs((prev) => {
      const currentList = prev[specKey] || []
      const nextList = currentList.includes(optionVal)
        ? currentList.filter((v) => v !== optionVal)
        : [...currentList, optionVal]

      const next = { ...prev, [specKey]: nextList }
      if (nextList.length === 0) {
        delete next[specKey]
      }
      return next
    })
  }

  // Client side filtration
  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      if (inStockOnly && product.stock <= 0) return false

      const displayPrice = product.discount_price || product.price || 0
      if (displayPrice < priceRange[0] || displayPrice > priceRange[1]) return false

      for (const [specKey, values] of Object.entries(selectedSpecs)) {
        if (!values || values.length === 0) continue

        let specs: any = {}
        if (product.specs_jsonb) {
          try {
            specs = typeof product.specs_jsonb === 'string'
              ? JSON.parse(product.specs_jsonb)
              : product.specs_jsonb
          } catch (e) {
            // ignore
          }
        }

        let productValue = ''
        for (const groupName of Object.keys(specs)) {
          const groupData = specs[groupName]
          if (groupData && typeof groupData === 'object') {
            const specInfo = groupData[specKey]
            if (specInfo) {
              productValue = String(specInfo.raw_value || specInfo.value || '').trim()
              break
            }
          }
        }

        if (!productValue || !values.includes(productValue)) {
          return false
        }
      }

      return true
    })
  }, [products, inStockOnly, priceRange, selectedSpecs])

  const handleQuickAddList = async (e: React.MouseEvent, product: Product) => {
    e.preventDefault()
    e.stopPropagation()
    if (product.stock === 0) return

    try {
      setLoadingCartMap((prev) => ({ ...prev, [product.id]: true }))
      const detailedProduct = await productAPI.getProductById(product.id)
      const targetVariant = detailedProduct.variants?.[0]

      if (!targetVariant) {
        alert('Sản phẩm này chưa cấu hình phiên bản bán hàng.')
        return
      }

      await addToCart(targetVariant.id, 1).unwrap()
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setLoadingCartMap((prev) => ({ ...prev, [product.id]: false }))
    }
  }

  const getPageNumbers = useMemo(() => {
    const total = pagination.total_pages
    const current = activePage
    const pages: (number | string)[] = []

    let left = current - 2
    let right = current + 2

    if (left < 1) {
      right += (1 - left)
      left = 1
    }
    if (right > total) {
      left -= (right - total)
      right = total
    }

    left = Math.max(1, left)
    right = Math.min(total, right)

    for (let i = left; i <= right; i++) {
      pages.push(i)
    }

    if (left > 1) {
      if (left > 2) {
        pages.unshift('...')
      }
      pages.unshift(1)
    }

    if (right < total) {
      if (right < total - 1) {
        pages.push('...')
      }
      pages.push(total)
    }

    return Array.from(new Set(pages))
  }, [pagination.total_pages, activePage])

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
              Hiển thị {filteredProducts.length} trên tổng số {pagination.total} sản phẩm
            </p>
          </div>

          <div className="flex items-center gap-4">
            {/* Layout Toggles */}
            <div className="flex items-center border border-neutral-200 rounded overflow-hidden">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 transition-colors ${viewMode === 'grid' ? 'bg-neutral-900 text-white' : 'bg-white text-neutral-600 hover:bg-neutral-100'}`}
                title="Dạng lưới"
              >
                <svg className="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`p-2 transition-colors ${viewMode === 'list' ? 'bg-neutral-900 text-white' : 'bg-white text-neutral-600 hover:bg-neutral-100'}`}
                title="Dạng danh sách"
              >
                <svg className="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
            </div>

            {/* Sort selection dropdown */}
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
        </div>

        {/* Layout Grid: Sidebar + Product Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          
          {/* Left Column: Sidebar Filters */}
          <div className="space-y-6 lg:sticky lg:top-4">
            
            {/* Active Filters list */}
            {(categoryIdParam || brandIdParam || searchQuery || inStockOnly || Object.keys(selectedSpecs).length > 0 || priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max) && (
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
                  {inStockOnly && (
                    <span className="inline-flex items-center gap-1 bg-neutral-100 text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded border border-neutral-200">
                      Còn hàng
                      <button onClick={() => setInStockOnly(false)} className="hover:text-black text-xs">✕</button>
                    </span>
                  )}
                </div>
              </div>
            )}

            {/* In Stock toggle */}
            <div className="border border-neutral-200 bg-white rounded-lg p-5 shadow-sm">
              <label className="flex items-center gap-2.5 text-xs text-neutral-700 font-bold cursor-pointer">
                <input
                  type="checkbox"
                  checked={inStockOnly}
                  onChange={(e) => setInStockOnly(e.target.checked)}
                  className="w-4 h-4 rounded text-black border-neutral-300 focus:ring-black"
                />
                Chỉ hiển thị Còn hàng
              </label>
            </div>

            {/* Price range filter */}
            {priceLimits.max > priceLimits.min && (
              <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-3 shadow-sm">
                <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">Khoảng giá (đ)</h3>
                <input
                  type="range"
                  min={priceLimits.min}
                  max={priceLimits.max}
                  value={priceRange[1]}
                  onChange={(e) => setPriceRange([priceRange[0], parseInt(e.target.value)])}
                  className="w-full h-1 bg-neutral-200 rounded-lg appearance-none cursor-pointer accent-black"
                />
                <div className="flex items-center justify-between text-[10px] font-bold text-neutral-600">
                  <span>{priceLimits.min.toLocaleString('vi-VN')} đ</span>
                  <span>{priceRange[1].toLocaleString('vi-VN')} đ</span>
                </div>
              </div>
            )}

            {/* Categories filter box */}
            <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-3.5 shadow-sm">
              <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">Danh Mục</h3>
              
              {!catalogLoading && categories.length > 8 && (
                <input
                  type="text"
                  placeholder="Tìm danh mục..."
                  value={categorySearch}
                  onChange={(e) => setCategorySearch(e.target.value)}
                  className="w-full text-xs border border-neutral-200 rounded px-2.5 py-1.5 focus:outline-none focus:border-black placeholder-neutral-400 transition-colors focus:ring-1 focus:ring-black bg-neutral-50"
                />
              )}

              {catalogLoading ? (
                <div className="space-y-2">
                  {[...Array(5)].map((_, i) => (
                    <div key={i} className="h-4 w-3/4 bg-neutral-100 animate-pulse rounded"></div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col gap-1.5">
                  <button
                    onClick={() => updateParam('category', null)}
                    className={`text-left text-xs py-1 font-bold transition-colors ${!categoryIdParam ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                  >
                    Tất cả danh mục
                  </button>
                  <div className="flex flex-col gap-1 max-h-48 overflow-y-auto pr-1">
                    {filteredCategories.map((c) => (
                      <button
                        key={c.id}
                        onClick={() => updateParam('category', String(c.id))}
                        className={`text-left text-xs py-1.5 font-bold transition-colors ${categoryIdParam === String(c.id) ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                      >
                        {c.name}
                      </button>
                    ))}
                    {filteredCategories.length === 0 && (
                      <span className="text-[10px] text-neutral-450 italic py-1">Không tìm thấy</span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Brands filter box */}
            <div className="border border-neutral-200 bg-white rounded-lg p-5 space-y-3.5 shadow-sm">
              <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">Thương Hiệu</h3>

              {!catalogLoading && brands.length > 8 && (
                <input
                  type="text"
                  placeholder="Tìm thương hiệu..."
                  value={brandSearch}
                  onChange={(e) => setBrandSearch(e.target.value)}
                  className="w-full text-xs border border-neutral-200 rounded px-2.5 py-1.5 focus:outline-none focus:border-black placeholder-neutral-400 transition-colors focus:ring-1 focus:ring-black bg-neutral-50"
                />
              )}

              {catalogLoading ? (
                <div className="space-y-2">
                  {[...Array(5)].map((_, i) => (
                    <div key={i} className="h-4 w-3/4 bg-neutral-100 animate-pulse rounded"></div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col gap-1.5">
                  <button
                    onClick={() => updateParam('brand', null)}
                    className={`text-left text-xs py-1 font-bold transition-colors ${!brandIdParam ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                  >
                    Tất cả thương hiệu
                  </button>
                  <div className="flex flex-col gap-1 max-h-48 overflow-y-auto pr-1">
                    {filteredBrands.map((b) => (
                      <button
                        key={b.id}
                        onClick={() => updateParam('brand', String(b.id))}
                        className={`text-left text-xs py-1.5 font-bold transition-colors ${brandIdParam === String(b.id) ? 'text-black font-extrabold' : 'text-neutral-500 hover:text-black'}`}
                      >
                        {b.name}
                      </button>
                    ))}
                    {filteredBrands.length === 0 && (
                      <span className="text-[10px] text-neutral-450 italic py-1">Không tìm thấy</span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Dynamic Specs filters */}
            {filterSpecsList.map((spec) => (
              <div key={spec.key} className="border border-neutral-200 bg-white rounded-lg p-5 space-y-3.5 shadow-sm">
                <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider">
                  {spec.key}
                </h3>
                <div className="flex flex-col gap-1.5 max-h-40 overflow-y-auto pr-1">
                  {spec.options.map((opt) => {
                    const isChecked = selectedSpecs[spec.key]?.includes(opt) || false
                    return (
                      <label
                        key={opt}
                        className="flex items-center gap-2 text-xs text-neutral-600 hover:text-black cursor-pointer"
                      >
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={() => handleToggleSpecOption(spec.key, opt)}
                          className="w-3.5 h-3.5 rounded border-neutral-300 text-black focus:ring-black"
                        />
                        <span className="truncate">{opt}</span>
                      </label>
                    )
                  })}
                </div>
              </div>
            ))}

          </div>

          {/* Right Column: Products List/Grid */}
          <div className="lg:col-span-3 space-y-8">
            
            {/* Loading state or products display */}
            {loadingProducts ? (
              <SearchSkeleton viewMode={viewMode} count={9} />
            ) : (
              <>
                {viewMode === 'grid' ? (
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
                    {filteredProducts.map((product) => (
                      <ProductCard key={product.id} product={product} />
                    ))}
                  </div>
                ) : (
                  <div className="space-y-4">
                    {filteredProducts.map((product) => {
                      const isAddedLoading = loadingCartMap[product.id] || false
                      const displayPrice = product.discount_price || product.price || 0
                      const originalPrice = product.price || 0
                      const hasDiscount = !!(product.discount_price && product.price && product.discount_price < product.price)

                      // Extract specifications snapshot
                      let parsedSpecs: any = {}
                      if (product.specs_jsonb) {
                        try {
                          parsedSpecs = typeof product.specs_jsonb === 'string'
                            ? JSON.parse(product.specs_jsonb)
                            : product.specs_jsonb
                        } catch (e) {
                          // ignore
                        }
                      }
                      
                      const flatSpecs: string[] = []
                      Object.keys(parsedSpecs).slice(0, 2).forEach(g => {
                        const group = parsedSpecs[g]
                        if (group && typeof group === 'object') {
                          Object.keys(group).slice(0, 2).forEach(k => {
                            const item = group[k]
                            const text = `${k}: ${item?.raw_value || item?.value || ''}`
                            if (text.length > 5 && text.length < 32) flatSpecs.push(text)
                          })
                        }
                      })

                      return (
                        <div
                          key={product.id}
                          className="bg-white border border-neutral-200 rounded-xl overflow-hidden shadow-sm flex flex-col sm:flex-row p-4 gap-6 hover:shadow-md transition-all group"
                        >
                          {/* Image */}
                          <div className="w-full sm:w-40 aspect-square bg-neutral-50 flex items-center justify-center p-3 relative rounded-lg overflow-hidden flex-shrink-0">
                            <Link to={`/products/${product.id}`} className="w-full h-full flex items-center justify-center">
                              <img
                                src={product.image || '/placeholder-product.png'}
                                alt={product.name}
                                className="object-contain max-h-full max-w-full mix-blend-multiply group-hover:scale-105 transition-transform duration-300"
                              />
                            </Link>
                            {product.stock === 0 && (
                              <span className="absolute inset-0 bg-white/75 flex items-center justify-center text-xs font-bold text-neutral-800 uppercase">
                                Hết hàng
                              </span>
                            )}
                          </div>

                          {/* Details */}
                          <div className="flex-1 flex flex-col justify-between py-1.5 space-y-3">
                            <div className="space-y-1">
                              {product.brand && (
                                <span className="text-[10px] font-bold text-neutral-400 uppercase tracking-widest">
                                  {product.brand.name}
                                </span>
                              )}
                              <Link to={`/products/${product.id}`}>
                                <h3 className="text-sm font-bold text-neutral-800 group-hover:text-black transition-colors leading-tight">
                                  {product.name}
                                </h3>
                              </Link>
                              
                              <div className="flex items-center gap-1.5 text-xs text-neutral-500 pt-1">
                                <span className="text-amber-500 font-bold">★ {product.rating ? product.rating.toFixed(1) : '0.0'}</span>
                                <span className="text-neutral-300">|</span>
                                <span>{product.review_count} Đánh giá</span>
                              </div>

                              {flatSpecs.length > 0 && (
                                <div className="flex flex-wrap gap-1.5 pt-2">
                                  {flatSpecs.map((s, i) => (
                                    <span key={i} className="bg-neutral-100 text-neutral-600 text-[10px] px-2 py-0.5 rounded font-medium border border-neutral-200/50">
                                      {s}
                                    </span>
                                  ))}
                                </div>
                              )}
                            </div>

                            <div className="text-[11px] font-semibold text-neutral-500">
                              Tình trạng: <span className={product.stock > 0 ? 'text-emerald-600' : 'text-neutral-400'}>
                                {product.stock > 0 ? `Còn ${product.stock} hàng` : 'Hết hàng'}
                              </span>
                            </div>
                          </div>

                          {/* Pricing column */}
                          <div className="sm:w-48 sm:border-l border-neutral-100 sm:pl-6 flex flex-row sm:flex-col justify-between sm:justify-center items-center sm:items-end gap-4">
                            <div className="text-right">
                              {hasDiscount && (
                                <span className="text-xs text-neutral-400 line-through block">
                                  {originalPrice.toLocaleString('vi-VN')} đ
                                </span>
                              )}
                              <span className="text-base font-black text-neutral-900 block mt-0.5">
                                {displayPrice.toLocaleString('vi-VN')} đ
                              </span>
                            </div>

                            <button
                              onClick={(e) => handleQuickAddList(e, product)}
                              disabled={product.stock === 0 || isAddedLoading}
                              className="bg-black hover:bg-neutral-800 text-white disabled:opacity-40 text-xs font-bold px-4 py-2.5 rounded-lg flex items-center justify-center gap-2 transition-all duration-200 active:scale-95"
                            >
                              {isAddedLoading ? (
                                <svg className="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                                </svg>
                              ) : (
                                <>
                                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
                                  </svg>
                                  Mua ngay
                                </>
                              )}
                            </button>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}

                {filteredProducts.length === 0 && (
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
                    {getPageNumbers.map((pageNum, idx) => {
                      if (pageNum === '...') {
                        return (
                          <span key={`dots-${idx}`} className="px-2 text-neutral-400 text-xs font-bold select-none">
                            ...
                          </span>
                        )
                      }
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
