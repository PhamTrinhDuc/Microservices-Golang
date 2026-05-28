import { api } from './api'
import type { LoginRequest, LoginResponse, RegisterRequest, User } from '../types'

export const authAPI = {
  login: async (payload: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/login', payload)
    return response.data
  },
  register: async (payload: RegisterRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/register', payload)
    return response.data
  },
  getProfile: async (): Promise<User> => {
    const response = await api.get<User>('/auth/profile')
    return response.data
  },
}
