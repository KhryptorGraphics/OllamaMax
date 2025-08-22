import type { Meta, StoryObj } from '@storybook/react'
import { Layout } from './Layout'
import { Button } from '../Button/Button'
import { Card } from '../Card/Card'
import { Badge } from '../Badge/Badge'
import { Menu, Home, Settings, User, Bell, Search } from 'lucide-react'

const meta: Meta<typeof Layout> = {
  title: 'Design System/Layout',
  component: Layout,
  parameters: {
    docs: {
      description: {
        component: 'Layout component provides the foundational structure for pages and applications. Includes header, sidebar, main content area, and footer sections with responsive behavior.'
      }
    },
    layout: 'fullscreen'
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'sidebar', 'centered', 'fullwidth'],
      description: 'Layout variant for different page structures'
    },
    sidebarCollapsible: {
      control: 'boolean',
      description: 'Whether the sidebar can be collapsed'
    },
    sidebarOpen: {
      control: 'boolean',
      description: 'Initial sidebar open state'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Layout>

// Default layout
export const Default: Story = {
  args: {
    children: (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Welcome to OllamaMax</h1>
        <p className="text-muted-foreground mb-6">
          This is the default layout with header and main content area.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <Card>
            <Card.Header>
              <Card.Title>Feature 1</Card.Title>
              <Card.Description>Description of feature 1</Card.Description>
            </Card.Header>
          </Card>
          <Card>
            <Card.Header>
              <Card.Title>Feature 2</Card.Title>
              <Card.Description>Description of feature 2</Card.Description>
            </Card.Header>
          </Card>
          <Card>
            <Card.Header>
              <Card.Title>Feature 3</Card.Title>
              <Card.Description>Description of feature 3</Card.Description>
            </Card.Header>
          </Card>
        </div>
      </div>
    )
  }
}

// Sidebar layout
export const WithSidebar: Story = {
  args: {
    variant: 'sidebar',
    sidebar: (
      <div className="p-4 space-y-2">
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-4">Navigation</h2>
        </div>
        
        <nav className="space-y-1">
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md bg-primary text-primary-foreground">
            <Home className="w-4 h-4" />
            <span>Dashboard</span>
          </a>
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-muted">
            <User className="w-4 h-4" />
            <span>Profile</span>
          </a>
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-muted">
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </a>
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-muted">
            <Bell className="w-4 h-4" />
            <span>Notifications</span>
            <Badge variant="destructive" size="sm">3</Badge>
          </a>
        </nav>
      </div>
    ),
    children: (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
        <p className="text-muted-foreground mb-6">
          This layout includes a sidebar for navigation.
        </p>
        
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <Card.Header>
              <Card.Title>Quick Stats</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold">1,234</div>
                  <div className="text-sm text-muted-foreground">Users</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold">5,678</div>
                  <div className="text-sm text-muted-foreground">Sessions</div>
                </div>
              </div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Recent Activity</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="space-y-3">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm">User logged in</span>
                  <span className="text-xs text-muted-foreground ml-auto">2m ago</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                  <span className="text-sm">New deployment</span>
                  <span className="text-xs text-muted-foreground ml-auto">5m ago</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
                  <span className="text-sm">Alert resolved</span>
                  <span className="text-xs text-muted-foreground ml-auto">10m ago</span>
                </div>
              </div>
            </Card.Content>
          </Card>
        </div>
      </div>
    )
  }
}

// Collapsible sidebar
export const CollapsibleSidebar: Story = {
  args: {
    variant: 'sidebar',
    sidebarCollapsible: true,
    sidebarOpen: false,
    sidebar: (
      <div className="p-4 space-y-2">
        <nav className="space-y-1">
          <a href="#" className="flex items-center justify-center w-full px-3 py-2 rounded-md bg-primary text-primary-foreground" title="Dashboard">
            <Home className="w-4 h-4" />
          </a>
          <a href="#" className="flex items-center justify-center w-full px-3 py-2 rounded-md hover:bg-muted" title="Profile">
            <User className="w-4 h-4" />
          </a>
          <a href="#" className="flex items-center justify-center w-full px-3 py-2 rounded-md hover:bg-muted" title="Settings">
            <Settings className="w-4 h-4" />
          </a>
          <a href="#" className="flex items-center justify-center w-full px-3 py-2 rounded-md hover:bg-muted relative" title="Notifications">
            <Bell className="w-4 h-4" />
            <div className="absolute -top-1 -right-1 w-2 h-2 bg-destructive rounded-full"></div>
          </a>
        </nav>
      </div>
    ),
    children: (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Collapsed Sidebar Layout</h1>
        <p className="text-muted-foreground mb-6">
          The sidebar is collapsed by default. Click the menu button to expand it.
        </p>
        <Card>
          <Card.Header>
            <Card.Title>Responsive Sidebar</Card.Title>
            <Card.Description>
              The sidebar automatically collapses on smaller screens and can be toggled manually.
            </Card.Description>
          </Card.Header>
        </Card>
      </div>
    )
  }
}

// Centered layout
export const Centered: Story = {
  args: {
    variant: 'centered',
    children: (
      <div className="max-w-2xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-4">Centered Layout</h1>
          <p className="text-lg text-muted-foreground">
            Perfect for forms, documentation, or content-focused pages.
          </p>
        </div>
        
        <Card>
          <Card.Header>
            <Card.Title>Sign Up</Card.Title>
            <Card.Description>
              Create your account to get started with OllamaMax.
            </Card.Description>
          </Card.Header>
          <Card.Content className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium">First Name</label>
                <input className="w-full mt-1 px-3 py-2 border rounded-md" placeholder="John" />
              </div>
              <div>
                <label className="text-sm font-medium">Last Name</label>
                <input className="w-full mt-1 px-3 py-2 border rounded-md" placeholder="Doe" />
              </div>
            </div>
            <div>
              <label className="text-sm font-medium">Email</label>
              <input className="w-full mt-1 px-3 py-2 border rounded-md" placeholder="john@example.com" />
            </div>
            <div>
              <label className="text-sm font-medium">Password</label>
              <input type="password" className="w-full mt-1 px-3 py-2 border rounded-md" placeholder="••••••••" />
            </div>
          </Card.Content>
          <Card.Footer>
            <Button className="w-full">Create Account</Button>
          </Card.Footer>
        </Card>
      </div>
    )
  }
}

// Full width layout
export const FullWidth: Story = {
  args: {
    variant: 'fullwidth',
    children: (
      <div className="p-6">
        <div className="mb-8">
          <h1 className="text-2xl font-bold mb-4">Full Width Layout</h1>
          <p className="text-muted-foreground">
            Utilizes the full viewport width for data-heavy interfaces.
          </p>
        </div>
        
        <div className="overflow-x-auto">
          <table className="w-full border-collapse border border-border">
            <thead>
              <tr className="bg-muted">
                <th className="border border-border px-4 py-2 text-left">Name</th>
                <th className="border border-border px-4 py-2 text-left">Email</th>
                <th className="border border-border px-4 py-2 text-left">Role</th>
                <th className="border border-border px-4 py-2 text-left">Status</th>
                <th className="border border-border px-4 py-2 text-left">Last Login</th>
                <th className="border border-border px-4 py-2 text-left">Actions</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: 10 }, (_, i) => (
                <tr key={i}>
                  <td className="border border-border px-4 py-2">User {i + 1}</td>
                  <td className="border border-border px-4 py-2">user{i + 1}@example.com</td>
                  <td className="border border-border px-4 py-2">
                    <Badge variant={i % 3 === 0 ? 'default' : i % 3 === 1 ? 'secondary' : 'success'}>
                      {i % 3 === 0 ? 'Admin' : i % 3 === 1 ? 'User' : 'Manager'}
                    </Badge>
                  </td>
                  <td className="border border-border px-4 py-2">
                    <Badge variant={i % 2 === 0 ? 'success' : 'secondary'} dot>
                      {i % 2 === 0 ? 'Active' : 'Inactive'}
                    </Badge>
                  </td>
                  <td className="border border-border px-4 py-2">2024-12-{String(i + 1).padStart(2, '0')}</td>
                  <td className="border border-border px-4 py-2">
                    <div className="flex space-x-1">
                      <Button size="sm" variant="ghost">Edit</Button>
                      <Button size="sm" variant="ghost">Delete</Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    )
  }
}

// With header
export const WithHeader: Story = {
  args: {
    header: (
      <div className="flex items-center justify-between px-6 py-4 border-b">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm">
            <Menu className="w-4 h-4" />
          </Button>
          <h1 className="text-xl font-semibold">OllamaMax</h1>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
            <input 
              className="pl-9 pr-4 py-2 w-64 border rounded-md bg-background" 
              placeholder="Search..." 
            />
          </div>
          
          <Button variant="ghost" size="sm">
            <Bell className="w-4 h-4" />
          </Button>
          
          <Button variant="ghost" size="sm">
            <User className="w-4 h-4" />
          </Button>
        </div>
      </div>
    ),
    children: (
      <div className="p-6">
        <h2 className="text-xl font-bold mb-4">Page Content</h2>
        <p className="text-muted-foreground mb-6">
          This layout includes a header with navigation and search functionality.
        </p>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card>
            <Card.Header>
              <Card.Title>Analytics</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">12,345</div>
              <div className="text-sm text-muted-foreground">Total Views</div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Users</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">1,234</div>
              <div className="text-sm text-muted-foreground">Active Users</div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Revenue</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">$45,678</div>
              <div className="text-sm text-muted-foreground">This Month</div>
            </Card.Content>
          </Card>
        </div>
      </div>
    )
  }
}

// With footer
export const WithFooter: Story = {
  args: {
    footer: (
      <div className="px-6 py-4 border-t bg-muted/50">
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            © 2024 OllamaMax. All rights reserved.
          </div>
          <div className="flex items-center space-x-4 text-sm">
            <a href="#" className="text-muted-foreground hover:text-foreground">Privacy</a>
            <a href="#" className="text-muted-foreground hover:text-foreground">Terms</a>
            <a href="#" className="text-muted-foreground hover:text-foreground">Contact</a>
          </div>
        </div>
      </div>
    ),
    children: (
      <div className="p-6 min-h-[60vh]">
        <h2 className="text-xl font-bold mb-4">Content with Footer</h2>
        <p className="text-muted-foreground mb-6">
          This layout includes a footer that stays at the bottom of the page.
        </p>
        
        <Card>
          <Card.Header>
            <Card.Title>Main Content</Card.Title>
            <Card.Description>
              The footer will appear at the bottom regardless of content height.
            </Card.Description>
          </Card.Header>
          <Card.Content>
            <p>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod 
              tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim 
              veniam, quis nostrud exercitation ullamco laboris.
            </p>
          </Card.Content>
        </Card>
      </div>
    )
  }
}

// Complete layout
export const CompleteLayout: Story = {
  args: {
    variant: 'sidebar',
    sidebarCollapsible: true,
    header: (
      <div className="flex items-center justify-between px-6 py-4 border-b">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm">
            <Menu className="w-4 h-4" />
          </Button>
          <h1 className="text-xl font-semibold">OllamaMax Dashboard</h1>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
            <input 
              className="pl-9 pr-4 py-2 w-64 border rounded-md bg-background" 
              placeholder="Search..." 
            />
          </div>
          
          <Button variant="ghost" size="sm">
            <Bell className="w-4 h-4" />
          </Button>
          
          <Button variant="ghost" size="sm">
            <User className="w-4 h-4" />
          </Button>
        </div>
      </div>
    ),
    sidebar: (
      <div className="p-4 space-y-2">
        <nav className="space-y-1">
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md bg-primary text-primary-foreground">
            <Home className="w-4 h-4" />
            <span>Dashboard</span>
          </a>
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-muted">
            <User className="w-4 h-4" />
            <span>Users</span>
          </a>
          <a href="#" className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-muted">
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </a>
        </nav>
      </div>
    ),
    footer: (
      <div className="px-6 py-4 border-t bg-muted/50">
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            © 2024 OllamaMax. All rights reserved.
          </div>
          <div className="flex items-center space-x-4 text-sm">
            <a href="#" className="text-muted-foreground hover:text-foreground">Help</a>
            <a href="#" className="text-muted-foreground hover:text-foreground">Support</a>
          </div>
        </div>
      </div>
    ),
    children: (
      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">Complete Application Layout</h2>
        
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 mb-8">
          <Card>
            <Card.Header>
              <Card.Title>Total Users</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">1,234</div>
              <div className="text-sm text-muted-foreground">+12% from last month</div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Active Sessions</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">567</div>
              <div className="text-sm text-muted-foreground">Currently online</div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Revenue</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">$12,345</div>
              <div className="text-sm text-muted-foreground">This month</div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Conversion</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="text-2xl font-bold">3.2%</div>
              <div className="text-sm text-muted-foreground">+0.5% increase</div>
            </Card.Content>
          </Card>
        </div>
        
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <Card.Header>
              <Card.Title>Recent Activity</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="space-y-4">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                    <User className="w-4 h-4" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">New user registered</p>
                    <p className="text-xs text-muted-foreground">2 minutes ago</p>
                  </div>
                </div>
                
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-success/10 rounded-full flex items-center justify-center">
                    <Settings className="w-4 h-4" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">System updated</p>
                    <p className="text-xs text-muted-foreground">5 minutes ago</p>
                  </div>
                </div>
                
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-warning/10 rounded-full flex items-center justify-center">
                    <Bell className="w-4 h-4" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">New notification</p>
                    <p className="text-xs text-muted-foreground">10 minutes ago</p>
                  </div>
                </div>
              </div>
            </Card.Content>
          </Card>
          
          <Card>
            <Card.Header>
              <Card.Title>Quick Actions</Card.Title>
            </Card.Header>
            <Card.Content>
              <div className="grid grid-cols-2 gap-3">
                <Button variant="outline" className="justify-start">
                  <User className="w-4 h-4 mr-2" />
                  Add User
                </Button>
                <Button variant="outline" className="justify-start">
                  <Settings className="w-4 h-4 mr-2" />
                  Settings
                </Button>
                <Button variant="outline" className="justify-start">
                  <Bell className="w-4 h-4 mr-2" />
                  Notifications
                </Button>
                <Button variant="outline" className="justify-start">
                  <Search className="w-4 h-4 mr-2" />
                  Search
                </Button>
              </div>
            </Card.Content>
          </Card>
        </div>
      </div>
    )
  }
}

// Responsive behavior demo
export const ResponsiveBehavior: Story = {
  render: () => (
    <div className="space-y-6">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-4">Responsive Layout Behavior</h2>
        <p className="text-muted-foreground">
          Resize your browser window to see how the layout adapts to different screen sizes.
        </p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card>
          <Card.Header>
            <Card.Title>Mobile First</Card.Title>
            <Card.Description>
              Layout starts with mobile design and scales up
            </Card.Description>
          </Card.Header>
        </Card>
        
        <Card className="md:col-span-2 lg:col-span-1">
          <Card.Header>
            <Card.Title>Flexible Grid</Card.Title>
            <Card.Description>
              Grid adapts to available space
            </Card.Description>
          </Card.Header>
        </Card>
        
        <Card className="lg:col-span-2">
          <Card.Header>
            <Card.Title>Responsive Content</Card.Title>
            <Card.Description>
              Content reflows based on screen size
            </Card.Description>
          </Card.Header>
        </Card>
      </div>
      
      <div className="text-center text-sm text-muted-foreground">
        <p>Breakpoints: sm (640px), md (768px), lg (1024px), xl (1280px), 2xl (1536px)</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Demonstrates responsive behavior across different screen sizes and breakpoints.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    variant: 'default',
    sidebarCollapsible: false,
    sidebarOpen: true,
    children: (
      <div className="p-6">
        <h2 className="text-xl font-bold mb-4">Layout Playground</h2>
        <p className="text-muted-foreground">
          Use the controls to experiment with different layout configurations.
        </p>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different layout configurations.'
      }
    }
  }
}