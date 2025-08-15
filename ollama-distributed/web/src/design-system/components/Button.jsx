/**
 * Button Component - OllamaMax Design System
 *
 * A flexible button component with multiple variants, sizes, and states.
 * Fully compliant with WCAG 2.1 AA accessibility standards.
 */

import React, { forwardRef, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';
import accessibilityService from '../../services/accessibilityService.js';
import i18nService from '../../services/i18nService.js';

const Button = forwardRef(({
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  fullWidth = false,
  leftIcon,
  rightIcon,
  onClick,
  type = 'button',
  className = '',
  'aria-label': ariaLabel,
  'aria-describedby': ariaDescribedBy,
  'aria-pressed': ariaPressed,
  'aria-expanded': ariaExpanded,
  'aria-haspopup': ariaHasPopup,
  role,
  tabIndex,
  autoFocus = false,
  ...props
}, ref) => {
  const [isFocused, setIsFocused] = useState(false);
  const [isPressed, setIsPressed] = useState(false);

  // Handle accessibility announcements
  useEffect(() => {
    if (loading && accessibilityService.screenReaderActive) {
      accessibilityService.announce(i18nService.t('a11y.loading'));
    }
  }, [loading]);
  // Base styles
  const baseStyles = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: tokens.spacing[2],
    fontFamily: tokens.typography.fontFamily.sans.join(', '),
    fontWeight: tokens.typography.fontWeight.medium,
    borderRadius: tokens.borderRadius.md,
    border: 'none',
    cursor: disabled || loading ? 'not-allowed' : 'pointer',
    transition: `all ${tokens.animation.duration.normal} ${tokens.animation.easing.easeInOut}`,
    textDecoration: 'none',
    userSelect: 'none',
    position: 'relative',
    overflow: 'hidden',
    width: fullWidth ? '100%' : 'auto',
    opacity: disabled ? 0.6 : 1,
    transform: 'translateY(0)',
    boxShadow: 'none'
  };

  // Size variants
  const sizeStyles = {
    sm: {
      height: tokens.components.button.height.sm,
      padding: tokens.components.button.padding.sm,
      fontSize: tokens.components.button.fontSize.sm,
      lineHeight: tokens.typography.lineHeight.tight
    },
    md: {
      height: tokens.components.button.height.md,
      padding: tokens.components.button.padding.md,
      fontSize: tokens.components.button.fontSize.md,
      lineHeight: tokens.typography.lineHeight.normal
    },
    lg: {
      height: tokens.components.button.height.lg,
      padding: tokens.components.button.padding.lg,
      fontSize: tokens.components.button.fontSize.lg,
      lineHeight: tokens.typography.lineHeight.normal
    }
  };

  // Color variants
  const variantStyles = {
    primary: {
      backgroundColor: tokens.colors.primary[500],
      color: tokens.colors.neutral[0],
      ':hover': {
        backgroundColor: tokens.colors.primary[600],
        transform: 'translateY(-1px)',
        boxShadow: tokens.shadows.md
      },
      ':active': {
        backgroundColor: tokens.colors.primary[700],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.primary[300]}`,
        outlineOffset: '2px'
      }
    },
    secondary: {
      backgroundColor: tokens.colors.neutral[100],
      color: tokens.colors.neutral[900],
      border: `1px solid ${tokens.colors.neutral[300]}`,
      ':hover': {
        backgroundColor: tokens.colors.neutral[200],
        borderColor: tokens.colors.neutral[400],
        transform: 'translateY(-1px)',
        boxShadow: tokens.shadows.sm
      },
      ':active': {
        backgroundColor: tokens.colors.neutral[300],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.primary[300]}`,
        outlineOffset: '2px'
      }
    },
    outline: {
      backgroundColor: 'transparent',
      color: tokens.colors.primary[600],
      border: `1px solid ${tokens.colors.primary[300]}`,
      ':hover': {
        backgroundColor: tokens.colors.primary[50],
        borderColor: tokens.colors.primary[400],
        transform: 'translateY(-1px)'
      },
      ':active': {
        backgroundColor: tokens.colors.primary[100],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.primary[300]}`,
        outlineOffset: '2px'
      }
    },
    ghost: {
      backgroundColor: 'transparent',
      color: tokens.colors.neutral[700],
      ':hover': {
        backgroundColor: tokens.colors.neutral[100],
        transform: 'translateY(-1px)'
      },
      ':active': {
        backgroundColor: tokens.colors.neutral[200],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.primary[300]}`,
        outlineOffset: '2px'
      }
    },
    danger: {
      backgroundColor: tokens.colors.error[500],
      color: tokens.colors.neutral[0],
      ':hover': {
        backgroundColor: tokens.colors.error[600],
        transform: 'translateY(-1px)',
        boxShadow: tokens.shadows.md
      },
      ':active': {
        backgroundColor: tokens.colors.error[700],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.error[300]}`,
        outlineOffset: '2px'
      }
    },
    success: {
      backgroundColor: tokens.colors.success[500],
      color: tokens.colors.neutral[0],
      ':hover': {
        backgroundColor: tokens.colors.success[600],
        transform: 'translateY(-1px)',
        boxShadow: tokens.shadows.md
      },
      ':active': {
        backgroundColor: tokens.colors.success[700],
        transform: 'translateY(0)'
      },
      ':focus': {
        outline: `2px solid ${tokens.colors.success[300]}`,
        outlineOffset: '2px'
      }
    }
  };

  // Combine styles
  const buttonStyles = {
    ...baseStyles,
    ...sizeStyles[size],
    ...variantStyles[variant]
  };

  // Loading spinner component
  const LoadingSpinner = () => (
    <div
      style={{
        width: '1em',
        height: '1em',
        border: '2px solid currentColor',
        borderTopColor: 'transparent',
        borderRadius: '50%',
        animation: `spin ${tokens.animation.duration.slow} linear infinite`
      }}
      aria-hidden="true"
    />
  );

  // Enhanced event handlers
  const handleClick = (event) => {
    if (disabled || loading) {
      event.preventDefault();
      return;
    }

    // Announce action for screen readers
    if (accessibilityService.screenReaderActive && ariaLabel) {
      accessibilityService.announce(`${ariaLabel} activated`);
    }

    onClick?.(event);
  };

  const handleKeyDown = (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleClick(event);
    }
  };

  const handleFocus = (event) => {
    setIsFocused(true);
    props.onFocus?.(event);
  };

  const handleBlur = (event) => {
    setIsFocused(false);
    setIsPressed(false);
    props.onBlur?.(event);
  };

  const handleMouseDown = (event) => {
    setIsPressed(true);
    props.onMouseDown?.(event);
  };

  const handleMouseUp = (event) => {
    setIsPressed(false);
    props.onMouseUp?.(event);
  };

  // Enhanced button styles with focus and pressed states
  const enhancedButtonStyles = {
    ...buttonStyles,
    outline: isFocused ? `2px solid ${tokens.colors.primary[500]}` : 'none',
    outlineOffset: '2px',
    transform: isPressed ? 'translateY(1px)' : 'translateY(0)',
    boxShadow: isFocused ? `0 0 0 3px ${tokens.colors.primary[500]}20` : 'none'
  };

  return (
    <button
      ref={ref}
      type={type}
      role={role}
      disabled={disabled || loading}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      onFocus={handleFocus}
      onBlur={handleBlur}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      className={className}
      style={enhancedButtonStyles}
      aria-label={ariaLabel || (typeof children === 'string' ? children : undefined)}
      aria-describedby={ariaDescribedBy}
      aria-pressed={ariaPressed}
      aria-expanded={ariaExpanded}
      aria-haspopup={ariaHasPopup}
      aria-disabled={disabled || loading}
      aria-busy={loading}
      tabIndex={disabled ? -1 : (tabIndex ?? 0)}
      autoFocus={autoFocus}
      {...props}
    >
      {loading && <LoadingSpinner />}
      {!loading && leftIcon && (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {leftIcon}
        </span>
      )}
      {!loading && children && (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {children}
        </span>
      )}
      {!loading && rightIcon && (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {rightIcon}
        </span>
      )}
    </button>
  );
});

Button.displayName = 'Button';

Button.propTypes = {
  children: PropTypes.node,
  variant: PropTypes.oneOf(['primary', 'secondary', 'outline', 'ghost', 'danger', 'success']),
  size: PropTypes.oneOf(['sm', 'md', 'lg']),
  disabled: PropTypes.bool,
  loading: PropTypes.bool,
  fullWidth: PropTypes.bool,
  leftIcon: PropTypes.node,
  rightIcon: PropTypes.node,
  onClick: PropTypes.func,
  onFocus: PropTypes.func,
  onBlur: PropTypes.func,
  onMouseDown: PropTypes.func,
  onMouseUp: PropTypes.func,
  type: PropTypes.oneOf(['button', 'submit', 'reset']),
  className: PropTypes.string,
  role: PropTypes.string,
  tabIndex: PropTypes.number,
  autoFocus: PropTypes.bool,
  'aria-label': PropTypes.string,
  'aria-describedby': PropTypes.string,
  'aria-pressed': PropTypes.oneOfType([PropTypes.bool, PropTypes.string]),
  'aria-expanded': PropTypes.oneOfType([PropTypes.bool, PropTypes.string]),
  'aria-haspopup': PropTypes.oneOfType([PropTypes.bool, PropTypes.string])
};

// CSS-in-JS styles for animations
const styles = `
  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
`;

// Inject styles
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.textContent = styles;
  document.head.appendChild(styleSheet);
}

export default Button;
