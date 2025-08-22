/**
 * Design Tokens: Typography
 * Comprehensive typography system with fluid scales, accessibility features, and responsive design
 */

// Font families
export const fontFamilies = {
  sans: [
    'Inter',
    '-apple-system',
    'BlinkMacSystemFont',
    '"Segoe UI"',
    'Roboto',
    '"Helvetica Neue"',
    'Arial',
    'sans-serif'
  ],
  serif: [
    'Charter',
    'Georgia',
    '"Times New Roman"',
    'serif'
  ],
  mono: [
    '"Fira Code"',
    'Consolas',
    '"Liberation Mono"',
    'Menlo',
    'Monaco',
    'monospace'
  ]
} as const

// Font weights
export const fontWeights = {
  thin: 100,
  extralight: 200,
  light: 300,
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
  extrabold: 800,
  black: 900
} as const

// Base font sizes (rem values)
export const fontSizes = {
  xs: '0.75rem',    // 12px
  sm: '0.875rem',   // 14px
  base: '1rem',     // 16px
  lg: '1.125rem',   // 18px
  xl: '1.25rem',    // 20px
  '2xl': '1.5rem',  // 24px
  '3xl': '1.875rem', // 30px
  '4xl': '2.25rem', // 36px
  '5xl': '3rem',    // 48px
  '6xl': '3.75rem', // 60px
  '7xl': '4.5rem',  // 72px
  '8xl': '6rem',    // 96px
  '9xl': '8rem'     // 128px
} as const

// Line heights
export const lineHeights = {
  none: '1',
  tight: '1.25',
  snug: '1.375',
  normal: '1.5',
  relaxed: '1.625',
  loose: '2'
} as const

// Letter spacing
export const letterSpacing = {
  tighter: '-0.05em',
  tight: '-0.025em',
  normal: '0em',
  wide: '0.025em',
  wider: '0.05em',
  widest: '0.1em'
} as const

// Typography scale with semantic naming
export const typographyScale = {
  // Display text (hero, landing pages)
  display: {
    large: {
      fontSize: fontSizes['7xl'],
      lineHeight: lineHeights.none,
      fontWeight: fontWeights.bold,
      letterSpacing: letterSpacing.tighter,
      fontFamily: fontFamilies.sans.join(', ')
    },
    medium: {
      fontSize: fontSizes['6xl'],
      lineHeight: lineHeights.none,
      fontWeight: fontWeights.bold,
      letterSpacing: letterSpacing.tighter,
      fontFamily: fontFamilies.sans.join(', ')
    },
    small: {
      fontSize: fontSizes['5xl'],
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.bold,
      letterSpacing: letterSpacing.tight,
      fontFamily: fontFamilies.sans.join(', ')
    }
  },

  // Headings
  heading: {
    h1: {
      fontSize: fontSizes['4xl'],
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.bold,
      letterSpacing: letterSpacing.tight,
      fontFamily: fontFamilies.sans.join(', ')
    },
    h2: {
      fontSize: fontSizes['3xl'],
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.semibold,
      letterSpacing: letterSpacing.tight,
      fontFamily: fontFamilies.sans.join(', ')
    },
    h3: {
      fontSize: fontSizes['2xl'],
      lineHeight: lineHeights.snug,
      fontWeight: fontWeights.semibold,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    h4: {
      fontSize: fontSizes.xl,
      lineHeight: lineHeights.snug,
      fontWeight: fontWeights.semibold,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    h5: {
      fontSize: fontSizes.lg,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    h6: {
      fontSize: fontSizes.base,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.wide,
      fontFamily: fontFamilies.sans.join(', ')
    }
  },

  // Body text
  body: {
    large: {
      fontSize: fontSizes.lg,
      lineHeight: lineHeights.relaxed,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    medium: {
      fontSize: fontSizes.base,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    small: {
      fontSize: fontSizes.sm,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    }
  },

  // Labels and captions
  label: {
    large: {
      fontSize: fontSizes.base,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    medium: {
      fontSize: fontSizes.sm,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    small: {
      fontSize: fontSizes.xs,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.wide,
      fontFamily: fontFamilies.sans.join(', ')
    }
  },

  // Code and monospace
  code: {
    large: {
      fontSize: fontSizes.base,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.mono.join(', ')
    },
    medium: {
      fontSize: fontSizes.sm,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.mono.join(', ')
    },
    small: {
      fontSize: fontSizes.xs,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.normal,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.mono.join(', ')
    }
  },

  // Button text
  button: {
    large: {
      fontSize: fontSizes.base,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    medium: {
      fontSize: fontSizes.sm,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.normal,
      fontFamily: fontFamilies.sans.join(', ')
    },
    small: {
      fontSize: fontSizes.xs,
      lineHeight: lineHeights.normal,
      fontWeight: fontWeights.medium,
      letterSpacing: letterSpacing.wide,
      fontFamily: fontFamilies.sans.join(', ')
    }
  }
} as const

// Responsive typography scale
export const responsiveTypography = {
  // Fluid typography using clamp()
  display: {
    large: {
      fontSize: 'clamp(3rem, 8vw, 4.5rem)',
      lineHeight: lineHeights.none,
      fontWeight: fontWeights.bold
    },
    medium: {
      fontSize: 'clamp(2.25rem, 6vw, 3.75rem)',
      lineHeight: lineHeights.none,
      fontWeight: fontWeights.bold
    },
    small: {
      fontSize: 'clamp(1.875rem, 5vw, 3rem)',
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.bold
    }
  },

  heading: {
    h1: {
      fontSize: 'clamp(1.875rem, 4vw, 2.25rem)',
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.bold
    },
    h2: {
      fontSize: 'clamp(1.5rem, 3vw, 1.875rem)',
      lineHeight: lineHeights.tight,
      fontWeight: fontWeights.semibold
    },
    h3: {
      fontSize: 'clamp(1.25rem, 2.5vw, 1.5rem)',
      lineHeight: lineHeights.snug,
      fontWeight: fontWeights.semibold
    }
  },

  body: {
    large: {
      fontSize: 'clamp(1rem, 1.5vw, 1.125rem)',
      lineHeight: lineHeights.relaxed
    },
    medium: {
      fontSize: 'clamp(0.875rem, 1vw, 1rem)',
      lineHeight: lineHeights.normal
    }
  }
} as const

// Typography utilities
export const typographyUtils = {
  // Get typography style by semantic name
  getStyle: (
    category: keyof typeof typographyScale,
    variant: string
  ) => {
    const styles = typographyScale[category] as any
    return styles[variant] || styles.medium || null
  },

  // Get responsive typography style
  getResponsiveStyle: (
    category: keyof typeof responsiveTypography,
    variant: string
  ) => {
    const styles = responsiveTypography[category] as any
    return styles[variant] || null
  },

  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}

    // Font families
    Object.entries(fontFamilies).forEach(([key, value]) => {
      cssVars[`--font-family-${key}`] = value.join(', ')
    })

    // Font sizes
    Object.entries(fontSizes).forEach(([key, value]) => {
      cssVars[`--font-size-${key}`] = value
    })

    // Font weights
    Object.entries(fontWeights).forEach(([key, value]) => {
      cssVars[`--font-weight-${key}`] = value.toString()
    })

    // Line heights
    Object.entries(lineHeights).forEach(([key, value]) => {
      cssVars[`--line-height-${key}`] = value
    })

    // Letter spacing
    Object.entries(letterSpacing).forEach(([key, value]) => {
      cssVars[`--letter-spacing-${key}`] = value
    })

    return cssVars
  },

  // Apply typography style to element
  applyStyle: (element: HTMLElement, style: any) => {
    if (!element || !style) return

    Object.entries(style).forEach(([property, value]) => {
      if (typeof value === 'string' || typeof value === 'number') {
        element.style[property as any] = value.toString()
      }
    })
  }
} as const

// Accessibility considerations
export const accessibilityTypography = {
  // Minimum font sizes for readability
  minimumSizes: {
    mobile: '16px',   // iOS minimum for no zoom
    desktop: '14px'   // Generally readable
  },

  // Recommended line heights for accessibility
  recommendedLineHeights: {
    body: '1.5',      // WCAG recommendation
    heading: '1.3',   // Tighter for headings
    code: '1.4'       // Monospace adjustment
  },

  // High contrast adjustments
  highContrast: {
    fontWeight: {
      normal: fontWeights.medium,  // Heavier for better visibility
      medium: fontWeights.semibold,
      bold: fontWeights.extrabold
    },
    letterSpacing: {
      normal: letterSpacing.wide,  // More spacing for clarity
      wide: letterSpacing.wider,
      wider: letterSpacing.widest
    }
  },

  // Dyslexia-friendly adjustments
  dyslexiaFriendly: {
    fontFamily: [
      'OpenDyslexic',
      'Arial',
      'Verdana',
      'sans-serif'
    ],
    letterSpacing: letterSpacing.wide,
    lineHeight: lineHeights.relaxed,
    fontWeight: fontWeights.normal
  }
} as const

// Export types
export type FontFamily = keyof typeof fontFamilies
export type FontWeight = keyof typeof fontWeights
export type FontSize = keyof typeof fontSizes
export type LineHeight = keyof typeof lineHeights
export type LetterSpacing = keyof typeof letterSpacing
export type TypographyCategory = keyof typeof typographyScale
export type TypographyStyle = typeof typographyScale.body.medium

export default {
  fontFamilies,
  fontWeights,
  fontSizes,
  lineHeights,
  letterSpacing,
  typographyScale,
  responsiveTypography,
  typographyUtils,
  accessibilityTypography
}