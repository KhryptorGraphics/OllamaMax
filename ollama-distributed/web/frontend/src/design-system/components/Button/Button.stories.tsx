import type { Meta, StoryObj } from '@storybook/react'
import { Button, ButtonGroup, IconButton, ToggleButton } from './Button'
import { Play, Download, Heart, Settings, Plus, ArrowRight, ExternalLink } from 'lucide-react'

const meta: Meta<typeof Button> = {
  title: 'Design System/Button',
  component: Button,
  parameters: {
    docs: {
      description: {
        component: 'A versatile button component with multiple variants, states, and accessibility features. Built with class-variance-authority for consistent styling and responsive design.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline', 'ghost', 'link', 'destructive'],
      description: 'Visual style variant of the button'
    },
    size: {
      control: 'select',
      options: ['xs', 'sm', 'md', 'lg', 'xl', 'icon'],
      description: 'Size of the button'
    },
    loading: {
      control: 'boolean',
      description: 'Show loading spinner and disable interaction'
    },
    disabled: {
      control: 'boolean',
      description: 'Disable the button'
    },
    fullWidth: {
      control: 'boolean',
      description: 'Make button take full width of container'
    },
    loadingText: {
      control: 'text',
      description: 'Text to show during loading state'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Button>

// Default story
export const Default: Story = {
  args: {
    children: 'Button'
  }
}

// Variant stories
export const Primary: Story = {
  args: {
    variant: 'primary',
    children: 'Primary Button'
  }
}

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    children: 'Secondary Button'
  }
}

export const Outline: Story = {
  args: {
    variant: 'outline',
    children: 'Outline Button'
  }
}

export const Ghost: Story = {
  args: {
    variant: 'ghost',
    children: 'Ghost Button'
  }
}

export const Link: Story = {
  args: {
    variant: 'link',
    children: 'Link Button'
  }
}

export const Destructive: Story = {
  args: {
    variant: 'destructive',
    children: 'Destructive Button'
  }
}

// Size variations
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4 flex-wrap">
      <Button size="xs">Extra Small</Button>
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
      <Button size="xl">Extra Large</Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Available button sizes from extra small to extra large.'
      }
    }
  }
}

// State stories
export const Loading: Story = {
  args: {
    loading: true,
    children: 'Loading...'
  }
}

export const LoadingWithText: Story = {
  args: {
    loading: true,
    loadingText: 'Processing...',
    children: 'Submit'
  }
}

export const Disabled: Story = {
  args: {
    disabled: true,
    children: 'Disabled Button'
  }
}

// Icon examples
export const WithLeftIcon: Story = {
  args: {
    leftIcon: <Play className="w-4 h-4" />,
    children: 'Play Video'
  }
}

export const WithRightIcon: Story = {
  args: {
    rightIcon: <Download className="w-4 h-4" />,
    children: 'Download'
  }
}

export const WithBothIcons: Story = {
  args: {
    leftIcon: <ExternalLink className="w-4 h-4" />,
    rightIcon: <ArrowRight className="w-4 h-4" />,
    children: 'Open External'
  }
}

// Full width
export const FullWidth: Story = {
  args: {
    fullWidth: true,
    children: 'Full Width Button'
  }
}

// All variants showcase
export const AllVariants: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="flex gap-2 flex-wrap">
        <Button variant="primary">Primary</Button>
        <Button variant="secondary">Secondary</Button>
        <Button variant="outline">Outline</Button>
        <Button variant="ghost">Ghost</Button>
        <Button variant="link">Link</Button>
        <Button variant="destructive">Destructive</Button>
      </div>
      
      <div className="flex gap-2 flex-wrap">
        <Button variant="primary" disabled>Primary</Button>
        <Button variant="secondary" disabled>Secondary</Button>
        <Button variant="outline" disabled>Outline</Button>
        <Button variant="ghost" disabled>Ghost</Button>
        <Button variant="link" disabled>Link</Button>
        <Button variant="destructive" disabled>Destructive</Button>
      </div>
      
      <div className="flex gap-2 flex-wrap">
        <Button variant="primary" loading>Primary</Button>
        <Button variant="secondary" loading>Secondary</Button>
        <Button variant="outline" loading>Outline</Button>
        <Button variant="ghost" loading>Ghost</Button>
        <Button variant="destructive" loading>Destructive</Button>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Showcase of all button variants in normal, disabled, and loading states.'
      }
    }
  }
}

// Button Group examples
export const ButtonGroupExample: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Horizontal Button Group</h3>
        <ButtonGroup>
          <Button variant="outline" leftIcon={<Plus />}>Add</Button>
          <Button variant="outline" leftIcon={<Settings />}>Settings</Button>
          <Button variant="outline" leftIcon={<Download />}>Export</Button>
        </ButtonGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Vertical Button Group</h3>
        <ButtonGroup orientation="vertical">
          <Button variant="ghost" className="justify-start">Dashboard</Button>
          <Button variant="ghost" className="justify-start">Analytics</Button>
          <Button variant="ghost" className="justify-start">Reports</Button>
          <Button variant="ghost" className="justify-start">Settings</Button>
        </ButtonGroup>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Different Spacing</h3>
        <div className="space-y-3">
          <ButtonGroup spacing="sm">
            <Button size="sm">Small</Button>
            <Button size="sm">Spacing</Button>
            <Button size="sm">Tight</Button>
          </ButtonGroup>
          
          <ButtonGroup spacing="lg">
            <Button size="sm">Large</Button>
            <Button size="sm">Spacing</Button>
            <Button size="sm">Loose</Button>
          </ButtonGroup>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Button groups for organizing related actions with different orientations and spacing options.'
      }
    }
  }
}

// Icon Button examples
export const IconButtonExample: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-3">Icon Button Sizes</h3>
        <div className="flex items-center gap-2">
          <IconButton 
            icon={<Heart />} 
            size="xs" 
            aria-label="Like (extra small)"
            variant="outline"
          />
          <IconButton 
            icon={<Heart />} 
            size="sm" 
            aria-label="Like (small)"
            variant="outline"
          />
          <IconButton 
            icon={<Heart />} 
            size="md" 
            aria-label="Like (medium)"
            variant="outline"
          />
          <IconButton 
            icon={<Heart />} 
            size="lg" 
            aria-label="Like (large)"
            variant="outline"
          />
          <IconButton 
            icon={<Heart />} 
            size="xl" 
            aria-label="Like (extra large)"
            variant="outline"
          />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Icon Button Variants</h3>
        <div className="flex items-center gap-2">
          <IconButton icon={<Settings />} variant="primary" aria-label="Settings" />
          <IconButton icon={<Settings />} variant="secondary" aria-label="Settings" />
          <IconButton icon={<Settings />} variant="outline" aria-label="Settings" />
          <IconButton icon={<Settings />} variant="ghost" aria-label="Settings" />
          <IconButton icon={<Settings />} variant="destructive" aria-label="Settings" />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Icon buttons for actions that only need an icon. Always include aria-label for accessibility.'
      }
    }
  }
}

// Toggle Button examples
export const ToggleButtonExample: Story = {
  render: () => {
    const [pressed1, setPressed1] = React.useState(false)
    const [pressed2, setPressed2] = React.useState(true)
    const [pressed3, setPressed3] = React.useState(false)

    return (
      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-medium mb-3">Toggle Buttons</h3>
          <div className="flex items-center gap-2">
            <ToggleButton 
              pressed={pressed1}
              onPressedChange={setPressed1}
              leftIcon={<Heart />}
            >
              Favorite
            </ToggleButton>
            
            <ToggleButton 
              pressed={pressed2}
              onPressedChange={setPressed2}
              leftIcon={<Download />}
            >
              Downloaded
            </ToggleButton>
            
            <ToggleButton 
              pressed={pressed3}
              onPressedChange={setPressed3}
              leftIcon={<Plus />}
            >
              Follow
            </ToggleButton>
          </div>
        </div>
        
        <div>
          <h3 className="text-sm font-medium mb-3">Toggle Button Group</h3>
          <ButtonGroup>
            <ToggleButton pressed={pressed1} onPressedChange={setPressed1}>
              Bold
            </ToggleButton>
            <ToggleButton pressed={pressed2} onPressedChange={setPressed2}>
              Italic
            </ToggleButton>
            <ToggleButton pressed={pressed3} onPressedChange={setPressed3}>
              Underline
            </ToggleButton>
          </ButtonGroup>
        </div>
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Toggle buttons for binary states like favorites, selections, or formatting options.'
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
          Use Tab to navigate between buttons, Enter or Space to activate them.
        </p>
        <div className="flex gap-2 flex-wrap">
          <Button>First Button</Button>
          <Button variant="secondary">Second Button</Button>
          <Button variant="outline">Third Button</Button>
          <IconButton icon={<Settings />} aria-label="Settings" />
          <Button disabled>Disabled Button</Button>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Loading States</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Loading buttons are properly disabled and announced to screen readers.
        </p>
        <div className="flex gap-2 flex-wrap">
          <Button loading loadingText="Saving...">Save Changes</Button>
          <Button loading variant="secondary">Loading...</Button>
          <Button loading variant="destructive">Deleting...</Button>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Accessibility features including keyboard navigation, ARIA attributes, and screen reader support.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    variant: 'primary',
    size: 'md',
    children: 'Interactive Button',
    disabled: false,
    loading: false,
    fullWidth: false,
    loadingText: 'Loading...'
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different button configurations.'
      }
    }
  }
}