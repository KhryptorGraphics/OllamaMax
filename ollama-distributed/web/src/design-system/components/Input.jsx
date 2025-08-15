/**
 * Input Component - OllamaMax Design System
 * 
 * A flexible input component with validation, icons, and accessibility features.
 */

import React, { forwardRef, useState } from 'react';
import PropTypes from 'prop-types';
import { tokens } from '../tokens.js';

const Input = forwardRef(({
  type = 'text',
  size = 'md',
  variant = 'default',
  placeholder,
  value,
  defaultValue,
  onChange,
  onFocus,
  onBlur,
  disabled = false,
  readOnly = false,
  required = false,
  error = false,
  success = false,
  leftIcon,
  rightIcon,
  leftAddon,
  rightAddon,
  helperText,
  errorText,
  label,
  id,
  name,
  autoComplete,
  autoFocus = false,
  maxLength,
  minLength,
  pattern,
  className = '',
  'aria-describedby': ariaDescribedBy,
  ...props
}, ref) => {
  const [focused, setFocused] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Generate unique IDs
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
  const helperTextId = `${inputId}-helper`;
  const errorTextId = `${inputId}-error`;

  // Base styles
  const baseStyles = {
    fontFamily: tokens.typography.fontFamily.sans.join(', '),
    fontSize: tokens.typography.fontSize.base[0],
    lineHeight: tokens.typography.lineHeight.normal,
    borderRadius: tokens.borderRadius.md,
    border: '1px solid',
    transition: `all ${tokens.animation.duration.normal} ${tokens.animation.easing.easeInOut}`,
    outline: 'none',
    width: '100%',
    backgroundColor: disabled ? tokens.colors.neutral[50] : tokens.colors.neutral[0]
  };

  // Size variants
  const sizeStyles = {
    sm: {
      height: tokens.components.input.height.sm,
      padding: tokens.components.input.padding.sm,
      fontSize: tokens.typography.fontSize.sm[0]
    },
    md: {
      height: tokens.components.input.height.md,
      padding: tokens.components.input.padding.md,
      fontSize: tokens.typography.fontSize.base[0]
    },
    lg: {
      height: tokens.components.input.height.lg,
      padding: tokens.components.input.padding.lg,
      fontSize: tokens.typography.fontSize.lg[0]
    }
  };

  // State-based styles
  const getStateStyles = () => {
    if (error) {
      return {
        borderColor: tokens.colors.error[500],
        boxShadow: `0 0 0 3px ${tokens.colors.error[100]}`,
        color: tokens.colors.neutral[900]
      };
    }
    
    if (success) {
      return {
        borderColor: tokens.colors.success[500],
        boxShadow: `0 0 0 3px ${tokens.colors.success[100]}`,
        color: tokens.colors.neutral[900]
      };
    }
    
    if (focused) {
      return {
        borderColor: tokens.colors.primary[500],
        boxShadow: `0 0 0 3px ${tokens.colors.primary[100]}`,
        color: tokens.colors.neutral[900]
      };
    }
    
    return {
      borderColor: tokens.colors.neutral[300],
      color: tokens.colors.neutral[900]
    };
  };

  // Container styles for addons and icons
  const containerStyles = {
    position: 'relative',
    display: 'flex',
    alignItems: 'center',
    width: '100%'
  };

  // Input wrapper styles
  const inputWrapperStyles = {
    position: 'relative',
    display: 'flex',
    alignItems: 'center',
    flex: 1
  };

  // Icon styles
  const iconStyles = {
    position: 'absolute',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: tokens.colors.neutral[400],
    pointerEvents: 'none',
    zIndex: 1
  };

  const leftIconStyles = {
    ...iconStyles,
    left: tokens.spacing[3]
  };

  const rightIconStyles = {
    ...iconStyles,
    right: tokens.spacing[3],
    pointerEvents: type === 'password' ? 'auto' : 'none',
    cursor: type === 'password' ? 'pointer' : 'default'
  };

  // Addon styles
  const addonStyles = {
    display: 'flex',
    alignItems: 'center',
    padding: `0 ${tokens.spacing[3]}`,
    backgroundColor: tokens.colors.neutral[50],
    border: '1px solid',
    borderColor: tokens.colors.neutral[300],
    color: tokens.colors.neutral[600],
    fontSize: sizeStyles[size].fontSize,
    whiteSpace: 'nowrap'
  };

  const leftAddonStyles = {
    ...addonStyles,
    borderRight: 'none',
    borderTopLeftRadius: tokens.borderRadius.md,
    borderBottomLeftRadius: tokens.borderRadius.md
  };

  const rightAddonStyles = {
    ...addonStyles,
    borderLeft: 'none',
    borderTopRightRadius: tokens.borderRadius.md,
    borderBottomRightRadius: tokens.borderRadius.md
  };

  // Adjust input styles for icons and addons
  const getInputStyles = () => {
    const styles = {
      ...baseStyles,
      ...sizeStyles[size],
      ...getStateStyles()
    };

    if (leftIcon || leftAddon) {
      styles.paddingLeft = leftIcon ? tokens.spacing[10] : tokens.spacing[3];
      if (leftAddon) {
        styles.borderTopLeftRadius = 0;
        styles.borderBottomLeftRadius = 0;
        styles.borderLeft = 'none';
      }
    }

    if (rightIcon || rightAddon || type === 'password') {
      styles.paddingRight = rightIcon || type === 'password' ? tokens.spacing[10] : tokens.spacing[3];
      if (rightAddon) {
        styles.borderTopRightRadius = 0;
        styles.borderBottomRightRadius = 0;
        styles.borderRight = 'none';
      }
    }

    if (disabled) {
      styles.cursor = 'not-allowed';
      styles.opacity = 0.6;
    }

    return styles;
  };

  // Handle focus
  const handleFocus = (event) => {
    setFocused(true);
    onFocus?.(event);
  };

  // Handle blur
  const handleBlur = (event) => {
    setFocused(false);
    onBlur?.(event);
  };

  // Toggle password visibility
  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  // Password toggle icon
  const PasswordToggleIcon = () => (
    <button
      type="button"
      onClick={togglePasswordVisibility}
      style={{
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        padding: 0,
        display: 'flex',
        alignItems: 'center',
        color: tokens.colors.neutral[400]
      }}
      aria-label={showPassword ? 'Hide password' : 'Show password'}
    >
      {showPassword ? (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
        </svg>
      ) : (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 7c2.76 0 5 2.24 5 5 0 .65-.13 1.26-.36 1.83l2.92 2.92c1.51-1.26 2.7-2.89 3.43-4.75-1.73-4.39-6-7.5-11-7.5-1.4 0-2.74.25-3.98.7l2.16 2.16C10.74 7.13 11.35 7 12 7zM2 4.27l2.28 2.28.46.46C3.08 8.3 1.78 10.02 1 12c1.73 4.39 6 7.5 11 7.5 1.55 0 3.03-.3 4.38-.84l.42.42L19.73 22 21 20.73 3.27 3 2 4.27zM7.53 9.8l1.55 1.55c-.05.21-.08.43-.08.65 0 1.66 1.34 3 3 3 .22 0 .44-.03.65-.08l1.55 1.55c-.67.33-1.41.53-2.2.53-2.76 0-5-2.24-5-5 0-.79.2-1.53.53-2.2zm4.31-.78l3.15 3.15.02-.16c0-1.66-1.34-3-3-3l-.17.01z"/>
        </svg>
      )}
    </button>
  );

  return (
    <div className={className}>
      {label && (
        <label
          htmlFor={inputId}
          style={{
            display: 'block',
            fontSize: tokens.typography.fontSize.sm[0],
            fontWeight: tokens.typography.fontWeight.medium,
            color: tokens.colors.neutral[700],
            marginBottom: tokens.spacing[1]
          }}
        >
          {label}
          {required && (
            <span style={{ color: tokens.colors.error[500], marginLeft: tokens.spacing[1] }}>
              *
            </span>
          )}
        </label>
      )}
      
      <div style={containerStyles}>
        {leftAddon && (
          <div style={leftAddonStyles}>
            {leftAddon}
          </div>
        )}
        
        <div style={inputWrapperStyles}>
          {leftIcon && (
            <div style={leftIconStyles}>
              {leftIcon}
            </div>
          )}
          
          <input
            ref={ref}
            id={inputId}
            name={name}
            type={type === 'password' ? (showPassword ? 'text' : 'password') : type}
            placeholder={placeholder}
            value={value}
            defaultValue={defaultValue}
            onChange={onChange}
            onFocus={handleFocus}
            onBlur={handleBlur}
            disabled={disabled}
            readOnly={readOnly}
            required={required}
            autoComplete={autoComplete}
            autoFocus={autoFocus}
            maxLength={maxLength}
            minLength={minLength}
            pattern={pattern}
            style={getInputStyles()}
            aria-describedby={[
              helperText ? helperTextId : null,
              errorText ? errorTextId : null,
              ariaDescribedBy
            ].filter(Boolean).join(' ') || undefined}
            aria-invalid={error}
            {...props}
          />
          
          {(rightIcon || type === 'password') && (
            <div style={rightIconStyles}>
              {type === 'password' ? <PasswordToggleIcon /> : rightIcon}
            </div>
          )}
        </div>
        
        {rightAddon && (
          <div style={rightAddonStyles}>
            {rightAddon}
          </div>
        )}
      </div>
      
      {(helperText || errorText) && (
        <div
          id={errorText ? errorTextId : helperTextId}
          style={{
            marginTop: tokens.spacing[1],
            fontSize: tokens.typography.fontSize.sm[0],
            color: errorText ? tokens.colors.error[600] : tokens.colors.neutral[600]
          }}
        >
          {errorText || helperText}
        </div>
      )}
    </div>
  );
});

Input.displayName = 'Input';

Input.propTypes = {
  type: PropTypes.oneOf(['text', 'email', 'password', 'number', 'tel', 'url', 'search']),
  size: PropTypes.oneOf(['sm', 'md', 'lg']),
  variant: PropTypes.oneOf(['default', 'filled', 'flushed']),
  placeholder: PropTypes.string,
  value: PropTypes.string,
  defaultValue: PropTypes.string,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  onBlur: PropTypes.func,
  disabled: PropTypes.bool,
  readOnly: PropTypes.bool,
  required: PropTypes.bool,
  error: PropTypes.bool,
  success: PropTypes.bool,
  leftIcon: PropTypes.node,
  rightIcon: PropTypes.node,
  leftAddon: PropTypes.node,
  rightAddon: PropTypes.node,
  helperText: PropTypes.string,
  errorText: PropTypes.string,
  label: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  autoComplete: PropTypes.string,
  autoFocus: PropTypes.bool,
  maxLength: PropTypes.number,
  minLength: PropTypes.number,
  pattern: PropTypes.string,
  className: PropTypes.string,
  'aria-describedby': PropTypes.string
};

export default Input;
