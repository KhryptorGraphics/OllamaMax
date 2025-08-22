import type { Meta, StoryObj } from '@storybook/react'
import { useState, useEffect } from 'react'
import { Progress, LinearProgress, CircularProgress, StepsProgress } from './Progress'
import { Button } from '../Button/Button'
import { Card } from '../Card/Card'
import { 
  Upload, 
  Download, 
  CheckCircle, 
  AlertCircle,
  Loader2,
  FileUp,
  Wifi,
  Battery,
  HardDrive
} from 'lucide-react'

const meta = {
  title: 'Design System/Progress',
  component: Progress,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
The Progress component provides visual feedback about the duration and progression of a process. 
It supports linear bars, circular indicators, and step-based progress tracking with full accessibility support.

## Features
- **Multiple types**: Linear, circular, and steps progress indicators
- **Size variants**: xs, sm, md, lg, xl for different use cases
- **Color variants**: Primary, secondary, success, warning, error, info, neutral
- **Animation modes**: Static, animated, striped, and indeterminate states
- **Value display**: Percentage, fraction, or custom formatting
- **Full accessibility**: WCAG 2.1 AA compliant with ARIA attributes
- **Dark mode**: Automatic theme adaptation
- **Performance**: Optimized animations and rendering
        `
      }
    }
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Progress>

export default meta
type Story = StoryObj<typeof meta>

// Basic Linear Progress
export const Default: Story = {
  args: {
    value: 60,
    variant: 'primary',
    size: 'md',
    showValue: true
  },
  render: (args) => (
    <div className="w-96">
      <LinearProgress {...args} />
    </div>
  )
}

// All Linear Variants
export const LinearVariants: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-2">Primary</h3>
        <LinearProgress value={70} variant="primary" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Secondary</h3>
        <LinearProgress value={60} variant="secondary" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Success</h3>
        <LinearProgress value={90} variant="success" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Warning</h3>
        <LinearProgress value={40} variant="warning" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Error</h3>
        <LinearProgress value={25} variant="error" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Info</h3>
        <LinearProgress value={55} variant="info" showValue />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Neutral</h3>
        <LinearProgress value={80} variant="neutral" showValue />
      </div>
    </div>
  )
}

// All Sizes
export const Sizes: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <div>
        <h3 className="text-xs font-medium mb-2">Extra Small (xs)</h3>
        <LinearProgress value={60} size="xs" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Small (sm)</h3>
        <LinearProgress value={60} size="sm" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Medium (md)</h3>
        <LinearProgress value={60} size="md" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Large (lg)</h3>
        <LinearProgress value={60} size="lg" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Extra Large (xl)</h3>
        <LinearProgress value={60} size="xl" />
      </div>
    </div>
  )
}

// Animated Progress
export const Animated: Story = {
  render: () => {
    const [progress, setProgress] = useState(0)
    
    useEffect(() => {
      const timer = setInterval(() => {
        setProgress((prev) => {
          if (prev >= 100) return 0
          return prev + 10
        })
      }, 500)
      
      return () => clearInterval(timer)
    }, [])
    
    return (
      <div className="w-96 space-y-6">
        <div>
          <h3 className="text-sm font-medium mb-2">Animated Progress</h3>
          <LinearProgress value={progress} animated showValue />
        </div>
        <div>
          <h3 className="text-sm font-medium mb-2">Striped Progress</h3>
          <LinearProgress value={progress} striped showValue variant="success" />
        </div>
        <div>
          <h3 className="text-sm font-medium mb-2">Animated Striped</h3>
          <LinearProgress value={progress} animated striped showValue variant="info" />
        </div>
      </div>
    )
  }
}

// Indeterminate Loading
export const Indeterminate: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-2">Indeterminate Linear</h3>
        <LinearProgress indeterminate variant="primary" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Indeterminate with Stripes</h3>
        <LinearProgress indeterminate striped variant="info" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Indeterminate Circular</h3>
        <div className="flex gap-4">
          <CircularProgress indeterminate size="sm" />
          <CircularProgress indeterminate size="md" />
          <CircularProgress indeterminate size="lg" />
        </div>
      </div>
    </div>
  )
}

// Value Display Options
export const ValueDisplay: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-2">Percentage (default)</h3>
        <LinearProgress value={75} showValue valueFormat="percentage" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Fraction</h3>
        <LinearProgress value={7} max={10} showValue valueFormat="fraction" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Custom Format (GB)</h3>
        <LinearProgress 
          value={1.5} 
          max={2} 
          showValue 
          formatValue={(val, max) => `${val}GB / ${max}GB`}
        />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Custom Label</h3>
        <LinearProgress value={60} showValue label="Processing..." />
      </div>
    </div>
  )
}

// Label Positions
export const LabelPositions: Story = {
  render: () => (
    <div className="w-96 space-y-8">
      <div>
        <h3 className="text-sm font-medium mb-2">Inside</h3>
        <LinearProgress value={75} showValue labelPosition="inside" size="xl" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Outside (default)</h3>
        <LinearProgress value={60} showValue labelPosition="outside" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Outside Start</h3>
        <LinearProgress value={45} showValue labelPosition="outside-start" />
      </div>
      <div>
        <h3 className="text-sm font-medium mb-2">Inline</h3>
        <LinearProgress value={80} showValue labelPosition="inline" />
      </div>
    </div>
  )
}

// Circular Progress
export const Circular: Story = {
  render: () => (
    <div className="flex flex-wrap gap-8">
      <div className="text-center">
        <CircularProgress value={25} size="xs" showValue />
        <p className="text-xs mt-2">Extra Small</p>
      </div>
      <div className="text-center">
        <CircularProgress value={50} size="sm" showValue variant="secondary" />
        <p className="text-xs mt-2">Small</p>
      </div>
      <div className="text-center">
        <CircularProgress value={75} size="md" showValue variant="success" />
        <p className="text-xs mt-2">Medium</p>
      </div>
      <div className="text-center">
        <CircularProgress value={85} size="lg" showValue variant="warning" />
        <p className="text-xs mt-2">Large</p>
      </div>
      <div className="text-center">
        <CircularProgress value={95} size="xl" showValue variant="error" />
        <p className="text-xs mt-2">Extra Large</p>
      </div>
    </div>
  )
}

// Circular Custom Styling
export const CircularCustom: Story = {
  render: () => (
    <div className="flex flex-wrap gap-8">
      <div className="text-center">
        <CircularProgress value={60} thickness={2} showValue />
        <p className="text-xs mt-2">Thin Stroke</p>
      </div>
      <div className="text-center">
        <CircularProgress value={70} thickness={8} showValue variant="success" />
        <p className="text-xs mt-2">Thick Stroke</p>
      </div>
      <div className="text-center">
        <CircularProgress value={80} showTrack={false} showValue variant="info" />
        <p className="text-xs mt-2">No Track</p>
      </div>
      <div className="text-center">
        <CircularProgress value={90} startAngle={0} showValue variant="warning" />
        <p className="text-xs mt-2">Start from Top</p>
      </div>
    </div>
  )
}

// Steps Progress
export const Steps: Story = {
  render: () => {
    const [currentStep, setCurrentStep] = useState(2)
    
    return (
      <div className="space-y-8">
        <div>
          <h3 className="text-sm font-medium mb-4">Horizontal Steps</h3>
          <StepsProgress 
            steps={5} 
            currentStep={currentStep}
            onStepClick={setCurrentStep}
          />
        </div>
        
        <div>
          <h3 className="text-sm font-medium mb-4">With Labels</h3>
          <StepsProgress 
            steps={4} 
            currentStep={currentStep}
            stepLabels={['Setup', 'Configure', 'Review', 'Deploy']}
            onStepClick={setCurrentStep}
            size="lg"
          />
        </div>
        
        <div className="flex gap-8">
          <div>
            <h3 className="text-sm font-medium mb-4">Vertical Steps</h3>
            <StepsProgress 
              steps={4} 
              currentStep={2}
              orientation="vertical"
              variant="success"
            />
          </div>
          
          <div>
            <h3 className="text-sm font-medium mb-4">Different Variants</h3>
            <div className="space-y-4">
              <StepsProgress steps={3} currentStep={1} variant="primary" />
              <StepsProgress steps={3} currentStep={1} variant="success" />
              <StepsProgress steps={3} currentStep={1} variant="warning" />
            </div>
          </div>
        </div>
      </div>
    )
  }
}

// File Upload Example
export const FileUpload: Story = {
  render: () => {
    const [uploadProgress, setUploadProgress] = useState(0)
    const [uploading, setUploading] = useState(false)
    
    const startUpload = () => {
      setUploading(true)
      setUploadProgress(0)
      
      const interval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 100) {
            clearInterval(interval)
            setUploading(false)
            return 100
          }
          return prev + Math.random() * 15
        })
      }, 300)
    }
    
    return (
      <Card className="w-96 p-6">
        <div className="flex items-center gap-3 mb-4">
          <FileUp className="w-8 h-8 text-primary-500" />
          <div className="flex-1">
            <h3 className="font-medium">document.pdf</h3>
            <p className="text-sm text-muted-foreground">2.4 MB</p>
          </div>
        </div>
        
        <LinearProgress 
          value={uploadProgress} 
          animated={uploading}
          striped={uploading}
          variant={uploadProgress === 100 ? 'success' : 'primary'}
          showValue
          formatValue={(val) => `${Math.round(val)}%`}
        />
        
        <div className="mt-4 flex gap-2">
          <Button 
            onClick={startUpload} 
            disabled={uploading}
            size="sm"
            leftIcon={uploading ? <Loader2 className="animate-spin" /> : <Upload />}
          >
            {uploading ? 'Uploading...' : 'Start Upload'}
          </Button>
          {uploadProgress === 100 && (
            <span className="flex items-center gap-1 text-success-600 text-sm">
              <CheckCircle className="w-4 h-4" />
              Upload complete
            </span>
          )}
        </div>
      </Card>
    )
  }
}

// Form Completion Example
export const FormCompletion: Story = {
  render: () => {
    const steps = ['Personal Info', 'Address', 'Payment', 'Review']
    const [currentStep, setCurrentStep] = useState(0)
    const progress = ((currentStep + 1) / steps.length) * 100
    
    return (
      <Card className="w-[500px] p-6">
        <h3 className="text-lg font-semibold mb-6">Account Setup</h3>
        
        <StepsProgress 
          steps={steps.length}
          currentStep={currentStep}
          stepLabels={steps}
          onStepClick={setCurrentStep}
          className="mb-6"
        />
        
        <div className="mb-6">
          <div className="flex justify-between text-sm mb-2">
            <span className="font-medium">{steps[currentStep]}</span>
            <span className="text-muted-foreground">Step {currentStep + 1} of {steps.length}</span>
          </div>
          <LinearProgress 
            value={progress} 
            variant="primary"
            animated
          />
        </div>
        
        <div className="p-8 bg-neutral-50 dark:bg-neutral-900 rounded-lg mb-6 text-center text-muted-foreground">
          Form content for: {steps[currentStep]}
        </div>
        
        <div className="flex justify-between">
          <Button 
            onClick={() => setCurrentStep(Math.max(0, currentStep - 1))}
            disabled={currentStep === 0}
            variant="outline"
          >
            Previous
          </Button>
          <Button 
            onClick={() => setCurrentStep(Math.min(steps.length - 1, currentStep + 1))}
            disabled={currentStep === steps.length - 1}
          >
            {currentStep === steps.length - 1 ? 'Complete' : 'Next'}
          </Button>
        </div>
      </Card>
    )
  }
}

// System Resources Dashboard
export const SystemResources: Story = {
  render: () => {
    const [cpu, setCpu] = useState(45)
    const [memory, setMemory] = useState(67)
    const [disk, setDisk] = useState(82)
    const [network, setNetwork] = useState(23)
    
    useEffect(() => {
      const timer = setInterval(() => {
        setCpu(40 + Math.random() * 30)
        setMemory(60 + Math.random() * 20)
        setDisk(80 + Math.random() * 10)
        setNetwork(15 + Math.random() * 40)
      }, 2000)
      
      return () => clearInterval(timer)
    }, [])
    
    return (
      <div className="grid grid-cols-2 gap-6 w-[600px]">
        <Card className="p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium">CPU Usage</span>
            <span className="text-2xl font-bold">{Math.round(cpu)}%</span>
          </div>
          <CircularProgress 
            value={cpu} 
            variant={cpu > 80 ? 'error' : cpu > 60 ? 'warning' : 'success'}
            animated
          />
        </Card>
        
        <Card className="p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium">Memory</span>
            <span className="text-2xl font-bold">{Math.round(memory)}%</span>
          </div>
          <CircularProgress 
            value={memory} 
            variant={memory > 80 ? 'error' : memory > 60 ? 'warning' : 'info'}
            animated
          />
        </Card>
        
        <Card className="p-4 col-span-2">
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <HardDrive className="w-4 h-4" />
                  <span className="text-sm font-medium">Disk Space</span>
                </div>
                <span className="text-sm">{Math.round(disk)}%</span>
              </div>
              <LinearProgress 
                value={disk} 
                variant={disk > 90 ? 'error' : disk > 75 ? 'warning' : 'primary'}
                size="sm"
              />
            </div>
            
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Wifi className="w-4 h-4" />
                  <span className="text-sm font-medium">Network</span>
                </div>
                <span className="text-sm">{Math.round(network)} Mbps</span>
              </div>
              <LinearProgress 
                value={network} 
                max={100}
                variant="info"
                size="sm"
                animated
                striped
              />
            </div>
          </div>
        </Card>
      </div>
    )
  }
}

// Accessibility Features
export const AccessibilityShowcase: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <Card className="p-4">
        <h3 className="font-medium mb-4">Screen Reader Support</h3>
        <LinearProgress 
          value={75} 
          aria-label="Loading user data"
          aria-describedby="loading-description"
          showValue
        />
        <p id="loading-description" className="text-sm text-muted-foreground mt-2">
          This progress bar has proper ARIA attributes for screen readers
        </p>
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-4">Keyboard Navigation</h3>
        <StepsProgress 
          steps={4}
          currentStep={1}
          onStepClick={(step) => console.log('Step clicked:', step)}
          stepLabels={['Start', 'Middle', 'Almost', 'Done']}
        />
        <p className="text-sm text-muted-foreground mt-2">
          Steps are keyboard accessible and can be activated with Enter/Space
        </p>
      </Card>
      
      <Card className="p-4">
        <h3 className="font-medium mb-4">High Contrast Mode</h3>
        <div className="space-y-3">
          <LinearProgress value={60} variant="primary" showValue />
          <LinearProgress value={40} variant="error" showValue />
          <LinearProgress value={80} variant="success" showValue />
        </div>
        <p className="text-sm text-muted-foreground mt-2">
          Colors meet WCAG AA standards for contrast
        </p>
      </Card>
    </div>
  )
}

// Performance Considerations
export const PerformanceDemo: Story = {
  render: () => {
    const [items, setItems] = useState(10)
    
    return (
      <div className="w-[600px] space-y-6">
        <Card className="p-4">
          <h3 className="font-medium mb-4">Optimized Rendering</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Multiple progress bars with smooth animations
          </p>
          
          <div className="mb-4">
            <Button 
              onClick={() => setItems(items + 10)} 
              size="sm"
              className="mr-2"
            >
              Add 10 More
            </Button>
            <span className="text-sm text-muted-foreground">
              Current: {items} progress bars
            </span>
          </div>
          
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {Array.from({ length: items }, (_, i) => (
              <LinearProgress 
                key={i}
                value={Math.random() * 100}
                variant={['primary', 'secondary', 'success', 'warning', 'error', 'info'][i % 6] as any}
                size="sm"
                animated
                striped={i % 2 === 0}
              />
            ))}
          </div>
        </Card>
      </div>
    )
  }
}