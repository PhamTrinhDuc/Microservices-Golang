import { api } from './api'
import type { CartItem, AddToCartRequest } from '../types'

export const cartAPI = {
  getCart: async (): Promise<CartItem[]> => {
    const res = await api.get<CartItem[]>('/cart')
    return res.data || []
  },

  addToCart: async (payload: AddToCartRequest): Promise<any> => {
    const res = await api.post<any>('/cart', payload)
    return res.data
  },

  updateItemQty: async (itemId: number, quantity: number): Promise<any> => {
    const res = await api.put<any>(`/cart/items/${itemId}`, { quantity })
    return res.data
  },

  removeItem: async (itemId: number): Promise<void> => {
    await api.delete(`/cart/items/${itemId}`)
  },

  clearCart: async (): Promise<void> => {
    await api.delete('/cart')
  },

  mergeCart: async (sessionId: string): Promise<any> => {
    const res = await api.post<any>('/cart/merge', { session_id: sessionId })
    return res.data
  },
}
