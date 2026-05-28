import { FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { GoogleLogin } from '@react-oauth/google'
import { useAuth } from '../hooks/useAuth'
import { validateEmail, validatePassword } from '../utils/validation'

const LoginPage = () => {
  const navigate = useNavigate()
  const { login, isAuthenticated, loading, error, clearError, googleAuth, googleLoading } = useAuth()

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
    <section className="mx-auto flex w-full max-w-md flex-1 items-center px-4 py-10">
      <form className="w-full rounded-lg border bg-white p-6 shadow-sm" onSubmit={handleSubmit}>
        <h1 className="mb-6 text-center text-2xl font-bold text-gray-900">Login</h1>

        <label htmlFor="email" className="mb-1 block text-sm font-medium text-gray-700">
          Email
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mb-1 w-full rounded border px-3 py-2"
          autoComplete="email"
        />
        {emailError && <p className="mb-3 text-sm text-red-600">{emailError}</p>}

        <label htmlFor="password" className="mb-1 block text-sm font-medium text-gray-700">
          Password
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="mb-1 w-full rounded border px-3 py-2"
          autoComplete="current-password"
        />
        {passwordError && <p className="mb-3 text-sm text-red-600">{passwordError}</p>}

        {error && <p className="mb-3 text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-gray-900 px-3 py-2 text-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Logging in...' : 'Login'}
        </button>

        <div className="my-4 flex items-center">
          <div className="flex-1 border-t border-gray-300"></div>
          <span className="px-2 text-sm text-gray-500">Or</span>
          <div className="flex-1 border-t border-gray-300"></div>
        </div>

        <div className="mb-4 flex justify-center">
          <GoogleLogin
            onSuccess={handleGoogleSuccess}
            onError={() => clearError()}
            useOneTap
          />
        </div>

        <p className="text-center text-sm text-gray-600">
          No account?{' '}
          <Link to="/register" className="font-medium text-gray-900">
            Register
          </Link>
        </p>
      </form>
    </section>
  )
}

export default LoginPage
