export interface User {
  id: number
  email: string
  full_name: string
  phone?: string
  gender?: string
  dob?: string
  role: string
  avatar?: string
  is_lock: boolean
  is_verified: boolean
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  full_name: string
  email: string
  password: string
}

export interface GoogleLoginRequest {
  credential: string
}

export interface LoginResponse {
  token: string
  user: User
}
