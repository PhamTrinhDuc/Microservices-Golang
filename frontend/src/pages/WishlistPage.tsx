import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useWishlist } from '../hooks/useWishlist'
import { useCart } from '../hooks/useCart'

const WishlistPage = () => {
  const { items, loading, error, removeFromWishlist } = useWishlist()
  const { addToCart } = useCart()
  
  const [loadingCartMap, setLoadingCartMap] = useState<Record<number, boolean>>({})
  const [loadingRemoveMap, setLoadingRemoveMap] = useState<Record<number, boolean>>({})

  const handleAddToCart = async (variantId: number) => {
    try {
      setLoadingCartMap(prev => ({ ...prev, [variantId]: true }))
      await addToCart(variantId, 1).unwrap()
      alert('Sản phẩm đã được thêm vào giỏ hàng!')
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setLoadingCartMap(prev => ({ ...prev, [variantId]: false }))
    }
  }

  const handleRemoveFromWishlist = async (variantId: number) => {
    try {
      setLoadingRemoveMap(prev => ({ ...prev, [variantId]: true }))
      await removeFromWishlist(variantId).unwrap()
    } catch (err: any) {
      alert(err || 'Không thể xóa sản phẩm khỏi danh sách yêu thích')
    } finally {
      setLoadingRemoveMap(prev => ({ ...prev, [variantId]: false }))
    }
  }

  if (loading && items.length === 0) {
    return (
      <div className="flex-1 bg-neutral-50 py-12 flex items-center justify-center min-h-[400px]">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto"></div>
          <p className="text-xs font-semibold text-neutral-500 uppercase tracking-wider">Đang tải danh sách yêu thích...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 bg-neutral-50 py-8 font-sans">
      <div className="mx-auto max-w-7xl px-4">
        {/* Breadcrumb */}
        <nav className="mb-6 flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">Trang chủ</Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Sản phẩm yêu thích</span>
        </nav>

        <h1 className="text-2xl font-black text-neutral-900 tracking-tight uppercase mb-8">
          Sản phẩm yêu thích ({items.length})
        </h1>

        {error && (
          <div className="mb-6 border border-red-200 bg-red-50 text-red-650 text-xs font-semibold px-4 py-3 rounded-lg">
            {error}
          </div>
        )}

        {items.length === 0 ? (
          /* EMPTY STATE */
          <div className="max-w-md mx-auto text-center border border-neutral-200 border-dashed rounded-lg p-12 bg-white shadow-sm space-y-5">
            <div className="w-16 h-16 rounded-full bg-red-50 flex items-center justify-center text-red-500 mx-auto">
              <svg className="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
                <path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </div>
            <div>
              <h3 className="text-sm font-bold text-neutral-800 uppercase tracking-wider">Danh sách yêu thích của bạn trống</h3>
              <p className="text-xs text-neutral-400 leading-relaxed mt-1 max-w-xs mx-auto">
                Hãy khám phá các sản phẩm và nhấn biểu tượng trái tim để lưu lại những sản phẩm yêu thích của bạn tại đây!
              </p>
            </div>
            <Link
              to="/browse"
              className="inline-block bg-black text-white text-xs font-extrabold uppercase px-8 py-3 rounded hover:bg-neutral-850 transition-colors"
            >
              Khám phá sản phẩm
            </Link>
          </div>
        ) : (
          /* GRID VIEW */
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {items.map((item) => {
              const displayPrice = item.sell_price || 0
              const originalPrice = item.compare_price || 0
              const hasDiscount = !!(originalPrice && originalPrice > displayPrice)
              const discountPercent = hasDiscount 
                ? Math.round(((originalPrice - displayPrice) / originalPrice) * 100)
                : 0

              return (
                <div key={item.id} className="relative flex flex-col h-full bg-white rounded-lg border border-neutral-200 overflow-hidden hover:shadow-premium hover:-translate-y-0.5 transition-all duration-300 group">
                  {/* Image Container */}
                  <div className="relative w-full aspect-square bg-neutral-100/70 flex items-center justify-center p-3">
                    <Link to={`/products/${item.product_id}`} className="w-full h-full flex items-center justify-center overflow-hidden rounded">
                      <img
                        src={item.image_url || '/placeholder-product.png'}
                        alt={item.product_name}
                        className="object-contain max-h-full max-w-full mix-blend-multiply group-hover:scale-105 transition-transform duration-550 ease-out"
                      />
                    </Link>

                    {/* Remove Heart Button Overlay */}
                    <button
                      type="button"
                      disabled={loadingRemoveMap[item.variant_id]}
                      onClick={() => handleRemoveFromWishlist(item.variant_id)}
                      className="absolute top-2.5 right-2.5 w-8 h-8 rounded-full bg-white shadow-sm flex items-center justify-center text-red-500 hover:scale-105 transition-all disabled:opacity-50"
                      title="Xóa khỏi danh sách yêu thích"
                    >
                      {loadingRemoveMap[item.variant_id] ? (
                        <svg className="animate-spin h-4 w-4 text-red-500" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                        </svg>
                      ) : (
                        <svg className="w-4 h-4 fill-red-500 text-red-500" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                        </svg>
                      )}
                    </button>

                    {/* Discount Tag */}
                    {hasDiscount && (
                      <div className="absolute top-2.5 left-2.5 bg-red-500 text-white text-[9px] font-extrabold px-1.5 py-0.5 rounded-sm uppercase tracking-wide">
                        -{discountPercent}%
                      </div>
                    )}

                    {/* Out of Stock Overlay */}
                    {item.stock === 0 && (
                      <div className="absolute inset-0 bg-white/80 backdrop-blur-[1px] flex items-center justify-center">
                        <span className="bg-neutral-900 text-white text-[10px] font-extrabold uppercase px-3 py-1 rounded-sm tracking-wider">
                          Hết hàng
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Info Details Section */}
                  <div className="flex flex-col flex-1 p-3.5">
                    {/* Title */}
                    <Link to={`/products/${item.product_id}`} className="block mb-2">
                      <h3 className="line-clamp-2 text-[13px] font-semibold text-neutral-800 leading-snug hover:text-black transition-colors min-h-[36px]">
                        {item.product_name}
                      </h3>
                    </Link>

                    {/* Variant Name Badges */}
                    {item.variant_name && (
                      <span className="inline-block text-[10px] text-neutral-500 bg-neutral-100 rounded px-2 py-0.5 self-start uppercase font-bold tracking-wide mb-2.5">
                        Phân loại: {item.variant_name}
                      </span>
                    )}

                    {/* Rating badge */}
                    <div className="flex items-center gap-2 mb-3 text-[11px] text-neutral-550">
                      <div className="flex items-center text-amber-500">
                        <span className="text-xs mr-0.5">★</span>
                        <span className="font-bold text-neutral-700">{item.rating ? item.rating.toFixed(1) : '0.0'}</span>
                      </div>
                      <span className="text-neutral-300">|</span>
                      <span>SKU: {item.sku || 'N/A'}</span>
                    </div>

                    {/* Price & Cart Actions */}
                    <div className="mt-auto pt-3 border-t border-neutral-100 flex items-center justify-between gap-2">
                      <div className="flex flex-col">
                        {hasDiscount && (
                          <span className="text-[10px] text-neutral-400 line-through mb-0.5">
                            {originalPrice.toLocaleString('vi-VN')} đ
                          </span>
                        )}
                        <span className="text-sm font-black text-neutral-900">
                          {displayPrice.toLocaleString('vi-VN')} đ
                        </span>
                      </div>

                      <button
                        type="button"
                        disabled={item.stock === 0 || loadingCartMap[item.variant_id]}
                        onClick={() => handleAddToCart(item.variant_id)}
                        className="w-8 h-8 rounded-full border border-neutral-250 flex items-center justify-center text-neutral-750 hover:bg-black hover:text-white hover:border-black active:scale-95 disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-neutral-750 disabled:hover:border-neutral-250 transition-all shrink-0"
                        title="Thêm vào giỏ hàng"
                      >
                        {loadingCartMap[item.variant_id] ? (
                          <svg className="animate-spin h-4 w-4 text-black" fill="none" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                          </svg>
                        ) : (
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
                          </svg>
                        )}
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default WishlistPage
