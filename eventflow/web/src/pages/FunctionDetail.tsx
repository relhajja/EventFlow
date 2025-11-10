import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { functionsApi } from '@/services/api'
import {
  ArrowLeft,
  Play,
  Pause,
  Trash2,
  Settings,
  Activity,
  Clock,
  Cpu,
  MemoryStick,
  Network,
  Save,
  X,
  AlertCircle,
  CheckCircle2,
  TrendingUp,
  Server
} from 'lucide-react'

interface FunctionConfig {
  replicas: number
  env: Record<string, string>
  resources?: {
    requests: {
      cpu: string
      memory: string
    }
    limits: {
      cpu: string
      memory: string
    }
  }
}

export default function FunctionDetail() {
  const { name } = useParams<{ name: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'overview' | 'config' | 'metrics' | 'logs'>('overview')
  const [isEditing, setIsEditing] = useState(false)
  const [config, setConfig] = useState<FunctionConfig>({
    replicas: 1,
    env: {},
  })

  const { data: func, isLoading } = useQuery({
    queryKey: ['function', name],
    queryFn: () => functionsApi.get(name!),
    refetchInterval: 5000,
  })

  const deleteMutation = useMutation({
    mutationFn: functionsApi.delete,
    onSuccess: () => {
      navigate('/dashboard')
    },
  })

  const undeployMutation = useMutation({
    mutationFn: functionsApi.undeploy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['function', name] })
    },
  })

  useEffect(() => {
    if (func) {
      setConfig({
        replicas: func.replicas || 1,
        env: {},
      })
    }
  }, [func])

  const handleDelete = () => {
    if (confirm(`Are you sure you want to delete "${name}"?`)) {
      deleteMutation.mutate(name!)
    }
  }

  const handleUndeploy = () => {
    if (confirm(`Undeploy "${name}"? This removes it from Kubernetes but keeps the configuration.`)) {
      undeployMutation.mutate(name!)
    }
  }

  const handleSaveConfig = () => {
    // TODO: Implement update function API
    setIsEditing(false)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
      </div>
    )
  }

  if (!func) {
    return (
      <div className="bg-red-900/20 border border-red-500 rounded-lg p-4">
        <p className="text-red-400">Function not found</p>
      </div>
    )
  }

  const getStatusBadge = (status: string) => {
    const badges = {
      Running: 'bg-green-500/20 text-green-400 border-green-500/50',
      Pending: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/50',
      Failed: 'bg-red-500/20 text-red-400 border-red-500/50',
    }
    return badges[status as keyof typeof badges] || badges.Pending
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate('/dashboard')}
            className="p-2 hover:bg-slate-700 rounded-lg transition"
          >
            <ArrowLeft className="w-5 h-5 text-slate-400" />
          </button>
          <div>
            <h2 className="text-3xl font-bold text-white">{func.name}</h2>
            <p className="text-slate-400 mt-1">{func.image}</p>
          </div>
          <span className={`px-3 py-1 rounded-full text-sm border ${getStatusBadge(func.status)}`}>
            {func.status}
          </span>
        </div>

        <div className="flex space-x-2">
          <button
            onClick={handleUndeploy}
            disabled={undeployMutation.isPending}
            className="flex items-center space-x-2 px-4 py-2 bg-orange-600 hover:bg-orange-700 disabled:bg-slate-700 text-white rounded-lg transition"
          >
            <Pause className="w-4 h-4" />
            <span>Undeploy</span>
          </button>
          <button
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="flex items-center space-x-2 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-slate-700 text-white rounded-lg transition"
          >
            <Trash2 className="w-4 h-4" />
            <span>Delete</span>
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 border-b border-slate-700">
        {[
          { id: 'overview', label: 'Overview', icon: Activity },
          { id: 'config', label: 'Configuration', icon: Settings },
          { id: 'metrics', label: 'Metrics', icon: TrendingUp },
          { id: 'logs', label: 'Logs', icon: Server },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as any)}
            className={`flex items-center space-x-2 px-4 py-3 border-b-2 transition ${
              activeTab === tab.id
                ? 'border-primary-500 text-primary-400'
                : 'border-transparent text-slate-400 hover:text-slate-300'
            }`}
          >
            <tab.icon className="w-4 h-4" />
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div>
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {/* Status Card */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-slate-400">Status</h3>
                <CheckCircle2 className="w-5 h-5 text-green-400" />
              </div>
              <div className="text-2xl font-bold text-white">{func.status}</div>
              <p className="text-sm text-slate-400 mt-1">
                {func.ready_replicas}/{func.replicas} ready
              </p>
            </div>

            {/* Replicas Card */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-slate-400">Replicas</h3>
                <Server className="w-5 h-5 text-blue-400" />
              </div>
              <div className="text-2xl font-bold text-white">{func.replicas}</div>
              <p className="text-sm text-slate-400 mt-1">
                Available: {func.available_replicas}
              </p>
            </div>

            {/* Namespace Card */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-slate-400">Namespace</h3>
                <Network className="w-5 h-5 text-purple-400" />
              </div>
              <div className="text-lg font-bold text-white break-all">{func.namespace}</div>
            </div>

            {/* Created Card */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-slate-400">Created</h3>
                <Clock className="w-5 h-5 text-yellow-400" />
              </div>
              <div className="text-sm font-bold text-white">
                {new Date(func.created_at).toLocaleDateString()}
              </div>
              <p className="text-sm text-slate-400 mt-1">
                {new Date(func.created_at).toLocaleTimeString()}
              </p>
            </div>

            {/* Image Info */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 md:col-span-2">
              <h3 className="text-sm font-medium text-slate-400 mb-3">Container Image</h3>
              <code className="text-sm text-primary-400 bg-slate-900 px-3 py-2 rounded block break-all">
                {func.image}
              </code>
            </div>

            {/* Quick Actions */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 md:col-span-2">
              <h3 className="text-sm font-medium text-slate-400 mb-4">Quick Actions</h3>
              <div className="grid grid-cols-2 gap-3">
                <button className="flex items-center justify-center space-x-2 px-4 py-3 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition">
                  <Play className="w-4 h-4" />
                  <span>Invoke</span>
                </button>
                <button className="flex items-center justify-center space-x-2 px-4 py-3 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition">
                  <Activity className="w-4 h-4" />
                  <span>View Logs</span>
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'config' && (
          <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-white">Function Configuration</h3>
              {!isEditing ? (
                <button
                  onClick={() => setIsEditing(true)}
                  className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition"
                >
                  <Settings className="w-4 h-4" />
                  <span>Edit</span>
                </button>
              ) : (
                <div className="flex space-x-2">
                  <button
                    onClick={() => setIsEditing(false)}
                    className="flex items-center space-x-2 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
                  >
                    <X className="w-4 h-4" />
                    <span>Cancel</span>
                  </button>
                  <button
                    onClick={handleSaveConfig}
                    className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition"
                  >
                    <Save className="w-4 h-4" />
                    <span>Save</span>
                  </button>
                </div>
              )}
            </div>

            <div className="space-y-6">
              {/* Replicas */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Replicas
                </label>
                <input
                  type="number"
                  value={config.replicas}
                  onChange={(e) => setConfig({ ...config, replicas: parseInt(e.target.value) })}
                  disabled={!isEditing}
                  className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white disabled:opacity-50 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  min="0"
                  max="10"
                />
              </div>

              {/* Environment Variables */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Environment Variables
                </label>
                <div className="bg-slate-900 border border-slate-700 rounded-lg p-4">
                  {Object.keys(config.env).length > 0 ? (
                    <div className="space-y-2">
                      {Object.entries(config.env).map(([key, value]) => (
                        <div key={key} className="flex items-center space-x-2">
                          <code className="flex-1 text-sm text-slate-300">
                            {key} = {value}
                          </code>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-slate-500">No environment variables configured</p>
                  )}
                </div>
              </div>

              {/* Resources (placeholder) */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Resource Limits
                </label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">CPU Request</label>
                    <input
                      type="text"
                      placeholder="100m"
                      disabled={!isEditing}
                      className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm disabled:opacity-50"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">CPU Limit</label>
                    <input
                      type="text"
                      placeholder="500m"
                      disabled={!isEditing}
                      className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm disabled:opacity-50"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Memory Request</label>
                    <input
                      type="text"
                      placeholder="128Mi"
                      disabled={!isEditing}
                      className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm disabled:opacity-50"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Memory Limit</label>
                    <input
                      type="text"
                      placeholder="256Mi"
                      disabled={!isEditing}
                      className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm disabled:opacity-50"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'metrics' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">CPU Usage</h3>
                <Cpu className="w-5 h-5 text-blue-400" />
              </div>
              <div className="text-3xl font-bold text-white mb-2">--</div>
              <div className="w-full bg-slate-700 rounded-full h-2">
                <div className="bg-blue-500 h-2 rounded-full" style={{ width: '0%' }}></div>
              </div>
              <p className="text-sm text-slate-400 mt-2">Metrics coming soon</p>
            </div>

            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">Memory Usage</h3>
                <MemoryStick className="w-5 h-5 text-green-400" />
              </div>
              <div className="text-3xl font-bold text-white mb-2">--</div>
              <div className="w-full bg-slate-700 rounded-full h-2">
                <div className="bg-green-500 h-2 rounded-full" style={{ width: '0%' }}></div>
              </div>
              <p className="text-sm text-slate-400 mt-2">Metrics coming soon</p>
            </div>

            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 md:col-span-2">
              <h3 className="text-lg font-semibold text-white mb-4">Invocation History</h3>
              <div className="text-center py-8 text-slate-500">
                <AlertCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>No invocation data available</p>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'logs' && (
          <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Function Logs</h3>
            <div className="bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300 h-96 overflow-y-auto">
              <p className="text-slate-500">Logs coming soon...</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
