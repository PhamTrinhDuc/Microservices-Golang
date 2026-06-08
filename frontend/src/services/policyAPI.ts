import { api } from './api'
import type { Policy, PolicySearchResult } from '../types'

export interface ListPoliciesResponse {
  policies: Policy[]
  total: number
  limit: number
  offset: number
}

export const policyAPI = {
  listPolicies: async (params?: { category?: string; limit?: number; offset?: number }): Promise<ListPoliciesResponse> => {
    const res = await api.get<ListPoliciesResponse>('/policies', { params })
    return res.data
  },

  getPolicyBySlug: async (slug: string): Promise<Policy> => {
    const res = await api.get<Policy>(`/policies/${slug}`)
    return res.data
  },

  adminCreatePolicy: async (payload: Omit<Policy, 'id' | 'created_at' | 'updated_at'>): Promise<Policy> => {
    const res = await api.post<Policy>('/admin/policies', payload)
    return res.data
  },

  adminUpdatePolicy: async (id: string, payload: Omit<Policy, 'id' | 'created_at' | 'updated_at'>): Promise<Policy> => {
    const res = await api.put<Policy>(`/admin/policies/${id}`, payload)
    return res.data
  },

  adminDeletePolicy: async (id: string): Promise<void> => {
    await api.delete(`/admin/policies/${id}`)
  },

  askChatbot: async (query: string, limit?: number): Promise<PolicySearchResult[]> => {
    const res = await api.post<PolicySearchResult[]>('/policies/chat', { query, limit })
    return res.data || []
  }
}
