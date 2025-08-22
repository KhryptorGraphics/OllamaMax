/**
 * Sidebar - Collapsible sidebar navigation
 * Features: Collapsible sidebar, menu items with icons and badges, active state management
 */

import React from 'react'
import styled from 'styled-components'
import { NavLink, useLocation } from 'react-router-dom'
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
  Zap,
  Lock,
  Globe,
  ChevronLeft,
  ChevronRight
} from 'lucide-react'

import { Badge } from '../../design-system/components/Badge/Badge'

// Styled Components
const SidebarContainer = styled.aside<{ 
  $collapsed: boolean 
  $mobileOpen: boolean 
}>`
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  width: ${({ $collapsed }) => ($collapsed ? '64px' : '240px')};
  background-color: ${({ theme }) => theme.colors.background.secondary};
  border-right: 1px solid ${({ theme }) => theme.colors.border.primary};
  transition: width 0.3s ease;
  z-index: 50;
  display: flex;
  flex-direction: column;

  @media (max-width: 768px) {
    transform: ${({ $mobileOpen }) =>
      $mobileOpen ? 'translateX(0)' : 'translateX(-100%)'};
    width: 240px;
    transition: transform 0.3s ease;
  }
`

const SidebarHeader = styled.div<{ $collapsed: boolean }>`
  display: flex;
  align-items: center;
  justify-content: ${({ $collapsed }) => ($collapsed ? 'center' : 'space-between')};
  padding: 1rem;
  border-bottom: 1px solid ${({ theme }) => theme.colors.border.primary};
  min-height: 64px;
`

const Logo = styled.div<{ $collapsed: boolean }>`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: ${({ theme }) => theme.colors.text.primary};

  .logo-text {
    opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
    transition: opacity 0.2s ease;
  }
`

const LogoIcon = styled.div`
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border-radius: ${({ theme }) => theme.radius.md};
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  flex-shrink: 0;
`

const ToggleButton = styled.button<{ $collapsed: boolean }>`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: ${({ theme }) => theme.colors.text.secondary};
  border-radius: ${({ theme }) => theme.radius.sm};
  transition: ${({ theme }) => theme.transitions.colors};
  opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};

  &:hover {
    background-color: ${({ theme }) => theme.colors.interactive.ghost.hover};
    color: ${({ theme }) => theme.colors.text.primary};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  @media (max-width: 768px) {
    display: none;
  }
`

const Navigation = styled.nav`
  flex: 1;
  padding: 1rem 0;
  overflow-y: auto;
`

const NavGroup = styled.div`
  margin-bottom: 1.5rem;
`

const NavGroupTitle = styled.h3<{ $collapsed: boolean }>`
  font-size: 0.75rem;
  font-weight: 600;
  color: ${({ theme }) => theme.colors.text.tertiary};
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 0.5rem 0;
  padding: 0 1rem;
  opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
  transition: opacity 0.2s ease;
`

const NavList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
`

const NavItem = styled.li``

const NavItemLink = styled(NavLink)<{ $collapsed: boolean }>`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: ${({ theme }) => theme.colors.text.secondary};
  text-decoration: none;
  transition: ${({ theme }) => theme.transitions.colors};
  position: relative;
  border-radius: 0;
  margin: 0 0.5rem;
  border-radius: ${({ theme }) => theme.radius.sm};

  &:hover {
    background-color: ${({ theme }) => theme.colors.interactive.ghost.hover};
    color: ${({ theme }) => theme.colors.text.primary};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  &.active {
    background-color: ${({ theme }) => theme.colors.interactive.primary.default};
    color: ${({ theme }) => theme.colors.text.inverse};

    &:hover {
      background-color: ${({ theme }) => theme.colors.interactive.primary.hover};
    }

    .nav-badge {
      background-color: ${({ theme }) => theme.colors.text.inverse};
      color: ${({ theme }) => theme.colors.interactive.primary.default};
    }
  }

  .nav-icon {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .nav-text {
    font-size: 0.875rem;
    font-weight: 500;
    opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
    transition: opacity 0.2s ease;
    flex: 1;
  }

  .nav-badge {
    margin-left: auto;
    opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
    transition: opacity 0.2s ease;
  }
`

const TooltipWrapper = styled.div<{ $collapsed: boolean }>`
  position: relative;
  
  &:hover .tooltip {
    opacity: ${({ $collapsed }) => ($collapsed ? 1 : 0)};
    visibility: ${({ $collapsed }) => ($collapsed ? 'visible' : 'hidden')};
  }
`

const Tooltip = styled.div`
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: ${({ theme }) => theme.colors.background.inverse};
  color: ${({ theme }) => theme.colors.text.inverse};
  font-size: 0.75rem;
  border-radius: ${({ theme }) => theme.radius.sm};
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all 0.2s ease;
  z-index: 60;
  pointer-events: none;

  &::before {
    content: '';
    position: absolute;
    right: 100%;
    top: 50%;
    transform: translateY(-50%);
    border: 4px solid transparent;
    border-right-color: ${({ theme }) => theme.colors.background.inverse};
  }
`

const SidebarFooter = styled.div`
  padding: 1rem;
  border-top: 1px solid ${({ theme }) => theme.colors.border.primary};
`

const UserInfo = styled.div<{ $collapsed: boolean }>`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-radius: ${({ theme }) => theme.radius.sm};
  background-color: ${({ theme }) => theme.colors.background.tertiary};
`

const UserAvatar = styled.div`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #22c55e, #16a34a);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 0.875rem;
  flex-shrink: 0;
`

const UserDetails = styled.div<{ $collapsed: boolean }>`
  flex: 1;
  opacity: ${({ $collapsed }) => ($collapsed ? 0 : 1)};
  transition: opacity 0.2s ease;
`

const UserName = styled.div`
  font-size: 0.875rem;
  font-weight: 500;
  color: ${({ theme }) => theme.colors.text.primary};
`

const UserRole = styled.div`
  font-size: 0.75rem;
  color: ${({ theme }) => theme.colors.text.tertiary};
`

// Types
interface MenuItem {
  id: string
  label: string
  icon: React.ComponentType<{ size?: number; className?: string }>
  path: string
  badge?: {
    text: string
    variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error'
  }
  external?: boolean
}

interface MenuGroup {
  id: string
  title: string
  items: MenuItem[]
}

interface SidebarProps {
  collapsed: boolean
  mobileOpen: boolean
  onToggle: () => void
  onMobileClose: () => void
}

// Menu Configuration
const MENU_GROUPS: MenuGroup[] = [
  {
    id: 'main',
    title: 'Main',
    items: [
      {
        id: 'dashboard',
        label: 'Dashboard',
        icon: Home,
        path: '/dashboard'
      },
      {
        id: 'models',
        label: 'Models',
        icon: Cpu,
        path: '/models',
        badge: { text: '12', variant: 'primary' }
      },
      {
        id: 'nodes',
        label: 'Nodes',
        icon: Network,
        path: '/nodes',
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
        path: '/performance'
      },
      {
        id: 'analytics',
        label: 'Analytics',
        icon: BarChart3,
        path: '/analytics'
      },
      {
        id: 'logs',
        label: 'Logs',
        icon: FileText,
        path: '/logs'
      }
    ]
  },
  {
    id: 'management',
    title: 'Management',
    items: [
      {
        id: 'security',
        label: 'Security',
        icon: Shield,
        path: '/security',
        badge: { text: '3', variant: 'warning' }
      },
      {
        id: 'users',
        label: 'Users',
        icon: Users,
        path: '/users'
      },
      {
        id: 'storage',
        label: 'Storage',
        icon: Database,
        path: '/storage'
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
        path: '/settings'
      },
      {
        id: 'api',
        label: 'API Keys',
        icon: Lock,
        path: '/api-keys'
      },
      {
        id: 'network',
        label: 'Network',
        icon: Globe,
        path: '/network'
      }
    ]
  }
]

// Component
export const Sidebar: React.FC<SidebarProps> = ({
  collapsed,
  mobileOpen,
  onToggle,
  onMobileClose
}) => {
  const location = useLocation()

  // Mock user data - in real app, this would come from auth store
  const user = {
    name: 'John Doe',
    role: 'Administrator',
    avatar: 'JD'
  }

  return (
    <SidebarContainer $collapsed={collapsed} $mobileOpen={mobileOpen}>
      {/* Header */}
      <SidebarHeader $collapsed={collapsed}>
        <Logo $collapsed={collapsed}>
          <LogoIcon>O</LogoIcon>
          <span className="logo-text">OllamaMax</span>
        </Logo>
        
        <ToggleButton
          $collapsed={collapsed}
          onClick={onToggle}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </ToggleButton>
      </SidebarHeader>

      {/* Navigation */}
      <Navigation>
        {MENU_GROUPS.map((group) => (
          <NavGroup key={group.id}>
            <NavGroupTitle $collapsed={collapsed}>
              {group.title}
            </NavGroupTitle>
            
            <NavList>
              {group.items.map((item) => (
                <NavItem key={item.id}>
                  <TooltipWrapper $collapsed={collapsed}>
                    <NavItemLink
                      to={item.path}
                      $collapsed={collapsed}
                      onClick={onMobileClose}
                      aria-label={item.label}
                      className={({ isActive }) => (isActive ? 'active' : '')}
                    >
                      <item.icon size={20} className="nav-icon" />
                      <span className="nav-text">{item.label}</span>
                      {item.badge && (
                        <Badge
                          variant={item.badge.variant || 'primary'}
                          size="sm"
                          className="nav-badge"
                        >
                          {item.badge.text}
                        </Badge>
                      )}
                    </NavItemLink>
                    
                    <div className="tooltip">
                      <Tooltip>{item.label}</Tooltip>
                    </div>
                  </TooltipWrapper>
                </NavItem>
              ))}
            </NavList>
          </NavGroup>
        ))}
      </Navigation>

      {/* Footer */}
      <SidebarFooter>
        <UserInfo $collapsed={collapsed}>
          <UserAvatar>{user.avatar}</UserAvatar>
          <UserDetails $collapsed={collapsed}>
            <UserName>{user.name}</UserName>
            <UserRole>{user.role}</UserRole>
          </UserDetails>
        </UserInfo>
      </SidebarFooter>
    </SidebarContainer>
  )
}

export default Sidebar