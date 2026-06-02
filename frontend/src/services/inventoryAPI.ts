import { api } from './api'
import type { Store, Supplier, ProductInventory, ImportInvoiceResponse, ImportInvoiceDetailsResponse, LowStockAlertResponse, InventoryLogResponse } from '../types'

export const inventoryAPI = {
  // Stores CRUD (public / admin)
  listStores: async (): Promise<Store[]> => {
    const res = await api.get<Store[]>('/stores')
    return res.data || []
  },

  adminCreateStore: async (payload: {
    name: string
    hotline?: string | null
    district: string
    province: string
    ward: string
    road?: string | null
    email?: string | null
    lat?: number | null
    lng?: number | null
  }): Promise<Store> => {
    const res = await api.post<Store>('/admin/stores', payload)
    return res.data
  },

  adminUpdateStore: async (id: number, payload: {
    name: string
    hotline?: string | null
    district: string
    province: string
    ward: string
    road?: string | null
    email?: string | null
    lat?: number | null
    lng?: number | null
    is_active: boolean
  }): Promise<Store> => {
    const res = await api.put<Store>(`/admin/stores/${id}`, payload)
    return res.data
  },

  adminDeactivateStore: async (id: number): Promise<void> => {
    await api.delete(`/admin/stores/${id}`)
  },

  // Suppliers CRUD
  adminListSuppliers: async (): Promise<Supplier[]> => {
    const res = await api.get<Supplier[]>('/admin/suppliers')
    return res.data || []
  },

  adminCreateSupplier: async (payload: { name: string; address?: string | null; phone?: string | null }): Promise<Supplier> => {
    const res = await api.post<Supplier>('/admin/suppliers', payload)
    return res.data
  },

  adminUpdateSupplier: async (id: number, payload: { name: string; address?: string | null; phone?: string | null }): Promise<Supplier> => {
    const res = await api.put<Supplier>(`/admin/suppliers/${id}`, payload)
    return res.data
  },

  adminDeleteSupplier: async (id: number): Promise<void> => {
    await api.delete(`/admin/suppliers/${id}`)
  },

  // Inventory Management
  adminListStoreInventory: async (storeId: number): Promise<ProductInventory[]> => {
    const res = await api.get<ProductInventory[]>(`/admin/stores/${storeId}/inventory`)
    return res.data || []
  },

  adminAdjustInventory: async (storeId: number, payload: { adjustments: { variant_id: number; new_quantity: number }[] }): Promise<void> => {
    await api.put(`/admin/stores/${storeId}/inventory`, payload)
  },

  // Goods Import Invoices
  adminImportGoods: async (payload: {
    supplier_id: number
    store_id: number
    note?: string | null
    items: { variant_id: number; quantity: number; price_import: number }[]
  }): Promise<any> => {
    const res = await api.post<any>('/admin/inventory/import', payload)
    return res.data
  },

  adminListImportInvoices: async (storeId?: number, page = 1, limit = 10): Promise<{ data: ImportInvoiceResponse[]; total: number; page: number; limit: number }> => {
    const params: any = { page, limit }
    if (storeId) {
      params.store_id = storeId
    }
    const res = await api.get<{ data: ImportInvoiceResponse[]; total: number; page: number; limit: number }>('/admin/inventory/imports', {
      params,
    })
    return res.data
  },

  adminGetImportInvoiceDetails: async (invoiceId: number): Promise<ImportInvoiceDetailsResponse> => {
    const res = await api.get<ImportInvoiceDetailsResponse>(`/admin/inventory/imports/${invoiceId}`)
    return res.data
  },

  // Alerts & Logs
  adminGetLowStockAlerts: async (storeId?: number): Promise<LowStockAlertResponse[]> => {
    const params: any = {}
    if (storeId) {
      params.store_id = storeId
    }
    const res = await api.get<LowStockAlertResponse[]>('/admin/inventory/low-stock', { params })
    return res.data || []
  },

  adminGetInventoryLogs: async (params?: { store_id?: number; variant_id?: number; reason?: string; page?: number; limit?: number }): Promise<{ logs: InventoryLogResponse[]; total_count: number; page: number; limit: number }> => {
    const res = await api.get<{ logs: InventoryLogResponse[]; total_count: number; page: number; limit: number }>('/admin/inventory/logs', { params })
    return res.data
  },
}
