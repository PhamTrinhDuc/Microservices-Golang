import { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import type { AppDispatch, RootState } from '../store'
import { fetchCart, addToCart, updateItemQty, removeItem, clearCart } from '../store/slices/cartSlice'

export const useCart = () => {
  const dispatch = useDispatch<AppDispatch>()
  const cart = useSelector((state: RootState) => state.cart)

  const totalItems = cart.items.reduce((sum, item) => sum + item.quantity, 0)
  const cartSubtotal = cart.items.reduce((sum, item) => sum + (item.price * item.quantity), 0)

  const handleFetchCart = useCallback(() => dispatch(fetchCart()), [dispatch])
  const handleAddToCart = useCallback((variantId: number, quantity: number) => 
    dispatch(addToCart({ variant_id: variantId, quantity })), [dispatch])
  const handleUpdateItemQty = useCallback((itemId: number, quantity: number) => 
    dispatch(updateItemQty({ itemId, quantity })), [dispatch])
  const handleRemoveItem = useCallback((itemId: number) => 
    dispatch(removeItem(itemId)), [dispatch])
  const handleClearCart = useCallback(() => 
    dispatch(clearCart()), [dispatch])

  return {
    ...cart,
    totalItems,
    cartSubtotal,
    fetchCart: handleFetchCart,
    addToCart: handleAddToCart,
    updateItemQty: handleUpdateItemQty,
    removeItem: handleRemoveItem,
    clearCart: handleClearCart,
  }
}
