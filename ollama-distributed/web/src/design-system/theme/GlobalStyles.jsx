/**
 * Global Styles Component
 * 
 * Provides global CSS styles and resets for the application.
 */

import React, { useEffect } from 'react';
import { useTheme } from './ThemeProvider.jsx';
import { tokens } from '../tokens.js';

const GlobalStyles = () => {
  const { theme, currentTheme } = useTheme();

  useEffect(() => {
    // Create or update global styles
    const styleId = 'ollama-global-styles';
    let styleElement = document.getElementById(styleId);
    
    if (!styleElement) {
      styleElement = document.createElement('style');
      styleElement.id = styleId;
      document.head.appendChild(styleElement);
    }

    const globalCSS = `
      /* CSS Reset and Base Styles */
      *, *::before, *::after {
        box-sizing: border-box;
      }

      * {
        margin: 0;
        padding: 0;
      }

      html, body {
        height: 100%;
      }

      body {
        font-family: ${tokens.typography.fontFamily.sans.join(', ')};
        font-size: ${tokens.typography.fontSize.base[0]};
        line-height: ${tokens.typography.lineHeight.normal};
        color: ${theme.colors.text};
        background-color: ${theme.colors.background};
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        transition: color 0.3s ease, background-color 0.3s ease;
      }

      #root {
        height: 100%;
      }

      /* Typography */
      h1, h2, h3, h4, h5, h6 {
        font-family: ${tokens.typography.fontFamily.display.join(', ')};
        font-weight: ${tokens.typography.fontWeight.bold};
        line-height: ${tokens.typography.lineHeight.tight};
        color: ${theme.colors.text};
      }

      h1 {
        font-size: ${tokens.typography.fontSize['3xl'][0]};
      }

      h2 {
        font-size: ${tokens.typography.fontSize['2xl'][0]};
      }

      h3 {
        font-size: ${tokens.typography.fontSize.xl[0]};
      }

      h4 {
        font-size: ${tokens.typography.fontSize.lg[0]};
      }

      h5 {
        font-size: ${tokens.typography.fontSize.base[0]};
      }

      h6 {
        font-size: ${tokens.typography.fontSize.sm[0]};
      }

      p {
        margin-bottom: ${tokens.spacing[4]};
        color: ${theme.colors.textSecondary};
        line-height: ${tokens.typography.lineHeight.relaxed};
      }

      a {
        color: ${theme.colors.primary};
        text-decoration: none;
        transition: color ${tokens.animation.duration.fast} ${tokens.animation.easing.easeInOut};
      }

      a:hover {
        color: ${theme.colors.primaryVariant};
        text-decoration: underline;
      }

      /* Form Elements */
      button {
        font-family: inherit;
        font-size: inherit;
        line-height: inherit;
        margin: 0;
      }

      input, textarea, select {
        font-family: inherit;
        font-size: inherit;
        line-height: inherit;
        margin: 0;
      }

      /* Focus Styles */
      *:focus {
        outline: 2px solid ${theme.colors.primary};
        outline-offset: 2px;
      }

      *:focus:not(:focus-visible) {
        outline: none;
      }

      /* Scrollbar Styles */
      ::-webkit-scrollbar {
        width: 8px;
        height: 8px;
      }

      ::-webkit-scrollbar-track {
        background: ${theme.colors.surfaceVariant};
        border-radius: 4px;
      }

      ::-webkit-scrollbar-thumb {
        background: ${theme.colors.border};
        border-radius: 4px;
        transition: background-color 0.2s ease;
      }

      ::-webkit-scrollbar-thumb:hover {
        background: ${theme.colors.textMuted};
      }

      /* Selection Styles */
      ::selection {
        background-color: ${theme.colors.primary}40;
        color: ${theme.colors.text};
      }

      ::-moz-selection {
        background-color: ${theme.colors.primary}40;
        color: ${theme.colors.text};
      }

      /* Animations */
      @keyframes spin {
        from {
          transform: rotate(0deg);
        }
        to {
          transform: rotate(360deg);
        }
      }

      @keyframes fadeIn {
        from {
          opacity: 0;
        }
        to {
          opacity: 1;
        }
      }

      @keyframes slideUp {
        from {
          opacity: 0;
          transform: translateY(10px);
        }
        to {
          opacity: 1;
          transform: translateY(0);
        }
      }

      @keyframes pulse {
        0%, 100% {
          opacity: 1;
        }
        50% {
          opacity: 0.5;
        }
      }

      /* Utility Classes */
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

      .fade-in {
        animation: fadeIn ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut};
      }

      .slide-up {
        animation: slideUp ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut};
      }

      .pulse {
        animation: pulse ${tokens.animation.duration.slow} ${tokens.animation.easing.easeInOut} infinite;
      }

      /* Theme-specific styles */
      .theme-light {
        color-scheme: light;
      }

      .theme-dark {
        color-scheme: dark;
      }

      /* Print Styles */
      @media print {
        * {
          background: transparent !important;
          color: black !important;
          box-shadow: none !important;
          text-shadow: none !important;
        }

        a, a:visited {
          text-decoration: underline;
        }

        a[href]:after {
          content: " (" attr(href) ")";
        }

        abbr[title]:after {
          content: " (" attr(title) ")";
        }

        .no-print {
          display: none !important;
        }
      }

      /* Responsive Design Helpers */
      @media (max-width: ${tokens.breakpoints.sm}) {
        body {
          font-size: ${tokens.typography.fontSize.sm[0]};
        }

        h1 {
          font-size: ${tokens.typography.fontSize['2xl'][0]};
        }

        h2 {
          font-size: ${tokens.typography.fontSize.xl[0]};
        }

        h3 {
          font-size: ${tokens.typography.fontSize.lg[0]};
        }
      }

      /* Reduced Motion */
      @media (prefers-reduced-motion: reduce) {
        *, *::before, *::after {
          animation-duration: 0.01ms !important;
          animation-iteration-count: 1 !important;
          transition-duration: 0.01ms !important;
          scroll-behavior: auto !important;
        }
      }

      /* High Contrast Mode */
      @media (prefers-contrast: high) {
        * {
          border-color: currentColor !important;
        }
      }

      /* Dark Mode Media Query */
      @media (prefers-color-scheme: dark) {
        .auto-theme {
          color-scheme: dark;
        }
      }

      /* Loading States */
      .loading {
        pointer-events: none;
        opacity: 0.6;
      }

      .loading * {
        cursor: wait !important;
      }

      /* Error States */
      .error {
        color: ${theme.colors.error} !important;
        border-color: ${theme.colors.error} !important;
      }

      /* Success States */
      .success {
        color: ${theme.colors.success} !important;
        border-color: ${theme.colors.success} !important;
      }

      /* Warning States */
      .warning {
        color: ${theme.colors.warning} !important;
        border-color: ${theme.colors.warning} !important;
      }
    `;

    styleElement.textContent = globalCSS;
  }, [theme, currentTheme]);

  return null;
};

export default GlobalStyles;
