import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/services/api'

interface AuthContextType {
  isAuthenticated: boolean
  token: string | null
  login: () => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [token, setToken] = useState<string | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    const storedToken = localStorage.getItem('eventflow_token')
    setToken(storedToken)
    setIsAuthenticated(!!storedToken)
  }, [])

  const login = async () => {
    try {
      const { token: newToken } = await authApi.getToken()
      localStorage.setItem('eventflow_token', newToken)
      setToken(newToken)
      setIsAuthenticated(true)
      navigate('/')
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  const logout = () => {
    localStorage.removeItem('eventflow_token')
    setToken(null)
    setIsAuthenticated(false)
    navigate('/login')
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, token, login, logout }}>
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
