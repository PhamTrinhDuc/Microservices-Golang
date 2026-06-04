import { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import type { AppDispatch, RootState } from '../store'
import { fetchWishlist, addToWishlist, removeFromWishlist } from '../store/slices/wishlistSlice'

export const useWishlist = () => {
  const dispatch = useDispatch<AppDispatch>()
  const wishlist = useSelector((state: RootState) => state.wishlist)

  const handleFetchWishlist = useCallback(() => dispatch(fetchWishlist()), [dispatch])
  
  const handleAddToWishlist = useCallback((variantId: number) => 
    dispatch(addToWishlist(variantId)), [dispatch])
    
  const handleRemoveFromWishlist = useCallback((variantId: number) => 
    dispatch(removeFromWishlist(variantId)), [dispatch])

  const isWishlisted = useCallback((variantId: number) => {
    return wishlist.items.some((item) => item.variant_id === variantId)
  }, [wishlist.items])

  const toggleWishlist = useCallback(async (variantId: number) => {
    if (isWishlisted(variantId)) {
      return dispatch(removeFromWishlist(variantId)).unwrap()
    } else {
      return dispatch(addToWishlist(variantId)).unwrap()
    }
  }, [dispatch, isWishlisted])

  return {
    ...wishlist,
    fetchWishlist: handleFetchWishlist,
    addToWishlist: handleAddToWishlist,
    removeFromWishlist: handleRemoveFromWishlist,
    isWishlisted,
    toggleWishlist,
  }
}
