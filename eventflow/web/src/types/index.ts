export interface FunctionStatus {
  name: string
  namespace?: string
  image: string
  replicas: number
  available_replicas: number
  ready_replicas: number
  updated_replicas: number
  status: 'Running' | 'Pending' | 'Failed'
  created_at: string
}

export interface CreateFunctionRequest {
  name: string
  image: string
  command?: string[]
  env?: Record<string, string>
  replicas: number
}

export interface AuthToken {
  token: string
  user_id: string
  username: string
  email?: string
  namespace: string
}

export interface User {
  user_id: string
  username: string
  email?: string
  namespace: string
}
