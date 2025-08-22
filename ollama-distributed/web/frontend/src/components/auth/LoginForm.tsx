import React, { useState, useCallback } from 'react'
import { Eye, EyeOff, Lock, User, Shield, AlertCircle, Loader2 } from 'lucide-react'
import { useStore } from '@/stores'
import type { LoginCredentials } from '@/types/auth'

interface LoginFormProps {
  onSuccess?: () => void
  onRegisterClick?: () => void
  onForgotPasswordClick?: () => void
  className?: string
}

export const LoginForm: React.FC<LoginFormProps> = ({
  onSuccess,
  onRegisterClick,
  onForgotPasswordClick,
  className = ''
}) => {
  const { login, auth } = useStore()
  const [showPassword, setShowPassword] = useState(false)
  const [rememberMe, setRememberMe] = useState(false)
  const [formData, setFormData] = useState<LoginCredentials>({
    username: '',
    password: '',
    rememberMe: false
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {}

    if (!formData.username.trim()) {
      newErrors.username = 'Username or email is required'
    }

    if (!formData.password) {
      newErrors.password = 'Password is required'
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [formData])

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) return

    setIsSubmitting(true)
    setErrors({})

    try {
      await login({
        ...formData,
        rememberMe
      })
      
      onSuccess?.()
    } catch (error) {
      setErrors({
        general: error instanceof Error ? error.message : 'Login failed. Please try again.'
      })
    } finally {
      setIsSubmitting(false)
    }
  }, [formData, rememberMe, validateForm, login, onSuccess])

  const handleInputChange = useCallback((field: keyof LoginCredentials, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }))
    }
  }, [errors])

  return (
    <div className={`w-full max-w-md mx-auto ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="flex justify-center mb-4">
            <div className="p-3 bg-blue-100 dark:bg-blue-900 rounded-full">
              <Shield className="w-8 h-8 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Welcome Back
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Sign in to your Ollama Distributed account
          </p>
        </div>

        {/* General Error */}
        {errors.general && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">
                {errors.general}
              </span>
            </div>
          </div>
        )}

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Username/Email Field */}
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Username or Email
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <User className="h-4 w-4 text-gray-400" />
              </div>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                required
                value={formData.username}
                onChange={(e) => handleInputChange('username', e.target.value)}
                className={`
                  block w-full pl-10 pr-3 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.username
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Enter your username or email"
                disabled={isSubmitting}
                data-testid="username-input"
              />
            </div>
            {errors.username && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.username}
              </p>
            )}
          </div>

          {/* Password Field */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Password
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Lock className="h-4 w-4 text-gray-400" />
              </div>
              <input
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                required
                value={formData.password}
                onChange={(e) => handleInputChange('password', e.target.value)}
                className={`
                  block w-full pl-10 pr-10 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.password
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Enter your password"
                disabled={isSubmitting}
                data-testid="password-input"
              />
              <button
                type="button"
                className="absolute inset-y-0 right-0 pr-3 flex items-center"
                onClick={() => setShowPassword(!showPassword)}
                disabled={isSubmitting}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                ) : (
                  <Eye className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                )}
              </button>
            </div>
            {errors.password && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.password}
              </p>
            )}
          </div>

          {/* Remember Me & Forgot Password */}
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="remember-me"
                name="remember-me"
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                disabled={isSubmitting}
              />
              <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-700 dark:text-gray-300">
                Remember me
              </label>
            </div>

            <button
              type="button"
              onClick={onForgotPasswordClick}
              className="text-sm text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300"
              disabled={isSubmitting}
            >
              Forgot password?
            </button>
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isSubmitting}
            className={`
              w-full flex justify-center items-center py-2 px-4 border border-transparent 
              rounded-md shadow-sm text-sm font-medium text-white
              transition-colors duration-200
              ${isSubmitting
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700 focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
              }
              dark:focus:ring-offset-gray-800
            `}
            data-testid="login-button"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Signing in...
              </>
            ) : (
              'Sign In'
            )}
          </button>
        </form>

        {/* Register Link */}
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Don't have an account?{' '}
            <button
              type="button"
              onClick={onRegisterClick}
              className="text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300 font-medium"
              disabled={isSubmitting}
            >
              Sign up
            </button>
          </p>
        </div>

        {/* SSO Options */}
        <div className="mt-6">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300 dark:border-gray-600" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400">
                Or continue with
              </span>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-3">
            <button
              type="button"
              className="w-full inline-flex justify-center py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors duration-200"
              disabled={isSubmitting}
            >
              <span>Google</span>
            </button>
            <button
              type="button"
              className="w-full inline-flex justify-center py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors duration-200"
              disabled={isSubmitting}
            >
              <span>GitHub</span>
            </button>
          </div>
        </div>
      </div>

      {/* Loading Overlay */}
      {auth.loading && (
        <div className="absolute inset-0 bg-black bg-opacity-20 flex items-center justify-center rounded-lg">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-lg">
            <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
          </div>
        </div>
      )}
    </div>
  )
}

export default LoginForm