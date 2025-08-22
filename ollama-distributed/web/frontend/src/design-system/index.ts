/**
 * Design System Entry Point
 * Complete design system export with tokens, components, and utilities
 */

// Design tokens
export * from './tokens'

// Components
export * from './components'

// Utilities
export { cn } from '@/utils/cn'

// Design system configuration
export const designSystemConfig = {
  name: 'Ollama Distributed Design System',
  version: '1.0.0',
  components: [
    'Button',
    'Input',
    'Card',
    'Badge', 
    'Alert',
    'Layout',
    'Slider'
  ],
  tokens: [
    'colors',
    'typography',
    'spacing'
  ],
  features: [
    'Dark mode support',
    'Accessibility compliance (WCAG 2.1 AA)',
    'Responsive design',
    'TypeScript support',
    'Customizable variants',
    'Icon integration',
    'Form validation',
    'Animation support'
  ]
} as const

// Theme provider interface
export interface DesignSystemTheme {
  mode: 'light' | 'dark'
  colors: Record<string, string>
  typography: Record<string, any>
  spacing: Record<string, string>
}

// Design system provider setup
export const createDesignSystemTheme = async (mode: 'light' | 'dark' = 'light'): Promise<DesignSystemTheme> => {
  const [colors, typography, spacing] = await Promise.all([
    import('./tokens/colors').then(m => m.colorUtils.generateCSSVariables(mode)),
    import('./tokens/typography').then(m => m.typographyUtils.generateCSSVariables()),
    import('./tokens/spacing').then(m => m.spacingUtils.generateCSSVariables())
  ])

  return {
    mode,
    colors,
    typography,
    spacing
  }
}

export default {
  designSystemConfig,
  createDesignSystemTheme
}