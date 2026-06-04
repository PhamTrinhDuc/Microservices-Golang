import { api } from './api'
import type { AnalyticsSummary } from '../types'

export interface GetAnalyticsParams {
  start_date?: string
  end_date?: string
  store_id?: number
}

export const analyticsAPI = {
  getAnalytics: async (params: GetAnalyticsParams): Promise<AnalyticsSummary> => {
    const queryParams = new URLSearchParams()
    if (params.start_date) queryParams.append('start_date', params.start_date)
    if (params.end_date) queryParams.append('end_date', params.end_date)
    if (params.store_id !== undefined) queryParams.append('store_id', params.store_id.toString())

    const url = `/admin/analytics?${queryParams.toString()}`
    const res = await api.get<AnalyticsSummary>(url)
    return res.data
  },
}
