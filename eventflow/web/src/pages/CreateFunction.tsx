import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { functionsApi } from '@/services/api'
import { ArrowLeft, AlertCircle } from 'lucide-react'
import CodeEditor from '@/components/CodeEditor'
import { DeploymentTypeSelector, GitForm, ImageForm } from '@/components/DeploymentForms'
import type { CreateFunctionRequest } from '@/types'

export default function CreateFunction() {
  const navigate = useNavigate()
  const [deploymentType, setDeploymentType] = useState<'git' | 'code' | 'image'>('image')
  const [functionName, setFunctionName] = useState('')
  const [replicas, setReplicas] = useState(1)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const createMutation = useMutation({
    mutationFn: functionsApi.create,
    onSuccess: () => {
      alert(`Function "${functionName}" created successfully!`)
      navigate('/dashboard')
    },
    onError: (error: any) => {
      setErrors({ general: error.message || 'Failed to create function' })
    },
  })

  const validateName = () => {
    if (!functionName.trim()) {
      setErrors({ name: 'Name is required' })
      return false
    }
    if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(functionName)) {
      setErrors({ name: 'Name must be lowercase alphanumeric with hyphens' })
      return false
    }
    setErrors({})
    return true
  }

  const handleCodeDeploy = async (runtime: string, code: string) => {
    if (!validateName()) return

    const request: CreateFunctionRequest = {
      name: functionName,
      deployment_type: 'code',
      runtime,
      source_code: btoa(code), // Base64 encode
      replicas,
    }

    createMutation.mutate(request)
  }

  const handleGitDeploy = async (gitConfig: any) => {
    if (!validateName()) return

    const request: CreateFunctionRequest = {
      name: functionName,
      deployment_type: 'git',
      git_config: gitConfig,
      replicas,
    }

    createMutation.mutate(request)
  }

  const handleImageDeploy = async (image: string) => {
    if (!validateName()) return

    const request: CreateFunctionRequest = {
      name: functionName,
      deployment_type: 'image',
      image,
      replicas,
    }

    createMutation.mutate(request)
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

      <div className="space-y-6">
        {/* Function Name and Replicas */}
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Function Details</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Function Name *</label>
              <input
                type="text"
                value={functionName}
                onChange={(e) => setFunctionName(e.target.value.toLowerCase())}
                placeholder="my-function"
                className={`w-full px-4 py-2 bg-slate-900 border rounded-lg text-white ${errors.name ? 'border-red-500' : 'border-slate-700'}`}
              />
              {errors.name && <p className="text-red-400 text-sm mt-1">{errors.name}</p>}
              <p className="text-slate-500 text-xs mt-1">Must be lowercase alphanumeric with hyphens</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Replicas</label>
              <input
                type="number"
                value={replicas}
                onChange={(e) => setReplicas(parseInt(e.target.value) || 1)}
                min="1"
                max="10"
                className="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white"
              />
              <p className="text-slate-500 text-xs mt-1">Number of pod replicas (1-10)</p>
            </div>
          </div>
        </div>

        {/* Deployment Type Selector */}
        <DeploymentTypeSelector selected={deploymentType} onSelect={setDeploymentType} />

        {/* Conditional Forms Based on Deployment Type */}
        {deploymentType === 'code' && (
          <CodeEditor onDeploy={handleCodeDeploy} />
        )}

        {deploymentType === 'git' && (
          <GitForm onSubmit={handleGitDeploy} />
        )}

        {deploymentType === 'image' && (
          <ImageForm onSubmit={handleImageDeploy} />
        )}

        {/* Cancel Button */}
        <div className="flex justify-start">
          <button
            type="button"
            onClick={() => navigate('/dashboard')}
            className="px-6 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}
