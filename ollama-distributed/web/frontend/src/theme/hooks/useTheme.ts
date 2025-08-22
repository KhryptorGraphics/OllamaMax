import { useTheme as useStyledTheme } from 'styled-components';
import { Theme } from '../theme';

// Re-export useTheme from providers to avoid circular imports  
export { useTheme as useThemeContext } from '../providers/ThemeProvider';

// Hook to get styled-components theme
export const useTheme = (): Theme => {
  return useStyledTheme();
};

// Hook to get both context and theme
export const useThemeState = () => {
  const context = useThemeContext();
  const theme = useTheme();
  
  return {
    ...context,
    theme
  };
};