/**
 * Theme Store - Theme state management with persistence
 * Features: Light/dark mode, system preference detection, persistence
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// Types
export type ThemeMode = 'light' | 'dark' | 'system'
export type ResolvedTheme = 'light' | 'dark'

interface ThemeState {
  mode: ThemeMode
  resolvedTheme: ResolvedTheme
  systemTheme: ResolvedTheme
  setMode: (mode: ThemeMode) => void
  toggleTheme: () => void
  updateSystemTheme: (theme: ResolvedTheme) => void
}

// System theme detection
const getSystemTheme = (): ResolvedTheme => {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

// Resolve theme based on mode and system preference
const resolveTheme = (mode: ThemeMode, systemTheme: ResolvedTheme): ResolvedTheme => {
  if (mode === 'system') return systemTheme
  return mode
}

// Theme store
export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => {
      const initialSystemTheme = getSystemTheme()
      
      return {
        mode: 'system',
        systemTheme: initialSystemTheme,
        resolvedTheme: initialSystemTheme,
        
        setMode: (mode: ThemeMode) => {
          const { systemTheme } = get()
          const resolvedTheme = resolveTheme(mode, systemTheme)
          
          set({ mode, resolvedTheme })
          
          // Update document class for CSS
          document.documentElement.classList.remove('light', 'dark')
          document.documentElement.classList.add(resolvedTheme)
        },
        
        toggleTheme: () => {
          const { mode, resolvedTheme } = get()
          
          // If in system mode, switch to opposite of current resolved theme
          if (mode === 'system') {
            const newMode = resolvedTheme === 'light' ? 'dark' : 'light'
            get().setMode(newMode)
          } else {
            // If in manual mode, toggle between light and dark
            const newMode = mode === 'light' ? 'dark' : 'light'
            get().setMode(newMode)
          }
        },
        
        updateSystemTheme: (systemTheme: ResolvedTheme) => {
          const { mode } = get()
          const resolvedTheme = resolveTheme(mode, systemTheme)
          
          set({ systemTheme, resolvedTheme })
          
          // Only update document if in system mode
          if (mode === 'system') {
            document.documentElement.classList.remove('light', 'dark')
            document.documentElement.classList.add(resolvedTheme)
          }
        }
      }
    },
    {
      name: 'ollama-theme-storage',
      partialize: (state) => ({ mode: state.mode }), // Only persist mode
      onRehydrateStorage: () => (state) => {
        if (state) {
          // Re-initialize system theme after hydration
          const systemTheme = getSystemTheme()
          const resolvedTheme = resolveTheme(state.mode, systemTheme)
          
          state.systemTheme = systemTheme
          state.resolvedTheme = resolvedTheme
          
          // Set initial document class
          document.documentElement.classList.remove('light', 'dark')
          document.documentElement.classList.add(resolvedTheme)
        }
      }
    }
  )
)

// Theme utilities
export const themeUtils = {
  // Initialize theme system
  initialize: () => {
    const store = useThemeStore.getState()
    
    // Set initial document class
    document.documentElement.classList.add(store.resolvedTheme)
    
    // Listen for system theme changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      const systemTheme = e.matches ? 'dark' : 'light'
      store.updateSystemTheme(systemTheme)
    }
    
    mediaQuery.addEventListener('change', handleChange)
    
    return () => {
      mediaQuery.removeEventListener('change', handleChange)
    }
  },
  
  // Get current theme colors
  getCurrentTheme: () => {
    const { resolvedTheme } = useThemeStore.getState()
    return resolvedTheme
  },
  
  // Check if dark mode is active
  isDarkMode: () => {
    const { resolvedTheme } = useThemeStore.getState()
    return resolvedTheme === 'dark'
  },
  
  // Get theme-aware color
  getThemeColor: (lightColor: string, darkColor: string) => {
    const { resolvedTheme } = useThemeStore.getState()
    return resolvedTheme === 'dark' ? darkColor : lightColor
  }
}

// React hook for theme
export const useTheme = () => {
  const { mode, resolvedTheme, setMode, toggleTheme } = useThemeStore()
  
  return {
    mode,
    theme: resolvedTheme,
    setMode,
    toggleTheme,
    isDark: resolvedTheme === 'dark',
    isLight: resolvedTheme === 'light',
    isSystem: mode === 'system'
  }
}

// Initialize theme on app start
if (typeof window !== 'undefined') {
  // Initialize immediately
  themeUtils.initialize()
}

export default useThemeStore