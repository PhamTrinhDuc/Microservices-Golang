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
    </div>
  )
}

export default LoginPage
