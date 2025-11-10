import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { functionsApi } from '@/services/api'
import { ArrowLeft, Plus, X, AlertCircle, Loader2 } from 'lucide-react'

interface Runtime {
  id: string
  name: string
  image: string
  description: string
  icon: string
}

const RUNTIMES: Runtime[] = [
  {
    id: 'python',
    name: 'Python 3.11',
    image: 'python:3.11-slim',
    description: 'Python runtime for general purpose applications',
    icon: '🐍',
  },
  {
    id: 'nodejs',
    name: 'Node.js 20',
    image: 'node:20-alpine',
    description: 'JavaScript/TypeScript runtime',
    icon: '📦',
  },
  {
    id: 'go',
    name: 'Go 1.21',
    image: 'golang:1.21-alpine',
    description: 'Go runtime for high-performance applications',
    icon: '🔷',
  },
  {
    id: 'rust',
    name: 'Rust',
    image: 'rust:latest',
    description: 'Rust runtime for systems programming',
    icon: '🦀',
  },
  {
    id: 'java',
    name: 'Java 17',
    image: 'openjdk:17-slim',
    description: 'Java runtime for enterprise applications',
    icon: '☕',
  },
  {
    id: 'dotnet',
    name: '.NET 8',
    image: 'mcr.microsoft.com/dotnet/runtime:8.0',
    description: 'C# and .NET runtime',
    icon: '💠',
  },
  {
    id: 'custom',
    name: 'Custom',
    image: '',
    description: 'Use your own Docker image',
    icon: '🐳',
  },
]

export default function CreateFunction() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    name: '',
    runtime: 'python',
    customImage: '',
    replicas: 1,
    env: [] as { key: string; value: string }[],
    command: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  const createMutation = useMutation({
    mutationFn: functionsApi.create,
    onSuccess: () => {
      navigate('/dashboard')
    },
    onError: (error: any) => {
      setErrors({ general: error.message || 'Failed to create function' })
    },
  })

  const selectedRuntime = RUNTIMES.find((r) => r.id === formData.runtime)

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    } else if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(formData.name)) {
      newErrors.name = 'Name must be lowercase alphanumeric with hyphens'
    }

    if (formData.runtime === 'custom' && !formData.customImage.trim()) {
      newErrors.customImage = 'Custom image is required'
    }

    if (formData.replicas < 0 || formData.replicas > 10) {
      newErrors.replicas = 'Replicas must be between 0 and 10'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) return

    const image = formData.runtime === 'custom' 
      ? formData.customImage 
      : selectedRuntime?.image || ''

    const envObj: Record<string, string> = {}
    formData.env.forEach((item) => {
      if (item.key && item.value) {
        envObj[item.key] = item.value
      }
    })

    const command = formData.command 
      ? formData.command.split(' ').filter((c) => c.trim())
      : undefined

    createMutation.mutate({
      name: formData.name,
      image,
      replicas: formData.replicas,
      env: envObj,
      command,
    })
  }

  const addEnvVar = () => {
    setFormData({
      ...formData,
      env: [...formData.env, { key: '', value: '' }],
    })
  }

  const removeEnvVar = (index: number) => {
    setFormData({
      ...formData,
      env: formData.env.filter((_, i) => i !== index),
    })
  }

  const updateEnvVar = (index: number, field: 'key' | 'value', value: string) => {
    const newEnv = [...formData.env]
    newEnv[index][field] = value
    setFormData({ ...formData, env: newEnv })
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center space-x-4">
        <button
          onClick={() => navigate('/dashboard')}
          className="p-2 hover:bg-slate-700 rounded-lg transition"
        >
          <ArrowLeft className="w-5 h-5 text-slate-400" />
        </button>
        <div>
          <h2 className="text-3xl font-bold text-white">Create Function</h2>
          <p className="text-slate-400 mt-1">Deploy a new serverless function</p>
        </div>
      </div>

      {errors.general && (
        <div className="bg-red-900/20 border border-red-500 rounded-lg p-4 flex items-start space-x-3">
          <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="text-red-400 font-semibold">Error</h3>
            <p className="text-red-300 text-sm mt-1">{errors.general}</p>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Function Details</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Function Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value.toLowerCase() })}
                placeholder="my-function"
                className={`w-full px-4 py-2 bg-slate-900 border rounded-lg text-white ${errors.name ? 'border-red-500' : 'border-slate-700'}`}
              />
              {errors.name && <p className="text-red-400 text-sm mt-1">{errors.name}</p>}
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Replicas *</label>
              <input
                type="number"
                value={formData.replicas}
                onChange={(e) => setFormData({ ...formData, replicas: parseInt(e.target.value) || 0 })}
                min="0"
                max="10"
                className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white"
              />
            </div>
          </div>
        </div>

        <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Select Runtime</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {RUNTIMES.map((runtime) => (
              <button
                key={runtime.id}
                type="button"
                onClick={() => setFormData({ ...formData, runtime: runtime.id })}
                className={`p-4 border-2 rounded-lg transition text-left ${formData.runtime === runtime.id ? 'border-primary-500 bg-primary-500/10' : 'border-slate-700 bg-slate-900'}`}
              >
                <div className="flex items-start space-x-3">
                  <span className="text-3xl">{runtime.icon}</span>
                  <div>
                    <h4 className="text-white font-semibold">{runtime.name}</h4>
                    <p className="text-slate-400 text-sm">{runtime.description}</p>
                  </div>
                </div>
              </button>
            ))}
          </div>
          {formData.runtime === 'custom' && (
            <div className="mt-4">
              <input
                type="text"
                value={formData.customImage}
                onChange={(e) => setFormData({ ...formData, customImage: e.target.value })}
                placeholder="my-registry/my-image:tag"
                className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white"
              />
            </div>
          )}
        </div>

        <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white">Environment Variables</h3>
            <button type="button" onClick={addEnvVar} className="px-3 py-1 bg-slate-700 text-white text-sm rounded">
              <Plus className="w-4 h-4 inline" /> Add
            </button>
          </div>
          <div className="space-y-3">
            {formData.env.map((envVar, index) => (
              <div key={index} className="flex items-center space-x-2">
                <input
                  type="text"
                  value={envVar.key}
                  onChange={(e) => updateEnvVar(index, 'key', e.target.value)}
                  placeholder="KEY"
                  className="flex-1 px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm"
                />
                <span className="text-slate-500">=</span>
                <input
                  type="text"
                  value={envVar.value}
                  onChange={(e) => updateEnvVar(index, 'value', e.target.value)}
                  placeholder="value"
                  className="flex-1 px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white text-sm"
                />
                <button type="button" onClick={() => removeEnvVar(index)} className="p-2 text-red-400">
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end space-x-3">
          <button
            type="button"
            onClick={() => navigate('/dashboard')}
            className="px-6 py-2 bg-slate-700 text-white rounded-lg"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="flex items-center space-x-2 px-6 py-2 bg-primary-600 text-white rounded-lg"
          >
            {createMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Creating...</span>
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                <span>Create Function</span>
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  )
}
