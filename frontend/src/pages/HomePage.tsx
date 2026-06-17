import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import ProductCard from '../components/ProductCard'
import { useCatalog } from '../hooks/useCatalog'
import { useFeaturedProducts } from '../hooks/useProducts'
import { bannerAPI } from '../services/bannerAPI'
import { voucherAPI } from '../services/voucherAPI'
import type { Banner, Promotion, Product } from '../types'
import { keycloak } from '../utils/keycloak'

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
  const { brands, loading: catalogLoading } = useCatalog()
  const { products: featuredProducts, loading: productsLoading } = useFeaturedProducts(24)

  const [activeTab, setActiveTab] = useState<'for_you' | 'best_seller' | 'discount' | 'newest'>('for_you')
  const [currentSlide, setCurrentSlide] = useState(0)

  const [timeLeft, setTimeLeft] = useState({ hours: 0, minutes: 0, seconds: 0 })
  const [promotions, setPromotions] = useState<Promotion[]>([])
  const [countdownTarget, setCountdownTarget] = useState<Date | null>(null)
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

  useEffect(() => {
    const slideInterval = setInterval(() => {
      if (slides.length > 0) {
        setCurrentSlide(prev => (prev + 1) % slides.length)
      }
    }, 6000)
    return () => clearInterval(slideInterval)
  }, [slides])

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

  const activePromotions = promotions.filter(promo => {
    if (!promo.is_active || promo.is_deleted) return false
    const now = new Date()
    const startDate = new Date(promo.start_date)
    const endDate = new Date(promo.end_date)
    return now >= startDate && now <= endDate
  })

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
    <div className="bg-[#FAF9F5] min-h-screen pb-16 font-sans">
      
      {/* 1. Main Slider Block - Styled in print layout format */}
      {slides.length > 0 && (
        <div className="w-full bg-[#FAF9F5] py-6 border-b border-[#E4E4E7]">
          <div className="mx-auto max-w-7xl px-4">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              
              {/* Left Main Banner Slider */}
              <div className="lg:col-span-2 relative overflow-hidden rounded-none bg-white border border-[#E4E4E7] aspect-[21/9] lg:aspect-[16/7] flex items-center group">
                <Link to={slides[currentSlide]?.link_url || '/browse'} className="w-full h-full block">
                  <img
                    src={slides[currentSlide]?.image_url}
                    alt={slides[currentSlide]?.title}
                    className="w-full h-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.01]"
                  />
                </Link>
                
                {/* Elegant overlay text */}
                <div className="absolute inset-0 bg-gradient-to-r from-black/60 via-black/25 to-transparent flex flex-col justify-end p-8 text-white pointer-events-none">
                  {slides[currentSlide]?.tag && (
                    <span className="bg-[#FAF9F5] text-[#18181B] text-[8px] font-bold uppercase tracking-[0.2em] px-2 py-0.5 w-fit mb-3">
                      {slides[currentSlide].tag}
                    </span>
                  )}
                  <h1 className="font-serif-display text-2xl md:text-4xl font-light leading-tight max-w-lg">
                    {slides[currentSlide]?.title}
                  </h1>
                  {slides[currentSlide]?.subtitle && (
                    <p className="text-[10px] md:text-xs text-[#FAF9F5]/80 mt-2 tracking-widest uppercase font-semibold">
                      {slides[currentSlide].subtitle}
                    </p>
                  )}
                </div>

                {/* Left/Right Buttons */}
                <button
                  onClick={() => setCurrentSlide(prev => (prev - 1 + slides.length) % slides.length)}
                  className="absolute left-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-none border border-white/20 bg-black/40 hover:bg-black/60 flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                >
                  ‹
                </button>
                <button
                  onClick={() => setCurrentSlide(prev => (prev + 1) % slides.length)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-none border border-white/20 bg-black/40 hover:bg-black/60 flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                >
                  ›
                </button>

                {/* Indicators */}
                <div className="absolute bottom-4 left-8 flex gap-2 z-10">
                  {slides.map((_, idx) => (
                    <button
                      key={idx}
                      onClick={() => setCurrentSlide(idx)}
                      className={`h-1 transition-all rounded-none ${idx === currentSlide ? 'w-6 bg-[#FAF9F5]' : 'w-2 bg-[#FAF9F5]/40 hover:bg-white'}`}
                    />
                  ))}
                </div>
              </div>

              {/* Right Stacked Promos */}
              <div className="flex flex-col gap-4 h-full justify-between">
                {/* Promo 1 */}
                <div className="relative flex-1 rounded-none overflow-hidden border border-[#E4E4E7] bg-[#18181B] text-white p-5 flex flex-col justify-between group min-h-[140px]">
                  <div className="absolute inset-0 bg-cover bg-center opacity-20 mix-blend-overlay group-hover:scale-[1.01] transition-transform duration-500" style={{ backgroundImage: slides[1] ? `url(${slides[1].image_url})` : 'none' }}></div>
                  <div className="relative z-10 space-y-1">
                    <span className="text-[8px] font-bold uppercase tracking-[0.2em] text-[#8C8273]">Độc quyền</span>
                    <h3 className="font-serif-display text-lg mt-1 font-light">{slides[1]?.title || 'Sản Phẩm Độc Quyền'}</h3>
                    <p className="text-[10px] text-[#A1A1AA] font-light line-clamp-1">{slides[1]?.subtitle || 'Tuyển chọn các thiết kế giới hạn'}</p>
                  </div>
                  <Link to={slides[1]?.link_url || '/browse'} className="relative z-10 text-[9px] font-bold uppercase tracking-widest text-[#FAF9F5] hover:text-[#8C8273] transition-colors mt-4 flex items-center gap-1 w-fit">
                    Khám phá <span className="text-xs">→</span>
                  </Link>
                </div>

                {/* Promo 2 */}
                <div className="relative flex-1 rounded-none overflow-hidden border border-[#E4E4E7] bg-[#F6F5F0] text-[#18181B] p-5 flex flex-col justify-between group min-h-[140px]">
                  <div className="absolute inset-0 bg-cover bg-center opacity-10 mix-blend-overlay group-hover:scale-[1.01] transition-transform duration-500" style={{ backgroundImage: slides[2] ? `url(${slides[2].image_url})` : 'none' }}></div>
                  <div className="relative z-10 space-y-1">
                    <span className="text-[8px] font-bold uppercase tracking-[0.2em] text-[#8C8273]">Điểm hẹn</span>
                    <h3 className="font-serif-display text-lg mt-1 font-light">{slides[2]?.title || 'Tuyển chọn Giờ Vàng'}</h3>
                    <p className="text-[10px] text-[#71717A] font-light line-clamp-1">{slides[2]?.subtitle || 'Số lượng giới hạn dành cho hội viên'}</p>
                  </div>
                  <Link to={slides[2]?.link_url || '/browse'} className="relative z-10 text-[9px] font-bold uppercase tracking-widest text-[#18181B] hover:text-[#8C8273] transition-colors mt-4 flex items-center gap-1 w-fit">
                    Xem ngay <span className="text-xs">→</span>
                  </Link>
                </div>
              </div>

            </div>
          </div>
        </div>
      )}

      {/* 2. Main Grid Content */}
      <div className="mx-auto max-w-7xl px-4 py-6 space-y-8">

        {/* Brand Values Row */}
        <div className="py-2 border-b border-[#E4E4E7] pb-6">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {[
              { title: 'Miễn phí giao hàng', desc: 'Cho các đơn hàng từ 299K' },
              { title: '100% Chính hãng', desc: 'Cam kết chất lượng thiết kế' },
              { title: 'Bảo mật thanh toán', desc: 'Mã hóa dữ liệu giao dịch 24/7' },
              { title: '7 ngày đổi trả', desc: 'Đổi trả nhanh chóng, thủ tục tối giản' }
            ].map((val, idx) => (
              <div key={idx} className="flex flex-col p-4 bg-white border border-[#E4E4E7]">
                <h5 className="text-[9px] font-bold uppercase tracking-[0.2em] text-[#18181B]">{val.title}</h5>
                <p className="text-[10px] text-[#8C8273] mt-1 font-light">{val.desc}</p>
              </div>
            ))}
          </div>
        </div>

        {/* Flash Sale Section */}
        {flashSaleProducts.length > 0 && (
          <div className="border border-[#E4E4E7] rounded-none p-6 bg-white space-y-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border-b border-[#E4E4E7] pb-4">
              <div className="flex items-baseline gap-4">
                <span className="font-serif-display text-2xl font-medium tracking-tight text-[#18181B]">
                  BeliBeli // Giờ Vàng
                </span>
                
                {/* Clean Timer */}
                <div className="flex items-center gap-1 font-mono text-xs text-[#18181B] font-bold">
                  <span className="bg-[#18181B] text-[#FAF9F5] px-2 py-0.5 rounded-none">{String(timeLeft.hours).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-[#18181B] text-[#FAF9F5] px-2 py-0.5 rounded-none">{String(timeLeft.minutes).padStart(2, '0')}</span>
                  <span>:</span>
                  <span className="bg-[#18181B] text-[#FAF9F5] px-2 py-0.5 rounded-none">{String(timeLeft.seconds).padStart(2, '0')}</span>
                </div>
              </div>
              <span className="text-[9px] uppercase tracking-[0.25em] text-[#8C8273] font-bold">
                KẾT THÚC LÚC {countdownTarget ? countdownTarget.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }) : '24:00'}
              </span>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {flashSaleProducts.map(product => (
                <div key={product.id} className="relative flex flex-col justify-between bg-white border border-[#E4E4E7] p-1 group">
                  <ProductCard product={product} />
                  
                  {/* Slim, clean progress bar */}
                  <div className="mt-3 px-3 pb-3">
                    {(() => {
                      const totalStock = product.stock + Math.max(5, (product.stock % 7) + 3)
                      const soldCount = totalStock - product.stock
                      const soldPercentage = Math.round((soldCount / totalStock) * 100)
                      return (
                        <div className="space-y-1">
                          <div className="flex justify-between items-center text-[8px] font-bold uppercase tracking-wider text-[#8C8273]">
                            <span>Đã bán {soldCount}</span>
                            <span>Còn {product.stock}</span>
                          </div>
                          <div className="w-full bg-[#FAF9F5] border border-[#E4E4E7] h-1.5 relative overflow-hidden">
                            <div 
                              className="bg-[#18181B] h-full transition-all duration-700 ease-out" 
                              style={{ width: `${soldPercentage}%` }}
                            >
                            </div>
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

        {/* Luxe Editorial Section: Apple Zone */}
        {featuredProducts && featuredProducts.some(p => p.brand?.name?.toLowerCase().includes('apple') || p.name?.toLowerCase().includes('iphone') || p.name?.toLowerCase().includes('ipad')) && (
          <div className="relative rounded-none bg-[#18181B] text-[#FAF9F5] p-8 border border-[#18181B] space-y-6">
            <div className="flex items-center justify-between flex-wrap gap-4 border-b border-[#FAF9F5]/10 pb-4">
              <span className="font-serif-display text-2xl font-light tracking-wide text-white">
                 Apple Reseller // <span className="italic">Góc trưng bày thiết kế</span>
              </span>
              <Link to="/browse?brand=1" className="text-[9px] font-bold uppercase tracking-[0.25em] text-[#8C8273] hover:text-[#FAF9F5] transition-colors">
                Xem tất cả sản phẩm Apple →
              </Link>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {featuredProducts
                .filter(p => p.brand?.name?.toLowerCase().includes('apple') || p.name?.toLowerCase().includes('iphone') || p.name?.toLowerCase().includes('ipad') || p.name?.toLowerCase().includes('macbook'))
                .slice(0, 5)
                .map(product => (
                  <div key={product.id} className="bg-white/5 border border-white/10 p-1 hover:border-white/20 transition-all duration-300">
                    <ProductCard product={product} />
                  </div>
                ))
              }
            </div>
          </div>
        )}

        {/* Minimal Slogan Block */}
        <div className="w-full bg-[#FAF9F5] text-[#18181B] py-10 px-8 border border-[#E4E4E7] flex flex-col md:flex-row items-center justify-between gap-6 relative overflow-hidden">
          <div className="space-y-2 max-w-xl">
            <h3 className="font-serif-display text-2xl font-light leading-snug">
              "Ngôn ngữ của sự tĩnh lặng và nét thanh lịch thuần khiết."
            </h3>
            <p className="text-[11px] text-[#8C8273] font-light leading-relaxed">
              Tất cả các sản phẩm trưng bày tại BeliBeli đều trải qua quy trình đánh giá nghiêm ngặt để đảm bảo sự hòa hợp tuyệt đối giữa thiết kế mỹ thuật và hiệu năng vận hành.
            </p>
          </div>
          <Link to="/browse" className="bg-[#18181B] text-[#FAF9F5] text-[10px] font-semibold uppercase tracking-[0.2em] px-8 py-3.5 hover:bg-transparent hover:text-[#18181B] border border-[#18181B] transition-colors shrink-0">
            Xem Bộ Sưu Tập
          </Link>
        </div>

        {/* Tabs Grid */}
        <div className="space-y-6">
          <div className="border-b border-[#E4E4E7]">
            <div className="flex gap-6 overflow-x-auto pb-px scrollbar-none">
              {[
                { id: 'for_you', label: 'Gợi ý tuyển chọn' },
                { id: 'best_seller', label: 'Bán chạy nhất' },
                { id: 'discount', label: 'Ưu đãi thành viên' },
                { id: 'newest', label: 'Hàng mới về' },
              ].map(tab => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`pb-3 text-[10px] uppercase tracking-[0.2em] font-bold border-b transition-all shrink-0 cursor-pointer ${activeTab === tab.id ? 'border-[#18181B] text-[#18181B]' : 'border-transparent text-[#8C8273] hover:text-[#18181B]'}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {productsLoading ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              {[...Array(10)].map((_, i) => (
                <div key={i} className="aspect-[3/4] bg-white border border-[#E4E4E7] animate-pulse"></div>
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

        {/* Gian hàng nổi bật (Featured Store Cards) */}
        {!catalogLoading && brands.length > 0 && (
          <div className="space-y-4">
            <h2 className="font-serif-display text-2xl font-light text-[#18181B] tracking-tight">Gian Hàng Nổi Bật</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {brands.slice(0, 3).map(brand => {
                const brandProducts = featuredProducts 
                  ? featuredProducts.filter(p => p.brand?.id === brand.id || p.brand_id === brand.id)
                  : []
                
                return (
                  <div key={brand.id} className="border border-[#E4E4E7] bg-white p-6 transition-all duration-300 flex flex-col justify-between">
                    <div>
                      <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 bg-[#FAF9F5] border border-[#E4E4E7] text-[#18181B] flex items-center justify-center font-bold text-xs uppercase">
                            {brandLogos[brand.name.toLowerCase()]?.slice(0, 2) || brand.name.slice(0, 2)}
                          </div>
                          <div>
                            <div className="flex items-center gap-1">
                              <span className="font-serif-display font-medium text-sm text-[#18181B] uppercase tracking-wider">{brand.name}</span>
                              <span className="text-[#8C8273] text-xs">✓</span>
                            </div>
                            <p className="text-[8px] text-[#A1A1AA] uppercase tracking-[0.2em] font-semibold">Chính hãng</p>
                          </div>
                        </div>
                        <Link to={`/browse?brand=${brand.id}`} className="border border-[#E4E4E7] text-[#18181B] text-[9px] font-bold px-3 py-1.5 hover:bg-[#18181B] hover:text-[#FAF9F5] transition-all uppercase tracking-wider">
                          Khám phá
                        </Link>
                      </div>
                      
                      <div className="grid grid-cols-3 gap-2">
                        {brandProducts.slice(0, 3).map((p) => (
                          <Link key={p.id} to={`/products/${p.id}`} className="aspect-square bg-[#FAF9F5] border border-[#E4E4E7] p-1 flex items-center justify-center hover:border-[#18181B] transition-colors">
                            <img src={p.image || '/placeholder-product.png'} alt={p.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
                          </Link>
                        ))}
                        {brandProducts.length === 0 && (
                          [...Array(3)].map((_, idx) => {
                            const idxProduct = featuredProducts ? featuredProducts[idx * 2 + (brand.id % 3)] : null
                            if (!idxProduct) return <div key={idx} className="aspect-square bg-[#FAF9F5] border border-[#E4E4E7]" />
                            return (
                              <Link key={idx} to={`/products/${idxProduct.id}`} className="aspect-square bg-[#FAF9F5] border border-[#E4E4E7] p-1 flex items-center justify-center hover:border-[#18181B] transition-colors">
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

        {/* Bottom Editorial Columns */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4">
          <div className="bg-[#18181B] text-white p-8 border border-[#18181B] flex flex-col justify-between min-h-[180px]">
            <div>
              <span className="text-[8px] font-bold tracking-[0.2em] bg-white/10 px-2 py-1 uppercase text-[#FAF9F5]">Hội viên</span>
              <h4 className="font-serif-display text-lg font-light mt-4">Thành Viên Premium</h4>
              <p className="text-[11px] text-[#A1A1AA] mt-2 font-light leading-relaxed">Đăng ký thành viên để nhận ngay chính sách hoàn tiền 5% trọn đời và giao hàng ưu tiên.</p>
            </div>
            <button
              onClick={() => void keycloak.register()}
              className="text-[9px] font-bold uppercase text-[#FAF9F5] tracking-[0.2em] hover:text-[#8C8273] transition-colors mt-6 w-fit bg-transparent border-0 p-0 cursor-pointer text-left font-sans"
            >
              Đăng ký ngay →
            </button>
          </div>

          <div className="bg-white border border-[#E4E4E7] p-8 flex flex-col justify-between min-h-[180px]">
            <div>
              <span className="text-[8px] font-bold tracking-[0.2em] bg-[#FAF9F5] text-[#8C8273] px-2 py-1 uppercase">Tài chính</span>
              <h4 className="font-serif-display text-lg font-light mt-4">BeliBeli Credit Card</h4>
              <p className="text-[11px] text-[#71717A] mt-2 font-light leading-relaxed">Nhận ưu đãi giảm giá thêm tới 200.000đ khi liên kết thanh toán với thẻ VISA / MasterCard của ngân hàng đối tác.</p>
            </div>
            <a href="#" className="text-[9px] font-bold uppercase text-[#18181B] hover:text-[#8C8273] transition-colors tracking-[0.2em] mt-6 w-fit">Tìm hiểu thêm →</a>
          </div>

          <div className="bg-white border border-[#E4E4E7] p-8 flex flex-col justify-between min-h-[180px]">
            <div>
              <span className="text-[8px] font-bold tracking-[0.2em] bg-[#FAF9F5] text-[#8C8273] px-2 py-1 uppercase">Dịch vụ</span>
              <h4 className="font-serif-display text-lg font-light mt-4">Giao Hàng 2 Giờ</h4>
              <p className="text-[11px] text-[#71717A] mt-2 font-light leading-relaxed">Dịch vụ chuyển phát hỏa tốc trong vòng 2 giờ nội thành Hồ Chí Minh & Hà Nội cho các đơn hàng tuyển chọn.</p>
            </div>
            <a href="#" className="text-[9px] font-bold uppercase text-[#18181B] hover:text-[#8C8273] transition-colors tracking-[0.2em] mt-6 w-fit">Xem khu vực hỗ trợ →</a>
          </div>
        </div>

      </div>
    </div>
  )
}

export default HomePage
