import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { functionsApi } from '@/services/api'
import { Plus, Play, Trash2, AlertCircle, Server, PowerOff } from 'lucide-react'

export default function Dashboard() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: functions, isLoading, error } = useQuery({
    queryKey: ['functions'],
    queryFn: functionsApi.list,
    refetchInterval: 5000, // Auto-refresh every 5 seconds
  })

  const deleteMutation = useMutation({
    mutationFn: functionsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] })
    },
  })

  const undeployMutation = useMutation({
    mutationFn: functionsApi.undeploy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] })
    },
  })

  const invokeMutation = useMutation({
    mutationFn: functionsApi.invoke,
  })

  const handleDelete = async (name: string) => {
    if (confirm(`Are you sure you want to delete function "${name}"?`)) {
      await deleteMutation.mutateAsync(name)
    }
  }

  const handleUndeploy = async (name: string) => {
    if (confirm(`Are you sure you want to undeploy function "${name}"? This will remove it from Kubernetes but keep it in the database.`)) {
      await undeployMutation.mutateAsync(name)
    }
  }

  const handleInvoke = async (name: string) => {
    await invokeMutation.mutateAsync(name)
    alert(`Function "${name}" invoked successfully!`)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Running':
        return 'bg-green-500'
      case 'Pending':
        return 'bg-yellow-500'
      case 'Failed':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-900/20 border border-red-500 rounded-lg p-4 flex items-start space-x-3">
        <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
        <div>
          <h3 className="text-red-400 font-semibold">Error loading functions</h3>
          <p className="text-red-300 text-sm mt-1">{(error as Error).message}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold text-white">Functions</h2>
          <p className="text-slate-400 mt-1">
            {functions?.length || 0} function{functions?.length !== 1 ? 's' : ''} deployed
          </p>
        </div>
        <button
          onClick={() => navigate('/functions/new')}
          className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition"
        >
          <Plus className="w-5 h-5" />
          <span>Create Function</span>
        </button>
      </div>

      {/* Functions Grid */}
      {functions && functions.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {functions.map((func) => (
            <div
              key={func.name}
              className="bg-slate-800 border border-slate-700 rounded-lg p-6 hover:border-primary-500 transition cursor-pointer"
              onClick={() => navigate(`/functions/${func.name}`)}
            >
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-white">{func.name}</h3>
                  <p className="text-sm text-slate-400 mt-1 truncate">{func.image}</p>
                </div>
                <div className={`w-3 h-3 rounded-full ${getStatusColor(func.status)}`} />
              </div>

              <div className="space-y-2 mb-4">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Status:</span>
                  <span className="text-white font-medium">{func.status}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Replicas:</span>
                  <span className="text-white">
                    {func.ready_replicas}/{func.replicas}
                  </span>
                </div>
              </div>

              <div className="flex space-x-2" onClick={(e) => e.stopPropagation()}>
                <button
                  onClick={() => handleInvoke(func.name)}
                  disabled={invokeMutation.isPending}
                  className="flex-1 flex items-center justify-center space-x-1 px-3 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-700 text-white text-sm rounded transition"
                >
                  <Play className="w-4 h-4" />
                  <span>Invoke</span>
                </button>
                <button
                  onClick={() => handleUndeploy(func.name)}
                  disabled={undeployMutation.isPending}
                  className="px-3 py-2 bg-orange-600 hover:bg-orange-700 disabled:bg-slate-700 text-white rounded transition"
                  title="Undeploy (remove from Kubernetes)"
                >
                  <PowerOff className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleDelete(func.name)}
                  disabled={deleteMutation.isPending}
                  className="px-3 py-2 bg-red-600 hover:bg-red-700 disabled:bg-slate-700 text-white rounded transition"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-12 bg-slate-800 rounded-lg border border-slate-700">
          <Server className="w-16 h-16 text-slate-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-300 mb-2">No functions deployed</h3>
          <p className="text-slate-500 mb-6">Get started by creating your first function</p>
          <button
            onClick={() => navigate('/functions/new')}
            className="inline-flex items-center space-x-2 px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition"
          >
            <Plus className="w-5 h-5" />
            <span>Create Function</span>
          </button>
        </div>
      )}
    </div>
  )
}
