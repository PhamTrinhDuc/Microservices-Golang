import { useEffect } from 'react'
import { useAuth } from '../hooks/useAuth'

const HomePage = () => {
  const { user, fetchProfile, error } = useAuth()

  useEffect(() => {
    if (!user) {
      void fetchProfile()
    }
  }, [fetchProfile, user])

  return (
    <section className="mx-auto w-full max-w-5xl flex-1 px-4 py-10">
      <h1 className="mb-4 text-3xl font-bold text-gray-900">Welcome to E-Commerce</h1>
      <p className="mb-6 text-gray-700">Your frontend foundation is ready.</p>

      <div className="rounded-lg border bg-white p-5 shadow-sm">
        <h2 className="mb-3 text-lg font-semibold text-gray-900">User Information</h2>
        {error && <p className="mb-3 text-sm text-red-600">{error}</p>}
        <dl className="space-y-2 text-sm text-gray-700">
          <div>
            <dt className="font-medium">Name</dt>
            <dd>{user?.name ?? '-'}</dd>
          </div>
          <div>
            <dt className="font-medium">Email</dt>
            <dd>{user?.email ?? '-'}</dd>
          </div>
          <div>
            <dt className="font-medium">User ID</dt>
            <dd>{user?.id ?? '-'}</dd>
          </div>
        </dl>
      </div>
    </section>
  )
}

export default HomePage
