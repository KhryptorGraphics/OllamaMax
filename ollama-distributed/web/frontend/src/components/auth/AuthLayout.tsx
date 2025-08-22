import React, { useState } from 'react'
import { Shield, ArrowLeft, Globe, Monitor, Smartphone } from 'lucide-react'
import LoginForm from './LoginForm'
import RegisterForm from './RegisterForm'
import ForgotPasswordForm from './ForgotPasswordForm'

interface AuthLayoutProps {
  initialView?: 'login' | 'register' | 'forgot-password'
  onSuccess?: () => void
  className?: string
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({
  initialView = 'login',
  onSuccess,
  className = ''
}) => {
  const [currentView, setCurrentView] = useState(initialView)

  const handleSuccess = () => {
    onSuccess?.()
  }

  const renderCurrentView = () => {
    switch (currentView) {
      case 'login':
        return (
          <LoginForm
            onSuccess={handleSuccess}
            onRegisterClick={() => setCurrentView('register')}
            onForgotPasswordClick={() => setCurrentView('forgot-password')}
          />
        )
      case 'register':
        return (
          <RegisterForm
            onSuccess={handleSuccess}
            onLoginClick={() => setCurrentView('login')}
          />
        )
      case 'forgot-password':
        return (
          <ForgotPasswordForm
            onSuccess={() => setCurrentView('login')}
            onBackToLogin={() => setCurrentView('login')}
          />
        )
      default:
        return null
    }
  }

  return (
    <div className={`min-h-screen bg-gray-50 dark:bg-gray-900 flex flex-col justify-center py-12 sm:px-6 lg:px-8 ${className}`}>
      {/* Background Pattern */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 opacity-50" />
        <div className="absolute inset-0 bg-grid-pattern opacity-5" />
      </div>

      {/* Header */}
      <div className="sm:mx-auto sm:w-full sm:max-w-md relative z-10">
        <div className="text-center">
          <div className="flex justify-center">
            <div className="p-4 bg-white dark:bg-gray-800 rounded-full shadow-lg">
              <Shield className="w-12 h-12 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
          <h1 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-white">
            Ollama Distributed
          </h1>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Distributed AI Model Management Platform
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md relative z-10">
        {currentView !== 'login' && (
          <div className="mb-4">
            <button
              onClick={() => setCurrentView('login')}
              className="flex items-center text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Sign In
            </button>
          </div>
        )}

        {renderCurrentView()}
      </div>

      {/* Security Features */}
      <div className="mt-12 sm:mx-auto sm:w-full sm:max-w-md relative z-10">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
            Enterprise Security Features
          </h3>
          <div className="grid grid-cols-3 gap-4 text-center">
            <div className="flex flex-col items-center">
              <div className="p-2 bg-green-100 dark:bg-green-900 rounded-full mb-2">
                <Shield className="w-4 h-4 text-green-600 dark:text-green-400" />
              </div>
              <span className="text-xs text-gray-600 dark:text-gray-400">End-to-End Encryption</span>
            </div>
            <div className="flex flex-col items-center">
              <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-full mb-2">
                <Monitor className="w-4 h-4 text-blue-600 dark:text-blue-400" />
              </div>
              <span className="text-xs text-gray-600 dark:text-gray-400">Multi-Factor Auth</span>
            </div>
            <div className="flex flex-col items-center">
              <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-full mb-2">
                <Globe className="w-4 h-4 text-purple-600 dark:text-purple-400" />
              </div>
              <span className="text-xs text-gray-600 dark:text-gray-400">SSO Integration</span>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="mt-8 text-center relative z-10">
        <div className="flex justify-center space-x-6 text-sm text-gray-500 dark:text-gray-400">
          <a href="/privacy" className="hover:text-gray-700 dark:hover:text-gray-300">
            Privacy Policy
          </a>
          <a href="/terms" className="hover:text-gray-700 dark:hover:text-gray-300">
            Terms of Service
          </a>
          <a href="/security" className="hover:text-gray-700 dark:hover:text-gray-300">
            Security
          </a>
          <a href="/support" className="hover:text-gray-700 dark:hover:text-gray-300">
            Support
          </a>
        </div>
        <p className="mt-2 text-xs text-gray-400 dark:text-gray-500">
          © 2025 Ollama Distributed. Built with enterprise security.
        </p>
      </div>
    </div>
  )
}

export default AuthLayout