import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import QueryLog from './pages/QueryLog'
import Blocklists from './pages/Blocklists'
import Rules from './pages/Rules'
import Dhcp from './pages/Dhcp'
import Network from './pages/Network'
import Settings from './pages/Settings'
import Statistics from './pages/Statistics'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="queries" element={<QueryLog />} />
        <Route path="blocklists" element={<Blocklists />} />
        <Route path="rules" element={<Rules />} />
        <Route path="dhcp" element={<Dhcp />} />
        <Route path="network" element={<Network />} />
        <Route path="settings" element={<Settings />} />
        <Route path="statistics" element={<Statistics />} />
      </Route>
    </Routes>
  )
}
