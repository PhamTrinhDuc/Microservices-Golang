import { FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { GoogleLogin } from '@react-oauth/google'
import { useAuth } from '../hooks/useAuth'
import { validateEmail, validatePassword } from '../utils/validation'

const LoginPage = () => {
  const navigate = useNavigate()
  const { login, isAuthenticated, loading, error, clearError, googleAuth } = useAuth()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [emailError, setEmailError] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [modalMessage, setModalMessage] = useState('')

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/')
    }
  }, [isAuthenticated, navigate])

  useEffect(() => {
    return () => {
      clearError()
    }
  }, [clearError])

  useEffect(() => {
    if (error) {
      let friendlyMsg = error
      const lowerErr = error.toLowerCase()
      if (lowerErr.includes('network') || lowerErr.includes('conn') || lowerErr.includes('refused')) {
        friendlyMsg = 'Lỗi kết nối: Không thể kết nối tới máy chủ. Vui lòng đảm bảo rằng backend đã khởi chạy.'
      } else if (lowerErr.includes('credential') || lowerErr.includes('unauthorized') || lowerErr.includes('invalid')) {
        friendlyMsg = 'Đăng nhập thất bại: Email hoặc mật khẩu không chính xác.'
      }
      setModalMessage(friendlyMsg)
      setModalOpen(true)
    } else {
      setModalOpen(false)
    }
  }, [error, clearError])

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const nextEmailError = validateEmail(email)
    const nextPasswordError = validatePassword(password)

    setEmailError(nextEmailError)
    setPasswordError(nextPasswordError)

    if (nextEmailError || nextPasswordError) {
      return
    }

    await login(email, password)
  }

  const handleGoogleSuccess = async (credentialResponse: any) => {
    if (credentialResponse.credential) {
      await googleAuth(credentialResponse.credential)
    }
  }

  return (
    <div className="bg-mesh flex-1 flex items-center justify-center py-16 px-6">
      <form className="w-full max-w-md rounded-2xl border border-slate-100 bg-white p-8 shadow-premium animate-fade-in-up" onSubmit={handleSubmit}>
        
        {/* Title Header */}
        <h1 className="text-2xl font-extrabold text-slate-800 text-center tracking-tight">Chào mừng quay lại</h1>
        <p className="text-xs text-slate-500 text-center mt-1.5 mb-6">Đăng nhập vào tài khoản Jiyuu Store của bạn</p>

        {/* Email Field */}
        <div className="mb-4">
          <label htmlFor="email" className="mb-1.5 block text-xs font-bold uppercase tracking-wider text-slate-500">
            Địa chỉ Email
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-4 py-2.5 text-sm transition-all focus:border-brand-500 focus:bg-white focus:outline-none focus:ring-4 focus:ring-brand-500/10"
            autoComplete="email"
            placeholder="name@example.com"
          />
          {emailError && <p className="mt-1 text-xs font-semibold text-red-500">{emailError}</p>}
        </div>

        {/* Password Field */}
        <div className="mb-4">
          <label htmlFor="password" className="mb-1.5 block text-xs font-bold uppercase tracking-wider text-slate-500">
            Mật khẩu
          </label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-4 py-2.5 text-sm transition-all focus:border-brand-500 focus:bg-white focus:outline-none focus:ring-4 focus:ring-brand-500/10"
            autoComplete="current-password"
            placeholder="••••••••"
          />
          {passwordError && <p className="mt-1 text-xs font-semibold text-red-500">{passwordError}</p>}
        </div>

        {error && <p className="mb-4 text-xs font-bold text-red-500 text-center">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-xl bg-slate-950 py-3 text-xs font-extrabold tracking-wider uppercase text-white shadow-md shadow-slate-950/10 transition-all hover:bg-brand-600 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Đang đăng nhập...' : 'Đăng nhập'}
        </button>

        {/* Divider */}
        <div className="my-5 flex items-center">
          <div className="flex-1 border-t border-slate-100"></div>
          <span className="px-3 text-xs font-bold uppercase tracking-widest text-slate-400">Hoặc</span>
          <div className="flex-1 border-t border-slate-100"></div>
        </div>

        {/* Google Authentication */}
        <div className="mb-5 flex justify-center">
          <GoogleLogin
            onSuccess={handleGoogleSuccess}
            onError={() => clearError()}
            useOneTap
          />
        </div>

        {/* Footer Navigation link */}
        <p className="text-center text-xs text-slate-500">
          Chưa có tài khoản?{' '}
          <Link to="/register" className="font-bold text-brand-600 hover:text-brand-700 transition-colors">
            Đăng ký ngay
          </Link>
        </p>

      </form>

      {/* Connection & Auth Error Modal */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="bg-white border border-neutral-200 rounded-2xl shadow-2xl w-full max-w-sm p-6 space-y-4 text-center">
            <div className="w-12 h-12 bg-red-50 text-red-600 rounded-full flex items-center justify-center mx-auto border border-red-100">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            
            <div className="space-y-1.5">
              <h3 className="text-sm font-black uppercase tracking-wide text-neutral-900">Thông báo</h3>
              <p className="text-xs text-neutral-550 leading-relaxed font-semibold">{modalMessage}</p>
            </div>

            <button
              type="button"
              onClick={() => {
                setModalOpen(false)
                clearError()
              }}
              className="w-full bg-black hover:bg-neutral-850 text-white text-xs font-black uppercase tracking-wider py-2.5 rounded-xl transition-all active:scale-[0.98]"
            >
              Đóng
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default LoginPage
