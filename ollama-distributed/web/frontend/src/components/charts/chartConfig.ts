/**
 * Chart Configuration System
 * Global configuration for chart themes, responsive breakpoints, animations,
 * and accessibility features
 */

import { colorUtils, responsiveUtils, type ChartTheme } from '@/utils/chartUtils'

// Animation configurations
export const animationConfig = {
  // Animation durations in milliseconds
  durations: {
    fast: 300,
    normal: 600,
    slow: 1000,
    very_slow: 1500
  },
  
  // Animation easing functions
  easing: {
    ease: 'ease',
    linear: 'linear',
    easeIn: 'ease-in',
    easeOut: 'ease-out',
    easeInOut: 'ease-in-out',
    bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
    smooth: 'cubic-bezier(0.4, 0, 0.2, 1)'
  },
  
  // Animation delays for staggered effects
  delays: {
    none: 0,
    short: 100,
    medium: 200,
    long: 400
  },
  
  // Performance settings
  performance: {
    // Disable animations on low-end devices
    respectMotionPreference: true,
    // Maximum number of animated elements
    maxAnimatedElements: 100,
    // Fallback to CSS animations for large datasets
    useCSSAnimations: true
  }
} as const

// Responsive breakpoints for charts
export const responsiveConfig = {
  breakpoints: {
    xs: 320,
    sm: 640,
    md: 768,
    lg: 1024,
    xl: 1280,
    xxl: 1536
  },
  
  // Chart dimensions by breakpoint
  chartDimensions: {
    xs: { width: 280, height: 180 },
    sm: { width: 400, height: 250 },
    md: { width: 600, height: 350 },
    lg: { width: 800, height: 450 },
    xl: { width: 1000, height: 550 },
    xxl: { width: 1200, height: 650 }
  },
  
  // Margin adjustments by breakpoint
  margins: {
    xs: { top: 10, right: 10, bottom: 30, left: 30 },
    sm: { top: 20, right: 20, bottom: 40, left: 40 },
    md: { top: 30, right: 30, bottom: 50, left: 60 },
    lg: { top: 40, right: 40, bottom: 60, left: 80 },
    xl: { top: 50, right: 50, bottom: 70, left: 100 },
    xxl: { top: 60, right: 60, bottom: 80, left: 120 }
  },
  
  // Font sizes by breakpoint
  fontSizes: {
    xs: { axis: 10, legend: 10, tooltip: 11, title: 14 },
    sm: { axis: 11, legend: 11, tooltip: 12, title: 16 },
    md: { axis: 12, legend: 12, tooltip: 13, title: 18 },
    lg: { axis: 13, legend: 13, tooltip: 14, title: 20 },
    xl: { axis: 14, legend: 14, tooltip: 15, title: 22 },
    xxl: { axis: 15, legend: 15, tooltip: 16, title: 24 }
  }
} as const

// Accessibility configuration
export const accessibilityConfig = {
  // WCAG 2.1 AA compliant colors
  contrastRatios: {
    // Normal text: 4.5:1
    normalText: 4.5,
    // Large text: 3:1
    largeText: 3.0,
    // UI components: 3:1
    uiComponents: 3.0
  },
  
  // Keyboard navigation
  keyboard: {
    // Enable keyboard navigation for interactive charts
    enabled: true,
    // Tab order for chart elements
    tabOrder: ['chart', 'legend', 'controls', 'export'],
    // Keyboard shortcuts
    shortcuts: {
      export: 'e',
      fullscreen: 'f',
      refresh: 'r',
      zoomIn: '+',
      zoomOut: '-',
      reset: '0'
    }
  },
  
  // Screen reader support
  screenReader: {
    // Enable ARIA labels and descriptions
    enabled: true,
    // Include data summaries for screen readers
    includeSummary: true,
    // Maximum number of data points to announce
    maxDataPoints: 20
  },
  
  // Motion settings
  motion: {
    // Respect user's motion preferences
    respectPreference: true,
    // Reduced motion alternatives
    reducedMotion: {
      disableAnimations: true,
      useStaticTransitions: true,
      simplifyInteractions: true
    }
  },
  
  // Color accessibility
  colorAccessibility: {
    // Ensure sufficient contrast ratios
    enforceContrast: true,
    // Provide alternative indicators for color-blind users
    usePatterns: true,
    // Include color names in tooltips
    includeColorNames: true
  }
} as const

// Performance optimization settings
export const performanceConfig = {
  // Data point limits for different chart types
  dataLimits: {
    line: 1000,
    area: 800,
    bar: 500,
    pie: 50,
    scatter: 2000,
    heatmap: 10000
  },
  
  // Virtualization settings
  virtualization: {
    // Enable virtualization for large datasets
    enabled: true,
    // Threshold for enabling virtualization
    threshold: 500,
    // Chunk size for virtual rendering
    chunkSize: 100
  },
  
  // Debouncing for interactive features
  debouncing: {
    // Resize debounce delay
    resize: 150,
    // Hover debounce delay
    hover: 50,
    // Search/filter debounce delay
    search: 300
  },
  
  // Memory management
  memory: {
    // Maximum cache size in MB
    maxCacheSize: 50,
    // Cache TTL in milliseconds
    cacheTTL: 5 * 60 * 1000, // 5 minutes
    // Enable garbage collection hints
    enableGC: true
  }
} as const

// Theme configurations
export const themeConfig = {
  // Predefined color palettes
  palettes: {
    default: ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316'],
    pastel: ['#a78bfa', '#fbbf24', '#34d399', '#f87171', '#60a5fa', '#fb7185', '#fde047', '#a3e635'],
    vibrant: ['#1e40af', '#dc2626', '#059669', '#d97706', '#7c3aed', '#0284c7', '#65a30d', '#ea580c'],
    monochrome: ['#374151', '#6b7280', '#9ca3af', '#d1d5db', '#e5e7eb', '#f3f4f6', '#f9fafb', '#ffffff'],
    semantic: {
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neutral: '#6b7280'
    }
  },
  
  // Dark/light mode configurations
  modes: {
    light: {
      background: '#ffffff',
      surface: '#f9fafb',
      text: '#111827',
      textSecondary: '#6b7280',
      border: '#e5e7eb',
      grid: '#f3f4f6'
    },
    dark: {
      background: '#111827',
      surface: '#1f2937',
      text: '#f9fafb',
      textSecondary: '#d1d5db',
      border: '#374151',
      grid: '#2d3748'
    }
  },
  
  // Font configurations
  typography: {
    families: {
      primary: 'Inter, system-ui, sans-serif',
      mono: 'JetBrains Mono, Consolas, monospace'
    },
    weights: {
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700
    }
  }
} as const

// Chart-specific default configurations
export const chartDefaults = {
  // LineChart defaults
  lineChart: {
    strokeWidth: 2,
    dot: false,
    activeDot: { r: 4 },
    connectNulls: false,
    animationDuration: animationConfig.durations.normal
  },
  
  // AreaChart defaults
  areaChart: {
    strokeWidth: 1,
    fillOpacity: 0.6,
    connectNulls: false,
    animationDuration: animationConfig.durations.normal
  },
  
  // BarChart defaults
  barChart: {
    radius: 0,
    maxBarSize: 100,
    animationDuration: animationConfig.durations.normal
  },
  
  // PieChart defaults
  pieChart: {
    startAngle: 90,
    endAngle: -270,
    innerRadius: 0,
    outerRadius: '80%',
    paddingAngle: 0,
    animationDuration: animationConfig.durations.slow
  },
  
  // Heatmap defaults
  heatmap: {
    cellSize: 'auto',
    cellGap: 1,
    cellRadius: 2,
    showBorder: false
  },
  
  // GaugeChart defaults
  gaugeChart: {
    thickness: 20,
    startAngle: 180,
    endAngle: 0,
    animationDuration: animationConfig.durations.slow
  }
} as const

// Tooltip configurations
export const tooltipConfig = {
  // Default tooltip styles
  styles: {
    backgroundColor: 'rgba(0, 0, 0, 0.9)',
    color: 'white',
    padding: '8px 12px',
    borderRadius: '6px',
    fontSize: '12px',
    maxWidth: '300px',
    zIndex: 1000
  },
  
  // Animation settings
  animation: {
    duration: animationConfig.durations.fast,
    easing: animationConfig.easing.easeOut
  },
  
  // Positioning
  positioning: {
    offset: 10,
    strategy: 'follow' as 'follow' | 'fixed',
    boundary: 'viewport'
  }
} as const

// Export configurations
export const exportConfig = {
  // Default file formats
  formats: {
    png: {
      quality: 0.9,
      scale: 2, // High DPI
      backgroundColor: '#ffffff'
    },
    svg: {
      preserveAspectRatio: true,
      encoding: 'utf-8'
    },
    pdf: {
      format: 'a4' as const,
      orientation: 'landscape' as const,
      margin: 10
    },
    csv: {
      delimiter: ',',
      encoding: 'utf-8',
      includeHeaders: true
    }
  },
  
  // Default filenames
  filenames: {
    timestamp: true,
    format: 'YYYY-MM-DD_HH-mm-ss',
    prefix: 'chart'
  }
} as const

// Utility functions for configuration
export const configUtils = {
  /**
   * Get responsive configuration based on container width
   */
  getResponsiveConfig: (width: number) => {
    const { breakpoints, chartDimensions, margins, fontSizes } = responsiveConfig
    
    let breakpoint: keyof typeof breakpoints = 'xs'
    
    if (width >= breakpoints.xxl) breakpoint = 'xxl'
    else if (width >= breakpoints.xl) breakpoint = 'xl'
    else if (width >= breakpoints.lg) breakpoint = 'lg'
    else if (width >= breakpoints.md) breakpoint = 'md'
    else if (width >= breakpoints.sm) breakpoint = 'sm'
    
    return {
      breakpoint,
      dimensions: chartDimensions[breakpoint],
      margins: margins[breakpoint],
      fontSizes: fontSizes[breakpoint]
    }
  },
  
  /**
   * Create chart theme with mode and palette
   */
  createTheme: (mode: 'light' | 'dark', palette?: string[]): ChartTheme => {
    const baseTheme = colorUtils.generateChartTheme(mode)
    
    if (palette) {
      return {
        ...baseTheme,
        colors: {
          ...baseTheme.colors,
          primary: palette
        }
      }
    }
    
    return baseTheme
  },
  
  /**
   * Check if animations should be disabled based on user preferences
   */
  shouldDisableAnimations: (): boolean => {
    if (!accessibilityConfig.motion.respectPreference) return false
    
    // Check for reduced motion preference
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  },
  
  /**
   * Get chart-specific defaults
   */
  getChartDefaults: (chartType: keyof typeof chartDefaults) => {
    const defaults = chartDefaults[chartType]
    const shouldDisableAnimations = configUtils.shouldDisableAnimations()
    
    return {
      ...defaults,
      animationDuration: shouldDisableAnimations ? 0 : defaults.animationDuration
    }
  },
  
  /**
   * Validate data size against performance limits
   */
  validateDataSize: (chartType: keyof typeof performanceConfig.dataLimits, dataSize: number) => {
    const limit = performanceConfig.dataLimits[chartType]
    
    return {
      isValid: dataSize <= limit,
      limit,
      dataSize,
      shouldVirtualize: performanceConfig.virtualization.enabled && 
                      dataSize >= performanceConfig.virtualization.threshold
    }
  }
}

export type ResponsiveBreakpoint = keyof typeof responsiveConfig.breakpoints
export type ChartType = keyof typeof chartDefaults
export type ColorPalette = keyof typeof themeConfig.palettes
export type ThemeMode = keyof typeof themeConfig.modes

export {
  type ChartTheme
}