/**
 * [FRONTEND-UPDATE] Slider Component
 * Comprehensive slider component with single/range values, orientations, and full accessibility
 */

import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { colors, semanticColors } from '../../tokens/colors'

// Types
export interface SliderProps {
  // Value management
  value?: number | [number, number]
  defaultValue?: number | [number, number]
  onChange?: (value: number | [number, number]) => void
  onChangeEnd?: (value: number | [number, number]) => void
  
  // Range configuration
  min?: number
  max?: number
  step?: number
  marks?: boolean | { value: number; label?: string }[]
  
  // Visual configuration
  orientation?: 'horizontal' | 'vertical'
  size?: 'sm' | 'md' | 'lg'
  variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error'
  showTooltip?: boolean | 'always' | 'hover' | 'focus'
  formatTooltip?: (value: number) => string
  
  // State
  disabled?: boolean
  readOnly?: boolean
  
  // Styling
  className?: string
  trackClassName?: string
  thumbClassName?: string
  markClassName?: string
  
  // Accessibility
  ariaLabel?: string | [string, string]
  ariaLabelledby?: string
  ariaDescribedby?: string
  ariaValueText?: (value: number) => string
  
  // Advanced
  inverted?: boolean
  trackClickable?: boolean
  name?: string
  id?: string
}

// Constants
const THUMB_SIZE = {
  sm: 16,
  md: 20,
  lg: 24
}

const TRACK_HEIGHT = {
  sm: 4,
  md: 6,
  lg: 8
}

const MARK_SIZE = {
  sm: 8,
  md: 10,
  lg: 12
}

// Utility functions
const clamp = (value: number, min: number, max: number) => {
  return Math.min(Math.max(value, min), max)
}

const roundToStep = (value: number, step: number, min: number) => {
  const rounded = Math.round((value - min) / step) * step + min
  return Math.round(rounded * 100) / 100 // Avoid floating point issues
}

const getPercentage = (value: number, min: number, max: number) => {
  return ((value - min) / (max - min)) * 100
}

const getValue = (percentage: number, min: number, max: number) => {
  return (percentage / 100) * (max - min) + min
}

export const Slider: React.FC<SliderProps> = ({
  value: controlledValue,
  defaultValue,
  onChange,
  onChangeEnd,
  min = 0,
  max = 100,
  step = 1,
  marks = false,
  orientation = 'horizontal',
  size = 'md',
  variant = 'primary',
  showTooltip = 'hover',
  formatTooltip = (v) => v.toString(),
  disabled = false,
  readOnly = false,
  className = '',
  trackClassName = '',
  thumbClassName = '',
  markClassName = '',
  ariaLabel,
  ariaLabelledby,
  ariaDescribedby,
  ariaValueText,
  inverted = false,
  trackClickable = true,
  name,
  id
}) => {
  // Determine if range mode
  const isRange = Array.isArray(controlledValue || defaultValue)
  
  // State management
  const [internalValue, setInternalValue] = useState<number | [number, number]>(() => {
    if (controlledValue !== undefined) return controlledValue
    if (defaultValue !== undefined) return defaultValue
    return isRange ? [min, min + (max - min) / 3] : min
  })
  
  const [isDragging, setIsDragging] = useState<false | 0 | 1>(false)
  const [isFocused, setIsFocused] = useState<false | 0 | 1>(false)
  const [isHovered, setIsHovered] = useState(false)
  const [tooltipPosition, setTooltipPosition] = useState<{ x: number; y: number } | null>(null)
  
  // Refs
  const trackRef = useRef<HTMLDivElement>(null)
  const thumb0Ref = useRef<HTMLDivElement>(null)
  const thumb1Ref = useRef<HTMLDivElement>(null)
  const touchId = useRef<number | null>(null)
  
  // Controlled/uncontrolled value
  const value = controlledValue !== undefined ? controlledValue : internalValue
  const values = Array.isArray(value) ? value : [value]
  
  // Calculate positions
  const percentages = useMemo(() => {
    return values.map(v => getPercentage(v, min, max))
  }, [values, min, max])
  
  // Generate marks array
  const marksArray = useMemo(() => {
    if (!marks) return []
    if (marks === true) {
      // Auto-generate marks
      const count = Math.min(11, Math.floor((max - min) / step) + 1)
      const markStep = (max - min) / (count - 1)
      return Array.from({ length: count }, (_, i) => ({
        value: min + markStep * i,
        label: undefined
      }))
    }
    return marks
  }, [marks, min, max, step])
  
  // Update value
  const updateValue = useCallback((newValue: number | [number, number], isEnd = false) => {
    if (disabled || readOnly) return
    
    // Clamp and round values
    if (Array.isArray(newValue)) {
      newValue = [
        clamp(roundToStep(newValue[0], step, min), min, max),
        clamp(roundToStep(newValue[1], step, min), min, max)
      ]
      // Ensure proper ordering
      if (newValue[0] > newValue[1]) {
        newValue = [newValue[1], newValue[0]]
      }
    } else {
      newValue = clamp(roundToStep(newValue, step, min), min, max)
    }
    
    if (controlledValue === undefined) {
      setInternalValue(newValue)
    }
    
    if (isEnd && onChangeEnd) {
      onChangeEnd(newValue)
    } else if (onChange) {
      onChange(newValue)
    }
  }, [disabled, readOnly, controlledValue, onChange, onChangeEnd, step, min, max])
  
  // Get position from mouse/touch event
  const getPositionFromEvent = useCallback((e: MouseEvent | TouchEvent | React.MouseEvent | React.TouchEvent) => {
    if (!trackRef.current) return null
    
    const rect = trackRef.current.getBoundingClientRect()
    const clientX = 'touches' in e ? e.touches[0]?.clientX ?? 0 : e.clientX
    const clientY = 'touches' in e ? e.touches[0]?.clientY ?? 0 : e.clientY
    
    let percentage: number
    if (orientation === 'horizontal') {
      const x = clientX - rect.left
      percentage = (x / rect.width) * 100
      if (inverted) percentage = 100 - percentage
    } else {
      const y = clientY - rect.top
      percentage = (y / rect.height) * 100
      if (!inverted) percentage = 100 - percentage
    }
    
    return clamp(percentage, 0, 100)
  }, [orientation, inverted])
  
  // Handle track click
  const handleTrackClick = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    if (!trackClickable || disabled || readOnly || isDragging !== false) return
    
    const percentage = getPositionFromEvent(e)
    if (percentage === null) return
    
    const newValue = getValue(percentage, min, max)
    
    if (isRange) {
      const [v0, v1] = values
      const d0 = Math.abs(newValue - v0)
      const d1 = Math.abs(newValue - v1)
      if (d0 < d1) {
        updateValue([newValue, v1], true)
      } else {
        updateValue([v0, newValue], true)
      }
    } else {
      updateValue(newValue, true)
    }
  }, [trackClickable, disabled, readOnly, isDragging, getPositionFromEvent, isRange, values, min, max, updateValue])
  
  // Handle thumb drag
  const handleThumbMouseDown = useCallback((thumbIndex: 0 | 1) => (e: React.MouseEvent | React.TouchEvent) => {
    if (disabled || readOnly) return
    
    e.stopPropagation()
    e.preventDefault()
    
    setIsDragging(thumbIndex)
    
    if ('touches' in e && e.touches[0]) {
      touchId.current = e.touches[0].identifier
    }
  }, [disabled, readOnly])
  
  // Handle mouse/touch move
  useEffect(() => {
    if (isDragging === false) return
    
    const handleMove = (e: MouseEvent | TouchEvent) => {
      if ('touches' in e && touchId.current !== null) {
        const touch = Array.from(e.touches).find(t => t.identifier === touchId.current)
        if (!touch) return
      }
      
      const percentage = getPositionFromEvent(e)
      if (percentage === null) return
      
      const newValue = getValue(percentage, min, max)
      
      if (isRange) {
        const [v0, v1] = values
        if (isDragging === 0) {
          updateValue([newValue, v1])
        } else {
          updateValue([v0, newValue])
        }
      } else {
        updateValue(newValue)
      }
      
      // Update tooltip position
      if (showTooltip === 'always' || (showTooltip !== false && isHovered)) {
        const clientX = 'touches' in e ? e.touches[0]?.clientX ?? 0 : e.clientX
        const clientY = 'touches' in e ? e.touches[0]?.clientY ?? 0 : e.clientY
        setTooltipPosition({ x: clientX, y: clientY })
      }
    }
    
    const handleEnd = () => {
      if (isDragging !== false) {
        const endValue = isRange ? values : values[0]
        updateValue(endValue, true)
        setIsDragging(false)
        touchId.current = null
        setTooltipPosition(null)
      }
    }
    
    document.addEventListener('mousemove', handleMove)
    document.addEventListener('mouseup', handleEnd)
    document.addEventListener('touchmove', handleMove, { passive: false })
    document.addEventListener('touchend', handleEnd)
    document.addEventListener('touchcancel', handleEnd)
    
    return () => {
      document.removeEventListener('mousemove', handleMove)
      document.removeEventListener('mouseup', handleEnd)
      document.removeEventListener('touchmove', handleMove)
      document.removeEventListener('touchend', handleEnd)
      document.removeEventListener('touchcancel', handleEnd)
    }
  }, [isDragging, isRange, values, getPositionFromEvent, min, max, updateValue, showTooltip, isHovered])
  
  // Handle keyboard navigation
  const handleKeyDown = useCallback((thumbIndex: 0 | 1) => (e: React.KeyboardEvent) => {
    if (disabled || readOnly) return
    
    const stepSize = e.shiftKey ? step * 10 : e.ctrlKey ? step * 0.1 : step
    let handled = true
    
    if (isRange) {
      const [v0, v1] = values
      let newValue: [number, number] = [v0, v1]
      
      switch (e.key) {
        case 'ArrowLeft':
        case 'ArrowDown':
          if (thumbIndex === 0) {
            newValue = [v0 - stepSize, v1]
          } else {
            newValue = [v0, v1 - stepSize]
          }
          break
        case 'ArrowRight':
        case 'ArrowUp':
          if (thumbIndex === 0) {
            newValue = [v0 + stepSize, v1]
          } else {
            newValue = [v0, v1 + stepSize]
          }
          break
        case 'Home':
          if (thumbIndex === 0) {
            newValue = [min, v1]
          } else {
            newValue = [v0, min]
          }
          break
        case 'End':
          if (thumbIndex === 0) {
            newValue = [max, v1]
          } else {
            newValue = [v0, max]
          }
          break
        case 'PageDown':
          const pageDownStep = (max - min) / 10
          if (thumbIndex === 0) {
            newValue = [v0 - pageDownStep, v1]
          } else {
            newValue = [v0, v1 - pageDownStep]
          }
          break
        case 'PageUp':
          const pageUpStep = (max - min) / 10
          if (thumbIndex === 0) {
            newValue = [v0 + pageUpStep, v1]
          } else {
            newValue = [v0, v1 + pageUpStep]
          }
          break
        default:
          handled = false
      }
      
      if (handled) {
        e.preventDefault()
        updateValue(newValue)
      }
    } else {
      const v = values[0]
      let newValue = v
      
      switch (e.key) {
        case 'ArrowLeft':
        case 'ArrowDown':
          newValue = v - stepSize
          break
        case 'ArrowRight':
        case 'ArrowUp':
          newValue = v + stepSize
          break
        case 'Home':
          newValue = min
          break
        case 'End':
          newValue = max
          break
        case 'PageDown':
          newValue = v - (max - min) / 10
          break
        case 'PageUp':
          newValue = v + (max - min) / 10
          break
        default:
          handled = false
      }
      
      if (handled) {
        e.preventDefault()
        updateValue(newValue)
      }
    }
  }, [disabled, readOnly, isRange, values, step, min, max, updateValue])
  
  // Get color based on variant
  const getVariantColor = () => {
    switch (variant) {
      case 'secondary':
        return semanticColors.light.interactive.secondary.default
      case 'success':
        return colors.success[500]
      case 'warning':
        return colors.warning[500]
      case 'error':
        return colors.error[500]
      default:
        return semanticColors.light.interactive.primary.default
    }
  }
  
  const variantColor = getVariantColor()
  
  // Styles
  const containerStyle: React.CSSProperties = {
    position: 'relative',
    display: orientation === 'horizontal' ? 'block' : 'inline-block',
    width: orientation === 'horizontal' ? '100%' : THUMB_SIZE[size],
    height: orientation === 'vertical' ? 300 : THUMB_SIZE[size],
    padding: orientation === 'horizontal' ? `${THUMB_SIZE[size] / 2}px 0` : `0 ${THUMB_SIZE[size] / 2}px`,
    opacity: disabled ? 0.5 : 1,
    cursor: disabled ? 'not-allowed' : 'default'
  }
  
  const trackStyle: React.CSSProperties = {
    position: 'absolute',
    backgroundColor: semanticColors.light.border.primary,
    borderRadius: TRACK_HEIGHT[size] / 2,
    cursor: trackClickable && !disabled && !readOnly ? 'pointer' : 'default',
    ...(orientation === 'horizontal' ? {
      left: 0,
      right: 0,
      top: '50%',
      transform: 'translateY(-50%)',
      height: TRACK_HEIGHT[size]
    } : {
      top: 0,
      bottom: 0,
      left: '50%',
      transform: 'translateX(-50%)',
      width: TRACK_HEIGHT[size]
    })
  }
  
  const fillStyle: React.CSSProperties = {
    position: 'absolute',
    backgroundColor: variantColor,
    borderRadius: TRACK_HEIGHT[size] / 2,
    transition: isDragging ? 'none' : 'all 0.2s ease',
    ...(orientation === 'horizontal' ? {
      top: 0,
      bottom: 0,
      left: isRange ? `${percentages[0]}%` : 0,
      width: isRange ? `${percentages[1] - percentages[0]}%` : `${percentages[0]}%`
    } : {
      left: 0,
      right: 0,
      bottom: isRange ? `${percentages[0]}%` : 0,
      height: isRange ? `${percentages[1] - percentages[0]}%` : `${percentages[0]}%`
    })
  }
  
  const getThumbStyle = (index: 0 | 1): React.CSSProperties => ({
    position: 'absolute',
    width: THUMB_SIZE[size],
    height: THUMB_SIZE[size],
    backgroundColor: '#fff',
    border: `2px solid ${variantColor}`,
    borderRadius: '50%',
    cursor: disabled || readOnly ? 'default' : 'grab',
    transition: isDragging === index ? 'none' : 'all 0.2s ease',
    boxShadow: isFocused === index ? `0 0 0 4px ${variantColor}33` : '0 2px 4px rgba(0,0,0,0.1)',
    transform: orientation === 'horizontal' 
      ? `translate(-50%, -50%)` 
      : `translate(-50%, 50%)`,
    ...(orientation === 'horizontal' ? {
      left: `${percentages[index]}%`,
      top: '50%'
    } : {
      bottom: `${percentages[index]}%`,
      left: '50%'
    }),
    '&:hover': {
      boxShadow: '0 2px 8px rgba(0,0,0,0.2)'
    },
    '&:active': {
      cursor: 'grabbing'
    }
  })
  
  const getMarkStyle = (markValue: number): React.CSSProperties => {
    const percentage = getPercentage(markValue, min, max)
    return {
      position: 'absolute',
      width: MARK_SIZE[size],
      height: MARK_SIZE[size],
      backgroundColor: semanticColors.light.border.secondary,
      borderRadius: '50%',
      transform: orientation === 'horizontal' 
        ? 'translate(-50%, -50%)' 
        : 'translate(-50%, 50%)',
      ...(orientation === 'horizontal' ? {
        left: `${percentage}%`,
        top: '50%'
      } : {
        bottom: `${percentage}%`,
        left: '50%'
      })
    }
  }
  
  // Tooltip component
  const Tooltip: React.FC<{ value: number; thumbRef: React.RefObject<HTMLDivElement> }> = ({ value, thumbRef }) => {
    const shouldShow = showTooltip === 'always' || 
      (showTooltip === 'hover' && isHovered) ||
      (showTooltip === 'focus' && isFocused !== false) ||
      isDragging !== false
    
    if (!shouldShow || !thumbRef.current) return null
    
    const rect = thumbRef.current.getBoundingClientRect()
    
    return (
      <div
        style={{
          position: 'fixed',
          left: rect.left + rect.width / 2,
          top: rect.top - 30,
          transform: 'translateX(-50%)',
          backgroundColor: semanticColors.light.background.inverse,
          color: semanticColors.light.text.inverse,
          padding: '4px 8px',
          borderRadius: 4,
          fontSize: 12,
          whiteSpace: 'nowrap',
          pointerEvents: 'none',
          zIndex: 1000
        }}
      >
        {formatTooltip(value)}
      </div>
    )
  }
  
  return (
    <div 
      className={className}
      style={containerStyle}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Track */}
      <div
        ref={trackRef}
        style={trackStyle}
        className={trackClassName}
        onClick={handleTrackClick}
        onTouchEnd={handleTrackClick}
      >
        {/* Fill */}
        <div style={fillStyle} />
        
        {/* Marks */}
        {marksArray.map((mark, i) => (
          <div
            key={i}
            style={getMarkStyle(mark.value)}
            className={markClassName}
          />
        ))}
      </div>
      
      {/* Thumbs */}
      <div
        ref={thumb0Ref}
        role="slider"
        tabIndex={disabled ? -1 : 0}
        aria-label={Array.isArray(ariaLabel) ? ariaLabel[0] : ariaLabel}
        aria-labelledby={ariaLabelledby}
        aria-describedby={ariaDescribedby}
        aria-valuemin={min}
        aria-valuemax={max}
        aria-valuenow={values[0]}
        aria-valuetext={ariaValueText ? ariaValueText(values[0]) : values[0].toString()}
        aria-disabled={disabled}
        aria-readonly={readOnly}
        aria-orientation={orientation}
        style={getThumbStyle(0)}
        className={thumbClassName}
        onMouseDown={handleThumbMouseDown(0)}
        onTouchStart={handleThumbMouseDown(0)}
        onKeyDown={handleKeyDown(0)}
        onFocus={() => setIsFocused(0)}
        onBlur={() => setIsFocused(false)}
      />
      
      {isRange && (
        <div
          ref={thumb1Ref}
          role="slider"
          tabIndex={disabled ? -1 : 0}
          aria-label={Array.isArray(ariaLabel) ? ariaLabel[1] : `${ariaLabel} end`}
          aria-labelledby={ariaLabelledby}
          aria-describedby={ariaDescribedby}
          aria-valuemin={min}
          aria-valuemax={max}
          aria-valuenow={values[1]}
          aria-valuetext={ariaValueText ? ariaValueText(values[1]) : values[1].toString()}
          aria-disabled={disabled}
          aria-readonly={readOnly}
          aria-orientation={orientation}
          style={getThumbStyle(1)}
          className={thumbClassName}
          onMouseDown={handleThumbMouseDown(1)}
          onTouchStart={handleThumbMouseDown(1)}
          onKeyDown={handleKeyDown(1)}
          onFocus={() => setIsFocused(1)}
          onBlur={() => setIsFocused(false)}
        />
      )}
      
      {/* Tooltips */}
      <Tooltip value={values[0]} thumbRef={thumb0Ref} />
      {isRange && <Tooltip value={values[1]} thumbRef={thumb1Ref} />}
      
      {/* Hidden inputs for form submission */}
      {name && (
        <>
          <input
            type="hidden"
            name={isRange ? `${name}[0]` : name}
            value={values[0]}
            id={id}
          />
          {isRange && (
            <input
              type="hidden"
              name={`${name}[1]`}
              value={values[1]}
            />
          )}
        </>
      )}
    </div>
  )
}

export default Slider