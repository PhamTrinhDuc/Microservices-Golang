import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { orderAPI } from '../services/orderAPI'
import type { OrderResponse } from '../types'

export default function OrderSuccessPage() {
  const [searchParams] = useSearchParams()
  const orderIdStr = searchParams.get('id')
  const orderCode = searchParams.get('code')

  const [orderData, setOrderData] = useState<OrderResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!orderIdStr) {
      setLoading(false)
      return
    }

    const fetchOrderDetails = async () => {
      try {
        setLoading(true)
        const orderId = parseInt(orderIdStr, 10)
        if (isNaN(orderId)) {
          throw new Error('Mã đơn hàng không hợp lệ')
        }
        const data = await orderAPI.getMyOrderDetails(orderId)
        setOrderData(data)
      } catch (err: any) {
        console.error(err)
        setError(err.message || 'Không thể tải chi tiết đơn hàng')
      } finally {
        setLoading(false)
      }
    }

    void fetchOrderDetails()
  }, [orderIdStr])

  if (loading) {
    return (
      <div className="flex-1 bg-neutral-50 py-16 flex items-center justify-center min-h-[450px]">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto"></div>
          <p className="text-xs font-bold text-neutral-500 uppercase tracking-widest">Đang tải thông tin đơn hàng...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex-1 bg-neutral-50 py-16 flex items-center justify-center min-h-[450px]">
        <div className="text-center space-y-4">
          <p className="text-xs text-red-500 font-semibold uppercase tracking-wider">{error}</p>
          <Link
            to="/browse"
            className="inline-block bg-black text-white text-xs font-black uppercase tracking-widest px-6 py-2.5 rounded hover:bg-neutral-850 transition-colors shadow-sm"
          >
            Quay lại mua sắm
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 bg-neutral-50 py-12 font-sans">
      <div className="mx-auto max-w-3xl px-4 text-center">
        
        {/* CHECKMARK CARD */}
        <div className="bg-white border border-neutral-200 rounded-lg p-8 shadow-sm space-y-6">
          <div className="w-16 h-16 bg-neutral-100 text-neutral-850 rounded-full flex items-center justify-center mx-auto border border-neutral-250 select-none">
            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl font-black text-neutral-900 uppercase tracking-tight">Đặt Hàng Thành Công!</h1>
            <p className="text-xs text-neutral-450 leading-relaxed max-w-md mx-auto">
              Cảm ơn bạn đã mua hàng tại <strong>Jiyuu Store</strong>. Đơn hàng của bạn đang được hệ thống xử lý.
            </p>
          </div>

          <div className="py-4 border-t border-b border-neutral-150 max-w-md mx-auto grid grid-cols-2 gap-4 text-left text-xs">
            <div>
              <p className="text-neutral-400 font-bold uppercase text-[9px] tracking-wider">Mã đơn hàng</p>
              <p className="font-mono font-bold text-neutral-850 mt-1">{orderCode || orderData?.order.order_code || 'N/A'}</p>
            </div>
            <div>
              <p className="text-neutral-400 font-bold uppercase text-[9px] tracking-wider">Phương thức thanh toán</p>
              <p className="font-bold text-neutral-850 uppercase mt-1">
                {orderData?.order.payment_method === 'payos'
                  ? 'PayOS'
                  : orderData?.order.payment_method === 'bank_transfer'
                  ? 'Chuyển khoản'
                  : 'Tiền mặt (COD)'}
              </p>
            </div>
          </div>

          {/* Payment Prompt if PayOS is Unpaid */}
          {orderData?.order.payment_method === 'payos' && orderData.payment_status_label === 'unpaid' && orderData.checkout_url && (
            <div className="bg-sky-50 border border-sky-200 text-sky-850 rounded-lg p-5 max-w-md mx-auto text-left space-y-3">
              <p className="text-xs font-bold uppercase tracking-wide flex items-center gap-1.5">
                Đơn hàng chưa được thanh toán
              </p>
              <p className="text-[11px] leading-relaxed text-sky-700">
                Bạn đã chọn thanh toán qua PayOS. Vui lòng nhấn vào nút bên dưới để thanh toán và giữ chỗ sản phẩm.
              </p>
              <a
                href={orderData.checkout_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-block bg-sky-600 hover:bg-sky-700 text-white text-xs font-black uppercase tracking-wider px-6 py-2.5 rounded shadow-sm transition-colors"
              >
                Thanh toán ngay qua PayOS
              </a>
            </div>
          )}

          {/* Detailed Invoice Info */}
          {orderData && (
            <div className="text-left space-y-4 max-w-xl mx-auto border border-neutral-200 rounded-lg p-5">
              <h3 className="text-xs font-black text-neutral-850 uppercase tracking-widest border-b border-neutral-100 pb-2">
                Thông Tin Giao Hàng & Đơn Hàng
              </h3>

              {/* Shipping address info */}
              <div className="text-[11px] space-y-1 text-neutral-600">
                <p><strong>Người nhận:</strong> {orderData.order.receiver_name}</p>
                <p><strong>Số điện thoại:</strong> {orderData.order.receiver_phone}</p>
                <p><strong>Địa chỉ giao:</strong> {orderData.order.receiver_address}</p>
                {orderData.order.shipping_provider && (
                  <p><strong>Đơn vị vận chuyển:</strong> {orderData.order.shipping_provider.toUpperCase()}</p>
                )}
              </div>

              {/* Items List */}
              <div className="border-t border-neutral-150 pt-3.5 space-y-2">
                {orderData.items.map((item) => (
                  <div key={item.id} className="flex justify-between items-center text-xs">
                    <div className="min-w-0 flex-1 pr-4">
                      <p className="font-bold text-neutral-800 truncate">{item.variant_name}</p>
                      <p className="text-[10px] text-neutral-400 mt-0.5">SL: {item.quantity} x {item.unit_price.toLocaleString('vi-VN')} đ</p>
                    </div>
                    <span className="font-mono font-bold text-neutral-850 shrink-0">
                      {item.total_cost.toLocaleString('vi-VN')} đ
                    </span>
                  </div>
                ))}
              </div>

              {/* Cost breakdown */}
              <div className="border-t border-neutral-150 pt-3 space-y-1.5 text-xs">
                <div className="flex justify-between text-neutral-500 text-[11px]">
                  <span>Phí vận chuyển</span>
                  <span>{orderData.order.shipping_price === 0 ? 'Miễn phí' : `${orderData.order.shipping_price.toLocaleString('vi-VN')} đ`}</span>
                </div>
                {orderData.order.voucher_discount > 0 && (
                  <div className="flex justify-between text-green-700 text-[11px] font-semibold">
                    <span>Mã giảm giá đã áp dụng</span>
                    <span>-{orderData.order.voucher_discount.toLocaleString('vi-VN')} đ</span>
                  </div>
                )}
                <div className="flex justify-between text-neutral-850 font-black pt-1.5 border-t border-neutral-100 text-sm">
                  <span>Tổng tiền</span>
                  <span>{orderData.order.total_amount.toLocaleString('vi-VN')} đ</span>
                </div>
              </div>
            </div>
          )}

          {/* Action buttons */}
          <div className="flex flex-col sm:flex-row gap-3 justify-center pt-2 max-w-sm mx-auto">
            <Link
              to="/browse"
              className="flex-1 bg-black text-white text-xs font-black uppercase tracking-widest py-3 rounded hover:bg-neutral-850 transition-colors shadow-sm"
            >
              Tiếp tục mua sắm
            </Link>
          </div>
        </div>

      </div>
    </div>
  )
}
