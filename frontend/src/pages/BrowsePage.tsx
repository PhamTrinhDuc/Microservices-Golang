import { useEffect, useState, useTransition, useMemo } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import SearchSkeleton from '../components/SearchSkeleton'
import { useCatalog } from '../hooks/useCatalog'
import { productAPI } from '../services/productAPI'
import { bannerAPI } from '../services/bannerAPI'
import { useCart } from '../hooks/useCart'
import type { Product, Banner } from '../types'


const brandLogos: { [key: string]: string } = {
  apple: ' Apple',
  samsung: 'SAMSUNG',
  xiaomi: 'Xiaomi',
  oppo: 'OPPO',
  realme: 'realme',
  vivo: 'vivo',
  asus: 'ASUS',
  hp: 'HP',
  lenovo: 'Lenovo',
  dell: 'DELL',
  msi: 'msi',
  acer: 'acer'
}

const priceOptions = [
  { label: 'Dưới 2 triệu', min: 0, max: 2000000 },
  { label: 'Từ 2 - 4 triệu', min: 2000000, max: 4000000 },
  { label: 'Từ 4 - 7 triệu', min: 4000000, max: 7000000 },
  { label: 'Từ 7 - 13 triệu', min: 7000000, max: 13000000 },
  { label: 'Từ 13 - 20 triệu', min: 13000000, max: 20000000 },
  { label: 'Trên 20 triệu', min: 20000000, max: 100000000 },
]

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

  // Layout & Client Side Filtering State
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
  
  // Banners & Modal State
  const [categoryBanners, setCategoryBanners] = useState<Banner[]>([])
  const [currentSlide, setCurrentSlide] = useState(0)
  const [isFilterModalOpen, setIsFilterModalOpen] = useState(false)

  // Fetch category-specific banners when categoryParam changes
  useEffect(() => {
    const fetchCategoryBanners = async () => {
      try {
        const params: any = {}
        if (categoryIdParam) {
          params.category_id = parseInt(categoryIdParam)
        }
        const data = await bannerAPI.listBanners(params)
        // If we are looking for a specific category, filter banners belonging to it
        if (categoryIdParam) {
          const catId = parseInt(categoryIdParam)
          const filtered = data.filter(b => b.category_id === catId)
          setCategoryBanners(filtered)
        } else {
          // If viewing "All Products", show general banners
          const general = data.filter(b => !b.category_id)
          setCategoryBanners(general)
        }
        setCurrentSlide(0)
      } catch (err) {
        console.error('Failed to load category banners', err)
      }
    }
    void fetchCategoryBanners()
  }, [categoryIdParam])

  // Auto rotate banners
  useEffect(() => {
    if (categoryBanners.length <= 1) return
    const timer = setInterval(() => {
      setCurrentSlide(prev => (prev + 1) % categoryBanners.length)
    }, 5000)
    return () => clearInterval(timer)
  }, [categoryBanners])

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
    setSelectedPriceRanges([])
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

  const updateCategoryAndBrand = (categoryId: string | null, brandId: string | null) => {
    startTransition(() => {
      setSearchParams(prev => {
        const next = new URLSearchParams(prev)
        if (categoryId === null) {
          next.delete('category')
        } else {
          next.set('category', categoryId)
        }
        if (brandId === null) {
          next.delete('brand')
        } else {
          next.set('brand', brandId)
        }
        next.delete('page')
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
    setSelectedPriceRanges([])
    setPriceRange([priceLimits.min, priceLimits.max])
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

  const openFilterModal = () => {
    setTempSelectedSpecs(JSON.parse(JSON.stringify(selectedSpecs)))
    setTempSelectedPriceRanges([...selectedPriceRanges])
    setTempPriceRange([...priceRange])
    setIsFilterModalOpen(true)
  }

  const selectedCategory = categories.find(c => String(c.id) === categoryIdParam)
  const selectedBrand = brands.find(b => String(b.id) === brandIdParam)


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

  // Derive brands that actually appear in the loaded products for this category
  const availableBrands = useMemo(() => {
    if (!products.length) return []
    const brandIdSet = new Set<number>()
    products.forEach(p => { if (p.brand?.id) brandIdSet.add(p.brand.id) })
    return brands.filter(b => brandIdSet.has(b.id))
  }, [products, brands])


  // Client side filtration
  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      if (inStockOnly && product.stock <= 0) return false

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

      for (const [specKey, values] of Object.entries(selectedSpecs)) {
        if (!values || values.length === 0) continue

        // Special key for brand filtering from the modal
        if (specKey === '__brand__') {
          if (!values.includes(String(product.brand?.id))) return false
          continue
        }

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

  // Memoized count of matching products based on temp filters inside the modal
  const tempFilteredProductsCount = useMemo(() => {
    return products.filter((product) => {
      if (inStockOnly && product.stock <= 0) return false

      const displayPrice = product.discount_price || product.price || 0
      
      // Price Check
      if (tempSelectedPriceRanges.length > 0) {
        const matchesAnyRange = tempSelectedPriceRanges.some(rangeLabel => {
          const opt = priceOptions.find(o => o.label === rangeLabel)
          if (!opt) return false
          return displayPrice >= opt.min && displayPrice <= opt.max
        })
        if (!matchesAnyRange) return false
      } else {
        if (displayPrice < tempPriceRange[0] || displayPrice > tempPriceRange[1]) return false
      }

      // Specs Check
      for (const [specKey, values] of Object.entries(tempSelectedSpecs)) {
        if (!values || values.length === 0) continue

        // Special key for brand filtering
        if (specKey === '__brand__') {
          if (!values.includes(String(product.brand?.id))) return false
          continue
        }

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
    }).length
  }, [products, inStockOnly, tempSelectedPriceRanges, tempPriceRange, tempSelectedSpecs])

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
    <div className="bg-neutral-50 min-h-screen py-4 font-sans">
      <div className="mx-auto max-w-7xl px-4">
        {/* Breadcrumbs */}
        <nav className="mb-3 flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">
            Trang chủ
          </Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Tất cả sản phẩm</span>
        </nav>



        {/* Category Banner Slider */}
        {categoryBanners.length > 0 && (
          <div className="mb-4 relative rounded-2xl overflow-hidden shadow-md group h-[200px] sm:h-[300px] bg-neutral-200">
            <Link to={categoryBanners[currentSlide]?.link_url || '/browse'} className="w-full h-full block">
              <img
                src={categoryBanners[currentSlide]?.image_url}
                alt={categoryBanners[currentSlide]?.title}
                className="w-full h-full object-cover transition-transform duration-500 hover:scale-101"
              />
            </Link>
            
            {/* Banner Content Overlay */}
            <div className="absolute inset-0 bg-gradient-to-r from-black/60 via-black/20 to-transparent flex flex-col justify-center px-8 sm:px-12 text-white pointer-events-none">
              {categoryBanners[currentSlide]?.tag && (
                <span className="bg-amber-500 text-black text-[9px] font-extrabold uppercase px-2.5 py-1 rounded w-fit tracking-wider mb-2.5">
                  {categoryBanners[currentSlide].tag}
                </span>
              )}
              <h2 className="text-lg sm:text-2xl font-black max-w-md leading-tight drop-shadow-md">
                {categoryBanners[currentSlide]?.title}
              </h2>
              {categoryBanners[currentSlide]?.subtitle && (
                <p className="text-xs sm:text-sm text-neutral-200 font-bold mt-1 max-w-sm drop-shadow">
                  {categoryBanners[currentSlide].subtitle}
                </p>
              )}
            </div>

            {/* Slider Controls */}
            {categoryBanners.length > 1 && (
              <>
                <button
                  onClick={() => setCurrentSlide(prev => (prev - 1 + categoryBanners.length) % categoryBanners.length)}
                  className="absolute left-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/70 hover:bg-white text-black flex items-center justify-center shadow-sm opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto"
                >
                  ❮
                </button>
                <button
                  onClick={() => setCurrentSlide(prev => (prev + 1) % categoryBanners.length)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/70 hover:bg-white text-black flex items-center justify-center shadow-sm opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto"
                >
                  ❯
                </button>

                {/* Indicators */}
                <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-1.5 pointer-events-auto">
                  {categoryBanners.map((_, idx) => (
                    <button
                      key={idx}
                      onClick={() => setCurrentSlide(idx)}
                      className={`w-2 h-2 rounded-full transition-all ${currentSlide === idx ? 'bg-white w-4' : 'bg-white/40'}`}
                    />
                  ))}
                </div>
              </>
            )}
          </div>
        )}

        {/* Header Title with query results */}
        <div className="mb-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
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

        {/* Backdrop for filter dropdowns */}
        {openDropdown && (
          <div className="fixed inset-0 z-35 bg-transparent" onClick={() => setOpenDropdown(null)} />
        )}

        {/* Brand Logo Row — only shows brands present in current product listing */}
        {!catalogLoading && availableBrands.length > 1 && (
          <div className="mb-4">
            <div className="flex gap-2 overflow-x-auto pb-1.5 scrollbar-none">
              {/* "Tất cả" Brand Option */}
              <button
                onClick={() => updateParam('brand', null)}
                className={`flex-shrink-0 px-4 py-2 rounded-full border text-xs font-bold uppercase transition-all ${
                  !brandIdParam
                    ? 'border-brand-600 bg-brand-600 text-white shadow-sm'
                    : 'border-neutral-200 bg-white text-neutral-600 hover:border-neutral-300 hover:text-black'
                }`}
              >
                Tất cả
              </button>
              {availableBrands.map((b) => {
                const isActive = brandIdParam === String(b.id)
                const logoText = brandLogos[b.name.toLowerCase()] || b.name
                return (
                  <button
                    key={b.id}
                    onClick={() => updateParam('brand', isActive ? null : String(b.id))}
                    className={`flex-shrink-0 px-4 py-2 rounded-full border text-xs font-bold uppercase tracking-wider transition-all duration-200 ${
                      isActive
                        ? 'border-brand-600 bg-brand-600 text-white shadow-sm'
                        : 'border-neutral-200 bg-white text-neutral-600 hover:border-neutral-300 hover:text-neutral-900 hover:-translate-y-0.5'
                    }`}
                  >
                    {logoText}
                  </button>
                )
              })}
            </div>
          </div>
        )}

        {/* Horizontal Dropdown Filter Bar */}
        <div className="relative z-40 mb-4 flex flex-wrap items-center gap-2.5">
          {/* Funnel Icon "Tất cả bộ lọc" Button */}
          <button
            onClick={openFilterModal}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-lg border text-xs font-bold transition-all shadow-sm ${
              Object.keys(selectedSpecs).length > 0 || selectedPriceRanges.length > 0 || priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max
                ? 'border-brand-600 bg-brand-50/30 text-brand-700 font-extrabold ring-1 ring-brand-100'
                : 'border-neutral-200 bg-white text-neutral-700 hover:border-neutral-350'
            }`}
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            <span>Bộ lọc</span>
            {(Object.keys(selectedSpecs).length > 0 || selectedPriceRanges.length > 0) && (
              <span className="bg-amber-500 text-white text-[9px] w-4 h-4 rounded-full flex items-center justify-center font-black">
                {Object.keys(selectedSpecs).length + (selectedPriceRanges.length > 0 ? 1 : 0)}
              </span>
            )}
          </button>

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
                    className="text-[10px] font-bold text-neutral-455 hover:text-neutral-700 uppercase"
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

          {/* Dynamic Spec Filter Dropdowns - Render KEY SPEC dropdowns only */}
          {filterSpecsList
            .filter((spec) => ['ram', 'bộ nhớ trong', 'dung lượng pin', 'kích thước màn hình', 'độ phân giải'].includes(spec.key.toLowerCase()))
            .map((spec) => {
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
                ? 'border-brand-600 bg-brand-50/30 text-brand-700 ring-1 ring-brand-100'
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
        {(categoryIdParam || brandIdParam || searchQuery || inStockOnly || Object.keys(selectedSpecs).length > 0 || selectedPriceRanges.length > 0 || priceRange[0] !== priceLimits.min || priceRange[1] !== priceLimits.max) && (
          <div className="mb-6 flex flex-wrap items-center gap-2 bg-neutral-100/50 border border-neutral-200/60 p-3 rounded-lg">
            <span className="text-[10px] font-black text-neutral-500 uppercase tracking-wider mr-1">Đang lọc:</span>
            
            {searchQuery && (
              <span className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                Từ khóa: {searchQuery}
                <button onClick={() => updateParam('q', null)} className="hover:text-red-500 text-xs font-black">✕</button>
              </span>
            )}
            {selectedCategory && (
              <span className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                Danh mục: {selectedCategory.name}
                <button onClick={() => updateParam('category', null)} className="hover:text-red-500 text-xs font-black">✕</button>
              </span>
            )}
            {selectedBrand && (
              <span className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                Hãng: {selectedBrand.name}
                <button onClick={() => updateParam('brand', null)} className="hover:text-red-500 text-xs font-black">✕</button>
              </span>
            )}
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
              values.map((v) => {
                if (key === '__brand__') {
                  const brand = brands.find(b => String(b.id) === v)
                  return (
                    <span key={`${key}-${v}`} className="inline-flex items-center gap-1 bg-white text-neutral-800 text-[10px] font-bold px-2.5 py-1 rounded-md border border-neutral-200 shadow-sm">
                      Hãng: {brand?.name || v}
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
                  )
                }
                return (
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
                )
              })
            )}

            <button
              onClick={clearAllFilters}
              className="text-[10px] font-black text-red-500 hover:text-red-650 transition-colors uppercase ml-auto"
            >
              Xóa tất cả bộ lọc
            </button>
          </div>
        )}

        {/* Full-width Product Grid */}
        <div className="w-full space-y-8">
          
          {/* Loading state or products display */}
          {loadingProducts ? (
            <SearchSkeleton viewMode={viewMode} count={10} />
          ) : (
            <>
              {viewMode === 'grid' ? (
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 md:gap-5">
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

        {/* Redesigned Clean Filter Modal */}
        {isFilterModalOpen && (
          <div
            className="fixed inset-0 z-[100] bg-black/50 backdrop-blur-[2px] flex items-end sm:items-center justify-center"
            onClick={(e) => { if (e.target === e.currentTarget) setIsFilterModalOpen(false) }}
          >
            <div className="bg-white w-full sm:w-[480px] sm:rounded-2xl rounded-t-2xl max-h-[90vh] flex flex-col shadow-2xl animate-fade-in-up">
              {/* Modal Header */}
              <div className="flex items-center justify-between px-5 py-4 border-b border-neutral-150">
                <div>
                  <h2 className="text-base font-black text-neutral-900">Bộ lọc</h2>
                  <p className="text-[11px] text-neutral-400 mt-0.5">{tempFilteredProductsCount} sản phẩm phù hợp</p>
                </div>
                <button
                  onClick={() => setIsFilterModalOpen(false)}
                  className="w-8 h-8 rounded-full bg-neutral-100 hover:bg-neutral-200 flex items-center justify-center font-bold text-neutral-600 transition-colors text-sm"
                >
                  ✕
                </button>
              </div>

              {/* Modal Body — scrollable */}
              <div className="flex-1 overflow-y-auto divide-y divide-neutral-100">

                {/* 1. Price Section */}
                <div className="px-5 py-4">
                  <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider mb-3">Khoảng giá</h3>
                  <div className="grid grid-cols-2 gap-2">
                    {priceOptions.map((opt) => {
                      const isSelected = tempSelectedPriceRanges.includes(opt.label)
                      return (
                        <button
                          key={opt.label}
                          type="button"
                          onClick={() => {
                            setTempSelectedPriceRanges(prev =>
                              prev.includes(opt.label)
                                ? prev.filter(l => l !== opt.label)
                                : [...prev, opt.label]
                            )
                          }}
                          className={`px-3 py-2.5 rounded-xl border text-xs font-bold text-center transition-all ${
                            isSelected
                              ? 'border-brand-600 bg-brand-600 text-white shadow-sm'
                              : 'border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-350 hover:bg-white'
                          }`}
                        >
                          {opt.label}
                        </button>
                      )
                    })}
                  </div>
                </div>

                {/* 2. Brand Section (from available brands only) */}
                {availableBrands.length > 1 && (
                  <div className="px-5 py-4">
                    <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider mb-3">Thương hiệu</h3>
                    <div className="flex flex-wrap gap-2">
                      {availableBrands.map((b) => {
                        const isSelected = tempSelectedSpecs['__brand__']?.includes(String(b.id)) || false
                        return (
                          <button
                            key={b.id}
                            type="button"
                            onClick={() => {
                              setTempSelectedSpecs(prev => {
                                const currentList = prev['__brand__'] || []
                                const nextList = currentList.includes(String(b.id))
                                  ? currentList.filter(v => v !== String(b.id))
                                  : [...currentList, String(b.id)]
                                const next = { ...prev, ['__brand__']: nextList }
                                if (nextList.length === 0) delete next['__brand__']
                                return next
                              })
                            }}
                            className={`px-3.5 py-2 rounded-xl border text-xs font-bold transition-all ${
                              isSelected
                                ? 'border-brand-600 bg-brand-600 text-white shadow-sm'
                                : 'border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-350 hover:bg-white'
                            }`}
                          >
                            {b.name}
                          </button>
                        )
                      })}
                    </div>
                  </div>
                )}

                {/* 3. Spec Sections — only key specs with short, meaningful values */}
                {filterSpecsList
                  .filter(spec => spec.options.every(o => o.length <= 30))
                  .slice(0, 5)
                  .map((spec) => (
                  <div key={spec.key} className="px-5 py-4">
                    <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider mb-3">{spec.key}</h3>
                    <div className="flex flex-wrap gap-2">
                      {spec.options.slice(0, 8).map((opt) => {
                        const isSelected = tempSelectedSpecs[spec.key]?.includes(opt) || false
                        return (
                          <button
                            key={opt}
                            type="button"
                            onClick={() => {
                              setTempSelectedSpecs(prev => {
                                const currentList = prev[spec.key] || []
                                const nextList = currentList.includes(opt)
                                  ? currentList.filter(v => v !== opt)
                                  : [...currentList, opt]
                                const next = { ...prev, [spec.key]: nextList }
                                if (nextList.length === 0) delete next[spec.key]
                                return next
                              })
                            }}
                            className={`px-3.5 py-2 rounded-xl border text-xs font-bold transition-all ${
                              isSelected
                                ? 'border-brand-600 bg-brand-600 text-white shadow-sm'
                                : 'border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-350 hover:bg-white'
                            }`}
                          >
                            {opt}
                          </button>
                        )
                      })}
                    </div>
                  </div>
                ))}

                {/* 4. Stock toggle */}
                <div className="px-5 py-4">
                  <button
                    type="button"
                    onClick={() => setInStockOnly(v => !v)}
                    className={`w-full flex items-center justify-between px-4 py-3 rounded-xl border text-xs font-bold transition-all ${
                      inStockOnly
                        ? 'border-brand-600 bg-brand-600 text-white'
                        : 'border-neutral-200 bg-neutral-50 text-neutral-700 hover:bg-white'
                    }`}
                  >
                    <span>Chỉ hiển thị còn hàng</span>
                    <span className={`w-10 h-6 rounded-full relative transition-colors ${
                      inStockOnly ? 'bg-white/30' : 'bg-neutral-300'
                    }`}>
                      <span className={`absolute top-1 w-4 h-4 rounded-full bg-white shadow transition-all ${
                        inStockOnly ? 'left-5' : 'left-1'
                      }`} />
                    </span>
                  </button>
                </div>

              </div>

              {/* Modal Footer */}
              <div className="px-5 py-4 border-t border-neutral-150 flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setTempSelectedSpecs({})
                    setTempSelectedPriceRanges([])
                    setTempPriceRange([priceLimits.min, priceLimits.max])
                    setInStockOnly(false)
                  }}
                  className="flex-1 py-3 rounded-xl border border-neutral-300 text-xs font-bold text-neutral-700 hover:bg-neutral-50 transition-colors"
                >
                  Xóa bộ lọc
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setSelectedSpecs(tempSelectedSpecs)
                    setSelectedPriceRanges(tempSelectedPriceRanges)
                    setPriceRange(tempPriceRange)
                    setIsFilterModalOpen(false)
                  }}
                  className="flex-2 flex-grow-[2] py-3 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-xs font-black uppercase tracking-wider transition-all shadow-sm active:scale-[0.98]"
                >
                  Xem {tempFilteredProductsCount} kết quả
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default BrowsePage
