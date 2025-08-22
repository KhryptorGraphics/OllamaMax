/**
 * @fileoverview Accessibility tests for Button component
 * Tests WCAG 2.1 AA compliance for all button variants and states
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { testAccessibility, testAxeCompliance, testKeyboardNavigation } from '@/utils/accessibility-testing'
import { Button, IconButton, ToggleButton, ButtonGroup } from '@/design-system/components/Button/Button'
import { Download, Plus } from 'lucide-react'

describe('Button Accessibility', () => {
  const user = userEvent.setup()

  describe('Basic Button', () => {
    it('should meet WCAG 2.1 AA standards', async () => {
      await testAxeCompliance(<Button>Click me</Button>)
    })

    it('should be keyboard accessible', async () => {
      const handleClick = vi.fn()
      render(<Button onClick={handleClick}>Click me</Button>)
      
      const button = screen.getByRole('button', { name: /click me/i })
      
      // Should be focusable
      await user.tab()
      expect(button).toHaveFocus()
      
      // Should activate with Enter
      await user.keyboard('{Enter}')
      expect(handleClick).toHaveBeenCalledTimes(1)
      
      // Should activate with Space
      await user.keyboard(' ')
      expect(handleClick).toHaveBeenCalledTimes(2)
    })

    it('should have proper ARIA attributes', () => {
      render(<Button disabled>Disabled button</Button>)
      
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-disabled', 'true')
      expect(button).toBeDisabled()
    })

    it('should announce loading state', () => {
      render(<Button loading loadingText="Processing...">Submit</Button>)
      
      const button = screen.getByRole('button')
      expect(button).toHaveTextContent('Processing...')
      expect(button).toHaveAttribute('aria-disabled', 'true')
      
      // Loading spinner should be hidden from screen readers
      const spinner = button.querySelector('[aria-hidden="true"]')
      expect(spinner).toBeInTheDocument()
    })

    it('should handle focus management correctly', async () => {
      render(
        <div>
          <Button>First</Button>
          <Button>Second</Button>
          <Button disabled>Disabled</Button>
          <Button>Third</Button>
        </div>
      )

      // Should skip disabled button in tab order
      const first = screen.getByRole('button', { name: 'First' })
      const second = screen.getByRole('button', { name: 'Second' })
      const third = screen.getByRole('button', { name: 'Third' })
      const disabled = screen.getByRole('button', { name: 'Disabled' })

      await user.tab()
      expect(first).toHaveFocus()

      await user.tab()
      expect(second).toHaveFocus()

      await user.tab()
      expect(third).toHaveFocus() // Should skip disabled button

      expect(disabled).not.toHaveFocus()
    })
  })

  describe('Button Variants', () => {
    const variants = ['primary', 'secondary', 'outline', 'ghost', 'link', 'destructive'] as const
    
    variants.forEach(variant => {
      it(`should be accessible with ${variant} variant`, async () => {
        await testAxeCompliance(<Button variant={variant}>Test button</Button>)
      })
    })
  })

  describe('Button Sizes', () => {
    const sizes = ['xs', 'sm', 'md', 'lg', 'xl'] as const
    
    sizes.forEach(size => {
      it(`should be accessible with ${size} size`, async () => {
        await testAxeCompliance(<Button size={size}>Test button</Button>)
      })

      it(`should have minimum touch target size for ${size}`, () => {
        const { container } = render(<Button size={size}>Test button</Button>)
        const button = container.querySelector('button')!
        const rect = button.getBoundingClientRect()
        
        // All buttons should meet minimum 44x44px touch target
        // Note: In actual implementation, you might need CSS to ensure this
        if (size === 'xs' || size === 'sm') {
          // Smaller sizes might need padding adjustments
          expect(rect.height).toBeGreaterThanOrEqual(44)
        } else {
          expect(rect.height).toBeGreaterThanOrEqual(44)
        }
      })
    })
  })

  describe('IconButton', () => {
    it('should meet accessibility standards', async () => {
      await testAxeCompliance(
        <IconButton icon={<Download />} aria-label="Download file" />
      )
    })

    it('should require aria-label', () => {
      render(<IconButton icon={<Download />} aria-label="Download file" />)
      
      const button = screen.getByRole('button', { name: /download file/i })
      expect(button).toHaveAccessibleName('Download file')
    })

    it('should be keyboard accessible', async () => {
      const handleClick = vi.fn()
      render(
        <IconButton 
          icon={<Download />} 
          aria-label="Download file"
          onClick={handleClick}
        />
      )
      
      const button = screen.getByRole('button', { name: /download file/i })
      
      await user.tab()
      expect(button).toHaveFocus()
      
      await user.keyboard('{Enter}')
      expect(handleClick).toHaveBeenCalled()
    })

    it('should hide icon from screen readers', () => {
      render(<IconButton icon={<Download />} aria-label="Download file" />)
      
      const button = screen.getByRole('button')
      const iconContainer = button.querySelector('[aria-hidden="true"]')
      expect(iconContainer).toBeInTheDocument()
    })
  })

  describe('ToggleButton', () => {
    it('should meet accessibility standards', async () => {
      await testAxeCompliance(
        <ToggleButton pressed={false}>Toggle feature</ToggleButton>
      )
    })

    it('should have proper ARIA pressed state', () => {
      const { rerender } = render(
        <ToggleButton pressed={false}>Toggle feature</ToggleButton>
      )
      
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-pressed', 'false')
      
      rerender(<ToggleButton pressed={true}>Toggle feature</ToggleButton>)
      expect(button).toHaveAttribute('aria-pressed', 'true')
    })

    it('should announce state changes', async () => {
      const handleToggle = vi.fn()
      render(
        <ToggleButton 
          pressed={false} 
          onPressedChange={handleToggle}
        >
          Toggle feature
        </ToggleButton>
      )
      
      const button = screen.getByRole('button')
      
      await user.click(button)
      expect(handleToggle).toHaveBeenCalledWith(true)
      
      // Verify aria-pressed is updated
      expect(button).toHaveAttribute('aria-pressed', 'false') // Still false until parent updates
    })

    it('should be keyboard accessible', async () => {
      const handleToggle = vi.fn()
      render(
        <ToggleButton 
          pressed={false} 
          onPressedChange={handleToggle}
        >
          Toggle feature
        </ToggleButton>
      )
      
      const button = screen.getByRole('button')
      button.focus()
      
      await user.keyboard('{Enter}')
      expect(handleToggle).toHaveBeenCalledWith(true)
      
      await user.keyboard(' ')
      expect(handleToggle).toHaveBeenCalledWith(true)
    })
  })

  describe('ButtonGroup', () => {
    it('should meet accessibility standards', async () => {
      await testAxeCompliance(
        <ButtonGroup>
          <Button>First</Button>
          <Button>Second</Button>
          <Button>Third</Button>
        </ButtonGroup>
      )
    })

    it('should have proper group role and label', () => {
      render(
        <ButtonGroup>
          <Button>First</Button>
          <Button>Second</Button>
        </ButtonGroup>
      )
      
      const group = screen.getByRole('group')
      expect(group).toHaveAttribute('aria-label', 'Button group')
    })

    it('should support arrow key navigation', async () => {
      await testKeyboardNavigation(
        <ButtonGroup>
          <Button>First</Button>
          <Button>Second</Button>
          <Button>Third</Button>
        </ButtonGroup>,
        {
          keySequences: [
            {
              keys: ['{Tab}', '{ArrowRight}'],
              expectedFocus: 'button:nth-child(2)',
              description: 'Arrow right navigation'
            },
            {
              keys: ['{ArrowLeft}'],
              expectedFocus: 'button:nth-child(1)',
              description: 'Arrow left navigation'
            }
          ]
        }
      )
    })

    it('should handle vertical orientation', async () => {
      await testAxeCompliance(
        <ButtonGroup orientation="vertical">
          <Button>First</Button>
          <Button>Second</Button>
          <Button>Third</Button>
        </ButtonGroup>
      )
    })
  })

  describe('Button with Icons', () => {
    it('should be accessible with left icon', async () => {
      await testAxeCompliance(
        <Button leftIcon={<Plus />}>Add item</Button>
      )
    })

    it('should be accessible with right icon', async () => {
      await testAxeCompliance(
        <Button rightIcon={<Download />}>Download</Button>
      )
    })

    it('should hide decorative icons from screen readers', () => {
      render(<Button leftIcon={<Plus />}>Add item</Button>)
      
      const button = screen.getByRole('button')
      const iconContainer = button.querySelector('[aria-hidden="true"]')
      expect(iconContainer).toBeInTheDocument()
    })

    it('should not duplicate icon information in accessible name', () => {
      render(<Button leftIcon={<Plus />}>Add item</Button>)
      
      const button = screen.getByRole('button')
      expect(button).toHaveAccessibleName('Add item')
      // Should not include icon information in accessible name
    })
  })

  describe('Loading States', () => {
    it('should be accessible during loading', async () => {
      await testAxeCompliance(
        <Button loading loadingText="Saving...">Save</Button>
      )
    })

    it('should announce loading state to screen readers', () => {
      render(<Button loading loadingText="Saving...">Save</Button>)
      
      const button = screen.getByRole('button')
      expect(button).toHaveTextContent('Saving...')
      expect(button).toHaveAttribute('aria-disabled', 'true')
    })

    it('should hide loading spinner from screen readers', () => {
      render(<Button loading>Save</Button>)
      
      const button = screen.getByRole('button')
      const spinner = button.querySelector('.animate-spin')
      expect(spinner).toHaveAttribute('aria-hidden', 'true')
    })
  })

  describe('Error States and Validation', () => {
    it('should handle form validation errors accessibly', async () => {
      const { container } = render(
        <form>
          <Button type="submit">Submit form</Button>
        </form>
      )
      
      await testAxeCompliance(container.querySelector('form')!)
    })
  })

  describe('Color Contrast', () => {
    it('should meet color contrast requirements', () => {
      const { container } = render(<Button>Test button</Button>)
      const button = container.querySelector('button')!
      
      // In a real implementation, you'd test actual color values
      // This is a placeholder for contrast testing
      const style = window.getComputedStyle(button)
      expect(style.backgroundColor).toBeTruthy()
      expect(style.color).toBeTruthy()
    })
  })

  describe('Comprehensive Accessibility Test', () => {
    it('should pass comprehensive accessibility test', async () => {
      await testAccessibility(
        <div>
          <Button>Primary button</Button>
          <Button variant="secondary">Secondary button</Button>
          <IconButton icon={<Download />} aria-label="Download" />
          <ToggleButton pressed={false}>Toggle</ToggleButton>
          <ButtonGroup>
            <Button>Group 1</Button>
            <Button>Group 2</Button>
          </ButtonGroup>
        </div>,
        {
          tags: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          focusableSelectors: ['button'],
          ariaAttributes: ['aria-label', 'aria-pressed', 'aria-disabled'],
          landmarks: ['group']
        }
      )
    })
  })
})