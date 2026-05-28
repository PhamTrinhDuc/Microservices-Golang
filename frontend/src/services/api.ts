import axios, { AxiosError } from 'axios'
import { tokenManager } from '../utils/tokenManager'

interface APIErrorResponse {
  message?: string
}

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
    config.headers.Authorization = `JWT ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<APIErrorResponse>) => {
    if (error.response?.status === 401) {
      tokenManager.removeToken()
    }

    const message =
      error.response?.data?.message ??
      error.message ??
      'An unexpected error occurred. Please try again.'

    return Promise.reject(new Error(message))
  },
)
