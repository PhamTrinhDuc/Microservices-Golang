import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import { useCatalog } from '../hooks/useCatalog'
import { useFeaturedProducts } from '../hooks/useProducts'
import { bannerAPI } from '../services/bannerAPI'
import { voucherAPI } from '../services/voucherAPI'
import type { Banner, Promotion, Product } from '../types'

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
      
      {/* TGDĐ Split Hero Section */}
      {slides.length > 0 && (
        <div className="w-full bg-neutral-100/60 py-6 border-b border-neutral-200/80">
          <div className="mx-auto max-w-7xl px-4">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
              {/* Left 2/3 for Main Banner Slider */}
              <div className="lg:col-span-2 relative overflow-hidden rounded-xl bg-white border border-neutral-200/80 shadow-sm aspect-[21/9] lg:aspect-[16/7] flex items-center group">
                <Link to={slides[currentSlide]?.link_url || '/browse'} className="w-full h-full block">
                  <img
                    src={slides[currentSlide]?.image_url}
                    alt={slides[currentSlide]?.title}
                    className="w-full h-full object-cover transition-transform duration-700 ease-out group-hover:scale-101"
                  />
                </Link>
                
                {/* Glassmorphic overlay */}
                <div className="absolute inset-0 bg-gradient-to-r from-black/70 via-black/30 to-transparent flex flex-col justify-end p-6 md:p-10 text-white pointer-events-none">
                  {slides[currentSlide]?.tag && (
                    <span className="bg-amber-400 text-neutral-950 text-[10px] font-black uppercase tracking-wider px-2 py-0.5 rounded-sm w-fit mb-3">
                      {slides[currentSlide].tag}
                    </span>
                  )}
                  <h1 className="text-xl md:text-3xl font-black leading-tight drop-shadow-sm max-w-lg">
                    {slides[currentSlide]?.title}
                  </h1>
                  {slides[currentSlide]?.subtitle && (
                    <p className="text-xs md:text-sm text-neutral-200 mt-1 font-semibold drop-shadow-sm">
                      {slides[currentSlide].subtitle}
                    </p>
                  )}
                </div>

                {/* Left/Right Buttons */}
                <button
                  onClick={() => setCurrentSlide(prev => (prev - 1 + slides.length) % slides.length)}
                  className="absolute left-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-black/45 hover:bg-black/60 flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  ‹
                </button>
                <button
                  onClick={() => setCurrentSlide(prev => (prev + 1) % slides.length)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-black/45 hover:bg-black/60 flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  ›
                </button>

                {/* Navigation Indicators */}
                <div className="absolute bottom-3 left-6 md:left-10 flex gap-1.5 z-10">
                  {slides.map((_, idx) => (
                    <button
                      key={idx}
                      onClick={() => setCurrentSlide(idx)}
                      className={`h-1.5 rounded-full transition-all ${idx === currentSlide ? 'w-5 bg-amber-400' : 'w-1.5 bg-white/50 hover:bg-white'}`}
                    />
                  ))}
                </div>
              </div>

              {/* Right 1/3 for Stacked Promos */}
              <div className="flex flex-col gap-3 h-full justify-between">
                {/* Promo 1 */}
                <div className="relative flex-1 rounded-xl overflow-hidden border border-neutral-200/80 shadow-sm bg-gradient-to-br from-indigo-900 via-indigo-950 to-neutral-950 text-white p-5 flex flex-col justify-between group min-h-[145px]">
                  <div className="absolute inset-0 bg-cover bg-center opacity-30 mix-blend-overlay group-hover:scale-102 transition-transform duration-500" style={{ backgroundImage: slides[1] ? `url(${slides[1].image_url})` : 'none' }}></div>
                  <div className="relative z-10 space-y-1">
                    <span className="text-[9px] font-black uppercase tracking-wider bg-white/10 px-2 py-0.5 rounded text-indigo-200">Độc quyền online</span>
                    <h3 className="text-sm font-black mt-2 leading-snug">{slides[1]?.title || 'Sản Phẩm Độc Quyền'}</h3>
                    <p className="text-[10px] text-indigo-200/85 line-clamp-1">{slides[1]?.subtitle || 'Nhập mã giảm cực sâu hôm nay'}</p>
                  </div>
                  <Link to={slides[1]?.link_url || '/browse'} className="relative z-10 text-[10px] font-extrabold text-amber-300 hover:text-amber-250 uppercase tracking-wider mt-4 flex items-center gap-1">
                    Mua ngay <span className="text-xs">→</span>
                  </Link>
                </div>

                {/* Promo 2 */}
                <div className="relative flex-1 rounded-xl overflow-hidden border border-neutral-200/80 shadow-sm bg-gradient-to-br from-amber-600 via-red-650 to-neutral-950 text-white p-5 flex flex-col justify-between group min-h-[145px]">
                  <div className="absolute inset-0 bg-cover bg-center opacity-30 mix-blend-overlay group-hover:scale-102 transition-transform duration-500" style={{ backgroundImage: slides[2] ? `url(${slides[2].image_url})` : 'none' }}></div>
                  <div className="relative z-10 space-y-1">
                    <span className="text-[9px] font-black uppercase tracking-wider bg-black/10 px-2 py-0.5 rounded text-amber-100">Hot deal</span>
                    <h3 className="text-sm font-black mt-2 leading-snug">{slides[2]?.title || 'Deal Sốc Giờ Vàng'}</h3>
                    <p className="text-[10px] text-amber-100/85 line-clamp-1">{slides[2]?.subtitle || 'Số lượng có hạn, săn ngay kẻo lỡ'}</p>
                  </div>
                  <Link to={slides[2]?.link_url || '/browse'} className="relative z-10 text-[10px] font-extrabold text-white hover:underline uppercase tracking-wider mt-4 flex items-center gap-1">
                    Khám phá <span className="text-xs">→</span>
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Main Grid Content */}
      <div className="mx-auto max-w-7xl px-4 py-6 space-y-10">

        {/* Brand Value & Trust Section (Moved to top below Hero slider) */}
        <div className="py-2 border-b border-neutral-200/60 pb-6">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 text-center">
            <div className="flex items-center gap-3 p-3.5 rounded-xl bg-white border border-neutral-200/60 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-brand-50 flex items-center justify-center text-brand-600 shrink-0">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7" />
                </svg>
              </div>
              <div className="text-left">
                <h5 className="text-[11px] font-black uppercase text-neutral-800">Miễn phí vận chuyển</h5>
                <p className="text-[9px] text-neutral-450 mt-0.5">Mọi hóa đơn toàn quốc từ 299K</p>
              </div>
            </div>

            <div className="flex items-center gap-3 p-3.5 rounded-xl bg-white border border-neutral-200/60 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-brand-50 flex items-center justify-center text-brand-600 shrink-0">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <div className="text-left">
                <h5 className="text-[11px] font-black uppercase text-neutral-800">100% Chính hãng</h5>
                <p className="text-[9px] text-neutral-450 mt-0.5">Bồi hoàn gấp đôi nếu phát hiện giả</p>
              </div>
            </div>

            <div className="flex items-center gap-3 p-3.5 rounded-xl bg-white border border-neutral-200/60 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-brand-50 flex items-center justify-center text-brand-600 shrink-0">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <div className="text-left">
                <h5 className="text-[11px] font-black uppercase text-neutral-800">Bảo mật thanh toán</h5>
                <p className="text-[9px] text-neutral-450 mt-0.5">Mã hóa thông tin 24/7 an toàn</p>
              </div>
            </div>

            <div className="flex items-center gap-3 p-3.5 rounded-xl bg-white border border-neutral-200/60 shadow-sm">
              <div className="w-10 h-10 rounded-full bg-brand-50 flex items-center justify-center text-brand-600 shrink-0">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M4 4v5h.582m15.356 2A8.001 8.001 0 1121.21 8H18.5" />
                </svg>
              </div>
              <div className="text-left">
                <h5 className="text-[11px] font-black uppercase text-neutral-800">7 ngày đổi trả</h5>
                <p className="text-[9px] text-neutral-450 mt-0.5">Chính sách đổi trả dễ dàng nhanh chóng</p>
              </div>
            </div>
          </div>
        </div>

        {/* Categories Showcase Grid (TGDĐ style box grid) */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-[13px] font-black text-neutral-900 tracking-wider uppercase">Danh Mục Phổ Biến</h2>
            <Link to="/browse" className="text-xs font-bold text-brand-600 hover:text-brand-885 transition-colors uppercase">Xem tất cả</Link>
          </div>
          
          <div className="grid grid-cols-3 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-8 gap-3">
            {catalogLoading ? (
              [...Array(8)].map((_, i) => (
                <div key={i} className="flex flex-col items-center gap-2 bg-white border border-neutral-200 rounded-xl p-4 shrink-0 animate-pulse h-24">
                  <div className="w-10 h-10 rounded-full bg-neutral-200"></div>
                  <div className="w-12 h-3 bg-neutral-200 rounded mt-1"></div>
                </div>
              ))
            ) : (
              categories.map((category) => (
                <Link
                  key={category.id}
                  to={`/browse?category=${category.id}`}
                  className="flex flex-col items-center gap-2 text-center p-3 bg-white border border-neutral-200/80 rounded-xl hover:border-brand-500 hover:shadow-premium-soft hover:-translate-y-0.5 transition-all duration-300 group"
                >
                  <div className="w-10 h-10 flex items-center justify-center group-hover:scale-108 transition-transform">
                    {category.icon ? (
                      <img
                        src={category.icon}
                        alt={category.name}
                        className="w-full h-full object-contain mix-blend-multiply"
                      />
                    ) : (
                      <span className="font-black text-sm text-neutral-400 group-hover:text-brand-500 transition-colors">{category.name[0]}</span>
                    )}
                  </div>
                  <span className="text-[10px] font-black text-neutral-700 group-hover:text-black transition-colors leading-tight line-clamp-1">
                    {category.name}
                  </span>
                </Link>
              ))
            )}
          </div>
        </div>

        {/* Flash Sale Countdown Section with Neon Highlight Bar */}
        {flashSaleProducts.length > 0 && (
          <div className="border border-red-200/80 rounded-xl p-5 bg-gradient-to-b from-white to-red-50/10 space-y-4 shadow-sm relative overflow-hidden">
            <div className="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-orange-500 via-red-500 to-rose-600"></div>
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <span className="bg-red-500 text-white text-xs font-black tracking-wider px-2.5 py-1 rounded-sm uppercase flex items-center gap-1 shadow-sm">
                  <span>⚡</span> FLASH SALE
                </span>
                
                {/* Timer */}
                <div className="flex items-center gap-1.5 font-bold text-sm text-neutral-850">
                  <span className="bg-neutral-900 text-white px-2.5 py-0.8 rounded text-xs font-mono shadow-sm">{String(timeLeft.hours).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-neutral-900 text-white px-2.5 py-0.8 rounded text-xs font-mono shadow-sm">{String(timeLeft.minutes).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-red-650 text-white px-2.5 py-0.8 rounded text-xs font-mono shadow-sm animate-pulse">{String(timeLeft.seconds).padStart(2, '0')}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] font-black text-neutral-555 flex items-center gap-1.5 bg-neutral-100 px-3 py-1.2 rounded-full border border-neutral-200">
                  <span className="w-1.5 h-1.5 rounded-full bg-red-500 animate-ping"></span>
                  KẾT THÚC LÚC {countdownTarget ? countdownTarget.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }) : '24:00'}
                </span>
              </div>
            </div>

            {/* Flash Sale Product Cards and Progress Bars */}
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 pt-1">
              {flashSaleProducts.map(product => (
                <div key={product.id} className="relative flex flex-col justify-between bg-white border border-neutral-150 rounded-lg p-1.5 hover:shadow-md transition-shadow">
                  <div className="flex-1">
                    <ProductCard product={product} />
                  </div>
                  
                  {/* Glowing dynamic progress bar */}
                  <div className="mt-3 px-1.5 pb-2">
                    {(() => {
                      const totalStock = product.stock + Math.max(5, (product.stock % 7) + 3)
                      const soldCount = totalStock - product.stock
                      const soldPercentage = Math.round((soldCount / totalStock) * 100)
                      return (
                        <div className="space-y-1.5">
                          <div className="flex justify-between items-center text-[9px] font-black uppercase tracking-wider">
                            <span className="flex items-center gap-0.5 text-red-600">
                              <span>🔥</span>
                              Đã bán {soldCount}
                            </span>
                            <span className="text-neutral-450">Còn {product.stock}</span>
                          </div>
                          <div className="w-full bg-neutral-150 rounded-full h-3.5 relative overflow-hidden shadow-inner border border-neutral-200">
                            <div 
                              className="bg-gradient-to-r from-orange-500 via-red-500 to-rose-600 h-full rounded-full transition-all duration-700 ease-out shadow-[0_0_8px_rgba(239,68,68,0.35)]" 
                              style={{ width: `${soldPercentage}%` }}
                            >
                            </div>
                            <span className="absolute inset-0 flex items-center justify-center text-[9px] font-black text-neutral-800 leading-none">
                              {soldPercentage >= 50 ? <span className="text-white">Hot - Bán {soldPercentage}%</span> : `Đã bán ${soldPercentage}%`}
                            </span>
                          </div>
                        </div>
                      )
                    })()}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Thematic Promotion Block: Apple Zone */}
        {featuredProducts && featuredProducts.some(p => p.brand?.name?.toLowerCase().includes('apple') || p.name?.toLowerCase().includes('iphone') || p.name?.toLowerCase().includes('ipad')) && (
          <div className="relative rounded-2xl overflow-hidden bg-gradient-to-br from-neutral-900 via-neutral-950 to-neutral-900 p-6 md:p-8 border border-neutral-800 shadow-xl space-y-6">
            {/* Background luxury highlights */}
            <div className="absolute top-0 right-0 w-80 h-80 bg-neutral-850/20 rounded-full blur-3xl pointer-events-none"></div>
            <div className="absolute -bottom-20 -left-20 w-80 h-80 bg-brand-500/5 rounded-full blur-3xl pointer-events-none"></div>

            <div className="relative z-10 flex items-center justify-between flex-wrap gap-4 border-b border-neutral-800 pb-4">
              <span className="text-white text-lg font-black flex items-center gap-1.5">
                 <span className="bg-gradient-to-r from-neutral-100 to-neutral-400 bg-clip-text text-transparent tracking-tight">Apple Authorized Reseller</span>
              </span>
              <Link to="/browse?brand=1" className="text-xs font-black uppercase tracking-wider text-amber-300 hover:text-amber-250 transition-colors">
                Xem tất cả sản phẩm Apple →
              </Link>
            </div>

            {/* Apple Products List */}
            <div className="relative z-10 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {featuredProducts
                .filter(p => p.brand?.name?.toLowerCase().includes('apple') || p.name?.toLowerCase().includes('iphone') || p.name?.toLowerCase().includes('ipad') || p.name?.toLowerCase().includes('macbook'))
                .slice(0, 5)
                .map(product => (
                  <div key={product.id} className="bg-neutral-950/40 rounded-xl border border-neutral-800/80 p-1 hover:border-neutral-700 transition-all duration-300">
                    <ProductCard product={product} />
                  </div>
                ))
              }
            </div>
          </div>
        )}

        {/* Slogan Banner Block */}
        <div className="w-full bg-neutral-900 text-white py-6 px-8 rounded-xl flex flex-col md:flex-row items-center justify-between gap-4 border border-neutral-800 shadow-md relative overflow-hidden">
          <div className="absolute -right-10 -top-10 w-40 h-40 bg-white/5 rounded-full blur-2xl"></div>
          <div className="relative z-10">
            <h3 className="text-base font-black tracking-wider uppercase text-amber-400">Let's Shop Beyond Boundaries</h3>
            <p className="text-[11px] text-neutral-450 mt-1 max-w-xl">Đặt mua thiết bị công nghệ chính hãng từ các tập đoàn hàng đầu thế giới với chính sách hậu mãi và chăm sóc khách hàng tốt nhất tại BeliBeli.</p>
          </div>
          <Link to="/browse" className="relative z-10 bg-white text-neutral-900 text-xs font-black uppercase px-6 py-3 rounded-lg hover:bg-neutral-150 transition-colors shadow-sm shrink-0">
            Mua ngay
          </Link>
        </div>

        {/* Today's For You Tabs & Grid */}
        <div className="space-y-6">
          <div className="border-b border-neutral-200">
            <div className="flex gap-6 overflow-x-auto pb-px scrollbar-none">
              {[
                { id: 'for_you', label: 'Gợi ý cho bạn' },
                { id: 'best_seller', label: 'Bán chạy nhất' },
                { id: 'discount', label: 'Khuyến mãi cực sâu' },
                { id: 'newest', label: 'Hàng mới về' },
              ].map(tab => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`pb-3 text-xs font-black border-b-2 transition-all shrink-0 uppercase tracking-wider ${activeTab === tab.id ? 'border-brand-600 text-brand-600' : 'border-transparent text-neutral-400 hover:text-neutral-600'}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {productsLoading ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {[...Array(10)].map((_, i) => (
                <div key={i} className="aspect-[3/4] bg-neutral-200 animate-pulse rounded-lg border border-neutral-200"></div>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {getFilteredProducts().slice(0, 10).map(product => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          )}
        </div>

        {/* Best Selling Store (Brand Panels) */}
        {!catalogLoading && brands.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-[13px] font-black text-neutral-900 uppercase tracking-wider">Gian Hàng Nổi Bật</h2>
              <Link to="/browse" className="text-xs font-bold text-brand-600 hover:text-brand-850 transition-colors uppercase">Xem tất cả</Link>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {brands.slice(0, 3).map(brand => {
                const brandProducts = featuredProducts 
                  ? featuredProducts.filter(p => p.brand?.id === brand.id || p.brand_id === brand.id)
                  : []
                
                return (
                  <div key={brand.id} className="border border-neutral-200/80 rounded-xl p-4.5 bg-white hover:shadow-premium transition-all duration-300 flex flex-col justify-between">
                    <div>
                      <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-2">
                          {brand.logo ? (
                            <img src={brand.logo} alt={brand.name} className="w-10 h-10 object-contain rounded-full border border-neutral-100 p-0.5" />
                          ) : (
                            <div className="w-10 h-10 rounded-full bg-neutral-950 text-white flex items-center justify-center font-black text-sm">{brandLogos[brand.name.toLowerCase()]?.slice(0, 2) || brand.name.slice(0, 2)}</div>
                          )}
                          <div>
                            <div className="flex items-center gap-1">
                              <span className="font-bold text-xs text-neutral-800 uppercase tracking-wider">{brand.name}</span>
                              <span className="text-blue-500 text-xs" title="Official Store">✓</span>
                            </div>
                            <p className="text-[9px] text-neutral-400 uppercase tracking-widest font-semibold">Chính hãng</p>
                          </div>
                        </div>
                        <Link to={`/browse?brand=${brand.id}`} className="border border-neutral-250 text-neutral-800 text-[10px] font-bold px-3 py-1 rounded hover:bg-black hover:text-white hover:border-black transition-all uppercase">
                          Ghé thăm
                        </Link>
                      </div>
                      
                      {/* Product Previews */}
                      <div className="grid grid-cols-3 gap-2">
                        {brandProducts.slice(0, 3).map((p) => (
                          <Link key={p.id} to={`/products/${p.id}`} className="aspect-square bg-neutral-50 rounded-lg border border-neutral-200 p-1 flex items-center justify-center hover:border-neutral-400 transition-colors">
                            <img src={p.image || '/placeholder-product.png'} alt={p.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
                          </Link>
                        ))}
                        {brandProducts.length === 0 && (
                          [...Array(3)].map((_, idx) => {
                            const idxProduct = featuredProducts ? featuredProducts[idx * 2 + (brand.id % 3)] : null
                            if (!idxProduct) return <div key={idx} className="aspect-square bg-neutral-50 rounded-lg border border-neutral-200" />
                            return (
                              <Link key={idx} to={`/products/${idxProduct.id}`} className="aspect-square bg-neutral-50 rounded-lg border border-neutral-200 p-1 flex items-center justify-center hover:border-neutral-400 transition-colors">
                                <img src={idxProduct.image || '/placeholder-product.png'} alt={idxProduct.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
                              </Link>
                            )
                          })
                        )}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Promotional & Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-2">
          <div className="bg-gradient-to-tr from-neutral-900 to-neutral-800 text-white rounded-xl p-6 flex flex-col justify-between min-h-[160px] border border-neutral-800 shadow-sm relative overflow-hidden">
            <div className="absolute -right-6 -bottom-6 w-24 h-24 bg-white/5 rounded-full blur-xl"></div>
            <div>
              <span className="text-[10px] font-black tracking-wider bg-white/10 px-2.5 py-1 rounded uppercase text-amber-300">Đặc quyền VIP</span>
              <h4 className="text-xs font-black mt-3 uppercase tracking-wider">Thành Viên Premium</h4>
              <p className="text-[10.5px] text-neutral-400 mt-1 leading-relaxed">Nhận ngay hoàn tiền 5% cho tất cả đơn hàng & miễn phí giao hàng trọn đời.</p>
            </div>
            <Link to="/register" className="text-[10px] font-black uppercase text-amber-300 tracking-wider hover:underline mt-4 w-fit">Đăng ký ngay →</Link>
          </div>

          <div className="bg-white border border-neutral-200 rounded-xl p-6 flex flex-col justify-between min-h-[160px] shadow-sm">
            <div>
              <span className="text-[10px] font-black tracking-wider bg-neutral-100 text-neutral-600 px-2.5 py-1 rounded uppercase">Ưu đãi thanh toán</span>
              <h4 className="text-xs font-black mt-3 text-neutral-900 uppercase tracking-wider">BeliBeli Credit Card</h4>
              <p className="text-[10.5px] text-neutral-500 mt-1 leading-relaxed">Giảm thêm tới 200.000đ khi thanh toán bằng thẻ VISA/MasterCard của các ngân hàng đối tác.</p>
            </div>
            <a href="#" className="text-[10px] font-black uppercase text-neutral-900 hover:text-brand-600 tracking-wider mt-4 w-fit">Tìm hiểu thêm →</a>
          </div>

          <div className="bg-white border border-neutral-200 rounded-xl p-6 flex flex-col justify-between min-h-[160px] shadow-sm">
            <div>
              <span className="text-[10px] font-black tracking-wider bg-neutral-100 text-neutral-600 px-2.5 py-1 rounded uppercase">Dịch vụ hỏa tốc</span>
              <h4 className="text-xs font-black mt-3 text-neutral-900 uppercase tracking-wider">Giao hàng 2 Giờ</h4>
              <p className="text-[10.5px] text-neutral-500 mt-1 leading-relaxed">Nhận hàng nhanh chóng trong 2h tại khu vực nội thành. Miễn phí cho đơn hàng từ 500.000đ.</p>
            </div>
            <a href="#" className="text-[10px] font-black uppercase text-neutral-900 hover:text-brand-600 tracking-wider mt-4 w-fit">Xem khu vực áp dụng →</a>
          </div>
        </div>

      </div>
    </div>
  )
}

export default HomePage
