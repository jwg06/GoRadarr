import { Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import AddMovie from './pages/AddMovie'
import CalendarPage from './pages/Calendar'
import History from './pages/History'
import Movies from './pages/Movies'
import Queue from './pages/Queue'
import Settings from './pages/Settings'
import System from './pages/System'
import Wanted from './pages/Wanted'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
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
