/**
 * Design Tokens: Spacing
 * Comprehensive spacing system with logical units, responsive patterns, and layout utilities
 */

// Base spacing unit (rem values)
export const spacing = {
  0: '0',
  px: '1px',
  0.5: '0.125rem',  // 2px
  1: '0.25rem',     // 4px
  1.5: '0.375rem',  // 6px
  2: '0.5rem',      // 8px
  2.5: '0.625rem',  // 10px
  3: '0.75rem',     // 12px
  3.5: '0.875rem',  // 14px
  4: '1rem',        // 16px
  5: '1.25rem',     // 20px
  6: '1.5rem',      // 24px
  7: '1.75rem',     // 28px
  8: '2rem',        // 32px
  9: '2.25rem',     // 36px
  10: '2.5rem',     // 40px
  11: '2.75rem',    // 44px
  12: '3rem',       // 48px
  14: '3.5rem',     // 56px
  16: '4rem',       // 64px
  20: '5rem',       // 80px
  24: '6rem',       // 96px
  28: '7rem',       // 112px
  32: '8rem',       // 128px
  36: '9rem',       // 144px
  40: '10rem',      // 160px
  44: '11rem',      // 176px
  48: '12rem',      // 192px
  52: '13rem',      // 208px
  56: '14rem',      // 224px
  60: '15rem',      // 240px
  64: '16rem',      // 256px
  72: '18rem',      // 288px
  80: '20rem',      // 320px
  96: '24rem'       // 384px
} as const

// Semantic spacing tokens
export const semanticSpacing = {
  // Component spacing
  component: {
    // Internal padding
    padding: {
      xs: spacing[1],    // 4px
      sm: spacing[2],    // 8px
      md: spacing[4],    // 16px
      lg: spacing[6],    // 24px
      xl: spacing[8],    // 32px
      '2xl': spacing[12] // 48px
    },

    // Margins between components
    margin: {
      xs: spacing[2],    // 8px
      sm: spacing[4],    // 16px
      md: spacing[6],    // 24px
      lg: spacing[8],    // 32px
      xl: spacing[12],   // 48px
      '2xl': spacing[16] // 64px
    },

    // Gaps in flex/grid layouts
    gap: {
      xs: spacing[1],    // 4px
      sm: spacing[2],    // 8px
      md: spacing[4],    // 16px
      lg: spacing[6],    // 24px
      xl: spacing[8],    // 32px
      '2xl': spacing[12] // 48px
    }
  },

  // Layout spacing
  layout: {
    // Container padding
    container: {
      xs: spacing[4],    // 16px
      sm: spacing[6],    // 24px
      md: spacing[8],    // 32px
      lg: spacing[12],   // 48px
      xl: spacing[16],   // 64px
      '2xl': spacing[24] // 96px
    },

    // Section spacing
    section: {
      xs: spacing[8],    // 32px
      sm: spacing[12],   // 48px
      md: spacing[16],   // 64px
      lg: spacing[24],   // 96px
      xl: spacing[32],   // 128px
      '2xl': spacing[48] // 192px
    },

    // Grid gutters
    gutter: {
      xs: spacing[2],    // 8px
      sm: spacing[4],    // 16px
      md: spacing[6],    // 24px
      lg: spacing[8],    // 32px
      xl: spacing[12],   // 48px
      '2xl': spacing[16] // 64px
    }
  },

  // Interactive element spacing
  interactive: {
    // Button padding
    button: {
      xs: `${spacing[2]} ${spacing[3]}`,      // 8px 12px
      sm: `${spacing[2]} ${spacing[4]}`,      // 8px 16px
      md: `${spacing[3]} ${spacing[6]}`,      // 12px 24px
      lg: `${spacing[4]} ${spacing[8]}`,      // 16px 32px
      xl: `${spacing[5]} ${spacing[10]}`      // 20px 40px
    },

    // Input padding
    input: {
      xs: `${spacing[2]} ${spacing[3]}`,      // 8px 12px
      sm: `${spacing[2.5]} ${spacing[3]}`,    // 10px 12px
      md: `${spacing[3]} ${spacing[4]}`,      // 12px 16px
      lg: `${spacing[4]} ${spacing[5]}`,      // 16px 20px
      xl: `${spacing[5]} ${spacing[6]}`       // 20px 24px
    },

    // Focus ring offset
    focusOffset: {
      default: spacing[1],   // 4px
      large: spacing[2]      // 8px
    }
  }
} as const

// Responsive spacing
export const responsiveSpacing = {
  // Mobile-first breakpoint spacing
  breakpoints: {
    mobile: {
      container: spacing[4],    // 16px
      section: spacing[8],      // 32px
      component: spacing[4],    // 16px
      gutter: spacing[4]        // 16px
    },
    tablet: {
      container: spacing[6],    // 24px
      section: spacing[12],     // 48px
      component: spacing[6],    // 24px
      gutter: spacing[6]        // 24px
    },
    desktop: {
      container: spacing[8],    // 32px
      section: spacing[16],     // 64px
      component: spacing[8],    // 32px
      gutter: spacing[8]        // 32px
    },
    wide: {
      container: spacing[12],   // 48px
      section: spacing[24],     // 96px
      component: spacing[12],   // 48px
      gutter: spacing[12]       // 48px
    }
  },

  // Fluid spacing using clamp()
  fluid: {
    container: 'clamp(1rem, 4vw, 3rem)',          // 16px to 48px
    section: 'clamp(2rem, 8vw, 6rem)',            // 32px to 96px
    component: 'clamp(1rem, 2vw, 2rem)',          // 16px to 32px
    gap: 'clamp(0.5rem, 2vw, 1.5rem)'             // 8px to 24px
  }
} as const

// Logical spacing properties (for international layouts)
export const logicalSpacing = {
  // Inline (horizontal) spacing
  inline: {
    start: 'margin-inline-start',
    end: 'margin-inline-end',
    both: 'margin-inline'
  },

  // Block (vertical) spacing
  block: {
    start: 'margin-block-start',
    end: 'margin-block-end',
    both: 'margin-block'
  },

  // Padding equivalents
  padding: {
    inline: {
      start: 'padding-inline-start',
      end: 'padding-inline-end',
      both: 'padding-inline'
    },
    block: {
      start: 'padding-block-start',
      end: 'padding-block-end',
      both: 'padding-block'
    }
  }
} as const

// Special spacing values
export const specialSpacing = {
  // Auto values
  auto: 'auto',
  
  // Negative spacing (for overlapping elements)
  negative: Object.fromEntries(
    Object.entries(spacing).map(([key, value]) => [
      key, value === '0' ? '0' : `-${value}`
    ])
  ),

  // Percentage-based spacing
  percentage: {
    '1/12': '8.333333%',
    '1/6': '16.666667%',
    '1/4': '25%',
    '1/3': '33.333333%',
    '1/2': '50%',
    '2/3': '66.666667%',
    '3/4': '75%',
    '5/6': '83.333333%',
    full: '100%'
  },

  // Viewport-relative spacing
  viewport: {
    '1vw': '1vw',
    '2vw': '2vw',
    '5vw': '5vw',
    '10vw': '10vw',
    '1vh': '1vh',
    '2vh': '2vh',
    '5vh': '5vh',
    '10vh': '10vh'
  }
} as const

// Spacing utilities
export const spacingUtils = {
  // Get semantic spacing value
  get: (
    category: keyof typeof semanticSpacing,
    subcategory: string,
    size: string
  ) => {
    const categoryObj = semanticSpacing[category] as any
    const subcategoryObj = categoryObj?.[subcategory] as any
    return subcategoryObj?.[size] || spacing[4] // fallback to medium
  },

  // Get responsive spacing for breakpoint
  getResponsive: (
    property: keyof typeof responsiveSpacing.breakpoints.mobile,
    breakpoint: keyof typeof responsiveSpacing.breakpoints = 'mobile'
  ) => {
    return responsiveSpacing.breakpoints[breakpoint][property]
  },

  // Get fluid spacing value
  getFluid: (property: keyof typeof responsiveSpacing.fluid) => {
    return responsiveSpacing.fluid[property]
  },

  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}

    // Base spacing
    Object.entries(spacing).forEach(([key, value]) => {
      cssVars[`--spacing-${key}`] = value
    })

    // Semantic spacing
    Object.entries(semanticSpacing).forEach(([category, subcategories]) => {
      Object.entries(subcategories).forEach(([subcategory, sizes]) => {
        if (typeof sizes === 'object') {
          Object.entries(sizes).forEach(([size, value]) => {
            cssVars[`--spacing-${category}-${subcategory}-${size}`] = value
          })
        }
      })
    })

    // Responsive spacing
    Object.entries(responsiveSpacing.breakpoints).forEach(([breakpoint, values]) => {
      Object.entries(values).forEach(([property, value]) => {
        cssVars[`--spacing-${breakpoint}-${property}`] = value
      })
    })

    // Fluid spacing
    Object.entries(responsiveSpacing.fluid).forEach(([property, value]) => {
      cssVars[`--spacing-fluid-${property}`] = value
    })

    return cssVars
  },

  // Create spacing classes
  createClasses: () => {
    const classes: Record<string, any> = {}

    // Margin classes
    Object.entries(spacing).forEach(([key, value]) => {
      classes[`.m-${key}`] = { margin: value }
      classes[`.mt-${key}`] = { marginTop: value }
      classes[`.mr-${key}`] = { marginRight: value }
      classes[`.mb-${key}`] = { marginBottom: value }
      classes[`.ml-${key}`] = { marginLeft: value }
      classes[`.mx-${key}`] = { marginLeft: value, marginRight: value }
      classes[`.my-${key}`] = { marginTop: value, marginBottom: value }
    })

    // Padding classes
    Object.entries(spacing).forEach(([key, value]) => {
      classes[`.p-${key}`] = { padding: value }
      classes[`.pt-${key}`] = { paddingTop: value }
      classes[`.pr-${key}`] = { paddingRight: value }
      classes[`.pb-${key}`] = { paddingBottom: value }
      classes[`.pl-${key}`] = { paddingLeft: value }
      classes[`.px-${key}`] = { paddingLeft: value, paddingRight: value }
      classes[`.py-${key}`] = { paddingTop: value, paddingBottom: value }
    })

    // Gap classes
    Object.entries(spacing).forEach(([key, value]) => {
      classes[`.gap-${key}`] = { gap: value }
      classes[`.gap-x-${key}`] = { columnGap: value }
      classes[`.gap-y-${key}`] = { rowGap: value }
    })

    return classes
  }
} as const

// Accessibility spacing considerations
export const accessibilitySpacing = {
  // Minimum touch target sizes
  touchTarget: {
    minimum: spacing[11],  // 44px (iOS/Android minimum)
    recommended: spacing[12] // 48px (recommended)
  },

  // Focus ring spacing
  focusRing: {
    offset: spacing[1],    // 4px offset
    width: spacing[0.5]    // 2px width
  },

  // Reading comfort spacing
  reading: {
    lineHeight: '1.5',     // Minimum for readability
    paragraphSpacing: spacing[4], // 16px between paragraphs
    letterSpacing: '0.02em' // Subtle letter spacing
  },

  // Form spacing
  form: {
    fieldSpacing: spacing[4],      // 16px between fields
    labelSpacing: spacing[1],      // 4px between label and input
    helpTextSpacing: spacing[1]    // 4px between input and help text
  }
} as const

// Export types
export type SpacingKey = keyof typeof spacing
export type SemanticSpacingCategory = keyof typeof semanticSpacing
export type ResponsiveBreakpoint = keyof typeof responsiveSpacing.breakpoints
export type SpacingValue = typeof spacing[keyof typeof spacing]

export default {
  spacing,
  semanticSpacing,
  responsiveSpacing,
  logicalSpacing,
  specialSpacing,
  spacingUtils,
  accessibilitySpacing
}