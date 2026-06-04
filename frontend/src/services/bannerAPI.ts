import { api } from './api'
import type { Banner } from '../types'

export const bannerAPI = {
  listBanners: async (params?: { category_id?: number }): Promise<Banner[]> => {
    const res = await api.get<Banner[]>('/banners', { params })
    return res.data || []
  },

  adminListBanners: async (): Promise<Banner[]> => {
    const res = await api.get<Banner[]>('/admin/banners')
    return res.data || []
  },

  adminCreateBanner: async (payload: Omit<Banner, 'id' | 'created_at' | 'updated_at'>): Promise<Banner> => {
    const res = await api.post<Banner>('/admin/banners', payload)
    return res.data
  },

  adminUpdateBanner: async (id: number, payload: Omit<Banner, 'id' | 'created_at' | 'updated_at'>): Promise<Banner> => {
    const res = await api.put<Banner>(`/admin/banners/${id}`, payload)
    return res.data
  },

  adminDeleteBanner: async (id: number): Promise<void> => {
    await api.delete(`/admin/banners/${id}`)
  }
}
