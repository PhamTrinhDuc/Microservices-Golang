import { api } from './api'
import type { Voucher, ApplyVoucherRequest, ApplyVoucherResponse, CreateVoucherRequest, UpdateVoucherRequest, CreatePromotionRequest, Promotion } from '../types'

export const voucherAPI = {
  getVouchers: async (): Promise<Voucher[]> => {
    const res = await api.get<Voucher[]>('/vouchers')
    return res.data || []
  },

  applyVoucher: async (payload: ApplyVoucherRequest): Promise<ApplyVoucherResponse> => {
    const res = await api.post<ApplyVoucherResponse>('/vouchers/apply', payload)
    return res.data
  },

  // Admin Vouchers
  adminListVouchers: async (): Promise<Voucher[]> => {
    const res = await api.get<Voucher[]>('/admin/vouchers')
    return res.data || []
  },

  adminCreateVoucher: async (payload: CreateVoucherRequest): Promise<Voucher> => {
    const res = await api.post<Voucher>('/admin/vouchers', payload)
    return res.data
  },

  adminUpdateVoucher: async (id: number, payload: UpdateVoucherRequest): Promise<Voucher> => {
    const res = await api.put<Voucher>(`/admin/vouchers/${id}`, payload)
    return res.data
  },

  adminDeleteVoucher: async (id: number): Promise<void> => {
    await api.delete(`/admin/vouchers/${id}`)
  },

  // Admin Promotions
  adminListPromotions: async (): Promise<Promotion[]> => {
    const res = await api.get<Promotion[]>('/admin/promotions')
    return res.data || []
  },

  adminCreatePromotion: async (payload: CreatePromotionRequest): Promise<Promotion> => {
    const res = await api.post<Promotion>('/admin/promotions', payload)
    return res.data
  },

  adminDeletePromotion: async (id: number): Promise<void> => {
    await api.delete(`/admin/promotions/${id}`)
  },
}

