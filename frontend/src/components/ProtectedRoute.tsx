import { Navigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { keycloak } from '../utils/keycloak'

interface ProtectedRouteProps {
  children: JSX.Element
  adminOnly?: boolean
}

const ProtectedRoute = ({ children, adminOnly = false }: ProtectedRouteProps) => {
  const { isAuthenticated, user } = useAuth()

  if (!isAuthenticated) {
    void keycloak.login()
    return (
      <div className="flex-1 flex items-center justify-center min-h-[400px] bg-neutral-50 font-sans py-12">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-neutral-900 border-t-transparent"></div>
          <span className="text-xs text-slate-500 font-medium">Đang chuyển hướng đến trang đăng nhập...</span>
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="flex-1 flex items-center justify-center min-h-[400px] bg-neutral-50 font-sans py-12">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-neutral-900 border-t-transparent"></div>
          <span className="text-xs text-neutral-500 font-medium">Đang tải thông tin...</span>
        </div>
      </div>
    )
  }

  if (adminOnly && user.role !== 'admin') {
    return <Navigate to="/" replace />
  }

  return children
}

export default ProtectedRoute
