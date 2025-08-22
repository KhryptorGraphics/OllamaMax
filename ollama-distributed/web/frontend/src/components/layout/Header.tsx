import React from 'react'
import { useLocation, Link } from 'react-router-dom'
import { useStore } from '@/stores'
import { Button } from '@/design-system/components/Button/Button'
import { ThemeToggle } from '@/theme/components/ThemeToggle'
import { Badge } from '@/design-system/components/Badge/Badge'
import { useMediaQuery } from '@/theme/hooks/useMediaQuery'

// Breadcrumb configuration
const breadcrumbConfig: Record<string, { label: string; parent?: string }> = {
  '/dashboard': { label: 'Dashboard' },
  '/models': { label: 'Models' },
  '/nodes': { label: 'Nodes' },
  '/monitoring': { label: 'Monitoring' },
  '/tasks': { label: 'Tasks' },
  '/transfers': { label: 'Transfers' },
  '/security': { label: 'Security' },
  '/performance': { label: 'Performance' },
  '/settings': { label: 'Settings' },
  '/profile': { label: 'Profile' }
}

interface HeaderProps {
  className?: string
}

const Header: React.FC<HeaderProps> = ({ className }) => {
  const location = useLocation()
  const { ui, auth, notifications } = useStore()
  const isMobile = useMediaQuery('(max-width: 1024px)')
  
  // Generate breadcrumbs
  const generateBreadcrumbs = () => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const breadcrumbs = []
    
    let currentPath = ''
    for (const segment of pathSegments) {
      currentPath += `/${segment}`
      const config = breadcrumbConfig[currentPath]
      
      if (config) {
        breadcrumbs.push({
          path: currentPath,
          label: config.label,
          isLast: currentPath === location.pathname
        })
      }
    }
    
    return breadcrumbs
  }
  
  const breadcrumbs = generateBreadcrumbs()
  const currentPage = breadcrumbConfig[location.pathname]
  const totalNotifications = (notifications.alerts?.length || 0) + 
                             (notifications.pendingTasks || 0) + 
                             (notifications.securityAlerts || 0)

  const toggleSidebar = () => {
    useStore.getState().setSidebarOpen(!ui.sidebarOpen)
  }

  const handleLogout = async () => {
    const { authService } = await import('@/services/auth/authService')
    await authService.logout()
  }

  return (
    <header className={`header ${className || ''}`}>
      <div className="header-main">
        {/* Mobile menu button and breadcrumbs */}
        <div className="header-left">
          {isMobile && (
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleSidebar}
              className="mobile-menu-btn"
              aria-label={ui.sidebarOpen ? 'Close menu' : 'Open menu'}
            >
              <span className="menu-icon">
                {ui.sidebarOpen ? '✕' : '☰'}
              </span>
            </Button>
          )}
          
          {/* Breadcrumbs */}
          {breadcrumbs.length > 0 && (
            <nav className="breadcrumbs" aria-label="Breadcrumb navigation">
              <ol className="breadcrumb-list">
                <li className="breadcrumb-item">
                  <Link to="/dashboard" className="breadcrumb-link">
                    🏠 Home
                  </Link>
                </li>
                {breadcrumbs.map((crumb, index) => (
                  <li key={crumb.path} className="breadcrumb-item">
                    <span className="breadcrumb-separator" aria-hidden="true">&gt;</span>
                    {crumb.isLast ? (
                      <span className="breadcrumb-current" aria-current="page">
                        {crumb.label}
                      </span>
                    ) : (
                      <Link to={crumb.path} className="breadcrumb-link">
                        {crumb.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ol>
            </nav>
          )}
        </div>
        
        {/* Page title for mobile */}
        {isMobile && currentPage && (
          <h1 className="page-title-mobile">{currentPage.label}</h1>
        )}
        
        {/* Header actions */}
        <div className="header-right">
          {/* Notifications */}
          <Button
            variant="ghost"
            size="sm"
            className="notifications-btn"
            aria-label={`${totalNotifications} notifications`}
            title="View notifications"
          >
            <span className="notification-icon">🔔</span>
            {totalNotifications > 0 && (
              <Badge 
                variant="error" 
                size="sm" 
                className="notification-badge"
              >
                {totalNotifications > 99 ? '99+' : totalNotifications}
              </Badge>
            )}
          </Button>
          
          {/* Theme toggle */}
          <ThemeToggle />
          
          {/* User menu */}
          <div className="user-menu">
            <Button
              variant="ghost"
              size="sm"
              className="user-btn"
              title={`Logged in as ${auth.user?.username || 'User'}`}
            >
              <span className="user-avatar">👤</span>
              {!isMobile && (
                <span className="user-name">
                  {auth.user?.username || 'User'}
                </span>
              )}
            </Button>
            
            {/* User dropdown menu would go here */}
            <div className="user-dropdown">
              <Link to="/profile" className="dropdown-item">
                👤 Profile
              </Link>
              <Link to="/settings" className="dropdown-item">
                ⚙️ Settings
              </Link>
              <button onClick={handleLogout} className="dropdown-item logout-btn">
                😪 Logout
              </button>
            </div>
          </div>
        </div>
      </div>
      
      {/* Page title for desktop */}
      {!isMobile && currentPage && (
        <div className="page-header">
          <h1 className="page-title">{currentPage.label}</h1>
        </div>
      )}
    </header>
  )
}

export { Header }