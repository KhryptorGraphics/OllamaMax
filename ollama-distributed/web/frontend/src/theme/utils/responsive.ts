import { breakpoints, mediaQueries } from '../tokens';

/**
 * Responsive utility functions for CSS-in-JS
 */

// Create responsive styles with mobile-first approach
export const responsive = {
  xs: (styles: string) => `
    @media (min-width: ${breakpoints.xs}) {
      ${styles}
    }
  `,
  sm: (styles: string) => `
    @media (min-width: ${breakpoints.sm}) {
      ${styles}
    }
  `,
  md: (styles: string) => `
    @media (min-width: ${breakpoints.md}) {
      ${styles}
    }
  `,
  lg: (styles: string) => `
    @media (min-width: ${breakpoints.lg}) {
      ${styles}
    }
  `,
  xl: (styles: string) => `
    @media (min-width: ${breakpoints.xl}) {
      ${styles}
    }
  `,
  '2xl': (styles: string) => `
    @media (min-width: ${breakpoints['2xl']}) {
      ${styles}
    }
  `,
  
  // Max-width queries (desktop-first approach)
  maxXs: (styles: string) => `
    @media (max-width: ${breakpoints.xs}) {
      ${styles}
    }
  `,
  maxSm: (styles: string) => `
    @media (max-width: ${breakpoints.sm}) {
      ${styles}
    }
  `,
  maxMd: (styles: string) => `
    @media (max-width: ${breakpoints.md}) {
      ${styles}
    }
  `,
  maxLg: (styles: string) => `
    @media (max-width: ${breakpoints.lg}) {
      ${styles}
    }
  `,
  maxXl: (styles: string) => `
    @media (max-width: ${breakpoints.xl}) {
      ${styles}
    }
  `,
  
  // Between breakpoints
  between: (min: keyof typeof breakpoints, max: keyof typeof breakpoints) => (styles: string) => `
    @media (min-width: ${breakpoints[min]}) and (max-width: ${breakpoints[max]}) {
      ${styles}
    }
  `,
  
  // Device-specific queries
  mobile: (styles: string) => `
    @media (max-width: ${breakpoints.sm}) {
      ${styles}
    }
  `,
  tablet: (styles: string) => `
    @media (min-width: ${breakpoints.sm}) and (max-width: ${breakpoints.lg}) {
      ${styles}
    }
  `,
  desktop: (styles: string) => `
    @media (min-width: ${breakpoints.lg}) {
      ${styles}
    }
  `,
  
  // Touch and hover
  touch: (styles: string) => `
    @media (hover: none) and (pointer: coarse) {
      ${styles}
    }
  `,
  hover: (styles: string) => `
    @media (hover: hover) and (pointer: fine) {
      ${styles}
    }
  `,
  
  // Motion preferences
  motion: (styles: string) => `
    @media (prefers-reduced-motion: no-preference) {
      ${styles}
    }
  `,
  reduceMotion: (styles: string) => `
    @media (prefers-reduced-motion: reduce) {
      ${styles}
    }
  `,
  
  // Color scheme
  darkMode: (styles: string) => `
    @media (prefers-color-scheme: dark) {
      ${styles}
    }
  `,
  lightMode: (styles: string) => `
    @media (prefers-color-scheme: light) {
      ${styles}
    }
  `,
  
  // High resolution displays
  highDPI: (styles: string) => `
    @media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
      ${styles}
    }
  `
};

// Create responsive values helper
export const responsiveValue = <T>(values: {
  base?: T;
  xs?: T;
  sm?: T;
  md?: T;
  lg?: T;
  xl?: T;
  '2xl'?: T;
}) => {
  let css = '';
  
  if (values.base !== undefined) {
    css += `value: ${values.base};`;
  }
  
  if (values.xs !== undefined) {
    css += responsive.xs(`value: ${values.xs};`);
  }
  
  if (values.sm !== undefined) {
    css += responsive.sm(`value: ${values.sm};`);
  }
  
  if (values.md !== undefined) {
    css += responsive.md(`value: ${values.md};`);
  }
  
  if (values.lg !== undefined) {
    css += responsive.lg(`value: ${values.lg};`);
  }
  
  if (values.xl !== undefined) {
    css += responsive.xl(`value: ${values.xl};`);
  }
  
  if (values['2xl'] !== undefined) {
    css += responsive['2xl'](`value: ${values['2xl']};`);
  }
  
  return css;
};

// Grid system helpers
export const gridSystem = {
  container: (maxWidth?: keyof typeof breakpoints) => `
    width: 100%;
    margin-left: auto;
    margin-right: auto;
    padding-left: 1rem;
    padding-right: 1rem;
    
    ${maxWidth ? `max-width: ${breakpoints[maxWidth]};` : ''}
    
    ${responsive.sm('padding-left: 1.5rem; padding-right: 1.5rem;')}
    ${responsive.lg('padding-left: 2rem; padding-right: 2rem;')}
  `,
  
  row: `
    display: flex;
    flex-wrap: wrap;
    margin-left: -0.5rem;
    margin-right: -0.5rem;
  `,
  
  col: (span?: number, total: number = 12) => `
    flex: ${span ? `0 0 ${(span / total) * 100}%` : '1 1 0%'};
    padding-left: 0.5rem;
    padding-right: 0.5rem;
  `
};

// Common responsive patterns
export const patterns = {
  flexCenter: `
    display: flex;
    align-items: center;
    justify-content: center;
  `,
  
  absoluteCenter: `
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  `,
  
  aspectRatio: (ratio: string) => `
    position: relative;
    width: 100%;
    height: 0;
    padding-bottom: ${ratio};
    
    > * {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
    }
  `,
  
  stack: (gap: string = '1rem') => `
    display: flex;
    flex-direction: column;
    gap: ${gap};
  `,
  
  inline: (gap: string = '1rem') => `
    display: flex;
    align-items: center;
    gap: ${gap};
    flex-wrap: wrap;
  `
};

export type ResponsiveValues<T> = {
  base?: T;
  xs?: T;
  sm?: T;
  md?: T;
  lg?: T;
  xl?: T;
  '2xl'?: T;
};