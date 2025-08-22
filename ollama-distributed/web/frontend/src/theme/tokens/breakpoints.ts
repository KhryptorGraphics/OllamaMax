export const breakpoints = {
  xs: '475px',      // Mobile small
  sm: '640px',      // Mobile large
  md: '768px',      // Tablet
  lg: '1024px',     // Desktop small
  xl: '1280px',     // Desktop large
  '2xl': '1536px'   // Desktop extra large
} as const;

// Media queries for responsive design
export const mediaQueries = {
  xs: `@media (min-width: ${breakpoints.xs})`,
  sm: `@media (min-width: ${breakpoints.sm})`,
  md: `@media (min-width: ${breakpoints.md})`,
  lg: `@media (min-width: ${breakpoints.lg})`,
  xl: `@media (min-width: ${breakpoints.xl})`,
  '2xl': `@media (min-width: ${breakpoints['2xl']})`,

  // Max-width queries (for mobile-first approach)
  'max-xs': `@media (max-width: ${breakpoints.xs})`,
  'max-sm': `@media (max-width: ${breakpoints.sm})`,
  'max-md': `@media (max-width: ${breakpoints.md})`,
  'max-lg': `@media (max-width: ${breakpoints.lg})`,
  'max-xl': `@media (max-width: ${breakpoints.xl})`,

  // Between breakpoints
  'xs-sm': `@media (min-width: ${breakpoints.xs}) and (max-width: ${breakpoints.sm})`,
  'sm-md': `@media (min-width: ${breakpoints.sm}) and (max-width: ${breakpoints.md})`,
  'md-lg': `@media (min-width: ${breakpoints.md}) and (max-width: ${breakpoints.lg})`,
  'lg-xl': `@media (min-width: ${breakpoints.lg}) and (max-width: ${breakpoints.xl})`,

  // Utility queries
  touch: '@media (hover: none) and (pointer: coarse)',
  hover: '@media (hover: hover) and (pointer: fine)',
  'high-res': '@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi)',
  
  // Motion preferences
  motion: '@media (prefers-reduced-motion: no-preference)',
  'no-motion': '@media (prefers-reduced-motion: reduce)',
  
  // Color scheme preferences
  'dark-mode': '@media (prefers-color-scheme: dark)',
  'light-mode': '@media (prefers-color-scheme: light)',
  
  // Contrast preferences
  'high-contrast': '@media (prefers-contrast: high)',
  'low-contrast': '@media (prefers-contrast: low)'
} as const;

// Container max-widths for each breakpoint
export const containerSizes = {
  xs: '100%',
  sm: '640px',
  md: '768px',
  lg: '1024px',
  xl: '1280px',
  '2xl': '1536px'
} as const;

// Grid system configuration
export const grid = {
  columns: 12,
  gutter: {
    xs: '16px',
    sm: '20px',
    md: '24px',
    lg: '32px',
    xl: '40px'
  },
  margin: {
    xs: '16px',
    sm: '24px',
    md: '32px',
    lg: '40px',
    xl: '48px'
  }
} as const;

export type Breakpoints = typeof breakpoints;
export type MediaQueries = typeof mediaQueries;
export type ContainerSizes = typeof containerSizes;
export type Grid = typeof grid;