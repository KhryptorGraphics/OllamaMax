import {
  colors,
  lightSemanticColors,
  darkSemanticColors,
  typography,
  spacing,
  breakpoints,
  mediaQueries,
  shadows,
  darkShadows,
  elevation,
  darkElevation,
  animation,
  radii,
  zIndex,
  opacity,
  blur
} from './tokens';

// Define the complete theme structure
export interface Theme {
  mode: 'light' | 'dark';
  colors: typeof lightSemanticColors;
  typography: typeof typography;
  spacing: typeof spacing;
  breakpoints: typeof breakpoints;
  mediaQueries: typeof mediaQueries;
  shadows: typeof shadows;
  elevation: typeof elevation;
  animation: typeof animation;
  radii: typeof radii;
  zIndex: typeof zIndex;
  opacity: typeof opacity;
  blur: typeof blur;
  raw: {
    colors: typeof colors;
  };
}

// Create the theme object
export const theme = {
  // Raw color palette (for advanced usage)
  raw: {
    colors
  },

  // Design tokens
  typography,
  spacing,
  breakpoints,
  mediaQueries,
  animation,
  radii,
  zIndex,
  opacity,
  blur,

  // Theme variants
  colors: {
    light: lightSemanticColors,
    dark: darkSemanticColors
  },

  shadows: {
    light: shadows,
    dark: darkShadows
  },

  elevation: {
    light: elevation,
    dark: darkElevation
  }
} as const;

// Helper function to get theme-aware values
export const getThemeValue = (
  lightValue: string,
  darkValue: string,
  currentTheme: 'light' | 'dark'
): string => {
  return currentTheme === 'dark' ? darkValue : lightValue;
};

// CSS-in-JS helpers
export const css = {
  // Responsive helpers
  mobile: (content: string) => `
    @media (max-width: ${breakpoints.sm}) {
      ${content}
    }
  `,
  
  tablet: (content: string) => `
    @media (min-width: ${breakpoints.sm}) and (max-width: ${breakpoints.lg}) {
      ${content}
    }
  `,
  
  desktop: (content: string) => `
    @media (min-width: ${breakpoints.lg}) {
      ${content}
    }
  `,

  // Motion helpers
  motion: (content: string) => `
    @media (prefers-reduced-motion: no-preference) {
      ${content}
    }
  `,

  noMotion: (content: string) => `
    @media (prefers-reduced-motion: reduce) {
      ${content}
    }
  `,

  // Touch device helpers
  touch: (content: string) => `
    @media (hover: none) and (pointer: coarse) {
      ${content}
    }
  `,

  hover: (content: string) => `
    @media (hover: hover) and (pointer: fine) {
      ${content}
    }
  `,

  // Common focus styles
  focus: (color: string = 'var(--color-border-focus)') => `
    outline: 2px solid ${color};
    outline-offset: 2px;
  `,

  // Common button reset
  buttonReset: `
    border: none;
    background: none;
    padding: 0;
    margin: 0;
    font: inherit;
    cursor: pointer;
    outline: none;
  `,

  // Visually hidden (for accessibility)
  visuallyHidden: `
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  `,

  // Truncate text
  truncate: `
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  `,

  // Multi-line truncate
  truncateLines: (lines: number) => `
    display: -webkit-box;
    -webkit-line-clamp: ${lines};
    -webkit-box-orient: vertical;
    overflow: hidden;
  `
};

export type ThemeType = typeof theme;