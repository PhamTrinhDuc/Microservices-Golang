import { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { clearError, login, logout, register, googleAuth, fetchProfile } from '../store/slices/authSlice'
import type { AppDispatch, RootState } from '../store'

export const useAuth = () => {
  const dispatch = useDispatch<AppDispatch>()
  const auth = useSelector((state: RootState) => state.auth)

  const handleLogin = useCallback((email: string, password: string) => 
    dispatch(login({ email, password })), [dispatch])

  const handleRegister = useCallback((fullName: string, email: string, password: string) =>
    dispatch(register({ full_name: fullName, email, password })), [dispatch])

  const handleGoogleAuth = useCallback((credential: string) => 
    dispatch(googleAuth({ credential })), [dispatch])

  const handleLogout = useCallback(() => 
    dispatch(logout()), [dispatch])

  const handleFetchProfile = useCallback(() => 
    dispatch(fetchProfile()), [dispatch])

  const handleClearError = useCallback(() => 
    dispatch(clearError()), [dispatch])

  return {
    ...auth,
    login: handleLogin,
    register: handleRegister,
    googleAuth: handleGoogleAuth,
    logout: handleLogout,
    fetchProfile: handleFetchProfile,
    clearError: handleClearError,
  }
}
