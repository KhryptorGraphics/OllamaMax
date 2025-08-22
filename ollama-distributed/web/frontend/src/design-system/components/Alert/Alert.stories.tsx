import type { Meta, StoryObj } from '@storybook/react'
import { Alert, AlertTitle, AlertDescription, AlertActions, ToastAlert, BannerAlert } from './Alert'
import { Button } from '../Button/Button'
import { Info, AlertTriangle, CheckCircle, AlertCircle, RefreshCw } from 'lucide-react'
import { useState } from 'react'

const meta: Meta<typeof Alert> = {
  title: 'Design System/Alert',
  component: Alert,
  parameters: {
    docs: {
      description: {
        component: 'Alert component for displaying important messages, notifications, and system status. Supports multiple variants, dismissible alerts, and action buttons.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'info', 'success', 'warning', 'destructive'],
      description: 'Visual style variant of the alert'
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size of the alert'
    },
    dismissible: {
      control: 'boolean',
      description: 'Whether the alert can be dismissed'
    },
    showIcon: {
      control: 'boolean',
      description: 'Whether to show the default variant icon'
    },
    title: {
      control: 'text',
      description: 'Alert title'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Alert>

// Default story
export const Default: Story = {
  args: {
    children: 'This is a default alert message.'
  }
}

// Variant stories
export const Info: Story = {
  args: {
    variant: 'info',
    title: 'Information',
    children: 'This is an informational alert that provides helpful context.'
  }
}

export const Success: Story = {
  args: {
    variant: 'success',
    title: 'Success',
    children: 'Your action was completed successfully!'
  }
}

export const Warning: Story = {
  args: {
    variant: 'warning',
    title: 'Warning',
    children: 'Please review the following information before proceeding.'
  }
}

export const Destructive: Story = {
  args: {
    variant: 'destructive',
    title: 'Error',
    children: 'An error occurred while processing your request.'
  }
}

// Size variations
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <Alert variant="info" size="sm" title="Small Alert">
        This is a small alert with compact spacing.
      </Alert>
      
      <Alert variant="success" size="md" title="Medium Alert">
        This is a medium alert with standard spacing.
      </Alert>
      
      <Alert variant="warning" size="lg" title="Large Alert">
        This is a large alert with generous spacing and larger text.
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Available alert sizes: small, medium, and large.'
      }
    }
  }
}

// Dismissible alerts
export const Dismissible: Story = {
  render: () => {
    const [alerts, setAlerts] = useState([
      { id: 1, variant: 'info' as const, title: 'Info Alert', message: 'This alert can be dismissed.' },
      { id: 2, variant: 'success' as const, title: 'Success Alert', message: 'Operation completed successfully.' },
      { id: 3, variant: 'warning' as const, title: 'Warning Alert', message: 'Please be careful with this action.' }
    ])

    const dismissAlert = (id: number) => {
      setAlerts(alerts.filter(alert => alert.id !== id))
    }

    return (
      <div className="space-y-4">
        {alerts.map(alert => (
          <Alert
            key={alert.id}
            variant={alert.variant}
            title={alert.title}
            dismissible
            onDismiss={() => dismissAlert(alert.id)}
          >
            {alert.message}
          </Alert>
        ))}
        
        {alerts.length === 0 && (
          <Alert variant="success" title="All Clear">
            All alerts have been dismissed.
          </Alert>
        )}
        
        <Button 
          onClick={() => setAlerts([
            { id: Date.now() + 1, variant: 'info', title: 'New Info', message: 'A new info alert.' },
            { id: Date.now() + 2, variant: 'warning', title: 'New Warning', message: 'A new warning alert.' }
          ])}
          variant="outline"
          size="sm"
        >
          Restore Alerts
        </Button>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Dismissible alerts that can be closed by the user.'
      }
    }
  }
}

// With custom icons
export const WithCustomIcons: Story = {
  render: () => (
    <div className="space-y-4">
      <Alert variant="info" icon={<RefreshCw className="w-4 h-4" />} title="Syncing">
        Your data is being synchronized with the server.
      </Alert>
      
      <Alert variant="warning" icon={<AlertTriangle className="w-4 h-4" />} title="Maintenance">
        System maintenance is scheduled for tonight at 2 AM.
      </Alert>
      
      <Alert variant="success" showIcon={false} title="No Icon">
        This alert doesn't show an icon.
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Alerts with custom icons or no icons at all.'
      }
    }
  }
}

// With actions
export const WithActions: Story = {
  render: () => (
    <div className="space-y-4">
      <Alert 
        variant="info" 
        title="Update Available"
        actions={
          <div className="flex gap-2">
            <Button size="sm" variant="outline">
              Update Now
            </Button>
            <Button size="sm" variant="ghost">
              Remind Later
            </Button>
          </div>
        }
      >
        A new version of the application is available.
      </Alert>
      
      <Alert 
        variant="warning" 
        title="Storage Almost Full"
        actions={
          <div className="flex gap-2">
            <Button size="sm" variant="primary">
              Upgrade Plan
            </Button>
            <Button size="sm" variant="outline">
              Manage Files
            </Button>
          </div>
        }
      >
        You're using 95% of your storage space.
      </Alert>
      
      <Alert 
        variant="destructive" 
        title="Connection Lost"
        actions={
          <Button size="sm" variant="outline">
            Retry Connection
          </Button>
        }
      >
        Unable to connect to the server. Please check your internet connection.
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Alerts with action buttons for user interaction.'
      }
    }
  }
}

// Compound components
export const CompoundComponents: Story = {
  render: () => (
    <div className="space-y-4">
      <Alert variant="info">
        <Alert.Title>Using Compound Components</Alert.Title>
        <Alert.Description>
          You can use Alert.Title and Alert.Description for more control over the content structure.
          This approach gives you flexibility in layout and styling.
        </Alert.Description>
        <Alert.Actions>
          <Button size="sm" variant="outline">Learn More</Button>
          <Button size="sm" variant="ghost">Dismiss</Button>
        </Alert.Actions>
      </Alert>
      
      <Alert variant="success">
        <Alert.Title level={2}>Custom Heading Level</Alert.Title>
        <Alert.Description>
          The title component accepts a level prop to control the heading semantics.
        </Alert.Description>
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Using compound components for more control over alert structure.'
      }
    }
  }
}

// Toast alerts
export const ToastAlerts: Story = {
  render: () => {
    const [toasts, setToasts] = useState<Array<{ id: number; variant: any; title: string; message: string }>>([])

    const addToast = (variant: 'info' | 'success' | 'warning' | 'destructive', title: string, message: string) => {
      const id = Date.now()
      setToasts(prev => [...prev, { id, variant, title, message }])
    }

    const removeToast = (id: number) => {
      setToasts(prev => prev.filter(toast => toast.id !== id))
    }

    return (
      <div className="space-y-4">
        <div className="flex gap-2 flex-wrap">
          <Button 
            onClick={() => addToast('info', 'Info', 'This is an info toast')}
            variant="outline"
            size="sm"
          >
            Show Info Toast
          </Button>
          
          <Button 
            onClick={() => addToast('success', 'Success', 'Operation completed successfully')}
            variant="outline"
            size="sm"
          >
            Show Success Toast
          </Button>
          
          <Button 
            onClick={() => addToast('warning', 'Warning', 'Please be careful')}
            variant="outline"
            size="sm"
          >
            Show Warning Toast
          </Button>
          
          <Button 
            onClick={() => addToast('destructive', 'Error', 'Something went wrong')}
            variant="outline"
            size="sm"
          >
            Show Error Toast
          </Button>
        </div>
        
        {toasts.map(toast => (
          <ToastAlert
            key={toast.id}
            variant={toast.variant}
            title={toast.title}
            position="top-right"
            autoHideDuration={3000}
            onDismiss={() => removeToast(toast.id)}
          >
            {toast.message}
          </ToastAlert>
        ))}
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Toast alerts that appear in fixed positions and auto-dismiss after a timeout.'
      }
    }
  }
}

// Banner alerts
export const BannerAlerts: Story = {
  render: () => (
    <div className="space-y-4">
      <BannerAlert variant="info" title="System Maintenance">
        Scheduled maintenance will occur tonight from 2:00 AM to 4:00 AM EST.
      </BannerAlert>
      
      <BannerAlert 
        variant="warning" 
        title="Limited Functionality"
        dismissible
        actions={
          <Button size="sm" variant="outline">
            Learn More
          </Button>
        }
      >
        Some features may be temporarily unavailable due to ongoing updates.
      </BannerAlert>
      
      <BannerAlert 
        variant="success" 
        title="Welcome!"
        sticky
      >
        Thank you for joining our platform. Explore the features to get started.
      </BannerAlert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Banner alerts for page-level notifications and announcements.'
      }
    }
  }
}

// All variants showcase
export const AllVariants: Story = {
  render: () => (
    <div className="space-y-4">
      <Alert variant="default" title="Default">
        This is a default alert for general information.
      </Alert>
      
      <Alert variant="info" title="Information">
        This is an info alert for helpful context.
      </Alert>
      
      <Alert variant="success" title="Success">
        This is a success alert for positive feedback.
      </Alert>
      
      <Alert variant="warning" title="Warning">
        This is a warning alert for cautionary information.
      </Alert>
      
      <Alert variant="destructive" title="Error">
        This is an error alert for problems and failures.
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Showcase of all alert variants with their distinctive styling.'
      }
    }
  }
}

// Accessibility demonstration
export const AccessibilityDemo: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Screen Reader Announcement</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Alerts use the role="alert" attribute to be announced by screen readers.
        </p>
        <Alert variant="info" title="Accessible Alert">
          This alert will be announced to screen readers immediately when it appears.
        </Alert>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Focus moves to dismiss button and action buttons with Tab key.
        </p>
        <Alert 
          variant="warning" 
          title="Interactive Alert"
          dismissible
          actions={
            <Button size="sm" variant="outline">
              Take Action
            </Button>
          }
        >
          Use Tab to navigate to the action button and dismiss button.
        </Alert>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Accessibility features including ARIA attributes and keyboard navigation.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    variant: 'info',
    size: 'md',
    title: 'Alert Title',
    children: 'This is the alert message content.',
    dismissible: false,
    showIcon: true
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different alert configurations.'
      }
    }
  }
}