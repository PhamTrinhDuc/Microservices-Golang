import { createAsyncThunk, createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { authAPI } from '../../services/authAPI'
import type { LoginRequest, LoginResponse, RegisterRequest, User, GoogleLoginRequest } from '../../types'
import { tokenManager } from '../../utils/tokenManager'
import { fetchCart, mergeCart, clearCartState } from './cartSlice'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  googleLoading: boolean
  error: string | null
}

const initialState: AuthState = {
  user: null,
  token: tokenManager.getToken(),
  isAuthenticated: tokenManager.isAuthenticated(),
  loading: false,
  googleLoading: false,
  error: null,
}

export const login = createAsyncThunk<LoginResponse, LoginRequest, { rejectValue: string }>(
  'auth/login',
  async (payload, { dispatch, rejectWithValue }) => {
    try {
      const res = await authAPI.login(payload)
      const guestSession = localStorage.getItem('guest_session_id')
      if (guestSession) {
        void dispatch(mergeCart(guestSession))
      } else {
        void dispatch(fetchCart())
      }
      return res
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Login failed')
    }
  },
)

export const register = createAsyncThunk<LoginResponse, RegisterRequest, { rejectValue: string }>(
  'auth/register',
  async (payload, { dispatch, rejectWithValue }) => {
    try {
      const res = await authAPI.register(payload)
      const guestSession = localStorage.getItem('guest_session_id')
      if (guestSession) {
        void dispatch(mergeCart(guestSession))
      } else {
        void dispatch(fetchCart())
      }
      return res
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Registration failed')
    }
  },
)

export const googleAuth = createAsyncThunk<LoginResponse, GoogleLoginRequest, { rejectValue: string }>(
  'auth/googleAuth',
  async (payload, { dispatch, rejectWithValue }) => {
    try {
      const res = await authAPI.googleAuth(payload)
      const guestSession = localStorage.getItem('guest_session_id')
      if (guestSession) {
        void dispatch(mergeCart(guestSession))
      } else {
        void dispatch(fetchCart())
      }
      return res
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Google authentication failed')
    }
  },
)

export const logout = createAsyncThunk('auth/logout', async (_, { dispatch }) => {
  tokenManager.removeToken()
  dispatch(clearCartState())
})

export const fetchProfile = createAsyncThunk<User, void, { rejectValue: string }>(
  'auth/fetchProfile',
  async (_, { rejectWithValue }) => {
    try {
      return await authAPI.getProfile()
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to fetch profile')
    }
  },
)

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null
    },
    setCredentials: (state, action: PayloadAction<LoginResponse>) => {
      const { token, user } = action.payload
      state.token = token
      state.user = user
      state.isAuthenticated = true
      state.error = null
      tokenManager.setToken(token)
    },
  },
  extraReducers: (builder) => {
    builder
      // Login handlers
      .addCase(login.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(login.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.user = action.payload.user
        state.token = action.payload.token
        tokenManager.setToken(action.payload.token)
      })
      .addCase(login.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload ?? 'Login failed'
      })
      // Register handlers
      .addCase(register.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(register.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.user = action.payload.user
        state.token = action.payload.token
        tokenManager.setToken(action.payload.token)
      })
      .addCase(register.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload ?? 'Registration failed'
      })
      // Google auth handlers
      .addCase(googleAuth.pending, (state) => {
        state.googleLoading = true
        state.error = null
      })
      .addCase(googleAuth.fulfilled, (state, action) => {
        state.googleLoading = false
        state.isAuthenticated = true
        state.user = action.payload.user
        state.token = action.payload.token
        tokenManager.setToken(action.payload.token)
      })
      .addCase(googleAuth.rejected, (state, action) => {
        state.googleLoading = false
        state.error = action.payload ?? 'Google authentication failed'
      })
      // Logout handler
      .addCase(logout.fulfilled, (state) => {
        state.isAuthenticated = false
        state.user = null
        state.token = null
        state.error = null
      })
      // Profile handler
      .addCase(fetchProfile.fulfilled, (state, action) => {
        state.user = action.payload
      })
      .addCase(fetchProfile.rejected, (state, action) => {
        state.isAuthenticated = false
        state.user = null
        state.token = null
        state.error = action.payload ?? 'Failed to fetch profile'
        tokenManager.removeToken()
      })
  },
})

export const { clearError, setCredentials } = authSlice.actions
export default authSlice.reducer
