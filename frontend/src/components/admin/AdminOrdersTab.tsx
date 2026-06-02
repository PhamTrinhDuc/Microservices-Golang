import React, { useEffect, useState } from 'react'
import { orderAPI } from '../../services/orderAPI'
import type { OrderResponse, Store } from '../../types'

interface AdminOrdersTabProps {
  stores: Store[]
}

export default function AdminOrdersTab({ stores }: AdminOrdersTabProps) {
  const [orders, setOrders] = useState<OrderResponse[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [storeFilter, setStoreFilter] = useState<number | undefined>(undefined)
  const [search, setSearch] = useState('')
  const [orderStatusFilter, setOrderStatusFilter] = useState('')
  const [paymentStatusFilter, setPaymentStatusFilter] = useState('')
  const [shippingStatusFilter, setShippingStatusFilter] = useState('')
  const [loading, setLoading] = useState(false)

  // Selected Order for Detail Modal
  const [selectedOrder, setSelectedOrder] = useState<OrderResponse | null>(null)
  const [showStatusModal, setShowStatusModal] = useState(false)

  // Form for Updating Status
  const [statusForm, setStatusForm] = useState({
    orderStatus: '',
    paymentStatus: '',
    shippingStatus: '',
    shippingProvider: '',
    shippingCode: '',
    note: '',
  })

  const loadOrders = async () => {
    try {
      setLoading(true)
      const res = await orderAPI.adminListOrders(page, 10, {
        storeId: storeFilter,
        q: search,
        orderStatus: orderStatusFilter,
        paymentStatus: paymentStatusFilter,
        shippingStatus: shippingStatusFilter,
      })
      setOrders(res.data || [])
      setTotal(res.total || 0)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadOrders()
  }, [page, storeFilter, search, orderStatusFilter, paymentStatusFilter, shippingStatusFilter])

  const handleCopyCode = (code: string) => {
    void navigator.clipboard.writeText(code)
  }

  const handleOpenStatusModal = async (order: OrderResponse) => {
    try {
      setLoading(true)
      const details = await orderAPI.adminGetOrderDetails(order.order.id)
      setSelectedOrder(details)
      setStatusForm({
        orderStatus: '',
        paymentStatus: '',
        shippingStatus: '',
        shippingProvider: details.order.shipping_provider || '',
        shippingCode: details.order.shipping_code || '',
        note: '',
      })
      setShowStatusModal(true)
    } catch (err: any) {
      alert(err.message || 'Lỗi khi tải chi tiết đơn hàng')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateStatus = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedOrder) return
    try {
      setLoading(true)
      const payload: any = {}
      if (statusForm.orderStatus) payload.order_status_code = statusForm.orderStatus
      if (statusForm.paymentStatus) payload.payment_status_code = statusForm.paymentStatus
      if (statusForm.shippingStatus) payload.shipping_status_code = statusForm.shippingStatus
      if (statusForm.shippingProvider) payload.shipping_provider = statusForm.shippingProvider
      if (statusForm.shippingCode) payload.shipping_code = statusForm.shippingCode
      if (statusForm.note) payload.note = statusForm.note

      await orderAPI.adminUpdateOrderStatus(selectedOrder.order.id, payload)
      alert('Cập nhật trạng thái đơn hàng thành công!')
      setShowStatusModal(false)
      setSelectedOrder(null)
      void loadOrders()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi cập nhật trạng thái')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Đơn hàng</h1>
          <p className="text-xs text-neutral-500 mt-1">Duyệt đơn, trừ tồn kho và cập nhật vận chuyển</p>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="grid grid-cols-1 sm:grid-cols-5 gap-3 bg-neutral-50 p-4 rounded-lg border border-neutral-200">
        <div className="sm:col-span-1">
          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tìm kiếm</label>
          <input
            type="text"
            placeholder="Mã đơn, tên khách, SĐT..."
            className="w-full border border-neutral-300 rounded px-2.5 py-1.5 text-xs bg-white focus:outline-none"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
          />
        </div>

        <div>
          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Cửa hàng</label>
          <select
            className="w-full border border-neutral-300 rounded px-2.5 py-1.5 text-xs bg-white focus:outline-none"
            value={storeFilter || ''}
            onChange={(e) => {
              setStoreFilter(e.target.value ? Number(e.target.value) : undefined)
              setPage(1)
            }}
          >
            <option value="">Tất cả cửa hàng</option>
            {stores.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Trạng thái đơn</label>
          <select
            className="w-full border border-neutral-300 rounded px-2.5 py-1.5 text-xs bg-white focus:outline-none"
            value={orderStatusFilter}
            onChange={(e) => {
              setOrderStatusFilter(e.target.value)
              setPage(1)
            }}
          >
            <option value="">Tất cả</option>
            <option value="pending">Chờ xử lý (Pending)</option>
            <option value="confirmed">Xác nhận (Confirmed)</option>
            <option value="processing">Đang đóng gói (Processing)</option>
            <option value="shipping">Đang vận chuyển (Shipping)</option>
            <option value="delivered">Đã giao hàng (Delivered)</option>
            <option value="cancelled">Hủy bỏ (Cancelled)</option>
          </select>
        </div>

        <div>
          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Thanh toán</label>
          <select
            className="w-full border border-neutral-300 rounded px-2.5 py-1.5 text-xs bg-white focus:outline-none"
            value={paymentStatusFilter}
            onChange={(e) => {
              setPaymentStatusFilter(e.target.value)
              setPage(1)
            }}
          >
            <option value="">Tất cả</option>
            <option value="unpaid">Chưa thanh toán (Unpaid)</option>
            <option value="paid">Đã thanh toán (Paid)</option>
            <option value="refunded">Hoàn tiền (Refunded)</option>
          </select>
        </div>

        <div>
          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Giao hàng</label>
          <select
            className="w-full border border-neutral-300 rounded px-2.5 py-1.5 text-xs bg-white focus:outline-none"
            value={shippingStatusFilter}
            onChange={(e) => {
              setShippingStatusFilter(e.target.value)
              setPage(1)
            }}
          >
            <option value="">Tất cả</option>
            <option value="not_shipped">Chưa giao (Not Shipped)</option>
            <option value="shipped">Đang vận chuyển (Shipped)</option>
            <option value="delivered">Đã hoàn tất giao (Delivered)</option>
          </select>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
        </div>
      ) : (
        <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                <th className="p-4">Đơn hàng</th>
                <th className="p-4">Khách hàng</th>
                <th className="p-4">Cửa hàng</th>
                <th className="p-4 text-right">Tổng tiền</th>
                <th className="p-4">Đơn hàng</th>
                <th className="p-4">Thanh toán</th>
                <th className="p-4">Giao hàng</th>
                <th className="p-4 text-center">Hành động</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-150">
              {orders.map((o) => (
                <tr key={o.order.id} className="hover:bg-neutral-50 transition-colors">
                  <td className="p-4">
                    <span
                      onClick={() => handleCopyCode(o.order.order_code)}
                      className="font-bold text-neutral-900 cursor-pointer hover:text-blue-650 select-none"
                      title="Click để sao chép"
                    >
                      {o.order.order_code}
                    </span>
                    <p className="text-[10px] text-neutral-400 mt-0.5">
                      {new Date(o.order.created_at).toLocaleDateString('vi-VN')}
                    </p>
                  </td>
                  <td className="p-4">
                    <p className="font-semibold">{o.order.receiver_name}</p>
                    <p className="text-[10px] text-neutral-400 font-mono mt-0.5">{o.order.receiver_phone}</p>
                  </td>
                  <td className="p-4 text-neutral-600 font-medium">
                    {stores.find((s) => s.id === o.order.store_id)?.name || `Store #${o.order.store_id}`}
                  </td>
                  <td className="p-4 text-right font-mono font-bold text-neutral-850">
                    {o.order.total_amount.toLocaleString('vi-VN')} đ
                  </td>
                  <td className="p-4">
                    <span className={`px-2 py-0.5 rounded-[4px] text-[10px] font-black uppercase tracking-wider ${
                      o.order_status_label.includes('Chờ') || o.order_status_label.includes('pending')
                        ? 'bg-amber-100 text-amber-700'
                        : o.order_status_label.includes('Hủy') || o.order_status_label.includes('cancelled')
                        ? 'bg-red-100 text-red-700'
                        : 'bg-green-100 text-green-700'
                    }`}>
                      {o.order_status_label}
                    </span>
                  </td>
                  <td className="p-4">
                    <span className={`px-2 py-0.5 rounded-[4px] text-[10px] font-black uppercase tracking-wider ${
                      o.payment_status_label.includes('paid') || o.payment_status_label.includes('Đã')
                        ? 'bg-green-100 text-green-700'
                        : 'bg-neutral-100 text-neutral-600'
                    }`}>
                      {o.payment_status_label}
                    </span>
                  </td>
                  <td className="p-4">
                    <span className="px-2 py-0.5 rounded-[4px] bg-neutral-100 text-neutral-600 text-[10px] font-bold">
                      {o.shipping_status_label}
                    </span>
                  </td>
                  <td className="p-4 text-center">
                    <button
                      onClick={() => handleOpenStatusModal(o)}
                      className="bg-black hover:bg-neutral-800 text-white text-[10px] uppercase font-black tracking-wider px-3.5 py-1.5 rounded transition-colors"
                    >
                      👁 Chi tiết
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination controls */}
      {total > 10 && (
        <div className="flex gap-2 justify-end text-xs font-semibold">
          <button
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
            className="border border-neutral-350 px-3 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
          >
            ← Trước
          </button>
          <span className="px-3 py-1 flex items-center">Trang {page}</span>
          <button
            disabled={page * 10 >= total}
            onClick={() => setPage((p) => p + 1)}
            className="border border-neutral-350 px-3 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
          >
            Sau →
          </button>
        </div>
      )}

      {/* UPDATE STATUS MODAL */}
      {showStatusModal && selectedOrder && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="bg-white border border-neutral-200 rounded-lg shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto p-6 space-y-5 flex flex-col">
            <div className="flex justify-between items-center border-b border-neutral-100 pb-3">
              <h3 className="text-xs font-black uppercase tracking-wider text-neutral-900">
                Chi Tiết Đơn Hàng: {selectedOrder.order.order_code}
              </h3>
              <button
                onClick={() => setShowStatusModal(false)}
                className="text-neutral-400 hover:text-black"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleUpdateStatus} className="space-y-4 text-xs flex-1">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {/* Order Status */}
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Trạng thái đơn hàng</label>
                  <select
                    className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                    value={statusForm.orderStatus}
                    onChange={(e) => setStatusForm((p) => ({ ...p, orderStatus: e.target.value }))}
                  >
                    <option value="">Giữ nguyên</option>
                    <option value="pending">Chờ xử lý (Pending)</option>
                    <option value="confirmed">Xác nhận (Confirmed)</option>
                    <option value="processing">Đang đóng gói (Processing)</option>
                    <option value="shipping">Đang vận chuyển (Shipping)</option>
                    <option value="delivered">Đã giao hàng (Delivered)</option>
                    <option value="cancelled">Hủy bỏ (Cancelled)</option>
                  </select>
                </div>

                {/* Payment Status */}
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Thanh toán</label>
                  <select
                    className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                    value={statusForm.paymentStatus}
                    onChange={(e) => setStatusForm((p) => ({ ...p, paymentStatus: e.target.value }))}
                  >
                    <option value="">Giữ nguyên</option>
                    <option value="unpaid">Chưa thanh toán (Unpaid)</option>
                    <option value="paid">Đã thanh toán (Paid)</option>
                    <option value="refunded">Hoàn tiền (Refunded)</option>
                  </select>
                </div>

                {/* Shipping Status */}
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Giao hàng</label>
                  <select
                    className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                    value={statusForm.shippingStatus}
                    onChange={(e) => setStatusForm((p) => ({ ...p, shippingStatus: e.target.value }))}
                  >
                    <option value="">Giữ nguyên</option>
                    <option value="not_shipped">Chưa giao (Not Shipped)</option>
                    <option value="shipped">Đang vận chuyển (Shipped)</option>
                    <option value="delivered">Đã hoàn tất giao (Delivered)</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {/* Shipping Provider */}
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Nhà vận chuyển</label>
                  <input
                    type="text"
                    placeholder="Ví dụ: ghn, ghtk"
                    className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                    value={statusForm.shippingProvider}
                    onChange={(e) => setStatusForm((p) => ({ ...p, shippingProvider: e.target.value }))}
                  />
                </div>

                {/* Shipping Tracking Code */}
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Mã vận đơn</label>
                  <input
                    type="text"
                    placeholder="Mã tracking vận đơn"
                    className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                    value={statusForm.shippingCode}
                    onChange={(e) => setStatusForm((p) => ({ ...p, shippingCode: e.target.value }))}
                  />
                </div>
              </div>

              {/* Note */}
              <div>
                <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Ghi chú cập nhật</label>
                <textarea
                  rows={2}
                  placeholder="Ghi chú nội bộ cho lịch sử..."
                  className="w-full border border-neutral-300 rounded px-2.5 py-1.5"
                  value={statusForm.note}
                  onChange={(e) => setStatusForm((p) => ({ ...p, note: e.target.value }))}
                />
              </div>

              {/* Show items list */}
              <div className="border border-neutral-200 rounded p-4 space-y-2 bg-neutral-50">
                <p className="text-[10px] font-bold text-neutral-700 uppercase tracking-wide">Chi tiết sản phẩm</p>
                {selectedOrder.items.map((i) => (
                  <div key={i.id} className="flex justify-between text-neutral-800">
                    <span>{i.variant_name} x {i.quantity}</span>
                    <span className="font-mono font-semibold">{i.total_cost.toLocaleString('vi-VN')} đ</span>
                  </div>
                ))}
                <div className="flex justify-between border-t border-neutral-200 pt-1.5 font-bold text-neutral-900">
                  <span>Tổng tiền</span>
                  <span>{selectedOrder.order.total_amount.toLocaleString('vi-VN')} đ</span>
                </div>
              </div>

              {/* Lịch sử cập nhật trạng thái (Timeline) */}
              {selectedOrder.history && selectedOrder.history.length > 0 && (
                <div className="border border-neutral-200 rounded p-4 space-y-3 bg-white">
                  <p className="text-[10px] font-bold text-neutral-700 uppercase tracking-wide">
                    Timeline lịch sử trạng thái
                  </p>
                  <div className="relative border-l border-neutral-200 pl-4 ml-2 space-y-4">
                    {selectedOrder.history.map((h) => (
                      <div key={h.id} className="relative">
                        <span className="absolute -left-[21px] top-1 bg-black rounded-full w-2 h-2 border-2 border-white ring-2 ring-neutral-200"></span>
                        <div className="text-[11px] text-neutral-800 space-y-1">
                          <div className="flex flex-wrap items-center gap-1.5">
                            <span className="font-bold uppercase tracking-wider text-[8px] px-1 py-0.2 bg-neutral-100 rounded text-neutral-500">
                              {h.status_type === 'order' ? 'Đơn hàng' : h.status_type === 'payment' ? 'Thanh toán' : 'Giao hàng'}
                            </span>
                            {h.from_status && (
                              <>
                                <span className="text-neutral-400 font-mono text-[10px]">{h.from_status}</span>
                                <span className="text-neutral-400">→</span>
                              </>
                            )}
                            <span className="font-bold font-mono text-[10px] text-neutral-900">{h.to_status}</span>
                          </div>
                          {h.note && <p className="italic text-neutral-600 bg-neutral-50 px-2 py-1 rounded mt-0.5">{h.note}</p>}
                          <div className="text-[9px] text-neutral-400 flex justify-between">
                            <span>
                              Bởi: <span className="font-semibold text-neutral-600">{h.changed_by_name || 'Hệ thống'}</span>
                            </span>
                            <span>{new Date(h.changed_at).toLocaleString('vi-VN')}</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex gap-2 justify-end pt-3">
                <button
                  type="button"
                  onClick={() => setShowStatusModal(false)}
                  className="text-[10px] border border-neutral-350 px-5 py-2 font-bold uppercase rounded text-neutral-600 hover:text-black"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="text-[10px] bg-black text-white px-6 py-2 font-black uppercase rounded hover:bg-neutral-800"
                >
                  Lưu thay đổi
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
