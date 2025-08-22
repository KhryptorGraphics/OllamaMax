import React, { useState, useCallback, useEffect } from 'react'
import { QrCode, Smartphone, Mail, Shield, Key, Copy, Check, AlertCircle, Loader2 } from 'lucide-react'
import { authService } from '@/services/auth/authService'
import type { MFAMethod } from '@/types/auth'

interface MFASetupProps {
  onComplete?: () => void
  onCancel?: () => void
  className?: string
}

interface TOTPSetupData {
  secret: string
  qrCode: string
  backupCodes: string[]
}

type MFAType = 'totp' | 'sms' | 'email'

export const MFASetup: React.FC<MFASetupProps> = ({
  onComplete,
  onCancel,
  className = ''
}) => {
  const [selectedType, setSelectedType] = useState<MFAType>('totp')
  const [step, setStep] = useState<'select' | 'setup' | 'verify' | 'backup'>('select')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [totpData, setTotpData] = useState<TOTPSetupData | null>(null)
  const [verificationCode, setVerificationCode] = useState('')
  const [challengeId, setChallengeId] = useState<string | null>(null)
  const [phoneNumber, setPhoneNumber] = useState('')
  const [copiedBackupCodes, setCopiedBackupCodes] = useState(false)

  const mfaOptions: Array<{
    type: MFAType
    name: string
    description: string
    icon: React.ReactNode
    recommended?: boolean
  }> = [
    {
      type: 'totp',
      name: 'Authenticator App',
      description: 'Use an app like Google Authenticator or Authy',
      icon: <Smartphone className="w-6 h-6" />,
      recommended: true
    },
    {
      type: 'sms',
      name: 'SMS',
      description: 'Receive codes via text message',
      icon: <Mail className="w-6 h-6" />
    },
    {
      type: 'email',
      name: 'Email',
      description: 'Receive codes via email',
      icon: <Mail className="w-6 h-6" />
    }
  ]

  const handleSetupMFA = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await authService.setupMFA(selectedType)
      
      if (selectedType === 'totp') {
        setTotpData(response)
      } else if (selectedType === 'sms') {
        setChallengeId(response.challengeId)
      } else if (selectedType === 'email') {
        setChallengeId(response.challengeId)
      }
      
      setStep('setup')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to setup MFA')
    } finally {
      setIsLoading(false)
    }
  }, [selectedType])

  const handleVerifyCode = useCallback(async () => {
    if (!verificationCode.trim() || !challengeId) return

    setIsLoading(true)
    setError(null)

    try {
      await authService.verifyMFA(challengeId, verificationCode)
      
      if (selectedType === 'totp' && totpData) {
        setStep('backup')
      } else {
        onComplete?.()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid verification code')
    } finally {
      setIsLoading(false)
    }
  }, [verificationCode, challengeId, selectedType, totpData, onComplete])

  const handleCopySecret = useCallback(() => {
    if (totpData?.secret) {
      navigator.clipboard.writeText(totpData.secret)
    }
  }, [totpData?.secret])

  const handleCopyBackupCodes = useCallback(() => {
    if (totpData?.backupCodes) {
      const codes = totpData.backupCodes.join('\n')
      navigator.clipboard.writeText(codes)
      setCopiedBackupCodes(true)
      setTimeout(() => setCopiedBackupCodes(false), 2000)
    }
  }, [totpData?.backupCodes])

  const handleNext = useCallback(() => {
    if (step === 'select') {
      handleSetupMFA()
    } else if (step === 'setup') {
      setStep('verify')
    } else if (step === 'verify') {
      handleVerifyCode()
    } else if (step === 'backup') {
      onComplete?.()
    }
  }, [step, handleSetupMFA, handleVerifyCode, onComplete])

  // Auto-focus verification code input
  useEffect(() => {
    if (step === 'verify') {
      const input = document.getElementById('verification-code')
      input?.focus()
    }
  }, [step])

  return (
    <div className={`w-full max-w-lg mx-auto ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center">
            <div className="p-3 bg-blue-100 dark:bg-blue-900 rounded-full mr-4">
              <Shield className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                Setup Two-Factor Authentication
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Add an extra layer of security to your account
              </p>
            </div>
          </div>
        </div>

        <div className="p-6">
          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
              <div className="flex items-center">
                <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
                <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
              </div>
            </div>
          )}

          {/* Step 1: Select MFA Type */}
          {step === 'select' && (
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                Choose your authentication method
              </h3>
              
              <div className="space-y-3">
                {mfaOptions.map((option) => (
                  <div
                    key={option.type}
                    className={`
                      relative rounded-lg border-2 p-4 cursor-pointer transition-colors
                      ${selectedType === option.type
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                        : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                      }
                    `}
                    onClick={() => setSelectedType(option.type)}
                  >
                    <div className="flex items-start">
                      <div className="flex-shrink-0">
                        <div className={`
                          p-2 rounded-lg
                          ${selectedType === option.type
                            ? 'bg-blue-100 dark:bg-blue-800 text-blue-600 dark:text-blue-400'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                          }
                        `}>
                          {option.icon}
                        </div>
                      </div>
                      <div className="ml-3 flex-1">
                        <div className="flex items-center">
                          <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                            {option.name}
                          </h4>
                          {option.recommended && (
                            <span className="ml-2 px-2 py-1 text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-full">
                              Recommended
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {option.description}
                        </p>
                      </div>
                      <div className="ml-3">
                        <input
                          type="radio"
                          checked={selectedType === option.type}
                          onChange={() => setSelectedType(option.type)}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Step 2: Setup Instructions */}
          {step === 'setup' && (
            <div className="space-y-6">
              {selectedType === 'totp' && totpData && (
                <>
                  <div>
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                      Scan QR Code
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                      Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                    </p>
                    
                    <div className="flex justify-center mb-4">
                      <div className="p-4 bg-white rounded-lg">
                        <img
                          src={totpData.qrCode}
                          alt="QR Code for TOTP setup"
                          className="w-48 h-48"
                        />
                      </div>
                    </div>
                    
                    <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                      <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                        Can't scan? Enter this code manually:
                      </p>
                      <div className="flex items-center justify-between bg-white dark:bg-gray-800 rounded border p-2">
                        <code className="text-sm font-mono text-gray-900 dark:text-white break-all">
                          {totpData.secret}
                        </code>
                        <button
                          type="button"
                          onClick={handleCopySecret}
                          className="ml-2 p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                          title="Copy secret key"
                        >
                          <Copy className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </>
              )}

              {selectedType === 'sms' && (
                <div>
                  <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                    Phone Number
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                    Enter your phone number to receive verification codes via SMS
                  </p>
                  <input
                    type="tel"
                    value={phoneNumber}
                    onChange={(e) => setPhoneNumber(e.target.value)}
                    placeholder="+1 (555) 123-4567"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  />
                </div>
              )}

              {selectedType === 'email' && (
                <div>
                  <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                    Email Verification
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Verification codes will be sent to your registered email address.
                  </p>
                </div>
              )}
            </div>
          )}

          {/* Step 3: Verify Code */}
          {step === 'verify' && (
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                Enter Verification Code
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {selectedType === 'totp' && 'Enter the 6-digit code from your authenticator app'}
                {selectedType === 'sms' && 'Enter the code sent to your phone'}
                {selectedType === 'email' && 'Enter the code sent to your email'}
              </p>
              
              <div>
                <input
                  id="verification-code"
                  type="text"
                  value={verificationCode}
                  onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="000000"
                  className="w-full px-4 py-3 text-center text-xl font-mono border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  maxLength={6}
                  autoComplete="one-time-code"
                />
              </div>
            </div>
          )}

          {/* Step 4: Backup Codes */}
          {step === 'backup' && totpData && (
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                Save Backup Codes
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Save these backup codes in a safe place. You can use them to access your account if you lose access to your authenticator app.
              </p>
              
              <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    Backup Codes
                  </span>
                  <button
                    type="button"
                    onClick={handleCopyBackupCodes}
                    className={`
                      flex items-center px-3 py-1 text-sm rounded-md transition-colors
                      ${copiedBackupCodes
                        ? 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300'
                        : 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 hover:bg-blue-200 dark:hover:bg-blue-800'
                      }
                    `}
                  >
                    {copiedBackupCodes ? (
                      <>
                        <Check className="w-4 h-4 mr-1" />
                        Copied!
                      </>
                    ) : (
                      <>
                        <Copy className="w-4 h-4 mr-1" />
                        Copy All
                      </>
                    )}
                  </button>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {totpData.backupCodes.map((code, index) => (
                    <div
                      key={index}
                      className="bg-white dark:bg-gray-800 rounded border p-2 text-center"
                    >
                      <code className="text-sm font-mono text-gray-900 dark:text-white">
                        {code}
                      </code>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3">
                <div className="flex">
                  <AlertCircle className="w-5 h-5 text-yellow-400 mt-0.5 mr-2" />
                  <div className="text-sm text-yellow-700 dark:text-yellow-300">
                    <strong>Important:</strong> Each backup code can only be used once. Store them securely and treat them like passwords.
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-between pt-6 border-t border-gray-200 dark:border-gray-700">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
              disabled={isLoading}
            >
              Cancel
            </button>
            
            <div className="flex space-x-3">
              {step !== 'select' && step !== 'backup' && (
                <button
                  type="button"
                  onClick={() => {
                    if (step === 'verify') setStep('setup')
                    else if (step === 'setup') setStep('select')
                  }}
                  className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md"
                  disabled={isLoading}
                >
                  Back
                </button>
              )}
              
              <button
                type="button"
                onClick={handleNext}
                disabled={
                  isLoading ||
                  (step === 'verify' && verificationCode.length !== 6) ||
                  (step === 'setup' && selectedType === 'sms' && !phoneNumber.trim())
                }
                className={`
                  flex items-center px-4 py-2 text-sm font-medium text-white rounded-md
                  transition-colors duration-200
                  ${isLoading ||
                    (step === 'verify' && verificationCode.length !== 6) ||
                    (step === 'setup' && selectedType === 'sms' && !phoneNumber.trim())
                    ? 'bg-gray-400 cursor-not-allowed'
                    : 'bg-blue-600 hover:bg-blue-700 focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
                  }
                `}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    {step === 'select' ? 'Setting up...' : 'Verifying...'}
                  </>
                ) : (
                  <>
                    {step === 'select' && 'Continue'}
                    {step === 'setup' && 'Next'}
                    {step === 'verify' && 'Verify'}
                    {step === 'backup' && 'Complete Setup'}
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MFASetup