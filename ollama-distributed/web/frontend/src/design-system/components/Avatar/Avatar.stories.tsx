import type { Meta, StoryObj } from '@storybook/react'
import { Avatar, AvatarGroup, AvatarWithStatus } from './Avatar'
import { Button } from '../Button/Button'
import { Card } from '../Card/Card'
import { Badge } from '../Badge/Badge'
import { User, Settings, Crown, Star } from 'lucide-react'

const meta: Meta<typeof Avatar> = {
  title: 'Design System/Avatar',
  component: Avatar,
  parameters: {
    docs: {
      description: {
        component: 'Avatar component with comprehensive fallback system, multiple sizes, shapes, and status indicators. Includes automatic initial generation and accessibility features.'
      }
    }
  },
  argTypes: {
    size: {
      control: 'select',
      options: ['xs', 'sm', 'md', 'lg', 'xl', '2xl'],
      description: 'Size variant of the avatar'
    },
    shape: {
      control: 'select',
      options: ['circle', 'rounded', 'square'],
      description: 'Shape variant of the avatar'
    },
    variant: {
      control: 'select',
      options: ['default', 'outline', 'ghost'],
      description: 'Style variant of the avatar'
    },
    src: {
      control: 'text',
      description: 'Image source URL'
    },
    name: {
      control: 'text',
      description: 'Name for generating initials and alt text'
    },
    colorful: {
      control: 'boolean',
      description: 'Whether to use colored background for initials'
    },
    loading: {
      control: 'boolean',
      description: 'Loading state'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Avatar>

// Basic avatars
export const Default: Story = {
  args: {
    src: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop&crop=face',
    alt: 'User avatar',
    name: 'John Doe'
  }
}

export const WithInitials: Story = {
  args: {
    name: 'Jane Smith'
  }
}

export const IconFallback: Story = {
  args: {
    alt: 'User avatar'
  }
}

// Size variants
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <Avatar size="xs" name="XS" />
      <Avatar size="sm" name="SM" />
      <Avatar size="md" name="MD" />
      <Avatar size="lg" name="LG" />
      <Avatar size="xl" name="XL" />
      <Avatar size="2xl" name="2XL" />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Available size variants from extra small to 2xl.'
      }
    }
  }
}

// Shape variants
export const Shapes: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <div className="text-center">
        <Avatar shape="circle" name="Circle" className="mb-2" />
        <p className="text-sm text-muted-foreground">Circle</p>
      </div>
      <div className="text-center">
        <Avatar shape="rounded" name="Rounded" className="mb-2" />
        <p className="text-sm text-muted-foreground">Rounded</p>
      </div>
      <div className="text-center">
        <Avatar shape="square" name="Square" className="mb-2" />
        <p className="text-sm text-muted-foreground">Square</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different shape variants for various design needs.'
      }
    }
  }
}

// Style variants
export const Variants: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <div className="text-center">
        <Avatar variant="default" name="Default" className="mb-2" />
        <p className="text-sm text-muted-foreground">Default</p>
      </div>
      <div className="text-center">
        <Avatar variant="outline" name="Outline" className="mb-2" />
        <p className="text-sm text-muted-foreground">Outline</p>
      </div>
      <div className="text-center">
        <Avatar variant="ghost" name="Ghost" className="mb-2" />
        <p className="text-sm text-muted-foreground">Ghost</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Style variants with different visual treatments.'
      }
    }
  }
}

// Loading states
export const LoadingStates: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <div className="text-center">
        <Avatar loading={true} className="mb-2" />
        <p className="text-sm text-muted-foreground">Loading</p>
      </div>
      <div className="text-center">
        <Avatar 
          src="https://broken-url.jpg" 
          name="Failed Load"
          className="mb-2" 
        />
        <p className="text-sm text-muted-foreground">Failed Image</p>
      </div>
      <div className="text-center">
        <Avatar name="Success" className="mb-2" />
        <p className="text-sm text-muted-foreground">Initials Fallback</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different loading and fallback states.'
      }
    }
  }
}

// Custom fallback content
export const CustomFallback: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <Avatar
        fallback={<Crown className="w-5 h-5" />}
        colorful={false}
        fallbackBg="hsl(var(--warning))"
      />
      <Avatar
        fallback={<Settings className="w-5 h-5" />}
        colorful={false}
        fallbackBg="hsl(var(--primary))"
      />
      <Avatar
        fallback={<Star className="w-5 h-5" />}
        colorful={false}
        fallbackBg="hsl(var(--success))"
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Custom fallback content with icons and colors.'
      }
    }
  }
}

// Avatar with status
export const WithStatus: Story = {
  render: () => (
    <div className="flex items-center space-x-6">
      <div className="text-center">
        <AvatarWithStatus
          name="Online User"
          status="online"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Online</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Away User"
          status="away"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Away</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Busy User"
          status="busy"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Busy</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Offline User"
          status="offline"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Offline</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Avatar with status indicators for different user states.'
      }
    }
  }
}

// Status positions
export const StatusPositions: Story = {
  render: () => (
    <div className="flex items-center space-x-6">
      <div className="text-center">
        <AvatarWithStatus
          name="Top Right"
          status="online"
          statusPosition="top-right"
          size="lg"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Top Right</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Bottom Right"
          status="away"
          statusPosition="bottom-right"
          size="lg"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Bottom Right</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Top Left"
          status="busy"
          statusPosition="top-left"
          size="lg"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Top Left</p>
      </div>
      <div className="text-center">
        <AvatarWithStatus
          name="Bottom Left"
          status="offline"
          statusPosition="bottom-left"
          size="lg"
          className="mb-2"
        />
        <p className="text-sm text-muted-foreground">Bottom Left</p>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different positions for status indicators.'
      }
    }
  }
}

// Avatar groups
export const Groups: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Default Group (max 3)</h3>
        <AvatarGroup>
          <Avatar name="John Doe" />
          <Avatar name="Jane Smith" />
          <Avatar name="Mike Johnson" />
          <Avatar name="Sarah Wilson" />
          <Avatar name="Alex Brown" />
        </AvatarGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Larger Group (max 5)</h3>
        <AvatarGroup max={5} spacing="normal">
          <Avatar name="Alice Cooper" />
          <Avatar name="Bob Martin" />
          <Avatar name="Carol Davis" />
          <Avatar name="David Lee" />
          <Avatar name="Emma Taylor" />
          <Avatar name="Frank Miller" />
          <Avatar name="Grace Kim" />
        </AvatarGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Tight Spacing</h3>
        <AvatarGroup spacing="tight" size="sm">
          <Avatar name="User 1" />
          <Avatar name="User 2" />
          <Avatar name="User 3" />
          <Avatar name="User 4" />
          <Avatar name="User 5" />
        </AvatarGroup>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Avatar groups with different configurations and spacing options.'
      }
    }
  }
}

// Color generation demo
export const ColorGeneration: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-3">Colorful Initials (Based on Name)</h3>
        <div className="flex items-center space-x-2">
          <Avatar name="Alice Johnson" />
          <Avatar name="Bob Smith" />
          <Avatar name="Carol Williams" />
          <Avatar name="David Brown" />
          <Avatar name="Emma Davis" />
          <Avatar name="Frank Miller" />
          <Avatar name="Grace Wilson" />
          <Avatar name="Henry Taylor" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Without Colors</h3>
        <div className="flex items-center space-x-2">
          <Avatar name="Alice Johnson" colorful={false} />
          <Avatar name="Bob Smith" colorful={false} />
          <Avatar name="Carol Williams" colorful={false} />
          <Avatar name="David Brown" colorful={false} />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Automatic color generation based on name for better visual distinction.'
      }
    }
  }
}

// Real-world examples
export const TeamMemberCard: Story = {
  render: () => (
    <Card className="max-w-sm">
      <Card.Header className="text-center">
        <AvatarWithStatus
          src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop&crop=face"
          name="John Doe"
          status="online"
          size="xl"
          className="mx-auto mb-4"
        />
        <Card.Title>John Doe</Card.Title>
        <Card.Description>Senior Frontend Developer</Card.Description>
      </Card.Header>
      <Card.Content>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Location:</span>
          <span>San Francisco, CA</span>
        </div>
        <div className="flex items-center justify-between text-sm mt-2">
          <span className="text-muted-foreground">Team:</span>
          <span>Engineering</span>
        </div>
      </Card.Content>
      <Card.Footer className="flex space-x-2">
        <Button size="sm" className="flex-1">
          <User className="w-4 h-4 mr-2" />
          Profile
        </Button>
        <Button size="sm" variant="outline" className="flex-1">
          Message
        </Button>
      </Card.Footer>
    </Card>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Team member card using avatar with status indicator.'
      }
    }
  }
}

export const CommentThread: Story = {
  render: () => (
    <div className="max-w-md space-y-4">
      <div className="flex space-x-3">
        <Avatar name="Sarah Wilson" size="sm" />
        <div className="flex-1">
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium">Sarah Wilson</span>
            <Badge variant="secondary" size="sm">Author</Badge>
            <span className="text-xs text-muted-foreground">2h ago</span>
          </div>
          <p className="text-sm mt-1">
            This looks great! I really like the new design direction we're taking.
          </p>
        </div>
      </div>
      
      <div className="flex space-x-3">
        <Avatar name="Mike Johnson" size="sm" />
        <div className="flex-1">
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium">Mike Johnson</span>
            <span className="text-xs text-muted-foreground">1h ago</span>
          </div>
          <p className="text-sm mt-1">
            Agreed! The color palette works really well with our brand.
          </p>
        </div>
      </div>
      
      <div className="flex space-x-3">
        <Avatar name="Emma Taylor" size="sm" />
        <div className="flex-1">
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium">Emma Taylor</span>
            <span className="text-xs text-muted-foreground">30m ago</span>
          </div>
          <p className="text-sm mt-1">
            Should we consider the accessibility implications of these colors?
          </p>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Comment thread with small avatars for user identification.'
      }
    }
  }
}

export const ProjectCollaborators: Story = {
  render: () => (
    <Card>
      <Card.Header>
        <Card.Title>Project Collaborators</Card.Title>
        <Card.Description>
          Team members working on this project
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <AvatarWithStatus
                name="Alice Cooper"
                status="online"
                size="sm"
              />
              <div>
                <p className="text-sm font-medium">Alice Cooper</p>
                <p className="text-xs text-muted-foreground">Project Manager</p>
              </div>
            </div>
            <Badge variant="success">Owner</Badge>
          </div>
          
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <AvatarWithStatus
                name="Bob Martin"
                status="away"
                size="sm"
              />
              <div>
                <p className="text-sm font-medium">Bob Martin</p>
                <p className="text-xs text-muted-foreground">Lead Developer</p>
              </div>
            </div>
            <Badge variant="default">Admin</Badge>
          </div>
          
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <AvatarWithStatus
                name="Carol Davis"
                status="online"
                size="sm"
              />
              <div>
                <p className="text-sm font-medium">Carol Davis</p>
                <p className="text-xs text-muted-foreground">UI Designer</p>
              </div>
            </div>
            <Badge variant="secondary">Member</Badge>
          </div>
        </div>
        
        <div className="mt-4 pt-4 border-t">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">
              Team ({3} members)
            </span>
            <AvatarGroup size="xs" max={4}>
              <Avatar name="Alice Cooper" />
              <Avatar name="Bob Martin" />
              <Avatar name="Carol Davis" />
              <Avatar name="David Lee" />
              <Avatar name="Emma Taylor" />
            </AvatarGroup>
          </div>
        </div>
      </Card.Content>
    </Card>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Project collaborators list with avatars, status indicators, and avatar group summary.'
      }
    }
  }
}

// Accessibility demo
export const AccessibilityFeatures: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Avatars are properly accessible with screen readers and support keyboard navigation when interactive.
        </p>
        <div className="flex items-center space-x-4">
          <Avatar 
            name="User 1" 
            aria-label="User 1's profile picture"
          />
          <Avatar 
            name="User 2"
            aria-label="User 2's profile picture"
          />
          <Avatar 
            name="User 3"
            aria-label="User 3's profile picture"
          />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">High Contrast Support</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Avatars maintain proper contrast ratios and work well in high contrast mode.
        </p>
        <div className="flex items-center space-x-4 high-contrast p-4 rounded-lg border">
          <Avatar name="HC User 1" />
          <Avatar name="HC User 2" variant="outline" />
          <Avatar name="HC User 3" variant="ghost" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Reduced Motion</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Avatars respect reduced motion preferences for loading states.
        </p>
        <div className="reduce-motion">
          <Avatar loading={true} />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Demonstrates accessibility features including keyboard navigation, high contrast support, and reduced motion preferences.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    name: 'John Doe',
    size: 'md',
    shape: 'circle',
    variant: 'default',
    colorful: true,
    loading: false
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different avatar configurations.'
      }
    }
  }
}