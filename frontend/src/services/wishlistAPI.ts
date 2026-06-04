import { api } from './api'
import type { WishlistItemResponse } from '../types'

export const wishlistAPI = {
  getWishlist: async (): Promise<WishlistItemResponse[]> => {
    const res = await api.get<WishlistItemResponse[]>('/wishlist')
    return res.data || []
  },

  addToWishlist: async (variantId: number): Promise<any> => {
    const res = await api.post<any>('/wishlist', { variant_id: variantId })
    return res.data
  },

  removeFromWishlist: async (variantId: number): Promise<void> => {
    await api.delete(`/wishlist/${variantId}`)
  },
}
