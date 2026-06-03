import React, { useEffect, useState } from 'react'
import { voucherAPI } from '../../services/voucherAPI'
import type { Voucher } from '../../types'

export default function AdminVouchersTab() {
  const [vouchers, setVouchers] = useState<Voucher[]>([])
  const [loading, setLoading] = useState(false)

  // Vouchers form state
  const [showVoucherForm, setShowVoucherForm] = useState(false)
  const [newVoucher, setNewVoucher] = useState({
    code: '',
    name: '',
    startDate: '',
    endDate: '',
    discountType: 'percentage' as 'percentage' | 'fixed',
    discountValue: 0,
    discountTarget: 'order' as 'order' | 'shipping',
    minOrderValue: 0,
    maxDiscountAmount: '',
    maxUsageTotal: '',
    maxUsagePerUser: 1,
  })

  const loadData = async () => {
    try {
      setLoading(true)
      const vList = await voucherAPI.adminListVouchers()
      setVouchers(vList)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadData()
  }, [])

  const handleCreateVoucher = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      const payload = {
        code: newVoucher.code.trim().toUpperCase(),
        name: newVoucher.name.trim(),
        start_date: new Date(newVoucher.startDate).toISOString(),
        end_date: new Date(newVoucher.endDate).toISOString(),
        discount_type: newVoucher.discountType,
        discount_value: Number(newVoucher.discountValue),
        discount_target: newVoucher.discountTarget,
        min_order_value: Number(newVoucher.minOrderValue),
        max_discount_amount: newVoucher.maxDiscountAmount ? Number(newVoucher.maxDiscountAmount) : null,
        max_usage_total: newVoucher.maxUsageTotal ? Number(newVoucher.maxUsageTotal) : null,
        max_usage_per_user: Number(newVoucher.maxUsagePerUser),
      }

      await voucherAPI.adminCreateVoucher(payload)
      alert('Tạo mã giảm giá thành công!')
      setShowVoucherForm(false)
      setNewVoucher({
        code: '',
        name: '',
        startDate: '',
        endDate: '',
        discountType: 'percentage',
        discountValue: 0,
        discountTarget: 'order',
        minOrderValue: 0,
        maxDiscountAmount: '',
        maxUsageTotal: '',
        maxUsagePerUser: 1,
      })
      void loadData()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi tạo voucher')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteVoucher = async (id: number) => {
    if (!confirm('Bạn chắc chắn muốn xóa voucher này? (Xóa mềm)')) return
    try {
      setLoading(true)
      await voucherAPI.adminDeleteVoucher(id)
      alert('Đã xóa voucher thành công!')
      void loadData()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xóa voucher')
    } finally {
      setLoading(false)
    }
  }



  return (
    <div className="space-y-8">
      {/* Vouchers section */}
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Mã Giảm Giá (Vouchers)</h1>
            <p className="text-xs text-neutral-500 mt-1">Quản lý và kích hoạt mã giảm giá toàn hệ thống</p>
          </div>
          <button
            onClick={() => setShowVoucherForm(!showVoucherForm)}
            className="bg-black hover:bg-neutral-800 text-white text-xs font-black uppercase tracking-wider px-4 py-2 rounded transition-colors"
          >
            {showVoucherForm ? 'Đóng form' : '+ Tạo Voucher mới'}
          </button>
        </div>

        {/* Voucher Creator form */}
        {showVoucherForm && (
          <form onSubmit={handleCreateVoucher} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs">
            <h3 className="text-xs font-black uppercase tracking-wide">Nhập thông tin Voucher</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <input
                type="text"
                placeholder="Mã Voucher * (Ví dụ: BLACKFRIDAY)"
                className="border border-neutral-350 rounded px-3 py-2 uppercase font-mono font-bold w-full"
                value={newVoucher.code}
                onChange={(e) => setNewVoucher((p) => ({ ...p, code: e.target.value }))}
                required
              />
              <input
                type="text"
                placeholder="Tên chương trình *"
                className="border border-neutral-350 rounded px-3 py-2 w-full col-span-2"
                value={newVoucher.name}
                onChange={(e) => setNewVoucher((p) => ({ ...p, name: e.target.value }))}
                required
              />
              <select
                className="border border-neutral-350 rounded px-2.5 py-2 w-full"
                value={newVoucher.discountTarget}
                onChange={(e) => setNewVoucher((p) => ({ ...p, discountTarget: e.target.value as 'order' | 'shipping' }))}
              >
                <option value="order">Khấu trừ Đơn hàng</option>
                <option value="shipping">Khấu trừ Ship (Free Ship)</option>
              </select>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <select
                className="border border-neutral-350 rounded px-2.5 py-2 w-full"
                value={newVoucher.discountType}
                onChange={(e) => setNewVoucher((p) => ({ ...p, discountType: e.target.value as 'percentage' | 'fixed' }))}
              >
                <option value="percentage">Giảm theo Phần trăm (%)</option>
                <option value="fixed">Số tiền cố định (VND)</option>
              </select>
              <input
                type="number"
                placeholder="Giá trị giảm *"
                className="border border-neutral-350 rounded px-3 py-2 w-full font-bold"
                value={newVoucher.discountValue || ''}
                onChange={(e) => setNewVoucher((p) => ({ ...p, discountValue: Number(e.target.value) }))}
                required
              />
              <input
                type="number"
                placeholder="Giảm tối đa (Tùy chọn)"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={newVoucher.maxDiscountAmount}
                onChange={(e) => setNewVoucher((p) => ({ ...p, maxDiscountAmount: e.target.value }))}
              />
              <input
                type="number"
                placeholder="Đơn tối thiểu (Min value) *"
                className="border border-neutral-350 rounded px-3 py-2 w-full font-bold"
                value={newVoucher.minOrderValue || ''}
                onChange={(e) => setNewVoucher((p) => ({ ...p, minOrderValue: Number(e.target.value) }))}
                required
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <input
                type="datetime-local"
                placeholder="Thời gian bắt đầu *"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={newVoucher.startDate}
                onChange={(e) => setNewVoucher((p) => ({ ...p, startDate: e.target.value }))}
                required
              />
              <input
                type="datetime-local"
                placeholder="Thời gian kết thúc *"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={newVoucher.endDate}
                onChange={(e) => setNewVoucher((p) => ({ ...p, endDate: e.target.value }))}
                required
              />
              <input
                type="number"
                placeholder="Tổng số lượt tối đa (Tùy chọn)"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={newVoucher.maxUsageTotal}
                onChange={(e) => setNewVoucher((p) => ({ ...p, maxUsageTotal: e.target.value }))}
              />
              <input
                type="number"
                placeholder="Lượt dùng tối đa / 1 user *"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={newVoucher.maxUsagePerUser}
                onChange={(e) => setNewVoucher((p) => ({ ...p, maxUsagePerUser: Number(e.target.value) }))}
                required
              />
            </div>

            <div className="flex justify-end gap-2">
              <button
                type="submit"
                disabled={loading}
                className="bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider px-6 py-2.5 rounded disabled:opacity-40"
              >
                Tạo Voucher
              </button>
            </div>
          </form>
        )}

        {/* Vouchers Table */}
        <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                <th className="p-4">Mã / Tên</th>
                <th className="p-4">Hình thức giảm</th>
                <th className="p-4">Áp dụng cho</th>
                <th className="p-4">Lượt đã dùng</th>
                <th className="p-4">Hạn sử dụng</th>
                <th className="p-4 text-center">Xóa</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-150">
              {vouchers.map((v) => (
                <tr key={v.id} className="hover:bg-neutral-50 transition-colors">
                  <td className="p-4">
                    <p className="font-mono font-bold text-neutral-850 uppercase bg-neutral-100 border inline-block px-1.5 py-0.5 rounded text-[10px]">
                      {v.code}
                    </p>
                    <p className="font-bold text-neutral-800 mt-1">{v.name}</p>
                  </td>
                  <td className="p-4">
                    <p className="font-bold text-neutral-900">
                      {v.discount_type === 'percentage' ? `${v.discount_value}%` : `${v.discount_value.toLocaleString('vi-VN')} đ`}
                    </p>
                    {v.max_discount_amount && (
                      <p className="text-[10px] text-neutral-400 mt-0.5">Giảm tối đa {v.max_discount_amount.toLocaleString('vi-VN')} đ</p>
                    )}
                  </td>
                  <td className="p-4">
                    <p className="text-neutral-600">Đơn từ: {v.min_order_value.toLocaleString('vi-VN')} đ</p>
                    <p className="text-[10px] text-neutral-400 uppercase font-black tracking-wide mt-0.5">{v.discount_target === 'shipping' ? 'Free Ship' : 'Đơn hàng'}</p>
                  </td>
                  <td className="p-4 font-mono font-medium text-neutral-600">
                    {v.used_count} / {v.max_usage_total || '∞'}
                  </td>
                  <td className="p-4 font-medium text-neutral-600">
                    {new Date(v.end_date).toLocaleDateString('vi-VN')}
                  </td>
                  <td className="p-4 text-center">
                    <button
                      onClick={() => handleDeleteVoucher(v.id)}
                      className="text-neutral-400 hover:text-red-655 hover:bg-red-50 p-2 rounded transition-all inline-flex items-center justify-center"
                      title="Xóa"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
