import React, { useEffect, useState } from 'react'
import { bannerAPI } from '../../services/bannerAPI'
import type { Banner } from '../../types'
import ImageUploader from './ImageUploader'

export default function AdminBannersTab() {
  const [banners, setBanners] = useState<Banner[]>([])
  const [loading, setLoading] = useState(false)
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)
  const [bannerForm, setBannerForm] = useState({
    id: 0,
    title: '',
    subtitle: '',
    description: '',
    image_url: '',
    tag: '',
    link_url: '',
    sort_order: 0,
    is_active: true
  })

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const handleDrop = async (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === targetIndex) return

    const updatedBanners = [...banners]
    const [draggedItem] = updatedBanners.splice(draggedIndex, 1)
    updatedBanners.splice(targetIndex, 0, draggedItem)

    // Re-assign sequence orders starting at 1
    const finalBanners = updatedBanners.map((b, idx) => ({
      ...b,
      sort_order: idx + 1
    }))

    setBanners(finalBanners)
    setDraggedIndex(null)

    try {
      setLoading(true)
      for (const banner of finalBanners) {
        const payload = {
          title: banner.title,
          subtitle: banner.subtitle || null,
          description: banner.description || null,
          image_url: banner.image_url,
          tag: banner.tag || null,
          link_url: banner.link_url || null,
          sort_order: banner.sort_order,
          is_active: banner.is_active
        }
        await bannerAPI.adminUpdateBanner(banner.id, payload)
      }
    } catch (err: any) {
      console.error(err)
      alert(err.message || 'Lỗi khi cập nhật thứ tự banner')
      void loadBanners()
    } finally {
      setLoading(false)
    }
  }

  const loadBanners = async () => {
    try {
      setLoading(true)
      const data = await bannerAPI.adminListBanners()
      // Sort initially by sort_order
      setBanners((data || []).sort((a, b) => a.sort_order - b.sort_order))
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadBanners()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!bannerForm.title || !bannerForm.image_url) {
      alert('Vui lòng nhập đầy đủ tiêu đề và hình ảnh banner')
      return
    }

    const payload = {
      title: bannerForm.title,
      subtitle: bannerForm.subtitle || null,
      description: bannerForm.description || null,
      image_url: bannerForm.image_url,
      tag: bannerForm.tag || null,
      link_url: bannerForm.link_url || null,
      sort_order: bannerForm.sort_order,
      is_active: bannerForm.is_active
    } as any

    try {
      setLoading(true)
      if (bannerForm.id > 0) {
        await bannerAPI.adminUpdateBanner(bannerForm.id, payload)
        alert('Cập nhật banner thành công!')
      } else {
        await bannerAPI.adminCreateBanner(payload)
        alert('Tạo banner thành công!')
      }
      setBannerForm({
        id: 0,
        title: '',
        subtitle: '',
        description: '',
        image_url: '',
        tag: '',
        link_url: '',
        sort_order: 0,
        is_active: true
      })
      void loadBanners()
    } catch (err: any) {
      alert(err.message || 'Không thể lưu banner')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (b: Banner) => {
    setBannerForm({
      id: b.id,
      title: b.title,
      subtitle: b.subtitle || '',
      description: b.description || '',
      image_url: b.image_url,
      tag: b.tag || '',
      link_url: b.link_url || '',
      sort_order: b.sort_order,
      is_active: b.is_active
    })
  }

  const handleDelete = async (id: number) => {
    if (!window.confirm('Bạn có chắc chắn muốn xóa banner này?')) return
    try {
      setLoading(true)
      await bannerAPI.adminDeleteBanner(id)
      alert('Xóa banner thành công!')
      void loadBanners()
    } catch (err: any) {
      alert(err.message || 'Không thể xóa banner')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Banners</h1>
        <p className="text-xs text-neutral-555 mt-1">Tạo mới, sắp xếp và kích hoạt các banner hiển thị trên trang chủ</p>
      </div>

      {loading && (
        <div className="flex justify-center py-6">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Creator Form */}
        <form onSubmit={handleSubmit} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs">
          <h3 className="text-xs font-black uppercase tracking-wide">
            {bannerForm.id > 0 ? 'Cập Nhật Banner' : 'Thêm Banner Mới'}
          </h3>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tiêu đề (Title) *</label>
            <input
              type="text"
              placeholder="Ví dụ: LIMIT TIME OFFER..."
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white font-bold"
              value={bannerForm.title}
              onChange={(e) => setBannerForm(p => ({ ...p, title: e.target.value }))}
              required
            />
          </div>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tiêu đề phụ (Subtitle)</label>
            <input
              type="text"
              placeholder="Ví dụ: Giảm giá công nghệ tới 50%..."
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
              value={bannerForm.subtitle}
              onChange={(e) => setBannerForm(p => ({ ...p, subtitle: e.target.value }))}
            />
          </div>

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Mô tả (Description)</label>
            <textarea
              placeholder="Mô tả nội dung banner..."
              rows={2}
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white resize-none"
              value={bannerForm.description}
              onChange={(e) => setBannerForm(p => ({ ...p, description: e.target.value }))}
            />
          </div>

          <ImageUploader
            label="Hình ảnh Banner *"
            value={bannerForm.image_url}
            onChange={(url) => setBannerForm(p => ({ ...p, image_url: url }))}
            placeholder="Kéo thả ảnh banner hoặc nhấp để chọn"
          />

          <div>
            <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Nhãn phân loại (Tag)</label>
            <input
              type="text"
              placeholder="Ví dụ: Fashion..."
              className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
              value={bannerForm.tag}
              onChange={(e) => setBannerForm(p => ({ ...p, tag: e.target.value }))}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Thứ tự hiển thị</label>
              <input
                type="number"
                min={0}
                className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                value={bannerForm.sort_order}
                onChange={(e) => setBannerForm(p => ({ ...p, sort_order: Number(e.target.value) }))}
              />
            </div>
            <div className="flex items-end pb-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={bannerForm.is_active}
                  onChange={(e) => setBannerForm(p => ({ ...p, is_active: e.target.checked }))}
                />
                <span className="text-[10px] font-bold uppercase text-neutral-700">Kích hoạt</span>
              </label>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider py-2.5 rounded transition-colors"
            >
              {bannerForm.id > 0 ? 'Cập Nhật' : 'Tạo Mới'}
            </button>
            {bannerForm.id > 0 && (
              <button
                type="button"
                onClick={() => setBannerForm({
                  id: 0,
                  title: '',
                  subtitle: '',
                  description: '',
                  image_url: '',
                  tag: '',
                  link_url: '',
                  sort_order: 0,
                  is_active: true
                })}
                className="bg-neutral-200 hover:bg-neutral-300 text-neutral-700 text-[10px] font-bold uppercase py-2.5 px-4 rounded transition-colors"
              >
                Hủy
              </button>
            )}
          </div>
        </form>

        {/* Banners List Table */}
        <div className="lg:col-span-2 bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                <th className="p-4 w-10 text-center"></th>
                <th className="p-4">Hình ảnh</th>
                <th className="p-4">Nội dung Banner</th>
                <th className="p-4 text-center">Thứ tự</th>
                <th className="p-4 text-center">Trạng thái</th>
                <th className="p-4 text-center">Hành động</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-150">
              {banners.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-8 text-center text-neutral-400">
                    Chưa có banner nào được tạo.
                  </td>
                </tr>
              ) : (
                banners.map((b, index) => (
                  <tr
                    key={b.id}
                    draggable
                    onDragStart={(e) => {
                      setDraggedIndex(index)
                      e.dataTransfer.effectAllowed = 'move'
                    }}
                    onDragOver={handleDragOver}
                    onDrop={(e) => void handleDrop(e, index)}
                    onDragEnd={() => setDraggedIndex(null)}
                    className={`transition-all ${
                      draggedIndex === index ? 'opacity-40 bg-neutral-100' : 'hover:bg-neutral-50'
                    }`}
                  >
                    <td className="p-4 text-center cursor-grab active:cursor-grabbing text-neutral-450">
                      <svg className="w-4 h-4 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 8h16M4 16h16" />
                      </svg>
                    </td>
                    <td className="p-4 w-24">
                      <img
                        src={b.image_url}
                        alt={b.title}
                        className="w-16 h-10 object-cover rounded border border-neutral-100"
                      />
                    </td>
                    <td className="p-4">
                      <div>
                        {b.tag && (
                          <span className="bg-neutral-100 text-neutral-600 px-1.5 py-0.5 rounded text-[8px] font-bold uppercase">
                            {b.tag}
                          </span>
                        )}
                        <h4 className="font-extrabold text-neutral-900 mt-1">{b.title}</h4>
                        {b.subtitle && <p className="text-neutral-500 text-[10px] mt-0.5">{b.subtitle}</p>}

                      </div>
                    </td>
                    <td className="p-4 text-center font-mono font-bold text-neutral-600">
                      {b.sort_order}
                    </td>
                    <td className="p-4 text-center">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase ${b.is_active ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                        {b.is_active ? 'Hoạt động' : 'Tạm ngưng'}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <div className="flex gap-2 justify-center">
                        <button
                          type="button"
                          onClick={() => handleEdit(b)}
                          className="text-neutral-800 hover:text-black font-bold uppercase text-[9px] tracking-wider"
                        >
                          Sửa
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(b.id)}
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
        </div>
      </div>
    </div>
  )
}
