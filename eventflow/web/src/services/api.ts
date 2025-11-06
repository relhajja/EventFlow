import axios from 'axios'
import type { FunctionStatus, CreateFunctionRequest, AuthToken } from '@/types'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('eventflow_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export const authApi = {
  getToken: async (): Promise<AuthToken> => {
    const { data } = await axios.post<AuthToken>('/auth/token')
    return data
  },
}

export const functionsApi = {
  list: async (): Promise<FunctionStatus[]> => {
    const { data } = await api.get<FunctionStatus[]>('/v1/functions')
    return data
  },

  get: async (name: string): Promise<FunctionStatus> => {
    const { data } = await api.get<FunctionStatus>(`/v1/functions/${name}`)
    return data
  },

  create: async (functionData: CreateFunctionRequest): Promise<void> => {
    await api.post('/v1/functions', functionData)
  },

  delete: async (name: string): Promise<void> => {
    await api.delete(`/v1/functions/${name}`)
  },

  invoke: async (name: string, payload?: Record<string, unknown>): Promise<void> => {
    await api.post(`/v1/functions/${name}:invoke`, { payload })
  },

  getLogs: async (name: string): Promise<string> => {
    const { data } = await api.get<string>(`/v1/functions/${name}/logs`)
    return data
  },
}
