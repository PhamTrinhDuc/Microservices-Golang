import { api } from './api'
import type { User } from '../types'

export const userAPI = {
  adminLockUser: async (userId: number, isLock: boolean): Promise<{ message: string }> => {
    const res = await api.put<{ message: string }>(`/admin/users/${userId}/lock`, { is_lock: isLock })
    return res.data
  },

  adminListUsers: async (
    page = 1,
    limit = 10,
    query?: string
  ): Promise<{
    data: User[]
    pagination: {
      page: number
      limit: number
      total: number
      total_pages: number
    }
  }> => {
    const res = await api.get<{
      data: User[]
      pagination: {
        page: number
        limit: number
        total: number
        total_pages: number
      }
    }>('/admin/users', {
      params: { page, limit, q: query },
    })
    return res.data
  },
}
