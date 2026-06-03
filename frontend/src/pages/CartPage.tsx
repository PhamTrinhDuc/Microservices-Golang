import { Link } from 'react-router-dom'
import { useCart } from '../hooks/useCart'
import { useState } from 'react'

const CartPage = () => {
  const { items, loading, error, cartSubtotal, updateItemQty, removeItem, clearCart } = useCart()
  const [updatingId, setUpdatingId] = useState<number | null>(null)

  // Free shipping threshold configuration: 299,000 đ
  const FREE_SHIPPING_THRESHOLD = 299000
  const SHIPPING_FEE = 30000

  const isFreeShipping = cartSubtotal >= FREE_SHIPPING_THRESHOLD
  const amountToFreeShipping = FREE_SHIPPING_THRESHOLD - cartSubtotal
  const freeShippingProgress = Math.min(100, (cartSubtotal / FREE_SHIPPING_THRESHOLD) * 100)

  const shippingCost = items.length === 0 ? 0 : (isFreeShipping ? 0 : SHIPPING_FEE)
  const cartTotal = cartSubtotal + shippingCost

  const handleQtyChange = async (itemId: number, currentQty: number, change: number) => {
    const newQty = currentQty + change
    if (newQty < 1) return
    try {
      setUpdatingId(itemId)
      await updateItemQty(itemId, newQty).unwrap()
    } catch (err: any) {
      alert(err || 'Không thể cập nhật số lượng')
    } finally {
      setUpdatingId(null)
    }
  }

  const handleRemove = async (itemId: number) => {
    if (!confirm('Bạn chắc chắn muốn xóa sản phẩm này khỏi giỏ hàng?')) return
    try {
      setUpdatingId(itemId)
      await removeItem(itemId).unwrap()
    } catch (err: any) {
      alert(err || 'Không thể xóa sản phẩm')
    } finally {
      setUpdatingId(null)
    }
  }

  const handleClear = async () => {
    if (!confirm('Bạn chắc chắn muốn xóa toàn bộ giỏ hàng?')) return
    try {
      await clearCart().unwrap()
    } catch (err: any) {
      alert(err || 'Không thể xóa giỏ hàng')
    }
  }

  if (loading && items.length === 0) {
    return (
      <div className="flex-1 bg-neutral-50 py-12 flex items-center justify-center min-h-[400px]">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto"></div>
          <p className="text-xs font-semibold text-neutral-500 uppercase tracking-wider">Đang tải giỏ hàng...</p>
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
          <span className="text-neutral-800 font-semibold">Giỏ hàng của bạn</span>
        </nav>

        <h1 className="text-2xl font-black text-neutral-900 tracking-tight uppercase mb-8">
          Giỏ Hàng ({items.length})
        </h1>

        {items.length === 0 ? (
          /* EMPTY STATE */
          <div className="max-w-md mx-auto text-center border border-neutral-200 border-dashed rounded-lg p-12 bg-white shadow-sm space-y-5">
            <div className="w-16 h-16 rounded-full bg-neutral-100 flex items-center justify-center text-neutral-450 mx-auto">
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
            </div>
            <div>
              <h3 className="text-sm font-bold text-neutral-800 uppercase tracking-wider">Giỏ hàng của bạn đang trống</h3>
              <p className="text-xs text-neutral-400 leading-relaxed mt-1 max-w-xs mx-auto">
                Hiện chưa có sản phẩm nào được thêm vào giỏ hàng. Hãy quay lại cửa hàng và chọn các sản phẩm bạn ưng ý nhé!
              </p>
            </div>
            <Link
              to="/browse"
              className="inline-block bg-black text-white text-xs font-extrabold uppercase px-8 py-3 rounded hover:bg-neutral-850 transition-colors"
            >
              Tiếp tục mua sắm
            </Link>
          </div>
        ) : (
          /* ACTIVE CART VIEW */
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            
            {/* Left: Cart Items List */}
            <div className="lg:col-span-8 space-y-4">
              {error && (
                <div className="border border-red-200 bg-red-50 text-red-650 text-xs font-semibold px-4 py-3 rounded-lg flex items-center justify-between">
                  <span>{error}</span>
                </div>
              )}

              {/* Clear all action */}
              <div className="flex justify-end">
                <button
                  type="button"
                  onClick={handleClear}
                  className="text-xs font-bold text-red-500 hover:text-red-650 transition-colors uppercase tracking-wider"
                >
                  Xóa sạch giỏ hàng
                </button>
              </div>

              {/* Items Card List */}
              <div className="bg-white border border-neutral-200 rounded-lg divide-y divide-neutral-150 overflow-hidden shadow-sm">
                {items.map((item) => (
                  <div key={item.id} className="p-5 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between relative">
                    {/* Overlay loader when updating */}
                    {updatingId === item.id && (
                      <div className="absolute inset-0 bg-white/60 flex items-center justify-center z-10 transition-opacity">
                        <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-black"></div>
                      </div>
                    )}

                    {/* Product Image & Meta */}
                    <div className="flex gap-4 items-center flex-1">
                      <Link to={`/products/${item.product_id}`} className="w-16 h-16 bg-neutral-50 border border-neutral-200 p-1 flex items-center justify-center rounded shrink-0 overflow-hidden">
                        <img
                          src={item.image_url || '/placeholder-product.png'}
                          alt={item.product_name}
                          className="max-h-full max-w-full object-contain mix-blend-multiply"
                        />
                      </Link>
                      
                      <div className="space-y-1">
                        <Link to={`/products/${item.product_id}`} className="text-xs font-bold text-neutral-850 hover:underline hover:text-black line-clamp-1">
                          {item.product_name}
                        </Link>
                        {item.variant_name && (
                          <div className="text-[10px] text-neutral-450 uppercase font-extrabold tracking-wide">
                            Phân loại: {item.variant_name}
                          </div>
                        )}
                        <div className="text-[10px] text-neutral-400 font-medium">SKU: {item.sku || 'N/A'}</div>
                      </div>
                    </div>

                    {/* Quantity selectors + Price fields */}
                    <div className="flex items-center justify-between sm:justify-end gap-6 w-full sm:w-auto mt-4 sm:mt-0 pt-4 sm:pt-0 border-t sm:border-t-0 border-neutral-150">
                      
                      {/* Quantity adjustment buttons */}
                      <div className="flex items-center border border-neutral-250 rounded bg-white shrink-0">
                        <button
                          type="button"
                          disabled={item.quantity <= 1 || updatingId === item.id}
                          onClick={() => handleQtyChange(item.id, item.quantity, -1)}
                          className="w-8 h-8 flex items-center justify-center font-bold text-neutral-600 hover:bg-neutral-100 disabled:opacity-30 transition-colors"
                        >
                          −
                        </button>
                        <span className="w-8 text-center text-xs font-bold text-neutral-850">
                          {item.quantity}
                        </span>
                        <button
                          type="button"
                          disabled={updatingId === item.id}
                          onClick={() => handleQtyChange(item.id, item.quantity, 1)}
                          className="w-8 h-8 flex items-center justify-center font-bold text-neutral-600 hover:bg-neutral-100 disabled:opacity-30 transition-colors"
                        >
                          +
                        </button>
                      </div>

                      {/* Pricing block */}
                      <div className="flex flex-col items-end shrink-0 min-w-[100px]">
                        <span className="text-xs font-black text-neutral-900">
                          {(item.sell_price * item.quantity).toLocaleString('vi-VN')} đ
                        </span>
                        {item.quantity > 1 && (
                          <span className="text-[10px] text-neutral-400 mt-0.5">
                            {item.sell_price.toLocaleString('vi-VN')} đ / sp
                          </span>
                        )}
                      </div>

                      {/* Delete icon button */}
                      <button
                        type="button"
                        onClick={() => handleRemove(item.id)}
                        disabled={updatingId === item.id}
                        className="text-neutral-400 hover:text-red-500 hover:bg-red-50 p-2 rounded-full transition-all shrink-0"
                        title="Xóa sản phẩm"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>

                    </div>

                  </div>
                ))}
              </div>
            </div>

            {/* Right: Summary panel */}
            <div className="lg:col-span-4 space-y-5">
              
              {/* Free shipping progress card */}
              <div className="border border-neutral-200 bg-white rounded-lg p-5 shadow-sm space-y-3">
                <div className="flex items-center justify-between text-xs font-bold text-neutral-800 uppercase tracking-wider">
                  <span>🚚 Vận Chuyển Hỏa Tốc</span>
                  {isFreeShipping ? (
                    <span className="text-green-600 text-[10px] bg-green-50 px-2 py-0.5 rounded font-black">Miễn phí</span>
                  ) : (
                    <span className="text-neutral-500 text-[10px]">Chỉ từ 30kđ</span>
                  )}
                </div>
                
                {/* Progress bar */}
                <div className="w-full bg-neutral-100 rounded-full h-1.5 overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${isFreeShipping ? 'bg-green-500' : 'bg-black'}`}
                    style={{ width: `${freeShippingProgress}%` }}
                  ></div>
                </div>

                {/* Progress text */}
                <p className="text-[11px] text-neutral-500 leading-relaxed">
                  {isFreeShipping ? (
                    <span>🎉 Chúc mừng! Đơn hàng của bạn đủ điều kiện nhận <strong>Miễn phí vận chuyển</strong>.</span>
                  ) : (
                    <span>Mua thêm <strong>{amountToFreeShipping.toLocaleString('vi-VN')} đ</strong> để được <strong>Miễn phí vận chuyển</strong> toàn quốc!</span>
                  )}
                </p>
              </div>

              {/* Order total estimation block */}
              <div className="border border-neutral-200 bg-white rounded-lg p-5 shadow-sm space-y-4">
                <h3 className="text-xs font-black text-neutral-800 uppercase tracking-wider pb-3 border-b border-neutral-150">
                  Tóm Tắt Đơn Hàng
                </h3>

                <div className="space-y-2.5 text-xs">
                  <div className="flex justify-between text-neutral-550">
                    <span>Tạm tính ({items.reduce((sum, i) => sum + i.quantity, 0)} sp)</span>
                    <span className="font-semibold text-neutral-850">{cartSubtotal.toLocaleString('vi-VN')} đ</span>
                  </div>
                  <div className="flex justify-between text-neutral-550">
                    <span>Phí vận chuyển</span>
                    <span className="font-semibold text-neutral-850">
                      {shippingCost === 0 ? 'Miễn phí' : `${shippingCost.toLocaleString('vi-VN')} đ`}
                    </span>
                  </div>
                </div>

                <div className="border-t border-neutral-150 pt-4 flex justify-between items-baseline">
                  <span className="text-xs font-bold text-neutral-800 uppercase tracking-wide">Tổng cộng</span>
                  <span className="text-xl font-black text-neutral-900">
                    {cartTotal.toLocaleString('vi-VN')} đ
                  </span>
                </div>

                <Link
                  to="/checkout"
                  className="block w-full text-center bg-black text-white text-xs font-black uppercase tracking-wider py-3.5 rounded hover:bg-neutral-850 transition-colors shadow-sm mt-3"
                >
                  Tiến hành thanh toán
                </Link>

                <div className="text-[10px] text-neutral-400 text-center leading-relaxed mt-2.5">
                  🛡️ Thanh toán bảo mật SSL. Hoàn trả tiền dễ dàng trong 7 ngày nếu do lỗi sản xuất.
                </div>
              </div>

            </div>

          </div>
        )}
      </div>
    </div>
  )
}

export default CartPage
