/**
 * Design Tokens: Shadows
 * Comprehensive shadow system with elevation scales, focus rings, and design variants
 */

// Base shadow definitions
export const shadows = {
  // No shadow
  none: 'none',
  
  // Subtle shadows for cards and elevated content
  xs: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  sm: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
  
  // Inner shadows for pressed states
  inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
  innerLg: 'inset 0 4px 8px 0 rgba(0, 0, 0, 0.1)',
  
  // Colored shadows for brand elements
  primary: '0 4px 14px 0 rgba(59, 130, 246, 0.15)',
  primaryLg: '0 10px 25px 0 rgba(59, 130, 246, 0.2)',
  
  // Glow effects
  glow: '0 0 20px rgba(59, 130, 246, 0.4)',
  glowLg: '0 0 40px rgba(59, 130, 246, 0.3)'
} as const

// Semantic shadow scales
export const semanticShadows = {
  // Elevation levels (Material Design inspired)
  elevation: {
    0: shadows.none,
    1: shadows.xs,
    2: shadows.sm,
    3: shadows.md,
    4: shadows.lg,
    5: shadows.xl,
    6: shadows['2xl']
  },
  
  // Interactive states
  interactive: {
    none: shadows.none,
    hover: shadows.md,
    active: shadows.inner,
    focus: '0 0 0 2px rgba(59, 130, 246, 0.5)',
    focusWithin: '0 0 0 1px rgba(59, 130, 246, 0.3)',
    disabled: shadows.none
  },
  
  // Component specific shadows
  component: {
    // Button shadows
    button: {
      default: shadows.sm,
      hover: shadows.md,
      active: shadows.inner,
      focus: '0 0 0 2px rgba(59, 130, 246, 0.5)',
      primary: shadows.primary,
      primaryHover: shadows.primaryLg
    },
    
    // Card shadows
    card: {
      default: shadows.sm,
      hover: shadows.md,
      elevated: shadows.lg,
      floating: shadows.xl
    },
    
    // Modal/Dialog shadows
    modal: {
      backdrop: 'rgba(0, 0, 0, 0.5)',
      content: shadows['2xl']
    },
    
    // Dropdown shadows
    dropdown: {
      default: shadows.lg,
      floating: shadows.xl
    },
    
    // Tooltip shadows
    tooltip: {
      default: shadows.md
    }
  }
} as const

// Dark mode shadow variants
export const darkModeShadows = {
  elevation: {
    0: 'none',
    1: '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
    2: '0 1px 3px 0 rgba(0, 0, 0, 0.4), 0 1px 2px 0 rgba(0, 0, 0, 0.2)',
    3: '0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
    4: '0 10px 15px -3px rgba(0, 0, 0, 0.4), 0 4px 6px -2px rgba(0, 0, 0, 0.2)',
    5: '0 20px 25px -5px rgba(0, 0, 0, 0.4), 0 10px 10px -5px rgba(0, 0, 0, 0.1)',
    6: '0 25px 50px -12px rgba(0, 0, 0, 0.6)'
  },
  
  glow: {
    primary: '0 0 20px rgba(96, 165, 250, 0.6)',
    success: '0 0 20px rgba(34, 197, 94, 0.6)',
    warning: '0 0 20px rgba(245, 158, 11, 0.6)',
    error: '0 0 20px rgba(239, 68, 68, 0.6)'
  }
} as const

// Shadow utility functions
export const shadowUtils = {
  // Get shadow based on elevation level
  getElevation: (level: keyof typeof semanticShadows.elevation, isDark = false) => {
    return isDark ? darkModeShadows.elevation[level] : semanticShadows.elevation[level]
  },
  
  // Get component shadow
  getComponentShadow: (
    component: keyof typeof semanticShadows.component,
    variant: string,
    isDark = false
  ) => {
    const componentShadows = semanticShadows.component[component] as any
    const shadow = componentShadows[variant] || componentShadows.default
    
    // For dark mode, apply darker variant if available
    if (isDark && component === 'card') {
      return darkModeShadows.elevation[2] // Default to elevation 2 for cards in dark mode
    }
    
    return shadow
  },
  
  // Get interactive shadow based on state
  getInteractiveShadow: (state: keyof typeof semanticShadows.interactive) => {
    return semanticShadows.interactive[state]
  },
  
  // Create custom shadow with specified color
  createColoredShadow: (
    size: keyof typeof shadows,
    color: string,
    opacity = 0.15
  ) => {
    const baseShadow = shadows[size]
    if (baseShadow === 'none') return 'none'
    
    // Extract shadow properties and replace with custom color
    return baseShadow.replace(/rgba\([^)]+\)/g, `${color.replace('#', '')}${Math.round(opacity * 255).toString(16).padStart(2, '0')}`)
  },
  
  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}
    
    // Base shadows
    Object.entries(shadows).forEach(([key, value]) => {
      cssVars[`--shadow-${key}`] = value
    })
    
    // Elevation shadows
    Object.entries(semanticShadows.elevation).forEach(([key, value]) => {
      cssVars[`--shadow-elevation-${key}`] = value
    })
    
    // Interactive shadows
    Object.entries(semanticShadows.interactive).forEach(([key, value]) => {
      cssVars[`--shadow-interactive-${key}`] = value
    })
    
    // Component shadows
    Object.entries(semanticShadows.component).forEach(([component, variants]) => {
      Object.entries(variants).forEach(([variant, value]) => {
        cssVars[`--shadow-${component}-${variant}`] = value
      })
    })
    
    // Dark mode elevation shadows
    Object.entries(darkModeShadows.elevation).forEach(([key, value]) => {
      cssVars[`--shadow-dark-elevation-${key}`] = value
    })
    
    // Dark mode glow shadows
    Object.entries(darkModeShadows.glow).forEach(([key, value]) => {
      cssVars[`--shadow-dark-glow-${key}`] = value
    })
    
    return cssVars
  },
  
  // Apply shadow to element
  applyShadow: (element: HTMLElement, shadow: string) => {
    if (!element) return
    element.style.boxShadow = shadow
  }
} as const

// Accessibility considerations for shadows
export const accessibilityShadows = {
  // High contrast mode adjustments
  highContrast: {
    // Replace subtle shadows with borders for better visibility
    elevation: {
      0: 'none',
      1: '0 0 0 1px rgba(0, 0, 0, 0.1)',
      2: '0 0 0 1px rgba(0, 0, 0, 0.2)',
      3: '0 0 0 2px rgba(0, 0, 0, 0.2)',
      4: '0 0 0 2px rgba(0, 0, 0, 0.3)',
      5: '0 0 0 3px rgba(0, 0, 0, 0.3)',
      6: '0 0 0 3px rgba(0, 0, 0, 0.4)'
    }
  },
  
  // Reduced motion preferences
  reducedMotion: {
    // Simpler shadows that don't imply motion
    static: shadows.sm,
    focus: '0 0 0 2px rgba(59, 130, 246, 0.8)', // Higher opacity for clarity
    none: 'none'
  }
} as const

// Export types
export type ShadowKey = keyof typeof shadows
export type ElevationLevel = keyof typeof semanticShadows.elevation
export type InteractiveState = keyof typeof semanticShadows.interactive
export type ComponentShadow = keyof typeof semanticShadows.component
export type ShadowValue = typeof shadows[keyof typeof shadows]

export default {
  shadows,
  semanticShadows,
  darkModeShadows,
  shadowUtils,
  accessibilityShadows
}