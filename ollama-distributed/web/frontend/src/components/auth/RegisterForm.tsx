import React, { useState, useCallback } from 'react'
import { Eye, EyeOff, User, Mail, Lock, Shield, AlertCircle, Loader2, CheckCircle } from 'lucide-react'
import { useStore } from '@/stores'
import type { RegisterData } from '@/types/auth'

interface RegisterFormProps {
  onSuccess?: () => void
  onLoginClick?: () => void
  className?: string
}

interface PasswordStrength {
  score: number
  feedback: string[]
  isValid: boolean
}

export const RegisterForm: React.FC<RegisterFormProps> = ({
  onSuccess,
  onLoginClick,
  className = ''
}) => {
  const { register } = useStore()
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [acceptTerms, setAcceptTerms] = useState(false)
  const [formData, setFormData] = useState<RegisterData>({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    firstName: '',
    lastName: ''
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [passwordStrength, setPasswordStrength] = useState<PasswordStrength>({
    score: 0,
    feedback: [],
    isValid: false
  })

  const validatePassword = useCallback((password: string): PasswordStrength => {
    const feedback: string[] = []
    let score = 0

    if (password.length >= 8) {
      score += 1
    } else {
      feedback.push('At least 8 characters')
    }

    if (/[a-z]/.test(password)) {
      score += 1
    } else {
      feedback.push('Include lowercase letters')
    }

    if (/[A-Z]/.test(password)) {
      score += 1
    } else {
      feedback.push('Include uppercase letters')
    }

    if (/\d/.test(password)) {
      score += 1
    } else {
      feedback.push('Include numbers')
    }

    if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      score += 1
    } else {
      feedback.push('Include special characters')
    }

    return {
      score,
      feedback,
      isValid: score >= 4
    }
  }, [])

  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {}

    // Username validation
    if (!formData.username.trim()) {
      newErrors.username = 'Username is required'
    } else if (formData.username.length < 3) {
      newErrors.username = 'Username must be at least 3 characters'
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.username)) {
      newErrors.username = 'Username can only contain letters, numbers, underscore, and hyphen'
    }

    // Email validation
    if (!formData.email.trim()) {
      newErrors.email = 'Email is required'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address'
    }

    // Password validation
    if (!formData.password) {
      newErrors.password = 'Password is required'
    } else if (!passwordStrength.isValid) {
      newErrors.password = 'Password does not meet security requirements'
    }

    // Confirm password validation
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password'
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match'
    }

    // Name validation
    if (!formData.firstName.trim()) {
      newErrors.firstName = 'First name is required'
    }

    if (!formData.lastName.trim()) {
      newErrors.lastName = 'Last name is required'
    }

    // Terms acceptance
    if (!acceptTerms) {
      newErrors.terms = 'You must accept the terms and conditions'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [formData, passwordStrength.isValid, acceptTerms])

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) return

    setIsSubmitting(true)
    setErrors({})

    try {
      await register(formData)
      onSuccess?.()
    } catch (error) {
      setErrors({
        general: error instanceof Error ? error.message : 'Registration failed. Please try again.'
      })
    } finally {
      setIsSubmitting(false)
    }
  }, [formData, validateForm, register, onSuccess])

  const handleInputChange = useCallback((field: keyof RegisterData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    
    // Update password strength for password field
    if (field === 'password') {
      setPasswordStrength(validatePassword(value))
    }
    
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }))
    }
  }, [errors, validatePassword])

  const getPasswordStrengthColor = (score: number): string => {
    if (score <= 1) return 'bg-red-500'
    if (score <= 2) return 'bg-yellow-500'
    if (score <= 3) return 'bg-blue-500'
    return 'bg-green-500'
  }

  const getPasswordStrengthText = (score: number): string => {
    if (score <= 1) return 'Weak'
    if (score <= 2) return 'Fair'
    if (score <= 3) return 'Good'
    return 'Strong'
  }

  return (
    <div className={`w-full max-w-md mx-auto ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="flex justify-center mb-4">
            <div className="p-3 bg-green-100 dark:bg-green-900 rounded-full">
              <Shield className="w-8 h-8 text-green-600 dark:text-green-400" />
            </div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Create Account
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Join Ollama Distributed to get started
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

        {/* Registration Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Name Fields */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label htmlFor="firstName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                First Name
              </label>
              <input
                id="firstName"
                name="firstName"
                type="text"
                autoComplete="given-name"
                required
                value={formData.firstName}
                onChange={(e) => handleInputChange('firstName', e.target.value)}
                className={`
                  block w-full px-3 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.firstName
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="First name"
                disabled={isSubmitting}
              />
              {errors.firstName && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                  {errors.firstName}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="lastName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Last Name
              </label>
              <input
                id="lastName"
                name="lastName"
                type="text"
                autoComplete="family-name"
                required
                value={formData.lastName}
                onChange={(e) => handleInputChange('lastName', e.target.value)}
                className={`
                  block w-full px-3 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.lastName
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Last name"
                disabled={isSubmitting}
              />
              {errors.lastName && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                  {errors.lastName}
                </p>
              )}
            </div>
          </div>

          {/* Username Field */}
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Username
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
                placeholder="Choose a username"
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

          {/* Email Field */}
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
                value={formData.email}
                onChange={(e) => handleInputChange('email', e.target.value)}
                className={`
                  block w-full pl-10 pr-3 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.email
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Enter your email"
                disabled={isSubmitting}
                data-testid="email-input"
              />
            </div>
            {errors.email && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.email}
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
                autoComplete="new-password"
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
                placeholder="Create a strong password"
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

            {/* Password Strength Indicator */}
            {formData.password && (
              <div className="mt-2">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-gray-600 dark:text-gray-400">
                    Password strength: {getPasswordStrengthText(passwordStrength.score)}
                  </span>
                  {passwordStrength.isValid && (
                    <CheckCircle className="w-4 h-4 text-green-500" />
                  )}
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-1">
                  <div
                    className={`h-1 rounded-full transition-all duration-300 ${getPasswordStrengthColor(passwordStrength.score)}`}
                    style={{ width: `${(passwordStrength.score / 5) * 100}%` }}
                  />
                </div>
                {passwordStrength.feedback.length > 0 && (
                  <ul className="mt-1 text-xs text-gray-600 dark:text-gray-400">
                    {passwordStrength.feedback.map((item, index) => (
                      <li key={index}>• {item}</li>
                    ))}
                  </ul>
                )}
              </div>
            )}

            {errors.password && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.password}
              </p>
            )}
          </div>

          {/* Confirm Password Field */}
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Confirm Password
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Lock className="h-4 w-4 text-gray-400" />
              </div>
              <input
                id="confirmPassword"
                name="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                autoComplete="new-password"
                required
                value={formData.confirmPassword}
                onChange={(e) => handleInputChange('confirmPassword', e.target.value)}
                className={`
                  block w-full pl-10 pr-10 py-2 border rounded-md text-sm
                  transition-colors duration-200
                  ${errors.confirmPassword
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                  }
                  dark:bg-gray-700 dark:border-gray-600 dark:text-white
                  focus:outline-none focus:ring-1
                `}
                placeholder="Confirm your password"
                disabled={isSubmitting}
                data-testid="confirm-password-input"
              />
              <button
                type="button"
                className="absolute inset-y-0 right-0 pr-3 flex items-center"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                disabled={isSubmitting}
                aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
              >
                {showConfirmPassword ? (
                  <EyeOff className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                ) : (
                  <Eye className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                )}
              </button>
            </div>
            {errors.confirmPassword && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.confirmPassword}
              </p>
            )}
          </div>

          {/* Terms and Conditions */}
          <div>
            <div className="flex items-start">
              <input
                id="accept-terms"
                name="accept-terms"
                type="checkbox"
                checked={acceptTerms}
                onChange={(e) => {
                  setAcceptTerms(e.target.checked)
                  if (errors.terms) {
                    setErrors(prev => ({ ...prev, terms: '' }))
                  }
                }}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded mt-0.5"
                disabled={isSubmitting}
              />
              <label htmlFor="accept-terms" className="ml-2 block text-sm text-gray-700 dark:text-gray-300">
                I agree to the{' '}
                <a href="/terms" className="text-blue-600 hover:text-blue-500 dark:text-blue-400">
                  Terms of Service
                </a>
                {' '}and{' '}
                <a href="/privacy" className="text-blue-600 hover:text-blue-500 dark:text-blue-400">
                  Privacy Policy
                </a>
              </label>
            </div>
            {errors.terms && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.terms}
              </p>
            )}
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isSubmitting || !passwordStrength.isValid}
            className={`
              w-full flex justify-center items-center py-2 px-4 border border-transparent 
              rounded-md shadow-sm text-sm font-medium text-white
              transition-colors duration-200
              ${isSubmitting || !passwordStrength.isValid
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-green-600 hover:bg-green-700 focus:ring-2 focus:ring-offset-2 focus:ring-green-500'
              }
              dark:focus:ring-offset-gray-800
            `}
            data-testid="register-button"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Creating account...
              </>
            ) : (
              'Create Account'
            )}
          </button>
        </form>

        {/* Login Link */}
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Already have an account?{' '}
            <button
              type="button"
              onClick={onLoginClick}
              className="text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300 font-medium"
              disabled={isSubmitting}
            >
              Sign in
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}

export default RegisterForm