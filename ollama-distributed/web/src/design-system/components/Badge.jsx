/**
 * Badge Component - OllamaMax Design System
 * 
 * A flexible badge component for displaying status, labels, and notifications.
 */

import React from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';

const Badge = ({
  children,
  variant = 'default',
  size = 'md',
  dot = false,
  className = '',
  ...props
}) => {
  // Base styles
  const baseStyles = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontFamily: tokens.typography.fontFamily.sans.join(', '),
    fontWeight: tokens.typography.fontWeight.medium,
    borderRadius: tokens.borderRadius.full,
    border: 'none',
    whiteSpace: 'nowrap',
    textDecoration: 'none',
    userSelect: 'none',
    transition: `all ${tokens.animation.duration.normal} ${tokens.animation.easing.easeInOut}`
  };

  // Size variants
  const sizeStyles = {
    sm: {
      height: dot ? '8px' : '20px',
      minWidth: dot ? '8px' : '20px',
      padding: dot ? '0' : '0 6px',
      fontSize: dot ? '0' : tokens.typography.fontSize.xs[0],
      lineHeight: tokens.typography.lineHeight.none
    },
    md: {
      height: dot ? '10px' : '24px',
      minWidth: dot ? '10px' : '24px',
      padding: dot ? '0' : '0 8px',
      fontSize: dot ? '0' : tokens.typography.fontSize.sm[0],
      lineHeight: tokens.typography.lineHeight.none
    },
    lg: {
      height: dot ? '12px' : '28px',
      minWidth: dot ? '12px' : '28px',
      padding: dot ? '0' : '0 10px',
      fontSize: dot ? '0' : tokens.typography.fontSize.sm[0],
      lineHeight: tokens.typography.lineHeight.none
    }
  };

  // Variant styles
  const variantStyles = {
    default: {
      backgroundColor: tokens.colors.neutral[100],
      color: tokens.colors.neutral[800]
    },
    primary: {
      backgroundColor: tokens.colors.primary[500],
      color: tokens.colors.neutral[0]
    },
    secondary: {
      backgroundColor: tokens.colors.secondary[500],
      color: tokens.colors.neutral[0]
    },
    success: {
      backgroundColor: tokens.colors.success[500],
      color: tokens.colors.neutral[0]
    },
    warning: {
      backgroundColor: tokens.colors.warning[500],
      color: tokens.colors.neutral[0]
    },
    error: {
      backgroundColor: tokens.colors.error[500],
      color: tokens.colors.neutral[0]
    },
    info: {
      backgroundColor: tokens.colors.info[500],
      color: tokens.colors.neutral[0]
    },
    outline: {
      backgroundColor: 'transparent',
      color: tokens.colors.neutral[700],
      border: `1px solid ${tokens.colors.neutral[300]}`
    },
    'outline-primary': {
      backgroundColor: 'transparent',
      color: tokens.colors.primary[600],
      border: `1px solid ${tokens.colors.primary[300]}`
    },
    'outline-success': {
      backgroundColor: 'transparent',
      color: tokens.colors.success[600],
      border: `1px solid ${tokens.colors.success[300]}`
    },
    'outline-warning': {
      backgroundColor: 'transparent',
      color: tokens.colors.warning[600],
      border: `1px solid ${tokens.colors.warning[300]}`
    },
    'outline-error': {
      backgroundColor: 'transparent',
      color: tokens.colors.error[600],
      border: `1px solid ${tokens.colors.error[300]}`
    }
  };

  // Combine styles
  const badgeStyles = {
    ...baseStyles,
    ...sizeStyles[size],
    ...variantStyles[variant]
  };

  return (
    <span
      className={className}
      style={badgeStyles}
      {...props}
    >
      {!dot && children}
    </span>
  );
};

Badge.propTypes = {
  children: PropTypes.node,
  variant: PropTypes.oneOf([
    'default',
    'primary',
    'secondary',
    'success',
    'warning',
    'error',
    'info',
    'outline',
    'outline-primary',
    'outline-success',
    'outline-warning',
    'outline-error'
  ]),
  size: PropTypes.oneOf(['sm', 'md', 'lg']),
  dot: PropTypes.bool,
  className: PropTypes.string
};

export default Badge;
