import React, { useState, useCallback, useEffect } from 'react'
import {
  Users, Plus, Search, Filter, MoreVertical, Edit, Trash2, Shield, Eye, EyeOff,
  CheckCircle, XCircle, Clock, AlertTriangle, Ban, UserCheck, Download, Mail,
  Key, Settings, Calendar, Activity
} from 'lucide-react'
import { apiClient } from '@/services/api/client'
import type { User, UserRole, Permission } from '@/types/auth'

interface UserManagementProps {
  className?: string
}

interface UserFilter {
  search: string
  status: 'all' | 'active' | 'inactive' | 'suspended' | 'pending'
  role: string
  sortBy: 'name' | 'email' | 'created' | 'lastLogin'
  sortOrder: 'asc' | 'desc'
}

interface CreateUserData {
  firstName: string
  lastName: string
  email: string
  username: string
  roleId: string
  sendInvite: boolean
}

export const UserManagement: React.FC<UserManagementProps> = ({ className = '' }) => {
  const [users, setUsers] = useState<User[]>([])
  const [roles, setRoles] = useState<UserRole[]>([])
  const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [createUserData, setCreateUserData] = useState<CreateUserData>({
    firstName: '',
    lastName: '',
    email: '',
    username: '',
    roleId: '',
    sendInvite: true
  })

  const [filter, setFilter] = useState<UserFilter>({
    search: '',
    status: 'all',
    role: 'all',
    sortBy: 'name',
    sortOrder: 'asc'
  })

  // Load initial data
  useEffect(() => {
    loadUsers()
    loadRoles()
  }, [])

  const loadUsers = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await apiClient.request<User[]>('/admin/users')
      if (response.success) {
        setUsers(response.data)
      }
    } catch (err) {
      setError('Failed to load users')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const loadRoles = useCallback(async () => {
    try {
      const response = await apiClient.request<UserRole[]>('/admin/roles')
      if (response.success) {
        setRoles(response.data)
      }
    } catch (err) {
      console.error('Failed to load roles:', err)
    }
  }, [])

  const handleCreateUser = useCallback(async () => {
    if (!createUserData.firstName || !createUserData.lastName || !createUserData.email || !createUserData.roleId) {
      setError('Please fill in all required fields')
      return
    }

    try {
      const response = await apiClient.request<User>('/admin/users', {
        method: 'POST',
        body: createUserData
      })

      if (response.success) {
        setUsers(prev => [...prev, response.data])
        setShowCreateModal(false)
        setCreateUserData({
          firstName: '',
          lastName: '',
          email: '',
          username: '',
          roleId: '',
          sendInvite: true
        })
        setSuccess('User created successfully')
        setTimeout(() => setSuccess(null), 3000)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create user')
    }
  }, [createUserData])

  const handleUpdateUser = useCallback(async (userId: string, updates: Partial<User>) => {
    try {
      const response = await apiClient.request<User>(`/admin/users/${userId}`, {
        method: 'PUT',
        body: updates
      })

      if (response.success) {
        setUsers(prev => prev.map(user => 
          user.id === userId ? { ...user, ...response.data } : user
        ))
        setSuccess('User updated successfully')
        setTimeout(() => setSuccess(null), 3000)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update user')
    }
  }, [])

  const handleDeleteUser = useCallback(async (userId: string) => {
    if (!confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
      return
    }

    try {
      await apiClient.request(`/admin/users/${userId}`, {
        method: 'DELETE'
      })

      setUsers(prev => prev.filter(user => user.id !== userId))
      setSuccess('User deleted successfully')
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user')
    }
  }, [])

  const handleBulkAction = useCallback(async (action: 'activate' | 'deactivate' | 'delete') => {
    if (selectedUsers.size === 0) return

    const userIds = Array.from(selectedUsers)
    
    if (action === 'delete') {
      if (!confirm(`Are you sure you want to delete ${userIds.length} user(s)? This action cannot be undone.`)) {
        return
      }
    }

    try {
      await apiClient.request('/admin/users/bulk', {
        method: 'POST',
        body: {
          action,
          userIds
        }
      })

      if (action === 'delete') {
        setUsers(prev => prev.filter(user => !selectedUsers.has(user.id)))
      } else {
        const newStatus = action === 'activate' ? 'active' : 'inactive'
        setUsers(prev => prev.map(user => 
          selectedUsers.has(user.id) 
            ? { ...user, status: newStatus as any }
            : user
        ))
      }

      setSelectedUsers(new Set())
      setSuccess(`Successfully ${action}d ${userIds.length} user(s)`)
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${action} users`)
    }
  }, [selectedUsers])

  const filteredUsers = users.filter(user => {
    if (filter.search && !user.firstName.toLowerCase().includes(filter.search.toLowerCase()) &&
        !user.lastName.toLowerCase().includes(filter.search.toLowerCase()) &&
        !user.email.toLowerCase().includes(filter.search.toLowerCase()) &&
        !user.username.toLowerCase().includes(filter.search.toLowerCase())) {
      return false
    }

    if (filter.status !== 'all' && user.status !== filter.status) {
      return false
    }

    if (filter.role !== 'all' && user.role.id !== filter.role) {
      return false
    }

    return true
  }).sort((a, b) => {
    let aVal: any, bVal: any

    switch (filter.sortBy) {
      case 'name':
        aVal = `${a.firstName} ${a.lastName}`.toLowerCase()
        bVal = `${b.firstName} ${b.lastName}`.toLowerCase()
        break
      case 'email':
        aVal = a.email.toLowerCase()
        bVal = b.email.toLowerCase()
        break
      case 'created':
        aVal = new Date(a.createdAt)
        bVal = new Date(b.createdAt)
        break
      case 'lastLogin':
        aVal = a.lastLogin ? new Date(a.lastLogin) : new Date(0)
        bVal = b.lastLogin ? new Date(b.lastLogin) : new Date(0)
        break
      default:
        return 0
    }

    if (filter.sortOrder === 'asc') {
      return aVal < bVal ? -1 : aVal > bVal ? 1 : 0
    } else {
      return aVal > bVal ? -1 : aVal < bVal ? 1 : 0
    }
  })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-600" />
      case 'inactive':
        return <XCircle className="w-4 h-4 text-gray-400" />
      case 'suspended':
        return <Ban className="w-4 h-4 text-red-600" />
      case 'pending':
        return <Clock className="w-4 h-4 text-yellow-600" />
      default:
        return <AlertTriangle className="w-4 h-4 text-gray-400" />
    }
  }

  const getStatusBadge = (status: string) => {
    const baseClasses = "px-2 py-1 text-xs font-medium rounded-full"
    switch (status) {
      case 'active':
        return `${baseClasses} bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200`
      case 'inactive':
        return `${baseClasses} bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200`
      case 'suspended':
        return `${baseClasses} bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200`
      case 'pending':
        return `${baseClasses} bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200`
      default:
        return `${baseClasses} bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200`
    }
  }

  return (
    <div className={`w-full ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <Users className="w-6 h-6 text-blue-600 dark:text-blue-400 mr-3" />
              <div>
                <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                  User Management
                </h1>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Manage users, roles, and permissions
                </p>
              </div>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add User
            </button>
          </div>
        </div>

        {/* Status Messages */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertTriangle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}

        {success && (
          <div className="mx-6 mt-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md">
            <div className="flex items-center">
              <CheckCircle className="w-4 h-4 text-green-500 mr-2" />
              <span className="text-sm text-green-700 dark:text-green-400">{success}</span>
            </div>
          </div>
        )}

        {/* Filters and Search */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search users..."
                value={filter.search}
                onChange={(e) => setFilter(prev => ({ ...prev, search: e.target.value }))}
                className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>

            {/* Status Filter */}
            <select
              value={filter.status}
              onChange={(e) => setFilter(prev => ({ ...prev, status: e.target.value as any }))}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            >
              <option value="all">All Statuses</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="suspended">Suspended</option>
              <option value="pending">Pending</option>
            </select>

            {/* Role Filter */}
            <select
              value={filter.role}
              onChange={(e) => setFilter(prev => ({ ...prev, role: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            >
              <option value="all">All Roles</option>
              {roles.map(role => (
                <option key={role.id} value={role.id}>{role.name}</option>
              ))}
            </select>

            {/* Sort */}
            <div className="flex space-x-2">
              <select
                value={filter.sortBy}
                onChange={(e) => setFilter(prev => ({ ...prev, sortBy: e.target.value as any }))}
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              >
                <option value="name">Name</option>
                <option value="email">Email</option>
                <option value="created">Created</option>
                <option value="lastLogin">Last Login</option>
              </select>
              <button
                onClick={() => setFilter(prev => ({ 
                  ...prev, 
                  sortOrder: prev.sortOrder === 'asc' ? 'desc' : 'asc' 
                }))}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                {filter.sortOrder === 'asc' ? '↑' : '↓'}
              </button>
            </div>
          </div>
        </div>

        {/* Bulk Actions */}
        {selectedUsers.size > 0 && (
          <div className="px-6 py-3 bg-blue-50 dark:bg-blue-900/20 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <span className="text-sm text-blue-700 dark:text-blue-300">
                {selectedUsers.size} user(s) selected
              </span>
              <div className="flex space-x-2">
                <button
                  onClick={() => handleBulkAction('activate')}
                  className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                >
                  Activate
                </button>
                <button
                  onClick={() => handleBulkAction('deactivate')}
                  className="px-3 py-1 text-sm bg-yellow-600 text-white rounded hover:bg-yellow-700"
                >
                  Deactivate
                </button>
                <button
                  onClick={() => handleBulkAction('delete')}
                  className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Users Table */}
        <div className="overflow-x-auto">
          {isLoading ? (
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="w-12 px-6 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={selectedUsers.size === filteredUsers.length && filteredUsers.length > 0}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedUsers(new Set(filteredUsers.map(u => u.id)))
                        } else {
                          setSelectedUsers(new Set())
                        }
                      }}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Last Login
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {filteredUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4">
                      <input
                        type="checkbox"
                        checked={selectedUsers.has(user.id)}
                        onChange={(e) => {
                          const newSelected = new Set(selectedUsers)
                          if (e.target.checked) {
                            newSelected.add(user.id)
                          } else {
                            newSelected.delete(user.id)
                          }
                          setSelectedUsers(newSelected)
                        }}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center mr-3">
                          {user.avatar ? (
                            <img src={user.avatar} alt={user.firstName} className="w-10 h-10 rounded-full" />
                          ) : (
                            <span className="text-sm font-medium text-blue-600 dark:text-blue-400">
                              {user.firstName.charAt(0)}{user.lastName.charAt(0)}
                            </span>
                          )}
                        </div>
                        <div>
                          <div className="text-sm font-medium text-gray-900 dark:text-white">
                            {user.firstName} {user.lastName}
                          </div>
                          <div className="text-sm text-gray-500 dark:text-gray-400">
                            {user.email}
                          </div>
                          <div className="text-xs text-gray-400 dark:text-gray-500">
                            @{user.username}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        <Shield className="w-4 h-4 text-gray-400 mr-2" />
                        <span className="text-sm text-gray-900 dark:text-white">
                          {user.role.name}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        {getStatusIcon(user.status)}
                        <span className={`ml-2 ${getStatusBadge(user.status)}`}>
                          {user.status}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {user.lastLogin ? new Date(user.lastLogin).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {new Date(user.createdAt).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => {
                            setEditingUser(user)
                            setShowEditModal(true)
                          }}
                          className="p-1 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400"
                          title="Edit user"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteUser(user.id)}
                          className="p-1 text-gray-400 hover:text-red-600 dark:hover:text-red-400"
                          title="Delete user"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        <div className="px-6 py-3 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-700 dark:text-gray-300">
              Showing {filteredUsers.length} of {users.length} users
            </div>
            <div className="flex space-x-2">
              <button className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                Previous
              </button>
              <button className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                Next
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Create User Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg w-full max-w-md mx-4">
            <div className="p-6 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                Create New User
              </h2>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    First Name
                  </label>
                  <input
                    type="text"
                    value={createUserData.firstName}
                    onChange={(e) => setCreateUserData(prev => ({ ...prev, firstName: e.target.value }))}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Last Name
                  </label>
                  <input
                    type="text"
                    value={createUserData.lastName}
                    onChange={(e) => setCreateUserData(prev => ({ ...prev, lastName: e.target.value }))}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Email
                </label>
                <input
                  type="email"
                  value={createUserData.email}
                  onChange={(e) => setCreateUserData(prev => ({ ...prev, email: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Username
                </label>
                <input
                  type="text"
                  value={createUserData.username}
                  onChange={(e) => setCreateUserData(prev => ({ ...prev, username: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Role
                </label>
                <select
                  value={createUserData.roleId}
                  onChange={(e) => setCreateUserData(prev => ({ ...prev, roleId: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  <option value="">Select a role</option>
                  {roles.map(role => (
                    <option key={role.id} value={role.id}>{role.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="sendInvite"
                  checked={createUserData.sendInvite}
                  onChange={(e) => setCreateUserData(prev => ({ ...prev, sendInvite: e.target.checked }))}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label htmlFor="sendInvite" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                  Send invitation email
                </label>
              </div>
            </div>
            <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateUser}
                className="px-4 py-2 text-sm text-white bg-blue-600 rounded-md hover:bg-blue-700"
              >
                Create User
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default UserManagement