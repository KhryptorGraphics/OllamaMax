/**
 * AppLayout - Main application layout with sidebar and header
 * Features: Responsive navigation, theme integration, notification system
 */

import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import { Outlet, useLocation } from 'react-router-dom'
import { Menu, X, Bell, Search, Settings, User, LogOut } from 'lucide-react'

import { useThemeState } from '../../theme/hooks/useTheme'
import { useAuthStore } from '../../store/auth'
import { Sidebar } from './Sidebar'
import { PageHeader } from './PageHeader'
import { Navigation } from './Navigation'
import { NotificationCenter } from '../common/NotificationCenter'
import { AccessibilityAnnouncer } from '../accessibility/AccessibilityAnnouncer'
import { SkipLinks } from '../accessibility/SkipLinks'

// Styled Components
const LayoutContainer = styled.div<{ $theme: 'light' | 'dark' }>`
  display: flex;
  min-height: 100vh;
  background-color: ${({ theme }) => theme.colors.background.primary};
  color: ${({ theme }) => theme.colors.text.primary};
  transition: ${({ theme }) => theme.transitions.colors};
`

const MainContent = styled.main<{ $sidebarCollapsed: boolean }>`
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-left: ${({ $sidebarCollapsed }) => ($sidebarCollapsed ? '64px' : '240px')};
  transition: margin-left 0.3s ease;

  @media (max-width: 768px) {
    margin-left: 0;
  }
`

const Header = styled.header`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background-color: ${({ theme }) => theme.colors.background.secondary};
  border-bottom: 1px solid ${({ theme }) => theme.colors.border.primary};
  min-height: 64px;
  position: sticky;
  top: 0;
  z-index: 40;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 0.75rem;
`

const MobileMenuButton = styled.button`
  display: none;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: none;
  background: transparent;
  color: ${({ theme }) => theme.colors.text.primary};
  border-radius: ${({ theme }) => theme.radius.sm};
  transition: ${({ theme }) => theme.transitions.colors};

  &:hover {
    background-color: ${({ theme }) => theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  @media (max-width: 768px) {
    display: flex;
  }
`

const SearchContainer = styled.div`
  position: relative;
  max-width: 400px;
  width: 100%;

  @media (max-width: 768px) {
    display: none;
  }
`

const SearchInput = styled.input`
  width: 100%;
  padding: 0.5rem 2.5rem 0.5rem 1rem;
  border: 1px solid ${({ theme }) => theme.colors.border.primary};
  border-radius: ${({ theme }) => theme.radius.md};
  background-color: ${({ theme }) => theme.colors.background.primary};
  color: ${({ theme }) => theme.colors.text.primary};
  font-size: 0.875rem;
  transition: ${({ theme }) => theme.transitions.colors};

  &:focus {
    outline: none;
    border-color: ${({ theme }) => theme.colors.border.focus};
    box-shadow: 0 0 0 3px ${({ theme }) => theme.colors.border.focus}20;
  }

  &::placeholder {
    color: ${({ theme }) => theme.colors.text.tertiary};
  }
`

const SearchIcon = styled(Search)`
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 1rem;
  height: 1rem;
  color: ${({ theme }) => theme.colors.text.tertiary};
  pointer-events: none;
`

const HeaderButton = styled.button<{ $active?: boolean }>`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: none;
  background: ${({ $active, theme }) =>
    $active ? theme.colors.interactive.primary.default : 'transparent'};
  color: ${({ $active, theme }) =>
    $active ? theme.colors.text.inverse : theme.colors.text.primary};
  border-radius: ${({ theme }) => theme.radius.sm};
  transition: ${({ theme }) => theme.transitions.colors};
  position: relative;

  &:hover {
    background-color: ${({ $active, theme }) =>
      $active
        ? theme.colors.interactive.primary.hover
        : theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }
`

const NotificationBadge = styled.span`
  position: absolute;
  top: 0.125rem;
  right: 0.125rem;
  width: 0.5rem;
  height: 0.5rem;
  background-color: ${({ theme }) => theme.colors.status.error.icon};
  border-radius: 50%;
  border: 2px solid ${({ theme }) => theme.colors.background.secondary};
`

const UserMenu = styled.div<{ $open: boolean }>`
  position: relative;
`

const UserMenuDropdown = styled.div<{ $open: boolean }>`
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.5rem;
  width: 200px;
  background-color: ${({ theme }) => theme.colors.background.primary};
  border: 1px solid ${({ theme }) => theme.colors.border.primary};
  border-radius: ${({ theme }) => theme.radius.md};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  opacity: ${({ $open }) => ($open ? 1 : 0)};
  visibility: ${({ $open }) => ($open ? 'visible' : 'hidden')};
  transform: ${({ $open }) => ($open ? 'translateY(0)' : 'translateY(-10px)')};
  transition: all 0.2s ease;
  z-index: 50;
`

const UserMenuItem = styled.button`
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border: none;
  background: transparent;
  color: ${({ theme }) => theme.colors.text.primary};
  text-align: left;
  font-size: 0.875rem;
  border-radius: ${({ theme }) => theme.radius.sm};
  transition: ${({ theme }) => theme.transitions.colors};

  &:hover {
    background-color: ${({ theme }) => theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: -2px;
  }

  svg {
    width: 1rem;
    height: 1rem;
  }
`

const ContentArea = styled.div`
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
`

const MobileOverlay = styled.div<{ $visible: boolean }>`
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 45;
  opacity: ${({ $visible }) => ($visible ? 1 : 0)};
  visibility: ${({ $visible }) => ($visible ? 'visible' : 'hidden')};
  transition: all 0.3s ease;

  @media (min-width: 769px) {
    display: none;
  }
`

// Types
interface AppLayoutProps {
  children?: React.ReactNode
}

// Component
export const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const { theme, toggleTheme } = useThemeState()
  const { user, clear: logout } = useAuthStore()
  const location = useLocation()

  // State
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [notifications] = useState(3) // Mock notification count

  // Effects
  useEffect(() => {
    // Close mobile menu on route change
    setMobileMenuOpen(false)
  }, [location.pathname])

  useEffect(() => {
    // Close user menu when clicking outside
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element
      if (!target.closest('[data-user-menu]')) {
        setUserMenuOpen(false)
      }
    }

    document.addEventListener('click', handleClickOutside)
    return () => document.removeEventListener('click', handleClickOutside)
  }, [])

  // Handlers
  const handleSidebarToggle = () => {
    setSidebarCollapsed(!sidebarCollapsed)
  }

  const handleMobileMenuToggle = () => {
    setMobileMenuOpen(!mobileMenuOpen)
  }

  const handleUserMenuToggle = () => {
    setUserMenuOpen(!userMenuOpen)
  }

  const handleLogout = () => {
    logout()
    setUserMenuOpen(false)
  }

  const handleSearch = (event: React.FormEvent) => {
    event.preventDefault()
    // Implement search functionality
    console.log('Search:', searchQuery)
  }

  return (
    <LayoutContainer $theme={theme}>
      <SkipLinks />
      <AccessibilityAnnouncer />

      {/* Sidebar */}
      <Sidebar
        collapsed={sidebarCollapsed}
        mobileOpen={mobileMenuOpen}
        onToggle={handleSidebarToggle}
        onMobileClose={() => setMobileMenuOpen(false)}
      />

      {/* Mobile Overlay */}
      <MobileOverlay
        $visible={mobileMenuOpen}
        onClick={() => setMobileMenuOpen(false)}
        aria-hidden="true"
      />

      {/* Main Content */}
      <MainContent $sidebarCollapsed={sidebarCollapsed}>
        {/* Header */}
        <Header>
          <HeaderLeft>
            <MobileMenuButton
              onClick={handleMobileMenuToggle}
              aria-label="Toggle mobile menu"
              aria-expanded={mobileMenuOpen}
            >
              {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
            </MobileMenuButton>

            <SearchContainer>
              <form onSubmit={handleSearch}>
                <SearchInput
                  type="search"
                  placeholder="Search..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  aria-label="Search"
                />
                <SearchIcon />
              </form>
            </SearchContainer>
          </HeaderLeft>

          <HeaderRight>
            {/* Notifications */}
            <HeaderButton
              onClick={() => {}}
              aria-label={`Notifications${notifications > 0 ? ` (${notifications} unread)` : ''}`}
            >
              <Bell size={18} />
              {notifications > 0 && <NotificationBadge />}
            </HeaderButton>

            {/* Settings */}
            <HeaderButton
              onClick={() => {}}
              aria-label="Settings"
            >
              <Settings size={18} />
            </HeaderButton>

            {/* Theme Toggle */}
            <HeaderButton
              onClick={toggleTheme}
              aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
            >
              {theme === 'light' ? '🌙' : '☀️'}
            </HeaderButton>

            {/* User Menu */}
            <UserMenu $open={userMenuOpen} data-user-menu>
              <HeaderButton
                onClick={handleUserMenuToggle}
                aria-label="User menu"
                aria-expanded={userMenuOpen}
                aria-haspopup="true"
              >
                <User size={18} />
              </HeaderButton>

              <UserMenuDropdown $open={userMenuOpen}>
                <UserMenuItem onClick={() => setUserMenuOpen(false)}>
                  <User />
                  Profile
                </UserMenuItem>
                <UserMenuItem onClick={() => setUserMenuOpen(false)}>
                  <Settings />
                  Settings
                </UserMenuItem>
                <UserMenuItem onClick={handleLogout}>
                  <LogOut />
                  Sign Out
                </UserMenuItem>
              </UserMenuDropdown>
            </UserMenu>
          </HeaderRight>
        </Header>

        {/* Page Header */}
        <PageHeader />

        {/* Content Area */}
        <ContentArea>
          {children || <Outlet />}
        </ContentArea>
      </MainContent>

      {/* Notification Center */}
      <NotificationCenter />
    </LayoutContainer>
  )
}

export default AppLayout