import React from 'react'
import { Outlet } from 'react-router-dom'
import { useStore } from '@/stores'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { MobileNavigation } from '@/components/mobile/MobileNavigation'
import { useMediaQuery } from '@/theme/hooks/useMediaQuery'

interface MainLayoutProps {
  className?: string
}

const MainLayout: React.FC<MainLayoutProps> = ({ className }) => {
  const { ui } = useStore()
  const isMobile = useMediaQuery('(max-width: 1024px)')

  return (
    <div className={`main-layout ${className || ''}`} data-sidebar-open={ui.sidebarOpen}>
      {/* Desktop Sidebar */}
      {!isMobile && <Sidebar />}
      
      {/* Mobile Navigation */}
      {isMobile && <MobileNavigation />}
      
      <div className="main-content">
        <Header />
        
        <main className="page-content">
          <Outlet />
        </main>
      </div>
      
      {/* Mobile sidebar overlay */}
      {isMobile && ui.sidebarOpen && (
        <div 
          className="sidebar-overlay"
          onClick={() => useStore.getState().setSidebarOpen(false)}
          aria-hidden="true"
        />
      )}
    </div>
  )
}

export { MainLayout }