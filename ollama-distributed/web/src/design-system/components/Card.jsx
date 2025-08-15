/**
 * Card Component - OllamaMax Design System
 * 
 * A flexible card component for displaying content with consistent styling.
 */

import React from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';

const Card = ({
  children,
  variant = 'default',
  size = 'md',
  interactive = false,
  hover = false,
  shadow = 'base',
  padding,
  className = '',
  onClick,
  ...props
}) => {
  // Base styles
  const baseStyles = {
    backgroundColor: tokens.colors.neutral[0],
    borderRadius: tokens.borderRadius.lg,
    border: '1px solid',
    borderColor: tokens.colors.neutral[200],
    transition: `all ${tokens.animation.duration.normal} ${tokens.animation.easing.easeInOut}`,
    position: 'relative',
    overflow: 'hidden'
  };

  // Size variants
  const sizeStyles = {
    sm: {
      padding: padding || tokens.components.card.padding.sm
    },
    md: {
      padding: padding || tokens.components.card.padding.md
    },
    lg: {
      padding: padding || tokens.components.card.padding.lg
    }
  };

  // Variant styles
  const variantStyles = {
    default: {
      backgroundColor: tokens.colors.neutral[0],
      borderColor: tokens.colors.neutral[200]
    },
    elevated: {
      backgroundColor: tokens.colors.neutral[0],
      borderColor: 'transparent',
      boxShadow: tokens.shadows[shadow]
    },
    outlined: {
      backgroundColor: 'transparent',
      borderColor: tokens.colors.neutral[300],
      borderWidth: '2px'
    },
    filled: {
      backgroundColor: tokens.colors.neutral[50],
      borderColor: tokens.colors.neutral[200]
    },
    gradient: {
      background: `linear-gradient(135deg, ${tokens.colors.primary[500]} 0%, ${tokens.colors.secondary[500]} 100%)`,
      borderColor: 'transparent',
      color: tokens.colors.neutral[0]
    }
  };

  // Interactive styles
  const interactiveStyles = interactive || onClick ? {
    cursor: 'pointer',
    ':hover': {
      transform: 'translateY(-2px)',
      boxShadow: tokens.shadows.lg,
      borderColor: tokens.colors.primary[300]
    },
    ':active': {
      transform: 'translateY(0)',
      boxShadow: tokens.shadows.md
    }
  } : {};

  // Hover styles
  const hoverStyles = hover ? {
    ':hover': {
      boxShadow: tokens.shadows.lg,
      transform: 'translateY(-1px)'
    }
  } : {};

  // Combine styles
  const cardStyles = {
    ...baseStyles,
    ...sizeStyles[size],
    ...variantStyles[variant],
    ...interactiveStyles,
    ...hoverStyles
  };

  // Handle click
  const handleClick = (event) => {
    if (onClick) {
      onClick(event);
    }
  };

  // Handle keyboard events for accessibility
  const handleKeyDown = (event) => {
    if ((interactive || onClick) && (event.key === 'Enter' || event.key === ' ')) {
      event.preventDefault();
      handleClick(event);
    }
  };

  return (
    <div
      className={className}
      style={cardStyles}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role={interactive || onClick ? 'button' : undefined}
      tabIndex={interactive || onClick ? 0 : undefined}
      {...props}
    >
      {children}
    </div>
  );
};

Card.propTypes = {
  children: PropTypes.node.isRequired,
  variant: PropTypes.oneOf(['default', 'elevated', 'outlined', 'filled', 'gradient']),
  size: PropTypes.oneOf(['sm', 'md', 'lg']),
  interactive: PropTypes.bool,
  hover: PropTypes.bool,
  shadow: PropTypes.oneOf(['sm', 'base', 'md', 'lg', 'xl', '2xl', 'none']),
  padding: PropTypes.string,
  className: PropTypes.string,
  onClick: PropTypes.func
};

// Card Header Component
const CardHeader = ({ children, className = '', ...props }) => (
  <div
    className={className}
    style={{
      padding: `${tokens.spacing[4]} ${tokens.spacing[6]} ${tokens.spacing[2]}`,
      borderBottom: `1px solid ${tokens.colors.neutral[200]}`,
      marginBottom: tokens.spacing[4]
    }}
    {...props}
  >
    {children}
  </div>
);

CardHeader.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string
};

// Card Body Component
const CardBody = ({ children, className = '', ...props }) => (
  <div
    className={className}
    style={{
      flex: 1
    }}
    {...props}
  >
    {children}
  </div>
);

CardBody.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string
};

// Card Footer Component
const CardFooter = ({ children, className = '', ...props }) => (
  <div
    className={className}
    style={{
      padding: `${tokens.spacing[2]} ${tokens.spacing[6]} ${tokens.spacing[4]}`,
      borderTop: `1px solid ${tokens.colors.neutral[200]}`,
      marginTop: tokens.spacing[4]
    }}
    {...props}
  >
    {children}
  </div>
);

CardFooter.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string
};

// Card Title Component
const CardTitle = ({ children, size = 'md', className = '', ...props }) => {
  const titleStyles = {
    sm: {
      fontSize: tokens.typography.fontSize.lg[0],
      fontWeight: tokens.typography.fontWeight.semibold,
      lineHeight: tokens.typography.lineHeight.tight
    },
    md: {
      fontSize: tokens.typography.fontSize.xl[0],
      fontWeight: tokens.typography.fontWeight.semibold,
      lineHeight: tokens.typography.lineHeight.tight
    },
    lg: {
      fontSize: tokens.typography.fontSize['2xl'][0],
      fontWeight: tokens.typography.fontWeight.bold,
      lineHeight: tokens.typography.lineHeight.tight
    }
  };

  return (
    <h3
      className={className}
      style={{
        margin: 0,
        color: tokens.colors.neutral[900],
        ...titleStyles[size]
      }}
      {...props}
    >
      {children}
    </h3>
  );
};

CardTitle.propTypes = {
  children: PropTypes.node.isRequired,
  size: PropTypes.oneOf(['sm', 'md', 'lg']),
  className: PropTypes.string
};

// Card Description Component
const CardDescription = ({ children, className = '', ...props }) => (
  <p
    className={className}
    style={{
      margin: `${tokens.spacing[2]} 0 0 0`,
      fontSize: tokens.typography.fontSize.sm[0],
      color: tokens.colors.neutral[600],
      lineHeight: tokens.typography.lineHeight.relaxed
    }}
    {...props}
  >
    {children}
  </p>
);

CardDescription.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string
};

// Export all components
Card.Header = CardHeader;
Card.Body = CardBody;
Card.Footer = CardFooter;
Card.Title = CardTitle;
Card.Description = CardDescription;

export default Card;
