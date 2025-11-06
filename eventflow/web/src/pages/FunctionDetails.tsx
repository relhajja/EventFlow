import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { functionsApi } from '@/services/api'
import { ArrowLeft, Play, Trash2, RefreshCw, Terminal } from 'lucide-react'

export default function FunctionDetails() {
  const { name } = useParams<{ name: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [logs, setLogs] = useState<string>('')

  const { data: functionData, isLoading } = useQuery({
    queryKey: ['function', name],
    queryFn: () => functionsApi.get(name!),
    refetchInterval: 3000,
    enabled: !!name,
  })

  const invokeMutation = useMutation({
    mutationFn: () => functionsApi.invoke(name!),
  })

  const deleteMutation = useMutation({
    mutationFn: () => functionsApi.delete(name!),
    onSuccess: () => {
      navigate('/')
    },
  })

  const loadLogs = async () => {
    if (name) {
      try {
        const logsData = await functionsApi.getLogs(name)
        setLogs(logsData)
      } catch (error) {
        setLogs(`Error loading logs: ${error}`)
      }
    }
  }

  useEffect(() => {
    loadLogs()
  }, [name])

  const handleInvoke = async () => {
    await invokeMutation.mutateAsync()
    alert('Function invoked successfully!')
    queryClient.invalidateQueries({ queryKey: ['function', name] })
  }

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete function "${name}"?`)) {
      await deleteMutation.mutateAsync()
    }
  }

  const getStatusColor = (status?: string) => {
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

  if (!functionData) {
    return (
      <div className="text-center py-12">
        <h3 className="text-lg text-slate-300">Function not found</h3>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate('/')}
            className="text-slate-400 hover:text-white transition"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h2 className="text-3xl font-bold text-white">{functionData.name}</h2>
            <p className="text-slate-400 mt-1">{functionData.image}</p>
          </div>
          <div className={`w-3 h-3 rounded-full ${getStatusColor(functionData.status)}`} />
        </div>
        
        <div className="flex space-x-2">
          <button
            onClick={handleInvoke}
            disabled={invokeMutation.isPending}
            className="flex items-center space-x-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-700 text-white rounded-lg transition"
          >
            <Play className="w-4 h-4" />
            <span>Invoke</span>
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

      {/* Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
          <div className="text-sm text-slate-400">Status</div>
          <div className="text-2xl font-bold text-white mt-1">{functionData.status}</div>
        </div>
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
          <div className="text-sm text-slate-400">Replicas</div>
          <div className="text-2xl font-bold text-white mt-1">
            {functionData.ready_replicas}/{functionData.replicas}
          </div>
        </div>
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
          <div className="text-sm text-slate-400">Available</div>
          <div className="text-2xl font-bold text-white mt-1">
            {functionData.available_replicas}
          </div>
        </div>
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
          <div className="text-sm text-slate-400">Updated</div>
          <div className="text-2xl font-bold text-white mt-1">
            {functionData.updated_replicas}
          </div>
        </div>
      </div>

      {/* Logs Section */}
      <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <Terminal className="w-5 h-5 text-slate-400" />
            <h3 className="text-lg font-semibold text-white">Logs</h3>
          </div>
          <button
            onClick={loadLogs}
            className="flex items-center space-x-2 px-3 py-1 text-sm text-slate-300 hover:text-white bg-slate-700 hover:bg-slate-600 rounded transition"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Refresh</span>
          </button>
        </div>
        
        <div className="bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300 overflow-x-auto max-h-96 overflow-y-auto">
          {logs || 'No logs available'}
        </div>
      </div>
    </div>
  )
}
