import { createAsyncThunk, createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { authAPI } from '../../services/authAPI'
import type { LoginRequest, LoginResponse, RegisterRequest, User } from '../../types'
import { tokenManager } from '../../utils/tokenManager'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null
}

const initialState: AuthState = {
  user: null,
  token: tokenManager.getToken(),
  isAuthenticated: tokenManager.isAuthenticated(),
  loading: false,
  error: null,
}

export const login = createAsyncThunk<LoginResponse, LoginRequest, { rejectValue: string }>(
  'auth/login',
  async (payload, { rejectWithValue }) => {
    try {
      return await authAPI.login(payload)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Login failed')
    }
  },
)

export const register = createAsyncThunk<LoginResponse, RegisterRequest, { rejectValue: string }>(
  'auth/register',
  async (payload, { rejectWithValue }) => {
    try {
      return await authAPI.register(payload)
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Registration failed')
    }
  },
)

export const logout = createAsyncThunk('auth/logout', async () => {
  tokenManager.removeToken()
})

export const fetchProfile = createAsyncThunk<User, void, { rejectValue: string }>(
  'auth/fetchProfile',
  async (_, { rejectWithValue }) => {
    try {
      return await authAPI.getProfile()
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch profile')
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
      .addCase(logout.fulfilled, (state) => {
        state.isAuthenticated = false
        state.user = null
        state.token = null
        state.error = null
      })
      .addCase(fetchProfile.fulfilled, (state, action) => {
        state.user = action.payload
      })
      .addCase(fetchProfile.rejected, (state, action) => {
        state.error = action.payload ?? 'Failed to fetch profile'
      })
  },
})

export const { clearError, setCredentials } = authSlice.actions
export default authSlice.reducer
