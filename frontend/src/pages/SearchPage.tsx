import { useEffect, useState, useTransition, useMemo } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import SearchBar from '../components/SearchBar'
import ProductCard from '../components/ProductCard'
import SearchSkeleton from '../components/SearchSkeleton'
import { productAPI } from '../services/productAPI'
import { useCart } from '../hooks/useCart'
import type { Product } from '../types'


const priceOptions = [
  { label: 'Dưới 2 triệu', min: 0, max: 2000000 },
  { label: 'Từ 2 - 4 triệu', min: 2000000, max: 4000000 },
  { label: 'Từ 4 - 7 triệu', min: 4000000, max: 7000000 },
  { label: 'Từ 7 - 13 triệu', min: 7000000, max: 13000000 },
  { label: 'Từ 13 - 20 triệu', min: 13000000, max: 20000000 },
  { label: 'Trên 20 triệu', min: 20000000, max: 100000000 },
]

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

  // Recommendations when query is empty or no results
  const [featuredProducts, setFeaturedProducts] = useState<Product[]>([])

  // Layout and Sidebar Filters State
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [inStockOnly, setInStockOnly] = useState(false)
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 100000000])
  const [selectedPriceRanges, setSelectedPriceRanges] = useState<string[]>([])
  const [selectedSpecs, setSelectedSpecs] = useState<{ [specKey: string]: string[] }>({})
  
  // Dropdown states for TGDĐ style filter bar
  const [openDropdown, setOpenDropdown] = useState<string | null>(null)
  const [tempSelectedSpecs, setTempSelectedSpecs] = useState<{ [specKey: string]: string[] }>({})
  const [tempPriceRange, setTempPriceRange] = useState<[number, number]>([0, 100000000])
  const [tempSelectedPriceRanges, setTempSelectedPriceRanges] = useState<string[]>([])

  const { addToCart } = useCart()
  const [loadingCartMap, setLoadingCartMap] = useState<{ [key: string]: boolean }>({})

  // Fetch recommendations
  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        const res = await productAPI.getProducts({ limit: 8, sort: 'popular' })
        setFeaturedProducts(res.data)
      } catch (err) {
        console.error('Failed to load recommendations', err)
      }
    }
    void fetchRecommendations()
  }, [])

  // Fetch search results
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
          limit: 24, // Fetch more for richer client-side filtering options
          sort: activeSort,
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
        console.error('Failed to load search results', err)
      } finally {
        setLoading(false)
      }
    }

    void fetchSearchResults()
    // Reset filters when a new search query is executed
    setInStockOnly(false)
    setSelectedSpecs({})
    setSelectedPriceRanges([])
  }, [searchQuery, activeSort, activePage])

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
      .filter((key) => specsMap[key].size > 1) // only filter by fields with varying choices
      .map((key) => ({
        key,
        options: Array.from(specsMap[key]).slice(0, 8), // cap values for design neatness
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

  // Get dynamic boundary prices from results
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

  // Perform Client-Side Filtering
  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      // 1. Stock filter
      if (inStockOnly && product.stock <= 0) return false

      // 2. Price filter
      const displayPrice = product.discount_price || product.price || 0
      
      // If we have selected price range pills, check if the price fits in any of them
      if (selectedPriceRanges.length > 0) {
        const matchesAnyRange = selectedPriceRanges.some(rangeLabel => {
          const opt = priceOptions.find(o => o.label === rangeLabel)
          if (!opt) return false
          return displayPrice >= opt.min && displayPrice <= opt.max
        })
        if (!matchesAnyRange) return false
      } else {
        // Fallback to slider range
        if (displayPrice < priceRange[0] || displayPrice > priceRange[1]) return false
      }

      // 3. Facet specifications filter
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
  }, [products, inStockOnly, priceRange, selectedSpecs, selectedPriceRanges])

  const handleSearchSubmit = (newQuery: string) => {
    startTransition(() => {
      setSearchParams({ q: newQuery })
    })
  }

  const updateParam = (key: string, value: string | null) => {
    startTransition(() => {
      setSearchParams((prev) => {
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

  // Dropdown open helper
  const handleOpenDropdown = (dropdownKey: string) => {
    setOpenDropdown(dropdownKey)
    if (dropdownKey === 'price') {
      setTempPriceRange([...priceRange] as [number, number])
      setTempSelectedPriceRanges([...selectedPriceRanges])
    } else {
      setTempSelectedSpecs({ ...selectedSpecs })
    }
  }


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

  // Predefined keyword shortcuts
  const trendingKeywords = ['iPhone 17', 'S24 Ultra', 'Loa Bluetooth', 'Tai nghe AVA', 'Laptop Asus']

  return (
    <div className="bg-neutral-50 min-h-screen py-8 font-sans transition-all duration-300">
      <div className="mx-auto max-w-7xl px-4 space-y-8">
        
        {/* Breadcrumbs */}
        <nav className="flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">
            Trang chủ
          </Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Tìm kiếm</span>
        </nav>

        {/* Dynamic Interactive Header Section */}
        <div className="relative rounded-2xl overflow-hidden bg-gradient-to-br from-neutral-900 via-neutral-950 to-neutral-900 text-white p-8 md:p-12 shadow-xl border border-neutral-800">
          <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-neutral-800/40 via-transparent to-transparent pointer-events-none"></div>
          <div className="max-w-2xl mx-auto text-center space-y-6 relative z-10">
            <span className="bg-amber-400/10 text-amber-300 border border-amber-400/20 text-[10px] uppercase font-bold tracking-widest px-3 py-1 rounded-full">
              Khám phá Kho Công Nghệ
            </span>
            <h1 className="text-3xl md:text-4xl font-extrabold tracking-tight bg-gradient-to-r from-white via-neutral-200 to-neutral-400 bg-clip-text text-transparent">
              Bắt đầu tìm kiếm sản phẩm
            </h1>
            <p className="text-xs text-neutral-400 max-w-md mx-auto leading-relaxed">
              Nhập từ khóa thiết bị, mã model hoặc thông số kỹ thuật để bắt đầu mua sắm ngay lập tức.
            </p>
            <div className="w-full">
              <SearchBar onSearch={handleSearchSubmit} initialValue={searchQuery} />
            </div>
            
            {/* Trending shortcuts */}
            <div className="flex flex-wrap items-center justify-center gap-2 pt-2 text-xs">
              <span className="text-neutral-500 font-medium mr-1">Tìm nhiều nhất:</span>
              {trendingKeywords.map((kw) => (
                <button
                  key={kw}
                  onClick={() => handleSearchSubmit(kw)}
                  className="bg-neutral-800/60 hover:bg-neutral-700 text-neutral-300 hover:text-white px-3 py-1 rounded-full border border-neutral-700/50 hover:border-neutral-500 transition-all duration-200 active:scale-95 text-[11px]"
                >
                  {kw}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Results Block */}
        {searchQuery ? (
          <div className="border-t border-neutral-200 pt-6 space-y-6">
            
            {/* Backdrop for filter dropdowns */}
            {openDropdown && (
              <div className="fixed inset-0 z-35 bg-transparent" onClick={() => setOpenDropdown(null)} />
            )}

            {/* Horizontal Dropdown Filter Bar */}
            <div className="relative z-45 flex flex-wrap items-center gap-2.5">
              {/* Precio (Price) Dropdown Button */}
              <div className="relative">
                <button
                  onClick={() => {
                    if (openDropdown === 'price') {
                      setOpenDropdown(null)
                    } else {
                      handleOpenDropdown('price')
                    }
                  }}
                  className={`flex items-center gap-1.5 px-4 py-2 rounded-lg border text-xs font-bold transition-all shadow-sm ${
                    openDropdown === 'price'
                      ? 'border-brand-600 bg-brand-50/30 text-brand-700 ring-1 ring-brand-100'
                      : selectedPriceRanges.length > 0 || priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max
                      ? 'border-amber-500 bg-amber-50/20 text-amber-700 font-extrabold'
                      : 'border-neutral-200 bg-white text-neutral-700 hover:border-neutral-350'
                  }`}
                >
                  <span>Giá</span>
                  {selectedPriceRanges.length > 0 && (
                    <span className="bg-amber-500 text-white text-[9px] w-4 h-4 rounded-full flex items-center justify-center font-black">
                      {selectedPriceRanges.length}
                    </span>
                  )}
                  <svg className={`w-3 h-3 transition-transform duration-200 ${openDropdown === 'price' ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M19 9l-7 7-7-7" />
                  </svg>
                </button>

                {/* Price Dropdown Panel */}
                {openDropdown === 'price' && (
                  <div className="absolute left-0 mt-2 w-72 bg-white border border-neutral-200 rounded-xl shadow-glass p-5 z-50 animate-fade-in-up">
                    <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider mb-3">Chọn khoảng giá</h3>
                    
                    {/* Price Pills */}
                    <div className="grid grid-cols-2 gap-2 mb-4">
                      {priceOptions.map((opt) => {
                        const isSelected = tempSelectedPriceRanges.includes(opt.label)
                        return (
                          <button
                            key={opt.label}
                            onClick={() => {
                              setTempSelectedPriceRanges(prev =>
                                prev.includes(opt.label)
                                  ? prev.filter(l => l !== opt.label)
                                  : [...prev, opt.label]
                              )
                            }}
                            className={`px-2 py-2 text-[10px] font-bold rounded-md border text-center transition-all ${
                              isSelected
                                ? 'border-brand-600 bg-brand-50/50 text-brand-700'
                                : 'border-neutral-200 bg-white text-neutral-600 hover:border-neutral-300'
                            }`}
                          >
                            {opt.label}
                          </button>
                        )
                      })}
                    </div>

                    {/* Range Slider (Fallback when no pills selected) */}
                    {tempSelectedPriceRanges.length === 0 && priceLimits.max > priceLimits.min && (
                      <div className="space-y-2 pt-2 border-t border-neutral-100">
                        <span className="text-[10px] font-black text-neutral-400 uppercase tracking-wider">Hoặc kéo khoảng giá</span>
                        <input
                          type="range"
                          min={priceLimits.min}
                          max={priceLimits.max}
                          value={tempPriceRange[1]}
                          onChange={(e) => setTempPriceRange([tempPriceRange[0], parseInt(e.target.value)])}
                          className="w-full h-1 bg-neutral-200 rounded-lg appearance-none cursor-pointer accent-brand-500"
                        />
                        <div className="flex items-center justify-between text-[9px] font-black text-neutral-500">
                          <span>{priceLimits.min.toLocaleString('vi-VN')} đ</span>
                          <span>{tempPriceRange[1].toLocaleString('vi-VN')} đ</span>
                        </div>
                      </div>
                    )}

                    {/* Footer Controls */}
                    <div className="flex items-center justify-between mt-5 pt-3.5 border-t border-neutral-100">
                      <button
                        onClick={() => {
                          setTempSelectedPriceRanges([])
                          setTempPriceRange([priceLimits.min, priceLimits.max])
                        }}
                        className="text-[10px] font-bold text-neutral-450 hover:text-neutral-700 uppercase"
                      >
                        Mặc định
                      </button>
                      <button
                        onClick={() => {
                          setPriceRange(tempPriceRange)
                          setSelectedPriceRanges(tempSelectedPriceRanges)
                          setOpenDropdown(null)
                        }}
                        className="bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase px-4 py-1.5 rounded-md transition-all shadow-sm"
                      >
                        Xem kết quả
                      </button>
                    </div>
                  </div>
                )}
              </div>

              {/* Dynamic Spec Filter Dropdowns */}
              {filterSpecsList.map((spec) => {
                const hasActiveFilter = selectedSpecs[spec.key]?.length > 0
                const isActiveDropdown = openDropdown === spec.key

                return (
                  <div key={spec.key} className="relative">
                    <button
                      onClick={() => {
                        if (isActiveDropdown) {
                          setOpenDropdown(null)
                        } else {
                          handleOpenDropdown(spec.key)
                        }
                      }}
                      className={`flex items-center gap-1.5 px-4 py-2 rounded-lg border text-xs font-bold transition-all shadow-sm ${
                        isActiveDropdown
                          ? 'border-brand-600 bg-brand-50/30 text-brand-700 ring-1 ring-brand-100'
                          : hasActiveFilter
                          ? 'border-amber-500 bg-amber-50/20 text-amber-700 font-extrabold'
                          : 'border-neutral-200 bg-white text-neutral-700 hover:border-neutral-350'
                      }`}
                    >
                      <span>{spec.key}</span>
                      {hasActiveFilter && (
                        <span className="bg-amber-500 text-white text-[9px] w-4 h-4 rounded-full flex items-center justify-center font-black">
                          {selectedSpecs[spec.key].length}
                        </span>
                      )}
                      <svg className={`w-3 h-3 transition-transform duration-200 ${isActiveDropdown ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M19 9l-7 7-7-7" />
                      </svg>
                    </button>

                    {/* Dropdown Panel */}
                    {isActiveDropdown && (
                      <div className="absolute left-0 mt-2 w-80 bg-white border border-neutral-200 rounded-xl shadow-glass p-5 z-50 animate-fade-in-up">
                        <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider mb-3">Lọc theo {spec.key}</h3>
                        
                        {/* Options Pills */}
                        <div className="flex flex-wrap gap-2 mb-4 max-h-48 overflow-y-auto pr-1">
                          {spec.options.map((opt) => {
                            const isSelected = tempSelectedSpecs[spec.key]?.includes(opt) || false
                            return (
                              <button
                                key={opt}
                                onClick={() => {
                                  setTempSelectedSpecs(prev => {
                                    const currentList = prev[spec.key] || []
                                    const nextList = currentList.includes(opt)
                                      ? currentList.filter(v => v !== opt)
                                      : [...currentList, opt]
                                    const next = { ...prev, [spec.key]: nextList }
                                    if (nextList.length === 0) {
                                      delete next[spec.key]
                                    }
                                    return next
                                  })
                                }}
                                className={`px-3 py-1.5 text-[10px] font-bold rounded-md border transition-all ${
                                  isSelected
                                    ? 'border-brand-600 bg-brand-50/50 text-brand-700'
                                    : 'border-neutral-200 bg-white text-neutral-600 hover:border-neutral-350'
                                }`}
                              >
                                {opt}
                              </button>
                            )
                          })}
                        </div>

                        {/* Footer Controls */}
                        <div className="flex items-center justify-between pt-3.5 border-t border-neutral-100">
                          <button
                            onClick={() => {
                              setTempSelectedSpecs(prev => {
                                const next = { ...prev }
                                delete next[spec.key]
                                return next
                              })
                            }}
                            className="text-[10px] font-bold text-neutral-455 hover:text-neutral-700 uppercase"
                          >
                            Hủy
                          </button>
                          <button
                            onClick={() => {
                              setSelectedSpecs(tempSelectedSpecs)
                              setOpenDropdown(null)
                            }}
                            className="bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase px-4 py-1.5 rounded-md transition-all shadow-sm"
                          >
                            Xem kết quả
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                )
              })}

              {/* Quick Toggle: Chỉ hiển thị Còn hàng */}
              <button
                onClick={() => setInStockOnly(!inStockOnly)}
                className={`flex items-center gap-1.5 px-4 py-2 rounded-lg border text-xs font-bold transition-all shadow-sm ${
                  inStockOnly
                    ? 'border-brand-600 bg-brand-50/30 text-brand-700 ring-1 ring-brand-100 font-extrabold'
                    : 'border-neutral-200 bg-white text-neutral-700 hover:border-neutral-350'
                }`}
              >
                <span>Còn hàng</span>
                {inStockOnly && (
                  <svg className="w-3 h-3 text-brand-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
            </div>

            {/* Active Filter Pills List */}
            {(inStockOnly || Object.keys(selectedSpecs).length > 0 || selectedPriceRanges.length > 0 || priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max) && (
              <div className="flex flex-wrap items-center gap-2 bg-neutral-100/50 border border-neutral-200/60 p-3 rounded-lg">
                <span className="text-[10px] font-black text-neutral-500 uppercase tracking-wider mr-1">Đang lọc:</span>
                
                {inStockOnly && (
                  <span className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                    Còn hàng
                    <button onClick={() => setInStockOnly(false)} className="hover:text-red-500 text-xs font-black">✕</button>
                  </span>
                )}
                {selectedPriceRanges.map((l) => (
                  <span key={l} className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                    Giá: {l}
                    <button
                      onClick={() => setSelectedPriceRanges(prev => prev.filter(v => v !== l))}
                      className="hover:text-red-500 text-xs font-black"
                    >
                      ✕
                    </button>
                  </span>
                ))}
                {selectedPriceRanges.length === 0 && (priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max) && (
                  <span className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                    Giá: {priceRange[0].toLocaleString('vi-VN')} - {priceRange[1].toLocaleString('vi-VN')} đ
                    <button
                      onClick={() => setPriceRange([priceLimits.min, priceLimits.max])}
                      className="hover:text-red-500 text-xs font-black"
                    >
                      ✕
                    </button>
                  </span>
                )}
                {Object.entries(selectedSpecs).map(([key, values]) =>
                  values.map((v) => (
                    <span key={`${key}-${v}`} className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                      {key}: {v}
                      <button
                        onClick={() => {
                          setSelectedSpecs(prev => {
                            const list = prev[key].filter(item => item !== v)
                            const next = { ...prev, [key]: list }
                            if (list.length === 0) delete next[key]
                            return next
                          })
                        }}
                        className="hover:text-red-500 text-xs font-black"
                      >
                        ✕
                      </button>
                    </span>
                  ))
                )}

                <button
                  onClick={() => {
                    setInStockOnly(false)
                    setSelectedSpecs({})
                    setSelectedPriceRanges([])
                    setPriceRange([priceLimits.min, priceLimits.max])
                  }}
                  className="text-[10px] font-black text-red-500 hover:text-red-650 transition-colors uppercase ml-auto"
                >
                  Xóa tất cả bộ lọc
                </button>
              </div>
            )}

            {/* Products Area */}
            <div className="w-full space-y-6">
              
              {/* Search result controls and quick layout toggles */}
              <div className="flex items-center justify-between gap-4 flex-wrap bg-white p-4 rounded-xl border border-neutral-200 shadow-sm">
                <div>
                  <h2 className="text-xs font-black uppercase tracking-wider text-neutral-800">
                    Kết quả tìm kiếm cho: <span className="text-neutral-500 font-semibold">"{searchQuery}"</span>
                  </h2>
                  <p className="text-[10px] text-neutral-400 mt-0.5">
                    Tìm thấy {filteredProducts.length} trên tổng số {pagination.total} sản phẩm
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
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                      </svg>
                    </button>
                    <button
                      onClick={() => setViewMode('list')}
                      className={`p-2 transition-colors ${viewMode === 'list' ? 'bg-neutral-900 text-white' : 'bg-white text-neutral-600 hover:bg-neutral-100'}`}
                      title="Dạng danh sách"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16" />
                      </svg>
                    </button>
                  </div>

                  {/* Sorting */}
                  <div className="flex items-center gap-2">
                    <select
                      value={activeSort}
                      onChange={(e) => updateParam('sort', e.target.value)}
                      className="bg-white border border-neutral-200 text-xs font-bold px-3 py-1.5 rounded focus:outline-none focus:border-black cursor-pointer shadow-sm"
                    >
                      <option value="newest">Mới nhất</option>
                      <option value="popular">Bán chạy nhất</option>
                      <option value="price_asc">Giá: Thấp đến Cao</option>
                      <option value="price_desc">Giá: Cao đến Thấp</option>
                    </select>
                  </div>
                </div>
              </div>

              {/* Main Results Wrapper */}
              <div className="min-h-[400px]">
                {loading ? (
                  <SearchSkeleton viewMode={viewMode} count={10} />
                ) : (
                  <>
                    {/* Empty State Grid */}
                    {filteredProducts.length === 0 && (
                      <div className="bg-white border border-neutral-200 border-dashed rounded-xl p-12 text-center max-w-lg mx-auto space-y-4 my-10">
                        <div className="w-16 h-16 bg-neutral-100 rounded-full flex items-center justify-center mx-auto text-neutral-400">
                          <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        </div>
                        <div>
                          <h3 className="text-sm font-bold text-neutral-800 uppercase tracking-wider">Không tìm thấy sản phẩm</h3>
                          <p className="text-xs text-neutral-450 mt-1 leading-relaxed">
                            Không có sản phẩm nào khớp với các bộ lọc hiện tại của bạn. Thử xóa bớt bộ lọc hoặc tìm từ khóa khác.
                          </p>
                        </div>
                        <button
                          onClick={() => {
                            setInStockOnly(false)
                            setPriceRange([priceLimits.min, priceLimits.max])
                            setSelectedSpecs({})
                            setSelectedPriceRanges([])
                          }}
                            className="bg-black hover:bg-neutral-800 text-white text-xs font-bold px-4 py-2 rounded transition-all active:scale-95"
                          >
                            Xóa bộ lọc để thử lại
                          </button>
                        </div>
                      )}

                      {/* Display Products */}
                      {viewMode === 'grid' ? (
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 md:gap-5">
                          {filteredProducts.map((product) => (
                            <div key={product.id} className="transition-all duration-300 hover:-translate-y-1">
                              <ProductCard product={product} />
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="space-y-4">
                          {filteredProducts.map((product) => {
                            const isAddedLoading = loadingCartMap[product.id] || false
                            const displayPrice = product.discount_price || product.price || 0
                            const originalPrice = product.price || 0
                            const hasDiscount = !!(product.discount_price && product.price && product.discount_price < product.price)

                            // Quick specifications preview for list layout
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
                                {/* Left thumbnail */}
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

                                {/* Middle details info */}
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
                                    
                                    {/* Star Rating */}
                                    <div className="flex items-center gap-1.5 text-xs text-neutral-500 pt-1">
                                      <span className="text-amber-500 font-bold">★ {product.rating ? product.rating.toFixed(1) : '0.0'}</span>
                                      <span className="text-neutral-300">|</span>
                                      <span>{product.review_count} Đánh giá</span>
                                    </div>

                                    {/* Specifications snippet preview */}
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

                                  {/* Availability stock level */}
                                  <div className="text-[11px] font-semibold text-neutral-500">
                                    Tình trạng: <span className={product.stock > 0 ? 'text-emerald-600' : 'text-neutral-400'}>
                                      {product.stock > 0 ? `Còn ${product.stock} hàng` : 'Hết hàng'}
                                    </span>
                                  </div>
                                </div>

                                {/* Right price column */}
                                <div className="sm:w-48 sm:border-l border-neutral-100 sm:pl-6 flex flex-row sm:flex-col justify-between sm:justify-center items-center sm:items-end gap-4">
                                  <div className="text-right flex flex-col justify-center sm:justify-start">
                                    {hasDiscount && (
                                      <span className="text-xs text-neutral-400 line-through block">
                                        {originalPrice.toLocaleString('vi-VN')} đ
                                      </span>
                                    )}
                                    <span className="text-base font-black text-neutral-900 block mt-0.5">
                                      {displayPrice.toLocaleString('vi-VN')} đ
                                    </span>
                                    {hasDiscount && (
                                      <span className="bg-red-50 text-red-500 text-[9px] font-extrabold uppercase px-1.5 py-0.5 rounded tracking-wide inline-block mt-1">
                                        Giảm {Math.round(((originalPrice - displayPrice) / originalPrice) * 100)}%
                                      </span>
                                    )}
                                  </div>

                                  <button
                                    onClick={(e) => handleQuickAddList(e, product)}
                                    disabled={product.stock === 0 || isAddedLoading}
                                    className="bg-black hover:bg-neutral-800 text-white disabled:opacity-40 disabled:hover:bg-black text-xs font-bold px-4 py-2.5 rounded-lg flex items-center justify-center gap-2 transition-all duration-200 active:scale-95"
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

                      {/* Pagination controls */}
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
            </div>
        ) : (
          /* Pre-Search Engagement Layout */
          <div className="space-y-12">
            
            {/* Promo grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="p-6 bg-amber-50 border border-amber-200/60 rounded-2xl flex items-start gap-4 hover:shadow-md transition-shadow">
                <span className="text-3xl text-amber-500">✨</span>
                <div className="space-y-1">
                  <h3 className="text-xs font-black uppercase text-amber-800">Thành viên mới</h3>
                  <p className="text-[11px] text-amber-700 leading-relaxed">
                    Giảm ngay 10% cho hóa đơn đầu tiên. Nhập mã code <strong className="text-amber-900 bg-amber-200/50 px-1 py-0.5 rounded">NEWCOMMER10</strong> khi thanh toán.
                  </p>
                </div>
              </div>
              <div className="p-6 bg-sky-50 border border-sky-200/60 rounded-2xl flex items-start gap-4 hover:shadow-md transition-shadow">
                <span className="text-3xl text-sky-500">🚚</span>
                <div className="space-y-1">
                  <h3 className="text-xs font-black uppercase text-sky-800">Miễn phí vận chuyển</h3>
                  <p className="text-[11px] text-sky-700 leading-relaxed">
                    Hỗ trợ giao hàng siêu tốc trong vòng 2H với tất cả đơn hàng trị giá từ 500K trở lên.
                  </p>
                </div>
              </div>
              <div className="p-6 bg-purple-50 border border-purple-200/60 rounded-2xl flex items-start gap-4 hover:shadow-md transition-shadow">
                <span className="text-3xl text-purple-500">🛡️</span>
                <div className="space-y-1">
                  <h3 className="text-xs font-black uppercase text-purple-800">Bảo hành 1 đổi 1</h3>
                  <p className="text-[11px] text-purple-700 leading-relaxed">
                    Cam kết chính hãng 100%. Bảo hành lỗi kỹ thuật phần cứng 1 đổi 1 trong thời hạn lên đến 12 tháng.
                  </p>
                </div>
              </div>
            </div>

            {/* Popular Categories Grid */}
            <div className="space-y-4">
              <div className="text-center md:text-left">
                <h2 className="text-xs font-black uppercase tracking-wider text-neutral-800">Danh mục được quan tâm</h2>
                <p className="text-[10px] text-neutral-400">Xem nhanh danh sách sản phẩm theo nhu cầu</p>
              </div>
              
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4">
                {[
                  { name: 'Điện thoại', q: 'iPhone', icon: '📱', color: 'bg-rose-50 hover:bg-rose-100 border-rose-200/40 text-rose-800' },
                  { name: 'Laptop', q: 'Asus', icon: '💻', color: 'bg-indigo-50 hover:bg-indigo-100 border-indigo-200/40 text-indigo-800' },
                  { name: 'Loa âm thanh', q: 'Loa', icon: '🔊', color: 'bg-teal-50 hover:bg-teal-100 border-teal-200/40 text-teal-800' },
                  { name: 'Tai nghe', q: 'AVA', icon: '🎧', color: 'bg-emerald-50 hover:bg-emerald-100 border-emerald-200/40 text-emerald-800' },
                  { name: 'Phụ kiện', q: 'Sạc', icon: '🔋', color: 'bg-amber-50 hover:bg-amber-100 border-amber-200/40 text-amber-800' },
                ].map((cat) => (
                  <button
                    key={cat.name}
                    onClick={() => handleSearchSubmit(cat.q)}
                    className={`p-5 rounded-2xl border text-center flex flex-col items-center justify-center gap-3 transition-all duration-300 hover:scale-105 active:scale-95 hover:shadow-premium-soft ${cat.color}`}
                  >
                    <span className="text-2xl">{cat.icon}</span>
                    <span className="text-xs font-black uppercase tracking-wide">{cat.name}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Featured Products/Recommendations section */}
            {featuredProducts.length > 0 && (
              <div className="space-y-4 pt-4">
                <div className="text-center md:text-left">
                  <h2 className="text-xs font-black uppercase tracking-wider text-neutral-800">Sản phẩm nổi bật bán chạy</h2>
                  <p className="text-[10px] text-neutral-400">Xu hướng mua sắm nổi bật nhất tuần qua</p>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                  {featuredProducts.map((prod) => (
                    <ProductCard key={prod.id} product={prod} />
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export default SearchPage
