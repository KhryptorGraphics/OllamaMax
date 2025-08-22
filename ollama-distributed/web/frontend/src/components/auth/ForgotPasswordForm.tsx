import React, { useState, useCallback } from 'react'
import { Mail, ArrowLeft, Send, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'
import { authService } from '@/services/auth/authService'

interface ForgotPasswordFormProps {
  onSuccess?: () => void
  onBackToLogin?: () => void
  className?: string
}

export const ForgotPasswordForm: React.FC<ForgotPasswordFormProps> = ({
  onSuccess,
  onBackToLogin,
  className = ''
}) => {
  const [email, setEmail] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSubmitted, setIsSubmitted] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const validateEmail = useCallback((email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(email)
  }, [])

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!email.trim()) {
      setError('Email address is required')
      return
    }

    if (!validateEmail(email)) {
      setError('Please enter a valid email address')
      return
    }

    setIsSubmitting(true)
    setError(null)

    try {
      await authService.resetPassword(email)
      setIsSubmitted(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send reset email. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }, [email, validateEmail])

  const handleResend = useCallback(async () => {
    setIsSubmitting(true)
    setError(null)

    try {
      await authService.resetPassword(email)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resend email. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }, [email])

  if (isSubmitted) {
    return (
      <div className={`w-full max-w-md mx-auto ${className}`}>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-gray-200 dark:border-gray-700">
          {/* Success Header */}
          <div className="text-center mb-6">
            <div className="flex justify-center mb-4">
              <div className="p-3 bg-green-100 dark:bg-green-900 rounded-full">
                <CheckCircle className="w-8 h-8 text-green-600 dark:text-green-400" />
              </div>
            </div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
              Check Your Email
            </h1>
            <p className="text-gray-600 dark:text-gray-400">
              We've sent a password reset link to
            </p>
            <p className="text-sm font-medium text-gray-900 dark:text-white mt-1">
              {email}
            </p>
          </div>

          {/* Instructions */}
          <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4 mb-6">
            <div className="flex items-start">
              <Mail className="w-5 h-5 text-blue-500 mt-0.5 mr-3" />
              <div className="text-sm text-blue-700 dark:text-blue-300">
                <p className="font-medium mb-1">Next steps:</p>
                <ol className="list-decimal list-inside space-y-1">
                  <li>Check your email inbox (and spam folder)</li>
                  <li>Click the reset link in the email</li>
                  <li>Create a new password</li>
                  <li>Sign in with your new password</li>
                </ol>
              </div>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
              <div className="flex items-center">
                <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
                <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="space-y-3">
            <button
              onClick={handleResend}
              disabled={isSubmitting}
              className={`
                w-full flex justify-center items-center py-2 px-4 border border-gray-300 dark:border-gray-600
                rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300
                transition-colors duration-200
                ${isSubmitting
                  ? 'bg-gray-100 cursor-not-allowed'
                  : 'bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'
                }
              `}
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Resending...
                </>
              ) : (
                <>
                  <Send className="w-4 h-4 mr-2" />
                  Resend Email
                </>
              )}
            </button>

            <button
              onClick={onBackToLogin}
              className="w-full flex justify-center items-center py-2 px-4 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300"
              disabled={isSubmitting}
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Sign In
            </button>
          </div>

          {/* Additional Help */}
          <div className="mt-6 text-center border-t border-gray-200 dark:border-gray-700 pt-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Didn't receive the email?{' '}
              <a href="/support" className="text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300">
                Contact support
              </a>
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`w-full max-w-md mx-auto ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="flex justify-center mb-4">
            <div className="p-3 bg-orange-100 dark:bg-orange-900 rounded-full">
              <Mail className="w-8 h-8 text-orange-600 dark:text-orange-400" />
            </div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Reset Password
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Enter your email address and we'll send you a link to reset your password
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}

        {/* Reset Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Email Address
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Mail className="h-4 w-4 text-gray-400" />
              </div>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value)
                  if (error) setError(null)
                }}
                className={`
                  block w-full pl-10 pr-3 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${error
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Enter your email address"
                disabled={isSubmitting}
                data-testid="email-input"
              />
            </div>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              We'll never share your email with anyone else
            </p>
          </div>

          <button
            type="submit"
            disabled={isSubmitting || !email.trim()}
            className={`
              w-full flex justify-center items-center py-2 px-4 border border-transparent 
              rounded-md shadow-sm text-sm font-medium text-white
              transition-colors duration-200
              ${isSubmitting || !email.trim()
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-orange-600 hover:bg-orange-700 focus:ring-2 focus:ring-offset-2 focus:ring-orange-500'
              }
              dark:focus:ring-offset-gray-800
            `}
            data-testid="reset-button"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Sending Reset Link...
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" />
                Send Reset Link
              </>
            )}
          </button>
        </form>

        {/* Back to Login */}
        <div className="mt-6 text-center">
          <button
            onClick={onBackToLogin}
            className="flex items-center justify-center w-full text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
            disabled={isSubmitting}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Sign In
          </button>
        </div>

        {/* Security Notice */}
        <div className="mt-6 bg-gray-50 dark:bg-gray-700 rounded-lg p-3">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <AlertCircle className="w-4 h-4 text-gray-400 mt-0.5" />
            </div>
            <div className="ml-3">
              <h3 className="text-xs font-medium text-gray-900 dark:text-white">
                Security Notice
              </h3>
              <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                For security reasons, we don't reveal whether an email address is registered in our system. 
                If you don't receive an email, please check your spam folder or contact support.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ForgotPasswordForm