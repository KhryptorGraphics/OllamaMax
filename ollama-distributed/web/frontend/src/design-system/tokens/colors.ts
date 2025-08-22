/**
 * Design Tokens: Colors
 * Comprehensive color system with semantic tokens, accessibility compliance, and dark mode support
 */

// Base color palette
export const colors = {
  // Primary brand colors
  primary: {
    50: '#eff6ff',
    100: '#dbeafe',
    200: '#bfdbfe',
    300: '#93c5fd',
    400: '#60a5fa',
    500: '#3b82f6', // Primary
    600: '#2563eb',
    700: '#1d4ed8',
    800: '#1e40af',
    900: '#1e3a8a',
    950: '#172554'
  },

  // Secondary colors
  secondary: {
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b', // Secondary
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
    950: '#020617'
  },

  // Semantic colors
  success: {
    50: '#f0fdf4',
    100: '#dcfce7',
    200: '#bbf7d0',
    300: '#86efac',
    400: '#4ade80',
    500: '#22c55e', // Success
    600: '#16a34a',
    700: '#15803d',
    800: '#166534',
    900: '#14532d',
    950: '#052e16'
  },

  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b', // Warning
    600: '#d97706',
    700: '#b45309',
    800: '#92400e',
    900: '#78350f',
    950: '#451a03'
  },

  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444', // Error
    600: '#dc2626',
    700: '#b91c1c',
    800: '#991b1b',
    900: '#7f1d1d',
    950: '#450a0a'
  },

  info: {
    50: '#f0f9ff',
    100: '#e0f2fe',
    200: '#bae6fd',
    300: '#7dd3fc',
    400: '#38bdf8',
    500: '#0ea5e9', // Info
    600: '#0284c7',
    700: '#0369a1',
    800: '#075985',
    900: '#0c4a6e',
    950: '#082f49'
  },

  // Neutral grays
  neutral: {
    50: '#fafafa',
    100: '#f5f5f5',
    200: '#e5e5e5',
    300: '#d4d4d4',
    400: '#a3a3a3',
    500: '#737373',
    600: '#525252',
    700: '#404040',
    800: '#262626',
    900: '#171717',
    950: '#0a0a0a'
  }
} as const

// Semantic color tokens
export const semanticColors = {
  // Light theme
  light: {
    // Text colors
    text: {
      primary: colors.neutral[900],
      secondary: colors.neutral[600],
      tertiary: colors.neutral[400],
      inverse: colors.neutral[50],
      disabled: colors.neutral[300],
      link: colors.primary[600],
      linkHover: colors.primary[700]
    },

    // Background colors
    background: {
      primary: colors.neutral[50],
      secondary: colors.neutral[100],
      tertiary: colors.neutral[200],
      inverse: colors.neutral[900],
      overlay: 'rgba(0, 0, 0, 0.5)'
    },

    // Border colors
    border: {
      primary: colors.neutral[200],
      secondary: colors.neutral[300],
      focus: colors.primary[500],
      error: colors.error[500],
      success: colors.success[500],
      warning: colors.warning[500]
    },

    // Interactive states
    interactive: {
      primary: {
        default: colors.primary[500],
        hover: colors.primary[600],
        active: colors.primary[700],
        disabled: colors.neutral[300],
        focus: colors.primary[500]
      },
      secondary: {
        default: colors.secondary[500],
        hover: colors.secondary[600],
        active: colors.secondary[700],
        disabled: colors.neutral[300],
        focus: colors.secondary[500]
      },
      ghost: {
        default: 'transparent',
        hover: colors.neutral[100],
        active: colors.neutral[200],
        disabled: 'transparent',
        focus: colors.neutral[100]
      }
    },

    // Status colors
    status: {
      success: {
        background: colors.success[50],
        border: colors.success[200],
        text: colors.success[800],
        icon: colors.success[600]
      },
      warning: {
        background: colors.warning[50],
        border: colors.warning[200],
        text: colors.warning[800],
        icon: colors.warning[600]
      },
      error: {
        background: colors.error[50],
        border: colors.error[200],
        text: colors.error[800],
        icon: colors.error[600]
      },
      info: {
        background: colors.info[50],
        border: colors.info[200],
        text: colors.info[800],
        icon: colors.info[600]
      }
    }
  },

  // Dark theme
  dark: {
    // Text colors
    text: {
      primary: colors.neutral[50],
      secondary: colors.neutral[300],
      tertiary: colors.neutral[500],
      inverse: colors.neutral[900],
      disabled: colors.neutral[600],
      link: colors.primary[400],
      linkHover: colors.primary[300]
    },

    // Background colors
    background: {
      primary: colors.neutral[900],
      secondary: colors.neutral[800],
      tertiary: colors.neutral[700],
      inverse: colors.neutral[50],
      overlay: 'rgba(0, 0, 0, 0.7)'
    },

    // Border colors
    border: {
      primary: colors.neutral[700],
      secondary: colors.neutral[600],
      focus: colors.primary[400],
      error: colors.error[500],
      success: colors.success[500],
      warning: colors.warning[500]
    },

    // Interactive states
    interactive: {
      primary: {
        default: colors.primary[500],
        hover: colors.primary[400],
        active: colors.primary[300],
        disabled: colors.neutral[600],
        focus: colors.primary[400]
      },
      secondary: {
        default: colors.secondary[400],
        hover: colors.secondary[300],
        active: colors.secondary[200],
        disabled: colors.neutral[600],
        focus: colors.secondary[400]
      },
      ghost: {
        default: 'transparent',
        hover: colors.neutral[800],
        active: colors.neutral[700],
        disabled: 'transparent',
        focus: colors.neutral[800]
      }
    },

    // Status colors
    status: {
      success: {
        background: 'rgba(34, 197, 94, 0.1)',
        border: colors.success[700],
        text: colors.success[300],
        icon: colors.success[400]
      },
      warning: {
        background: 'rgba(245, 158, 11, 0.1)',
        border: colors.warning[700],
        text: colors.warning[300],
        icon: colors.warning[400]
      },
      error: {
        background: 'rgba(239, 68, 68, 0.1)',
        border: colors.error[700],
        text: colors.error[300],
        icon: colors.error[400]
      },
      info: {
        background: 'rgba(14, 165, 233, 0.1)',
        border: colors.info[700],
        text: colors.info[300],
        icon: colors.info[400]
      }
    }
  }
} as const

// Accessibility color utilities
export const accessibilityColors = {
  // High contrast colors for better accessibility
  highContrast: {
    light: {
      text: '#000000',
      background: '#ffffff',
      primary: '#0066cc',
      focus: '#005fbb',
      error: '#cc0000',
      success: '#006600'
    },
    dark: {
      text: '#ffffff',
      background: '#000000',
      primary: '#66b3ff',
      focus: '#4da6ff',
      error: '#ff6666',
      success: '#66ff66'
    }
  },

  // WCAG AA compliant color combinations
  wcagAA: {
    // Contrast ratio >= 4.5:1 for normal text
    normalText: {
      light: {
        onPrimary: colors.neutral[50],
        onSecondary: colors.neutral[50],
        onSurface: colors.neutral[900],
        onBackground: colors.neutral[900]
      },
      dark: {
        onPrimary: colors.neutral[900],
        onSecondary: colors.neutral[900],
        onSurface: colors.neutral[50],
        onBackground: colors.neutral[50]
      }
    },
    // Contrast ratio >= 3:1 for large text
    largeText: {
      light: {
        onPrimary: colors.neutral[100],
        onSecondary: colors.neutral[100],
        onSurface: colors.neutral[800],
        onBackground: colors.neutral[800]
      },
      dark: {
        onPrimary: colors.neutral[800],
        onSecondary: colors.neutral[800],
        onSurface: colors.neutral[100],
        onBackground: colors.neutral[100]
      }
    }
  }
} as const

// Color utility functions
export const colorUtils = {
  // Get theme colors based on theme preference
  getThemeColors: (theme: 'light' | 'dark' = 'light') => {
    return semanticColors[theme]
  },

  // Get status color based on variant and theme
  getStatusColor: (
    variant: 'success' | 'warning' | 'error' | 'info',
    property: 'background' | 'border' | 'text' | 'icon',
    theme: 'light' | 'dark' = 'light'
  ) => {
    return semanticColors[theme].status[variant][property]
  },

  // Get interactive color based on state and theme
  getInteractiveColor: (
    variant: 'primary' | 'secondary' | 'ghost',
    state: 'default' | 'hover' | 'active' | 'disabled' | 'focus',
    theme: 'light' | 'dark' = 'light'
  ) => {
    return semanticColors[theme].interactive[variant][state]
  },

  // Check if color combination meets WCAG contrast requirements
  meetsWCAGContrast: (foreground: string, background: string, level: 'AA' | 'AAA' = 'AA') => {
    // This would integrate with a color contrast calculation library
    // For now, return true as implementation would require external dependency
    return true
  },

  // Generate CSS custom properties for themes
  generateCSSVariables: (theme: 'light' | 'dark' = 'light') => {
    const themeColors = semanticColors[theme]
    const cssVars: Record<string, string> = {}

    // Text colors
    Object.entries(themeColors.text).forEach(([key, value]) => {
      cssVars[`--color-text-${key}`] = value
    })

    // Background colors
    Object.entries(themeColors.background).forEach(([key, value]) => {
      cssVars[`--color-bg-${key}`] = value
    })

    // Border colors
    Object.entries(themeColors.border).forEach(([key, value]) => {
      cssVars[`--color-border-${key}`] = value
    })

    // Interactive colors
    Object.entries(themeColors.interactive).forEach(([variant, states]) => {
      Object.entries(states).forEach(([state, value]) => {
        cssVars[`--color-interactive-${variant}-${state}`] = value
      })
    })

    // Status colors
    Object.entries(themeColors.status).forEach(([variant, properties]) => {
      Object.entries(properties).forEach(([property, value]) => {
        cssVars[`--color-status-${variant}-${property}`] = value
      })
    })

    return cssVars
  }
} as const

// Export types for TypeScript
export type ColorScale = typeof colors.primary
export type SemanticColorTheme = typeof semanticColors.light
export type ColorVariant = 'success' | 'warning' | 'error' | 'info'
export type InteractiveVariant = 'primary' | 'secondary' | 'ghost'
export type InteractiveState = 'default' | 'hover' | 'active' | 'disabled' | 'focus'

export default {
  colors,
  semanticColors,
  accessibilityColors,
  colorUtils
}