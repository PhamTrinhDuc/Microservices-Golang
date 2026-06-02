import { api } from './api'
import type { OrderResponse, CheckoutOrderRequest, UpdateOrderStatusRequest } from '../types'

export const orderAPI = {
  checkout: async (payload: CheckoutOrderRequest): Promise<OrderResponse> => {
    const res = await api.post<OrderResponse>('/orders/checkout', payload)
    return res.data
  },

  listMyOrders: async (page = 1, limit = 10): Promise<{ data: OrderResponse[]; total: number; page: number; limit: number }> => {
    const res = await api.get<{ data: OrderResponse[]; total: number; page: number; limit: number }>('/orders', {
      params: { page, limit },
    })
    return res.data
  },

  getMyOrderDetails: async (orderId: number): Promise<OrderResponse> => {
    const res = await api.get<OrderResponse>(`/orders/${orderId}`)
    return res.data
  },

  cancelMyOrder: async (orderId: number): Promise<{ message: string }> => {
    const res = await api.post<{ message: string }>(`/orders/${orderId}/cancel`)
    return res.data
  },

  // Admin Orders
  adminListOrders: async (page = 1, limit = 10, storeId?: number): Promise<{ data: OrderResponse[]; total: number; page: number; limit: number }> => {
    const res = await api.get<{ data: OrderResponse[]; total: number; page: number; limit: number }>('/admin/orders', {
      params: { page, limit, store_id: storeId },
    })
    return res.data
  },

  adminUpdateOrderStatus: async (orderId: number, payload: UpdateOrderStatusRequest): Promise<{ message: string }> => {
    const res = await api.put<{ message: string }>(`/admin/orders/${orderId}/status`, payload)
    return res.data
  },
}

