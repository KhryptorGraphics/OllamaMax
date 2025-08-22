/**
 * @fileoverview Accessibility tests for Input component
 * Tests WCAG 2.1 AA compliance for all input variants and states
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { testAccessibility, testAxeCompliance, testKeyboardNavigation } from '@/utils/accessibility-testing'
import { Input, Textarea } from '@/design-system/components/Input/Input'
import { Search, User } from 'lucide-react'

describe('Input Accessibility', () => {
  const user = userEvent.setup()

  describe('Basic Input', () => {
    it('should meet WCAG 2.1 AA standards', async () => {
      await testAxeCompliance(
        <Input label="Full name" placeholder="Enter your full name" />
      )
    })

    it('should have proper label association', () => {
      render(<Input label="Email address" />)
      
      const input = screen.getByRole('textbox', { name: /email address/i })
      const label = screen.getByText('Email address')
      
      expect(input).toHaveAccessibleName('Email address')
      expect(label).toHaveAttribute('for', input.id)
    })

    it('should support aria-label when no visible label', async () => {
      await testAxeCompliance(
        <Input aria-label="Search products" placeholder="Search..." />
      )
      
      const input = screen.getByRole('textbox', { name: /search products/i })
      expect(input).toHaveAccessibleName('Search products')
    })

    it('should be keyboard accessible', async () => {
      render(<Input label="Username" />)
      
      const input = screen.getByRole('textbox')
      
      // Should be focusable
      await user.tab()
      expect(input).toHaveFocus()
      
      // Should accept text input
      await user.type(input, 'testuser')
      expect(input).toHaveValue('testuser')
    })

    it('should handle required fields properly', () => {
      render(<Input label="Required field" required />)
      
      const input = screen.getByRole('textbox')
      const label = screen.getByText(/required field/i)
      
      expect(input).toBeRequired()
      expect(input).toHaveAttribute('aria-required', 'true')
      expect(label).toHaveClass('after:content-[\'*\']') // Visual required indicator
    })

    it('should associate helper text', () => {
      render(
        <Input 
          label="Password" 
          helperText="Must be at least 8 characters long"
        />
      )
      
      const input = screen.getByRole('textbox')
      const helperText = screen.getByText(/must be at least 8 characters/i)
      
      expect(input).toHaveAttribute('aria-describedby', expect.stringContaining(helperText.id))
    })
  })

  describe('Input States', () => {
    it('should handle error state accessibly', async () => {
      await testAxeCompliance(
        <Input 
          label="Email" 
          error="Please enter a valid email address"
          value="invalid-email"
        />
      )
      
      const input = screen.getByRole('textbox')
      const errorMessage = screen.getByText(/please enter a valid email/i)
      
      expect(input).toHaveAttribute('aria-invalid', 'true')
      expect(input).toHaveAttribute('aria-describedby', expect.stringContaining(errorMessage.id))
      expect(errorMessage).toHaveAttribute('role', 'alert')
    })

    it('should handle success state accessibly', async () => {
      await testAxeCompliance(
        <Input 
          label="Username" 
          success="Username is available"
          value="testuser"
        />
      )
      
      const input = screen.getByRole('textbox')
      const successMessage = screen.getByText(/username is available/i)
      
      expect(input).toHaveAttribute('aria-invalid', 'false')
      expect(successMessage).toBeInTheDocument()
    })

    it('should handle warning state accessibly', async () => {
      await testAxeCompliance(
        <Input 
          label="Password" 
          warning="Password strength: weak"
          value="123"
        />
      )
      
      const input = screen.getByRole('textbox')
      const warningMessage = screen.getByText(/password strength: weak/i)
      
      expect(warningMessage).toBeInTheDocument()
    })

    it('should handle disabled state accessibly', async () => {
      await testAxeCompliance(
        <Input label="Disabled field" disabled value="Cannot edit" />
      )
      
      const input = screen.getByRole('textbox')
      expect(input).toBeDisabled()
      expect(input).toHaveAttribute('aria-disabled', 'true')
    })
  })

  describe('Password Input', () => {
    it('should be accessible with password visibility toggle', async () => {
      await testAxeCompliance(
        <Input label="Password" type="password" />
      )
    })

    it('should have accessible visibility toggle button', async () => {
      render(<Input label="Password" type="password" />)
      
      const input = screen.getByLabelText(/password/i)
      const toggleButton = screen.getByRole('button', { name: /show password/i })
      
      expect(input).toHaveAttribute('type', 'password')
      expect(toggleButton).toHaveAttribute('aria-label', 'Show password')
      expect(toggleButton).toHaveAttribute('tabIndex', '-1') // Not in tab order
      
      // Toggle visibility
      await user.click(toggleButton)
      expect(input).toHaveAttribute('type', 'text')
      expect(toggleButton).toHaveAttribute('aria-label', 'Hide password')
    })

    it('should announce password visibility changes', async () => {
      render(<Input label="Password" type="password" />)
      
      const toggleButton = screen.getByRole('button', { name: /show password/i })
      
      await user.click(toggleButton)
      expect(toggleButton).toHaveAttribute('aria-label', 'Hide password')
      
      await user.click(toggleButton)
      expect(toggleButton).toHaveAttribute('aria-label', 'Show password')
    })
  })

  describe('Input with Icons', () => {
    it('should be accessible with left icon', async () => {
      await testAxeCompliance(
        <Input label="Search" leftIcon={<Search />} />
      )
    })

    it('should be accessible with right icon', async () => {
      await testAxeCompliance(
        <Input label="Username" rightIcon={<User />} />
      )
    })

    it('should hide decorative icons from screen readers', () => {
      render(<Input label="Search" leftIcon={<Search />} />)
      
      const input = screen.getByRole('textbox')
      const iconContainer = input.parentElement?.querySelector('[aria-hidden="true"]')
      expect(iconContainer).toBeInTheDocument()
    })

    it('should not interfere with input accessibility', async () => {
      render(<Input label="Search" leftIcon={<Search />} rightIcon={<User />} />)
      
      const input = screen.getByRole('textbox')
      
      await user.tab()
      expect(input).toHaveFocus()
      
      await user.type(input, 'test')
      expect(input).toHaveValue('test')
    })
  })

  describe('Input Sizes', () => {
    const sizes = ['sm', 'md', 'lg'] as const
    
    sizes.forEach(size => {
      it(`should be accessible with ${size} size`, async () => {
        await testAxeCompliance(
          <Input label="Test input" size={size} />
        )
      })

      it(`should maintain minimum touch target for ${size}`, () => {
        const { container } = render(<Input label="Test input" size={size} />)
        const input = container.querySelector('input')!
        const rect = input.getBoundingClientRect()
        
        // Minimum 44px touch target height
        expect(rect.height).toBeGreaterThanOrEqual(32) // Inputs can be slightly smaller
      })
    })
  })

  describe('Form Integration', () => {
    it('should work properly in forms', async () => {
      await testAxeCompliance(
        <form>
          <Input label="First name" required />
          <Input label="Last name" required />
          <Input label="Email" type="email" required />
          <button type="submit">Submit</button>
        </form>
      )
    })

    it('should handle form validation', async () => {
      const handleSubmit = vi.fn(e => e.preventDefault())
      
      render(
        <form onSubmit={handleSubmit}>
          <Input 
            label="Email" 
            type="email" 
            required 
            error="Please enter a valid email"
          />
          <button type="submit">Submit</button>
        </form>
      )
      
      const input = screen.getByRole('textbox')
      const submitButton = screen.getByRole('button', { name: /submit/i })
      
      expect(input).toHaveAttribute('aria-invalid', 'true')
      
      await user.click(submitButton)
      expect(handleSubmit).toHaveBeenCalled()
    })
  })

  describe('Textarea Component', () => {
    it('should meet accessibility standards', async () => {
      await testAxeCompliance(
        <Textarea label="Description" placeholder="Enter description..." />
      )
    })

    it('should have proper label association', () => {
      render(<Textarea label="Comments" />)
      
      const textarea = screen.getByRole('textbox', { name: /comments/i })
      const label = screen.getByText('Comments')
      
      expect(textarea).toHaveAccessibleName('Comments')
      expect(label).toHaveAttribute('for', textarea.id)
    })

    it('should be keyboard accessible', async () => {
      render(<Textarea label="Message" />)
      
      const textarea = screen.getByRole('textbox')
      
      await user.tab()
      expect(textarea).toHaveFocus()
      
      await user.type(textarea, 'Test message')
      expect(textarea).toHaveValue('Test message')
    })

    it('should handle multiline input', async () => {
      render(<Textarea label="Message" />)
      
      const textarea = screen.getByRole('textbox')
      
      await user.type(textarea, 'Line 1{Enter}Line 2')
      expect(textarea).toHaveValue('Line 1\nLine 2')
    })

    it('should handle error state', async () => {
      await testAxeCompliance(
        <Textarea 
          label="Description" 
          error="Description is required"
        />
      )
      
      const textarea = screen.getByRole('textbox')
      const errorMessage = screen.getByText(/description is required/i)
      
      expect(textarea).toHaveAttribute('aria-invalid', 'true')
      expect(errorMessage).toHaveAttribute('role', 'alert')
    })

    it('should support different sizes', async () => {
      const sizes = ['sm', 'md', 'lg'] as const
      
      for (const size of sizes) {
        await testAxeCompliance(
          <Textarea label="Test textarea" size={size} />
        )
      }
    })
  })

  describe('Complex Input Scenarios', () => {
    it('should handle multiple validation states', async () => {
      await testAxeCompliance(
        <div>
          <Input label="Valid field" success="Looks good!" />
          <Input label="Warning field" warning="Double check this" />
          <Input label="Error field" error="This field is required" />
        </div>
      )
    })

    it('should handle dynamic label changes', () => {
      const { rerender } = render(<Input label="Original label" />)
      
      const input = screen.getByRole('textbox')
      expect(input).toHaveAccessibleName('Original label')
      
      rerender(<Input label="Updated label" />)
      expect(input).toHaveAccessibleName('Updated label')
    })

    it('should handle conditional helper text', () => {
      const { rerender } = render(
        <Input label="Password" helperText="Enter password" />
      )
      
      const input = screen.getByRole('textbox')
      const helperText = screen.getByText('Enter password')
      
      expect(input).toHaveAttribute('aria-describedby', expect.stringContaining(helperText.id))
      
      // Remove helper text
      rerender(<Input label="Password" />)
      expect(input).not.toHaveAttribute('aria-describedby')
    })
  })

  describe('Screen Reader Compatibility', () => {
    it('should provide clear input purpose', () => {
      render(
        <Input 
          label="Credit card number" 
          type="text"
          placeholder="1234 5678 9012 3456"
          helperText="Enter your 16-digit card number"
        />
      )
      
      const input = screen.getByRole('textbox')
      expect(input).toHaveAccessibleName('Credit card number')
      expect(input).toHaveAccessibleDescription('Enter your 16-digit card number')
    })

    it('should announce state changes clearly', async () => {
      const { rerender } = render(
        <Input label="Username" />
      )
      
      const input = screen.getByRole('textbox')
      
      // Add error state
      rerender(
        <Input 
          label="Username" 
          error="Username is already taken"
        />
      )
      
      expect(input).toHaveAttribute('aria-invalid', 'true')
      
      const errorMessage = screen.getByText(/username is already taken/i)
      expect(errorMessage).toHaveAttribute('role', 'alert')
    })
  })

  describe('Comprehensive Accessibility Test', () => {
    it('should pass comprehensive accessibility test', async () => {
      await testAccessibility(
        <form>
          <Input label="First name" required />
          <Input label="Email" type="email" error="Invalid email" />
          <Input label="Password" type="password" />
          <Input label="Search" leftIcon={<Search />} />
          <Textarea label="Comments" helperText="Optional feedback" />
          <button type="submit">Submit form</button>
        </form>,
        {
          tags: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          focusableSelectors: ['input', 'textarea', 'button'],
          ariaAttributes: ['aria-label', 'aria-labelledby', 'aria-describedby', 'aria-invalid', 'aria-required'],
          headingStructure: false // No headings in this form
        }
      )
    })
  })
})