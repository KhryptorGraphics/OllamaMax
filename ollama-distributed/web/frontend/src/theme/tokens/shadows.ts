export const shadows = {
  // Basic shadows
  xs: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  sm: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.25)',
  
  // Special shadows
  inner: 'inset 0 2px 4px 0 rgb(0 0 0 / 0.05)',
  none: '0 0 #0000',
  
  // Colored shadows for brand elements
  primary: '0 4px 14px 0 rgb(59 130 246 / 0.15)',
  'primary-lg': '0 10px 25px -3px rgb(59 130 246 / 0.2)',
  
  secondary: '0 4px 14px 0 rgb(16 185 129 / 0.15)',
  'secondary-lg': '0 10px 25px -3px rgb(16 185 129 / 0.2)',
  
  error: '0 4px 14px 0 rgb(239 68 68 / 0.15)',
  'error-lg': '0 10px 25px -3px rgb(239 68 68 / 0.2)',
  
  warning: '0 4px 14px 0 rgb(245 158 11 / 0.15)',
  'warning-lg': '0 10px 25px -3px rgb(245 158 11 / 0.2)',
  
  success: '0 4px 14px 0 rgb(34 197 94 / 0.15)',
  'success-lg': '0 10px 25px -3px rgb(34 197 94 / 0.2)'
} as const;

// Dark mode shadow variants
export const darkShadows = {
  xs: '0 1px 2px 0 rgb(0 0 0 / 0.3)',
  sm: '0 1px 3px 0 rgb(0 0 0 / 0.4), 0 1px 2px -1px rgb(0 0 0 / 0.4)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.4), 0 2px 4px -2px rgb(0 0 0 / 0.4)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.4)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.4), 0 8px 10px -6px rgb(0 0 0 / 0.4)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.6)',
  
  inner: 'inset 0 2px 4px 0 rgb(0 0 0 / 0.3)',
  none: '0 0 #0000',
  
  // Colored shadows for dark mode
  primary: '0 4px 14px 0 rgb(96 165 250 / 0.3)',
  'primary-lg': '0 10px 25px -3px rgb(96 165 250 / 0.4)',
  
  secondary: '0 4px 14px 0 rgb(52 211 153 / 0.3)',
  'secondary-lg': '0 10px 25px -3px rgb(52 211 153 / 0.4)',
  
  error: '0 4px 14px 0 rgb(248 113 113 / 0.3)',
  'error-lg': '0 10px 25px -3px rgb(248 113 113 / 0.4)',
  
  warning: '0 4px 14px 0 rgb(251 191 36 / 0.3)',
  'warning-lg': '0 10px 25px -3px rgb(251 191 36 / 0.4)',
  
  success: '0 4px 14px 0 rgb(74 222 128 / 0.3)',
  'success-lg': '0 10px 25px -3px rgb(74 222 128 / 0.4)'
} as const;

// Elevation levels for UI hierarchy
export const elevation = {
  0: shadows.none,         // Flat surfaces
  1: shadows.sm,           // Cards, buttons
  2: shadows.md,           // Raised elements
  3: shadows.lg,           // Modals, dropdowns
  4: shadows.xl,           // Navigation, overlays
  5: shadows['2xl']        // Floating elements
} as const;

export const darkElevation = {
  0: darkShadows.none,
  1: darkShadows.sm,
  2: darkShadows.md,
  3: darkShadows.lg,
  4: darkShadows.xl,
  5: darkShadows['2xl']
} as const;

export type Shadows = typeof shadows;
export type DarkShadows = typeof darkShadows;
export type Elevation = typeof elevation;