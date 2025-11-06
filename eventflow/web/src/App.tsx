import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import CreateFunction from './pages/CreateFunction'
import FunctionDetails from './pages/FunctionDetails'
import Layout from './components/Layout'

function App() {
  const { token } = useAuth()

  if (!token) {
    return (
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/functions/new" element={<CreateFunction />} />
        <Route path="/functions/:name" element={<FunctionDetails />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  )
}

export default App
