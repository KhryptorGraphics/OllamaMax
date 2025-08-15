/**
 * Theme Provider - OllamaMax Design System
 * 
 * Provides theme context and utilities for the entire application.
 */

import React, { createContext, useContext, useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';

// Theme Context
const ThemeContext = createContext();

// Theme Provider Component
export const ThemeProvider = ({ 
  children, 
  defaultTheme = 'light',
  storageKey = 'ollama-theme'
}) => {
  const [currentTheme, setCurrentTheme] = useState(defaultTheme);
  const [mounted, setMounted] = useState(false);

  // Load theme from localStorage on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem(storageKey);
    if (savedTheme && tokens.themes[savedTheme]) {
      setCurrentTheme(savedTheme);
    } else {
      // Detect system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      setCurrentTheme(prefersDark ? 'dark' : 'light');
    }
    setMounted(true);
  }, [storageKey]);

  // Save theme to localStorage when changed
  useEffect(() => {
    if (mounted) {
      localStorage.setItem(storageKey, currentTheme);
    }
  }, [currentTheme, mounted, storageKey]);

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e) => {
      if (!localStorage.getItem(storageKey)) {
        setCurrentTheme(e.matches ? 'dark' : 'light');
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [storageKey]);

  // Toggle between light and dark themes
  const toggleTheme = () => {
    setCurrentTheme(prev => prev === 'light' ? 'dark' : 'light');
  };

  // Set specific theme
  const setTheme = (theme) => {
    if (tokens.themes[theme]) {
      setCurrentTheme(theme);
    } else {
      console.warn(`Theme "${theme}" not found. Available themes:`, Object.keys(tokens.themes));
    }
  };

  // Get current theme object
  const theme = tokens.themes[currentTheme] || tokens.themes.light;

  // Theme utilities
  const themeUtils = {
    // Get color with theme context
    getColor: (colorPath) => {
      const pathArray = colorPath.split('.');
      let color = theme.colors;
      
      for (const key of pathArray) {
        color = color?.[key];
      }
      
      return color || colorPath;
    },

    // Check if current theme is dark
    isDark: currentTheme === 'dark',

    // Check if current theme is light
    isLight: currentTheme === 'light',

    // Get contrast color for given background
    getContrastColor: (backgroundColor) => {
      return themeUtils.isDark ? tokens.colors.neutral[0] : tokens.colors.neutral[900];
    },

    // Apply theme-aware styles
    applyTheme: (styles) => {
      const themedStyles = { ...styles };
      
      // Replace theme color references
      Object.keys(themedStyles).forEach(key => {
        const value = themedStyles[key];
        if (typeof value === 'string' && value.startsWith('theme.')) {
          const colorPath = value.replace('theme.', '');
          themedStyles[key] = themeUtils.getColor(colorPath);
        }
      });
      
      return themedStyles;
    }
  };

  // Context value
  const contextValue = {
    theme,
    currentTheme,
    setTheme,
    toggleTheme,
    utils: themeUtils,
    tokens
  };

  // Apply theme to document root
  useEffect(() => {
    if (mounted) {
      const root = document.documentElement;
      
      // Set CSS custom properties for theme colors
      Object.entries(theme.colors).forEach(([key, value]) => {
        root.style.setProperty(`--color-${key}`, value);
      });
      
      // Set theme class on body
      document.body.className = document.body.className
        .replace(/theme-\w+/g, '')
        .concat(` theme-${currentTheme}`)
        .trim();
    }
  }, [theme, currentTheme, mounted]);

  // Don't render until mounted to prevent hydration mismatch
  if (!mounted) {
    return null;
  }

  return (
    <ThemeContext.Provider value={contextValue}>
      {children}
    </ThemeContext.Provider>
  );
};

ThemeProvider.propTypes = {
  children: PropTypes.node.isRequired,
  defaultTheme: PropTypes.oneOf(['light', 'dark']),
  storageKey: PropTypes.string
};

// Hook to use theme context
export const useTheme = () => {
  const context = useContext(ThemeContext);
  
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  
  return context;
};

// Higher-order component for theme-aware components
export const withTheme = (Component) => {
  const ThemedComponent = (props) => {
    const theme = useTheme();
    return <Component {...props} theme={theme} />;
  };
  
  ThemedComponent.displayName = `withTheme(${Component.displayName || Component.name})`;
  return ThemedComponent;
};

// Hook for responsive values
export const useResponsive = () => {
  const [screenSize, setScreenSize] = useState('md');
  const [orientation, setOrientation] = useState('portrait');
  const [touchDevice, setTouchDevice] = useState(false);

  useEffect(() => {
    const updateScreenSize = () => {
      const width = window.innerWidth;
      const height = window.innerHeight;

      if (width < parseInt(tokens.breakpoints.sm)) {
        setScreenSize('xs');
      } else if (width < parseInt(tokens.breakpoints.md)) {
        setScreenSize('sm');
      } else if (width < parseInt(tokens.breakpoints.lg)) {
        setScreenSize('md');
      } else if (width < parseInt(tokens.breakpoints.xl)) {
        setScreenSize('lg');
      } else {
        setScreenSize('xl');
      }

      // Update orientation
      setOrientation(width > height ? 'landscape' : 'portrait');
    };

    const updateTouchDevice = () => {
      setTouchDevice('ontouchstart' in window || navigator.maxTouchPoints > 0);
    };

    updateScreenSize();
    updateTouchDevice();

    window.addEventListener('resize', updateScreenSize);
    window.addEventListener('orientationchange', updateScreenSize);

    return () => {
      window.removeEventListener('resize', updateScreenSize);
      window.removeEventListener('orientationchange', updateScreenSize);
    };
  }, []);

  return {
    screenSize,
    orientation,
    touchDevice,
    isMobile: screenSize === 'xs' || screenSize === 'sm',
    isTablet: screenSize === 'md',
    isDesktop: screenSize === 'lg' || screenSize === 'xl',
    isPortrait: orientation === 'portrait',
    isLandscape: orientation === 'landscape',
    breakpoints: tokens.breakpoints,

    // Utility functions
    isBreakpoint: (bp) => screenSize === bp,
    isBreakpointUp: (bp) => {
      const sizes = ['xs', 'sm', 'md', 'lg', 'xl'];
      const currentIndex = sizes.indexOf(screenSize);
      const targetIndex = sizes.indexOf(bp);
      return currentIndex >= targetIndex;
    },
    isBreakpointDown: (bp) => {
      const sizes = ['xs', 'sm', 'md', 'lg', 'xl'];
      const currentIndex = sizes.indexOf(screenSize);
      const targetIndex = sizes.indexOf(bp);
      return currentIndex <= targetIndex;
    }
  };
};

// Hook for color mode
export const useColorMode = () => {
  const { currentTheme, toggleTheme, setTheme } = useTheme();
  
  return {
    colorMode: currentTheme,
    toggleColorMode: toggleTheme,
    setColorMode: setTheme
  };
};

// Hook for theme-aware styles
export const useThemedStyles = (styles) => {
  const { utils } = useTheme();
  
  return React.useMemo(() => {
    return utils.applyTheme(styles);
  }, [styles, utils]);
};

export default ThemeProvider;
