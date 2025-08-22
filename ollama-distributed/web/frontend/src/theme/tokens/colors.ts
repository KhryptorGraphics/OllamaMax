export const colors = {
  // Primary brand colors - Blue spectrum
  primary: {
    50: '#eff6ff',   // Lightest
    100: '#dbeafe',
    200: '#bfdbfe',
    300: '#93c5fd',
    400: '#60a5fa',
    500: '#3b82f6',  // Base
    600: '#2563eb',
    700: '#1d4ed8',
    800: '#1e40af',
    900: '#1e3a8a',  // Darkest
    950: '#172554'
  },

  // Secondary colors - Emerald spectrum
  secondary: {
    50: '#ecfdf5',
    100: '#d1fae5',
    200: '#a7f3d0',
    300: '#6ee7b7',
    400: '#34d399',
    500: '#10b981',
    600: '#059669',
    700: '#047857',
    800: '#065f46',
    900: '#064e3b',
    950: '#022c22'
  },

  // Neutral colors - Slate spectrum
  neutral: {
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b',
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
    950: '#020617'
  },

  // Status colors
  success: {
    50: '#f0fdf4',
    100: '#dcfce7',
    200: '#bbf7d0',
    300: '#86efac',
    400: '#4ade80',
    500: '#22c55e',
    600: '#16a34a',
    700: '#15803d',
    800: '#166534',
    900: '#14532d',
    950: '#052e16'
  },

  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444',
    600: '#dc2626',
    700: '#b91c1c',
    800: '#991b1b',
    900: '#7f1d1d',
    950: '#450a0a'
  },

  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b',
    600: '#d97706',
    700: '#b45309',
    800: '#92400e',
    900: '#78350f',
    950: '#451a03'
  },

  info: {
    50: '#eff6ff',
    100: '#dbeafe',
    200: '#bfdbfe',
    300: '#93c5fd',
    400: '#60a5fa',
    500: '#3b82f6',
    600: '#2563eb',
    700: '#1d4ed8',
    800: '#1e40af',
    900: '#1e3a8a',
    950: '#172554'
  }
} as const;

// Semantic color mappings for light mode
export const lightSemanticColors = {
  background: {
    primary: colors.neutral[50],      // #f8fafc
    secondary: colors.neutral[100],   // #f1f5f9
    tertiary: colors.neutral[200],    // #e2e8f0
    inverse: colors.neutral[900],     // #0f172a
    elevated: '#ffffff',
    overlay: 'rgba(0, 0, 0, 0.8)',
    subtle: colors.neutral[50]
  },
  
  text: {
    primary: colors.neutral[900],     // #0f172a
    secondary: colors.neutral[600],   // #475569
    tertiary: colors.neutral[500],    // #64748b
    inverse: colors.neutral[50],      // #f8fafc
    muted: colors.neutral[400],       // #94a3b8
    disabled: colors.neutral[300]     // #cbd5e1
  },

  border: {
    default: colors.neutral[200],     // #e2e8f0
    muted: colors.neutral[100],       // #f1f5f9
    strong: colors.neutral[300],      // #cbd5e1
    focus: colors.primary[500],       // #3b82f6
    error: colors.error[500],         // #ef4444
    success: colors.success[500],     // #22c55e
    warning: colors.warning[500]      // #f59e0b
  },

  surface: {
    elevated: '#ffffff',
    overlay: 'rgba(0, 0, 0, 0.8)',
    subtle: colors.neutral[50],
    muted: colors.neutral[100],
    strong: colors.neutral[200]
  },

  button: {
    primary: {
      background: colors.primary[500],
      backgroundHover: colors.primary[600],
      backgroundActive: colors.primary[700],
      text: '#ffffff',
      border: colors.primary[500]
    },
    secondary: {
      background: colors.neutral[100],
      backgroundHover: colors.neutral[200],
      backgroundActive: colors.neutral[300],
      text: colors.neutral[700],
      border: colors.neutral[200]
    },
    ghost: {
      background: 'transparent',
      backgroundHover: colors.neutral[100],
      backgroundActive: colors.neutral[200],
      text: colors.neutral[700],
      border: 'transparent'
    },
    destructive: {
      background: colors.error[500],
      backgroundHover: colors.error[600],
      backgroundActive: colors.error[700],
      text: '#ffffff',
      border: colors.error[500]
    }
  }
} as const;

// Semantic color mappings for dark mode
export const darkSemanticColors = {
  background: {
    primary: colors.neutral[950],     // #020617
    secondary: colors.neutral[900],   // #0f172a
    tertiary: colors.neutral[800],    // #1e293b
    inverse: colors.neutral[50],      // #f8fafc
    elevated: colors.neutral[900],
    overlay: 'rgba(0, 0, 0, 0.9)',
    subtle: colors.neutral[950]
  },

  text: {
    primary: colors.neutral[50],      // #f8fafc
    secondary: colors.neutral[300],   // #cbd5e1
    tertiary: colors.neutral[400],    // #94a3b8
    inverse: colors.neutral[900],     // #0f172a
    muted: colors.neutral[500],       // #64748b
    disabled: colors.neutral[600]     // #475569
  },

  border: {
    default: colors.neutral[700],     // #334155
    muted: colors.neutral[800],       // #1e293b
    strong: colors.neutral[600],      // #475569
    focus: colors.primary[400],       // #60a5fa
    error: colors.error[400],         // #f87171
    success: colors.success[400],     // #4ade80
    warning: colors.warning[400]      // #fbbf24
  },

  surface: {
    elevated: colors.neutral[900],
    overlay: 'rgba(0, 0, 0, 0.9)',
    subtle: colors.neutral[950],
    muted: colors.neutral[900],
    strong: colors.neutral[800]
  },

  button: {
    primary: {
      background: colors.primary[600],
      backgroundHover: colors.primary[500],
      backgroundActive: colors.primary[700],
      text: '#ffffff',
      border: colors.primary[600]
    },
    secondary: {
      background: colors.neutral[800],
      backgroundHover: colors.neutral[700],
      backgroundActive: colors.neutral[600],
      text: colors.neutral[200],
      border: colors.neutral[700]
    },
    ghost: {
      background: 'transparent',
      backgroundHover: colors.neutral[800],
      backgroundActive: colors.neutral[700],
      text: colors.neutral[200],
      border: 'transparent'
    },
    destructive: {
      background: colors.error[600],
      backgroundHover: colors.error[500],
      backgroundActive: colors.error[700],
      text: '#ffffff',
      border: colors.error[600]
    }
  }
} as const;

export type ColorScale = typeof colors.primary;
export type SemanticColors = typeof lightSemanticColors;