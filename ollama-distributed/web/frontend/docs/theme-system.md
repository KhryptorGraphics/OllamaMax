# OllamaMax Theme System Documentation

A comprehensive design system with dark mode support, responsive utilities, and semantic tokens for the OllamaMax Distributed frontend.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Design Tokens](#design-tokens)
3. [Dark Mode](#dark-mode)
4. [Responsive Design](#responsive-design)
5. [Component Theming](#component-theming)
6. [Utilities](#utilities)
7. [Examples](#examples)

## Quick Start

### Installation & Setup

```tsx
import React from 'react';
import { ThemeProvider, GlobalStyles } from '@/theme';

function App() {
  return (
    <ThemeProvider defaultTheme="system">
      <GlobalStyles />
      {/* Your app components */}
    </ThemeProvider>
  );
}
```

### Basic Usage

```tsx
import styled from 'styled-components';
import { color, space, radius, shadow } from '@/theme/utils';

const Card = styled.div`
  background-color: ${color('surface.elevated')};
  padding: ${space(4)};
  border-radius: ${radius('lg')};
  box-shadow: ${shadow('sm')};
  color: ${color('text.primary')};
`;
```

## Design Tokens

### Color System

The theme uses a semantic color system with both light and dark variants:

```tsx
// Semantic colors (adapt to theme)
background.primary    // Main background
background.secondary  // Secondary background
text.primary         // Primary text
text.secondary       // Secondary text
border.default       // Default borders
surface.elevated     // Cards, modals

// Raw colors (static)
theme.raw.colors.primary[500]  // #3b82f6
theme.raw.colors.neutral[100]  // #f1f5f9
```

### Typography

```tsx
// Font families
fontFamily.sans      // Inter, system fonts
fontFamily.mono      // JetBrains Mono, monospace
fontFamily.display   // Cal Sans, display fonts

// Font sizes (tuple: [size, { lineHeight }])
fontSize.xs          // [0.75rem, { lineHeight: '1rem' }]
fontSize.base        // [1rem, { lineHeight: '1.5rem' }]
fontSize['2xl']      // [1.5rem, { lineHeight: '2rem' }]

// Font weights
fontWeight.normal    // 400
fontWeight.semibold  // 600
fontWeight.bold      // 700
```

### Spacing

```tsx
// Spacing scale (rem-based)
spacing[1]   // 0.25rem (4px)
spacing[4]   // 1rem (16px)
spacing[8]   // 2rem (32px)

// Semantic spacing
layoutSpacing.component.md  // 16px
layoutSpacing.section.lg    // 96px
```

### Shadows & Elevation

```tsx
// Basic shadows
shadows.sm   // Subtle shadow
shadows.md   // Standard shadow
shadows.lg   // Prominent shadow

// Elevation levels (0-5)
elevation[1] // Cards, buttons
elevation[3] // Modals, dropdowns
elevation[5] // Floating elements
```

## Dark Mode

### Theme Provider

```tsx
import { ThemeProvider } from '@/theme';

<ThemeProvider 
  defaultTheme="system"  // 'light' | 'dark' | 'system'
  storageKey="my-theme"  // Custom storage key
>
  <App />
</ThemeProvider>
```

### Theme Toggle Component

```tsx
import { ThemeToggle } from '@/theme';

// Icon toggle
<ThemeToggle variant="icon" size="md" />

// Dropdown with options
<ThemeToggle variant="dropdown" size="lg" />
```

### Using Theme Context

```tsx
import { useThemeContext } from '@/theme';

function MyComponent() {
  const { mode, effectiveTheme, setTheme, toggleTheme } = useThemeContext();
  
  return (
    <div>
      <p>Current mode: {mode}</p>
      <p>Effective theme: {effectiveTheme}</p>
      <button onClick={toggleTheme}>Toggle Theme</button>
      <button onClick={() => setTheme('dark')}>Set Dark</button>
    </div>
  );
}
```

## Responsive Design

### Breakpoints

```tsx
breakpoints = {
  xs: '475px',    // Mobile small
  sm: '640px',    // Mobile large  
  md: '768px',    // Tablet
  lg: '1024px',   // Desktop small
  xl: '1280px',   // Desktop large
  '2xl': '1536px' // Desktop extra large
}
```

### Responsive Utilities

```tsx
import { responsive } from '@/theme/utils';

const ResponsiveComponent = styled.div`
  font-size: 14px;
  
  ${responsive.sm('font-size: 16px;')}
  ${responsive.lg('font-size: 18px;')}
  
  // Device-specific
  ${responsive.mobile('padding: 8px;')}
  ${responsive.desktop('padding: 16px;')}
  
  // Motion-aware
  ${responsive.motion('transition: all 200ms;')}
  ${responsive.reduceMotion('transition: none;')}
`;
```

### Media Query Hooks

```tsx
import { useBreakpoint, useMediaQuery } from '@/theme';

function ResponsiveComponent() {
  const bp = useBreakpoint();
  const isMobile = useMediaQuery('(max-width: 640px)');
  
  return (
    <div>
      <p>Current: {bp.current}</p>
      <p>Is mobile: {bp.isMobile}</p>
      <p>Is desktop: {bp.isDesktop}</p>
    </div>
  );
}
```

## Component Theming

### Basic Styled Component

```tsx
import styled from 'styled-components';
import { color, space, radius } from '@/theme/utils';

const Button = styled.button<{ variant?: 'primary' | 'secondary' }>`
  padding: ${space(2)} ${space(4)};
  border-radius: ${radius('md')};
  border: 1px solid transparent;
  cursor: pointer;
  
  ${({ variant = 'primary', theme }) => {
    const colors = theme.colors.button[variant];
    return `
      background-color: ${colors.background};
      color: ${colors.text};
      border-color: ${colors.border};
      
      &:hover {
        background-color: ${colors.backgroundHover};
      }
    `;
  }}
  
  &:focus {
    outline: 2px solid ${color('border.focus')};
    outline-offset: 2px;
  }
`;
```

### Responsive Props

```tsx
import { responsiveProp } from '@/theme/utils';

const FlexBox = styled.div<{
  gap?: ResponsiveValues<number> | number;
  direction?: ResponsiveValues<string> | string;
}>`
  display: flex;
  
  ${({ gap }) => gap && responsiveProp(gap, (value) => `gap: ${space(value)};`)}
  ${({ direction }) => direction && responsiveProp(direction, (value) => `flex-direction: ${value};`)}
`;

// Usage
<FlexBox 
  gap={{ base: 2, md: 4, lg: 6 }}
  direction={{ base: 'column', md: 'row' }}
>
  Content
</FlexBox>
```

### Theme-Aware Components

```tsx
const ThemedCard = styled.div`
  background-color: ${color('surface.elevated')};
  border: 1px solid ${color('border.default')};
  border-radius: ${radius('lg')};
  padding: ${space(4)};
  box-shadow: ${shadow('sm')};
  
  // Use raw colors for specific needs
  &::before {
    background: linear-gradient(
      135deg,
      ${({ theme }) => theme.raw.colors.primary[500]},
      ${({ theme }) => theme.raw.colors.secondary[500]}
    );
  }
`;
```

## Utilities

### Common Patterns

```tsx
import { 
  focusRing, 
  truncate, 
  visuallyHidden, 
  buttonReset,
  motionSafe,
  patterns 
} from '@/theme/utils';

const Component = styled.div`
  ${patterns.flexCenter}  // Flex center alignment
  ${patterns.stack('1rem')}  // Vertical stack with gap
  
  button {
    ${buttonReset}
    ${focusRing()}
  }
  
  .truncated {
    ${truncate(2)}  // Truncate to 2 lines
  }
  
  .sr-only {
    ${visuallyHidden}
  }
  
  .animated {
    ${motionSafe(`
      transition: transform 200ms ease;
      &:hover { transform: scale(1.05); }
    `)}
  }
`;
```

### Grid System

```tsx
import { gridSystem } from '@/theme/utils';

const Container = styled.div`
  ${gridSystem.container('xl')}  // Max-width container
`;

const Row = styled.div`
  ${gridSystem.row}
`;

const Column = styled.div`
  ${gridSystem.col(6)}  // 6/12 columns
`;
```

## Examples

### Complete Component Example

```tsx
import React from 'react';
import styled from 'styled-components';
import { 
  color, 
  space, 
  radius, 
  shadow, 
  responsive, 
  focusRing,
  motionSafe 
} from '@/theme/utils';

interface CardProps {
  title: string;
  children: React.ReactNode;
  variant?: 'default' | 'highlighted';
}

const CardContainer = styled.div<{ variant: string }>`
  background-color: ${color('surface.elevated')};
  border: 1px solid ${color('border.default')};
  border-radius: ${radius('lg')};
  padding: ${space(4)};
  box-shadow: ${shadow('sm')};
  
  ${({ variant }) => variant === 'highlighted' && `
    border-color: ${color('primary.500')};
    box-shadow: ${shadow('primary')};
  `}
  
  ${motionSafe(`
    transition: all 200ms ease;
    
    &:hover {
      box-shadow: ${shadow('md')};
      transform: translateY(-2px);
    }
  `)}
  
  ${responsive.mobile(`
    padding: ${space(3)};
    border-radius: ${radius('md')};
  `)}
`;

const CardTitle = styled.h3`
  margin: 0 0 ${space(3)} 0;
  color: ${color('text.primary')};
  font-size: ${({ theme }) => theme.typography.fontSize.lg[0]};
  font-weight: ${({ theme }) => theme.typography.fontWeight.semibold};
`;

const CardContent = styled.div`
  color: ${color('text.secondary')};
  line-height: ${({ theme }) => theme.typography.lineHeight.relaxed};
`;

export const Card: React.FC<CardProps> = ({ 
  title, 
  children, 
  variant = 'default' 
}) => (
  <CardContainer variant={variant}>
    <CardTitle>{title}</CardTitle>
    <CardContent>{children}</CardContent>
  </CardContainer>
);
```

### App-level Integration

```tsx
import React from 'react';
import { ThemeProvider, GlobalStyles, ThemeToggle } from '@/theme';
import { Card } from './components/Card';

function App() {
  return (
    <ThemeProvider defaultTheme="system">
      <GlobalStyles />
      
      <header style={{ 
        display: 'flex', 
        justifyContent: 'space-between',
        padding: '1rem'
      }}>
        <h1>My App</h1>
        <ThemeToggle variant="dropdown" />
      </header>
      
      <main style={{ padding: '1rem' }}>
        <Card title="Welcome" variant="highlighted">
          <p>This card uses the theme system for consistent styling.</p>
        </Card>
      </main>
    </ThemeProvider>
  );
}

export default App;
```

## Best Practices

1. **Use semantic colors** over raw colors for theme consistency
2. **Prefer responsive utilities** over hardcoded breakpoints
3. **Implement motion-safe animations** for accessibility
4. **Use focus rings** on interactive elements
5. **Test both light and dark modes** during development
6. **Follow the spacing scale** for consistent layouts
7. **Leverage elevation levels** for proper UI hierarchy

## TypeScript Support

The theme system is fully typed with TypeScript:

```tsx
import { Theme } from '@/theme';

// Theme is automatically inferred in styled-components
const Component = styled.div`
  color: ${({ theme }) => theme.colors.text.primary}; // ✅ Typed
`;

// Use StyledProps for custom components
import { StyledProps } from '@/theme/utils';

const CustomComponent = styled.div<StyledProps<{ variant: string }>>`
  // Full theme typing available
`;
```