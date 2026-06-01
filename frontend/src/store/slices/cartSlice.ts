import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { cartAPI } from '../../services/cartAPI'
import { sessionManager } from '../../utils/sessionManager'
import type { CartItem, AddToCartRequest } from '../../types'

interface CartState {
  items: CartItem[]
  loading: boolean
  error: string | null
}

const initialState: CartState = {
  items: [],
  loading: false,
  error: null,
}

export const fetchCart = createAsyncThunk<CartItem[], void, { rejectValue: string }>(
  'cart/fetchCart',
  async (_, { rejectWithValue }) => {
    try {
      return await cartAPI.getCart()
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to load cart')
    }
  }
)

export const addToCart = createAsyncThunk<any, AddToCartRequest, { rejectValue: string }>(
  'cart/addToCart',
  async (payload, { dispatch, rejectWithValue }) => {
    try {
      const data = await cartAPI.addToCart(payload)
      void dispatch(fetchCart())
      return data
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to add item to cart')
    }
  }
)

export const updateItemQty = createAsyncThunk<any, { itemId: number; quantity: number }, { rejectValue: string }>(
  'cart/updateItemQty',
  async ({ itemId, quantity }, { dispatch, rejectWithValue }) => {
    try {
      const data = await cartAPI.updateItemQty(itemId, quantity)
      void dispatch(fetchCart())
      return data
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to update item quantity')
    }
  }
)

export const removeItem = createAsyncThunk<void, number, { rejectValue: string }>(
  'cart/removeItem',
  async (itemId, { dispatch, rejectWithValue }) => {
    try {
      await cartAPI.removeItem(itemId)
      void dispatch(fetchCart())
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to remove item')
    }
  }
)

export const clearCart = createAsyncThunk<void, void, { rejectValue: string }>(
  'cart/clearCart',
  async (_, { dispatch, rejectWithValue }) => {
    try {
      await cartAPI.clearCart()
      void dispatch(fetchCart())
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to clear cart')
    }
  }
)

export const mergeCart = createAsyncThunk<any, string, { rejectValue: string }>(
  'cart/mergeCart',
  async (sessionId, { dispatch, rejectWithValue }) => {
    try {
      const data = await cartAPI.mergeCart(sessionId)
      sessionManager.clearSessionId() // Clear local session ID after merge
      void dispatch(fetchCart())
      return data
    } catch (error: any) {
      return rejectWithValue(error?.message || 'Failed to merge cart')
    }
  }
)

const cartSlice = createSlice({
  name: 'cart',
  initialState,
  reducers: {
    clearCartState: (state) => {
      state.items = []
      state.error = null
      state.loading = false
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCart.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchCart.fulfilled, (state, action) => {
        state.loading = false
        state.items = action.payload
      })
      .addCase(fetchCart.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload ?? 'Failed to load cart'
      })
  },
})

export const { clearCartState } = cartSlice.actions
export default cartSlice.reducer
