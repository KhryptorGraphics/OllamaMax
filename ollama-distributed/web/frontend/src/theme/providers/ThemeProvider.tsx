import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { theme, Theme } from '../theme';
import { GlobalStyles } from '../GlobalStyles';

export type ThemeMode = 'light' | 'dark' | 'system';
export type EffectiveTheme = 'light' | 'dark';

interface ThemeContextType {
  mode: ThemeMode;
  effectiveTheme: EffectiveTheme;
  setTheme: (mode: ThemeMode) => void;
  toggleTheme: () => void;
  systemTheme: EffectiveTheme;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

interface ThemeProviderProps {
  children: ReactNode;
  defaultTheme?: ThemeMode;
  storageKey?: string;
}

// Detect system theme preference
const getSystemTheme = (): EffectiveTheme => {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

// Get stored theme from localStorage
const getStoredTheme = (storageKey: string): ThemeMode | null => {
  if (typeof window === 'undefined') return null;
  try {
    const stored = localStorage.getItem(storageKey);
    if (stored && ['light', 'dark', 'system'].includes(stored)) {
      return stored as ThemeMode;
    }
  } catch (error) {
    console.warn('Failed to read theme from localStorage:', error);
  }
  return null;
};

// Store theme to localStorage
const setStoredTheme = (storageKey: string, mode: ThemeMode): void => {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(storageKey, mode);
  } catch (error) {
    console.warn('Failed to store theme to localStorage:', error);
  }
};

export const ThemeProvider: React.FC<ThemeProviderProps> = ({
  children,
  defaultTheme = 'system',
  storageKey = 'ollama-theme'
}) => {
  const [systemTheme, setSystemTheme] = useState<EffectiveTheme>(() => getSystemTheme());
  const [mode, setMode] = useState<ThemeMode>(() => {
    const stored = getStoredTheme(storageKey);
    return stored || defaultTheme;
  });

  // Calculate effective theme based on mode and system preference
  const effectiveTheme: EffectiveTheme = mode === 'system' ? systemTheme : mode;

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleChange = (e: MediaQueryListEvent) => {
      setSystemTheme(e.matches ? 'dark' : 'light');
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  // Update HTML attributes and CSS custom properties
  useEffect(() => {
    const root = document.documentElement;
    
    // Set data attributes for CSS
    root.setAttribute('data-theme', effectiveTheme);
    root.setAttribute('data-theme-mode', mode);
    
    // Set CSS custom properties
    const themeColors = effectiveTheme === 'dark' 
      ? theme.colors.dark 
      : theme.colors.light;

    // Apply semantic colors as CSS custom properties
    Object.entries(themeColors.background).forEach(([key, value]) => {
      root.style.setProperty(`--color-background-${key}`, value);
    });
    
    Object.entries(themeColors.text).forEach(([key, value]) => {
      root.style.setProperty(`--color-text-${key}`, value);
    });
    
    Object.entries(themeColors.border).forEach(([key, value]) => {
      root.style.setProperty(`--color-border-${key}`, value);
    });
    
    Object.entries(themeColors.surface).forEach(([key, value]) => {
      root.style.setProperty(`--color-surface-${key}`, value);
    });

    // Set motion preferences
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    root.style.setProperty('--motion-scale', prefersReducedMotion ? '1.01' : '1.05');
    root.style.setProperty('--motion-duration', prefersReducedMotion ? '150ms' : '250ms');
    root.style.setProperty('--motion-easing', prefersReducedMotion ? 'ease-out' : 'cubic-bezier(0.4, 0, 0.2, 1)');

    // Update meta theme-color for mobile browsers
    const metaThemeColor = document.querySelector('meta[name="theme-color"]');
    if (metaThemeColor) {
      metaThemeColor.setAttribute('content', themeColors.background.primary);
    }
  }, [effectiveTheme, mode]);

  // Theme setter with persistence
  const setTheme = (newMode: ThemeMode) => {
    setMode(newMode);
    setStoredTheme(storageKey, newMode);
  };

  // Toggle between light and dark (ignores system)
  const toggleTheme = () => {
    if (mode === 'system') {
      setTheme(systemTheme === 'dark' ? 'light' : 'dark');
    } else {
      setTheme(mode === 'dark' ? 'light' : 'dark');
    }
  };

  const contextValue: ThemeContextType = {
    mode,
    effectiveTheme,
    setTheme,
    toggleTheme,
    systemTheme
  };

  // Get the current theme object
  const currentTheme: Theme = {
    ...theme,
    mode: effectiveTheme,
    colors: effectiveTheme === 'dark' ? theme.colors.dark : theme.colors.light,
    shadows: effectiveTheme === 'dark' ? theme.shadows.dark : theme.shadows.light,
    elevation: effectiveTheme === 'dark' ? theme.elevation.dark : theme.elevation.light
  };

  return (
    <ThemeContext.Provider value={contextValue}>
      <StyledThemeProvider theme={currentTheme}>
        <GlobalStyles />
        {children}
      </StyledThemeProvider>
    </ThemeContext.Provider>
  );
};

export const useTheme = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};