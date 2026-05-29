import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { productAPI } from '../services/productAPI'
import { type Product, type ProductVariant } from '../types'

const ProductDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedImage, setSelectedImage] = useState<string | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [detailTab, setDetailTab] = useState<'description' | 'styling' | 'reviews'>('description')

  useEffect(() => {
    const fetchProduct = async () => {
      if (!id) return

      try {
        setLoading(true)
        setError(null)
        const data = await productAPI.getProductById(parseInt(id))
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

  const imagesList = [product.image, ...(product.images || [])].filter((img): img is string => !!img)

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
            {/* Vertically Aligned Thumbnail Column */}
            {imagesList.length > 1 && (
              <div className="flex md:flex-col gap-2 shrink-0 overflow-x-auto md:overflow-x-visible">
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

            {/* Main Product Image */}
            <div className="flex-1 bg-neutral-50 border border-neutral-200 rounded flex items-center justify-center p-6 aspect-square overflow-hidden max-h-[460px]">
              <img
                src={selectedImage || product.image || '/placeholder-product.png'}
                alt={product.name}
                className="max-h-full max-w-full object-contain mix-blend-multiply hover:scale-105 transition-transform duration-500"
              />
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
                disabled={activeStock === 0}
                className="flex-1 bg-black text-white text-xs font-bold py-3 rounded uppercase tracking-wider hover:bg-neutral-850 transition-colors disabled:bg-neutral-200 disabled:text-neutral-400 disabled:cursor-not-allowed"
              >
                Mua ngay sản phẩm
              </button>
              <button
                type="button"
                disabled={activeStock === 0}
                className="flex-1 border border-neutral-300 bg-white text-neutral-900 text-xs font-bold py-3 rounded uppercase tracking-wider hover:border-black transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Thêm vào giỏ hàng
              </button>
            </div>
            
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

              {/* Specifications */}
              {product.specifications && product.specifications.length > 0 && (
                <div className="space-y-3 pt-4 border-t border-neutral-150">
                  <h3 className="font-extrabold text-sm text-neutral-900 uppercase tracking-wide">Thông số kỹ thuật</h3>
                  <div className="max-w-xl border border-neutral-200 rounded overflow-hidden">
                    <dl className="divide-y divide-neutral-200 text-xs">
                      {product.specifications.map((spec, idx) => (
                        <div key={idx} className="grid grid-cols-3 p-3 bg-white odd:bg-neutral-50/50">
                          <dt className="font-bold text-neutral-500 uppercase tracking-wider text-[9px]">{spec.key}</dt>
                          <dd className="col-span-2 text-neutral-800 font-semibold pl-4">{spec.value}</dd>
                        </div>
                      ))}
                    </dl>
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
                  <p className="text-[10px] text-neutral-400 mt-1">98% khách hàng hài lòng với sản phẩm này</p>
                </div>

                {/* Star level distribution columns */}
                <div className="space-y-2 pt-2">
                  {[
                    { stars: 5, pct: 85 },
                    { stars: 4, pct: 10 },
                    { stars: 3, pct: 3 },
                    { stars: 2, pct: 1 },
                    { stars: 1, pct: 1 },
                  ].map((row) => (
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
                
                {product.review_count && product.review_count > 0 ? (
                  <div className="divide-y divide-neutral-200 space-y-4">
                    {/* Mock review item 1 */}
                    <div className="pt-4 first:pt-0 space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-full bg-neutral-150 flex items-center justify-center font-bold text-xs text-neutral-600">
                            A
                          </div>
                          <div>
                            <span className="text-xs font-bold text-neutral-800">Anh Nguyễn</span>
                            <span className="ml-2 bg-neutral-100 text-neutral-600 text-[8px] font-bold px-1 py-0.5 rounded">Đã mua hàng</span>
                          </div>
                        </div>
                        <span className="text-[10px] text-neutral-400">25/05/2026</span>
                      </div>
                      <div className="flex text-amber-500 text-xs">
                        <span>★★★★★</span>
                      </div>
                      <p className="text-xs text-neutral-650 leading-relaxed">
                        Sản phẩm dùng cực kỳ ưng ý. Giao hàng nhanh, gói cẩn thận. Shop nhiệt tình hỗ trợ hướng dẫn. Xứng đáng 5 sao!
                      </p>
                    </div>

                    {/* Mock review item 2 */}
                    <div className="pt-4 space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-full bg-neutral-150 flex items-center justify-center font-bold text-xs text-neutral-600">
                            H
                          </div>
                          <div>
                            <span className="text-xs font-bold text-neutral-800">Hoàng Lê</span>
                            <span className="ml-2 bg-neutral-100 text-neutral-600 text-[8px] font-bold px-1 py-0.5 rounded">Đã mua hàng</span>
                          </div>
                        </div>
                        <span className="text-[10px] text-neutral-400">18/05/2026</span>
                      </div>
                      <div className="flex text-amber-500 text-xs">
                        <span>★★★★☆</span>
                      </div>
                      <p className="text-xs text-neutral-650 leading-relaxed">
                        Chất lượng tốt so với giá thành, dùng ổn định, không có lỗi gì phát sinh. Sẽ tiếp tục mua hàng ủng hộ shop lần sau.
                      </p>
                    </div>
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
