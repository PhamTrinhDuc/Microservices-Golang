import React, { useEffect, useState } from 'react'
import { policyAPI } from '../../services/policyAPI'
import type { Policy } from '../../types'

export default function AdminPoliciesTab() {
  const [policies, setPolicies] = useState<Policy[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [policyForm, setPolicyForm] = useState({
    id: '',
    title: '',
    slug: '',
    content: '',
    category: 'refund',
    is_active: true
  })

  // Predefined categories for standard ecommerce policies
  const categories = [
    { value: 'refund', label: 'Chính sách Đổi trả' },
    { value: 'shipping', label: 'Chính sách Vận chuyển' },
    { value: 'privacy', label: 'Chính sách Bảo mật' },
    { value: 'terms', label: 'Điều khoản Dịch vụ' },
    { value: 'payment', label: 'Chính sách Thanh toán' }
  ]

  const loadPolicies = async () => {
    try {
      setLoading(true)
      const res = await policyAPI.listPolicies({ limit: 100 })
      setPolicies(res.policies || [])
      setTotal(res.total || 0)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadPolicies()
  }, [])

  // Auto-generate slug from title
  const handleTitleChange = (val: string) => {
    const generatedSlug = val
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '') // remove Vietnamese signs
      .replace(/[đĐ]/g, 'd')
      .replace(/[^a-z0-9\s-]/g, '') // remove special characters
      .trim()
      .replace(/\s+/g, '-') // replace spaces with hyphens
      .replace(/-+/g, '-')

    setPolicyForm((prev) => ({
      ...prev,
      title: val,
      slug: prev.id ? prev.slug : generatedSlug // Only auto-fill slug if creating a new policy
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!policyForm.title || !policyForm.slug || !policyForm.content) {
      alert('Vui lòng nhập đầy đủ tiêu đề, slug và nội dung chính sách')
      return
    }

    const payload = {
      title: policyForm.title,
      slug: policyForm.slug,
      content: policyForm.content,
      category: policyForm.category,
      is_active: policyForm.is_active
    }

    try {
      setLoading(true)
      if (policyForm.id) {
        await policyAPI.adminUpdatePolicy(policyForm.id, payload)
        alert('Cập nhật chính sách thành công!')
      } else {
        await policyAPI.adminCreatePolicy(payload)
        alert('Tạo chính sách thành công!')
      }
      resetForm()
      void loadPolicies()
    } catch (err: any) {
      alert(err.message || 'Không thể lưu chính sách')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (p: Policy) => {
    setPolicyForm({
      id: p.id,
      title: p.title,
      slug: p.slug,
      content: p.content,
      category: p.category,
      is_active: p.is_active
    })
  }

  const handleDelete = async (id: string) => {
    if (!window.confirm('Bạn có chắc chắn muốn xóa chính sách này? Tất cả các vector chunks phục vụ RAG liên quan cũng sẽ bị xóa.')) return
    try {
      setLoading(true)
      await policyAPI.adminDeletePolicy(id)
      alert('Xóa chính sách thành công!')
      resetForm()
      void loadPolicies()
    } catch (err: any) {
      alert(err.message || 'Không thể xóa chính sách')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setPolicyForm({
      id: '',
      title: '',
      slug: '',
      content: '',
      category: 'refund',
      is_active: true
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý chính sách & RAG KB</h1>
        <p className="text-xs text-neutral-555 mt-1">
          Soạn thảo chính sách hiển thị trên Web và tự động đồng bộ hóa cơ sở kiến thức RAG cho Chatbot
        </p>
      </div>

      {loading && (
        <div className="flex justify-center py-6">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
        {/* Editor Form */}
        <form onSubmit={handleSubmit} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs shadow-sm">
          <h3 className="text-xs font-black uppercase tracking-wide border-b pb-2 border-neutral-100">
            {policyForm.id ? 'Cập Nhật Chính Sách' : 'Soạn Thảo Chính Sách Mới'}
          </h3>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tiêu đề (Title) *</label>
            <input
              type="text"
              placeholder="Ví dụ: Chính sách vận chuyển và giao nhận"
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white font-bold"
              value={policyForm.title}
              onChange={(e) => handleTitleChange(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Đường dẫn tĩnh (Slug) *</label>
            <input
              type="text"
              placeholder="Ví dụ: chinh-sach-van-chuyen"
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white font-mono"
              value={policyForm.slug}
              onChange={(e) => setPolicyForm(p => ({ ...p, slug: e.target.value }))}
              required
            />
            <p className="text-[9px] text-neutral-400 mt-0.5">Sử dụng cho URL trang web (ví dụ: /policies/chinh-sach-van-chuyen)</p>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Phân loại (Category) *</label>
              <select
                className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                value={policyForm.category}
                onChange={(e) => setPolicyForm(p => ({ ...p, category: e.target.value }))}
              >
                {categories.map((c) => (
                  <option key={c.value} value={c.value}>
                    {c.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex items-end pb-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={policyForm.is_active}
                  onChange={(e) => setPolicyForm(p => ({ ...p, is_active: e.target.checked }))}
                />
                <span className="text-[10px] font-bold uppercase text-neutral-700">Kích hoạt hiển thị</span>
              </label>
            </div>
          </div>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Nội dung chi tiết *</label>
            <textarea
              placeholder="Soạn thảo nội dung chính sách tại đây. Hỗ trợ định dạng văn bản thô hoặc Markdown..."
              rows={12}
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white resize-none font-sans leading-relaxed"
              value={policyForm.content}
              onChange={(e) => setPolicyForm(p => ({ ...p, content: e.target.value }))}
              required
            />
            <p className="text-[9px] text-neutral-400 mt-0.5">
              Nội dung này sẽ tự động được chia nhỏ và cập nhật làm Cơ sở kiến thức cho Chatbot.
            </p>
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider py-2.5 rounded transition-colors"
            >
              {policyForm.id ? 'Cập Nhật & Sync RAG' : 'Tạo & Sync RAG'}
            </button>
            {policyForm.id && (
              <button
                type="button"
                onClick={resetForm}
                className="bg-neutral-200 hover:bg-neutral-300 text-neutral-700 text-[10px] font-bold uppercase py-2.5 px-4 rounded transition-colors"
              >
                Hủy
              </button>
            )}
          </div>
        </form>

        {/* Policies List Table */}
        <div className="xl:col-span-2 bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                <th className="p-4">Chính sách</th>
                <th className="p-4">Đường dẫn</th>
                <th className="p-4 text-center">Phân loại</th>
                <th className="p-4 text-center">Hiển thị</th>
                <th className="p-4 text-center">Hành động</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-150">
              {policies.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-neutral-400">
                    Chưa có chính sách nào được soạn thảo.
                  </td>
                </tr>
              ) : (
                policies.map((p) => (
                  <tr key={p.id} className="hover:bg-neutral-50 transition-colors">
                    <td className="p-4">
                      <div className="font-extrabold text-neutral-900">{p.title}</div>
                    </td>
                    <td className="p-4 font-mono text-neutral-500 text-[10px]">
                      /policies/{p.slug}
                    </td>
                    <td className="p-4 text-center">
                      <span className="bg-indigo-50 border border-indigo-100 text-indigo-700 px-2 py-0.5 rounded text-[9px] font-bold uppercase">
                        {categories.find(c => c.value === p.category)?.label || p.category}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase ${p.is_active ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                        {p.is_active ? 'Hoạt động' : 'Tạm ngưng'}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <div className="flex gap-3 justify-center">
                        <button
                          type="button"
                          onClick={() => handleEdit(p)}
                          className="text-neutral-800 hover:text-black font-bold uppercase text-[9px] tracking-wider"
                        >
                          Sửa
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(p.id)}
                          className="text-red-500 hover:text-red-700 font-bold uppercase text-[9px] tracking-wider"
                        >
                          Xóa
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
          <div className="bg-neutral-50 px-4 py-3 border-t border-neutral-150 flex justify-between items-center text-[10px] text-neutral-555 font-mono">
            <span>TỔNG SỐ CHÍNH SÁCH: {total}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
