import { useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { Navigate } from 'react-router-dom'
import { Server, Users } from 'lucide-react'

export default function Login() {
  const { isAuthenticated, login } = useAuth()
  const [selectedUser, setSelectedUser] = useState('alice')
  const [isLoading, setIsLoading] = useState(false)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  const demoUsers = [
    { id: 'alice', name: 'Alice (Product Team)', email: 'alice@company.com' },
    { id: 'bob', name: 'Bob (Engineering Team)', email: 'bob@company.com' },
    { id: 'charlie', name: 'Charlie (Marketing Team)', email: 'charlie@company.com' },
    { id: 'demo-user', name: 'Demo User', email: 'demo@eventflow.io' },
  ]

  const handleLogin = async () => {
    setIsLoading(true)
    try {
      const user = demoUsers.find(u => u.id === selectedUser)
      if (user) {
        await login(user.id, user.name, user.email)
      }
    } catch (error) {
      console.error('Login failed:', error)
      alert('Login failed. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center p-4">
      <div className="max-w-md w-full space-y-8 p-8 bg-slate-800 rounded-xl shadow-2xl">
        <div className="text-center">
          <div className="flex justify-center mb-4">
            <Server className="w-16 h-16 text-primary-400" />
          </div>
          <h2 className="text-3xl font-bold text-white">EventFlow</h2>
          <p className="mt-2 text-slate-400">Multi-Tenant FaaS Platform</p>
        </div>
        
        <div className="mt-8 space-y-6">
          {/* User Selection */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-3">
              <Users className="w-4 h-4 inline mr-2" />
              Select User (Tenant)
            </label>
            <div className="space-y-2">
              {demoUsers.map((user) => (
                <label
                  key={user.id}
                  className={`flex items-center p-3 rounded-lg border cursor-pointer transition ${
                    selectedUser === user.id
                      ? 'border-primary-500 bg-primary-500/10'
                      : 'border-slate-700 hover:border-slate-600'
                  }`}
                >
                  <input
                    type="radio"
                    name="user"
                    value={user.id}
                    checked={selectedUser === user.id}
                    onChange={(e) => setSelectedUser(e.target.value)}
                    className="mr-3"
                  />
                  <div className="flex-1">
                    <div className="text-white font-medium">{user.name}</div>
                    <div className="text-xs text-slate-400">{user.email}</div>
                    <div className="text-xs text-slate-500 mt-1">Namespace: tenant-{user.id}</div>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <button
            onClick={handleLogin}
            disabled={isLoading}
            className="w-full flex justify-center py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Generating Token...' : 'Generate Token & Login'}
          </button>
          
          <p className="text-xs text-center text-slate-500">
            Development mode - Multi-tenant JWT authentication
            <br />
            Each user has isolated functions in their own namespace
          </p>
        </div>
      </div>
    </div>
  )
}
