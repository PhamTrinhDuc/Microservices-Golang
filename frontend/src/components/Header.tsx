import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

const Header = () => {
  const { isAuthenticated, user, logout } = useAuth()

  const handleLogout = async () => {
    await logout()
  }

  return (
    <header className="border-b bg-white">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
        <Link to="/" className="text-lg font-semibold text-gray-900">
          E-Commerce
        </Link>

        <nav className="flex items-center gap-4 text-sm text-gray-700">
          {isAuthenticated ? (
            <>
              <span className="text-gray-900">{user?.name ?? user?.email}</span>
              <button
                type="button"
                onClick={handleLogout}
                className="rounded bg-gray-900 px-3 py-1.5 text-white hover:bg-black"
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="hover:text-gray-900">
                Login
              </Link>
              <Link to="/register" className="hover:text-gray-900">
                Register
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}

export default Header
