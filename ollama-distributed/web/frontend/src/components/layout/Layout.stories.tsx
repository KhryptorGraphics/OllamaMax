/**
 * Layout Components Storybook Stories
 * Interactive examples of layout components
 */

import type { Meta, StoryObj } from '@storybook/react'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'
import { AppLayout } from './AppLayout'
import { PageHeader } from './PageHeader'
import { Sidebar } from './Sidebar'
import { theme } from '../../theme/theme'

// Mock theme provider wrapper
const StoryWrapper = ({ children }: { children: React.ReactNode }) => (
  <BrowserRouter>
    <ThemeProvider theme={theme.light}>
      {children}
    </ThemeProvider>
  </BrowserRouter>
)

// AppLayout Stories
const meta: Meta<typeof AppLayout> = {
  title: 'Layout/AppLayout',
  component: AppLayout,
  decorators: [
    (Story) => (
      <StoryWrapper>
        <div style={{ height: '100vh', width: '100vw' }}>
          <Story />
        </div>
      </StoryWrapper>
    )
  ],
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: 'Main application layout with sidebar, header, and responsive navigation.'
      }
    }
  }
}

export default meta
type Story = StoryObj<typeof AppLayout>

export const Default: Story = {
  args: {},
  parameters: {
    docs: {
      description: {
        story: 'Default layout with all features enabled.'
      }
    }
  }
}

export const WithContent: Story = {
  args: {
    children: (
      <div style={{ padding: '2rem' }}>
        <h1>Dashboard</h1>
        <p>Welcome to the OllamaMax dashboard. This is the main content area.</p>
        <div style={{ 
          display: 'grid', 
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', 
          gap: '1rem',
          marginTop: '2rem'
        }}>
          <div style={{ 
            padding: '1rem', 
            backgroundColor: '#f5f5f5', 
            borderRadius: '8px',
            border: '1px solid #e5e5e5'
          }}>
            <h3>System Status</h3>
            <p>All systems operational</p>
          </div>
          <div style={{ 
            padding: '1rem', 
            backgroundColor: '#f5f5f5', 
            borderRadius: '8px',
            border: '1px solid #e5e5e5'
          }}>
            <h3>Active Models</h3>
            <p>12 models running</p>
          </div>
          <div style={{ 
            padding: '1rem', 
            backgroundColor: '#f5f5f5', 
            borderRadius: '8px',
            border: '1px solid #e5e5e5'
          }}>
            <h3>Network Nodes</h3>
            <p>5 nodes connected</p>
          </div>
        </div>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Layout with sample dashboard content.'
      }
    }
  }
}

// PageHeader Stories
export const PageHeaderDefault: Meta<typeof PageHeader> = {
  title: 'Layout/PageHeader',
  component: PageHeader,
  decorators: [
    (Story) => (
      <StoryWrapper>
        <Story />
      </StoryWrapper>
    )
  ],
  parameters: {
    docs: {
      description: {
        component: 'Dynamic page header with breadcrumbs, search, and action buttons.'
      }
    }
  }
}

export const PageHeaderWithFilters: StoryObj<typeof PageHeader> = {
  args: {
    title: 'Models',
    subtitle: 'Manage and monitor AI models',
    showSearch: true,
    showFilters: true,
    showViewToggle: true,
    showSort: true,
    onSearch: (query: string) => console.log('Search:', query),
    onFilter: () => console.log('Filter clicked'),
    onExport: () => console.log('Export clicked'),
    onRefresh: () => console.log('Refresh clicked'),
    onSort: (field: string, direction: 'asc' | 'desc') => console.log('Sort:', field, direction),
    onViewChange: (view: 'grid' | 'list') => console.log('View changed:', view)
  },
  parameters: {
    docs: {
      description: {
        story: 'Page header with all controls enabled for data pages.'
      }
    }
  }
}

// Sidebar Stories
export const SidebarExpanded: Meta<typeof Sidebar> = {
  title: 'Layout/Sidebar',
  component: Sidebar,
  decorators: [
    (Story) => (
      <StoryWrapper>
        <div style={{ height: '100vh', width: '300px', position: 'relative' }}>
          <Story />
        </div>
      </StoryWrapper>
    )
  ]
}

export const SidebarCollapsed: StoryObj<typeof Sidebar> = {
  args: {
    collapsed: true,
    mobileOpen: false,
    onToggle: () => console.log('Toggle sidebar'),
    onMobileClose: () => console.log('Close mobile menu')
  }
}

export const SidebarMobile: StoryObj<typeof Sidebar> = {
  args: {
    collapsed: false,
    mobileOpen: true,
    onToggle: () => console.log('Toggle sidebar'),
    onMobileClose: () => console.log('Close mobile menu')
  },
  parameters: {
    viewport: {
      defaultViewport: 'mobile1'
    }
  }
}