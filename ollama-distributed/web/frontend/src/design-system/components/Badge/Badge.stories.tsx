import type { Meta, StoryObj } from '@storybook/react'
import { Badge, BadgeGroup, StatusBadge, NotificationBadge } from './Badge'
import { Star, Tag, User, Calendar, CheckCircle, AlertTriangle } from 'lucide-react'
import { useState } from 'react'

const meta: Meta<typeof Badge> = {
  title: 'Design System/Badge',
  component: Badge,
  parameters: {
    docs: {
      description: {
        component: 'Badge component for labels, status indicators, and notifications. Supports multiple variants, sizes, icons, and interactive states including removable badges.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'secondary', 'success', 'warning', 'destructive', 'outline', 'ghost'],
      description: 'Visual style variant of the badge'
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size of the badge'
    },
    interactive: {
      control: 'boolean',
      description: 'Whether the badge is clickable'
    },
    removable: {
      control: 'boolean',
      description: 'Whether the badge can be removed'
    },
    dot: {
      control: 'boolean',
      description: 'Show as dot indicator instead of full badge'
    },
    pulse: {
      control: 'boolean',
      description: 'Add pulse animation for notifications'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Badge>

// Default story
export const Default: Story = {
  args: {
    children: 'Badge'
  }
}

// Variant showcase
export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="default">Default</Badge>
      <Badge variant="secondary">Secondary</Badge>
      <Badge variant="success">Success</Badge>
      <Badge variant="warning">Warning</Badge>
      <Badge variant="destructive">Destructive</Badge>
      <Badge variant="outline">Outline</Badge>
      <Badge variant="ghost">Ghost</Badge>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'All available badge variants with their distinctive colors and styles.'
      }
    }
  }
}

// Size variations
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <div className="flex flex-col items-center gap-2">
        <Badge size="sm">Small</Badge>
        <span className="text-xs text-muted-foreground">Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Badge size="md">Medium</Badge>
        <span className="text-xs text-muted-foreground">Medium</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Badge size="lg">Large</Badge>
        <span className="text-xs text-muted-foreground">Large</span>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different badge sizes for various use cases and visual hierarchy.'
      }
    }
  }
}

// With icons
export const WithIcons: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-3">Icons with Text</h3>
        <div className="flex flex-wrap gap-2">
          <Badge icon={<Star className="w-3 h-3" />}>Featured</Badge>
          <Badge icon={<Tag className="w-3 h-3" />} variant="secondary">Tagged</Badge>
          <Badge icon={<User className="w-3 h-3" />} variant="success">Verified</Badge>
          <Badge icon={<Calendar className="w-3 h-3" />} variant="outline">Scheduled</Badge>
          <Badge icon={<CheckCircle className="w-3 h-3" />} variant="success">Completed</Badge>
          <Badge icon={<AlertTriangle className="w-3 h-3" />} variant="warning">Warning</Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Icon Only Badges</h3>
        <div className="flex flex-wrap gap-2">
          <Badge icon={<Star className="w-3 h-3" />} />
          <Badge icon={<Tag className="w-3 h-3" />} variant="secondary" />
          <Badge icon={<User className="w-3 h-3" />} variant="success" />
          <Badge icon={<Calendar className="w-3 h-3" />} variant="outline" />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Badges with icons for enhanced visual communication.'
      }
    }
  }
}

// Interactive badges
export const Interactive: Story = {
  render: () => {
    const [clickCount, setClickCount] = useState(0)

    return (
      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-medium mb-3">Clickable Badges</h3>
          <div className="flex flex-wrap gap-2">
            <Badge 
              interactive 
              onClick={() => setClickCount(count => count + 1)}
              variant="outline"
            >
              Click me ({clickCount})
            </Badge>
            <Badge interactive variant="secondary">Category</Badge>
            <Badge interactive variant="success">Available</Badge>
            <Badge interactive variant="warning">Pending</Badge>
          </div>
        </div>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive badges that respond to clicks and keyboard navigation.'
      }
    }
  }
}

// Removable badges
export const Removable: Story = {
  render: () => {
    const [tags, setTags] = useState([
      { id: 1, text: 'React', variant: 'default' as const },
      { id: 2, text: 'TypeScript', variant: 'secondary' as const },
      { id: 3, text: 'Tailwind', variant: 'success' as const },
      { id: 4, text: 'Storybook', variant: 'outline' as const }
    ])

    const removeTag = (id: number) => {
      setTags(tags.filter(tag => tag.id !== id))
    }

    const resetTags = () => {
      setTags([
        { id: 1, text: 'React', variant: 'default' as const },
        { id: 2, text: 'TypeScript', variant: 'secondary' as const },
        { id: 3, text: 'Tailwind', variant: 'success' as const },
        { id: 4, text: 'Storybook', variant: 'outline' as const }
      ])
    }

    return (
      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-medium mb-3">Removable Tags</h3>
          <div className="flex flex-wrap gap-2 min-h-[32px]">
            {tags.map(tag => (
              <Badge
                key={tag.id}
                variant={tag.variant}
                removable
                onRemove={() => removeTag(tag.id)}
                icon={<Tag className="w-3 h-3" />}
              >
                {tag.text}
              </Badge>
            ))}
            {tags.length === 0 && (
              <p className="text-sm text-muted-foreground italic">
                All tags removed
              </p>
            )}
          </div>
          
          {tags.length === 0 && (
            <button
              onClick={resetTags}
              className="mt-2 text-xs text-primary hover:underline"
            >
              Reset tags
            </button>
          )}
        </div>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Removable badges with close buttons for tag management.'
      }
    }
  }
}

// Dot indicators
export const DotIndicators: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-3">Status Dots</h3>
        <div className="space-y-2">
          <Badge dot variant="success">Online</Badge>
          <Badge dot variant="warning">Away</Badge>
          <Badge dot variant="destructive">Offline</Badge>
          <Badge dot variant="secondary">Idle</Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Pulse Animation</h3>
        <div className="space-y-2">
          <Badge dot variant="success" pulse>Live</Badge>
          <Badge dot variant="warning" pulse>Recording</Badge>
          <Badge dot variant="destructive" pulse>Alert</Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Dot Only</h3>
        <div className="flex gap-2">
          <Badge dot variant="success" />
          <Badge dot variant="warning" />
          <Badge dot variant="destructive" />
          <Badge dot variant="secondary" />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Dot indicators for status and live updates with optional pulse animation.'
      }
    }
  }
}

// Badge groups
export const BadgeGroups: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Horizontal Group</h3>
        <BadgeGroup spacing="md">
          <Badge variant="default">Frontend</Badge>
          <Badge variant="secondary">Backend</Badge>
          <Badge variant="success">DevOps</Badge>
          <Badge variant="outline">Design</Badge>
        </BadgeGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Vertical Group</h3>
        <BadgeGroup orientation="vertical" spacing="sm">
          <Badge variant="default">High Priority</Badge>
          <Badge variant="warning">Medium Priority</Badge>
          <Badge variant="secondary">Low Priority</Badge>
        </BadgeGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Truncated Group (max 3)</h3>
        <BadgeGroup max={3} spacing="sm">
          <Badge variant="default">JavaScript</Badge>
          <Badge variant="secondary">TypeScript</Badge>
          <Badge variant="success">React</Badge>
          <Badge variant="outline">Vue</Badge>
          <Badge variant="warning">Angular</Badge>
          <Badge variant="destructive">Svelte</Badge>
        </BadgeGroup>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Badge groups for organizing multiple badges with different layouts and truncation.'
      }
    }
  }
}

// Status badges
export const StatusBadges: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-3">User Status</h3>
        <div className="flex flex-wrap gap-2">
          <StatusBadge status="online" />
          <StatusBadge status="offline" />
          <StatusBadge status="busy" />
          <StatusBadge status="away" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Activity Status</h3>
        <div className="flex flex-wrap gap-2">
          <StatusBadge status="active" />
          <StatusBadge status="inactive" />
          <StatusBadge status="pending" />
          <StatusBadge status="approved" />
          <StatusBadge status="rejected" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Custom Status Text</h3>
        <div className="flex flex-wrap gap-2">
          <StatusBadge status="online">Available</StatusBadge>
          <StatusBadge status="busy">In Meeting</StatusBadge>
          <StatusBadge status="away">Lunch Break</StatusBadge>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Pre-configured status badges for common application states.'
      }
    }
  }
}

// Notification badges
export const NotificationBadges: Story = {
  render: () => {
    const [count1, setCount1] = useState(5)
    const [count2, setCount2] = useState(99)
    const [count3, setCount3] = useState(150)
    const [count4, setCount4] = useState(0)

    return (
      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-medium mb-3">Notification Counts</h3>
          <div className="flex items-center gap-6">
            <div className="relative">
              <div className="w-8 h-8 bg-muted rounded-lg flex items-center justify-center">
                📧
              </div>
              <div className="absolute -top-2 -right-2">
                <NotificationBadge count={count1} />
              </div>
            </div>
            
            <div className="relative">
              <div className="w-8 h-8 bg-muted rounded-lg flex items-center justify-center">
                🔔
              </div>
              <div className="absolute -top-2 -right-2">
                <NotificationBadge count={count2} />
              </div>
            </div>
            
            <div className="relative">
              <div className="w-8 h-8 bg-muted rounded-lg flex items-center justify-center">
                💬
              </div>
              <div className="absolute -top-2 -right-2">
                <NotificationBadge count={count3} max={99} />
              </div>
            </div>
            
            <div className="relative">
              <div className="w-8 h-8 bg-muted rounded-lg flex items-center justify-center">
                ⭐
              </div>
              <div className="absolute -top-2 -right-2">
                <NotificationBadge count={count4} showZero />
              </div>
            </div>
          </div>
        </div>
        
        <div>
          <h3 className="text-sm font-medium mb-3">Controls</h3>
          <div className="flex gap-2 flex-wrap">
            <button
              onClick={() => setCount1(c => Math.max(0, c - 1))}
              className="px-2 py-1 text-xs bg-muted rounded hover:bg-muted/80"
            >
              -1 Email
            </button>
            <button
              onClick={() => setCount1(c => c + 1)}
              className="px-2 py-1 text-xs bg-muted rounded hover:bg-muted/80"
            >
              +1 Email
            </button>
            <button
              onClick={() => setCount2(c => Math.max(0, c - 10))}
              className="px-2 py-1 text-xs bg-muted rounded hover:bg-muted/80"
            >
              -10 Notifications
            </button>
            <button
              onClick={() => setCount2(c => c + 10)}
              className="px-2 py-1 text-xs bg-muted rounded hover:bg-muted/80"
            >
              +10 Notifications
            </button>
          </div>
        </div>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Notification badges for displaying counts with automatic truncation (99+).'
      }
    }
  }
}

// Real-world examples
export const RealWorldExamples: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">E-commerce Product</h3>
        <div className="p-4 border rounded-lg">
          <div className="flex items-start justify-between mb-2">
            <h4 className="font-medium">Wireless Headphones</h4>
            <Badge variant="success" size="sm">In Stock</Badge>
          </div>
          <p className="text-sm text-muted-foreground mb-3">
            High-quality wireless headphones with noise cancellation
          </p>
          <div className="flex gap-2 flex-wrap">
            <Badge variant="outline" size="sm">Bluetooth</Badge>
            <Badge variant="outline" size="sm">Noise Canceling</Badge>
            <Badge variant="outline" size="sm">25h Battery</Badge>
            <Badge variant="warning" size="sm">Limited Edition</Badge>
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">User Profile</h3>
        <div className="p-4 border rounded-lg">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
              <User className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-medium">John Doe</span>
                <Badge dot variant="success" pulse>Online</Badge>
              </div>
              <p className="text-sm text-muted-foreground">Software Engineer</p>
            </div>
          </div>
          <div className="flex gap-2 flex-wrap">
            <Badge variant="secondary" size="sm" icon={<CheckCircle className="w-3 h-3" />}>Verified</Badge>
            <Badge variant="outline" size="sm">Pro Member</Badge>
            <Badge variant="success" size="sm">5+ Years</Badge>
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Task Management</h3>
        <div className="space-y-2">
          <div className="flex items-center justify-between p-3 border rounded">
            <span>Complete design system documentation</span>
            <div className="flex items-center gap-2">
              <Badge variant="warning" size="sm">High Priority</Badge>
              <Badge variant="outline" size="sm">Design</Badge>
              <Badge dot variant="warning">Due Soon</Badge>
            </div>
          </div>
          
          <div className="flex items-center justify-between p-3 border rounded">
            <span>Fix responsive layout issues</span>
            <div className="flex items-center gap-2">
              <Badge variant="destructive" size="sm">Bug</Badge>
              <Badge variant="outline" size="sm">Frontend</Badge>
              <Badge dot variant="destructive" pulse>Critical</Badge>
            </div>
          </div>
          
          <div className="flex items-center justify-between p-3 border rounded">
            <span>Update API documentation</span>
            <div className="flex items-center gap-2">
              <Badge variant="secondary" size="sm">Low Priority</Badge>
              <Badge variant="outline" size="sm">Backend</Badge>
              <Badge variant="success" size="sm">Ready</Badge>
            </div>
          </div>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Real-world usage examples in e-commerce, user profiles, and task management contexts.'
      }
    }
  }
}

// Accessibility demonstration
export const AccessibilityDemo: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Interactive badges can be focused and activated with keyboard.
        </p>
        <div className="flex gap-2 flex-wrap">
          <Badge interactive onClick={() => alert('First badge clicked!')}>
            Tab to me first
          </Badge>
          <Badge interactive onClick={() => alert('Second badge clicked!')}>
            Then to me
          </Badge>
          <Badge interactive onClick={() => alert('Third badge clicked!')}>
            Finally here
          </Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Screen Reader Support</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Badges provide proper context and states for assistive technologies.
        </p>
        <div className="flex gap-2 flex-wrap">
          <Badge variant="success" aria-label="Status: Active user account">
            Active
          </Badge>
          <Badge variant="warning" aria-label="Warning: Account requires verification">
            Needs Verification
          </Badge>
          <Badge variant="destructive" aria-label="Error: Account suspended">
            Suspended
          </Badge>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Accessibility features including keyboard navigation and screen reader support.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    variant: 'default',
    size: 'md',
    children: 'Interactive Badge',
    interactive: false,
    removable: false,
    dot: false,
    pulse: false
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different badge configurations.'
      }
    }
  }
}