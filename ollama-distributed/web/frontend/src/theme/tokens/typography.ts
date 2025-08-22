export const typography = {
  fontFamily: {
    sans: [
      'Inter',
      '-apple-system',
      'BlinkMacSystemFont',
      'Segoe UI',
      'Roboto',
      'Oxygen',
      'Ubuntu',
      'Cantarell',
      'Open Sans',
      'Helvetica Neue',
      'sans-serif'
    ],
    mono: [
      'JetBrains Mono',
      'Monaco',
      'Menlo',
      'Ubuntu Mono',
      'monospace'
    ],
    display: [
      'Cal Sans',
      'Inter',
      'system-ui',
      'sans-serif'
    ]
  },

  fontSize: {
    xs: ['0.75rem', { lineHeight: '1rem' }],      // 12px
    sm: ['0.875rem', { lineHeight: '1.25rem' }],  // 14px
    base: ['1rem', { lineHeight: '1.5rem' }],     // 16px
    lg: ['1.125rem', { lineHeight: '1.75rem' }],  // 18px
    xl: ['1.25rem', { lineHeight: '1.75rem' }],   // 20px
    '2xl': ['1.5rem', { lineHeight: '2rem' }],    // 24px
    '3xl': ['1.875rem', { lineHeight: '2.25rem' }], // 30px
    '4xl': ['2.25rem', { lineHeight: '2.5rem' }], // 36px
    '5xl': ['3rem', { lineHeight: '1' }],         // 48px
    '6xl': ['3.75rem', { lineHeight: '1' }],      // 60px
    '7xl': ['4.5rem', { lineHeight: '1' }],       // 72px
    '8xl': ['6rem', { lineHeight: '1' }],         // 96px
    '9xl': ['8rem', { lineHeight: '1' }]          // 128px
  },

  fontWeight: {
    thin: '100',
    extralight: '200',
    light: '300',
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
    extrabold: '800',
    black: '900'
  },

  letterSpacing: {
    tighter: '-0.05em',
    tight: '-0.025em',
    normal: '0em',
    wide: '0.025em',
    wider: '0.05em',
    widest: '0.1em'
  },

  lineHeight: {
    none: '1',
    tight: '1.25',
    snug: '1.375',
    normal: '1.5',
    relaxed: '1.625',
    loose: '2'
  }
} as const;

// Semantic typography mappings
export const semanticTypography = {
  heading: {
    h1: {
      fontSize: typography.fontSize['4xl'],
      fontWeight: typography.fontWeight.bold,
      lineHeight: typography.lineHeight.tight,
      letterSpacing: typography.letterSpacing.tight,
      fontFamily: typography.fontFamily.display
    },
    h2: {
      fontSize: typography.fontSize['3xl'],
      fontWeight: typography.fontWeight.bold,
      lineHeight: typography.lineHeight.tight,
      letterSpacing: typography.letterSpacing.tight,
      fontFamily: typography.fontFamily.display
    },
    h3: {
      fontSize: typography.fontSize['2xl'],
      fontWeight: typography.fontWeight.semibold,
      lineHeight: typography.lineHeight.tight,
      letterSpacing: typography.letterSpacing.tight,
      fontFamily: typography.fontFamily.display
    },
    h4: {
      fontSize: typography.fontSize.xl,
      fontWeight: typography.fontWeight.semibold,
      lineHeight: typography.lineHeight.snug,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    },
    h5: {
      fontSize: typography.fontSize.lg,
      fontWeight: typography.fontWeight.semibold,
      lineHeight: typography.lineHeight.snug,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    },
    h6: {
      fontSize: typography.fontSize.base,
      fontWeight: typography.fontWeight.semibold,
      lineHeight: typography.lineHeight.normal,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    }
  },

  body: {
    large: {
      fontSize: typography.fontSize.lg,
      fontWeight: typography.fontWeight.normal,
      lineHeight: typography.lineHeight.relaxed,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    },
    base: {
      fontSize: typography.fontSize.base,
      fontWeight: typography.fontWeight.normal,
      lineHeight: typography.lineHeight.normal,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    },
    small: {
      fontSize: typography.fontSize.sm,
      fontWeight: typography.fontWeight.normal,
      lineHeight: typography.lineHeight.normal,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.sans
    },
    xs: {
      fontSize: typography.fontSize.xs,
      fontWeight: typography.fontWeight.normal,
      lineHeight: typography.lineHeight.normal,
      letterSpacing: typography.letterSpacing.wide,
      fontFamily: typography.fontFamily.sans
    }
  },

  code: {
    inline: {
      fontSize: typography.fontSize.sm,
      fontWeight: typography.fontWeight.medium,
      lineHeight: typography.lineHeight.normal,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.mono
    },
    block: {
      fontSize: typography.fontSize.sm,
      fontWeight: typography.fontWeight.normal,
      lineHeight: typography.lineHeight.relaxed,
      letterSpacing: typography.letterSpacing.normal,
      fontFamily: typography.fontFamily.mono
    }
  },

  caption: {
    fontSize: typography.fontSize.xs,
    fontWeight: typography.fontWeight.medium,
    lineHeight: typography.lineHeight.tight,
    letterSpacing: typography.letterSpacing.wide,
    fontFamily: typography.fontFamily.sans
  },

  overline: {
    fontSize: typography.fontSize.xs,
    fontWeight: typography.fontWeight.semibold,
    lineHeight: typography.lineHeight.none,
    letterSpacing: typography.letterSpacing.widest,
    fontFamily: typography.fontFamily.sans,
    textTransform: 'uppercase' as const
  }
} as const;

export type Typography = typeof typography;
export type SemanticTypography = typeof semanticTypography;