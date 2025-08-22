/**
 * [FRONTEND-UPDATE] Slider Component Stories
 * Comprehensive examples demonstrating all slider features and use cases
 */

import React, { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { Slider } from './Slider'

const meta: Meta<typeof Slider> = {
  title: 'Design System/Components/Slider',
  component: Slider,
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        component: 'A comprehensive slider component supporting single values, ranges, multiple orientations, and full accessibility.'
      }
    }
  },
  argTypes: {
    value: {
      control: false,
      description: 'Controlled value (number or [number, number] for range)'
    },
    defaultValue: {
      control: false,
      description: 'Default value for uncontrolled usage'
    },
    onChange: {
      action: 'changed',
      description: 'Called when value changes during interaction'
    },
    onChangeEnd: {
      action: 'changeEnd',
      description: 'Called when interaction ends (mouse up, touch end)'
    },
    min: {
      control: { type: 'number' },
      description: 'Minimum value'
    },
    max: {
      control: { type: 'number' },
      description: 'Maximum value'
    },
    step: {
      control: { type: 'number' },
      description: 'Step increment'
    },
    marks: {
      control: { type: 'boolean' },
      description: 'Show marks on the track'
    },
    orientation: {
      control: { type: 'radio' },
      options: ['horizontal', 'vertical'],
      description: 'Slider orientation'
    },
    size: {
      control: { type: 'radio' },
      options: ['sm', 'md', 'lg'],
      description: 'Size variant'
    },
    variant: {
      control: { type: 'radio' },
      options: ['primary', 'secondary', 'success', 'warning', 'error'],
      description: 'Color variant'
    },
    showTooltip: {
      control: { type: 'radio' },
      options: [false, true, 'always', 'hover', 'focus'],
      description: 'Tooltip display mode'
    },
    disabled: {
      control: { type: 'boolean' },
      description: 'Disabled state'
    },
    readOnly: {
      control: { type: 'boolean' },
      description: 'Read-only state'
    },
    inverted: {
      control: { type: 'boolean' },
      description: 'Invert slider direction'
    },
    trackClickable: {
      control: { type: 'boolean' },
      description: 'Allow clicking on track to set value'
    }
  }
}

export default meta
type Story = StoryObj<typeof Slider>

// Basic Examples
export const Default: Story = {
  args: {
    defaultValue: 50,
    min: 0,
    max: 100
  }
}

export const Range: Story = {
  args: {
    defaultValue: [25, 75],
    min: 0,
    max: 100
  },
  parameters: {
    docs: {
      description: {
        story: 'Range slider with two handles for selecting a value range'
      }
    }
  }
}

// Size Variants
export const Sizes: Story = {
  render: () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <div>
        <h4 style={{ marginBottom: 16 }}>Small</h4>
        <Slider size="sm" defaultValue={30} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Medium (Default)</h4>
        <Slider size="md" defaultValue={50} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Large</h4>
        <Slider size="lg" defaultValue={70} />
      </div>
    </div>
  )
}

// Color Variants
export const Variants: Story = {
  render: () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <div>
        <h4 style={{ marginBottom: 16 }}>Primary</h4>
        <Slider variant="primary" defaultValue={60} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Secondary</h4>
        <Slider variant="secondary" defaultValue={60} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Success</h4>
        <Slider variant="success" defaultValue={60} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Warning</h4>
        <Slider variant="warning" defaultValue={60} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Error</h4>
        <Slider variant="error" defaultValue={60} />
      </div>
    </div>
  )
}

// Orientation
export const Vertical: Story = {
  args: {
    orientation: 'vertical',
    defaultValue: 50,
    min: 0,
    max: 100
  },
  decorators: [
    (Story) => (
      <div style={{ height: 400, display: 'flex', alignItems: 'center', gap: 32 }}>
        <Story />
        <div>
          <h4>Vertical Range Slider</h4>
          <Slider orientation="vertical" defaultValue={[20, 80]} />
        </div>
      </div>
    )
  ]
}

// Steps and Marks
export const StepsAndMarks: Story = {
  render: () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 48 }}>
      <div>
        <h4 style={{ marginBottom: 24 }}>Step: 10</h4>
        <Slider defaultValue={30} step={10} min={0} max={100} />
      </div>
      <div>
        <h4 style={{ marginBottom: 24 }}>Auto Marks</h4>
        <Slider defaultValue={50} marks={true} min={0} max={100} />
      </div>
      <div>
        <h4 style={{ marginBottom: 24 }}>Custom Marks with Labels</h4>
        <Slider
          defaultValue={25}
          min={0}
          max={100}
          marks={[
            { value: 0, label: 'Min' },
            { value: 25, label: 'Low' },
            { value: 50, label: 'Mid' },
            { value: 75, label: 'High' },
            { value: 100, label: 'Max' }
          ]}
        />
      </div>
    </div>
  )
}

// Tooltip Modes
export const TooltipModes: Story = {
  render: () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <div>
        <h4 style={{ marginBottom: 16 }}>Always Visible</h4>
        <Slider showTooltip="always" defaultValue={40} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>On Hover</h4>
        <Slider showTooltip="hover" defaultValue={60} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>On Focus</h4>
        <Slider showTooltip="focus" defaultValue={80} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Custom Format</h4>
        <Slider 
          showTooltip="always" 
          defaultValue={50}
          formatTooltip={(value) => `${value}%`}
        />
      </div>
    </div>
  )
}

// States
export const States: Story = {
  render: () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <div>
        <h4 style={{ marginBottom: 16 }}>Normal</h4>
        <Slider defaultValue={50} />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Disabled</h4>
        <Slider defaultValue={50} disabled />
      </div>
      <div>
        <h4 style={{ marginBottom: 16 }}>Read Only</h4>
        <Slider defaultValue={50} readOnly />
      </div>
    </div>
  )
}

// Controlled Component
export const Controlled: Story = {
  render: () => {
    const [value, setValue] = useState(50)
    const [rangeValue, setRangeValue] = useState<[number, number]>([25, 75])
    
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
        <div>
          <h4 style={{ marginBottom: 16 }}>Single Value: {value}</h4>
          <Slider
            value={value}
            onChange={(v) => setValue(v as number)}
            min={0}
            max={100}
          />
          <div style={{ marginTop: 16, display: 'flex', gap: 8 }}>
            <button onClick={() => setValue(0)}>Min</button>
            <button onClick={() => setValue(50)}>Center</button>
            <button onClick={() => setValue(100)}>Max</button>
          </div>
        </div>
        
        <div>
          <h4 style={{ marginBottom: 16 }}>Range: [{rangeValue[0]}, {rangeValue[1]}]</h4>
          <Slider
            value={rangeValue}
            onChange={(v) => setRangeValue(v as [number, number])}
            min={0}
            max={100}
          />
          <div style={{ marginTop: 16, display: 'flex', gap: 8 }}>
            <button onClick={() => setRangeValue([0, 100])}>Full Range</button>
            <button onClick={() => setRangeValue([40, 60])}>Center Range</button>
            <button onClick={() => setRangeValue([0, 50])}>Lower Half</button>
          </div>
        </div>
      </div>
    )
  }
}

// Real-World Examples
export const VolumeControl: Story = {
  render: () => {
    const [volume, setVolume] = useState(70)
    const [muted, setMuted] = useState(false)
    
    return (
      <div style={{ maxWidth: 400 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
          <button 
            onClick={() => setMuted(!muted)}
            style={{ width: 80 }}
          >
            {muted ? '🔇 Muted' : '🔊 Sound'}
          </button>
          <span style={{ minWidth: 50, textAlign: 'right' }}>{muted ? '0' : volume}%</span>
        </div>
        <Slider
          value={muted ? 0 : volume}
          onChange={(v) => {
            setVolume(v as number)
            if (muted) setMuted(false)
          }}
          min={0}
          max={100}
          step={1}
          variant="primary"
          showTooltip="hover"
          formatTooltip={(v) => `${v}%`}
          disabled={muted}
          ariaLabel="Volume control"
        />
      </div>
    )
  }
}

export const PriceRange: Story = {
  render: () => {
    const [priceRange, setPriceRange] = useState<[number, number]>([25000, 75000])
    
    return (
      <div style={{ maxWidth: 500 }}>
        <h4 style={{ marginBottom: 16 }}>
          Price Range: ${priceRange[0].toLocaleString()} - ${priceRange[1].toLocaleString()}
        </h4>
        <Slider
          value={priceRange}
          onChange={(v) => setPriceRange(v as [number, number])}
          min={0}
          max={100000}
          step={1000}
          marks={[
            { value: 0, label: '$0' },
            { value: 25000, label: '$25K' },
            { value: 50000, label: '$50K' },
            { value: 75000, label: '$75K' },
            { value: 100000, label: '$100K' }
          ]}
          variant="success"
          showTooltip="always"
          formatTooltip={(v) => `$${(v/1000).toFixed(0)}K`}
          ariaLabel={['Minimum price', 'Maximum price']}
        />
      </div>
    )
  }
}

export const ColorPicker: Story = {
  render: () => {
    const [hue, setHue] = useState(180)
    const [saturation, setSaturation] = useState(100)
    const [lightness, setLightness] = useState(50)
    
    const color = `hsl(${hue}, ${saturation}%, ${lightness}%)`
    
    return (
      <div style={{ maxWidth: 400 }}>
        <div 
          style={{ 
            width: '100%', 
            height: 100, 
            backgroundColor: color,
            borderRadius: 8,
            marginBottom: 24
          }}
        />
        
        <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
          <div>
            <label style={{ display: 'block', marginBottom: 8 }}>Hue: {hue}°</label>
            <Slider
              value={hue}
              onChange={(v) => setHue(v as number)}
              min={0}
              max={360}
              variant="primary"
              showTooltip="hover"
              formatTooltip={(v) => `${v}°`}
              ariaLabel="Hue"
            />
          </div>
          
          <div>
            <label style={{ display: 'block', marginBottom: 8 }}>Saturation: {saturation}%</label>
            <Slider
              value={saturation}
              onChange={(v) => setSaturation(v as number)}
              min={0}
              max={100}
              variant="secondary"
              showTooltip="hover"
              formatTooltip={(v) => `${v}%`}
              ariaLabel="Saturation"
            />
          </div>
          
          <div>
            <label style={{ display: 'block', marginBottom: 8 }}>Lightness: {lightness}%</label>
            <Slider
              value={lightness}
              onChange={(v) => setLightness(v as number)}
              min={0}
              max={100}
              variant="secondary"
              showTooltip="hover"
              formatTooltip={(v) => `${v}%`}
              ariaLabel="Lightness"
            />
          </div>
        </div>
        
        <div style={{ marginTop: 16, fontSize: 14, fontFamily: 'monospace' }}>
          {color}
        </div>
      </div>
    )
  }
}

export const MediaPlayer: Story = {
  render: () => {
    const [currentTime, setCurrentTime] = useState(45)
    const [volume, setVolume] = useState(70)
    const [playbackRate, setPlaybackRate] = useState(1)
    const duration = 180 // 3 minutes
    
    const formatTime = (seconds: number) => {
      const mins = Math.floor(seconds / 60)
      const secs = Math.floor(seconds % 60)
      return `${mins}:${secs.toString().padStart(2, '0')}`
    }
    
    return (
      <div style={{ maxWidth: 600, padding: 24, backgroundColor: '#f5f5f5', borderRadius: 8 }}>
        <h4 style={{ marginBottom: 24 }}>Media Player Controls</h4>
        
        {/* Progress Bar */}
        <div style={{ marginBottom: 32 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
            <span>{formatTime(currentTime)}</span>
            <span>{formatTime(duration)}</span>
          </div>
          <Slider
            value={currentTime}
            onChange={(v) => setCurrentTime(v as number)}
            min={0}
            max={duration}
            step={1}
            variant="primary"
            size="lg"
            showTooltip="hover"
            formatTooltip={formatTime}
            ariaLabel="Playback position"
          />
        </div>
        
        {/* Controls Row */}
        <div style={{ display: 'flex', gap: 32, alignItems: 'center' }}>
          {/* Volume */}
          <div style={{ flex: 1 }}>
            <label style={{ display: 'block', marginBottom: 8, fontSize: 12 }}>
              Volume: {volume}%
            </label>
            <Slider
              value={volume}
              onChange={(v) => setVolume(v as number)}
              min={0}
              max={100}
              size="sm"
              variant="secondary"
              showTooltip="hover"
              formatTooltip={(v) => `${v}%`}
              ariaLabel="Volume"
            />
          </div>
          
          {/* Playback Speed */}
          <div style={{ flex: 1 }}>
            <label style={{ display: 'block', marginBottom: 8, fontSize: 12 }}>
              Speed: {playbackRate}x
            </label>
            <Slider
              value={playbackRate}
              onChange={(v) => setPlaybackRate(v as number)}
              min={0.5}
              max={2}
              step={0.25}
              size="sm"
              variant="secondary"
              marks={[
                { value: 0.5 },
                { value: 1 },
                { value: 1.5 },
                { value: 2 }
              ]}
              showTooltip="hover"
              formatTooltip={(v) => `${v}x`}
              ariaLabel="Playback speed"
            />
          </div>
        </div>
      </div>
    )
  }
}

// Accessibility Demo
export const AccessibilityDemo: Story = {
  render: () => {
    const [value, setValue] = useState(50)
    const [rangeValue, setRangeValue] = useState<[number, number]>([20, 80])
    
    return (
      <div style={{ maxWidth: 600 }}>
        <h3 style={{ marginBottom: 24 }}>Accessibility Features Demo</h3>
        
        <div style={{ marginBottom: 32 }}>
          <h4 style={{ marginBottom: 16 }}>Keyboard Navigation</h4>
          <p style={{ marginBottom: 16, fontSize: 14, color: '#666' }}>
            Tab to focus, Arrow keys to adjust, Home/End for min/max, Page Up/Down for large steps
          </p>
          <Slider
            value={value}
            onChange={(v) => setValue(v as number)}
            showTooltip="focus"
            ariaLabel="Keyboard navigation demo"
            ariaValueText={(v) => `${v} percent`}
          />
        </div>
        
        <div style={{ marginBottom: 32 }}>
          <h4 style={{ marginBottom: 16 }}>Screen Reader Support</h4>
          <p style={{ marginBottom: 16, fontSize: 14, color: '#666' }}>
            Full ARIA labels and live value announcements
          </p>
          <Slider
            value={rangeValue}
            onChange={(v) => setRangeValue(v as [number, number])}
            ariaLabel={['Minimum value', 'Maximum value']}
            ariaValueText={(v) => `${v} out of 100`}
            ariaDescribedby="range-description"
          />
          <p id="range-description" style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
            Use this range slider to select minimum and maximum values
          </p>
        </div>
        
        <div>
          <h4 style={{ marginBottom: 16 }}>Touch/Mobile Support</h4>
          <p style={{ marginBottom: 16, fontSize: 14, color: '#666' }}>
            Optimized for touch interactions on mobile devices
          </p>
          <Slider
            defaultValue={[30, 70]}
            size="lg"
            showTooltip="always"
            variant="primary"
            ariaLabel={['Start', 'End']}
          />
        </div>
      </div>
    )
  }
}