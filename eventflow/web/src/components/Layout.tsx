import { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { Server, LogOut, User } from 'lucide-react'

export default function Layout({ children }: { children: ReactNode }) {
  const { isAuthenticated, user, logout } = useAuth()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="min-h-screen bg-slate-900">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-3">
              <Server className="w-8 h-8 text-primary-400" />
              <h1 className="text-2xl font-bold text-white">EventFlow</h1>
              <span className="text-sm text-slate-400">Multi-Tenant FaaS</span>
            </div>
            
            <div className="flex items-center space-x-4">
              {/* User Info */}
              {user && (
                <div className="flex items-center space-x-2 px-4 py-2 bg-slate-700/50 rounded-lg">
                  <User className="w-4 h-4 text-primary-400" />
                  <div className="text-sm">
                    <div className="text-white font-medium">{user.username}</div>
                    <div className="text-xs text-slate-400">{user.namespace}</div>
                  </div>
                </div>
              )}
              
              <button
                onClick={logout}
                className="flex items-center space-x-2 px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition"
              >
                <LogOut className="w-4 h-4" />
                <span>Logout</span>
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  )
}
