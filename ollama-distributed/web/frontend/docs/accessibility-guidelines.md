# Accessibility Guidelines for OllamaMax Frontend

## Overview

This document provides comprehensive accessibility guidelines for developing components and features in the OllamaMax distributed AI platform frontend. All components must meet WCAG 2.1 AA standards.

## Core Accessibility Principles

### 1. Perceivable
- All information and UI components must be presentable to users in ways they can perceive
- Provide text alternatives for non-text content
- Offer captions and other alternatives for multimedia
- Create content that can be presented in different ways without losing meaning
- Make it easier for users to see and hear content

### 2. Operable
- All UI components and navigation must be operable
- Make all functionality available from a keyboard
- Give users enough time to read and use content
- Don't use content that causes seizures or physical reactions
- Help users navigate and find content

### 3. Understandable
- Information and operation of UI must be understandable
- Make text readable and understandable
- Make content appear and operate in predictable ways
- Help users avoid and correct mistakes

### 4. Robust
- Content must be robust enough for interpretation by a wide variety of user agents
- Maximize compatibility with assistive technologies

## Implementation Standards

### Required ARIA Attributes

#### For Interactive Elements
```tsx
// Buttons
<button
  aria-label="Close dialog" // Required for icon-only buttons
  aria-pressed="false"      // For toggle buttons
  aria-expanded="false"     // For expandable buttons
  aria-describedby="help-text" // When additional description needed
>

// Links
<a
  aria-label="Download report"    // When link text isn't descriptive
  aria-describedby="file-info"   // For additional context
  aria-current="page"            // For current page in navigation
>

// Form Controls
<input
  aria-label="Email address"     // When no visible label
  aria-labelledby="email-label"  // Reference to label element
  aria-describedby="email-help"  // Reference to help text
  aria-invalid="true"            // When validation fails
  aria-required="true"           // For required fields
/>
```

#### For Complex UI Patterns
```tsx
// Modal/Dialog
<div
  role="dialog"
  aria-modal="true"
  aria-labelledby="modal-title"
  aria-describedby="modal-description"
>

// Tabs
<div role="tablist" aria-label="Settings sections">
  <button role="tab" aria-selected="true" aria-controls="panel1">
  <button role="tab" aria-selected="false" aria-controls="panel2">
</div>

// Menu
<ul role="menu" aria-label="Actions">
  <li role="menuitem">Edit</li>
  <li role="menuitem">Delete</li>
</ul>

// Live Regions
<div aria-live="polite" aria-atomic="true">
  Status updates appear here
</div>

<div role="alert">
  Error messages appear here
</div>
```

### Keyboard Navigation Requirements

#### Tab Order
- All interactive elements must be included in logical tab order
- Skip links should be provided for main content areas
- Tab order should follow visual layout and logical flow

#### Keyboard Shortcuts
```tsx
// Standard shortcuts that must be supported:
// Tab/Shift+Tab: Navigate between focusable elements
// Enter/Space: Activate buttons and controls
// Escape: Close dialogs, dropdowns, or cancel actions
// Arrow keys: Navigate within component groups (menus, tabs, etc.)
// Home/End: Move to first/last element in a group

// Example implementation:
const handleKeyDown = (event: KeyboardEvent) => {
  switch (event.key) {
    case 'Escape':
      onClose()
      break
    case 'Tab':
      if (event.shiftKey) {
        // Handle reverse tab navigation
      } else {
        // Handle forward tab navigation
      }
      break
    case 'ArrowDown':
    case 'ArrowUp':
      // Handle arrow navigation
      event.preventDefault()
      break
  }
}
```

#### Focus Management
```tsx
// Focus trapping in modals
const trapFocus = (container: HTMLElement) => {
  const focusableElements = getFocusableElements(container)
  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]

  container.addEventListener('keydown', (e) => {
    if (e.key === 'Tab') {
      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          e.preventDefault()
          lastElement.focus()
        }
      } else {
        if (document.activeElement === lastElement) {
          e.preventDefault()
          firstElement.focus()
        }
      }
    }
  })
}

// Focus restoration
const saveFocusBeforeModal = () => {
  return document.activeElement as HTMLElement
}

const restoreFocusAfterModal = (previouslyFocused: HTMLElement) => {
  if (previouslyFocused && previouslyFocused.focus) {
    previouslyFocused.focus()
  }
}
```

### Screen Reader Support

#### Semantic HTML
```tsx
// Use semantic HTML elements whenever possible
<nav aria-label="Main navigation">
  <ul>
    <li><a href="/dashboard">Dashboard</a></li>
    <li><a href="/models" aria-current="page">Models</a></li>
  </ul>
</nav>

<main>
  <h1>Page Title</h1>
  <article>
    <h2>Section Title</h2>
    <p>Content...</p>
  </article>
</main>

<aside aria-label="Related links">
  <h3>Quick Links</h3>
  <ul>...</ul>
</aside>
```

#### Heading Structure
```tsx
// Maintain proper heading hierarchy
<h1>Main Page Title</h1>       // Only one h1 per page
  <h2>Main Section</h2>
    <h3>Subsection</h3>
    <h3>Another Subsection</h3>
  <h2>Another Main Section</h2>
    <h3>Subsection</h3>
      <h4>Sub-subsection</h4>
```

#### Status and Error Announcements
```tsx
// Use live regions for dynamic content
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>

<div role="alert" aria-atomic="true">
  {errorMessage}
</div>

// Custom announcements
const { announce } = useAccessibilityContext()

// Announce state changes
announce('Data loaded successfully', 'polite')
announce('Form submission failed', 'assertive')
```

### Color and Contrast

#### Minimum Requirements
- **Normal text**: 4.5:1 contrast ratio
- **Large text** (18pt+ or 14pt+ bold): 3:1 contrast ratio
- **UI components**: 3:1 contrast ratio for borders, focus indicators
- **Graphical objects**: 3:1 contrast ratio for meaningful graphics

#### Implementation
```css
/* High contrast mode support */
@media (prefers-contrast: high) {
  .button {
    border: 2px solid currentColor;
    background: ButtonFace;
    color: ButtonText;
  }
}

/* Forced colors mode support */
@media (forced-colors: active) {
  .custom-button {
    forced-color-adjust: none;
    border: 1px solid ButtonBorder;
    background: ButtonFace;
    color: ButtonText;
  }
}
```

### Touch and Motor Accessibility

#### Touch Target Sizes
- Minimum 44x44 pixels for all interactive elements
- Adequate spacing between touch targets (8px minimum)

```css
.touch-target {
  min-height: 44px;
  min-width: 44px;
  margin: 4px; /* Provides 8px spacing between targets */
}
```

#### Motion Preferences
```css
/* Respect reduced motion preferences */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

## Component-Specific Guidelines

### Buttons
```tsx
// ✅ Good button implementation
<Button
  variant="primary"
  size="md"
  disabled={isLoading}
  aria-disabled={isLoading}
  aria-describedby={helpTextId}
  onClick={handleSubmit}
>
  {isLoading ? 'Saving...' : 'Save Changes'}
</Button>

// ✅ Icon button with proper labeling
<IconButton
  icon={<TrashIcon />}
  aria-label="Delete item"
  onClick={handleDelete}
/>

// ✅ Toggle button with state
<ToggleButton
  pressed={isToggled}
  aria-pressed={isToggled}
  onPressedChange={setIsToggled}
>
  {isToggled ? 'Hide' : 'Show'} details
</ToggleButton>
```

### Forms
```tsx
// ✅ Accessible form implementation
<form onSubmit={handleSubmit}>
  <fieldset>
    <legend>Personal Information</legend>
    
    <Input
      label="Full Name"
      required
      aria-required="true"
      error={errors.name}
      aria-invalid={!!errors.name}
      aria-describedby={errors.name ? 'name-error' : 'name-help'}
    />
    
    {errors.name && (
      <div id="name-error" role="alert" className="error">
        {errors.name}
      </div>
    )}
    
    <div id="name-help" className="help-text">
      Enter your first and last name
    </div>
  </fieldset>
  
  <Button type="submit" disabled={isSubmitting}>
    {isSubmitting ? 'Submitting...' : 'Submit Form'}
  </Button>
</form>
```

### Modals and Dialogs
```tsx
// ✅ Accessible modal implementation
<Modal
  isOpen={isOpen}
  onClose={onClose}
  aria-labelledby="modal-title"
  aria-describedby="modal-description"
>
  <div role="dialog" aria-modal="true">
    <h2 id="modal-title">Confirm Deletion</h2>
    <p id="modal-description">
      Are you sure you want to delete this item? This action cannot be undone.
    </p>
    
    <div className="modal-actions">
      <Button variant="destructive" onClick={handleConfirm}>
        Delete
      </Button>
      <Button variant="outline" onClick={onClose}>
        Cancel
      </Button>
    </div>
  </div>
</Modal>
```

### Data Tables
```tsx
// ✅ Accessible table implementation
<table role="table" aria-label="User list">
  <caption className="sr-only">
    List of users with their roles and status
  </caption>
  
  <thead>
    <tr>
      <th scope="col">Name</th>
      <th scope="col">Email</th>
      <th scope="col">Role</th>
      <th scope="col">Status</th>
      <th scope="col">Actions</th>
    </tr>
  </thead>
  
  <tbody>
    {users.map(user => (
      <tr key={user.id}>
        <th scope="row">{user.name}</th>
        <td>{user.email}</td>
        <td>{user.role}</td>
        <td>
          <Badge variant={user.isActive ? 'success' : 'secondary'}>
            {user.isActive ? 'Active' : 'Inactive'}
          </Badge>
        </td>
        <td>
          <Button
            size="sm"
            aria-label={`Edit ${user.name}`}
            onClick={() => handleEdit(user.id)}
          >
            Edit
          </Button>
        </td>
      </tr>
    ))}
  </tbody>
</table>
```

## Testing Guidelines

### Automated Testing
```tsx
// Use accessibility testing utilities
import { testAccessibility, testAxeCompliance } from '@/utils/accessibility-testing'

describe('Component Accessibility', () => {
  it('should meet WCAG 2.1 AA standards', async () => {
    await testAxeCompliance(<MyComponent />)
  })
  
  it('should support keyboard navigation', async () => {
    await testKeyboardNavigation(<MyComponent />)
  })
  
  it('should work with screen readers', () => {
    testScreenReaderCompatibility(<MyComponent />)
  })
})
```

### Manual Testing Checklist

#### Keyboard Testing
- [ ] All interactive elements are keyboard accessible
- [ ] Tab order is logical and matches visual layout
- [ ] Focus indicators are clearly visible
- [ ] Keyboard shortcuts work as expected
- [ ] Focus is properly trapped in modals
- [ ] Focus is restored when modals close

#### Screen Reader Testing
- [ ] Content is announced in logical order
- [ ] Form fields have proper labels
- [ ] Error messages are announced
- [ ] State changes are announced
- [ ] Images have appropriate alt text
- [ ] Heading structure is logical

#### Visual Testing
- [ ] Color contrast meets minimum requirements
- [ ] Content is usable at 200% zoom
- [ ] Focus indicators are visible in high contrast mode
- [ ] Touch targets are at least 44x44 pixels
- [ ] Content reflows properly on mobile devices

### Browser and Assistive Technology Testing

#### Required Testing Matrix
- **Browsers**: Chrome, Firefox, Safari, Edge
- **Screen Readers**: 
  - NVDA with Firefox (Windows)
  - JAWS with Chrome (Windows)
  - VoiceOver with Safari (macOS)
  - TalkBack with Chrome (Android)
- **Keyboard**: Test with keyboard-only navigation
- **Voice Control**: Test with Dragon NaturallySpeaking or Voice Control

## Development Workflow

### Pre-Development
1. Review designs for accessibility considerations
2. Identify required ARIA patterns
3. Plan keyboard navigation flow
4. Consider screen reader user experience

### During Development
1. Use semantic HTML first
2. Implement keyboard navigation
3. Add appropriate ARIA attributes
4. Test with accessibility tools
5. Run automated accessibility tests

### Code Review
1. Verify ARIA usage is correct
2. Check keyboard navigation implementation
3. Validate color contrast
4. Ensure proper error handling
5. Review with accessibility testing results

### Quality Assurance
1. Manual keyboard testing
2. Screen reader testing
3. High contrast mode testing
4. Mobile accessibility testing
5. Zoom testing (up to 200%)

## Common Accessibility Pitfalls

### ❌ Common Mistakes

```tsx
// Missing alt text
<img src="chart.png" />

// Poor button labeling
<button onClick={save}>💾</button>

// No focus management
<Modal isOpen={true}>
  <input autoFocus /> // Not sufficient for focus management
</Modal>

// Missing error association
<input type="email" />
<div className="error">Invalid email</div> // Not associated

// Poor heading structure
<h1>Page Title</h1>
<h3>Section</h3> // Skips h2

// No keyboard support
<div onClick={handleClick}>Clickable</div> // Not keyboard accessible
```

### ✅ Correct Implementations

```tsx
// Proper alt text
<img src="chart.png" alt="Revenue increased 15% over last quarter" />

// Accessible button
<Button aria-label="Save document" onClick={save}>
  <SaveIcon aria-hidden="true" />
</Button>

// Proper focus management
<Modal
  isOpen={true}
  onOpen={() => trapFocus(modalRef.current)}
  onClose={() => restoreFocus()}
>

// Associated error message
<Input
  type="email"
  aria-invalid={!!error}
  aria-describedby={error ? 'email-error' : undefined}
/>
{error && (
  <div id="email-error" role="alert">
    {error}
  </div>
)}

// Proper heading structure
<h1>Page Title</h1>
<h2>Main Section</h2>
<h3>Subsection</h3>

// Keyboard accessible custom control
<div
  role="button"
  tabIndex={0}
  onKeyDown={handleKeyDown}
  onClick={handleClick}
  aria-label="Custom action"
>
```

## Resources and Tools

### Development Tools
- **axe DevTools**: Browser extension for accessibility testing
- **WAVE**: Web accessibility evaluation tool
- **Lighthouse**: Includes accessibility audit
- **Color Oracle**: Color blindness simulator
- **Accessibility Insights**: Microsoft's accessibility testing tools

### Testing Libraries
- **@axe-core/react**: React integration for axe-core
- **jest-axe**: Jest matcher for accessibility testing
- **@testing-library/jest-dom**: Enhanced assertions for DOM testing

### Documentation
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [WAI-ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [MDN Accessibility Guide](https://developer.mozilla.org/en-US/docs/Web/Accessibility)

### Screen Readers
- **NVDA**: Free screen reader for Windows
- **VoiceOver**: Built into macOS and iOS
- **TalkBack**: Built into Android
- **JAWS**: Commercial screen reader for Windows

## Conclusion

Accessibility is not optional—it's a fundamental requirement for creating inclusive software. By following these guidelines and incorporating accessibility testing into your development workflow, you ensure that the OllamaMax platform is usable by everyone, regardless of their abilities or assistive technologies they may use.

Remember: Good accessibility benefits all users, not just those with disabilities. Clear navigation, logical structure, and robust keyboard support make the application better for everyone.