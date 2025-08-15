/**
 * Focus Trap Component - OllamaMax Design System
 * 
 * Traps focus within a container for accessibility compliance.
 * Essential for modals, dropdowns, and other overlay components.
 */

import React, { useEffect, useRef, useCallback } from 'react';
import PropTypes from 'prop-types';
import accessibilityService from '../../services/accessibilityService.js';

const FocusTrap = ({ 
  children, 
  active = true, 
  restoreFocus = true,
  autoFocus = true,
  className = '',
  ...props 
}) => {
  const containerRef = useRef(null);
  const previousActiveElement = useRef(null);
  const firstFocusableElement = useRef(null);
  const lastFocusableElement = useRef(null);

  // Get all focusable elements within the container
  const getFocusableElements = useCallback(() => {
    if (!containerRef.current) return [];
    
    const focusableSelectors = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      'a[href]',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]',
      'audio[controls]',
      'video[controls]',
      'iframe',
      'object',
      'embed',
      'area[href]',
      'summary'
    ].join(', ');

    const elements = Array.from(containerRef.current.querySelectorAll(focusableSelectors))
      .filter(element => {
        // Check if element is visible and not disabled
        const style = window.getComputedStyle(element);
        return (
          style.display !== 'none' &&
          style.visibility !== 'hidden' &&
          style.opacity !== '0' &&
          element.offsetWidth > 0 &&
          element.offsetHeight > 0 &&
          !element.hasAttribute('disabled') &&
          element.getAttribute('tabindex') !== '-1'
        );
      });

    return elements;
  }, []);

  // Update focusable elements
  const updateFocusableElements = useCallback(() => {
    const elements = getFocusableElements();
    firstFocusableElement.current = elements[0] || null;
    lastFocusableElement.current = elements[elements.length - 1] || null;
  }, [getFocusableElements]);

  // Handle tab key navigation
  const handleKeyDown = useCallback((event) => {
    if (!active || event.key !== 'Tab') return;

    updateFocusableElements();

    const { shiftKey } = event;
    const activeElement = document.activeElement;

    // If no focusable elements, prevent tabbing
    if (!firstFocusableElement.current) {
      event.preventDefault();
      return;
    }

    // If only one focusable element, keep focus on it
    if (firstFocusableElement.current === lastFocusableElement.current) {
      event.preventDefault();
      firstFocusableElement.current.focus();
      return;
    }

    // Handle tab navigation
    if (shiftKey) {
      // Shift + Tab (backward)
      if (activeElement === firstFocusableElement.current) {
        event.preventDefault();
        lastFocusableElement.current?.focus();
      }
    } else {
      // Tab (forward)
      if (activeElement === lastFocusableElement.current) {
        event.preventDefault();
        firstFocusableElement.current?.focus();
      }
    }
  }, [active, updateFocusableElements]);

  // Handle escape key
  const handleEscape = useCallback((event) => {
    if (active && event.key === 'Escape') {
      // Let parent components handle escape
      event.stopPropagation();
    }
  }, [active]);

  // Setup focus trap
  useEffect(() => {
    if (!active) return;

    // Store the previously focused element
    previousActiveElement.current = document.activeElement;

    // Update focusable elements
    updateFocusableElements();

    // Auto-focus first element if requested
    if (autoFocus && firstFocusableElement.current) {
      // Use a small delay to ensure the element is ready
      setTimeout(() => {
        firstFocusableElement.current?.focus();
      }, 10);
    }

    // Add event listeners
    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('keydown', handleEscape);

    // Cleanup function
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('keydown', handleEscape);

      // Restore focus to previously focused element
      if (restoreFocus && previousActiveElement.current) {
        // Check if the element still exists and is focusable
        if (document.contains(previousActiveElement.current)) {
          try {
            previousActiveElement.current.focus();
          } catch (error) {
            console.warn('Failed to restore focus:', error);
          }
        }
      }
    };
  }, [active, autoFocus, restoreFocus, handleKeyDown, handleEscape, updateFocusableElements]);

  // Update focusable elements when children change
  useEffect(() => {
    if (active) {
      updateFocusableElements();
    }
  }, [children, active, updateFocusableElements]);

  // Handle clicks outside focusable elements
  const handleClick = useCallback((event) => {
    if (!active) return;

    const focusableElements = getFocusableElements();
    const clickedElement = event.target;

    // Check if clicked element is focusable
    const isClickedElementFocusable = focusableElements.some(element => 
      element === clickedElement || element.contains(clickedElement)
    );

    // If clicked outside focusable elements, focus the first one
    if (!isClickedElementFocusable && firstFocusableElement.current) {
      firstFocusableElement.current.focus();
    }
  }, [active, getFocusableElements]);

  // Container styles
  const containerStyles = {
    outline: 'none'
  };

  return (
    <div
      ref={containerRef}
      className={className}
      style={containerStyles}
      onClick={handleClick}
      tabIndex={-1}
      {...props}
    >
      {children}
    </div>
  );
};

FocusTrap.propTypes = {
  children: PropTypes.node.isRequired,
  active: PropTypes.bool,
  restoreFocus: PropTypes.bool,
  autoFocus: PropTypes.bool,
  className: PropTypes.string
};

// Hook for using focus trap functionality
export const useFocusTrap = (active = true, options = {}) => {
  const {
    restoreFocus = true,
    autoFocus = true,
    onEscape
  } = options;

  const containerRef = useRef(null);
  const previousActiveElement = useRef(null);

  useEffect(() => {
    if (!active || !containerRef.current) return;

    const cleanup = accessibilityService.trapFocus(containerRef.current);
    
    // Store previous focus
    previousActiveElement.current = document.activeElement;

    // Handle escape key
    const handleEscape = (event) => {
      if (event.key === 'Escape') {
        onEscape?.(event);
      }
    };

    document.addEventListener('keydown', handleEscape);

    return () => {
      cleanup?.();
      document.removeEventListener('keydown', handleEscape);
      
      // Restore focus
      if (restoreFocus && previousActiveElement.current) {
        try {
          previousActiveElement.current.focus();
        } catch (error) {
          console.warn('Failed to restore focus:', error);
        }
      }
    };
  }, [active, restoreFocus, onEscape]);

  return containerRef;
};

export default FocusTrap;
