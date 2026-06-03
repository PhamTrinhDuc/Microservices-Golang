import { api } from './api'
import type { Address, CreateAddressRequest } from '../types'

export const addressAPI = {
  getAddresses: async (): Promise<Address[]> => {
    const res = await api.get<Address[]>('/addresses')
    return res.data || []
  },

  createAddress: async (payload: CreateAddressRequest): Promise<Address> => {
    const res = await api.post<Address>('/addresses', payload)
    return res.data
  },

  updateAddress: async (addressId: number, payload: CreateAddressRequest): Promise<Address> => {
    const res = await api.put<Address>(`/addresses/${addressId}`, payload)
    return res.data
  },

  setDefaultAddress: async (addressId: number): Promise<{ message: string }> => {
    const res = await api.put<{ message: string }>(`/addresses/${addressId}/default`)
    return res.data
  },

  deleteAddress: async (addressId: number): Promise<void> => {
    await api.delete(`/addresses/${addressId}`)
  },
}
