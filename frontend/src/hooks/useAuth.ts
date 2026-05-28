import { useDispatch, useSelector } from 'react-redux'
import { clearError, login, logout, register, fetchProfile } from '../store/slices/authSlice'
import type { AppDispatch, RootState } from '../store'

export const useAuth = () => {
  const dispatch = useDispatch<AppDispatch>()
  const auth = useSelector((state: RootState) => state.auth)

  return {
    ...auth,
    login: (email: string, password: string) => dispatch(login({ email, password })),
    register: (name: string, email: string, password: string) =>
      dispatch(register({ name, email, password })),
    logout: () => dispatch(logout()),
    fetchProfile: () => dispatch(fetchProfile()),
    clearError: () => dispatch(clearError()),
  }
}
