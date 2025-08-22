import React, { useState, useCallback, useEffect } from 'react'
import {
  User, Mail, Calendar, Shield, Key, Smartphone, Monitor,
  Globe, Bell, Save, Edit3, AlertCircle, Check, X, Eye, EyeOff,
  Settings, LogOut, Trash2, Plus
} from 'lucide-react'
import { useStore } from '@/stores'
import { authService } from '@/services/auth/authService'
import type { User as UserType, Session, SecurityEvent, UserPreferences } from '@/types/auth'

interface UserProfileProps {
  className?: string
}

type TabType = 'profile' | 'security' | 'sessions' | 'preferences' | 'activity'

export const UserProfile: React.FC<UserProfileProps> = ({ className = '' }) => {
  const { auth, logout } = useStore()
  const [activeTab, setActiveTab] = useState<TabType>('profile')
  const [isEditing, setIsEditing] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [securityEvents, setSecurityEvents] = useState<SecurityEvent[]>([])
  
  // Form states
  const [profileData, setProfileData] = useState<Partial<UserType>>({})
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  const [preferences, setPreferences] = useState<UserPreferences | null>(null)
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)

  const user = auth.user

  // Initialize profile data
  useEffect(() => {
    if (user) {
      setProfileData({
        firstName: user.firstName,
        lastName: user.lastName,
        email: user.email,
        username: user.username
      })
      setPreferences(user.preferences)
    }
  }, [user])

  // Load additional data
  useEffect(() => {
    if (activeTab === 'sessions') {
      loadSessions()
    } else if (activeTab === 'activity') {
      loadSecurityEvents()
    }
  }, [activeTab])

  const loadSessions = useCallback(async () => {
    setIsLoading(true)
    try {
      const sessionData = await authService.getSessions()
      setSessions(sessionData)
    } catch (err) {
      setError('Failed to load sessions')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const loadSecurityEvents = useCallback(async () => {
    setIsLoading(true)
    try {
      const events = await authService.getSecurityEvents()
      setSecurityEvents(events)
    } catch (err) {
      setError('Failed to load security events')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const handleProfileSave = useCallback(async () => {
    if (!profileData.firstName || !profileData.lastName || !profileData.email) {
      setError('Please fill in all required fields')
      return
    }

    setIsSaving(true)
    setError(null)

    try {
      await authService.updateProfile(profileData)
      setSuccess('Profile updated successfully')
      setIsEditing(false)
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update profile')
    } finally {
      setIsSaving(false)
    }
  }, [profileData])

  const handlePasswordChange = useCallback(async () => {
    if (!passwordData.currentPassword || !passwordData.newPassword || !passwordData.confirmPassword) {
      setError('Please fill in all password fields')
      return
    }

    if (passwordData.newPassword !== passwordData.confirmPassword) {
      setError('New passwords do not match')
      return
    }

    if (passwordData.newPassword.length < 8) {
      setError('New password must be at least 8 characters')
      return
    }

    setIsSaving(true)
    setError(null)

    try {
      await authService.changePassword(passwordData.currentPassword, passwordData.newPassword)
      setSuccess('Password changed successfully')
      setPasswordData({ currentPassword: '', newPassword: '', confirmPassword: '' })
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change password')
    } finally {
      setIsSaving(false)
    }
  }, [passwordData])

  const handleRevokeSession = useCallback(async (sessionId: string) => {
    try {
      await authService.revokeSession(sessionId)
      setSessions(prev => prev.filter(s => s.id !== sessionId))
      setSuccess('Session revoked successfully')
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to revoke session')
    }
  }, [])

  const handleRevokeAllSessions = useCallback(async () => {
    if (!confirm('This will log you out from all devices. Continue?')) return

    try {
      await authService.revokeAllSessions()
      // This will redirect to login
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to revoke all sessions')
    }
  }, [])

  const tabs = [
    { id: 'profile' as TabType, name: 'Profile', icon: User },
    { id: 'security' as TabType, name: 'Security', icon: Shield },
    { id: 'sessions' as TabType, name: 'Sessions', icon: Monitor },
    { id: 'preferences' as TabType, name: 'Preferences', icon: Settings },
    { id: 'activity' as TabType, name: 'Activity', icon: Globe }
  ]

  if (!user) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600 dark:text-gray-400">User not found</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`w-full max-w-4xl mx-auto ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="w-16 h-16 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center mr-4">
                {user.avatar ? (
                  <img src={user.avatar} alt={user.firstName} className="w-16 h-16 rounded-full" />
                ) : (
                  <User className="w-8 h-8 text-blue-600 dark:text-blue-400" />
                )}
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                  {user.firstName} {user.lastName}
                </h1>
                <p className="text-gray-600 dark:text-gray-400">@{user.username}</p>
                <p className="text-sm text-gray-500 dark:text-gray-500">{user.email}</p>
              </div>
            </div>
            <button
              onClick={() => logout()}
              className="flex items-center px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors"
            >
              <LogOut className="w-4 h-4 mr-2" />
              Sign Out
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 dark:border-gray-700">
          <nav className="flex space-x-8 px-6">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`
                  flex items-center py-4 px-1 border-b-2 font-medium text-sm transition-colors
                  ${activeTab === tab.id
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300'
                  }
                `}
              >
                <tab.icon className="w-4 h-4 mr-2" />
                {tab.name}
              </button>
            ))}
          </nav>
        </div>

        {/* Status Messages */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}

        {success && (
          <div className="mx-6 mt-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md">
            <div className="flex items-center">
              <Check className="w-4 h-4 text-green-500 mr-2" />
              <span className="text-sm text-green-700 dark:text-green-400">{success}</span>
            </div>
          </div>
        )}

        {/* Tab Content */}
        <div className="p-6">
          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                  Profile Information
                </h2>
                <button
                  onClick={() => setIsEditing(!isEditing)}
                  className="flex items-center px-3 py-2 text-sm text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-md"
                >
                  <Edit3 className="w-4 h-4 mr-2" />
                  {isEditing ? 'Cancel' : 'Edit'}
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    First Name
                  </label>
                  <input
                    type="text"
                    value={profileData.firstName || ''}
                    onChange={(e) => setProfileData(prev => ({ ...prev, firstName: e.target.value }))}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 dark:disabled:bg-gray-700 dark:bg-gray-700 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Last Name
                  </label>
                  <input
                    type="text"
                    value={profileData.lastName || ''}
                    onChange={(e) => setProfileData(prev => ({ ...prev, lastName: e.target.value }))}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 dark:disabled:bg-gray-700 dark:bg-gray-700 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Username
                  </label>
                  <input
                    type="text"
                    value={profileData.username || ''}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-50 dark:bg-gray-700 dark:text-white text-gray-500"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Username cannot be changed
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Email Address
                  </label>
                  <input
                    type="email"
                    value={profileData.email || ''}
                    onChange={(e) => setProfileData(prev => ({ ...prev, email: e.target.value }))}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 dark:disabled:bg-gray-700 dark:bg-gray-700 dark:text-white"
                  />
                </div>
              </div>

              {isEditing && (
                <div className="flex justify-end space-x-3">
                  <button
                    onClick={() => setIsEditing(false)}
                    className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleProfileSave}
                    disabled={isSaving}
                    className="flex items-center px-4 py-2 text-sm text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
                  >
                    <Save className="w-4 h-4 mr-2" />
                    {isSaving ? 'Saving...' : 'Save Changes'}
                  </button>
                </div>
              )}
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="space-y-8">
              {/* Change Password */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                  Change Password
                </h3>
                <div className="space-y-4 max-w-md">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Current Password
                    </label>
                    <div className="relative">
                      <input
                        type={showCurrentPassword ? 'text' : 'password'}
                        value={passwordData.currentPassword}
                        onChange={(e) => setPasswordData(prev => ({ ...prev, currentPassword: e.target.value }))}
                        className="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                      <button
                        type="button"
                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                        className="absolute inset-y-0 right-0 pr-3 flex items-center"
                      >
                        {showCurrentPassword ? (
                          <EyeOff className="h-4 w-4 text-gray-400" />
                        ) : (
                          <Eye className="h-4 w-4 text-gray-400" />
                        )}
                      </button>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      New Password
                    </label>
                    <div className="relative">
                      <input
                        type={showNewPassword ? 'text' : 'password'}
                        value={passwordData.newPassword}
                        onChange={(e) => setPasswordData(prev => ({ ...prev, newPassword: e.target.value }))}
                        className="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                      <button
                        type="button"
                        onClick={() => setShowNewPassword(!showNewPassword)}
                        className="absolute inset-y-0 right-0 pr-3 flex items-center"
                      >
                        {showNewPassword ? (
                          <EyeOff className="h-4 w-4 text-gray-400" />
                        ) : (
                          <Eye className="h-4 w-4 text-gray-400" />
                        )}
                      </button>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Confirm New Password
                    </label>
                    <input
                      type="password"
                      value={passwordData.confirmPassword}
                      onChange={(e) => setPasswordData(prev => ({ ...prev, confirmPassword: e.target.value }))}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                  </div>

                  <button
                    onClick={handlePasswordChange}
                    disabled={isSaving || !passwordData.currentPassword || !passwordData.newPassword || !passwordData.confirmPassword}
                    className="flex items-center px-4 py-2 text-sm text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
                  >
                    <Key className="w-4 h-4 mr-2" />
                    {isSaving ? 'Changing...' : 'Change Password'}
                  </button>
                </div>
              </div>

              {/* Two-Factor Authentication */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                  Two-Factor Authentication
                </h3>
                <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <Smartphone className="w-5 h-5 text-gray-600 dark:text-gray-400 mr-3" />
                      <div>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">
                          Authenticator App
                        </p>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {user.role.name === 'admin' ? 'Enabled' : 'Not configured'}
                        </p>
                      </div>
                    </div>
                    <button className="px-3 py-1 text-sm text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded">
                      {user.role.name === 'admin' ? 'Manage' : 'Setup'}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Sessions Tab */}
          {activeTab === 'sessions' && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                  Active Sessions
                </h2>
                <button
                  onClick={handleRevokeAllSessions}
                  className="flex items-center px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md"
                >
                  <LogOut className="w-4 h-4 mr-2" />
                  Sign Out All Devices
                </button>
              </div>

              {isLoading ? (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                </div>
              ) : (
                <div className="space-y-4">
                  {sessions.map((session) => (
                    <div
                      key={session.id}
                      className="border border-gray-200 dark:border-gray-600 rounded-lg p-4"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center">
                          <Monitor className="w-5 h-5 text-gray-600 dark:text-gray-400 mr-3" />
                          <div>
                            <p className="text-sm font-medium text-gray-900 dark:text-white">
                              {session.deviceInfo.browser} on {session.deviceInfo.platform}
                              {session.isActive && (
                                <span className="ml-2 px-2 py-1 text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-full">
                                  Current Session
                                </span>
                              )}
                            </p>
                            <p className="text-sm text-gray-600 dark:text-gray-400">
                              {session.ipAddress} • Last active: {new Date(session.lastActivity).toLocaleString()}
                            </p>
                          </div>
                        </div>
                        {!session.isActive && (
                          <button
                            onClick={() => handleRevokeSession(session.id)}
                            className="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Preferences Tab */}
          {activeTab === 'preferences' && preferences && (
            <div className="space-y-6">
              <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                Preferences
              </h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Theme
                  </label>
                  <select
                    value={preferences.theme}
                    onChange={(e) => setPreferences(prev => prev ? { ...prev, theme: e.target.value as any } : null)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    <option value="light">Light</option>
                    <option value="dark">Dark</option>
                    <option value="auto">Auto</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Language
                  </label>
                  <select
                    value={preferences.language}
                    onChange={(e) => setPreferences(prev => prev ? { ...prev, language: e.target.value } : null)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    <option value="en">English</option>
                    <option value="es">Spanish</option>
                    <option value="fr">French</option>
                    <option value="de">German</option>
                  </select>
                </div>
              </div>

              <div>
                <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                  Notifications
                </h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-700 dark:text-gray-300">Email notifications</span>
                    <input
                      type="checkbox"
                      checked={preferences.notifications.email}
                      onChange={(e) => setPreferences(prev => prev ? {
                        ...prev,
                        notifications: { ...prev.notifications, email: e.target.checked }
                      } : null)}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-700 dark:text-gray-300">Push notifications</span>
                    <input
                      type="checkbox"
                      checked={preferences.notifications.push}
                      onChange={(e) => setPreferences(prev => prev ? {
                        ...prev,
                        notifications: { ...prev.notifications, push: e.target.checked }
                      } : null)}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Activity Tab */}
          {activeTab === 'activity' && (
            <div className="space-y-6">
              <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                Security Activity
              </h2>

              {isLoading ? (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                </div>
              ) : (
                <div className="space-y-4">
                  {securityEvents.map((event) => (
                    <div
                      key={event.id}
                      className="border border-gray-200 dark:border-gray-600 rounded-lg p-4"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-start">
                          <div className={`
                            p-2 rounded-full mr-3 mt-1
                            ${event.severity === 'critical' ? 'bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-400' :
                              event.severity === 'high' ? 'bg-orange-100 dark:bg-orange-900 text-orange-600 dark:text-orange-400' :
                              event.severity === 'medium' ? 'bg-yellow-100 dark:bg-yellow-900 text-yellow-600 dark:text-yellow-400' :
                              'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400'
                            }
                          `}>
                            <Shield className="w-4 h-4" />
                          </div>
                          <div>
                            <p className="text-sm font-medium text-gray-900 dark:text-white">
                              {event.description}
                            </p>
                            <p className="text-sm text-gray-600 dark:text-gray-400">
                              {event.ipAddress} • {new Date(event.timestamp).toLocaleString()}
                            </p>
                          </div>
                        </div>
                        <span className={`
                          px-2 py-1 text-xs rounded-full
                          ${event.severity === 'critical' ? 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200' :
                            event.severity === 'high' ? 'bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200' :
                            event.severity === 'medium' ? 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200' :
                            'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200'
                          }
                        `}>
                          {event.severity}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default UserProfile