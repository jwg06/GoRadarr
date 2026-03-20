import { useEffect } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import axios from 'axios'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import AddMovie from './pages/AddMovie'
import CalendarPage from './pages/Calendar'
import History from './pages/History'
import Movies from './pages/Movies'
import Queue from './pages/Queue'
import Settings from './pages/Settings'
import System from './pages/System'
import Wanted from './pages/Wanted'
import { useAuthStore } from './stores/auth'

export default function App() {
  const token = useAuthStore((state) => state.token)
  const logout = useAuthStore((state) => state.logout)

  useEffect(() => {
    if (!token) return
    axios
      .get('/api/v1/auth/me', { headers: { Authorization: `Bearer ${token}` } })
      .catch(() => logout())
  }, [token, logout])

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Movies />} />
        <Route path="/add" element={<AddMovie />} />
        <Route path="/calendar" element={<CalendarPage />} />
        <Route path="/history" element={<History />} />
        <Route path="/queue" element={<Queue />} />
        <Route path="/wanted" element={<Wanted />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/system/status" element={<System />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
