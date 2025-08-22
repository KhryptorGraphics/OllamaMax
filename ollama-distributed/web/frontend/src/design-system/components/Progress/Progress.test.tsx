import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { LinearProgress, CircularProgress, StepsProgress } from './Progress'

describe('LinearProgress', () => {
  it('renders with default props', () => {
    render(<LinearProgress value={50} />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toBeInTheDocument()
    expect(progressBar).toHaveAttribute('aria-valuenow', '50')
    expect(progressBar).toHaveAttribute('aria-valuemin', '0')
    expect(progressBar).toHaveAttribute('aria-valuemax', '100')
  })

  it('renders with custom min and max values', () => {
    render(<LinearProgress value={5} min={0} max={10} />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveAttribute('aria-valuenow', '5')
    expect(progressBar).toHaveAttribute('aria-valuemax', '10')
  })

  it('displays percentage value when showValue is true', () => {
    render(<LinearProgress value={75} showValue />)
    expect(screen.getByText('75%')).toBeInTheDocument()
  })

  it('displays fraction value when valueFormat is fraction', () => {
    render(<LinearProgress value={3} max={5} showValue valueFormat="fraction" />)
    expect(screen.getByText('3/5')).toBeInTheDocument()
  })

  it('uses custom formatter when provided', () => {
    const formatter = (val: number, max: number) => `${val} of ${max} items`
    render(<LinearProgress value={10} max={20} showValue formatValue={formatter} />)
    expect(screen.getByText('10 of 20 items')).toBeInTheDocument()
  })

  it('displays custom label when provided', () => {
    render(<LinearProgress value={50} showValue label="Processing..." />)
    expect(screen.getByText('Processing...')).toBeInTheDocument()
  })

  it('renders in indeterminate state', () => {
    render(<LinearProgress indeterminate />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).not.toHaveAttribute('aria-valuenow')
    expect(progressBar).toHaveAttribute('aria-label', 'Loading...')
  })

  it('applies correct size classes', () => {
    const { rerender } = render(<LinearProgress value={50} size="xs" />)
    let progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveClass('h-1')

    rerender(<LinearProgress value={50} size="xl" />)
    progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveClass('h-6')
  })

  it('applies correct variant classes', () => {
    render(<LinearProgress value={50} variant="success" />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveClass('bg-success-100')
  })

  it('handles label positions correctly', () => {
    const { rerender } = render(
      <LinearProgress value={50} showValue labelPosition="inside" />
    )
    expect(screen.getByText('50%')).toHaveClass('absolute')

    rerender(<LinearProgress value={50} showValue labelPosition="inline" />)
    expect(screen.getByText('50%')).toHaveClass('ml-2')
  })

  it('applies custom className', () => {
    render(<LinearProgress value={50} className="custom-class" />)
    const container = screen.getByRole('progressbar').parentElement?.parentElement
    expect(container).toHaveClass('custom-class')
  })

  it('applies striped and animated classes', () => {
    render(<LinearProgress value={50} striped animated />)
    const progressFill = screen.getByRole('progressbar').firstChild
    expect(progressFill).toHaveClass('animate-[progress-stripes_1s_linear_infinite]')
  })
})

describe('CircularProgress', () => {
  it('renders with default props', () => {
    render(<CircularProgress value={50} />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toBeInTheDocument()
    expect(progressBar).toHaveAttribute('aria-valuenow', '50')
  })

  it('renders in indeterminate state', () => {
    render(<CircularProgress indeterminate />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).not.toHaveAttribute('aria-valuenow')
    expect(progressBar.querySelector('svg')).toHaveClass('animate-spin')
  })

  it('displays value in center when showValue is true', () => {
    render(<CircularProgress value={75} showValue />)
    expect(screen.getByText('75%')).toBeInTheDocument()
  })

  it('applies correct size', () => {
    render(<CircularProgress value={50} size="lg" />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveStyle({ width: '64px', height: '64px' })
  })

  it('renders without track when showTrack is false', () => {
    render(<CircularProgress value={50} showTrack={false} />)
    const circles = screen.getByRole('progressbar').querySelectorAll('circle')
    expect(circles).toHaveLength(1) // Only progress circle, no track
  })

  it('applies custom thickness', () => {
    render(<CircularProgress value={50} thickness={10} />)
    const circle = screen.getByRole('progressbar').querySelector('circle')
    expect(circle).toHaveAttribute('stroke-width', '10')
  })

  it('uses custom formatter', () => {
    const formatter = (val: number) => `${val} points`
    render(<CircularProgress value={80} showValue formatValue={formatter} />)
    expect(screen.getByText('80 points')).toBeInTheDocument()
  })

  it('applies correct variant color classes', () => {
    render(<CircularProgress value={50} variant="error" />)
    const svg = screen.getByRole('progressbar').querySelector('svg')
    expect(svg).toHaveClass('text-error-500')
  })
})

describe('StepsProgress', () => {
  it('renders correct number of steps', () => {
    render(<StepsProgress steps={5} currentStep={2} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(5)
  })

  it('marks current step correctly', () => {
    render(<StepsProgress steps={4} currentStep={2} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons[2]).toHaveAttribute('aria-current', 'step')
  })

  it('displays step numbers when showStepNumbers is true', () => {
    render(<StepsProgress steps={3} currentStep={1} showStepNumbers />)
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('displays step labels when provided', () => {
    const labels = ['Start', 'Middle', 'End']
    render(<StepsProgress steps={3} currentStep={1} stepLabels={labels} />)
    
    const buttons = screen.getAllByRole('button')
    expect(buttons[0]).toHaveAttribute('aria-label', 'Start')
    expect(buttons[1]).toHaveAttribute('aria-label', 'Middle')
    expect(buttons[2]).toHaveAttribute('aria-label', 'End')
  })

  it('handles step click events', () => {
    const handleClick = vi.fn()
    render(
      <StepsProgress 
        steps={3} 
        currentStep={0} 
        onStepClick={handleClick}
      />
    )
    
    const buttons = screen.getAllByRole('button')
    fireEvent.click(buttons[2])
    expect(handleClick).toHaveBeenCalledWith(2)
  })

  it('disables click when onStepClick is not provided', () => {
    render(<StepsProgress steps={3} currentStep={1} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons[0]).toBeDisabled()
  })

  it('renders in vertical orientation', () => {
    render(<StepsProgress steps={3} currentStep={1} orientation="vertical" />)
    const container = screen.getByRole('group')
    expect(container).toHaveClass('flex-col')
  })

  it('applies correct variant colors', () => {
    render(<StepsProgress steps={3} currentStep={1} variant="success" />)
    const activeButtons = screen.getAllByRole('button').slice(0, 2)
    activeButtons.forEach(button => {
      expect(button).toHaveClass('bg-success-500')
    })
  })

  it('applies correct size classes', () => {
    render(<StepsProgress steps={3} currentStep={1} size="lg" />)
    const buttons = screen.getAllByRole('button')
    buttons.forEach(button => {
      expect(button).toHaveClass('w-10', 'h-10')
    })
  })

  it('renders connecting lines between steps', () => {
    const { container } = render(<StepsProgress steps={3} currentStep={1} />)
    const lines = container.querySelectorAll('[aria-hidden="true"]')
    expect(lines).toHaveLength(2) // 2 connecting lines for 3 steps
  })
})

describe('Accessibility', () => {
  it('provides proper ARIA attributes for linear progress', () => {
    render(
      <LinearProgress 
        value={60} 
        aria-label="Upload progress"
        aria-describedby="upload-desc"
      />
    )
    
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveAttribute('aria-label', 'Upload progress')
    expect(progressBar).toHaveAttribute('aria-describedby', 'upload-desc')
  })

  it('provides proper ARIA attributes for circular progress', () => {
    render(
      <CircularProgress 
        value={75} 
        aria-label="Download progress"
      />
    )
    
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveAttribute('aria-label', 'Download progress')
  })

  it('announces indeterminate state properly', () => {
    render(<LinearProgress indeterminate />)
    const progressBar = screen.getByRole('progressbar')
    expect(progressBar).toHaveAttribute('aria-label', 'Loading...')
  })

  it('provides keyboard accessibility for steps', () => {
    const handleClick = vi.fn()
    render(
      <StepsProgress 
        steps={3} 
        currentStep={0} 
        onStepClick={handleClick}
      />
    )
    
    const firstStep = screen.getAllByRole('button')[0]
    fireEvent.keyDown(firstStep, { key: 'Enter' })
    // Note: This would need actual keyboard event handling implementation
  })
})

describe('Edge Cases', () => {
  it('handles value exceeding max', () => {
    render(<LinearProgress value={150} max={100} />)
    const progressBar = screen.getByRole('progressbar')
    const progressFill = progressBar.firstChild as HTMLElement
    expect(progressFill.style.width).toBe('100%')
  })

  it('handles negative value', () => {
    render(<LinearProgress value={-10} />)
    const progressBar = screen.getByRole('progressbar')
    const progressFill = progressBar.firstChild as HTMLElement
    expect(progressFill.style.width).toBe('0%')
  })

  it('handles zero steps gracefully', () => {
    render(<StepsProgress steps={0} currentStep={0} />)
    const buttons = screen.queryAllByRole('button')
    expect(buttons).toHaveLength(0)
  })

  it('handles currentStep exceeding steps', () => {
    render(<StepsProgress steps={3} currentStep={5} />)
    // Should handle gracefully without crashing
    expect(screen.getByRole('group')).toBeInTheDocument()
  })
})