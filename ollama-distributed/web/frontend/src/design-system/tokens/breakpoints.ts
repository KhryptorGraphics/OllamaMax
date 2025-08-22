/**
 * Design Tokens: Breakpoints
 * Responsive design breakpoints with mobile-first approach and container queries
 */

// Base breakpoint values (mobile-first)
export const breakpoints = {
  xs: '320px',    // Small mobile devices
  sm: '640px',    // Mobile devices
  md: '768px',    // Tablets
  lg: '1024px',   // Small desktop
  xl: '1280px',   // Desktop
  '2xl': '1536px', // Large desktop
  '3xl': '1920px', // Ultra-wide displays
  '4xl': '2560px'  // 4K displays
} as const

// Semantic breakpoint groups
export const semanticBreakpoints = {
  // Device categories
  device: {
    mobile: {
      min: breakpoints.xs,
      max: `calc(${breakpoints.md} - 1px)` // Up to 767px
    },
    tablet: {
      min: breakpoints.md,
      max: `calc(${breakpoints.lg} - 1px)` // 768px to 1023px
    },
    desktop: {
      min: breakpoints.lg,
      max: breakpoints['4xl'] // 1024px and up
    }
  },
  
  // Content width categories
  content: {
    narrow: {
      min: breakpoints.xs,
      max: breakpoints.sm,
      container: '100%',
      padding: '1rem'
    },
    medium: {
      min: breakpoints.sm,
      max: breakpoints.lg,
      container: '768px',
      padding: '1.5rem'
    },
    wide: {
      min: breakpoints.lg,
      max: breakpoints['2xl'],
      container: '1200px',
      padding: '2rem'
    },
    ultraWide: {
      min: breakpoints['2xl'],
      max: breakpoints['4xl'],
      container: '1400px',
      padding: '3rem'
    }
  },
  
  // Component-specific breakpoints
  component: {
    // Navigation breakpoints
    navigation: {
      mobileMenu: `(max-width: ${breakpoints.md})`,
      desktopMenu: `(min-width: ${breakpoints.md})`
    },
    
    // Sidebar breakpoints
    sidebar: {
      collapsed: `(max-width: ${breakpoints.lg})`,
      expanded: `(min-width: calc(${breakpoints.lg} + 1px))`
    },
    
    // Grid breakpoints
    grid: {
      singleColumn: `(max-width: ${breakpoints.sm})`,
      twoColumn: `(min-width: ${breakpoints.sm}) and (max-width: ${breakpoints.lg})`,
      multiColumn: `(min-width: calc(${breakpoints.lg} + 1px))`
    },
    
    // Modal breakpoints
    modal: {
      fullscreen: `(max-width: ${breakpoints.md})`,
      centered: `(min-width: calc(${breakpoints.md} + 1px))`
    }
  }
} as const

// Container queries (CSS Container Queries)
export const containerQueries = {
  // Standard container sizes
  container: {
    xs: '20rem',    // 320px
    sm: '24rem',    // 384px
    md: '28rem',    // 448px
    lg: '32rem',    // 512px
    xl: '36rem',    // 576px
    '2xl': '42rem', // 672px
    '3xl': '48rem', // 768px
    '4xl': '56rem', // 896px
    '5xl': '64rem', // 1024px
    '6xl': '72rem', // 1152px
    '7xl': '80rem'  // 1280px
  },
  
  // Component-specific container queries
  component: {
    card: {
      compact: '16rem',   // 256px
      comfortable: '24rem', // 384px
      spacious: '32rem'   // 512px
    },
    
    dataTable: {
      mobile: '20rem',    // 320px
      tablet: '48rem',    // 768px
      desktop: '64rem'    // 1024px
    },
    
    form: {
      single: '20rem',    // 320px
      double: '40rem',    // 640px
      triple: '60rem'     // 960px
    }
  }
} as const

// Media query utilities
export const mediaQueries = {
  // Generate min-width media query
  up: (breakpoint: keyof typeof breakpoints) => {
    return `(min-width: ${breakpoints[breakpoint]})`
  },
  
  // Generate max-width media query
  down: (breakpoint: keyof typeof breakpoints) => {
    const value = parseInt(breakpoints[breakpoint])
    return `(max-width: ${value - 1}px)`
  },
  
  // Generate range media query
  between: (
    minBreakpoint: keyof typeof breakpoints,
    maxBreakpoint: keyof typeof breakpoints
  ) => {
    const minValue = breakpoints[minBreakpoint]
    const maxValue = parseInt(breakpoints[maxBreakpoint]) - 1
    return `(min-width: ${minValue}) and (max-width: ${maxValue}px)`
  },
  
  // Generate only media query (exactly at breakpoint)
  only: (breakpoint: keyof typeof breakpoints) => {
    const current = parseInt(breakpoints[breakpoint])
    const breakpointKeys = Object.keys(breakpoints) as Array<keyof typeof breakpoints>
    const currentIndex = breakpointKeys.indexOf(breakpoint)
    
    if (currentIndex === breakpointKeys.length - 1) {
      // Last breakpoint, only min-width
      return `(min-width: ${breakpoints[breakpoint]})`
    }
    
    const nextBreakpoint = breakpointKeys[currentIndex + 1]
    const nextValue = parseInt(breakpoints[nextBreakpoint]) - 1
    
    return `(min-width: ${breakpoints[breakpoint]}) and (max-width: ${nextValue}px)`
  }
} as const

// Responsive utilities
export const responsiveUtils = {
  // Check if current viewport matches breakpoint
  matches: (query: string): boolean => {
    if (typeof window === 'undefined') return false
    return window.matchMedia(query).matches
  },
  
  // Get current breakpoint
  getCurrentBreakpoint: (): keyof typeof breakpoints | null => {
    if (typeof window === 'undefined') return null
    
    const width = window.innerWidth
    const breakpointEntries = Object.entries(breakpoints)
      .sort(([, a], [, b]) => parseInt(b) - parseInt(a)) // Sort descending
    
    for (const [key, value] of breakpointEntries) {
      if (width >= parseInt(value)) {
        return key as keyof typeof breakpoints
      }
    }
    
    return 'xs' // fallback to smallest breakpoint
  },
  
  // Check if viewport is mobile
  isMobile: (): boolean => {
    return responsiveUtils.matches(mediaQueries.down('md'))
  },
  
  // Check if viewport is tablet
  isTablet: (): boolean => {
    return responsiveUtils.matches(mediaQueries.between('md', 'lg'))
  },
  
  // Check if viewport is desktop
  isDesktop: (): boolean => {
    return responsiveUtils.matches(mediaQueries.up('lg'))
  },
  
  // Get container width for breakpoint
  getContainerWidth: (breakpoint: keyof typeof breakpoints) => {
    const breakpointValue = parseInt(breakpoints[breakpoint])
    
    // Define container max-widths based on breakpoint
    if (breakpointValue >= parseInt(breakpoints['2xl'])) return '1400px'
    if (breakpointValue >= parseInt(breakpoints.xl)) return '1200px'
    if (breakpointValue >= parseInt(breakpoints.lg)) return '1024px'
    if (breakpointValue >= parseInt(breakpoints.md)) return '768px'
    return '100%'
  }
} as const

// Breakpoint utility functions
export const breakpointUtils = {
  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}
    
    // Base breakpoints
    Object.entries(breakpoints).forEach(([key, value]) => {
      cssVars[`--breakpoint-${key}`] = value
    })
    
    // Container query values
    Object.entries(containerQueries.container).forEach(([key, value]) => {
      cssVars[`--container-${key}`] = value
    })
    
    // Component container queries
    Object.entries(containerQueries.component).forEach(([component, sizes]) => {
      Object.entries(sizes).forEach(([size, value]) => {
        cssVars[`--container-${component}-${size}`] = value
      })
    })
    
    return cssVars
  },
  
  // Generate Tailwind CSS responsive utilities
  generateTailwindConfig: () => {
    return {
      screens: {
        ...breakpoints,
        // Add container queries when supported
        '@container': containerQueries.container
      }
    }
  },
  
  // Generate media query CSS
  generateMediaQueries: () => {
    const queries: Record<string, string> = {}
    
    Object.entries(breakpoints).forEach(([key, value]) => {
      queries[`@media-${key}-up`] = `@media ${mediaQueries.up(key as keyof typeof breakpoints)}`
      queries[`@media-${key}-down`] = `@media ${mediaQueries.down(key as keyof typeof breakpoints)}`
      queries[`@media-${key}-only`] = `@media ${mediaQueries.only(key as keyof typeof breakpoints)}`
    })
    
    return queries
  }
} as const

// Accessibility considerations for breakpoints
export const accessibilityBreakpoints = {
  // Minimum touch target sizes at different breakpoints
  touchTargets: {
    mobile: {
      minimum: '44px',      // iOS/Android minimum
      recommended: '48px'   // Better usability
    },
    tablet: {
      minimum: '44px',
      recommended: '52px'
    },
    desktop: {
      minimum: '32px',      // Mouse precision allows smaller targets
      recommended: '40px'
    }
  },
  
  // Font size adjustments for different screen sizes
  fontScaling: {
    mobile: {
      base: '16px',         // Never go below 16px to prevent zoom
      scale: 0.9
    },
    tablet: {
      base: '16px',
      scale: 1.0
    },
    desktop: {
      base: '16px',
      scale: 1.1
    }
  },
  
  // Focus ring sizes for different breakpoints
  focusRing: {
    mobile: '3px',          // Larger for touch interfaces
    tablet: '2px',
    desktop: '2px'
  }
} as const

// Export types
export type Breakpoint = keyof typeof breakpoints
export type DeviceCategory = keyof typeof semanticBreakpoints.device
export type ContentCategory = keyof typeof semanticBreakpoints.content
export type ContainerSize = keyof typeof containerQueries.container
export type BreakpointValue = typeof breakpoints[keyof typeof breakpoints]

export default {
  breakpoints,
  semanticBreakpoints,
  containerQueries,
  mediaQueries,
  responsiveUtils,
  breakpointUtils,
  accessibilityBreakpoints
}