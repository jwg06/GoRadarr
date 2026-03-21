import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const login = useAuthStore((state) => state.login)
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const res = await axios.post<{ token: string; expires: string }>('/api/v1/auth/login', {
        username,
        password,
      })
      const { token } = res.data
      const meRes = await axios.get<{ username: string }>('/api/v1/auth/me', {
        headers: { Authorization: `Bearer ${token}` },
      })
      login(token, meRes.data.username)
      navigate('/')
    } catch (err) {
      if (axios.isAxiosError(err)) {
        setError((err.response?.data as { message?: string })?.message ?? 'Login failed')
      } else {
        setError('Login failed')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-950">
      <div className="w-full max-w-sm">
        <div className="panel p-8">
          <div className="mb-8 text-center">
            <p className="text-xs uppercase tracking-[0.3em] text-yellow-400">GoRadarr</p>
            <h1 className="mt-2 text-2xl font-semibold text-gray-50">Sign in</h1>
            <p className="mt-1 text-sm text-gray-400">Enter your credentials to continue</p>
          </div>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm text-gray-400">
                Username
              </label>
              <input
                id="username"
                className="field mt-2"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
              />
            </div>
            <div>
              <label htmlFor="password" className="block text-sm text-gray-400">
                Password
              </label>
              <input
                id="password"
                className="field mt-2"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </div>
            {error ? (
              <p className="rounded bg-red-900/40 px-3 py-2 text-sm text-red-400">{error}</p>
            ) : null}
            <button type="submit" className="btn-primary w-full justify-center" disabled={loading}>
              {loading ? 'Signing in…' : 'Sign in'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
