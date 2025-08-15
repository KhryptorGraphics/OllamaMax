/**
 * Skip Link Component - OllamaMax Design System
 * 
 * Provides keyboard navigation shortcuts for accessibility compliance.
 * Essential for WCAG 2.1 AA compliance.
 */

import React from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';
import i18nService from '../../services/i18nService.js';

const SkipLink = ({ 
  href = '#main-content', 
  children, 
  className = '',
  ...props 
}) => {
  // Skip link styles - hidden by default, visible on focus
  const skipLinkStyles = {
    position: 'absolute',
    top: '-40px',
    left: '6px',
    zIndex: tokens.zIndex.modal + 1000,
    backgroundColor: tokens.colors.neutral[900],
    color: tokens.colors.neutral[0],
    padding: `${tokens.spacing[2]} ${tokens.spacing[4]}`,
    borderRadius: tokens.borderRadius.md,
    fontSize: tokens.typography.fontSize.sm[0],
    fontWeight: tokens.typography.fontWeight.medium,
    textDecoration: 'none',
    border: `2px solid ${tokens.colors.primary[500]}`,
    transition: `top ${tokens.animation.duration.fast} ${tokens.animation.easing.easeOut}`,
    
    // Focus styles
    ':focus': {
      top: '6px',
      outline: 'none'
    }
  };

  const handleClick = (event) => {
    const target = document.querySelector(href);
    if (target) {
      event.preventDefault();
      
      // Make target focusable if it isn't already
      if (!target.hasAttribute('tabindex')) {
        target.setAttribute('tabindex', '-1');
      }
      
      // Focus the target
      target.focus();
      
      // Scroll to target
      target.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'start' 
      });
      
      // Remove tabindex after focus (if we added it)
      setTimeout(() => {
        if (target.getAttribute('tabindex') === '-1') {
          target.removeAttribute('tabindex');
        }
      }, 100);
    }
  };

  return (
    <a
      href={href}
      onClick={handleClick}
      className={`skip-link ${className}`}
      style={skipLinkStyles}
      {...props}
    >
      {children || i18nService.t('a11y.skipToMain')}
    </a>
  );
};

SkipLink.propTypes = {
  href: PropTypes.string,
  children: PropTypes.node,
  className: PropTypes.string
};

// Skip Links Container - provides multiple skip options
export const SkipLinks = ({ links = [], className = '' }) => {
  const defaultLinks = [
    { href: '#main-content', text: i18nService.t('a11y.skipToMain') },
    { href: '#navigation', text: 'Skip to navigation' },
    { href: '#footer', text: 'Skip to footer' }
  ];

  const skipLinks = links.length > 0 ? links : defaultLinks;

  const containerStyles = {
    position: 'absolute',
    top: 0,
    left: 0,
    zIndex: tokens.zIndex.modal + 1000
  };

  return (
    <div className={`skip-links ${className}`} style={containerStyles}>
      {skipLinks.map((link, index) => (
        <SkipLink
          key={index}
          href={link.href}
          style={{ left: `${6 + index * 150}px` }}
        >
          {link.text}
        </SkipLink>
      ))}
    </div>
  );
};

SkipLinks.propTypes = {
  links: PropTypes.arrayOf(PropTypes.shape({
    href: PropTypes.string.isRequired,
    text: PropTypes.string.isRequired
  })),
  className: PropTypes.string
};

// CSS for skip links
const styles = `
  .skip-link:focus {
    top: 6px !important;
  }
  
  .skip-link:not(:focus) {
    position: absolute !important;
    width: 1px !important;
    height: 1px !important;
    padding: 0 !important;
    margin: -1px !important;
    overflow: hidden !important;
    clip: rect(0, 0, 0, 0) !important;
    white-space: nowrap !important;
    border: 0 !important;
  }
`;

// Inject styles
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.textContent = styles;
  document.head.appendChild(styleSheet);
}

export default SkipLink;
