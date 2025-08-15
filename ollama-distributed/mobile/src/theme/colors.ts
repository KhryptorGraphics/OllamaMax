/**
 * Color System for OllamaMax Mobile
 * 
 * Consistent color palette across iOS and Android platforms.
 */

export const Colors = {
  // Primary brand colors
  primary: '#0ea5e9',
  primaryVariant: '#0284c7',
  secondary: '#d946ef',
  secondaryVariant: '#c026d3',

  // Semantic colors
  success: '#22c55e',
  successVariant: '#16a34a',
  warning: '#f59e0b',
  warningVariant: '#d97706',
  error: '#ef4444',
  errorVariant: '#dc2626',
  info: '#3b82f6',
  infoVariant: '#2563eb',

  // Neutral colors
  neutral: {
    0: '#ffffff',
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b',
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
    950: '#020617',
  },

  // Light theme
  light: {
    background: '#ffffff',
    surface: '#f8fafc',
    surfaceVariant: '#f1f5f9',
    text: '#0f172a',
    textSecondary: '#475569',
    textMuted: '#94a3b8',
    border: '#e2e8f0',
    divider: '#f1f5f9',
    overlay: 'rgba(15, 23, 42, 0.5)',
    shadow: 'rgba(15, 23, 42, 0.1)',
  },

  // Dark theme
  dark: {
    background: '#0f172a',
    surface: '#1e293b',
    surfaceVariant: '#334155',
    text: '#f8fafc',
    textSecondary: '#cbd5e1',
    textMuted: '#94a3b8',
    border: '#334155',
    divider: '#475569',
    overlay: 'rgba(0, 0, 0, 0.7)',
    shadow: 'rgba(0, 0, 0, 0.3)',
  },

  // Status colors with opacity variants
  status: {
    online: '#22c55e',
    offline: '#ef4444',
    warning: '#f59e0b',
    maintenance: '#8b5cf6',
  },

  // Chart colors
  chart: {
    primary: '#0ea5e9',
    secondary: '#d946ef',
    tertiary: '#22c55e',
    quaternary: '#f59e0b',
    quinary: '#ef4444',
    senary: '#8b5cf6',
  },

  // Gradient colors
  gradients: {
    primary: ['#0ea5e9', '#d946ef'],
    success: ['#22c55e', '#16a34a'],
    warning: ['#f59e0b', '#d97706'],
    error: ['#ef4444', '#dc2626'],
    neutral: ['#64748b', '#334155'],
  },

  // Platform-specific colors
  ios: {
    systemBlue: '#007AFF',
    systemGreen: '#34C759',
    systemIndigo: '#5856D6',
    systemOrange: '#FF9500',
    systemPink: '#FF2D92',
    systemPurple: '#AF52DE',
    systemRed: '#FF3B30',
    systemTeal: '#5AC8FA',
    systemYellow: '#FFCC00',
    systemGray: '#8E8E93',
    systemGray2: '#AEAEB2',
    systemGray3: '#C7C7CC',
    systemGray4: '#D1D1D6',
    systemGray5: '#E5E5EA',
    systemGray6: '#F2F2F7',
  },

  android: {
    materialBlue: '#2196F3',
    materialGreen: '#4CAF50',
    materialOrange: '#FF9800',
    materialRed: '#F44336',
    materialPurple: '#9C27B0',
    materialTeal: '#009688',
    materialYellow: '#FFEB3B',
    materialPink: '#E91E63',
    materialIndigo: '#3F51B5',
    materialCyan: '#00BCD4',
  },
};

// Color utility functions
export const colorUtils = {
  // Add opacity to a color
  withOpacity: (color: string, opacity: number): string => {
    if (color.startsWith('#')) {
      const alpha = Math.round(opacity * 255).toString(16).padStart(2, '0');
      return `${color}${alpha}`;
    }
    if (color.startsWith('rgb(')) {
      return color.replace('rgb(', 'rgba(').replace(')', `, ${opacity})`);
    }
    return color;
  },

  // Get contrast color (black or white) for a given background
  getContrastColor: (backgroundColor: string): string => {
    // Simple contrast calculation
    const hex = backgroundColor.replace('#', '');
    const r = parseInt(hex.substr(0, 2), 16);
    const g = parseInt(hex.substr(2, 2), 16);
    const b = parseInt(hex.substr(4, 2), 16);
    const brightness = (r * 299 + g * 587 + b * 114) / 1000;
    return brightness > 128 ? Colors.neutral[900] : Colors.neutral[0];
  },

  // Lighten a color
  lighten: (color: string, amount: number): string => {
    // Simplified lighten function
    const hex = color.replace('#', '');
    const num = parseInt(hex, 16);
    const amt = Math.round(2.55 * amount);
    const R = (num >> 16) + amt;
    const G = (num >> 8 & 0x00FF) + amt;
    const B = (num & 0x0000FF) + amt;
    return `#${(0x1000000 + (R < 255 ? R < 1 ? 0 : R : 255) * 0x10000 +
      (G < 255 ? G < 1 ? 0 : G : 255) * 0x100 +
      (B < 255 ? B < 1 ? 0 : B : 255)).toString(16).slice(1)}`;
  },

  // Darken a color
  darken: (color: string, amount: number): string => {
    return colorUtils.lighten(color, -amount);
  },

  // Get semantic color based on status
  getStatusColor: (status: 'healthy' | 'warning' | 'error' | 'offline' | 'maintenance'): string => {
    switch (status) {
      case 'healthy':
        return Colors.success;
      case 'warning':
        return Colors.warning;
      case 'error':
      case 'offline':
        return Colors.error;
      case 'maintenance':
        return Colors.status.maintenance;
      default:
        return Colors.neutral[500];
    }
  },

  // Get chart color by index
  getChartColor: (index: number): string => {
    const chartColors = Object.values(Colors.chart);
    return chartColors[index % chartColors.length];
  },
};

// Theme-aware color getter
export const getThemedColor = (lightColor: string, darkColor: string, isDark: boolean): string => {
  return isDark ? darkColor : lightColor;
};

// Export default theme colors
export const defaultTheme = {
  colors: Colors,
  utils: colorUtils,
};

export default Colors;
