import type { Meta, StoryObj } from '@storybook/react'
import React, { useState } from 'react'
import { Spinner, SpinnerOverlay, LoadingButton, Skeleton } from './Spinner'
import { Button } from '../Button/Button'
import { Card } from '../Card/Card'

/**
 * Spinner Component Stories
 * 
 * The Spinner component provides various loading indicators with different
 * animation types, sizes, colors, and display modes. It's fully accessible
 * and integrates with the design system tokens.
 */
const meta = {
  title: 'Design System/Components/Spinner',
  component: Spinner,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
The Spinner component is a versatile loading indicator that supports multiple animation types,
sizes, and color variants. It's designed to provide clear visual feedback during loading states
while maintaining accessibility standards.

## Features
- **Multiple animation types**: spin, pulse, dots, bars, ring, ripple
- **Size variants**: xs, sm, md, lg, xl
- **Color variants**: Integrates with design system color tokens
- **Speed control**: slow, normal, fast animation speeds
- **Loading text**: Optional label with flexible positioning
- **Display modes**: inline, centered, or fullscreen overlay
- **Accessibility**: Full ARIA support with screen reader announcements
- **Performance**: Optimized animations with GPU acceleration

## Usage Guidelines
- Use spinners to indicate loading states that take more than 300ms
- Choose animation type based on context (dots for inline, spin for overlays)
- Always provide descriptive loading text for better UX
- Use skeleton loaders for content that's loading in place
- Avoid multiple spinners on the same screen
        `
      }
    }
  },
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: { type: 'select' },
      options: ['xs', 'sm', 'md', 'lg', 'xl'],
      description: 'Size of the spinner',
      table: {
        type: { summary: 'xs | sm | md | lg | xl' },
        defaultValue: { summary: 'md' }
      }
    },
    variant: {
      control: { type: 'select' },
      options: ['primary', 'secondary', 'success', 'warning', 'danger', 'info', 'neutral', 'current', 'white', 'black'],
      description: 'Color variant of the spinner',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'primary' }
      }
    },
    type: {
      control: { type: 'select' },
      options: ['spin', 'pulse', 'dots', 'bars', 'ring', 'ripple'],
      description: 'Animation type',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'spin' }
      }
    },
    speed: {
      control: { type: 'select' },
      options: ['slow', 'normal', 'fast'],
      description: 'Animation speed',
      table: {
        type: { summary: 'slow | normal | fast' },
        defaultValue: { summary: 'normal' }
      }
    },
    label: {
      control: { type: 'text' },
      description: 'Loading text to display'
    },
    labelPosition: {
      control: { type: 'select' },
      options: ['top', 'bottom', 'left', 'right'],
      description: 'Position of the label relative to spinner',
      table: {
        type: { summary: 'top | bottom | left | right' },
        defaultValue: { summary: 'bottom' }
      }
    },
    overlay: {
      control: { type: 'boolean' },
      description: 'Show as fullscreen overlay'
    },
    centered: {
      control: { type: 'boolean' },
      description: 'Center in parent container'
    },
    inline: {
      control: { type: 'boolean' },
      description: 'Display inline with content'
    }
  }
} satisfies Meta<typeof Spinner>

export default meta
type Story = StoryObj<typeof meta>

/**
 * Default spinner with primary color and medium size
 */
export const Default: Story = {
  args: {}
}

/**
 * All size variants of the spinner
 */
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-8">
      <div className="text-center">
        <Spinner size="xs" />
        <p className="mt-2 text-xs text-muted-foreground">Extra Small</p>
      </div>
      <div className="text-center">
        <Spinner size="sm" />
        <p className="mt-2 text-xs text-muted-foreground">Small</p>
      </div>
      <div className="text-center">
        <Spinner size="md" />
        <p className="mt-2 text-xs text-muted-foreground">Medium</p>
      </div>
      <div className="text-center">
        <Spinner size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Large</p>
      </div>
      <div className="text-center">
        <Spinner size="xl" />
        <p className="mt-2 text-xs text-muted-foreground">Extra Large</p>
      </div>
    </div>
  )
}

/**
 * Different animation types available
 */
export const AnimationTypes: Story = {
  render: () => (
    <div className="grid grid-cols-3 gap-8">
      <div className="text-center">
        <Spinner type="spin" size="lg" />
        <p className="mt-4 text-sm font-medium">Spin</p>
        <p className="text-xs text-muted-foreground">Classic rotation</p>
      </div>
      <div className="text-center">
        <Spinner type="pulse" size="lg" />
        <p className="mt-4 text-sm font-medium">Pulse</p>
        <p className="text-xs text-muted-foreground">Breathing effect</p>
      </div>
      <div className="text-center">
        <Spinner type="dots" size="lg" />
        <p className="mt-4 text-sm font-medium">Dots</p>
        <p className="text-xs text-muted-foreground">Bouncing dots</p>
      </div>
      <div className="text-center">
        <Spinner type="bars" size="lg" />
        <p className="mt-4 text-sm font-medium">Bars</p>
        <p className="text-xs text-muted-foreground">Wave effect</p>
      </div>
      <div className="text-center">
        <Spinner type="ring" size="lg" />
        <p className="mt-4 text-sm font-medium">Ring</p>
        <p className="text-xs text-muted-foreground">Circular progress</p>
      </div>
      <div className="text-center">
        <Spinner type="ripple" size="lg" />
        <p className="mt-4 text-sm font-medium">Ripple</p>
        <p className="text-xs text-muted-foreground">Expanding rings</p>
      </div>
    </div>
  )
}

/**
 * Color variants matching the design system
 */
export const ColorVariants: Story = {
  render: () => (
    <div className="grid grid-cols-5 gap-6">
      <div className="text-center">
        <Spinner variant="primary" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Primary</p>
      </div>
      <div className="text-center">
        <Spinner variant="secondary" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Secondary</p>
      </div>
      <div className="text-center">
        <Spinner variant="success" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Success</p>
      </div>
      <div className="text-center">
        <Spinner variant="warning" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Warning</p>
      </div>
      <div className="text-center">
        <Spinner variant="danger" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Danger</p>
      </div>
      <div className="text-center">
        <Spinner variant="info" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Info</p>
      </div>
      <div className="text-center">
        <Spinner variant="neutral" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Neutral</p>
      </div>
      <div className="text-center bg-slate-800 p-4 rounded">
        <Spinner variant="white" size="lg" />
        <p className="mt-2 text-xs text-white">White</p>
      </div>
      <div className="text-center">
        <Spinner variant="black" size="lg" />
        <p className="mt-2 text-xs text-muted-foreground">Black</p>
      </div>
      <div className="text-center text-purple-500">
        <Spinner variant="current" size="lg" />
        <p className="mt-2 text-xs">Current Color</p>
      </div>
    </div>
  )
}

/**
 * Animation speed variations
 */
export const SpeedVariations: Story = {
  render: () => (
    <div className="flex items-center gap-8">
      <div className="text-center">
        <Spinner speed="slow" size="lg" />
        <p className="mt-2 text-sm font-medium">Slow</p>
        <p className="text-xs text-muted-foreground">1.5s duration</p>
      </div>
      <div className="text-center">
        <Spinner speed="normal" size="lg" />
        <p className="mt-2 text-sm font-medium">Normal</p>
        <p className="text-xs text-muted-foreground">1s duration</p>
      </div>
      <div className="text-center">
        <Spinner speed="fast" size="lg" />
        <p className="mt-2 text-sm font-medium">Fast</p>
        <p className="text-xs text-muted-foreground">0.5s duration</p>
      </div>
    </div>
  )
}

/**
 * Spinner with loading text in different positions
 */
export const WithLabel: Story = {
  render: () => (
    <div className="grid grid-cols-2 gap-8">
      <div className="text-center p-4 border rounded-lg">
        <Spinner label="Loading content..." labelPosition="bottom" />
        <p className="mt-4 text-xs text-muted-foreground">Label Bottom</p>
      </div>
      <div className="text-center p-4 border rounded-lg">
        <Spinner label="Loading content..." labelPosition="top" />
        <p className="mt-4 text-xs text-muted-foreground">Label Top</p>
      </div>
      <div className="text-center p-4 border rounded-lg">
        <Spinner label="Loading..." labelPosition="right" inline />
        <p className="mt-4 text-xs text-muted-foreground">Label Right (Inline)</p>
      </div>
      <div className="text-center p-4 border rounded-lg">
        <Spinner label="Loading..." labelPosition="left" inline />
        <p className="mt-4 text-xs text-muted-foreground">Label Left (Inline)</p>
      </div>
    </div>
  )
}

/**
 * Fullscreen overlay spinner
 */
export const OverlaySpinner: Story = {
  render: () => {
    const [isLoading, setIsLoading] = useState(false)
    
    return (
      <div className="space-y-4">
        <Button onClick={() => setIsLoading(true)}>
          Show Overlay Spinner
        </Button>
        <p className="text-sm text-muted-foreground">
          Click the button to show a fullscreen loading overlay
        </p>
        
        {isLoading && (
          <SpinnerOverlay
            label="Loading your content..."
            size="xl"
            onOverlayClick={() => setIsLoading(false)}
            closeOnClick
          />
        )}
      </div>
    )
  }
}

/**
 * Loading button with integrated spinner
 */
export const LoadingButtonExample: Story = {
  render: () => {
    const [isLoading, setIsLoading] = useState(false)
    
    const handleClick = () => {
      setIsLoading(true)
      setTimeout(() => setIsLoading(false), 3000)
    }
    
    return (
      <div className="space-y-4">
        <LoadingButton
          isLoading={isLoading}
          loadingText="Processing..."
          onClick={handleClick}
        >
          Submit Form
        </LoadingButton>
        
        <p className="text-sm text-muted-foreground">
          Click to see loading state (3 seconds)
        </p>
      </div>
    )
  }
}

/**
 * Spinner in card loading state
 */
export const CardLoadingState: Story = {
  render: () => (
    <div className="grid grid-cols-2 gap-4 w-[600px]">
      <Card className="p-6">
        <div className="flex flex-col items-center justify-center h-32">
          <Spinner size="lg" type="dots" />
          <p className="mt-4 text-sm text-muted-foreground">Loading data...</p>
        </div>
      </Card>
      
      <Card className="p-6">
        <div className="space-y-3">
          <Skeleton variant="text" width="60%" />
          <Skeleton variant="text" />
          <Skeleton variant="text" width="80%" />
          <div className="flex gap-2 pt-2">
            <Skeleton variant="rectangular" width={60} height={32} />
            <Skeleton variant="rectangular" width={60} height={32} />
          </div>
        </div>
      </Card>
    </div>
  )
}

/**
 * Skeleton loader variations
 */
export const SkeletonLoaders: Story = {
  render: () => (
    <div className="space-y-6 w-[400px]">
      <div>
        <h3 className="text-sm font-medium mb-3">Text Skeleton</h3>
        <div className="space-y-2">
          <Skeleton variant="text" />
          <Skeleton variant="text" width="80%" />
          <Skeleton variant="text" width="60%" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Avatar Skeleton</h3>
        <div className="flex items-center gap-3">
          <Skeleton variant="circular" width={40} height={40} />
          <div className="flex-1 space-y-2">
            <Skeleton variant="text" width="50%" />
            <Skeleton variant="text" width="30%" />
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Card Skeleton</h3>
        <Card className="p-4">
          <Skeleton variant="rectangular" height={120} className="mb-3" />
          <Skeleton variant="text" className="mb-2" />
          <Skeleton variant="text" width="60%" />
        </Card>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Wave Animation</h3>
        <div className="space-y-2">
          <Skeleton variant="text" animation="wave" />
          <Skeleton variant="text" animation="wave" width="75%" />
        </div>
      </div>
    </div>
  )
}

/**
 * Inline spinners in text and buttons
 */
export const InlineUsage: Story = {
  render: () => (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <span>Saving your changes</span>
        <Spinner size="sm" inline type="dots" />
      </div>
      
      <div className="flex items-center gap-2 text-success-600">
        <Spinner size="sm" variant="success" inline />
        <span>Upload complete!</span>
      </div>
      
      <div className="space-x-2">
        <Button disabled>
          <Spinner size="sm" variant="white" inline className="mr-2" />
          Processing...
        </Button>
        
        <Button variant="outline" disabled>
          <Spinner size="sm" inline type="dots" className="mr-2" />
          Loading...
        </Button>
      </div>
    </div>
  )
}

/**
 * Real-world loading scenarios
 */
export const RealWorldExamples: Story = {
  render: () => (
    <div className="space-y-8">
      {/* Data Table Loading */}
      <div className="border rounded-lg p-4">
        <h3 className="font-medium mb-4">Data Table Loading</h3>
        <div className="space-y-2">
          <div className="flex items-center justify-between p-2 bg-muted/50">
            <Skeleton variant="text" width={150} />
            <Skeleton variant="text" width={100} />
            <Skeleton variant="text" width={80} />
          </div>
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between p-2">
              <Skeleton variant="text" width={150} />
              <Skeleton variant="text" width={100} />
              <Skeleton variant="text" width={80} />
            </div>
          ))}
        </div>
      </div>
      
      {/* Form Submission */}
      <div className="border rounded-lg p-4">
        <h3 className="font-medium mb-4">Form Submission</h3>
        <div className="space-y-4">
          <input
            type="text"
            placeholder="Enter your email"
            className="w-full px-3 py-2 border rounded-md"
            disabled
          />
          <LoadingButton
            isLoading
            loadingText="Submitting..."
            className="w-full"
          >
            Submit
          </LoadingButton>
        </div>
      </div>
      
      {/* Image Gallery Loading */}
      <div className="border rounded-lg p-4">
        <h3 className="font-medium mb-4">Image Gallery Loading</h3>
        <div className="grid grid-cols-3 gap-2">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="aspect-square bg-muted rounded-md flex items-center justify-center">
              <Spinner type="pulse" variant="neutral" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

/**
 * Accessibility features demonstration
 */
export const AccessibilityFeatures: Story = {
  render: () => (
    <div className="space-y-6">
      <Card className="p-4">
        <h3 className="font-medium mb-2">Screen Reader Support</h3>
        <p className="text-sm text-muted-foreground mb-4">
          All spinners include proper ARIA attributes for screen readers
        </p>
        <Spinner 
          label="Loading user data..." 
          screenReaderText="Please wait while we load your information"
        />
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-2">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Loading buttons maintain focus and are keyboard accessible
        </p>
        <LoadingButton isLoading loadingText="Processing...">
          Tab to focus this button
        </LoadingButton>
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-2">Motion Preferences</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Respects prefers-reduced-motion for users sensitive to motion
        </p>
        <div className="flex gap-4">
          <Spinner type="spin" />
          <Spinner type="dots" />
          <Spinner type="pulse" />
        </div>
      </Card>
    </div>
  )
}

/**
 * Performance considerations
 */
export const PerformanceOptimization: Story = {
  render: () => (
    <div className="space-y-6">
      <Card className="p-4">
        <h3 className="font-medium mb-2">GPU Accelerated Animations</h3>
        <p className="text-sm text-muted-foreground mb-4">
          All animations use CSS transforms for optimal performance
        </p>
        <div className="flex gap-4">
          <Spinner type="spin" speed="fast" />
          <Spinner type="ring" speed="fast" />
          <Spinner type="ripple" speed="fast" />
        </div>
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-2">Conditional Rendering</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Spinners are only rendered when visible to save resources
        </p>
        <Spinner visible={true} label="Only renders when visible=true" />
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-2">Lightweight Variants</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Use simpler animations for better performance on low-end devices
        </p>
        <div className="flex gap-4">
          <div className="text-center">
            <Spinner type="pulse" />
            <p className="text-xs mt-2">Pulse (Lightest)</p>
          </div>
          <div className="text-center">
            <Spinner type="dots" />
            <p className="text-xs mt-2">Dots (Light)</p>
          </div>
          <div className="text-center">
            <Spinner type="spin" />
            <p className="text-xs mt-2">Spin (Standard)</p>
          </div>
        </div>
      </Card>
    </div>
  )
}