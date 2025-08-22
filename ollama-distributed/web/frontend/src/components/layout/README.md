# Layout Components - Sprint C

Comprehensive layout system for the OllamaMax distributed frontend application.

## Components Overview

### 🏗️ AppLayout
**Main application layout with sidebar and header**

```tsx
import { AppLayout } from '@/components/layout'

function App() {
  return (
    <AppLayout>
      {/* Your page content */}
    </AppLayout>
  )
}
```

**Features:**
- Responsive navigation for all screen sizes
- Theme integration and switching (light/dark/system)
- Notification system integration
- Mobile-first responsive design
- Accessibility compliance (WCAG AA)
- Search functionality in header
- User menu with profile actions

### 📋 PageHeader
**Dynamic page headers with breadcrumbs and actions**

```tsx
import { PageHeader } from '@/components/layout'

function ModelsPage() {
  return (
    <>
      <PageHeader
        title="Models"
        subtitle="Manage and monitor AI models"
        showSearch={true}
        showFilters={true}
        showViewToggle={true}
        onSearch={(query) => handleSearch(query)}
        onFilter={() => openFilters()}
        onExport={() => exportData()}
      />
      {/* Page content */}
    </>
  )
}
```

**Features:**
- Auto-generated breadcrumbs from route
- Dynamic page titles and descriptions
- Search and filter integration
- Export and settings controls
- View toggle (grid/list)
- Sort controls with direction indicators
- Action button slots

### 🧭 Sidebar
**Collapsible sidebar navigation**

```tsx
import { Sidebar } from '@/components/layout'

function Layout() {
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <Sidebar
      collapsed={collapsed}
      mobileOpen={mobileOpen}
      onToggle={() => setCollapsed(!collapsed)}
      onMobileClose={() => setMobileOpen(false)}
    />
  )
}
```

**Features:**
- Collapsible sidebar with smooth animations
- Menu items with icons and notification badges
- Active state management with React Router
- Mobile responsive with overlay
- Tooltips when collapsed
- User profile section
- Grouped navigation items

### 🧭 Navigation
**Navigation configuration and utilities**

```tsx
import { useNavigation, hasPermission } from '@/components/layout'

function CustomNav() {
  const { groups, userMenu, isActive, setActive } = useNavigation('admin')
  
  return (
    <nav>
      {groups.map(group => (
        <div key={group.id}>
          <h3>{group.title}</h3>
          {group.items.map(item => (
            <NavigationItem
              key={item.id}
              item={item}
              active={isActive(item.id)}
              onClick={() => setActive(item.id)}
            />
          ))}
        </div>
      ))}
    </nav>
  )
}
```

**Features:**
- Permission-based menu filtering
- Role-based access control
- Notification badges with pulse animation
- Breadcrumb generation
- Menu item configuration
- Quick actions menu

## Design System Integration

### Theme Support
All components integrate with the design system theme:

```tsx
// Automatic theme application
const Button = styled.button`
  background-color: ${({ theme }) => theme.colors.interactive.primary.default};
  color: ${({ theme }) => theme.colors.text.inverse};
  border-radius: ${({ theme }) => theme.radius.md};
  transition: ${({ theme }) => theme.transitions.colors};
`
```

### Color System
- **Light/Dark mode support** - Automatic switching
- **Semantic colors** - Success, warning, error, info variants
- **Interactive states** - Hover, active, focus, disabled
- **Accessibility** - WCAG AA contrast compliance

### Typography
- **Responsive sizing** - Scales across breakpoints
- **Font weights** - 400 (normal), 500 (medium), 600 (semibold), 700 (bold)
- **Line height** - Optimized for readability
- **Letter spacing** - Proper character spacing

### Spacing & Layout
- **8px grid system** - Consistent spacing units
- **Responsive breakpoints** - Mobile-first approach
- **Flexible layouts** - CSS Grid and Flexbox
- **Container queries** - Component-level responsiveness

## Responsive Design

### Breakpoints
```css
/* Mobile */
@media (max-width: 768px) {
  /* Mobile-specific styles */
}

/* Tablet */
@media (min-width: 769px) and (max-width: 1024px) {
  /* Tablet-specific styles */
}

/* Desktop */
@media (min-width: 1025px) {
  /* Desktop-specific styles */
}
```

### Mobile Features
- **Touch-friendly targets** - Minimum 44px touch targets
- **Swipe gestures** - Natural mobile interactions
- **Responsive navigation** - Collapsible mobile menu
- **Viewport optimization** - Proper scaling and zoom

## Accessibility Features

### WCAG Compliance
- **AA contrast ratios** - 4.5:1 for normal text, 3:1 for large text
- **Keyboard navigation** - Full keyboard accessibility
- **Screen reader support** - ARIA labels and descriptions
- **Focus management** - Visible focus indicators

### Screen Reader Support
```tsx
// Proper ARIA labeling
<button
  aria-label="Toggle sidebar navigation"
  aria-expanded={isExpanded}
  aria-controls="sidebar-navigation"
>
  <Menu />
</button>
```

### Keyboard Navigation
- **Tab order** - Logical tab sequence
- **Escape key** - Close modals and menus
- **Enter/Space** - Activate buttons and links
- **Arrow keys** - Navigate menu items

## Performance Optimizations

### Code Splitting
```tsx
// Lazy load heavy components
const AppLayout = lazy(() => import('./AppLayout'))
const PageHeader = lazy(() => import('./PageHeader'))
```

### Memoization
```tsx
// Prevent unnecessary re-renders
const Navigation = React.memo(({ items, activeItem }) => {
  return <nav>{/* render items */}</nav>
})
```

### Animation Performance
```css
/* Use transform and opacity for 60fps animations */
.sidebar {
  transform: translateX(0);
  transition: transform 300ms ease;
  will-change: transform;
}
```

## State Management

### Theme State
```tsx
import { useTheme } from '@/store/theme'

function ThemedComponent() {
  const { theme, toggleTheme, setMode } = useTheme()
  
  return (
    <button onClick={toggleTheme}>
      Switch to {theme === 'light' ? 'dark' : 'light'} mode
    </button>
  )
}
```

### Navigation State
```tsx
import { useNavigation } from '@/components/layout'

function NavigationMenu() {
  const { 
    groups, 
    activeItem, 
    setActive, 
    hasPermission 
  } = useNavigation(userRole)
  
  return (
    <nav>
      {groups.map(group => (
        hasPermission(group.permission) && (
          <MenuGroup key={group.id} {...group} />
        )
      ))}
    </nav>
  )
}
```

## Testing

### Unit Tests
```tsx
import { render, screen } from '@testing-library/react'
import { AppLayout } from './AppLayout'

test('renders sidebar navigation', () => {
  render(<AppLayout />)
  expect(screen.getByRole('navigation')).toBeInTheDocument()
})
```

### Accessibility Testing
```tsx
import { axe, toHaveNoViolations } from 'jest-axe'

test('has no accessibility violations', async () => {
  const { container } = render(<AppLayout />)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
})
```

### Visual Regression Testing
```tsx
// Storybook visual tests
export default {
  title: 'Layout/AppLayout',
  component: AppLayout,
  parameters: {
    chromatic: { 
      viewports: [320, 768, 1200] 
    }
  }
}
```

## Integration Examples

### Basic App Setup
```tsx
import { AppLayout } from '@/components/layout'
import { ThemeProvider } from '@/theme'
import { BrowserRouter } from 'react-router-dom'

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <AppLayout>
          <Routes>
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/models" element={<Models />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </AppLayout>
      </ThemeProvider>
    </BrowserRouter>
  )
}
```

### Custom Page Layout
```tsx
import { PageHeader } from '@/components/layout'

function CustomPage() {
  return (
    <>
      <PageHeader
        title="Custom Page"
        subtitle="Custom functionality"
        showSearch={true}
        onSearch={(query) => handleSearch(query)}
        actions={
          <Button variant="primary">
            Custom Action
          </Button>
        }
      />
      <main>
        {/* Page content */}
      </main>
    </>
  )
}
```

## Browser Support

- **Chrome 90+** - Full support
- **Safari 14+** - Full support
- **Firefox 88+** - Full support
- **Edge 90+** - Full support

## File Structure

```
src/components/layout/
├── AppLayout.tsx       # Main application layout
├── PageHeader.tsx      # Dynamic page headers
├── Sidebar.tsx         # Collapsible sidebar navigation
├── Navigation.tsx      # Navigation configuration
├── Layout.stories.tsx  # Storybook stories
├── README.md          # This documentation
└── index.ts           # Export index
```

## Contributing

1. Follow the existing code style and patterns
2. Include comprehensive TypeScript types
3. Add Storybook stories for new components
4. Write unit tests with accessibility testing
5. Update documentation for new features
6. Test across all supported browsers and devices

## Related Components

- **Design System** - `/src/design-system/` - Base components
- **Theme System** - `/src/theme/` - Theming and styling
- **Accessibility** - `/src/components/accessibility/` - A11y utilities
- **Common Components** - `/src/components/common/` - Shared utilities