/**
 * Navigation - Navigation menu configuration and utilities
 * Features: Permission-based menu items, notification badges, user menu
 */

import React from 'react'
import styled from 'styled-components'
import { 
  Home,
  Cpu,
  Network,
  Activity,
  Shield,
  Settings,
  Users,
  FileText,
  BarChart3,
  Database,
  Lock,
  Globe,
  Zap,
  Calendar,
  Download,
  Upload,
  HardDrive,
  Monitor,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  Server,
  Wifi,
  Cloud
} from 'lucide-react'

import { Badge } from '../../design-system/components/Badge/Badge'

// Types
export interface MenuItem {
  id: string
  label: string
  icon: React.ComponentType<{ size?: number; className?: string }>
  path: string
  badge?: {
    text: string
    variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error'
    pulse?: boolean
  }
  permission?: string
  external?: boolean
  children?: MenuItem[]
}

export interface MenuGroup {
  id: string
  title: string
  permission?: string
  items: MenuItem[]
}

export interface NavigationConfig {
  groups: MenuGroup[]
  userMenu: MenuItem[]
  quickActions: MenuItem[]
}

// Permission types
export type Permission = 
  | 'view_dashboard'
  | 'manage_models'
  | 'manage_nodes'
  | 'view_monitoring'
  | 'manage_security'
  | 'manage_users'
  | 'manage_settings'
  | 'view_analytics'
  | 'manage_storage'
  | 'view_logs'
  | 'manage_api_keys'
  | 'manage_network'
  | 'admin'

// User role type
export interface UserRole {
  id: string
  name: string
  permissions: Permission[]
}

// Common user roles
export const USER_ROLES: Record<string, UserRole> = {
  admin: {
    id: 'admin',
    name: 'Administrator',
    permissions: [
      'view_dashboard',
      'manage_models',
      'manage_nodes',
      'view_monitoring',
      'manage_security',
      'manage_users',
      'manage_settings',
      'view_analytics',
      'manage_storage',
      'view_logs',
      'manage_api_keys',
      'manage_network',
      'admin'
    ]
  },
  operator: {
    id: 'operator',
    name: 'Operator',
    permissions: [
      'view_dashboard',
      'manage_models',
      'manage_nodes',
      'view_monitoring',
      'view_analytics',
      'view_logs'
    ]
  },
  viewer: {
    id: 'viewer',
    name: 'Viewer',
    permissions: [
      'view_dashboard',
      'view_monitoring',
      'view_analytics',
      'view_logs'
    ]
  }
}

// Styled Components
const NavigationContainer = styled.nav`
  display: flex;
  flex-direction: column;
  height: 100%;
`

const NavSection = styled.div`
  margin-bottom: 1.5rem;
`

const NavSectionTitle = styled.h3`
  font-size: 0.75rem;
  font-weight: 600;
  color: ${({ theme }) => theme.colors.text.tertiary};
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 0.5rem 0;
  padding: 0 1rem;
`

const NavList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
`

const NavItemWrapper = styled.li`
  margin-bottom: 0.25rem;
`

// Navigation Configuration
export const NAVIGATION_CONFIG: NavigationConfig = {
  groups: [
    {
      id: 'main',
      title: 'Main',
      items: [
        {
          id: 'dashboard',
          label: 'Dashboard',
          icon: Home,
          path: '/dashboard',
          permission: 'view_dashboard'
        },
        {
          id: 'models',
          label: 'Models',
          icon: Cpu,
          path: '/models',
          permission: 'manage_models',
          badge: { text: '12', variant: 'primary' }
        },
        {
          id: 'nodes',
          label: 'Nodes',
          icon: Network,
          path: '/nodes',
          permission: 'manage_nodes',
          badge: { text: '5', variant: 'success' }
        }
      ]
    },
    {
      id: 'monitoring',
      title: 'Monitoring',
      items: [
        {
          id: 'performance',
          label: 'Performance',
          icon: Activity,
          path: '/performance',
          permission: 'view_monitoring'
        },
        {
          id: 'analytics',
          label: 'Analytics',
          icon: BarChart3,
          path: '/analytics',
          permission: 'view_analytics'
        },
        {
          id: 'logs',
          label: 'Logs',
          icon: FileText,
          path: '/logs',
          permission: 'view_logs'
        },
        {
          id: 'alerts',
          label: 'Alerts',
          icon: AlertTriangle,
          path: '/alerts',
          permission: 'view_monitoring',
          badge: { text: '3', variant: 'warning', pulse: true }
        }
      ]
    },
    {
      id: 'management',
      title: 'Management',
      permission: 'admin',
      items: [
        {
          id: 'security',
          label: 'Security',
          icon: Shield,
          path: '/security',
          permission: 'manage_security',
          badge: { text: '2', variant: 'error', pulse: true }
        },
        {
          id: 'users',
          label: 'Users',
          icon: Users,
          path: '/users',
          permission: 'manage_users'
        },
        {
          id: 'storage',
          label: 'Storage',
          icon: Database,
          path: '/storage',
          permission: 'manage_storage',
          children: [
            {
              id: 'volumes',
              label: 'Volumes',
              icon: HardDrive,
              path: '/storage/volumes',
              permission: 'manage_storage'
            },
            {
              id: 'backups',
              label: 'Backups',
              icon: Download,
              path: '/storage/backups',
              permission: 'manage_storage'
            }
          ]
        }
      ]
    },
    {
      id: 'system',
      title: 'System',
      items: [
        {
          id: 'settings',
          label: 'Settings',
          icon: Settings,
          path: '/settings',
          permission: 'manage_settings'
        },
        {
          id: 'api-keys',
          label: 'API Keys',
          icon: Lock,
          path: '/api-keys',
          permission: 'manage_api_keys'
        },
        {
          id: 'network',
          label: 'Network',
          icon: Globe,
          path: '/network',
          permission: 'manage_network'
        }
      ]
    }
  ],
  userMenu: [
    {
      id: 'profile',
      label: 'Profile',
      icon: Users,
      path: '/profile'
    },
    {
      id: 'preferences',
      label: 'Preferences',
      icon: Settings,
      path: '/preferences'
    },
    {
      id: 'api-tokens',
      label: 'API Tokens',
      icon: Lock,
      path: '/api-tokens'
    }
  ],
  quickActions: [
    {
      id: 'pull-model',
      label: 'Pull Model',
      icon: Download,
      path: '/models/pull',
      permission: 'manage_models'
    },
    {
      id: 'add-node',
      label: 'Add Node',
      icon: Network,
      path: '/nodes/add',
      permission: 'manage_nodes'
    },
    {
      id: 'create-backup',
      label: 'Create Backup',
      icon: Upload,
      path: '/storage/backup',
      permission: 'manage_storage'
    }
  ]
}

// Permission checking utilities
export const hasPermission = (userPermissions: Permission[], required?: string): boolean => {
  if (!required) return true
  if (userPermissions.includes('admin')) return true
  return userPermissions.includes(required as Permission)
}

export const filterMenuByPermissions = (
  menu: MenuItem[],
  userPermissions: Permission[]
): MenuItem[] => {
  return menu.filter(item => {
    if (!hasPermission(userPermissions, item.permission)) return false
    
    if (item.children) {
      item.children = filterMenuByPermissions(item.children, userPermissions)
    }
    
    return true
  })
}

export const filterGroupsByPermissions = (
  groups: MenuGroup[],
  userPermissions: Permission[]
): MenuGroup[] => {
  return groups
    .filter(group => hasPermission(userPermissions, group.permission))
    .map(group => ({
      ...group,
      items: filterMenuByPermissions(group.items, userPermissions)
    }))
    .filter(group => group.items.length > 0)
}

// Navigation state hook
export const useNavigation = (userRole?: string) => {
  const [activeItem, setActiveItem] = React.useState<string>('')
  const [expandedGroups, setExpandedGroups] = React.useState<Set<string>>(new Set())

  // Get user permissions
  const userPermissions = userRole ? USER_ROLES[userRole]?.permissions || [] : []

  // Filter navigation based on permissions
  const filteredGroups = React.useMemo(() => {
    return filterGroupsByPermissions(NAVIGATION_CONFIG.groups, userPermissions)
  }, [userPermissions])

  const filteredUserMenu = React.useMemo(() => {
    return filterMenuByPermissions(NAVIGATION_CONFIG.userMenu, userPermissions)
  }, [userPermissions])

  const filteredQuickActions = React.useMemo(() => {
    return filterMenuByPermissions(NAVIGATION_CONFIG.quickActions, userPermissions)
  }, [userPermissions])

  // Group expansion handlers
  const toggleGroup = (groupId: string) => {
    const newExpanded = new Set(expandedGroups)
    if (newExpanded.has(groupId)) {
      newExpanded.delete(groupId)
    } else {
      newExpanded.add(groupId)
    }
    setExpandedGroups(newExpanded)
  }

  const isGroupExpanded = (groupId: string) => expandedGroups.has(groupId)

  // Active item management
  const setActive = (itemId: string) => setActiveItem(itemId)
  const isActive = (itemId: string) => activeItem === itemId

  // Find item by path
  const findItemByPath = (path: string): MenuItem | null => {
    for (const group of filteredGroups) {
      for (const item of group.items) {
        if (item.path === path) return item
        if (item.children) {
          const found = item.children.find(child => child.path === path)
          if (found) return found
        }
      }
    }
    return null
  }

  // Get breadcrumbs for current path
  const getBreadcrumbs = (path: string): MenuItem[] => {
    const breadcrumbs: MenuItem[] = []
    const item = findItemByPath(path)
    
    if (item) {
      // Find parent group and item
      for (const group of filteredGroups) {
        for (const groupItem of group.items) {
          if (groupItem.path === path) {
            breadcrumbs.push(groupItem)
            break
          }
          if (groupItem.children) {
            const child = groupItem.children.find(c => c.path === path)
            if (child) {
              breadcrumbs.push(groupItem, child)
              break
            }
          }
        }
      }
    }

    return breadcrumbs
  }

  return {
    groups: filteredGroups,
    userMenu: filteredUserMenu,
    quickActions: filteredQuickActions,
    activeItem,
    setActive,
    isActive,
    expandedGroups,
    toggleGroup,
    isGroupExpanded,
    findItemByPath,
    getBreadcrumbs,
    hasPermission: (permission?: string) => hasPermission(userPermissions, permission)
  }
}

// Notification badge component
interface NotificationBadgeProps {
  count?: number
  variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error'
  pulse?: boolean
  max?: number
}

const BadgeWrapper = styled.div<{ $pulse?: boolean }>`
  position: relative;
  
  ${({ $pulse }) => $pulse && `
    &::after {
      content: '';
      position: absolute;
      top: 0;
      right: 0;
      width: 100%;
      height: 100%;
      border-radius: 50%;
      background-color: inherit;
      animation: pulse 2s infinite;
    }
    
    @keyframes pulse {
      0% {
        opacity: 1;
        transform: scale(1);
      }
      100% {
        opacity: 0;
        transform: scale(1.5);
      }
    }
  `}
`

export const NotificationBadge: React.FC<NotificationBadgeProps> = ({
  count = 0,
  variant = 'primary',
  pulse = false,
  max = 99
}) => {
  if (count === 0) return null

  const displayCount = count > max ? `${max}+` : count.toString()

  return (
    <BadgeWrapper $pulse={pulse}>
      <Badge variant={variant} size="sm">
        {displayCount}
      </Badge>
    </BadgeWrapper>
  )
}

// Navigation item component
interface NavigationItemProps {
  item: MenuItem
  active?: boolean
  collapsed?: boolean
  depth?: number
  onClick?: (item: MenuItem) => void
}

const NavItemContainer = styled.div<{ $depth: number; $collapsed?: boolean }>`
  padding-left: ${({ $depth, $collapsed }) => 
    $collapsed ? '0' : `${$depth * 1}rem`};
`

const NavItemButton = styled.button<{ $active?: boolean; $collapsed?: boolean }>`
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border: none;
  background: ${({ $active, theme }) =>
    $active ? theme.colors.interactive.primary.default : 'transparent'};
  color: ${({ $active, theme }) =>
    $active ? theme.colors.text.inverse : theme.colors.text.secondary};
  text-align: left;
  border-radius: ${({ theme }) => theme.radius.sm};
  transition: ${({ theme }) => theme.transitions.colors};
  margin: 0 0.5rem;

  &:hover {
    background-color: ${({ $active, theme }) =>
      $active
        ? theme.colors.interactive.primary.hover
        : theme.colors.interactive.ghost.hover};
    color: ${({ $active, theme }) =>
      $active ? theme.colors.text.inverse : theme.colors.text.primary};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  .nav-icon {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .nav-text {
    font-size: 0.875rem;
    font-weight: 500;
    flex: 1;
    opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
    transition: opacity 0.2s ease;
  }

  .nav-badge {
    margin-left: auto;
    opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
    transition: opacity 0.2s ease;
  }
`

export const NavigationItem: React.FC<NavigationItemProps> = ({
  item,
  active = false,
  collapsed = false,
  depth = 0,
  onClick
}) => {
  const handleClick = () => {
    onClick?.(item)
  }

  return (
    <NavItemContainer $depth={depth} $collapsed={collapsed}>
      <NavItemButton
        $active={active}
        $collapsed={collapsed}
        onClick={handleClick}
        aria-label={item.label}
      >
        <item.icon size={20} className="nav-icon" />
        <span className="nav-text">{item.label}</span>
        {item.badge && (
          <div className="nav-badge">
            <NotificationBadge
              count={parseInt(item.badge.text)}
              variant={item.badge.variant}
              pulse={item.badge.pulse}
            />
          </div>
        )}
      </NavItemButton>
    </NavItemContainer>
  )
}

export default {
  NAVIGATION_CONFIG,
  USER_ROLES,
  hasPermission,
  filterMenuByPermissions,
  filterGroupsByPermissions,
  useNavigation,
  NotificationBadge,
  NavigationItem
}