import { useEffect, useState, useMemo } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { productAPI } from '../services/productAPI'
import { type Product, type ProductVariant } from '../types'
import { useCart } from '../hooks/useCart'

const ProductDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedImage, setSelectedImage] = useState<string | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [detailTab, setDetailTab] = useState<'description' | 'styling' | 'reviews'>('description')

  const navigate = useNavigate()
  const { addToCart } = useCart()
  const [adding, setAdding] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const groupedSpecs = useMemo(() => {
    if (!product?.specifications) return {}
    return product.specifications.reduce<Record<string, typeof product.specifications>>((acc, spec) => {
      const groupName = spec.group || 'Thông số khác'
      if (!acc[groupName]) {
        acc[groupName] = []
      }
      acc[groupName].push(spec)
      return acc
    }, {})
  }, [product?.specifications])

  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})

  // Initialize all groups as expanded by default
  useEffect(() => {
    if (groupedSpecs) {
      const initial: Record<string, boolean> = {}
      Object.keys(groupedSpecs).forEach(group => {
        initial[group] = true
      })
      setExpandedGroups(initial)
    }
  }, [groupedSpecs])

  const toggleGroup = (groupName: string) => {
    setExpandedGroups(prev => ({
      ...prev,
      [groupName]: !prev[groupName]
    }))
  }

  const ratingStats = useMemo(() => {
    const reviews = product?.reviews || []
    const total = reviews.length
    const distribution = { 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 }
    
    reviews.forEach(r => {
      const rVal = Math.round(r.rating)
      if (rVal >= 1 && rVal <= 5) {
        distribution[rVal as 1|2|3|4|5]++
      }
    })
    
    // Calculate satisfied percentage
    const satisfiedCount = distribution[5] + distribution[4]
    const satisfiedPct = total > 0 ? Math.round((satisfiedCount / total) * 100) : 100
    
    return {
      total,
      satisfiedPct,
      distribution: Object.entries(distribution).map(([stars, count]) => ({
        stars: Number(stars),
        pct: total > 0 ? Math.round((count / total) * 100) : 0
      })).reverse()
    }
  }, [product?.reviews])

  const imagesList = useMemo(() => {
    if (!product) return []
    const rawList = [product.image, ...(product.images || [])].filter((img): img is string => !!img)
    return Array.from(new Set(rawList))
  }, [product?.image, product?.images])

  // Dynamic SEO meta tags update based on product meta_title and meta_description
  useEffect(() => {
    if (product) {
      document.title = product.meta_title || `${product.name} | Jiyuu Store`
      
      let metaDesc = document.querySelector('meta[name="description"]')
      if (!metaDesc) {
        metaDesc = document.createElement('meta')
        metaDesc.setAttribute('name', 'description')
        document.head.appendChild(metaDesc)
      }
      metaDesc.setAttribute('content', product.meta_description || `${product.name} - Mua sắm sản phẩm chất lượng tốt nhất tại Jiyuu Store.`)
    }
    
    return () => {
      document.title = 'Jiyuu Store'
    }
  }, [product])

  useEffect(() => {
    const fetchProduct = async () => {
      if (!id) return

      try {
        setLoading(true)
        setError(null)
        const data = await productAPI.getProductById(id)
        setProduct(data)
        setSelectedImage(data.image || null)
        if (data.variants && data.variants.length > 0) {
          setSelectedVariant(data.variants[0])
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load product')
      } finally {
        setLoading(false)
      }
    }

    void fetchProduct()
  }, [id])

  if (loading) {
    return (
      <section className="flex-1 bg-neutral-50 py-12">
        <div className="mx-auto max-w-7xl px-4">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
            <div className="lg:col-span-7 grid grid-cols-12 gap-4">
              <div className="col-span-2 space-y-3">
                {[...Array(4)].map((_, i) => (
                  <div key={i} className="aspect-square bg-neutral-200 animate-pulse rounded"></div>
                ))}
              </div>
              <div className="col-span-10 aspect-square bg-neutral-200 animate-pulse rounded"></div>
            </div>
            <div className="lg:col-span-5 space-y-4">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="h-4 w-3/4 animate-pulse rounded bg-neutral-200"></div>
              ))}
            </div>
          </div>
        </div>
      </section>
    )
  }

  if (error || !product) {
    return (
      <section className="flex-1 bg-neutral-50 py-12">
        <div className="mx-auto max-w-7xl px-4">
          <div className="rounded-lg border border-red-200 bg-red-50/50 p-8 text-center max-w-xl mx-auto">
            <p className="text-red-650 text-sm font-bold">{error || 'Không tìm thấy sản phẩm'}</p>
            <Link to="/browse" className="mt-4 inline-block text-xs font-bold underline text-neutral-850">
              Quay lại cửa hàng
            </Link>
          </div>
        </div>
      </section>
    )
  }

  // Determine current active prices and stock based on variant selection
  const originalPrice = selectedVariant ? selectedVariant.price : (product.price || 0)
  const displayPrice = selectedVariant 
    ? (selectedVariant.discount_price || selectedVariant.price)
    : (product.discount_price || product.price || 0)
  
  const hasDiscount = displayPrice < originalPrice
  const discountPercent = hasDiscount 
    ? Math.round(((originalPrice - displayPrice) / originalPrice) * 15) // Fallback discount percent if needed, or exact
    : 0

  const activeStock = selectedVariant ? selectedVariant.stock : product.stock

  const currentImgIdx = imagesList.indexOf(selectedImage || product.image || '')

  const handleAddToCart = async () => {
    if (!selectedVariant) {
      alert('Vui lòng chọn một phiên bản sản phẩm')
      return
    }
    try {
      setAdding(true)
      setSuccessMessage(null)
      await addToCart(selectedVariant.id, quantity).unwrap()
      setSuccessMessage(`Đã thêm ${quantity} sản phẩm vào giỏ hàng thành công!`)
      setTimeout(() => setSuccessMessage(null), 4000)
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setAdding(false)
    }
  }

  const handleBuyNow = async () => {
    if (!selectedVariant) return
    try {
      setAdding(true)
      await addToCart(selectedVariant.id, quantity).unwrap()
      navigate('/cart')
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setAdding(false)
    }
  }

  const handlePrevImage = () => {
    if (imagesList.length <= 1) return
    const nextIdx = (currentImgIdx - 1 + imagesList.length) % imagesList.length
    setSelectedImage(imagesList[nextIdx])
  }

  const handleNextImage = () => {
    if (imagesList.length <= 1) return
    const nextIdx = (currentImgIdx + 1) % imagesList.length
    setSelectedImage(imagesList[nextIdx])
  }

  return (
    <section className="flex-1 bg-neutral-55 py-8 font-sans">
      <div className="mx-auto max-w-7xl px-4">
        {/* Breadcrumb navigation */}
        <nav className="mb-6 flex items-center gap-1.5 text-xs text-neutral-450 font-medium">
          <Link to="/" className="hover:text-black transition-colors">
            Trang chủ
          </Link>
          <span>/</span>
          {product.category && (
            <>
              <Link to={`/browse?category=${product.category.id}`} className="hover:text-black transition-colors">
                {product.category.name}
              </Link>
              <span>/</span>
            </>
          )}
          <span className="text-neutral-800 font-semibold max-w-[200px] truncate">{product.name}</span>
        </nav>

        {/* Product Details Box */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 bg-white border border-neutral-200 rounded-lg p-6">
          
          {/* LEFT: Image Gallery Layout */}
          <div className="lg:col-span-7 flex flex-col-reverse md:flex-row gap-4">
            {/* Vertically Aligned Thumbnail Column (Scrollable on desktop, horizontal on mobile) */}
            {imagesList.length > 1 && (
              <div className="flex md:flex-col gap-2 shrink-0 overflow-x-auto md:overflow-y-auto md:max-h-[460px] pr-1.5 scrollbar-thin scrollbar-thumb-neutral-200">
                {imagesList.map((img, idx) => (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => setSelectedImage(img)}
                    className={`w-14 h-14 rounded border flex items-center justify-center p-1 bg-neutral-50 overflow-hidden shrink-0 transition-all ${
                      selectedImage === img ? 'border-black ring-1 ring-black' : 'border-neutral-200 hover:border-neutral-400'
                    }`}
                  >
                    <img
                      src={img}
                      alt={`${product.name} thumbnail ${idx + 1}`}
                      className="max-h-full max-w-full object-contain mix-blend-multiply"
                    />
                  </button>
                ))}
              </div>
            )}

            {/* Main Product Image with Next/Prev Arrow Controls */}
            <div className="relative flex-1 bg-neutral-50 border border-neutral-200 rounded flex items-center justify-center p-6 aspect-square overflow-hidden max-h-[460px] group">
              {imagesList.length > 1 && (
                <button
                  type="button"
                  onClick={handlePrevImage}
                  className="absolute left-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/80 hover:bg-white text-black shadow-md border border-neutral-250 flex items-center justify-center transition-all opacity-0 group-hover:opacity-100 focus:opacity-100 z-10 font-bold text-lg select-none hover:scale-105 active:scale-95"
                  aria-label="Previous image"
                >
                  ‹
                </button>
              )}

              <img
                src={selectedImage || product.image || '/placeholder-product.png'}
                alt={product.name}
                className="max-h-full max-w-full object-contain mix-blend-multiply hover:scale-105 transition-transform duration-500"
              />

              {imagesList.length > 1 && (
                <button
                  type="button"
                  onClick={handleNextImage}
                  className="absolute right-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/80 hover:bg-white text-black shadow-md border border-neutral-250 flex items-center justify-center transition-all opacity-0 group-hover:opacity-100 focus:opacity-100 z-10 font-bold text-lg select-none hover:scale-105 active:scale-95"
                  aria-label="Next image"
                >
                  ›
                </button>
              )}
            </div>
          </div>

          {/* RIGHT: Product Information Panel */}
          <div className="lg:col-span-5 flex flex-col gap-5">
            <div>
              {product.brand && (
                <span className="text-[10px] font-bold uppercase tracking-wider text-neutral-400">
                  {product.brand.name}
                </span>
              )}
              <h1 className="text-xl md:text-2xl font-black text-neutral-900 leading-tight mt-1">
                {product.name}
              </h1>
            </div>

            {/* Rating Stars and Stats */}
            <div className="flex items-center gap-3 text-xs">
              <div className="flex items-center text-amber-500 font-bold">
                <span className="text-sm mr-0.5">★</span>
                <span className="text-neutral-800">{product.rating ? product.rating.toFixed(1) : '0.0'}</span>
              </div>
              <span className="text-neutral-200">|</span>
              <span className="text-neutral-550">{product.review_count || 0} Đánh giá</span>
              <span className="text-neutral-200">|</span>
              <span className="text-neutral-550">Đã bán {((product.review_count || 0) * 4) + 12}</span>
            </div>

            {/* Price Section */}
            <div className="bg-neutral-50 rounded p-4 border border-neutral-150 flex flex-col gap-1">
              <div className="flex items-baseline gap-3">
                <span className="text-2xl font-black text-neutral-900">
                  {displayPrice.toLocaleString('vi-VN')} đ
                </span>
                {hasDiscount && (
                  <span className="text-sm text-neutral-400 line-through">
                    {originalPrice.toLocaleString('vi-VN')} đ
                  </span>
                )}
              </div>
              {hasDiscount && (
                <div className="text-[10px] font-bold uppercase text-red-500">
                  Tiết kiệm {discountPercent || 15}% so với giá gốc
                </div>
              )}
            </div>

            {/* Product Options (Variants Selector) */}
            {product.variants && product.variants.length > 0 && (
              <div className="space-y-2.5">
                <span className="text-xs font-bold text-neutral-850 uppercase tracking-wide">Chọn phiên bản:</span>
                <div className="flex flex-wrap gap-2">
                  {product.variants.map((variant) => (
                    <button
                      key={variant.id}
                      type="button"
                      onClick={() => setSelectedVariant(variant)}
                      className={`px-3 py-2 rounded text-xs font-semibold border transition-all ${
                        selectedVariant?.id === variant.id
                          ? 'border-black bg-neutral-950 text-white'
                          : 'border-neutral-250 bg-white text-neutral-700 hover:border-neutral-400'
                      }`}
                    >
                      {variant.name || `Phiên bản #${variant.id}`}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Quantity Selector */}
            <div className="space-y-2">
              <span className="text-xs font-bold text-neutral-850 uppercase tracking-wide">Số lượng:</span>
              <div className="flex items-center gap-4">
                <div className="flex items-center border border-neutral-300 rounded bg-white">
                  <button
                    type="button"
                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                    disabled={activeStock === 0}
                    className="w-9 h-9 flex items-center justify-center font-bold text-neutral-600 hover:bg-neutral-100 disabled:opacity-30"
                  >
                    −
                  </button>
                  <input
                    type="number"
                    min={1}
                    max={activeStock || 1}
                    value={quantity}
                    onChange={(e) => setQuantity(Math.max(1, Math.min(activeStock || 99, parseInt(e.target.value) || 1)))}
                    className="w-10 text-center text-xs font-bold text-neutral-850 bg-transparent focus:outline-none"
                  />
                  <button
                    type="button"
                    onClick={() => setQuantity(Math.min(activeStock || 99, quantity + 1))}
                    disabled={activeStock === 0}
                    className="w-9 h-9 flex items-center justify-center font-bold text-neutral-600 hover:bg-neutral-100 disabled:opacity-30"
                  >
                    +
                  </button>
                </div>
                
                {/* Stock info */}
                <span className={`text-xs font-semibold ${activeStock > 0 ? 'text-neutral-500' : 'text-red-500'}`}>
                  {activeStock > 0 ? `Còn lại ${activeStock} sản phẩm` : 'Tạm hết hàng'}
                </span>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 pt-3">
              <button
                type="button"
                disabled={activeStock === 0 || adding}
                onClick={handleBuyNow}
                className="flex-1 bg-black text-white text-xs font-bold py-3 rounded uppercase tracking-wider hover:bg-neutral-850 transition-colors disabled:bg-neutral-200 disabled:text-neutral-400 disabled:cursor-not-allowed"
              >
                {adding ? 'Đang xử lý...' : 'Mua ngay sản phẩm'}
              </button>
              <button
                type="button"
                disabled={activeStock === 0 || adding}
                onClick={handleAddToCart}
                className="flex-1 border border-neutral-300 bg-white text-neutral-900 text-xs font-bold py-3 rounded uppercase tracking-wider hover:border-black transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {adding ? 'Đang thêm...' : 'Thêm vào giỏ hàng'}
              </button>
            </div>

            {successMessage && (
              <div className="bg-neutral-900 text-white text-[11px] font-semibold py-2.5 px-4 rounded flex items-center justify-between shadow-md transition-all duration-300">
                <span>{successMessage}</span>
                <Link to="/cart" className="underline font-bold hover:text-neutral-300 ml-3 shrink-0 uppercase tracking-wider text-[10px]">
                  Xem giỏ hàng →
                </Link>
              </div>
            )}
            
            <div className="text-[11px] text-neutral-450 leading-relaxed pt-2 border-t border-neutral-150 space-y-1">
              <div className="flex items-center gap-1.5">
                <span>🛡️</span>
                <span>Bảo hành chính hãng 12 tháng tại hệ thống BeliBeli Care.</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span>🔄</span>
                <span>Lỗi 1 đổi 1 trong vòng 7 ngày nếu do lỗi nhà sản xuất.</span>
              </div>
            </div>
          </div>
        </div>

        {/* BOTTOM: Tabs Section for Description / Styling / Reviews */}
        <div className="mt-12 space-y-6">
          <div className="border-b border-neutral-200">
            <div className="flex gap-8">
              {[
                { id: 'description', label: 'Thông tin chi tiết' },
                { id: 'styling', label: 'Hướng dẫn sử dụng' },
                { id: 'reviews', label: `Đánh giá (${product.review_count || 0})` },
              ].map(tab => (
                <button
                  key={tab.id}
                  onClick={() => setDetailTab(tab.id as any)}
                  className={`pb-3 text-xs font-bold border-b-2 uppercase tracking-wide transition-all ${
                    detailTab === tab.id ? 'border-black text-black' : 'border-transparent text-neutral-400 hover:text-neutral-600'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Description Tab Content */}
          {detailTab === 'description' && (
            <div className="bg-white border border-neutral-200 rounded-lg p-6 space-y-6">
              {product.description && (
                <div className="space-y-2">
                  <h3 className="font-extrabold text-sm text-neutral-900 uppercase tracking-wide">Mô tả sản phẩm</h3>
                  <p className="text-xs text-neutral-600 leading-relaxed whitespace-pre-line">{product.description}</p>
                </div>
              )}

              {/* Specifications Grouped Accordion */}
              {Object.keys(groupedSpecs).length > 0 && (
                <div className="space-y-3 pt-4 border-t border-neutral-150">
                  <h3 className="font-extrabold text-sm text-neutral-900 uppercase tracking-wide">Thông số kỹ thuật</h3>
                  <div className="max-w-xl space-y-2">
                    {Object.entries(groupedSpecs).map(([groupName, specs]) => {
                      const isExpanded = !!expandedGroups[groupName]
                      return (
                        <div key={groupName} className="border border-neutral-200 rounded overflow-hidden">
                          {/* Accordion Header */}
                          <button
                            type="button"
                            onClick={() => toggleGroup(groupName)}
                            className="w-full flex items-center justify-between p-3 bg-neutral-50 hover:bg-neutral-100 transition-colors text-left"
                          >
                            <span className="text-xs font-bold text-neutral-800 uppercase tracking-wider">
                              {groupName}
                            </span>
                            <span 
                              className="text-xs text-neutral-450 transition-transform duration-200 inline-block"
                              style={{ transform: isExpanded ? 'rotate(180deg)' : 'rotate(0deg)' }}
                            >
                              ▼
                            </span>
                          </button>
                          
                          {/* Accordion Content */}
                          {isExpanded && (
                            <dl className="divide-y divide-neutral-200 text-xs">
                              {specs.map((spec, idx) => (
                                <div key={idx} className="grid grid-cols-3 p-3 bg-white odd:bg-neutral-50/50">
                                  <dt className="font-bold text-neutral-500 uppercase tracking-wider text-[9px]">{spec.key}</dt>
                                  <dd className="col-span-2 text-neutral-800 font-semibold pl-4">{spec.value}</dd>
                                </div>
                              ))}
                            </dl>
                          )}
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Styling Ideas Tab Content */}
          {detailTab === 'styling' && (
            <div className="bg-white border border-neutral-200 rounded-lg p-6 space-y-4">
              <h3 className="font-extrabold text-sm text-neutral-900 uppercase tracking-wide">Hướng dẫn sử dụng & Bảo quản</h3>
              <p className="text-xs text-neutral-600 leading-relaxed">
                Để đảm bảo thiết bị hoạt động tốt và bền bỉ nhất, vui lòng lưu ý một số điểm sau đây:
              </p>
              <ul className="list-disc list-inside text-xs text-neutral-600 space-y-2">
                <li>Đọc kỹ tài liệu hướng dẫn kèm theo hộp sản phẩm trước khi khởi động.</li>
                <li>Không sạc thiết bị qua đêm liên tục để tránh chai pin và ảnh hưởng tuổi thọ.</li>
                <li>Bảo quản ở nơi khô ráo, tránh ánh nắng chiếu trực tiếp hoặc nhiệt độ cao đột ngột.</li>
                <li>Sử dụng đúng cáp, củ sạc chính hãng kèm theo hoặc các phụ kiện đạt chuẩn MFi/QC/PD tương đương.</li>
              </ul>
            </div>
          )}

          {/* Reviews Tab Content */}
          {detailTab === 'reviews' && (
            <div className="bg-white border border-neutral-200 rounded-lg p-6 grid grid-cols-1 lg:grid-cols-12 gap-8">
              
              {/* Ratings breakdown side */}
              <div className="lg:col-span-4 space-y-4 border-b lg:border-b-0 lg:border-r border-neutral-200 pb-6 lg:pb-0 lg:pr-6">
                <div>
                  <h4 className="text-xs font-bold text-neutral-800 uppercase tracking-wider">Đánh giá chung</h4>
                  <div className="flex items-baseline gap-2 mt-2">
                    <span className="text-3xl font-black text-neutral-900">{product.rating ? product.rating.toFixed(1) : '0.0'}</span>
                    <span className="text-xs text-neutral-450">trên 5</span>
                  </div>
                  
                  <div className="flex text-amber-500 text-sm mt-1.5">
                    {[...Array(5)].map((_, i) => (
                      <span key={i}>{i < Math.floor(product.rating || 0) ? '★' : '☆'}</span>
                    ))}
                  </div>
                  <p className="text-[10px] text-neutral-400 mt-1">{ratingStats.satisfiedPct}% khách hàng hài lòng với sản phẩm này</p>
                </div>

                {/* Star level distribution columns */}
                <div className="space-y-2 pt-2">
                  {ratingStats.distribution.map((row) => (
                    <div key={row.stars} className="flex items-center gap-3 text-[11px] text-neutral-500">
                      <span className="w-3 text-right">{row.stars}★</span>
                      <div className="flex-1 bg-neutral-100 h-1.5 rounded-full overflow-hidden">
                        <div className="bg-neutral-900 h-full rounded-full" style={{ width: `${row.pct}%` }}></div>
                      </div>
                      <span className="w-7 text-right">{row.pct}%</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Review Lists side */}
              <div className="lg:col-span-8 space-y-5">
                <h4 className="text-xs font-bold text-neutral-800 uppercase tracking-wider">Đánh giá chi tiết</h4>
                
                {product.reviews && product.reviews.length > 0 ? (
                  <div className="divide-y divide-neutral-200 space-y-4">
                    {product.reviews.map((review) => {
                      const initial = review.user_full_name ? review.user_full_name.charAt(0).toUpperCase() : 'U'
                      const dateStr = new Date(review.created_at).toLocaleDateString('vi-VN')
                      return (
                        <div key={review.id} className="pt-4 first:pt-0 space-y-2">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <div className="w-8 h-8 rounded-full bg-neutral-150 flex items-center justify-center font-bold text-xs text-neutral-600">
                                {initial}
                              </div>
                              <div>
                                <span className="text-xs font-bold text-neutral-800">{review.user_full_name}</span>
                                <span className="ml-2 bg-neutral-100 text-neutral-600 text-[8px] font-bold px-1 py-0.5 rounded">Đã mua hàng</span>
                              </div>
                            </div>
                            <span className="text-[10px] text-neutral-400">{dateStr}</span>
                          </div>
                          <div className="flex text-amber-500 text-xs">
                            <span>{'★'.repeat(review.rating) + '☆'.repeat(5 - review.rating)}</span>
                          </div>
                          {review.comment && (
                            <p className="text-xs text-neutral-650 leading-relaxed">
                              {review.comment}
                            </p>
                          )}
                        </div>
                      )
                    })}
                  </div>
                ) : (
                  <div className="text-center py-6 text-neutral-450 text-xs border border-dashed border-neutral-250 rounded">
                    Chưa có đánh giá nào cho sản phẩm này. Hãy mua hàng và để lại đánh giá đầu tiên nhé!
                  </div>
                )}
              </div>

            </div>
          )}
        </div>

      </div>
    </section>
  )
}

export default ProductDetailPage
