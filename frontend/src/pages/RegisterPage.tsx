import { FormEvent, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { getPasswordStrength, validateEmail, validateName, validatePassword } from '../utils/validation'

const RegisterPage = () => {
  const navigate = useNavigate()
  const { register, isAuthenticated, loading, error, clearError } = useAuth()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const [nameError, setNameError] = useState<string | null>(null)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)

  const strength = useMemo(() => getPasswordStrength(password), [password])

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

    const nextNameError = validateName(name)
    const nextEmailError = validateEmail(email)
    const nextPasswordError = validatePassword(password)

    setNameError(nextNameError)
    setEmailError(nextEmailError)
    setPasswordError(nextPasswordError)

    if (nextNameError || nextEmailError || nextPasswordError) {
      return
    }

    await register(name, email, password)
  }

  const strengthColor =
    strength.label === 'Strong' ? 'bg-green-500' : strength.label === 'Medium' ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <section className="mx-auto flex w-full max-w-md flex-1 items-center px-4 py-10">
      <form className="w-full rounded-lg border bg-white p-6 shadow-sm" onSubmit={handleSubmit}>
        <h1 className="mb-6 text-center text-2xl font-bold text-gray-900">Register</h1>

        <label htmlFor="name" className="mb-1 block text-sm font-medium text-gray-700">
          Name
        </label>
        <input
          id="name"
          type="text"
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="mb-1 w-full rounded border px-3 py-2"
          autoComplete="name"
        />
        {nameError && <p className="mb-3 text-sm text-red-600">{nameError}</p>}

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
          className="mb-2 w-full rounded border px-3 py-2"
          autoComplete="new-password"
        />

        <div className="mb-1 h-2 w-full rounded bg-gray-200">
          <div
            className={`h-full rounded transition-all ${strengthColor}`}
            style={{ width: `${Math.max((strength.score / 4) * 100, 10)}%` }}
          />
        </div>
        <p className="mb-1 text-xs text-gray-600">Password strength: {strength.label}</p>

        {passwordError && <p className="mb-3 text-sm text-red-600">{passwordError}</p>}
        {error && <p className="mb-3 text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-gray-900 px-3 py-2 text-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Creating account...' : 'Register'}
        </button>

        <p className="mt-4 text-center text-sm text-gray-600">
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-gray-900">
            Login
          </Link>
        </p>
      </form>
    </section>
  )
}

export default RegisterPage
