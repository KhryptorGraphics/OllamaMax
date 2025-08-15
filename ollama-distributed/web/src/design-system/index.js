/**
 * OllamaMax Design System
 * 
 * A comprehensive design system providing consistent UI components,
 * design tokens, and styling utilities for the OllamaMax application.
 */

// Design Tokens
export { default as tokens } from './tokens.js';

// Core Components
export { default as Button } from './components/Button.jsx';
export { default as Input } from './components/Input.jsx';
export { default as Card } from './components/Card.jsx';
export { default as Badge } from './components/Badge.jsx';
export { default as Modal, ConfirmModal } from './components/Modal.jsx';
export { default as Toast, ToastContainer } from './components/Toast.jsx';

// Accessibility Components
export { default as SkipLink, SkipLinks } from './components/SkipLink.jsx';
export { default as FocusTrap, useFocusTrap } from './components/FocusTrap.jsx';

// Theme Provider and Utilities
export { ThemeProvider, useTheme } from './theme/ThemeProvider.jsx';
export { default as GlobalStyles } from './theme/GlobalStyles.jsx';

// Utility Functions
export * from './utils/responsive.js';
export * from './utils/accessibility.js';
export * from './utils/animations.js';

// Design System Version
export const VERSION = '1.0.0';

// Design System Configuration
export const config = {
  name: 'OllamaMax Design System',
  version: VERSION,
  description: 'A comprehensive design system for OllamaMax distributed applications',
  author: 'OllamaMax Team',
  license: 'MIT'
};

// Component Registry for Development Tools
export const componentRegistry = {
  Button: {
    name: 'Button',
    description: 'Interactive button component with multiple variants and states',
    category: 'Form',
    status: 'stable'
  },
  Input: {
    name: 'Input',
    description: 'Text input component with validation and accessibility features',
    category: 'Form',
    status: 'stable'
  },
  Card: {
    name: 'Card',
    description: 'Container component for displaying content with consistent styling',
    category: 'Layout',
    status: 'stable'
  }
};

// Design System Utilities
export const utils = {
  // Responsive breakpoint utilities
  breakpoint: {
    up: (size) => `@media (min-width: ${tokens.breakpoints[size]})`,
    down: (size) => {
      const breakpointValues = Object.values(tokens.breakpoints);
      const currentIndex = Object.keys(tokens.breakpoints).indexOf(size);
      const maxWidth = currentIndex > 0 ? 
        `${parseInt(breakpointValues[currentIndex]) - 1}px` : 
        tokens.breakpoints.xs;
      return `@media (max-width: ${maxWidth})`;
    },
    between: (min, max) => 
      `@media (min-width: ${tokens.breakpoints[min]}) and (max-width: ${parseInt(tokens.breakpoints[max]) - 1}px)`
  },

  // Spacing utilities
  spacing: {
    get: (value) => tokens.spacing[value] || value,
    responsive: (values) => {
      const breakpoints = Object.keys(tokens.breakpoints);
      return breakpoints.reduce((acc, bp, index) => {
        if (values[bp]) {
          acc[utils.breakpoint.up(bp)] = {
            padding: tokens.spacing[values[bp]]
          };
        }
        return acc;
      }, {});
    }
  },

  // Typography utilities
  typography: {
    heading: (level) => ({
      fontFamily: tokens.typography.fontFamily.display.join(', '),
      fontSize: tokens.typography.fontSize[`${level}xl`]?.[0] || tokens.typography.fontSize.xl[0],
      fontWeight: tokens.typography.fontWeight.bold,
      lineHeight: tokens.typography.lineHeight.tight,
      color: tokens.colors.neutral[900]
    }),
    body: (size = 'base') => ({
      fontFamily: tokens.typography.fontFamily.sans.join(', '),
      fontSize: tokens.typography.fontSize[size][0],
      lineHeight: tokens.typography.lineHeight.normal,
      color: tokens.colors.neutral[700]
    }),
    code: (size = 'sm') => ({
      fontFamily: tokens.typography.fontFamily.mono.join(', '),
      fontSize: tokens.typography.fontSize[size][0],
      backgroundColor: tokens.colors.neutral[100],
      padding: `${tokens.spacing[1]} ${tokens.spacing[2]}`,
      borderRadius: tokens.borderRadius.sm,
      color: tokens.colors.neutral[800]
    })
  },

  // Color utilities
  color: {
    alpha: (color, alpha) => {
      // Convert hex to rgba
      const hex = color.replace('#', '');
      const r = parseInt(hex.substr(0, 2), 16);
      const g = parseInt(hex.substr(2, 2), 16);
      const b = parseInt(hex.substr(4, 2), 16);
      return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    },
    contrast: (backgroundColor) => {
      // Simple contrast calculation
      const hex = backgroundColor.replace('#', '');
      const r = parseInt(hex.substr(0, 2), 16);
      const g = parseInt(hex.substr(2, 2), 16);
      const b = parseInt(hex.substr(4, 2), 16);
      const brightness = (r * 299 + g * 587 + b * 114) / 1000;
      return brightness > 128 ? tokens.colors.neutral[900] : tokens.colors.neutral[0];
    }
  },

  // Animation utilities
  animation: {
    fadeIn: {
      animation: `fadeIn ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut}`
    },
    slideUp: {
      animation: `slideUp ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut}`
    },
    pulse: {
      animation: `pulse ${tokens.animation.duration.slow} ${tokens.animation.easing.easeInOut} infinite`
    },
    spin: {
      animation: `spin ${tokens.animation.duration.slow} linear infinite`
    }
  },

  // Shadow utilities
  shadow: {
    get: (level) => tokens.shadows[level],
    colored: (color, level = 'md') => {
      const shadow = tokens.shadows[level];
      return shadow.replace('rgb(0 0 0', `${color}`);
    }
  }
};

// Validation utilities for development
export const validate = {
  // Validate component props against design system standards
  props: (componentName, props) => {
    const warnings = [];
    
    // Check for deprecated props
    if (props.variant && !componentRegistry[componentName]?.variants?.includes(props.variant)) {
      warnings.push(`Invalid variant "${props.variant}" for ${componentName}`);
    }
    
    // Check for accessibility issues
    if (componentName === 'Button' && !props.children && !props['aria-label']) {
      warnings.push('Button should have either children or aria-label for accessibility');
    }
    
    return warnings;
  },

  // Validate color contrast
  contrast: (foreground, background) => {
    // Simplified contrast ratio calculation
    const getLuminance = (color) => {
      const hex = color.replace('#', '');
      const r = parseInt(hex.substr(0, 2), 16) / 255;
      const g = parseInt(hex.substr(2, 2), 16) / 255;
      const b = parseInt(hex.substr(4, 2), 16) / 255;
      return 0.2126 * r + 0.7152 * g + 0.0722 * b;
    };
    
    const l1 = getLuminance(foreground);
    const l2 = getLuminance(background);
    const ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
    
    return {
      ratio,
      aa: ratio >= 4.5,
      aaa: ratio >= 7
    };
  }
};

// Development tools
export const dev = {
  // Log design system information
  info: () => {
    console.group('🎨 OllamaMax Design System');
    console.log(`Version: ${VERSION}`);
    console.log(`Components: ${Object.keys(componentRegistry).length}`);
    console.log(`Tokens: ${Object.keys(tokens).length} categories`);
    console.groupEnd();
  },

  // List all available components
  components: () => {
    console.table(componentRegistry);
  },

  // Show color palette
  colors: () => {
    console.group('🎨 Color Palette');
    Object.entries(tokens.colors).forEach(([name, shades]) => {
      console.group(name);
      if (typeof shades === 'object') {
        Object.entries(shades).forEach(([shade, value]) => {
          console.log(`%c${shade}: ${value}`, `color: ${value}; font-weight: bold;`);
        });
      } else {
        console.log(`%c${name}: ${shades}`, `color: ${shades}; font-weight: bold;`);
      }
      console.groupEnd();
    });
    console.groupEnd();
  },

  // Validate current theme
  validateTheme: (theme) => {
    const requiredTokens = ['colors', 'typography', 'spacing'];
    const missing = requiredTokens.filter(token => !theme[token]);
    
    if (missing.length > 0) {
      console.warn('Missing required theme tokens:', missing);
      return false;
    }
    
    console.log('✅ Theme validation passed');
    return true;
  }
};

// Export default design system object
export default {
  tokens,
  utils,
  validate,
  dev,
  config,
  VERSION
};
