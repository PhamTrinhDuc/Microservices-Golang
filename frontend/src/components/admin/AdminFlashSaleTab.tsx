import React, { useEffect, useState } from 'react'
import { voucherAPI } from '../../services/voucherAPI'
import { productAPI } from '../../services/productAPI'
import type { Promotion, Product, ProductVariant } from '../../types'

export default function AdminFlashSaleTab() {
  const [promotions, setPromotions] = useState<Promotion[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [selectedProductDetails, setSelectedProductDetails] = useState<Product | null>(null)
  const [loading, setLoading] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')

  // Form states
  const [showForm, setShowForm] = useState(false)
  const [editingPromoId, setEditingPromoId] = useState<number | null>(null)
  
  const [formData, setFormData] = useState({
    productId: '',
    variantId: '',
    name: '',
    description: '',
    discountType: 'percentage' as 'percentage' | 'fixed',
    discountValue: 0,
    startDate: '',
    endDate: '',
    isActive: true,
  })

  const loadData = async () => {
    try {
      setLoading(true)
      const pList = await voucherAPI.adminListPromotions()
      setPromotions(pList)
      
      // Load products for lookup
      const prodRes = await productAPI.getProducts({ limit: 100 })
      setProducts(prodRes.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadData()
  }, [])

  // When a product is chosen, fetch detailed info (including variants)
  useEffect(() => {
    const fetchVariants = async () => {
      if (!formData.productId) {
        setSelectedProductDetails(null)
        return
      }
      try {
        const details = await productAPI.getProductById(formData.productId)
        setSelectedProductDetails(details)
        
        // Auto-fill name if empty
        if (!formData.name || formData.name.startsWith('Flash Sale')) {
          setFormData(prev => ({
            ...prev,
            name: `Flash Sale ${details.name.slice(0, 30)}`
          }))
        }
      } catch (err) {
        console.error('Failed to load product variants:', err)
      }
    }
    void fetchVariants()
  }, [formData.productId])

  const handleOpenCreate = () => {
    setEditingPromoId(null)
    setFormData({
      productId: '',
      variantId: '',
      name: '',
      description: '',
      discountType: 'percentage',
      discountValue: 0,
      startDate: '',
      endDate: '',
      isActive: true,
    })
    setShowForm(true)
  }

  const handleOpenEdit = async (promo: Promotion) => {
    setEditingPromoId(promo.id)
    
    // Format dates for input field (YYYY-MM-DDThh:mm)
    const fmtDate = (dStr: string) => {
      const d = new Date(dStr)
      const pad = (n: number) => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
    }

    setFormData({
      productId: promo.product_id,
      variantId: promo.variant_id ? String(promo.variant_id) : '',
      name: promo.name,
      description: promo.description || '',
      discountType: promo.discount_type,
      discountValue: promo.discount_value,
      startDate: fmtDate(promo.start_date),
      endDate: fmtDate(promo.end_date),
      isActive: promo.is_active,
    })
    setShowForm(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.productId || !formData.startDate || !formData.endDate) {
      alert('Vui lòng điền đầy đủ các thông tin bắt buộc (*).')
      return
    }

    try {
      setLoading(true)
      const payload = {
        name: formData.name.trim(),
        description: formData.description ? formData.description.trim() : null,
        discount_type: formData.discountType,
        discount_value: Number(formData.discountValue),
        start_date: new Date(formData.startDate).toISOString(),
        end_date: new Date(formData.endDate).toISOString(),
        is_active: formData.isActive,
      }

      if (editingPromoId !== null) {
        // Update existing promotion
        await voucherAPI.adminUpdatePromotion(editingPromoId, payload)
        alert('Cập nhật khuyến mãi thành công!')
      } else {
        // Create new promotion
        const createPayload = {
          ...payload,
          product_id: formData.productId,
          variant_id: formData.variantId ? Number(formData.variantId) : null,
        }
        await voucherAPI.adminCreatePromotion(createPayload)
        alert('Tạo chương trình Flash Sale thành công!')
      }

      setShowForm(false)
      void loadData()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xử lý khuyến mãi')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Bạn chắc chắn muốn xóa chương trình Flash Sale này?')) return
    try {
      setLoading(true)
      await voucherAPI.adminDeletePromotion(id)
      alert('Đã xóa chương trình thành công!')
      void loadData()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xóa khuyến mãi')
    } finally {
      setLoading(false)
    }
  }

  const getPromoStatus = (promo: Promotion) => {
    if (!promo.is_active) return { label: 'Tạm khóa', css: 'bg-neutral-100 text-neutral-500 border border-neutral-300' }
    
    const now = new Date()
    const start = new Date(promo.start_date)
    const end = new Date(promo.end_date)
    
    if (now < start) {
      return { label: 'Sắp diễn ra', css: 'bg-blue-50 text-blue-600 border border-blue-200' }
    }
    if (now > end) {
      return { label: 'Đã hết hạn', css: 'bg-red-50 text-red-600 border border-red-200' }
    }
    return { label: 'Đang hoạt động', css: 'bg-green-50 text-green-600 border border-green-200 animate-pulse' }
  }

  // Filter products for select dropdown based on search
  const filteredProducts = products.filter(p => 
    p.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
    p.id.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Flash Sale & Khuyến Mãi</h1>
          <p className="text-xs text-neutral-500 mt-1">
            Thiết lập các chương trình giảm giá sốc và hiển thị đếm ngược trên trang chủ.
          </p>
        </div>
        <button
          onClick={showForm ? () => setShowForm(false) : handleOpenCreate}
          className="bg-black hover:bg-neutral-800 text-white text-xs font-black uppercase tracking-wider px-4 py-2.5 rounded transition-colors shadow-sm"
        >
          {showForm ? 'Đóng Form' : '+ Thiết lập Flash Sale'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="bg-white border border-neutral-200 rounded-lg p-6 space-y-5 text-xs shadow-sm">
          <h2 className="text-sm font-black uppercase tracking-wide border-b border-neutral-100 pb-2">
            {editingPromoId ? 'Cập nhật chương trình Flash Sale' : 'Tạo chương trình Flash Sale mới'}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Product Selector */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Chọn sản phẩm khuyến mãi *</label>
              {editingPromoId ? (
                <input
                  type="text"
                  className="border border-neutral-200 bg-neutral-50 text-neutral-500 rounded px-3 py-2 w-full font-bold"
                  value={products.find(p => p.id === formData.productId)?.name || formData.productId}
                  disabled
                />
              ) : (
                <div className="space-y-2">
                  <input
                    type="text"
                    placeholder="Tìm theo tên hoặc ID sản phẩm..."
                    className="border border-neutral-350 rounded px-3 py-2 w-full"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                  <select
                    className="border border-neutral-350 rounded px-3 py-2 w-full max-h-32"
                    value={formData.productId}
                    onChange={(e) => setFormData(p => ({ ...p, productId: e.target.value, variantId: '' }))}
                    required
                  >
                    <option value="">-- Chọn sản phẩm --</option>
                    {filteredProducts.map(p => (
                      <option key={p.id} value={p.id}>
                        {p.name} (Giá: {p.price.toLocaleString('vi-VN')} đ)
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {/* Variant Selector */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Chọn phiên bản / biến thể</label>
              <select
                className="border border-neutral-350 rounded px-3 py-2 w-full disabled:bg-neutral-50 disabled:text-neutral-400"
                value={formData.variantId}
                onChange={(e) => setFormData(p => ({ ...p, variantId: e.target.value }))}
                disabled={editingPromoId !== null || !formData.productId || !selectedProductDetails?.variants?.length}
              >
                <option value="">Áp dụng cho tất cả phiên bản</option>
                {selectedProductDetails?.variants?.map((v: ProductVariant) => (
                  <option key={v.id} value={v.id}>
                    {v.name} (Giá: {v.sell_price.toLocaleString('vi-VN')} đ)
                  </option>
                ))}
              </select>
              {!editingPromoId && formData.productId && (!selectedProductDetails?.variants || selectedProductDetails.variants.length === 0) && (
                <span className="text-[10px] text-neutral-450">Sản phẩm này chưa có cấu hình biến thể.</span>
              )}
            </div>

            {/* Promotion Campaign Name */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Tên chiến dịch Flash Sale *</label>
              <input
                type="text"
                placeholder="Ví dụ: Giảm giá sâu ngày hè"
                className="border border-neutral-355 rounded px-3 py-2 w-full"
                value={formData.name}
                onChange={(e) => setFormData(p => ({ ...p, name: e.target.value }))}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Discount Type */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Hình thức giảm giá *</label>
              <select
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={formData.discountType}
                onChange={(e) => setFormData(p => ({ ...p, discountType: e.target.value as 'percentage' | 'fixed' }))}
              >
                <option value="percentage">Theo phần trăm (%)</option>
                <option value="fixed">Số tiền trực tiếp (VND)</option>
              </select>
            </div>

            {/* Discount Value */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Giá trị giảm giá *</label>
              <input
                type="number"
                placeholder={formData.discountType === 'percentage' ? 'Ví dụ: 20' : 'Ví dụ: 100000'}
                className="border border-neutral-355 rounded px-3 py-2 w-full font-bold"
                value={formData.discountValue || ''}
                onChange={(e) => setFormData(p => ({ ...p, discountValue: Number(e.target.value) }))}
                required
              />
            </div>

            {/* Start Date */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Thời gian bắt đầu *</label>
              <input
                type="datetime-local"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={formData.startDate}
                onChange={(e) => setFormData(p => ({ ...p, startDate: e.target.value }))}
                required
              />
            </div>

            {/* End Date */}
            <div className="flex flex-col gap-1.5">
              <label className="font-bold text-neutral-700">Thời gian kết thúc *</label>
              <input
                type="datetime-local"
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={formData.endDate}
                onChange={(e) => setFormData(p => ({ ...p, endDate: e.target.value }))}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-center">
            {/* Description */}
            <div className="flex flex-col gap-1.5 md:col-span-2">
              <label className="font-bold text-neutral-700">Mô tả chi tiết</label>
              <input
                type="text"
                placeholder="Nhập mô tả cho chương trình khuyến mãi này..."
                className="border border-neutral-350 rounded px-3 py-2 w-full"
                value={formData.description}
                onChange={(e) => setFormData(p => ({ ...p, description: e.target.value }))}
              />
            </div>

            {/* Active Status Checkbox */}
            <div className="flex items-center gap-2 mt-4">
              <input
                type="checkbox"
                id="isActive"
                className="w-4 h-4 rounded text-black border-neutral-300 focus:ring-black"
                checked={formData.isActive}
                onChange={(e) => setFormData(p => ({ ...p, isActive: e.target.checked }))}
              />
              <label htmlFor="isActive" className="font-bold text-neutral-800 cursor-pointer">
                Kích hoạt chương trình ngay lập tức
              </label>
            </div>
          </div>

          <div className="flex justify-end gap-2 border-t border-neutral-100 pt-4">
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="border border-neutral-250 text-neutral-700 px-5 py-2 rounded hover:bg-neutral-50 transition-colors"
            >
              Hủy
            </button>
            <button
              type="submit"
              disabled={loading}
              className="bg-black hover:bg-neutral-800 text-white font-bold px-6 py-2 rounded disabled:opacity-40 shadow-sm"
            >
              {editingPromoId ? 'Cập nhật' : 'Tạo Khuyến Mãi'}
            </button>
          </div>
        </form>
      )}

      {/* Promotions Table */}
      <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
        <div className="p-4 border-b border-neutral-100 bg-neutral-50/50">
          <h3 className="font-bold text-neutral-800 text-sm">Danh sách chương trình Flash Sale ({promotions.length})</h3>
        </div>
        
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="bg-neutral-50/30 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
              <th className="p-4">Tên chương trình / Sản phẩm</th>
              <th className="p-4">Hình thức giảm giá</th>
              <th className="p-4">Thời gian diễn ra</th>
              <th className="p-4 text-center">Trạng thái</th>
              <th className="p-4 text-center">Thao tác</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-150">
            {promotions.map((p) => {
              const status = getPromoStatus(p)
              const matchedProduct = products.find(prod => prod.id === p.product_id)
              
              return (
                <tr key={p.id} className="hover:bg-neutral-50/50 transition-colors">
                  <td className="p-4">
                    <div className="flex items-center gap-3">
                      {matchedProduct?.image && (
                        <img 
                          src={matchedProduct.image} 
                          alt={matchedProduct.name} 
                          className="w-10 h-10 object-contain rounded border border-neutral-200 p-0.5 bg-white shrink-0"
                        />
                      )}
                      <div>
                        <p className="font-bold text-neutral-855 leading-tight">{p.name}</p>
                        <p className="text-[10px] text-neutral-450 mt-1">
                          Sản phẩm: <span className="font-semibold text-neutral-700">{matchedProduct?.name || p.product_id}</span>
                          {p.variant_id && <span className="text-neutral-500"> (Biến thể ID: {p.variant_id})</span>}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="p-4">
                    <span className="font-bold text-neutral-900">
                      {p.discount_type === 'percentage' 
                        ? `-${p.discount_value}%` 
                        : `-${p.discount_value.toLocaleString('vi-VN')} đ`
                      }
                    </span>
                    {matchedProduct?.price && (
                      <p className="text-[9px] text-neutral-400 mt-0.5">
                        Giá gốc: {matchedProduct.price.toLocaleString('vi-VN')} đ
                      </p>
                    )}
                  </td>
                  <td className="p-4 text-neutral-600 font-medium">
                    <div className="space-y-0.5">
                      <p><span className="text-neutral-400">Bắt đầu:</span> {new Date(p.start_date).toLocaleString('vi-VN')}</p>
                      <p><span className="text-neutral-400">Kết thúc:</span> {new Date(p.end_date).toLocaleString('vi-VN')}</p>
                    </div>
                  </td>
                  <td className="p-4 text-center">
                    <span className={`px-2 py-0.5 rounded-full text-[9px] font-black uppercase tracking-wider ${status.css}`}>
                      {status.label}
                    </span>
                  </td>
                  <td className="p-4">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => handleOpenEdit(p)}
                        className="text-neutral-400 hover:text-black hover:bg-neutral-100 p-2 rounded transition-all"
                        title="Chỉnh sửa"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                      <button
                        onClick={() => handleDelete(p.id)}
                        className="text-neutral-400 hover:text-red-655 hover:bg-red-50 p-2 rounded transition-all"
                        title="Xóa"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              )
            })}
            {promotions.length === 0 && (
              <tr>
                <td colSpan={5} className="p-8 text-center text-neutral-400">
                  Chưa có chương trình Flash Sale nào được thiết lập.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
