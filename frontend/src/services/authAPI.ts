import { api } from './api'
import type { LoginRequest, LoginResponse, RegisterRequest, User, GoogleLoginRequest } from '../types'

export const authAPI = {
  login: async (payload: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/login', payload)
    return response.data
  },
  register: async (payload: RegisterRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/register', payload)
    return response.data
  },
  googleAuth: async (payload: GoogleLoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/google', payload)
    return response.data
  },
  getProfile: async (): Promise<User> => {
    const response = await api.get<User>('/profile')
    return response.data
  },
}
