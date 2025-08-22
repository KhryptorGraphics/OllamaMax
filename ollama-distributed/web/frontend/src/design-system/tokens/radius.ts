/**
 * Design Tokens: Border Radius
 * Comprehensive border radius system with semantic naming and component-specific values
 */

// Base border radius values
export const radius = {
  none: '0',
  xs: '0.125rem',    // 2px
  sm: '0.25rem',     // 4px
  md: '0.375rem',    // 6px
  lg: '0.5rem',      // 8px
  xl: '0.75rem',     // 12px
  '2xl': '1rem',     // 16px
  '3xl': '1.5rem',   // 24px
  full: '9999px',    // Fully rounded
  
  // Special values
  inherit: 'inherit',
  initial: 'initial'
} as const

// Semantic radius tokens
export const semanticRadius = {
  // Component-specific border radius
  component: {
    // Button radius
    button: {
      xs: radius.xs,
      sm: radius.sm,
      md: radius.md,
      lg: radius.lg,
      xl: radius.xl,
      pill: radius.full
    },
    
    // Input field radius
    input: {
      xs: radius.xs,
      sm: radius.sm,
      md: radius.md,
      lg: radius.lg,
      xl: radius.xl
    },
    
    // Card radius
    card: {
      none: radius.none,
      sm: radius.sm,
      md: radius.lg,
      lg: radius.xl,
      xl: radius['2xl']
    },
    
    // Badge radius
    badge: {
      sm: radius.sm,
      md: radius.md,
      lg: radius.lg,
      pill: radius.full
    },
    
    // Avatar radius
    avatar: {
      none: radius.none,
      sm: radius.sm,
      md: radius.md,
      lg: radius.lg,
      xl: radius.xl,
      full: radius.full
    },
    
    // Modal/Dialog radius
    modal: {
      sm: radius.lg,
      md: radius.xl,
      lg: radius['2xl']
    },
    
    // Tooltip radius
    tooltip: {
      default: radius.md
    },
    
    // Dropdown radius
    dropdown: {
      default: radius.lg
    }
  },
  
  // Layout component radius
  layout: {
    // Container radius
    container: {
      none: radius.none,
      sm: radius.md,
      md: radius.lg,
      lg: radius.xl
    },
    
    // Section radius
    section: {
      none: radius.none,
      sm: radius.lg,
      md: radius.xl,
      lg: radius['2xl']
    }
  },
  
  // Interactive state radius
  interactive: {
    // Focus ring radius (usually matches element + offset)
    focus: {
      button: radius.lg,
      input: radius.lg,
      card: radius.xl
    },
    
    // Hover state radius adjustments
    hover: {
      subtle: radius.md,
      pronounced: radius.lg
    }
  }
} as const

// Responsive radius values
export const responsiveRadius = {
  // Mobile-first radius scaling
  breakpoints: {
    mobile: {
      button: radius.sm,
      card: radius.md,
      modal: radius.lg,
      input: radius.sm
    },
    tablet: {
      button: radius.md,
      card: radius.lg,
      modal: radius.xl,
      input: radius.md
    },
    desktop: {
      button: radius.md,
      card: radius.xl,
      modal: radius['2xl'],
      input: radius.md
    }
  },
  
  // Fluid radius using clamp()
  fluid: {
    button: 'clamp(0.25rem, 1vw, 0.5rem)',
    card: 'clamp(0.5rem, 2vw, 1rem)',
    modal: 'clamp(0.75rem, 3vw, 1.5rem)'
  }
} as const

// Special radius patterns
export const specialRadius = {
  // Asymmetric radius for specific use cases
  asymmetric: {
    // Top rounded for modals from bottom
    topRounded: {
      sm: `${radius.lg} ${radius.lg} 0 0`,
      md: `${radius.xl} ${radius.xl} 0 0`,
      lg: `${radius['2xl']} ${radius['2xl']} 0 0`
    },
    
    // Bottom rounded for modals from top
    bottomRounded: {
      sm: `0 0 ${radius.lg} ${radius.lg}`,
      md: `0 0 ${radius.xl} ${radius.xl}`,
      lg: `0 0 ${radius['2xl']} ${radius['2xl']}`
    },
    
    // Left rounded for dropdowns from right
    leftRounded: {
      sm: `${radius.lg} 0 0 ${radius.lg}`,
      md: `${radius.xl} 0 0 ${radius.xl}`,
      lg: `${radius['2xl']} 0 0 ${radius['2xl']}`
    },
    
    // Right rounded for dropdowns from left
    rightRounded: {
      sm: `0 ${radius.lg} ${radius.lg} 0`,
      md: `0 ${radius.xl} ${radius.xl} 0`,
      lg: `0 ${radius['2xl']} ${radius['2xl']} 0`
    }
  },
  
  // Percentage-based radius for specific ratios
  percentage: {
    '10': '10%',
    '20': '20%',
    '25': '25%',
    '30': '30%',
    '50': '50%'
  },
  
  // Organic/irregular radius for natural feel
  organic: {
    subtle: '0.3rem 0.4rem 0.35rem 0.45rem',
    moderate: '0.6rem 0.8rem 0.7rem 0.9rem',
    pronounced: '1rem 1.3rem 1.1rem 1.4rem'
  }
} as const

// Border radius utility functions
export const radiusUtils = {
  // Get component radius
  getComponentRadius: (
    component: keyof typeof semanticRadius.component,
    variant: string = 'md'
  ) => {
    const componentRadius = semanticRadius.component[component] as any
    return componentRadius[variant] || componentRadius.md || radius.md
  },
  
  // Get responsive radius
  getResponsiveRadius: (
    component: keyof typeof responsiveRadius.breakpoints.mobile,
    breakpoint: keyof typeof responsiveRadius.breakpoints = 'mobile'
  ) => {
    return responsiveRadius.breakpoints[breakpoint][component]
  },
  
  // Get fluid radius
  getFluidRadius: (component: keyof typeof responsiveRadius.fluid) => {
    return responsiveRadius.fluid[component]
  },
  
  // Create custom radius values
  createCustomRadius: (
    topLeft: string,
    topRight?: string,
    bottomRight?: string,
    bottomLeft?: string
  ) => {
    const values = [topLeft, topRight || topLeft, bottomRight || topLeft, bottomLeft || topRight || topLeft]
    return values.join(' ')
  },
  
  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}
    
    // Base radius values
    Object.entries(radius).forEach(([key, value]) => {
      cssVars[`--radius-${key}`] = value
    })
    
    // Component radius values
    Object.entries(semanticRadius.component).forEach(([component, variants]) => {
      Object.entries(variants).forEach(([variant, value]) => {
        cssVars[`--radius-${component}-${variant}`] = value
      })
    })
    
    // Layout radius values
    Object.entries(semanticRadius.layout).forEach(([component, variants]) => {
      Object.entries(variants).forEach(([variant, value]) => {
        cssVars[`--radius-layout-${component}-${variant}`] = value
      })
    })
    
    // Interactive radius values
    Object.entries(semanticRadius.interactive).forEach(([state, components]) => {
      Object.entries(components).forEach(([component, value]) => {
        cssVars[`--radius-${state}-${component}`] = value
      })
    })
    
    // Responsive radius values
    Object.entries(responsiveRadius.breakpoints).forEach(([breakpoint, components]) => {
      Object.entries(components).forEach(([component, value]) => {
        cssVars[`--radius-${breakpoint}-${component}`] = value
      })
    })
    
    // Fluid radius values
    Object.entries(responsiveRadius.fluid).forEach(([component, value]) => {
      cssVars[`--radius-fluid-${component}`] = value
    })
    
    // Special radius patterns
    Object.entries(specialRadius.asymmetric).forEach(([pattern, variants]) => {
      Object.entries(variants).forEach(([variant, value]) => {
        cssVars[`--radius-${pattern}-${variant}`] = value
      })
    })
    
    return cssVars
  },
  
  // Apply radius to element
  applyRadius: (element: HTMLElement, radiusValue: string) => {
    if (!element) return
    element.style.borderRadius = radiusValue
  },
  
  // Check if radius value is valid
  isValidRadius: (value: string): boolean => {
    return /^(\d+(\.\d+)?(px|rem|em|%)|none|inherit|initial|\d+)$/.test(value.trim())
  }
} as const

// Accessibility considerations for border radius
export const accessibilityRadius = {
  // Ensure adequate radius for touch targets
  touchTarget: {
    minimum: radius.sm,     // Minimum radius for touch targets
    recommended: radius.md  // Recommended radius for better usability
  },
  
  // Focus ring radius considerations
  focusRing: {
    // Focus ring should slightly exceed element radius
    getOffset: (elementRadius: string) => {
      const numericValue = parseFloat(elementRadius)
      if (isNaN(numericValue)) return radius.md
      return `${numericValue + 0.125}rem` // Add 2px
    }
  },
  
  // High contrast mode adjustments
  highContrast: {
    // Slightly larger radius for better definition
    enhanced: {
      button: radius.lg,
      input: radius.lg,
      card: radius.xl
    }
  }
} as const

// Export types
export type RadiusKey = keyof typeof radius
export type ComponentRadiusKey = keyof typeof semanticRadius.component
export type RadiusValue = typeof radius[keyof typeof radius]
export type SemanticRadiusCategory = keyof typeof semanticRadius

export default {
  radius,
  semanticRadius,
  responsiveRadius,
  specialRadius,
  radiusUtils,
  accessibilityRadius
}