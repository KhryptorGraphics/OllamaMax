/**
 * Accessibility Service
 * 
 * Provides accessibility features including focus management, screen reader support,
 * keyboard navigation, and WCAG 2.1 AA compliance utilities.
 */

class AccessibilityService {
  constructor() {
    this.focusHistory = [];
    this.announcements = [];
    this.reducedMotion = false;
    this.highContrast = false;
    this.screenReaderActive = false;
    
    this.init();
  }

  // Initialize accessibility service
  init() {
    this.detectReducedMotion();
    this.detectHighContrast();
    this.detectScreenReader();
    this.setupKeyboardNavigation();
    this.setupFocusManagement();
    this.createLiveRegion();
  }

  // Detect reduced motion preference
  detectReducedMotion() {
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
      this.reducedMotion = mediaQuery.matches;
      
      mediaQuery.addEventListener('change', (e) => {
        this.reducedMotion = e.matches;
        this.onReducedMotionChange(e.matches);
      });
    }
  }

  // Detect high contrast preference
  detectHighContrast() {
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-contrast: high)');
      this.highContrast = mediaQuery.matches;
      
      mediaQuery.addEventListener('change', (e) => {
        this.highContrast = e.matches;
        this.onHighContrastChange(e.matches);
      });
    }
  }

  // Detect screen reader
  detectScreenReader() {
    // Check for common screen reader indicators
    const indicators = [
      'JAWS',
      'NVDA',
      'DRAGON',
      'VoiceOver',
      'TalkBack'
    ];
    
    const userAgent = navigator.userAgent;
    this.screenReaderActive = indicators.some(indicator => 
      userAgent.includes(indicator)
    );
    
    // Also check for accessibility APIs
    if ('speechSynthesis' in window) {
      this.screenReaderActive = true;
    }
  }

  // Setup keyboard navigation
  setupKeyboardNavigation() {
    document.addEventListener('keydown', (event) => {
      this.handleKeyboardNavigation(event);
    });
    
    // Add visible focus indicators
    document.addEventListener('keydown', (event) => {
      if (event.key === 'Tab') {
        document.body.classList.add('keyboard-navigation');
      }
    });
    
    document.addEventListener('mousedown', () => {
      document.body.classList.remove('keyboard-navigation');
    });
  }

  // Handle keyboard navigation
  handleKeyboardNavigation(event) {
    const { key, ctrlKey, altKey, shiftKey } = event;
    
    // Skip navigation
    if (key === 'Tab' && ctrlKey) {
      event.preventDefault();
      this.skipToMainContent();
      return;
    }
    
    // Escape key handling
    if (key === 'Escape') {
      this.handleEscape();
      return;
    }
    
    // Arrow key navigation for custom components
    if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(key)) {
      this.handleArrowNavigation(event);
    }
    
    // Enter and Space for activation
    if (key === 'Enter' || key === ' ') {
      this.handleActivation(event);
    }
  }

  // Setup focus management
  setupFocusManagement() {
    // Track focus changes
    document.addEventListener('focusin', (event) => {
      this.onFocusChange(event.target);
    });
    
    // Handle focus trapping
    document.addEventListener('keydown', (event) => {
      if (event.key === 'Tab') {
        this.handleFocusTrapping(event);
      }
    });
  }

  // Create live region for announcements
  createLiveRegion() {
    const liveRegion = document.createElement('div');
    liveRegion.id = 'accessibility-live-region';
    liveRegion.setAttribute('aria-live', 'polite');
    liveRegion.setAttribute('aria-atomic', 'true');
    liveRegion.style.cssText = `
      position: absolute;
      left: -10000px;
      width: 1px;
      height: 1px;
      overflow: hidden;
    `;
    
    document.body.appendChild(liveRegion);
    this.liveRegion = liveRegion;
  }

  // Announce to screen readers
  announce(message, priority = 'polite') {
    if (!this.liveRegion) return;
    
    this.liveRegion.setAttribute('aria-live', priority);
    this.liveRegion.textContent = message;
    
    // Clear after announcement
    setTimeout(() => {
      this.liveRegion.textContent = '';
    }, 1000);
    
    this.announcements.push({
      message,
      priority,
      timestamp: new Date()
    });
  }

  // Focus management
  focusElement(element, options = {}) {
    if (!element) return false;
    
    const { preventScroll = false, announce = false } = options;
    
    try {
      element.focus({ preventScroll });
      
      if (announce) {
        const label = this.getElementLabel(element);
        if (label) {
          this.announce(`Focused on ${label}`);
        }
      }
      
      this.focusHistory.push(element);
      return true;
    } catch (error) {
      console.warn('Failed to focus element:', error);
      return false;
    }
  }

  // Get element label for announcements
  getElementLabel(element) {
    return (
      element.getAttribute('aria-label') ||
      element.getAttribute('aria-labelledby') ||
      element.getAttribute('title') ||
      element.textContent?.trim() ||
      element.getAttribute('placeholder') ||
      element.tagName.toLowerCase()
    );
  }

  // Focus trapping for modals
  trapFocus(container) {
    if (!container) return;
    
    const focusableElements = this.getFocusableElements(container);
    if (focusableElements.length === 0) return;
    
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];
    
    // Focus first element
    this.focusElement(firstElement);
    
    const trapHandler = (event) => {
      if (event.key === 'Tab') {
        if (event.shiftKey) {
          if (document.activeElement === firstElement) {
            event.preventDefault();
            this.focusElement(lastElement);
          }
        } else {
          if (document.activeElement === lastElement) {
            event.preventDefault();
            this.focusElement(firstElement);
          }
        }
      }
    };
    
    container.addEventListener('keydown', trapHandler);
    
    return () => {
      container.removeEventListener('keydown', trapHandler);
    };
  }

  // Get focusable elements
  getFocusableElements(container) {
    const selector = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      'a[href]',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]'
    ].join(', ');
    
    return Array.from(container.querySelectorAll(selector))
      .filter(element => this.isVisible(element));
  }

  // Check if element is visible
  isVisible(element) {
    const style = window.getComputedStyle(element);
    return (
      style.display !== 'none' &&
      style.visibility !== 'hidden' &&
      style.opacity !== '0' &&
      element.offsetWidth > 0 &&
      element.offsetHeight > 0
    );
  }

  // Skip to main content
  skipToMainContent() {
    const mainContent = document.querySelector('main, [role="main"], #main-content');
    if (mainContent) {
      this.focusElement(mainContent, { announce: true });
    }
  }

  // Handle escape key
  handleEscape() {
    // Close modals, dropdowns, etc.
    const openModal = document.querySelector('[role="dialog"][aria-hidden="false"]');
    if (openModal) {
      const closeButton = openModal.querySelector('[aria-label*="close"], [aria-label*="Close"]');
      if (closeButton) {
        closeButton.click();
      }
    }
    
    // Return focus to previous element
    if (this.focusHistory.length > 1) {
      const previousElement = this.focusHistory[this.focusHistory.length - 2];
      if (previousElement && this.isVisible(previousElement)) {
        this.focusElement(previousElement);
      }
    }
  }

  // Handle arrow navigation
  handleArrowNavigation(event) {
    const target = event.target;
    const role = target.getAttribute('role');
    
    // Handle specific ARIA patterns
    switch (role) {
      case 'tablist':
        this.handleTablistNavigation(event);
        break;
      case 'menu':
      case 'menubar':
        this.handleMenuNavigation(event);
        break;
      case 'listbox':
        this.handleListboxNavigation(event);
        break;
      case 'grid':
        this.handleGridNavigation(event);
        break;
    }
  }

  // Handle activation (Enter/Space)
  handleActivation(event) {
    const target = event.target;
    const role = target.getAttribute('role');
    
    // Only handle if not a native interactive element
    if (!['BUTTON', 'A', 'INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName)) {
      if (role === 'button' || target.hasAttribute('aria-pressed')) {
        event.preventDefault();
        target.click();
      }
    }
  }

  // Color contrast utilities
  checkColorContrast(foreground, background) {
    const getLuminance = (color) => {
      const rgb = this.hexToRgb(color);
      const [r, g, b] = rgb.map(c => {
        c = c / 255;
        return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
      });
      return 0.2126 * r + 0.7152 * g + 0.0722 * b;
    };
    
    const l1 = getLuminance(foreground);
    const l2 = getLuminance(background);
    const ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
    
    return {
      ratio,
      AA: ratio >= 4.5,
      AAA: ratio >= 7,
      AALarge: ratio >= 3,
      AAALarge: ratio >= 4.5
    };
  }

  // Convert hex to RGB
  hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? [
      parseInt(result[1], 16),
      parseInt(result[2], 16),
      parseInt(result[3], 16)
    ] : null;
  }

  // Generate accessible color
  generateAccessibleColor(baseColor, backgroundColor, targetRatio = 4.5) {
    // Implementation for generating accessible colors
    // This is a simplified version - full implementation would be more complex
    return baseColor;
  }

  // Event handlers (to be overridden)
  onReducedMotionChange(enabled) {
    console.log('Reduced motion:', enabled);
  }

  onHighContrastChange(enabled) {
    console.log('High contrast:', enabled);
  }

  onFocusChange(element) {
    // Override in app
  }

  // Get accessibility info
  getAccessibilityInfo() {
    return {
      reducedMotion: this.reducedMotion,
      highContrast: this.highContrast,
      screenReaderActive: this.screenReaderActive,
      focusHistoryLength: this.focusHistory.length,
      announcementsCount: this.announcements.length
    };
  }

  // Cleanup
  destroy() {
    if (this.liveRegion) {
      this.liveRegion.remove();
    }
    
    // Remove event listeners
    document.removeEventListener('keydown', this.handleKeyboardNavigation);
    document.removeEventListener('focusin', this.onFocusChange);
  }
}

// Create singleton instance
const accessibilityService = new AccessibilityService();

export default accessibilityService;
