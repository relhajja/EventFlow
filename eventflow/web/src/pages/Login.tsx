import { useAuth } from '@/context/AuthContext'
import { Navigate } from 'react-router-dom'
import { Server } from 'lucide-react'

export default function Login() {
  const { isAuthenticated, login } = useAuth()

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  const handleLogin = async () => {
    try {
      await login()
    } catch (error) {
      console.error('Login failed:', error)
      alert('Login failed. Please try again.')
    }
  }

  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center">
      <div className="max-w-md w-full space-y-8 p-8 bg-slate-800 rounded-xl shadow-2xl">
        <div className="text-center">
          <div className="flex justify-center mb-4">
            <Server className="w-16 h-16 text-primary-400" />
          </div>
          <h2 className="text-3xl font-bold text-white">EventFlow</h2>
          <p className="mt-2 text-slate-400">Functions as a Service Platform</p>
        </div>
        
        <div className="mt-8 space-y-4">
          <button
            onClick={handleLogin}
            className="w-full flex justify-center py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition"
          >
            Get Dev Token & Login
          </button>
          
          <p className="text-xs text-center text-slate-500">
            Development mode - JWT authentication
          </p>
        </div>
      </div>
    </div>
  )
}
