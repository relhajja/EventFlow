import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { functionsApi } from '@/services/api'
import type { CreateFunctionRequest } from '@/types'
import { ArrowLeft, Plus, Trash2 } from 'lucide-react'

export default function CreateFunction() {
  const navigate = useNavigate()
  const [envVars, setEnvVars] = useState<Array<{ key: string; value: string }>>([])
  const [commandArgs, setCommandArgs] = useState<string[]>([])

  const { register, handleSubmit, formState: { errors } } = useForm<CreateFunctionRequest>({
    defaultValues: {
      replicas: 1,
    },
  })

  const createMutation = useMutation({
    mutationFn: functionsApi.create,
    onSuccess: () => {
      navigate('/')
    },
  })

  const onSubmit = (data: CreateFunctionRequest) => {
    const env: Record<string, string> = {}
    envVars.forEach(({ key, value }) => {
      if (key && value) {
        env[key] = value
      }
    })

    const payload: CreateFunctionRequest = {
      ...data,
      replicas: Number(data.replicas) || 1, // Ensure replicas is a number
      env: Object.keys(env).length > 0 ? env : undefined,
      command: commandArgs.length > 0 ? commandArgs : undefined,
    }

    createMutation.mutate(payload)
  }

  const addEnvVar = () => {
    setEnvVars([...envVars, { key: '', value: '' }])
  }

  const removeEnvVar = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index))
  }

  const updateEnvVar = (index: number, field: 'key' | 'value', value: string) => {
    const updated = [...envVars]
    updated[index][field] = value
    setEnvVars(updated)
  }

  const addCommandArg = () => {
    setCommandArgs([...commandArgs, ''])
  }

  const removeCommandArg = (index: number) => {
    setCommandArgs(commandArgs.filter((_, i) => i !== index))
  }

  const updateCommandArg = (index: number, value: string) => {
    const updated = [...commandArgs]
    updated[index] = value
    setCommandArgs(updated)
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="mb-6">
        <button
          onClick={() => navigate('/')}
          className="flex items-center space-x-2 text-slate-400 hover:text-white transition"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Dashboard</span>
        </button>
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-lg p-8">
        <h2 className="text-2xl font-bold text-white mb-6">Create New Function</h2>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Function Name *
            </label>
            <input
              {...register('name', { 
                required: 'Name is required',
                pattern: {
                  value: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
                  message: 'Name must be lowercase alphanumeric with hyphens (e.g., my-function)'
                }
              })}
              type="text"
              className="w-full px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="my-function"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-400">{errors.name.message}</p>
            )}
            <p className="mt-1 text-xs text-slate-400">
              Use lowercase letters, numbers, and hyphens only. Must start and end with alphanumeric.
            </p>
          </div>

          {/* Image */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Container Image *
            </label>
            <input
              {...register('image', { required: 'Image is required' })}
              type="text"
              className="w-full px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="nginx:latest"
            />
            {errors.image && (
              <p className="mt-1 text-sm text-red-400">{errors.image.message}</p>
            )}
          </div>

          {/* Replicas */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Replicas
            </label>
            <input
              {...register('replicas', { 
                min: 1,
                valueAsNumber: true
              })}
              type="number"
              className="w-full px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>

          {/* Command */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-slate-300">
                Command (optional)
              </label>
              <button
                type="button"
                onClick={addCommandArg}
                className="text-sm text-primary-400 hover:text-primary-300"
              >
                + Add Argument
              </button>
            </div>
            <div className="space-y-2">
              {commandArgs.map((arg, index) => (
                <div key={index} className="flex space-x-2">
                  <input
                    type="text"
                    value={arg}
                    onChange={(e) => updateCommandArg(index, e.target.value)}
                    className="flex-1 px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                    placeholder={`Argument ${index + 1}`}
                  />
                  <button
                    type="button"
                    onClick={() => removeCommandArg(index)}
                    className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Environment Variables */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-slate-300">
                Environment Variables (optional)
              </label>
              <button
                type="button"
                onClick={addEnvVar}
                className="text-sm text-primary-400 hover:text-primary-300"
              >
                + Add Variable
              </button>
            </div>
            <div className="space-y-2">
              {envVars.map((envVar, index) => (
                <div key={index} className="flex space-x-2">
                  <input
                    type="text"
                    value={envVar.key}
                    onChange={(e) => updateEnvVar(index, 'key', e.target.value)}
                    className="flex-1 px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                    placeholder="KEY"
                  />
                  <input
                    type="text"
                    value={envVar.value}
                    onChange={(e) => updateEnvVar(index, 'value', e.target.value)}
                    className="flex-1 px-4 py-2 bg-slate-900 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                    placeholder="value"
                  />
                  <button
                    type="button"
                    onClick={() => removeEnvVar(index)}
                    className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Submit */}
          <div className="flex space-x-4 pt-4">
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="flex-1 flex items-center justify-center space-x-2 px-6 py-3 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-700 text-white rounded-lg transition"
            >
              <Plus className="w-5 h-5" />
              <span>{createMutation.isPending ? 'Creating...' : 'Create Function'}</span>
            </button>
            <button
              type="button"
              onClick={() => navigate('/')}
              className="px-6 py-3 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
            >
              Cancel
            </button>
          </div>

          {createMutation.isError && (
            <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg text-red-300 text-sm">
              Error creating function: {(createMutation.error as Error).message}
            </div>
          )}
        </form>
      </div>
    </div>
  )
}
