/**
 * @fileoverview Skip Links Component
 * Provides keyboard navigation shortcuts for screen reader users
 */

import React from 'react'

interface SkipLinksProps {
  /** Additional CSS classes */
  className?: string
  /** Custom skip link targets */
  links?: Array<{
    href: string
    label: string
  }>
}

/**
 * Skip Links Component
 * 
 * Provides keyboard shortcuts for screen reader users to quickly navigate
 * to main content areas. These links are hidden visually but accessible
 * to screen readers and become visible when focused.
 */
export const SkipLinks: React.FC<SkipLinksProps> = ({
  className,
  links = [
    { href: '#main-content', label: 'Skip to main content' },
    { href: '#navigation', label: 'Skip to navigation' },
    { href: '#footer', label: 'Skip to footer' }
  ]
}) => {
  return (
    <nav 
      className={`skip-links ${className || ''}`} 
      aria-label="Skip navigation links"
    >
      <ul className="skip-links-list">
        {links.map(({ href, label }) => (
          <li key={href}>
            <a
              href={href}
              className="skip-link"
              onFocus={(e) => {
                // Ensure the skip link is visible when focused
                e.target.scrollIntoView()
              }}
            >
              {label}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}

export default SkipLinks