import { createGlobalStyle } from 'styled-components';
import { Theme } from './theme';

export const GlobalStyles = createGlobalStyle<{ theme: Theme }>`
  /* CSS Reset and Base Styles */
  *,
  *::before,
  *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  /* HTML and Body */
  html {
    font-size: 16px;
    scroll-behavior: smooth;
    -webkit-text-size-adjust: 100%;
    -moz-text-size-adjust: 100%;
    text-size-adjust: 100%;
  }

  @media (prefers-reduced-motion: reduce) {
    html {
      scroll-behavior: auto;
    }
  }

  body {
    font-family: ${({ theme }) => theme.typography.fontFamily.sans.join(', ')};
    font-size: ${({ theme }) => theme.typography.fontSize.base[0]};
    line-height: ${({ theme }) => theme.typography.fontSize.base[1].lineHeight};
    font-weight: ${({ theme }) => theme.typography.fontWeight.normal};
    color: ${({ theme }) => theme.colors.text.primary};
    background-color: ${({ theme }) => theme.colors.background.primary};
    
    /* Smooth font rendering */
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-rendering: optimizeLegibility;
    
    /* Prevent horizontal scroll */
    overflow-x: hidden;
    
    /* Improve text selection */
    -webkit-text-select: text;
    -moz-user-select: text;
    -ms-user-select: text;
    user-select: text;
  }

  /* Focus styles */
  *:focus {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  *:focus:not(:focus-visible) {
    outline: none;
  }

  *:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  /* Typography Elements */
  h1, h2, h3, h4, h5, h6 {
    font-family: ${({ theme }) => theme.typography.fontFamily.display.join(', ')};
    color: ${({ theme }) => theme.colors.text.primary};
    margin: 0;
    font-weight: ${({ theme }) => theme.typography.fontWeight.semibold};
  }

  h1 {
    font-size: ${({ theme }) => theme.typography.fontSize['4xl'][0]};
    line-height: ${({ theme }) => theme.typography.fontSize['4xl'][1].lineHeight};
    font-weight: ${({ theme }) => theme.typography.fontWeight.bold};
  }

  h2 {
    font-size: ${({ theme }) => theme.typography.fontSize['3xl'][0]};
    line-height: ${({ theme }) => theme.typography.fontSize['3xl'][1].lineHeight};
    font-weight: ${({ theme }) => theme.typography.fontWeight.bold};
  }

  h3 {
    font-size: ${({ theme }) => theme.typography.fontSize['2xl'][0]};
    line-height: ${({ theme }) => theme.typography.fontSize['2xl'][1].lineHeight};
  }

  h4 {
    font-size: ${({ theme }) => theme.typography.fontSize.xl[0]};
    line-height: ${({ theme }) => theme.typography.fontSize.xl[1].lineHeight};
  }

  h5 {
    font-size: ${({ theme }) => theme.typography.fontSize.lg[0]};
    line-height: ${({ theme }) => theme.typography.fontSize.lg[1].lineHeight};
  }

  h6 {
    font-size: ${({ theme }) => theme.typography.fontSize.base[0]};
    line-height: ${({ theme }) => theme.typography.fontSize.base[1].lineHeight};
  }

  p {
    margin: 0;
    color: ${({ theme }) => theme.colors.text.secondary};
  }

  /* Links */
  a {
    color: ${({ theme }) => theme.raw.colors.primary[600]};
    text-decoration: none;
    transition: color ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  }

  a:hover {
    color: ${({ theme }) => theme.raw.colors.primary[700]};
    text-decoration: underline;
  }

  [data-theme="dark"] a {
    color: ${({ theme }) => theme.raw.colors.primary[400]};
  }

  [data-theme="dark"] a:hover {
    color: ${({ theme }) => theme.raw.colors.primary[300]};
  }

  /* Code elements */
  code {
    font-family: ${({ theme }) => theme.typography.fontFamily.mono.join(', ')};
    font-size: ${({ theme }) => theme.typography.fontSize.sm[0]};
    background-color: ${({ theme }) => theme.colors.surface.muted};
    padding: ${({ theme }) => theme.spacing[1]} ${({ theme }) => theme.spacing[2]};
    border-radius: ${({ theme }) => theme.radii.sm};
    border: 1px solid ${({ theme }) => theme.colors.border.muted};
  }

  pre {
    font-family: ${({ theme }) => theme.typography.fontFamily.mono.join(', ')};
    font-size: ${({ theme }) => theme.typography.fontSize.sm[0]};
    background-color: ${({ theme }) => theme.colors.surface.muted};
    padding: ${({ theme }) => theme.spacing[4]};
    border-radius: ${({ theme }) => theme.radii.lg};
    border: 1px solid ${({ theme }) => theme.colors.border.muted};
    overflow-x: auto;
    line-height: ${({ theme }) => theme.typography.lineHeight.relaxed};
  }

  pre code {
    background: none;
    padding: 0;
    border: none;
    border-radius: 0;
  }

  /* Form elements */
  input,
  textarea,
  select,
  button {
    font-family: inherit;
    font-size: inherit;
  }

  button {
    cursor: pointer;
    border: none;
    background: none;
    padding: 0;
    margin: 0;
  }

  button:disabled {
    cursor: not-allowed;
    opacity: ${({ theme }) => theme.opacity[50]};
  }

  /* Images and media */
  img,
  svg,
  video,
  canvas,
  audio,
  iframe,
  embed,
  object {
    display: block;
    max-width: 100%;
    height: auto;
  }

  /* Tables */
  table {
    border-collapse: collapse;
    border-spacing: 0;
    width: 100%;
  }

  th,
  td {
    text-align: left;
    vertical-align: top;
  }

  /* Lists */
  ul,
  ol {
    list-style: none;
  }

  /* Utility classes */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .truncate {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  /* Custom scrollbar */
  ::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  ::-webkit-scrollbar-track {
    background: ${({ theme }) => theme.colors.surface.muted};
  }

  ::-webkit-scrollbar-thumb {
    background: ${({ theme }) => theme.colors.border.default};
    border-radius: ${({ theme }) => theme.radii.full};
  }

  ::-webkit-scrollbar-thumb:hover {
    background: ${({ theme }) => theme.colors.border.strong};
  }

  /* Selection styles */
  ::selection {
    background-color: ${({ theme }) => theme.raw.colors.primary[200]};
    color: ${({ theme }) => theme.raw.colors.primary[900]};
  }

  [data-theme="dark"] ::selection {
    background-color: ${({ theme }) => theme.raw.colors.primary[800]};
    color: ${({ theme }) => theme.raw.colors.primary[100]};
  }

  /* Reduced motion preferences */
  @media (prefers-reduced-motion: reduce) {
    *,
    *::before,
    *::after {
      animation-duration: 0.01ms !important;
      animation-iteration-count: 1 !important;
      transition-duration: 0.01ms !important;
      scroll-behavior: auto !important;
    }
  }

  /* High contrast mode support */
  @media (prefers-contrast: high) {
    a {
      text-decoration: underline;
    }
    
    button,
    input,
    textarea,
    select {
      border: 2px solid currentColor;
    }
  }

  /* Print styles */
  @media print {
    *,
    *::before,
    *::after {
      background: transparent !important;
      color: black !important;
      box-shadow: none !important;
      text-shadow: none !important;
    }

    a,
    a:visited {
      text-decoration: underline;
    }

    a[href]::after {
      content: " (" attr(href) ")";
    }

    h2,
    h3 {
      page-break-after: avoid;
    }

    img {
      page-break-inside: avoid;
    }

    p,
    h2,
    h3 {
      orphans: 3;
      widows: 3;
    }
  }
`;