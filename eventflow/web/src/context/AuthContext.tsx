import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/services/api'
import type { User } from '@/types'

interface AuthContextType {
  isAuthenticated: boolean
  token: string | null
  user: User | null
  login: (userId: string, username: string, email?: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [token, setToken] = useState<string | null>(null)
  const [user, setUser] = useState<User | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    const storedToken = localStorage.getItem('eventflow_token')
    const storedUser = localStorage.getItem('eventflow_user')
    setToken(storedToken)
    if (storedUser) {
      setUser(JSON.parse(storedUser))
    }
    setIsAuthenticated(!!storedToken)
  }, [])

  const login = async (userId: string, username: string, email?: string) => {
    try {
      const authData = await authApi.getToken(userId, username, email)
      localStorage.setItem('eventflow_token', authData.token)
      const userData: User = {
        user_id: authData.user_id,
        username: authData.username,
        email: authData.email,
        namespace: authData.namespace
      }
      localStorage.setItem('eventflow_user', JSON.stringify(userData))
      setToken(authData.token)
      setUser(userData)
      setIsAuthenticated(true)
      navigate('/')
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  const logout = () => {
    localStorage.removeItem('eventflow_token')
    localStorage.removeItem('eventflow_user')
    setToken(null)
    setUser(null)
    setIsAuthenticated(false)
    navigate('/login')
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, token, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
