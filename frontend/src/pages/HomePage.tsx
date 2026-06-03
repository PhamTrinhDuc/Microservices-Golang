import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import { useCatalog } from '../hooks/useCatalog'
import { useFeaturedProducts } from '../hooks/useProducts'
import { bannerAPI } from '../services/bannerAPI'
import { voucherAPI } from '../services/voucherAPI'
import type { Banner, Promotion, Product } from '../types'

const HomePage = () => {
  const { categories, brands, loading: catalogLoading } = useCatalog()
  const { products: featuredProducts, loading: productsLoading } = useFeaturedProducts(24)

  const [activeTab, setActiveTab] = useState<'for_you' | 'best_seller' | 'discount' | 'newest'>('for_you')
  const [currentSlide, setCurrentSlide] = useState(0)

  // Flash Sale Countdown timer & promotions state
  const [timeLeft, setTimeLeft] = useState({ hours: 0, minutes: 0, seconds: 0 })
  const [promotions, setPromotions] = useState<Promotion[]>([])
  const [countdownTarget, setCountdownTarget] = useState<Date | null>(null)

  // Dynamic slides state
  const [slides, setSlides] = useState<Banner[]>([])

  const loadBanners = async () => {
    try {
      const data = await bannerAPI.listBanners()
      if (data.length > 0) {
        setSlides(data)
      }
    } catch (err) {
      console.error('Failed to load homepage banners, using fallback:', err)
    }
  }

  const loadPromotions = async () => {
    try {
      const data = await voucherAPI.getPromotions()
      setPromotions(data)
      
      // Find the closest active promotion that ends in the future
      const now = new Date()
      const activeFuturePromotions = data.filter(p => {
        if (!p.is_active || p.is_deleted) return false
        const startDate = new Date(p.start_date)
        const endDate = new Date(p.end_date)
        return now >= startDate && endDate > now
      })
      
      if (activeFuturePromotions.length > 0) {
        activeFuturePromotions.sort((a, b) => new Date(a.end_date).getTime() - new Date(b.end_date).getTime())
        setCountdownTarget(new Date(activeFuturePromotions[0].end_date))
      } else {
        setCountdownTarget(null)
      }
    } catch (err) {
      console.error('Failed to load promotions:', err)
      setCountdownTarget(null)
    }
  }

  useEffect(() => {
    void loadBanners()
    void loadPromotions()
  }, [])

  useEffect(() => {
    if (!countdownTarget) return

    const updateTimer = () => {
      const now = new Date().getTime()
      const distance = countdownTarget.getTime() - now

      if (distance <= 0) {
        setTimeLeft({ hours: 0, minutes: 0, seconds: 0 })
        return
      }

      const hours = Math.floor(distance / (1000 * 60 * 60))
      const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60))
      const seconds = Math.floor((distance % (1000 * 60)) / 1000)

      setTimeLeft({ hours, minutes, seconds })
    }

    updateTimer()
    const timer = setInterval(updateTimer, 1000)
    return () => clearInterval(timer)
  }, [countdownTarget])

  // Auto rotate slides
  useEffect(() => {
    const slideInterval = setInterval(() => {
      if (slides.length > 0) {
        setCurrentSlide(prev => (prev + 1) % slides.length)
      }
    }, 6000)
    return () => clearInterval(slideInterval)
  }, [slides])

  // Filter products for tabs
  const getFilteredProducts = () => {
    if (!featuredProducts) return []
    switch (activeTab) {
      case 'discount':
        return featuredProducts.filter(p => !!p.discount_price)
      case 'best_seller':
        return featuredProducts.filter(p => (p.rating || 0) >= 4.5)
      case 'newest':
        return [...featuredProducts].reverse().slice(0, 12)
      case 'for_you':
      default:
        return featuredProducts.slice(0, 12)
    }
  }

  // Find matching promotion for a product and override display discount price
  const getProductWithPromotion = (product: Product, promo: Promotion): Product => {
    let calculatedDiscountPrice = product.discount_price
    if (promo.discount_type === 'percentage') {
      calculatedDiscountPrice = product.price * (1 - promo.discount_value / 100)
    } else if (promo.discount_type === 'fixed') {
      calculatedDiscountPrice = product.price - promo.discount_value
    }
    return {
      ...product,
      discount_price: calculatedDiscountPrice,
    }
  }

  // Get active promotions
  const activePromotions = promotions.filter(promo => {
    if (!promo.is_active || promo.is_deleted) return false
    const now = new Date()
    const startDate = new Date(promo.start_date)
    const endDate = new Date(promo.end_date)
    return now >= startDate && now <= endDate
  })

  // Get discount products for Flash Sale
  const flashSaleProducts = featuredProducts 
    ? (activePromotions.length > 0
        ? activePromotions.map(promo => {
            const prod = featuredProducts.find(p => p.id === promo.product_id)
            if (!prod) return null
            return getProductWithPromotion(prod, promo)
          }).filter((p): p is Product => p !== null)
        : []
      ).slice(0, 5)
    : []

  return (
    <div className="bg-neutral-50 min-h-screen pb-20 font-sans">
      
      {/* Hero Slider Container */}
      {slides.length > 0 && (
        <div className="w-full bg-neutral-100 py-6 border-b border-neutral-200">
          <div className="mx-auto max-w-7xl px-4">
            <div className="relative overflow-hidden rounded-lg bg-white border border-neutral-200 shadow-sm min-h-[380px] flex flex-col md:flex-row items-center">
              
              {/* Slide Info */}
              <div className="flex-1 p-8 md:p-12 space-y-6">
                {slides[currentSlide]?.tag && (
                  <span className="inline-flex items-center gap-1.5 rounded-full bg-black px-3.5 py-1 text-[10px] font-extrabold uppercase tracking-wide text-white">
                    {slides[currentSlide].tag}
                  </span>
                )}
                <h1 className="text-3xl md:text-5xl font-black text-neutral-900 tracking-tight leading-none">
                  {slides[currentSlide]?.title} <br />
                  {slides[currentSlide]?.subtitle && (
                    <span className="text-neutral-500 text-2xl md:text-3xl font-semibold">
                      {slides[currentSlide].subtitle}
                    </span>
                  )}
                </h1>
                {slides[currentSlide]?.description && (
                  <p className="text-xs text-neutral-550 leading-relaxed max-w-md">
                    {slides[currentSlide].description}
                  </p>
                )}
                <div className="flex gap-3 pt-2">
                  <Link
                    to={slides[currentSlide]?.link_url || '/browse'}
                    className="bg-black text-white text-xs font-bold px-6 py-3 rounded hover:bg-neutral-800 transition-colors"
                  >
                    Mua ngay
                  </Link>
                  <Link
                    to={slides[currentSlide]?.link_url || '/browse'}
                    className="border border-neutral-250 text-neutral-800 text-xs font-bold px-6 py-3 rounded hover:bg-neutral-50 transition-colors"
                  >
                    Xem thêm
                  </Link>
                </div>
              </div>

              {/* Slide Visual Graphic */}
              <div className="flex-1 w-full md:w-auto h-[260px] md:h-[380px] p-6 flex justify-center items-center bg-neutral-50 md:bg-white border-t md:border-t-0 md:border-l border-neutral-200">
                <div className="w-full h-full relative overflow-hidden rounded">
                  <img
                    src={slides[currentSlide]?.image_url}
                    alt={slides[currentSlide]?.title}
                    className="w-full h-full object-cover transition-opacity duration-500"
                  />
                </div>
              </div>

              {/* Carousel Navigation Indicators */}
              <div className="absolute bottom-4 left-8 md:left-12 flex gap-1.5">
                {slides.map((_, idx) => (
                  <button
                    key={idx}
                    onClick={() => setCurrentSlide(idx)}
                    className={`h-1.5 rounded-full transition-all ${idx === currentSlide ? 'w-6 bg-black' : 'w-1.5 bg-neutral-300 hover:bg-neutral-400'}`}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Main Grid Content */}
      <div className="mx-auto max-w-7xl px-4 py-8 space-y-12">

        {/* Circular Categories Row */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-extrabold text-neutral-900 tracking-tight uppercase">Danh Mục Phổ Biến</h2>
            <Link to="/browse" className="text-xs font-bold text-neutral-550 hover:text-black transition-colors">Xem tất cả</Link>
          </div>
          
          <div className="flex items-center gap-6 overflow-x-auto pb-4 scrollbar-none">
            {catalogLoading ? (
              [...Array(8)].map((_, i) => (
                <div key={i} className="flex flex-col items-center gap-2 shrink-0">
                  <div className="w-16 h-16 rounded-full bg-neutral-200 animate-pulse"></div>
                  <div className="w-12 h-3 bg-neutral-200 animate-pulse rounded"></div>
                </div>
              ))
            ) : (
              categories.map((category) => (
                <Link
                  key={category.id}
                  to={`/browse?category=${category.id}`}
                  className="flex flex-col items-center gap-2 text-center group shrink-0"
                >
                  <div className="w-16 h-16 rounded-full bg-neutral-100/70 border border-neutral-200 flex items-center justify-center p-3.5 group-hover:border-black group-hover:bg-white transition-all shadow-sm">
                    {category.icon ? (
                      <img
                        src={category.icon}
                        alt={category.name}
                        className="w-full h-full object-contain mix-blend-multiply group-hover:scale-105 transition-transform"
                      />
                    ) : (
                      <span className="font-extrabold text-neutral-400 group-hover:text-black transition-colors">{category.name[0]}</span>
                    )}
                  </div>
                  <span className="text-[11px] font-bold text-neutral-600 group-hover:text-black transition-colors max-w-[80px] truncate">
                    {category.name}
                  </span>
                </Link>
              ))
            )}
          </div>
        </div>

        {/* Flash Sale Countdown Section */}
        {flashSaleProducts.length > 0 && (
          <div className="border border-neutral-200 rounded-lg p-5 bg-white space-y-4 shadow-sm">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <span className="bg-red-500 text-white text-xs font-black tracking-wider px-2.5 py-1 rounded-sm uppercase">
                  FLASH SALE
                </span>
                
                {/* Timer */}
                <div className="flex items-center gap-1.5 font-bold text-sm text-neutral-850">
                  <span className="bg-neutral-900 text-white px-2 py-0.5 rounded text-xs font-mono">{String(timeLeft.hours).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-neutral-900 text-white px-2 py-0.5 rounded text-xs font-mono">{String(timeLeft.minutes).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-red-500 text-white px-2 py-0.5 rounded text-xs font-mono">{String(timeLeft.seconds).padStart(2, '0')}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-neutral-550 flex items-center gap-1.5 bg-neutral-100 px-3 py-1 rounded-full">
                  <span className="w-2 h-2 rounded-full bg-red-500 animate-ping"></span>
                  Kết thúc lúc {countdownTarget ? countdownTarget.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }) : '24:00'} {countdownTarget ? `(${countdownTarget.toLocaleDateString('vi-VN')})` : 'hôm nay'}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {flashSaleProducts.map(product => (
                <div key={product.id} className="relative flex flex-col justify-between">
                  <div className="flex-1">
                    <ProductCard product={product} />
                  </div>
                  
                  {/* Stock Left Progress Bar */}
                  <div className="mt-3 px-1">
                    {(() => {
                      const totalStock = product.stock + Math.max(5, (product.stock % 7) + 3)
                      const soldCount = totalStock - product.stock
                      const soldPercentage = Math.round((soldCount / totalStock) * 100)
                      return (
                        <>
                          <div className="flex justify-between items-center text-[10px] text-neutral-600 font-bold mb-1 uppercase tracking-wider">
                            <span className="flex items-center gap-1">
                              <span className="text-red-500 animate-pulse">🔥</span>
                              Đã bán {soldCount}
                            </span>
                            <span className="text-neutral-400">Còn {product.stock}</span>
                          </div>
                          <div className="w-full bg-neutral-100 rounded-full h-3.5 relative overflow-hidden shadow-inner border border-neutral-200">
                            <div 
                              className="bg-gradient-to-r from-orange-500 via-red-500 to-rose-600 h-full rounded-full transition-all duration-500 shadow-[0_0_8px_rgba(239,68,68,0.3)]" 
                              style={{ width: `${soldPercentage}%` }}
                            >
                              <span className="absolute inset-0 flex items-center justify-center text-[9px] font-black text-white leading-none">
                                {soldPercentage}%
                              </span>
                            </div>
                          </div>
                        </>
                      )
                    })()}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Slogan Banner Block */}
        <div className="w-full bg-neutral-900 text-white py-6 px-8 rounded-lg flex flex-col md:flex-row items-center justify-between gap-4">
          <div>
            <h3 className="text-lg font-black tracking-tight">Let's Shop Beyond Boundaries</h3>
            <p className="text-[11px] text-neutral-400 mt-0.5">Đặt mua các sản phẩm từ các thương hiệu chính hãng toàn cầu một cách dễ dàng nhất.</p>
          </div>
          <Link to="/browse" className="bg-white text-black text-xs font-extrabold px-6 py-2.5 rounded hover:bg-neutral-200 transition-colors">
            Mua ngay
          </Link>
        </div>

        {/* Todays For You Tabs & Grid */}
        <div className="space-y-6">
          <div className="border-b border-neutral-200">
            <div className="flex gap-6 overflow-x-auto pb-px">
              {[
                { id: 'for_you', label: 'Gợi ý cho bạn' },
                { id: 'best_seller', label: 'Bán chạy nhất' },
                { id: 'discount', label: 'Khuyến mãi cực sâu' },
                { id: 'newest', label: 'Hàng mới về' },
              ].map(tab => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`pb-3 text-sm font-bold border-b-2 transition-all shrink-0 uppercase tracking-wide ${activeTab === tab.id ? 'border-black text-black' : 'border-transparent text-neutral-400 hover:text-neutral-600'}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {productsLoading ? (
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {[...Array(12)].map((_, i) => (
                <div key={i} className="aspect-[3/4] bg-neutral-200 animate-pulse rounded"></div>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {getFilteredProducts().map(product => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          )}
        </div>

        {/* Best Selling Store (Brand Panels) */}
        {!catalogLoading && brands.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-base font-extrabold text-neutral-900 uppercase">Gian Hàng Chính Hãng Nổi Bật</h2>
              <Link to="/browse" className="text-xs font-bold text-neutral-550 hover:text-black transition-colors">Xem tất cả</Link>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {brands.slice(0, 3).map(brand => {
                // Find products belonging to this brand
                const brandProducts = featuredProducts 
                  ? featuredProducts.filter(p => p.brand?.id === brand.id || p.brand_id === brand.id)
                  : []
                
                return (
                  <div key={brand.id} className="border border-neutral-200 rounded-lg p-4 bg-white hover:shadow-premium transition-all duration-300">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-2">
                        {brand.logo ? (
                          <img src={brand.logo} alt={brand.name} className="w-10 h-10 object-contain rounded-full border border-neutral-100 p-0.5" />
                        ) : (
                          <div className="w-10 h-10 rounded-full bg-neutral-950 text-white flex items-center justify-center font-extrabold text-sm">{brand.name[0]}</div>
                        )}
                        <div>
                          <div className="flex items-center gap-1">
                            <span className="font-bold text-sm text-neutral-850">{brand.name}</span>
                            <span className="text-blue-500 text-xs" title="Official Store">✓</span>
                          </div>
                          <p className="text-[10px] text-neutral-400">Official Brand Store</p>
                        </div>
                      </div>
                      <Link to={`/browse?brand=${brand.id}`} className="border border-neutral-250 text-neutral-800 text-[10px] font-bold px-3 py-1 rounded hover:bg-black hover:text-white hover:border-black transition-all">
                        Ghé thăm
                      </Link>
                    </div>
                    
                    {/* Product Previews */}
                    <div className="grid grid-cols-3 gap-2">
                      {brandProducts.slice(0, 3).map((p) => (
                        <Link key={p.id} to={`/products/${p.id}`} className="aspect-square bg-neutral-50 rounded border border-neutral-150 p-1 flex items-center justify-center hover:border-neutral-450 transition-colors">
                          <img src={p.image || '/placeholder-product.png'} alt={p.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
                        </Link>
                      ))}
                      {brandProducts.length === 0 && (
                        // Fallback elements
                        [...Array(3)].map((_, idx) => {
                          const idxProduct = featuredProducts ? featuredProducts[idx * 2 + (brand.id % 3)] : null
                          if (!idxProduct) return <div key={idx} className="aspect-square bg-neutral-50 rounded border border-neutral-100" />
                          return (
                            <Link key={idx} to={`/products/${idxProduct.id}`} className="aspect-square bg-neutral-50 rounded border border-neutral-150 p-1 flex items-center justify-center hover:border-neutral-450 transition-colors">
                              <img src={idxProduct.image || '/placeholder-product.png'} alt={idxProduct.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
                            </Link>
                          )
                        })
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Promotional & Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4">
          <div className="bg-gradient-to-tr from-neutral-900 to-neutral-800 text-white rounded-lg p-6 flex flex-col justify-between min-h-[160px] border border-neutral-800">
            <div>
              <span className="text-[10px] font-extrabold tracking-wider bg-white/10 px-2 py-0.5 rounded uppercase">Đặc quyền VIP</span>
              <h4 className="text-sm font-black mt-3">Thành Viên Premium</h4>
              <p className="text-[11px] text-neutral-450 mt-1 leading-relaxed">Nhận ngay hoàn tiền 5% cho tất cả đơn hàng & miễn phí giao hàng trọn đời.</p>
            </div>
            <Link to="/register" className="text-[11px] font-bold underline mt-4 hover:text-neutral-200 transition-colors w-fit">Đăng ký ngay →</Link>
          </div>

          <div className="bg-white border border-neutral-200 rounded-lg p-6 flex flex-col justify-between min-h-[160px]">
            <div>
              <span className="text-[10px] font-extrabold tracking-wider bg-neutral-100 text-neutral-600 px-2 py-0.5 rounded uppercase">Ưu đãi thanh toán</span>
              <h4 className="text-sm font-bold mt-3 text-neutral-900">BeliBeli Credit Card</h4>
              <p className="text-[11px] text-neutral-500 mt-1 leading-relaxed">Giảm thêm tới 200.000đ khi thanh toán bằng thẻ VISA/MasterCard của các ngân hàng đối tác.</p>
            </div>
            <a href="#" className="text-[11px] font-bold text-neutral-900 underline mt-4 hover:text-neutral-650 transition-colors w-fit">Tìm hiểu thêm →</a>
          </div>

          <div className="bg-white border border-neutral-200 rounded-lg p-6 flex flex-col justify-between min-h-[160px]">
            <div>
              <span className="text-[10px] font-extrabold tracking-wider bg-neutral-100 text-neutral-600 px-2 py-0.5 rounded uppercase">Dịch vụ hỏa tốc</span>
              <h4 className="text-sm font-bold mt-3 text-neutral-900">Giao hàng 2 Giờ</h4>
              <p className="text-[11px] text-neutral-500 mt-1 leading-relaxed">Nhận hàng nhanh chóng trong 2h tại khu vực nội thành. Miễn phí cho đơn hàng từ 500.000đ.</p>
            </div>
            <a href="#" className="text-[11px] font-bold text-neutral-900 underline mt-4 hover:text-neutral-650 transition-colors w-fit">Xem khu vực áp dụng →</a>
          </div>
        </div>

        {/* Brand Value & Trust Section */}
        <div className="border-t border-neutral-200 pt-8 mt-4">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-6 text-center">
            <div className="flex flex-col items-center p-4 rounded-lg bg-white border border-neutral-150 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center text-neutral-750 mb-3">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7" />
                </svg>
              </div>
              <h5 className="text-xs font-bold text-neutral-900">Miễn phí vận chuyển</h5>
              <p className="text-[10px] text-neutral-400 mt-1 max-w-[150px] mx-auto">Cho mọi hóa đơn từ 299kđ trên phạm vi toàn quốc.</p>
            </div>

            <div className="flex flex-col items-center p-4 rounded-lg bg-white border border-neutral-150 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center text-neutral-750 mb-3">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <h5 className="text-xs font-bold text-neutral-900">100% Chính hãng</h5>
              <p className="text-[10px] text-neutral-400 mt-1 max-w-[150px] mx-auto">Cam kết bồi thường gấp đôi nếu phát hiện hàng nhái.</p>
            </div>

            <div className="flex flex-col items-center p-4 rounded-lg bg-white border border-neutral-150 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center text-neutral-750 mb-3">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <h5 className="text-xs font-bold text-neutral-900">Bảo mật thanh toán</h5>
              <p className="text-[10px] text-neutral-400 mt-1 max-w-[150px] mx-auto">Chứng chỉ bảo mật SSL, bảo vệ thông tin 24/7.</p>
            </div>

            <div className="flex flex-col items-center p-4 rounded-lg bg-white border border-neutral-150 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center text-neutral-750 mb-3">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 1121.21 8H18.5" />
                </svg>
              </div>
              <h5 className="text-xs font-bold text-neutral-900">Đổi trả trong 7 ngày</h5>
              <p className="text-[10px] text-neutral-400 mt-1 max-w-[150px] mx-auto">Chính sách đổi hàng dễ dàng, hoàn tiền nhanh chóng.</p>
            </div>
          </div>
        </div>


      </div>
    </div>
  )
}

export default HomePage
