import { useEffect } from 'react'
import { GoogleOAuthProvider } from '@react-oauth/google'
import { Navigate, Route, Routes } from 'react-router-dom'
import Footer from './components/Footer'
import Header from './components/Header'
import ProtectedRoute from './components/ProtectedRoute'
import JiyuuChat from './components/JiyuuChat'
import BrowsePage from './pages/BrowsePage'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import ProductDetailPage from './pages/ProductDetailPage'
import RegisterPage from './pages/RegisterPage'
import SearchPage from './pages/SearchPage'
import CartPage from './pages/CartPage'
import CheckoutPage from './pages/CheckoutPage'
import OrderSuccessPage from './pages/OrderSuccessPage'
import AdminDashboardPage from './pages/AdminDashboardPage'
import ProfilePage from './pages/ProfilePage'
import WishlistPage from './pages/WishlistPage'
import PolicyPage from './pages/PolicyPage'
import { useCart } from './hooks/useCart'
import { useAuth } from './hooks/useAuth'
import { useWishlist } from './hooks/useWishlist'
import { tokenManager } from './utils/tokenManager'

const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID || ''

const App = () => {
  const { fetchCart } = useCart()
  const { isAuthenticated, user, fetchProfile, logout } = useAuth()
  const { fetchWishlist } = useWishlist()

  useEffect(() => {
    void fetchCart()
  }, [fetchCart])

  useEffect(() => {
    if (isAuthenticated) {
      void fetchWishlist()
    }
  }, [isAuthenticated, fetchWishlist])

  useEffect(() => {
    if (isAuthenticated && (!user || user.id === 0)) {
      void fetchProfile()
    }
  }, [isAuthenticated, user, fetchProfile])

  useEffect(() => {
    if (isAuthenticated && !tokenManager.getToken()) {
      void logout()
    }
  }, [isAuthenticated, logout])

  return (
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
      <div className="flex min-h-screen flex-col bg-gray-50">
        <Header />
        <main className="flex flex-col flex-1 w-full">
          <Routes>
            <Route
              path="/cart"
              element={
                <CartPage />
              }
            />
            <Route
              path="/"
              element={
                <HomePage />
              }
            />
            <Route
              path="/browse"
              element={
                <BrowsePage />
              }
            />
            <Route
              path="/search"
              element={
                <SearchPage />
              }
            />
            <Route
              path="/products/:id"
              element={
                <ProductDetailPage />
              }
            />
            <Route
              path="/checkout"
              element={
                <ProtectedRoute>
                  <CheckoutPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/order-success"
              element={
                <ProtectedRoute>
                  <OrderSuccessPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/wishlist"
              element={
                <ProtectedRoute>
                  <WishlistPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <ProtectedRoute adminOnly={true}>
                  <AdminDashboardPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/policies/:slug"
              element={
                <ProtectedRoute>
                  <PolicyPage />
                </ProtectedRoute>
              }
            />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
        <Footer />
        <JiyuuChat />
      </div>
    </GoogleOAuthProvider>
  )
}

export default App
