/**
 * Design Tokens Index
 * Centralized export of all design tokens for consistent theming
 */

export * from './colors'
export * from './typography'
export * from './spacing'
export * from './shadows'
export * from './radius'
export * from './breakpoints'
export * from './transitions'

// Re-export utilities for convenience
export { colorUtils } from './colors'
export { typographyUtils } from './typography'
export { spacingUtils } from './spacing'
export { shadowUtils } from './shadows'
export { radiusUtils } from './radius'
export { breakpointUtils } from './breakpoints'
export { transitionUtils } from './transitions'

// Combined token utilities
export const designTokens = {
  colors: () => import('./colors').then(m => m.colors),
  typography: () => import('./typography').then(m => m.typographyScale),
  spacing: () => import('./spacing').then(m => m.spacing),
  shadows: () => import('./shadows').then(m => m.shadows),
  radius: () => import('./radius').then(m => m.radius),
  breakpoints: () => import('./breakpoints').then(m => m.breakpoints),
  transitions: () => import('./transitions').then(m => m.transitions)
} as const

// CSS custom properties generator
export const generateAllCSSVariables = () => {
  return Promise.all([
    import('./colors').then(m => m.colorUtils.generateCSSVariables('light')),
    import('./typography').then(m => m.typographyUtils.generateCSSVariables()),
    import('./spacing').then(m => m.spacingUtils.generateCSSVariables()),
    import('./shadows').then(m => m.shadowUtils.generateCSSVariables()),
    import('./radius').then(m => m.radiusUtils.generateCSSVariables()),
    import('./breakpoints').then(m => m.breakpointUtils.generateCSSVariables()),
    import('./transitions').then(m => m.transitionUtils.generateCSSVariables())
  ]).then(([colors, typography, spacing, shadows, radius, breakpoints, transitions]) => ({
    ...colors,
    ...typography,
    ...spacing,
    ...shadows,
    ...radius,
    ...breakpoints,
    ...transitions
  }))
}

export default {
  designTokens,
  generateAllCSSVariables
}