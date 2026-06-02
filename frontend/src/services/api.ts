import axios from 'axios'
import { tokenManager } from '../utils/tokenManager'
import { sessionManager } from '../utils/sessionManager'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = tokenManager.getToken()
  if (token) {
    config.headers.Authorization = 'Bearer ' + token
  }
  
  config.headers['X-Session-ID'] = sessionManager.getSessionId()
  
  return config
})

const normalizeLogoUrl = (obj: any): any => {
  if (!obj || typeof obj !== 'object') return obj
  if (Array.isArray(obj)) {
    return obj.map(normalizeLogoUrl)
  }
  const newObj: any = {}
  for (const key of Object.keys(obj)) {
    if (key === 'logo_url') {
      newObj['logo'] = obj[key]
    }
    newObj[key] = normalizeLogoUrl(obj[key])
  }
  return newObj
}

api.interceptors.response.use(
  (response) => {
    if (response.data && typeof response.data === 'object') {
      const responseData = response.data as Record<string, any>
      if ('data' in responseData) {
        // If 'total' or 'total_count' are present at the root level, this is a root-level paginated response.
        // We normalize the logo URLs inside 'data' but do not unwrap it, keeping the page, limit, total metadata intact.
        if ('total' in responseData || 'total_count' in responseData) {
          responseData.data = normalizeLogoUrl(responseData.data)
          return response
        }

        const innerData = normalizeLogoUrl(responseData.data)

        // Check for paginated response structure from backend
        if (innerData && typeof innerData === 'object' && !Array.isArray(innerData)) {
          const arrayKey = Object.keys(innerData).find(
            (key) => Array.isArray((innerData as Record<string, any>)[key])
          )

          if (arrayKey && 'total_count' in innerData) {
            const list = (innerData as Record<string, any>)[arrayKey]
            const page = Number(innerData.page || 1)
            const limit = Number(innerData.limit || 10)
            const total = Number(innerData.total_count || 0)
            const totalPages = Math.ceil(total / limit)

            response.data = {
              data: list,
              pagination: {
                page,
                limit,
                total,
                total_pages: totalPages,
              },
            }
            return response
          }
        }

        response.data = innerData
      }
    }
    return response
  },
  (error: any) => {
    if (error?.response?.status === 401) {
      tokenManager.removeToken()
    }

    const message =
      error?.response?.data?.error ??
      error?.response?.data?.message ??
      error?.message ??
      'An unexpected error occurred. Please try again.'

    return Promise.reject(new Error(message))
  },
)
