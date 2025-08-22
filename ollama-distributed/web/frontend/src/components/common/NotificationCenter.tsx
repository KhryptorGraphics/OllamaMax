/**
 * NotificationCenter - Enhanced notification system
 * Features: Toast notifications, queue management, auto-dismiss
 */

import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import { X, CheckCircle, AlertTriangle, XCircle, Info } from 'lucide-react'
import { Button } from '../../design-system/components/Button/Button'

// Types
export type NotificationType = 'success' | 'warning' | 'error' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  title: string
  message?: string
  duration?: number // in milliseconds, 0 for persistent
  action?: {
    label: string
    onClick: () => void
  }
}

// Styled Components
const NotificationContainer = styled.div`
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 1000;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;

  @media (max-width: 480px) {
    left: 1rem;
    right: 1rem;
    max-width: none;
  }
`

const NotificationCard = styled.div<{ $type: NotificationType }>`
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  background-color: ${({ theme, $type }) => theme.colors.status[$type].background};
  border: 1px solid ${({ theme, $type }) => theme.colors.status[$type].border};
  border-radius: ${({ theme }) => theme.radius.md};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  transform: translateX(100%);
  animation: slideIn 0.3s ease forwards;

  @keyframes slideIn {
    to {
      transform: translateX(0);
    }
  }

  &.exiting {
    animation: slideOut 0.3s ease forwards;
  }

  @keyframes slideOut {
    to {
      transform: translateX(100%);
      opacity: 0;
    }
  }
`

const NotificationIcon = styled.div<{ $type: NotificationType }>`
  color: ${({ theme, $type }) => theme.colors.status[$type].icon};
  flex-shrink: 0;
  margin-top: 0.125rem;
`

const NotificationContent = styled.div`
  flex: 1;
  min-width: 0;
`

const NotificationTitle = styled.h4<{ $type: NotificationType }>`
  font-size: 0.875rem;
  font-weight: 600;
  color: ${({ theme, $type }) => theme.colors.status[$type].text};
  margin: 0 0 0.25rem 0;
  line-height: 1.4;
`

const NotificationMessage = styled.p<{ $type: NotificationType }>`
  font-size: 0.75rem;
  color: ${({ theme, $type }) => theme.colors.status[$type].text};
  margin: 0 0 0.75rem 0;
  line-height: 1.4;
  opacity: 0.8;
`

const NotificationActions = styled.div`
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
`

const CloseButton = styled.button<{ $type: NotificationType }>`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border: none;
  background: transparent;
  color: ${({ theme, $type }) => theme.colors.status[$type].text};
  border-radius: ${({ theme }) => theme.radius.sm};
  flex-shrink: 0;
  margin-top: 0.125rem;
  transition: ${({ theme }) => theme.transitions.colors};
  opacity: 0.6;

  &:hover {
    opacity: 1;
    background-color: rgba(0, 0, 0, 0.1);
  }

  &:focus-visible {
    outline: 2px solid ${({ theme, $type }) => theme.colors.status[$type].icon};
    outline-offset: 2px;
  }
`

// Icon mapping
const NOTIFICATION_ICONS = {
  success: CheckCircle,
  warning: AlertTriangle,
  error: XCircle,
  info: Info
}

// Notification store/context (simplified)
let notificationId = 0
const generateId = () => `notification-${++notificationId}`

// Create a simple notification store
class NotificationStore {
  private listeners: Set<(notifications: Notification[]) => void> = new Set()
  private notifications: Notification[] = []

  subscribe(listener: (notifications: Notification[]) => void) {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  getNotifications() {
    return this.notifications
  }

  add(notification: Omit<Notification, 'id'>) {
    const newNotification: Notification = {
      id: generateId(),
      duration: 5000, // Default 5 seconds
      ...notification
    }

    this.notifications = [...this.notifications, newNotification]
    this.notifyListeners()

    // Auto-dismiss if duration is set
    if (newNotification.duration && newNotification.duration > 0) {
      setTimeout(() => {
        this.remove(newNotification.id)
      }, newNotification.duration)
    }

    return newNotification.id
  }

  remove(id: string) {
    this.notifications = this.notifications.filter(n => n.id !== id)
    this.notifyListeners()
  }

  clear() {
    this.notifications = []
    this.notifyListeners()
  }

  private notifyListeners() {
    this.listeners.forEach(listener => listener(this.notifications))
  }
}

const notificationStore = new NotificationStore()

// Notification API
export const notifications = {
  success: (title: string, message?: string, options?: Partial<Notification>) =>
    notificationStore.add({ type: 'success', title, message, ...options }),
  
  warning: (title: string, message?: string, options?: Partial<Notification>) =>
    notificationStore.add({ type: 'warning', title, message, ...options }),
  
  error: (title: string, message?: string, options?: Partial<Notification>) =>
    notificationStore.add({ type: 'error', title, message, duration: 0, ...options }),
  
  info: (title: string, message?: string, options?: Partial<Notification>) =>
    notificationStore.add({ type: 'info', title, message, ...options }),
  
  dismiss: (id: string) => notificationStore.remove(id),
  
  clear: () => notificationStore.clear()
}

// Hook to use notifications
export const useNotifications = () => {
  const [notificationList, setNotifications] = useState<Notification[]>([])

  useEffect(() => {
    const unsubscribe = notificationStore.subscribe(setNotifications)
    setNotifications(notificationStore.getNotifications())
    return unsubscribe
  }, [])

  return {
    notifications: notificationList,
    add: notifications,
    dismiss: notifications.dismiss,
    clear: notifications.clear
  }
}

// Notification Item Component
interface NotificationItemProps {
  notification: Notification
  onDismiss: (id: string) => void
}

const NotificationItem: React.FC<NotificationItemProps> = ({
  notification,
  onDismiss
}) => {
  const [isExiting, setIsExiting] = useState(false)
  const Icon = NOTIFICATION_ICONS[notification.type]

  const handleDismiss = () => {
    setIsExiting(true)
    // Wait for animation to complete
    setTimeout(() => {
      onDismiss(notification.id)
    }, 300)
  }

  return (
    <NotificationCard
      $type={notification.type}
      className={isExiting ? 'exiting' : ''}
      role="alert"
      aria-live="polite"
    >
      <NotificationIcon $type={notification.type}>
        <Icon size={18} />
      </NotificationIcon>

      <NotificationContent>
        <NotificationTitle $type={notification.type}>
          {notification.title}
        </NotificationTitle>
        
        {notification.message && (
          <NotificationMessage $type={notification.type}>
            {notification.message}
          </NotificationMessage>
        )}

        {notification.action && (
          <NotificationActions>
            <Button
              variant="ghost"
              size="sm"
              onClick={notification.action.onClick}
            >
              {notification.action.label}
            </Button>
          </NotificationActions>
        )}
      </NotificationContent>

      <CloseButton
        $type={notification.type}
        onClick={handleDismiss}
        aria-label="Dismiss notification"
      >
        <X size={14} />
      </CloseButton>
    </NotificationCard>
  )
}

// Main NotificationCenter Component
export const NotificationCenter: React.FC = () => {
  const { notifications: notificationList, dismiss } = useNotifications()

  if (notificationList.length === 0) {
    return null
  }

  return (
    <NotificationContainer id="notification-center">
      {notificationList.map((notification) => (
        <NotificationItem
          key={notification.id}
          notification={notification}
          onDismiss={dismiss}
        />
      ))}
    </NotificationContainer>
  )
}

export default NotificationCenter