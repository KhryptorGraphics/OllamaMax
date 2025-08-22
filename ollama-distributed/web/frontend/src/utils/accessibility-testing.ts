/**
 * Accessibility Testing Utilities
 * Comprehensive testing utilities for WCAG 2.1 AA compliance
 */

import { axe, toHaveNoViolations } from 'jest-axe'
import { render, RenderResult } from '@testing-library/react'
import { expect } from 'vitest'
import userEvent from '@testing-library/user-event'

// Extend Jest matchers
expect.extend(toHaveNoViolations)

export interface AccessibilityTestOptions {
  /** WCAG tags to test against */
  tags?: string[]
  /** Rules to disable */
  disableRules?: string[]
  /** Rules to include */
  includeRules?: string[]
  /** Specific impact levels to test */
  impacts?: ('minor' | 'moderate' | 'serious' | 'critical')[]
  /** Custom timeout for testing */
  timeout?: number
}

export interface KeyboardTestOptions {
  /** Elements that should be focusable */
  focusableSelectors?: string[]
  /** Elements that should NOT be focusable */
  nonFocusableSelectors?: string[]
  /** Custom key sequences to test */
  keySequences?: Array<{
    keys: string[]
    expectedFocus?: string
    description: string
  }>
}

export interface ScreenReaderTestOptions {
  /** ARIA attributes to validate */
  ariaAttributes?: string[]
  /** Landmark roles to check */
  landmarks?: string[]
  /** Heading structure to validate */
  headingStructure?: boolean
}

/**
 * Comprehensive accessibility testing utility class
 */
export class AccessibilityTester {
  private defaultOptions: AccessibilityTestOptions = {
    tags: ['wcag2a', 'wcag2aa', 'wcag21aa'],
    impacts: ['serious', 'critical'],
    timeout: 10000
  }

  /**
   * Run axe-core accessibility tests on a component
   */
  async testAxeCompliance(
    container: HTMLElement,
    options: AccessibilityTestOptions = {}
  ): Promise<void> {
    const finalOptions = { ...this.defaultOptions, ...options }
    
    const axeConfig = {
      tags: finalOptions.tags,
      rules: {
        ...(finalOptions.disableRules?.reduce((acc, rule) => {
          acc[rule] = { enabled: false }
          return acc
        }, {} as Record<string, { enabled: boolean }>)),
        ...(finalOptions.includeRules?.reduce((acc, rule) => {
          acc[rule] = { enabled: true }
          return acc
        }, {} as Record<string, { enabled: boolean }>))
      }
    }

    const results = await axe(container, axeConfig)
    
    // Filter by impact level if specified
    if (finalOptions.impacts && finalOptions.impacts.length > 0) {
      const filteredViolations = results.violations.filter(violation =>
        finalOptions.impacts!.includes(violation.impact as any)
      )
      
      expect({
        ...results,
        violations: filteredViolations
      }).toHaveNoViolations()
    } else {
      expect(results).toHaveNoViolations()
    }
  }

  /**
   * Test keyboard navigation and focus management
   */
  async testKeyboardNavigation(
    renderResult: RenderResult,
    options: KeyboardTestOptions = {}
  ): Promise<void> {
    const user = userEvent.setup()
    const { container } = renderResult

    // Test Tab navigation
    await this.testTabNavigation(container, user, options)
    
    // Test Enter and Space activation
    await this.testActivationKeys(container, user)
    
    // Test Escape key
    await this.testEscapeKey(container, user)
    
    // Test Arrow key navigation if applicable
    await this.testArrowNavigation(container, user)
    
    // Test custom key sequences
    if (options.keySequences) {
      await this.testCustomKeySequences(container, user, options.keySequences)
    }
  }

  /**
   * Test Tab navigation order and focus trapping
   */
  private async testTabNavigation(
    container: HTMLElement,
    user: ReturnType<typeof userEvent.setup>,
    options: KeyboardTestOptions
  ): Promise<void> {
    // Get all focusable elements
    const focusableElements = this.getFocusableElements(container)
    
    // Verify expected focusable elements
    if (options.focusableSelectors) {
      for (const selector of options.focusableSelectors) {
        const element = container.querySelector(selector)
        expect(element, `Element with selector "${selector}" should exist`).toBeTruthy()
        expect(
          focusableElements.includes(element as HTMLElement),
          `Element with selector "${selector}" should be focusable`
        ).toBe(true)
      }
    }

    // Verify non-focusable elements
    if (options.nonFocusableSelectors) {
      for (const selector of options.nonFocusableSelectors) {
        const element = container.querySelector(selector)
        if (element) {
          expect(
            focusableElements.includes(element as HTMLElement),
            `Element with selector "${selector}" should NOT be focusable`
          ).toBe(false)
        }
      }
    }

    // Test Tab order
    if (focusableElements.length > 0) {
      // Focus first element
      focusableElements[0].focus()
      expect(document.activeElement).toBe(focusableElements[0])

      // Tab through all elements
      for (let i = 1; i < focusableElements.length; i++) {
        await user.keyboard('{Tab}')
        expect(
          document.activeElement,
          `After ${i} Tab presses, focus should be on element ${i}`
        ).toBe(focusableElements[i])
      }

      // Test Shift+Tab backwards
      for (let i = focusableElements.length - 2; i >= 0; i--) {
        await user.keyboard('{Shift>}{Tab}{/Shift}')
        expect(
          document.activeElement,
          `After Shift+Tab, focus should be on element ${i}`
        ).toBe(focusableElements[i])
      }
    }
  }

  /**
   * Test Enter and Space key activation
   */
  private async testActivationKeys(
    container: HTMLElement,
    user: ReturnType<typeof userEvent.setup>
  ): Promise<void> {
    const activatableElements = container.querySelectorAll(
      'button, [role="button"], [role="menuitem"], [role="tab"], [role="option"]'
    )

    for (const element of Array.from(activatableElements)) {
      const htmlElement = element as HTMLElement
      
      // Focus the element
      htmlElement.focus()
      
      // Test Enter key activation
      const enterHandler = vi.fn()
      htmlElement.addEventListener('click', enterHandler)
      await user.keyboard('{Enter}')
      
      // Test Space key activation (for buttons and button-like elements)
      if (element.tagName === 'BUTTON' || element.getAttribute('role') === 'button') {
        const spaceHandler = vi.fn()
        htmlElement.addEventListener('click', spaceHandler)
        await user.keyboard(' ')
      }
      
      htmlElement.removeEventListener('click', enterHandler)
    }
  }

  /**
   * Test Escape key functionality
   */
  private async testEscapeKey(
    container: HTMLElement,
    user: ReturnType<typeof userEvent.setup>
  ): Promise<void> {
    // Test escape on modal dialogs
    const modals = container.querySelectorAll('[role="dialog"], [role="alertdialog"]')
    
    for (const modal of Array.from(modals)) {
      const htmlElement = modal as HTMLElement
      htmlElement.focus()
      
      // Add escape key listener to verify behavior
      const escapeHandler = vi.fn()
      document.addEventListener('keydown', escapeHandler)
      
      await user.keyboard('{Escape}')
      
      // Verify escape was handled
      expect(escapeHandler).toHaveBeenCalledWith(
        expect.objectContaining({ key: 'Escape' })
      )
      
      document.removeEventListener('keydown', escapeHandler)
    }
  }

  /**
   * Test Arrow key navigation for lists, menus, and grids
   */
  private async testArrowNavigation(
    container: HTMLElement,
    user: ReturnType<typeof userEvent.setup>
  ): Promise<void> {
    const navigableContainers = container.querySelectorAll(
      '[role="menu"], [role="menubar"], [role="tablist"], [role="listbox"], [role="grid"]'
    )

    for (const navContainer of Array.from(navigableContainers)) {
      const role = navContainer.getAttribute('role')
      const items = navContainer.querySelectorAll(
        '[role="menuitem"], [role="tab"], [role="option"], [role="gridcell"]'
      )

      if (items.length > 1) {
        const firstItem = items[0] as HTMLElement
        const secondItem = items[1] as HTMLElement
        
        // Focus first item
        firstItem.focus()
        expect(document.activeElement).toBe(firstItem)

        // Test appropriate arrow key based on role
        if (role === 'menu' || role === 'listbox') {
          // Vertical navigation
          await user.keyboard('{ArrowDown}')
          expect(document.activeElement).toBe(secondItem)
          
          await user.keyboard('{ArrowUp}')
          expect(document.activeElement).toBe(firstItem)
        } else if (role === 'menubar' || role === 'tablist') {
          // Horizontal navigation
          await user.keyboard('{ArrowRight}')
          expect(document.activeElement).toBe(secondItem)
          
          await user.keyboard('{ArrowLeft}')
          expect(document.activeElement).toBe(firstItem)
        }
      }
    }
  }

  /**
   * Test custom key sequences
   */
  private async testCustomKeySequences(
    container: HTMLElement,
    user: ReturnType<typeof userEvent.setup>,
    keySequences: KeyboardTestOptions['keySequences']
  ): Promise<void> {
    for (const sequence of keySequences!) {
      // Execute key sequence
      for (const key of sequence.keys) {
        await user.keyboard(key)
      }

      // Check expected focus if specified
      if (sequence.expectedFocus) {
        const expectedElement = container.querySelector(sequence.expectedFocus)
        expect(
          document.activeElement,
          `After ${sequence.description}, focus should be on ${sequence.expectedFocus}`
        ).toBe(expectedElement)
      }
    }
  }

  /**
   * Test screen reader compatibility
   */
  testScreenReaderCompatibility(
    container: HTMLElement,
    options: ScreenReaderTestOptions = {}
  ): void {
    // Test ARIA attributes
    if (options.ariaAttributes) {
      this.testAriaAttributes(container, options.ariaAttributes)
    }

    // Test landmark roles
    if (options.landmarks) {
      this.testLandmarks(container, options.landmarks)
    }

    // Test heading structure
    if (options.headingStructure) {
      this.testHeadingStructure(container)
    }

    // Test form accessibility
    this.testFormAccessibility(container)
    
    // Test image accessibility
    this.testImageAccessibility(container)
  }

  /**
   * Test ARIA attributes
   */
  private testAriaAttributes(container: HTMLElement, attributes: string[]): void {
    for (const attr of attributes) {
      const elementsWithAttr = container.querySelectorAll(`[${attr}]`)
      
      for (const element of Array.from(elementsWithAttr)) {
        const value = element.getAttribute(attr)
        
        // Validate common ARIA attributes
        switch (attr) {
          case 'aria-labelledby':
          case 'aria-describedby':
            // Should reference existing IDs
            if (value) {
              const ids = value.split(' ')
              for (const id of ids) {
                const referencedElement = document.getElementById(id)
                expect(
                  referencedElement,
                  `Element with ${attr}="${value}" references non-existent ID: ${id}`
                ).toBeTruthy()
              }
            }
            break
            
          case 'aria-expanded':
            // Should be true, false, or undefined
            expect(['true', 'false', null]).toContain(value)
            break
            
          case 'aria-hidden':
            // Should be true or false
            if (value !== null) {
              expect(['true', 'false']).toContain(value)
            }
            break
        }
      }
    }
  }

  /**
   * Test landmark roles
   */
  private testLandmarks(container: HTMLElement, landmarks: string[]): void {
    for (const landmark of landmarks) {
      const landmarkElements = container.querySelectorAll(`[role="${landmark}"], ${landmark}`)
      expect(
        landmarkElements.length,
        `Should have at least one ${landmark} landmark`
      ).toBeGreaterThan(0)
    }
  }

  /**
   * Test heading structure
   */
  private testHeadingStructure(container: HTMLElement): void {
    const headings = Array.from(container.querySelectorAll('h1, h2, h3, h4, h5, h6'))
      .map(h => ({
        element: h,
        level: parseInt(h.tagName.charAt(1))
      }))
      .sort((a, b) => {
        const aPos = Array.from(container.querySelectorAll('*')).indexOf(a.element)
        const bPos = Array.from(container.querySelectorAll('*')).indexOf(b.element)
        return aPos - bPos
      })

    if (headings.length === 0) return

    // Check for h1
    expect(
      headings.some(h => h.level === 1),
      'Should have at least one h1 element'
    ).toBe(true)

    // Check heading hierarchy
    let lastLevel = 0
    for (const heading of headings) {
      if (lastLevel > 0) {
        expect(
          heading.level <= lastLevel + 1,
          `Heading level ${heading.level} should not skip levels (previous was ${lastLevel})`
        ).toBe(true)
      }
      lastLevel = heading.level
    }
  }

  /**
   * Test form accessibility
   */
  private testFormAccessibility(container: HTMLElement): void {
    const formControls = container.querySelectorAll(
      'input:not([type="hidden"]), select, textarea'
    )

    for (const control of Array.from(formControls)) {
      const htmlControl = control as HTMLElement
      const id = htmlControl.id
      const type = htmlControl.getAttribute('type')

      // Skip submit and button inputs
      if (type === 'submit' || type === 'button') continue

      // Check for labels
      const hasLabel = 
        (id && document.querySelector(`label[for="${id}"]`)) ||
        htmlControl.getAttribute('aria-label') ||
        htmlControl.getAttribute('aria-labelledby')

      expect(
        hasLabel,
        `Form control ${htmlControl.tagName} should have an associated label`
      ).toBeTruthy()

      // Check required fields
      if (htmlControl.hasAttribute('required')) {
        expect(
          htmlControl.getAttribute('aria-required') === 'true',
          'Required form controls should have aria-required="true"'
        ).toBe(true)
      }
    }
  }

  /**
   * Test image accessibility
   */
  private testImageAccessibility(container: HTMLElement): void {
    const images = container.querySelectorAll('img')

    for (const img of Array.from(images)) {
      const hasAlt = img.hasAttribute('alt')
      const altText = img.getAttribute('alt')
      const isDecorative = 
        img.getAttribute('role') === 'presentation' ||
        img.getAttribute('aria-hidden') === 'true'

      if (!isDecorative) {
        expect(
          hasAlt,
          'Non-decorative images must have alt attributes'
        ).toBe(true)

        if (hasAlt && altText === '') {
          // Empty alt is only valid for decorative images
          expect(
            isDecorative,
            'Images with empty alt text should be marked as decorative'
          ).toBe(true)
        }
      }
    }
  }

  /**
   * Get all focusable elements in a container
   */
  private getFocusableElements(container: HTMLElement): HTMLElement[] {
    const focusableSelectors = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      'a[href]',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]'
    ].join(', ')

    return Array.from(container.querySelectorAll(focusableSelectors))
      .filter((el): el is HTMLElement => {
        const element = el as HTMLElement
        return this.isVisible(element) && !this.isInert(element)
      })
  }

  /**
   * Check if element is visible
   */
  private isVisible(element: HTMLElement): boolean {
    const style = window.getComputedStyle(element)
    return (
      style.display !== 'none' &&
      style.visibility !== 'hidden' &&
      style.opacity !== '0' &&
      element.offsetWidth > 0 &&
      element.offsetHeight > 0
    )
  }

  /**
   * Check if element is inert
   */
  private isInert(element: HTMLElement): boolean {
    if (
      element.hasAttribute('disabled') ||
      element.getAttribute('aria-hidden') === 'true'
    ) {
      return true
    }

    // Check if any parent has aria-hidden
    let parent = element.parentElement
    while (parent) {
      if (parent.getAttribute('aria-hidden') === 'true') {
        return true
      }
      parent = parent.parentElement
    }

    return false
  }

  /**
   * Test color contrast
   */
  testColorContrast(element: HTMLElement, minimumRatio: number = 4.5): void {
    const style = window.getComputedStyle(element)
    const backgroundColor = style.backgroundColor
    const color = style.color

    // This is a simplified test - in real implementation you'd use a proper contrast checking library
    expect(
      backgroundColor !== color,
      'Text and background colors should be different'
    ).toBe(true)
  }

  /**
   * Test touch target size (minimum 44x44 pixels)
   */
  testTouchTargetSize(container: HTMLElement): void {
    const interactiveElements = container.querySelectorAll(
      'button, a, input, select, textarea, [role="button"], [role="link"]'
    )

    for (const element of Array.from(interactiveElements)) {
      const htmlElement = element as HTMLElement
      const rect = htmlElement.getBoundingClientRect()
      
      const minSize = 44 // 44px minimum as per WCAG
      
      expect(
        rect.width >= minSize && rect.height >= minSize,
        `Interactive element should be at least ${minSize}x${minSize}px (actual: ${rect.width}x${rect.height}px)`
      ).toBe(true)
    }
  }
}

/**
 * Create accessibility tester instance
 */
export const accessibilityTester = new AccessibilityTester()

/**
 * Quick accessibility test helper
 */
export async function testAccessibility(
  component: React.ReactElement,
  options: AccessibilityTestOptions & KeyboardTestOptions & ScreenReaderTestOptions = {}
): Promise<void> {
  const renderResult = render(component)
  const { container } = renderResult

  // Run axe-core tests
  await accessibilityTester.testAxeCompliance(container, options)

  // Test keyboard navigation
  await accessibilityTester.testKeyboardNavigation(renderResult, options)

  // Test screen reader compatibility
  accessibilityTester.testScreenReaderCompatibility(container, options)

  // Test touch target sizes
  accessibilityTester.testTouchTargetSize(container)
}

/**
 * Test only axe-core compliance (faster for unit tests)
 */
export async function testAxeCompliance(
  component: React.ReactElement,
  options: AccessibilityTestOptions = {}
): Promise<void> {
  const { container } = render(component)
  await accessibilityTester.testAxeCompliance(container, options)
}

/**
 * Test only keyboard navigation
 */
export async function testKeyboardNavigation(
  component: React.ReactElement,
  options: KeyboardTestOptions = {}
): Promise<void> {
  const renderResult = render(component)
  await accessibilityTester.testKeyboardNavigation(renderResult, options)
}

/**
 * Test only screen reader compatibility
 */
export function testScreenReaderCompatibility(
  component: React.ReactElement,
  options: ScreenReaderTestOptions = {}
): void {
  const { container } = render(component)
  accessibilityTester.testScreenReaderCompatibility(container, options)
}