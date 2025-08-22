import { css } from 'styled-components';
import { Theme } from '../theme';
import { responsive, ResponsiveValues } from './responsive';

/**
 * Styled-components utilities for theming
 */

// Theme-aware CSS helper
export const themeValue = (path: string) => ({ theme }: { theme: Theme }) => {
  const keys = path.split('.');
  let value: any = theme;
  
  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key];
    } else {
      console.warn(`Theme path "${path}" not found`);
      return undefined;
    }
  }
  
  return value;
};

// Responsive prop helper
export const responsiveProp = <T>(
  values: ResponsiveValues<T> | T,
  transform?: (value: T) => string
) => {
  if (typeof values !== 'object' || values === null) {
    const transformedValue = transform ? transform(values as T) : String(values);
    return css`${transformedValue}`;
  }
  
  const responsiveValues = values as ResponsiveValues<T>;
  
  return css`
    ${responsiveValues.base && css`
      ${transform ? transform(responsiveValues.base) : String(responsiveValues.base)}
    `}
    
    ${responsiveValues.xs && css`
      ${responsive.xs(transform ? transform(responsiveValues.xs) : String(responsiveValues.xs))}
    `}
    
    ${responsiveValues.sm && css`
      ${responsive.sm(transform ? transform(responsiveValues.sm) : String(responsiveValues.sm))}
    `}
    
    ${responsiveValues.md && css`
      ${responsive.md(transform ? transform(responsiveValues.md) : String(responsiveValues.md))}
    `}
    
    ${responsiveValues.lg && css`
      ${responsive.lg(transform ? transform(responsiveValues.lg) : String(responsiveValues.lg))}
    `}
    
    ${responsiveValues.xl && css`
      ${responsive.xl(transform ? transform(responsiveValues.xl) : String(responsiveValues.xl))}
    `}
    
    ${responsiveValues['2xl'] && css`
      ${responsive['2xl'](transform ? transform(responsiveValues['2xl']) : String(responsiveValues['2xl']))}
    `}
  `;
};

// Spacing helper
export const space = (value: string | number) => ({ theme }: { theme: Theme }) => {
  if (typeof value === 'number') {
    return theme.spacing[value] || `${value}px`;
  }
  
  if (typeof value === 'string' && value in theme.spacing) {
    return theme.spacing[value as keyof typeof theme.spacing];
  }
  
  return value;
};

// Color helper with theme awareness
export const color = (path: string, fallback?: string) => ({ theme }: { theme: Theme }) => {
  const keys = path.split('.');
  let value: any = theme.colors;
  
  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key];
    } else {
      // Try raw colors
      value = theme.raw.colors;
      for (const rawKey of keys) {
        if (value && typeof value === 'object' && rawKey in value) {
          value = value[rawKey];
        } else {
          return fallback || 'transparent';
        }
      }
      break;
    }
  }
  
  return value || fallback || 'transparent';
};

// Shadow helper
export const shadow = (level: keyof Theme['shadows']) => ({ theme }: { theme: Theme }) => {
  return theme.shadows[level];
};

// Border radius helper
export const radius = (size: keyof Theme['radii']) => ({ theme }: { theme: Theme }) => {
  return theme.radii[size];
};

// Typography helper
export const typography = (variant: string) => ({ theme }: { theme: Theme }) => {
  const keys = variant.split('.');
  let value: any = theme.typography;
  
  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key];
    } else {
      console.warn(`Typography variant "${variant}" not found`);
      return {};
    }
  }
  
  if (Array.isArray(value)) {
    // Handle fontSize arrays [size, { lineHeight }]
    return css`
      font-size: ${value[0]};
      ${value[1] && value[1].lineHeight && css`line-height: ${value[1].lineHeight};`}
    `;
  }
  
  if (typeof value === 'object') {
    return css`
      ${Object.entries(value).map(([prop, val]) => {
        const cssProp = prop.replace(/[A-Z]/g, letter => `-${letter.toLowerCase()}`);
        return `${cssProp}: ${Array.isArray(val) ? val.join(', ') : val};`;
      }).join('')}
    `;
  }
  
  return value;
};

// Animation helper
export const animation = (duration: keyof Theme['animation']['duration'], easing: keyof Theme['animation']['easing']) => 
  ({ theme }: { theme: Theme }) => css`
    transition-duration: ${theme.animation.duration[duration]};
    transition-timing-function: ${theme.animation.easing[easing]};
  `;

// Focus ring helper
export const focusRing = (color?: string) => ({ theme }: { theme: Theme }) => css`
  outline: 2px solid ${color || theme.colors.border.focus};
  outline-offset: 2px;
`;

// Truncate text helper
export const truncate = (lines?: number) => css`
  ${lines && lines > 1 ? css`
    display: -webkit-box;
    -webkit-line-clamp: ${lines};
    -webkit-box-orient: vertical;
    overflow: hidden;
  ` : css`
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  `}
`;

// Visually hidden helper
export const visuallyHidden = css`
  position: absolute !important;
  width: 1px !important;
  height: 1px !important;
  padding: 0 !important;
  margin: -1px !important;
  overflow: hidden !important;
  clip: rect(0, 0, 0, 0) !important;
  white-space: nowrap !important;
  border: 0 !important;
`;

// Button reset helper
export const buttonReset = css`
  border: none;
  background: none;
  padding: 0;
  margin: 0;
  font: inherit;
  cursor: pointer;
  outline: none;
  
  &:focus-visible {
    ${focusRing()}
  }
`;

// List reset helper
export const listReset = css`
  list-style: none;
  padding: 0;
  margin: 0;
`;

// Input reset helper
export const inputReset = css`
  border: none;
  background: none;
  padding: 0;
  margin: 0;
  font: inherit;
  outline: none;
  
  &:focus {
    outline: none;
  }
  
  &:focus-visible {
    ${focusRing()}
  }
`;

// Motion safe helper
export const motionSafe = (styles: ReturnType<typeof css>) => css`
  @media (prefers-reduced-motion: no-preference) {
    ${styles}
  }
`;

// Motion reduce helper
export const motionReduce = (styles: ReturnType<typeof css>) => css`
  @media (prefers-reduced-motion: reduce) {
    ${styles}
  }
`;

export type StyledProps<T = {}> = T & {
  theme: Theme;
};