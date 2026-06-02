import { useEffect, useState } from 'react'
import { voucherAPI } from '../services/voucherAPI'
import type { Voucher } from '../types'

interface VoucherSelectorProps {
  onSelect: (code: string) => void
  onClose: () => void
  orderAmount: number
  selectedCode?: string
}

export default function VoucherSelector({
  onSelect,
  onClose,
  orderAmount,
  selectedCode,
}: VoucherSelectorProps) {
  const [vouchers, setVouchers] = useState<Voucher[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchVouchers = async () => {
      try {
        setLoading(true)
        const data = await voucherAPI.getVouchers()
        // Filter out soft-deleted vouchers
        setVouchers(data.filter((v) => !v.is_deleted))
      } catch (err: any) {
        setError(err.message || 'Không thể tải danh sách mã giảm giá')
      } finally {
        setLoading(false)
      }
    }
    void fetchVouchers()
  }, [])

  const formatDiscount = (voucher: Voucher) => {
    if (voucher.discount_type === 'percentage') {
      return `${voucher.discount_value}%`
    }
    return `${voucher.discount_value.toLocaleString('vi-VN')} đ`
  }

  const formatExpiry = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('vi-VN', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm transition-opacity duration-300">
      <div className="relative w-full max-w-lg bg-white rounded-lg shadow-2xl flex flex-col max-h-[85vh] border border-neutral-200 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-neutral-150">
          <h2 className="text-sm font-black text-neutral-900 uppercase tracking-wider flex items-center gap-2">
            Chọn Mã Giảm Giá
          </h2>
          <button
            onClick={onClose}
            className="text-neutral-400 hover:text-black p-1 rounded-full hover:bg-neutral-100 transition-colors"
            title="Đóng"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4 min-h-[300px]">
          {loading ? (
            <div className="flex flex-col items-center justify-center py-16 space-y-3">
              <div className="animate-spin rounded-full h-7 w-7 border-b-2 border-black"></div>
              <p className="text-[10px] font-bold text-neutral-500 uppercase tracking-widest">Đang tải mã giảm giá...</p>
            </div>
          ) : error ? (
            <div className="text-center py-12 space-y-3">
              <p className="text-xs text-red-500 font-semibold">{error}</p>
              <button
                onClick={() => window.location.reload()}
                className="text-[10px] bg-black text-white px-4 py-2 font-black uppercase tracking-wider rounded hover:bg-neutral-850"
              >
                Thử lại
              </button>
            </div>
          ) : vouchers.length === 0 ? (
            <div className="text-center py-16 space-y-2">
              <p className="text-xs font-bold text-neutral-800 uppercase tracking-wider">Hiện không có mã giảm giá nào</p>
              <p className="text-[11px] text-neutral-400 max-w-xs mx-auto">Các chương trình khuyến mãi sẽ được cập nhật sớm nhất.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {vouchers.map((voucher) => {
                const isAmountMet = orderAmount >= voucher.min_order_value
                const isSelected = selectedCode === voucher.code
                const maxUsages = voucher.max_usage_total
                const isUsedUp = maxUsages !== null && voucher.used_count >= maxUsages

                const isDisabled = !isAmountMet || isUsedUp

                return (
                  <div
                    key={voucher.id}
                    className={`relative flex border rounded-lg overflow-hidden transition-all ${
                      isSelected
                        ? 'border-black ring-1 ring-black bg-neutral-50'
                        : isDisabled
                        ? 'border-neutral-200 bg-neutral-50 opacity-60'
                        : 'border-neutral-250 bg-white hover:border-neutral-400'
                    }`}
                  >
                    {/* Left Coupon Ticket Part */}
                    <div
                      className={`w-24 shrink-0 flex flex-col items-center justify-center text-center p-3 text-white select-none ${
                        isSelected
                          ? 'bg-black'
                          : isDisabled
                          ? 'bg-neutral-400'
                          : 'bg-neutral-900'
                      }`}
                    >
                      <span className="text-lg font-black tracking-tight leading-none">
                        {formatDiscount(voucher)}
                      </span>
                      <span className="text-[9px] uppercase font-bold tracking-wider mt-1 opacity-80">
                        GIẢM GIÁ
                      </span>
                    </div>

                    {/* Middle ticket dash divider */}
                    <div className="relative w-0 border-r border-dashed border-neutral-300 h-full flex flex-col justify-between items-center z-10 shrink-0">
                      <div className="absolute -top-1 -left-1 w-2 h-2 rounded-full bg-white border-b border-neutral-200"></div>
                      <div className="absolute -bottom-1 -left-1 w-2 h-2 rounded-full bg-white border-t border-neutral-200"></div>
                    </div>

                    {/* Right Coupon Details Part */}
                    <div className="flex-1 p-4 flex flex-col justify-between">
                      <div className="space-y-1">
                        <div className="flex items-start justify-between gap-2">
                          <h3 className="text-xs font-black text-neutral-850 uppercase tracking-wide">
                            {voucher.name}
                          </h3>
                          <span className="text-[10px] font-mono font-bold bg-neutral-100 border border-neutral-250 px-2 py-0.5 rounded text-neutral-700">
                            {voucher.code}
                          </span>
                        </div>
                        
                        <p className="text-[10px] text-neutral-500 font-medium leading-relaxed">
                          Áp dụng cho: Đơn hàng từ {voucher.min_order_value.toLocaleString('vi-VN')} đ
                          {voucher.max_discount_amount && (
                            <span> (Giảm tối đa {voucher.max_discount_amount.toLocaleString('vi-VN')} đ)</span>
                          )}
                        </p>
                      </div>

                      {/* Footer & Select Button inside details */}
                      <div className="flex items-end justify-between mt-3 pt-3 border-t border-neutral-100">
                        <div className="text-[9px] text-neutral-400 font-medium">
                          HSD: {formatExpiry(voucher.end_date)}
                        </div>

                        {isDisabled ? (
                          <span className="text-[9px] font-bold text-red-500 uppercase tracking-wide">
                            {!isAmountMet
                              ? `Chưa đủ đk (thiếu ${(voucher.min_order_value - orderAmount).toLocaleString('vi-VN')} đ)`
                              : 'Hết lượt dùng'}
                          </span>
                        ) : (
                          <button
                            type="button"
                            onClick={() => {
                              onSelect(voucher.code)
                              onClose()
                            }}
                            className={`text-[10px] font-black uppercase tracking-wider px-3.5 py-1.5 rounded transition-all border ${
                              isSelected
                                ? 'bg-white text-black border-black hover:bg-neutral-50'
                                : 'bg-black text-white border-transparent hover:bg-neutral-800'
                            }`}
                          >
                            {isSelected ? 'Đang chọn' : 'Áp dụng'}
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 bg-neutral-50 border-t border-neutral-150 flex justify-end">
          <button
            onClick={onClose}
            className="text-[10px] border border-neutral-350 text-neutral-600 hover:text-black font-black uppercase tracking-widest px-6 py-2.5 rounded transition-colors"
          >
            Đóng
          </button>
        </div>
      </div>
    </div>
  )
}
