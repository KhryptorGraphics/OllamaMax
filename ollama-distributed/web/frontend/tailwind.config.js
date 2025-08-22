/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx,js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      // Colors from design tokens
      colors: {
        // Brand colors
        brand: {
          DEFAULT: '#0ea5e9',
          dark: '#0284c7',
          light: '#7dd3fc',
        },
        
        // Semantic colors from design tokens
        primary: {
          50: 'var(--color-primary-50, #eff6ff)',
          100: 'var(--color-primary-100, #dbeafe)',
          200: 'var(--color-primary-200, #bfdbfe)',
          300: 'var(--color-primary-300, #93c5fd)',
          400: 'var(--color-primary-400, #60a5fa)',
          500: 'var(--color-primary-500, #3b82f6)',
          600: 'var(--color-primary-600, #2563eb)',
          700: 'var(--color-primary-700, #1d4ed8)',
          800: 'var(--color-primary-800, #1e40af)',
          900: 'var(--color-primary-900, #1e3a8a)',
          950: 'var(--color-primary-950, #172554)',
        },
        
        secondary: {
          50: 'var(--color-secondary-50, #f8fafc)',
          100: 'var(--color-secondary-100, #f1f5f9)',
          200: 'var(--color-secondary-200, #e2e8f0)',
          300: 'var(--color-secondary-300, #cbd5e1)',
          400: 'var(--color-secondary-400, #94a3b8)',
          500: 'var(--color-secondary-500, #64748b)',
          600: 'var(--color-secondary-600, #475569)',
          700: 'var(--color-secondary-700, #334155)',
          800: 'var(--color-secondary-800, #1e293b)',
          900: 'var(--color-secondary-900, #0f172a)',
          950: 'var(--color-secondary-950, #020617)',
        },
        
        // Semantic colors
        foreground: 'var(--color-text-primary)',
        background: 'var(--color-bg-primary)',
        muted: 'var(--color-text-secondary)',
        'muted-foreground': 'var(--color-text-tertiary)',
        border: 'var(--color-border-primary)',
        input: 'var(--color-border-secondary)',
        ring: 'var(--color-border-focus)',
        
        // Card specific colors
        card: {
          DEFAULT: 'var(--color-bg-elevated, #ffffff)',
          foreground: 'var(--color-text-primary)',
        },
        
        destructive: {
          DEFAULT: 'var(--color-error-500, #ef4444)',
          foreground: 'var(--color-status-error-text)',
        },
        
        accent: {
          DEFAULT: 'var(--color-secondary-100)',
          foreground: 'var(--color-secondary-900)',
        },
        
        success: {
          50: 'var(--color-success-50, #f0fdf4)',
          100: 'var(--color-success-100, #dcfce7)',
          200: 'var(--color-success-200, #bbf7d0)',
          300: 'var(--color-success-300, #86efac)',
          400: 'var(--color-success-400, #4ade80)',
          500: 'var(--color-success-500, #22c55e)',
          600: 'var(--color-success-600, #16a34a)',
          700: 'var(--color-success-700, #15803d)',
          800: 'var(--color-success-800, #166534)',
          900: 'var(--color-success-900, #14532d)',
        },
        
        warning: {
          50: 'var(--color-warning-50, #fffbeb)',
          100: 'var(--color-warning-100, #fef3c7)',
          200: 'var(--color-warning-200, #fde68a)',
          300: 'var(--color-warning-300, #fcd34d)',
          400: 'var(--color-warning-400, #fbbf24)',
          500: 'var(--color-warning-500, #f59e0b)',
          600: 'var(--color-warning-600, #d97706)',
          700: 'var(--color-warning-700, #b45309)',
          800: 'var(--color-warning-800, #92400e)',
          900: 'var(--color-warning-900, #78350f)',
        },
        
        error: {
          50: 'var(--color-error-50, #fef2f2)',
          100: 'var(--color-error-100, #fee2e2)',
          200: 'var(--color-error-200, #fecaca)',
          300: 'var(--color-error-300, #fca5a5)',
          400: 'var(--color-error-400, #f87171)',
          500: 'var(--color-error-500, #ef4444)',
          600: 'var(--color-error-600, #dc2626)',
          700: 'var(--color-error-700, #b91c1c)',
          800: 'var(--color-error-800, #991b1b)',
          900: 'var(--color-error-900, #7f1d1d)',
        },
        
        info: {
          50: 'var(--color-info-50, #eff6ff)',
          100: 'var(--color-info-100, #dbeafe)',
          200: 'var(--color-info-200, #bfdbfe)',
          300: 'var(--color-info-300, #93c5fd)',
          400: 'var(--color-info-400, #60a5fa)',
          500: 'var(--color-info-500, #3b82f6)',
          600: 'var(--color-info-600, #2563eb)',
          700: 'var(--color-info-700, #1d4ed8)',
          800: 'var(--color-info-800, #1e40af)',
          900: 'var(--color-info-900, #1e3a8a)',
        }
      },
      
      // Typography from design tokens
      fontFamily: {
        sans: ['var(--font-family-sans)', 'Inter', 'system-ui', 'sans-serif'],
        serif: ['var(--font-family-serif)', 'Georgia', 'serif'],
        mono: ['var(--font-family-mono)', 'Fira Code', 'monospace'],
      },
      
      fontSize: {
        xs: ['var(--font-size-xs, 0.75rem)', { lineHeight: 'var(--line-height-normal, 1.5)' }],
        sm: ['var(--font-size-sm, 0.875rem)', { lineHeight: 'var(--line-height-normal, 1.5)' }],
        base: ['var(--font-size-base, 1rem)', { lineHeight: 'var(--line-height-normal, 1.5)' }],
        lg: ['var(--font-size-lg, 1.125rem)', { lineHeight: 'var(--line-height-normal, 1.5)' }],
        xl: ['var(--font-size-xl, 1.25rem)', { lineHeight: 'var(--line-height-snug, 1.375)' }],
        '2xl': ['var(--font-size-2xl, 1.5rem)', { lineHeight: 'var(--line-height-snug, 1.375)' }],
        '3xl': ['var(--font-size-3xl, 1.875rem)', { lineHeight: 'var(--line-height-tight, 1.25)' }],
        '4xl': ['var(--font-size-4xl, 2.25rem)', { lineHeight: 'var(--line-height-tight, 1.25)' }],
        '5xl': ['var(--font-size-5xl, 3rem)', { lineHeight: 'var(--line-height-none, 1)' }],
        '6xl': ['var(--font-size-6xl, 3.75rem)', { lineHeight: 'var(--line-height-none, 1)' }],
        '7xl': ['var(--font-size-7xl, 4.5rem)', { lineHeight: 'var(--line-height-none, 1)' }],
        '8xl': ['var(--font-size-8xl, 6rem)', { lineHeight: 'var(--line-height-none, 1)' }],
        '9xl': ['var(--font-size-9xl, 8rem)', { lineHeight: 'var(--line-height-none, 1)' }],
      },
      
      fontWeight: {
        thin: 'var(--font-weight-thin, 100)',
        extralight: 'var(--font-weight-extralight, 200)',
        light: 'var(--font-weight-light, 300)',
        normal: 'var(--font-weight-normal, 400)',
        medium: 'var(--font-weight-medium, 500)',
        semibold: 'var(--font-weight-semibold, 600)',
        bold: 'var(--font-weight-bold, 700)',
        extrabold: 'var(--font-weight-extrabold, 800)',
        black: 'var(--font-weight-black, 900)',
      },
      
      // Spacing from design tokens
      spacing: {
        'px': 'var(--spacing-px, 1px)',
        '0': 'var(--spacing-0, 0)',
        '0.5': 'var(--spacing-0.5, 0.125rem)',
        '1': 'var(--spacing-1, 0.25rem)',
        '1.5': 'var(--spacing-1.5, 0.375rem)',
        '2': 'var(--spacing-2, 0.5rem)',
        '2.5': 'var(--spacing-2.5, 0.625rem)',
        '3': 'var(--spacing-3, 0.75rem)',
        '3.5': 'var(--spacing-3.5, 0.875rem)',
        '4': 'var(--spacing-4, 1rem)',
        '5': 'var(--spacing-5, 1.25rem)',
        '6': 'var(--spacing-6, 1.5rem)',
        '7': 'var(--spacing-7, 1.75rem)',
        '8': 'var(--spacing-8, 2rem)',
        '9': 'var(--spacing-9, 2.25rem)',
        '10': 'var(--spacing-10, 2.5rem)',
        '11': 'var(--spacing-11, 2.75rem)',
        '12': 'var(--spacing-12, 3rem)',
        '14': 'var(--spacing-14, 3.5rem)',
        '16': 'var(--spacing-16, 4rem)',
        '20': 'var(--spacing-20, 5rem)',
        '24': 'var(--spacing-24, 6rem)',
        '28': 'var(--spacing-28, 7rem)',
        '32': 'var(--spacing-32, 8rem)',
        '36': 'var(--spacing-36, 9rem)',
        '40': 'var(--spacing-40, 10rem)',
        '44': 'var(--spacing-44, 11rem)',
        '48': 'var(--spacing-48, 12rem)',
        '52': 'var(--spacing-52, 13rem)',
        '56': 'var(--spacing-56, 14rem)',
        '60': 'var(--spacing-60, 15rem)',
        '64': 'var(--spacing-64, 16rem)',
        '72': 'var(--spacing-72, 18rem)',
        '80': 'var(--spacing-80, 20rem)',
        '96': 'var(--spacing-96, 24rem)',
      },
      
      // Border radius from design tokens
      borderRadius: {
        none: 'var(--radius-none, 0)',
        xs: 'var(--radius-xs, 0.125rem)',
        sm: 'var(--radius-sm, 0.25rem)',
        md: 'var(--radius-md, 0.375rem)',
        lg: 'var(--radius-lg, 0.5rem)',
        xl: 'var(--radius-xl, 0.75rem)',
        '2xl': 'var(--radius-2xl, 1rem)',
        '3xl': 'var(--radius-3xl, 1.5rem)',
        full: 'var(--radius-full, 9999px)',
        card: 'var(--radius-card, 0.5rem)',
      },
      
      // Box shadows from design tokens
      boxShadow: {
        xs: 'var(--shadow-xs)',
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
        xl: 'var(--shadow-xl)',
        '2xl': 'var(--shadow-2xl)',
        inner: 'var(--shadow-inner)',
        none: 'var(--shadow-none)',
      },
      
      // Transitions from design tokens
      transitionDuration: {
        instant: 'var(--duration-instant, 0ms)',
        fast: 'var(--duration-fast, 150ms)',
        normal: 'var(--duration-normal, 200ms)',
        medium: 'var(--duration-medium, 300ms)',
        slow: 'var(--duration-slow, 500ms)',
        slower: 'var(--duration-slower, 700ms)',
        slowest: 'var(--duration-slowest, 1000ms)',
      },
      
      transitionTimingFunction: {
        linear: 'var(--easing-linear)',
        ease: 'var(--easing-ease)',
        'ease-in': 'var(--easing-easeIn)',
        'ease-out': 'var(--easing-easeOut)',
        'ease-in-out': 'var(--easing-easeInOut)',
        standard: 'var(--easing-standard)',
        decelerate: 'var(--easing-decelerate)',
        accelerate: 'var(--easing-accelerate)',
        spring: 'var(--easing-spring)',
        bounce: 'var(--easing-bounce)',
      },
      
      // Breakpoints from design tokens
      screens: {
        xs: 'var(--breakpoint-xs, 320px)',
        sm: 'var(--breakpoint-sm, 640px)',
        md: 'var(--breakpoint-md, 768px)',
        lg: 'var(--breakpoint-lg, 1024px)',
        xl: 'var(--breakpoint-xl, 1280px)',
        '2xl': 'var(--breakpoint-2xl, 1536px)',
        '3xl': 'var(--breakpoint-3xl, 1920px)',
        '4xl': 'var(--breakpoint-4xl, 2560px)',
      },
      
      // Animation utilities
      animation: {
        'spin': 'spin var(--duration-slowest, 1s) var(--easing-linear) infinite',
        'ping': 'ping var(--duration-slowest, 1s) var(--easing-standard) infinite',
        'pulse': 'pulse var(--duration-slower, 2s) var(--easing-ease-in-out) infinite',
        'bounce': 'bounce var(--duration-slowest, 1s) infinite',
        'fade-in': 'fadeIn var(--duration-medium, 300ms) var(--easing-decelerate)',
        'fade-out': 'fadeOut var(--duration-normal, 200ms) var(--easing-accelerate)',
        'slide-in-up': 'slideInUp var(--duration-medium, 300ms) var(--easing-decelerate)',
        'slide-in-down': 'slideInDown var(--duration-medium, 300ms) var(--easing-decelerate)',
        'scale-in': 'scaleIn var(--duration-medium, 300ms) var(--easing-spring)',
        'scale-out': 'scaleOut var(--duration-normal, 200ms) var(--easing-accelerate)',
      },
      
      // Container queries support
      container: {
        center: true,
        padding: {
          DEFAULT: 'var(--spacing-4, 1rem)',
          sm: 'var(--spacing-6, 1.5rem)',
          md: 'var(--spacing-8, 2rem)',
          lg: 'var(--spacing-12, 3rem)',
          xl: 'var(--spacing-16, 4rem)',
          '2xl': 'var(--spacing-24, 6rem)',
        },
        screens: {
          xs: 'var(--breakpoint-xs, 320px)',
          sm: 'var(--breakpoint-sm, 640px)',
          md: 'var(--breakpoint-md, 768px)',
          lg: 'var(--breakpoint-lg, 1024px)',
          xl: 'var(--breakpoint-xl, 1280px)',
          '2xl': 'var(--breakpoint-2xl, 1536px)',
        },
      },
    },
  },
  plugins: [
    // Add custom utilities for design tokens
    function({ addUtilities }) {
      addUtilities({
        '.transition-smooth': {
          transition: 'var(--transition-smooth)',
        },
        '.transition-smooth-fast': {
          transition: 'var(--transition-smoothFast)',
        },
        '.transition-smooth-slow': {
          transition: 'var(--transition-smoothSlow)',
        },
        '.shadow-elevation-1': {
          boxShadow: 'var(--shadow-elevation-1)',
        },
        '.shadow-elevation-2': {
          boxShadow: 'var(--shadow-elevation-2)',
        },
        '.shadow-elevation-3': {
          boxShadow: 'var(--shadow-elevation-3)',
        },
        '.shadow-elevation-4': {
          boxShadow: 'var(--shadow-elevation-4)',
        },
        '.shadow-elevation-5': {
          boxShadow: 'var(--shadow-elevation-5)',
        },
        '.shadow-elevation-6': {
          boxShadow: 'var(--shadow-elevation-6)',
        },
        // Line clamp utilities
        '.line-clamp-1': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '1',
        },
        '.line-clamp-2': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '2',
        },
        '.line-clamp-3': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '3',
        },
        '.line-clamp-4': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '4',
        },
        '.line-clamp-5': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '5',
        },
        '.line-clamp-6': {
          overflow: 'hidden',
          display: '-webkit-box',
          '-webkit-box-orient': 'vertical',
          '-webkit-line-clamp': '6',
        },
        // Focus ring utilities
        '.ring-focus': {
          '--tw-ring-color': 'var(--color-primary-500, #3b82f6)',
          '--tw-ring-opacity': '0.5',
          '--tw-ring-offset-width': '2px',
          '--tw-ring-offset-color': 'var(--color-bg-primary, #ffffff)',
          '--tw-ring-inset': '',
          '--tw-ring-width': '2px',
          'box-shadow': 'var(--tw-ring-inset) 0 0 0 calc(var(--tw-ring-width) + var(--tw-ring-offset-width)) var(--tw-ring-color)',
        },
      })
    }
  ],
}

