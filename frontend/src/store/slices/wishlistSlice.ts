import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { wishlistAPI } from '../../services/wishlistAPI'
import type { WishlistItemResponse } from '../../types'

interface WishlistState {
  items: WishlistItemResponse[]
  loading: boolean
  error: string | null
}

const initialState: WishlistState = {
  items: [],
  loading: false,
  error: null,
}

export const fetchWishlist = createAsyncThunk<WishlistItemResponse[], void, { rejectValue: string }>(
  'wishlist/fetchWishlist',
  async (_, { rejectWithValue }) => {
    try {
      return await wishlistAPI.getWishlist()
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to load wishlist')
    }
  }
)

export const addToWishlist = createAsyncThunk<any, number, { rejectValue: string }>(
  'wishlist/addToWishlist',
  async (variantId, { dispatch, rejectWithValue }) => {
    try {
      const data = await wishlistAPI.addToWishlist(variantId)
      void dispatch(fetchWishlist())
      return data
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to add to wishlist')
    }
  }
)

export const removeFromWishlist = createAsyncThunk<void, number, { rejectValue: string }>(
  'wishlist/removeFromWishlist',
  async (variantId, { dispatch, rejectWithValue }) => {
    try {
      await wishlistAPI.removeFromWishlist(variantId)
      void dispatch(fetchWishlist())
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to remove from wishlist')
    }
  }
)

const wishlistSlice = createSlice({
  name: 'wishlist',
  initialState,
  reducers: {
    clearWishlistState: (state) => {
      state.items = []
      state.error = null
      state.loading = false
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchWishlist.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchWishlist.fulfilled, (state, action) => {
        state.loading = false
        state.items = action.payload
      })
      .addCase(fetchWishlist.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload ?? 'Failed to load wishlist'
      })
  },
})

export const { clearWishlistState } = wishlistSlice.actions
export default wishlistSlice.reducer
