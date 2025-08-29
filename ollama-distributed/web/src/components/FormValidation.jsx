import React, { useState, useEffect, useCallback, createContext, useContext } from 'react';
import { Form, Alert, Spinner } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faCheckCircle,
  faExclamationCircle,
  faInfoCircle,
  faEye,
  faEyeSlash,
  faSpinner
} from '@fortawesome/free-solid-svg-icons';

// Validation Context
const ValidationContext = createContext();

export const useValidation = () => {
  const context = useContext(ValidationContext);
  if (!context) {
    throw new Error('useValidation must be used within a ValidationProvider');
  }
  return context;
};

// Validation rules
export const validationRules = {
  required: (message = 'This field is required') => (value) => {
    if (value === null || value === undefined || value === '') {
      return message;
    }
    if (typeof value === 'string' && value.trim() === '') {
      return message;
    }
    return null;
  },

  minLength: (min, message) => (value) => {
    if (!value) return null;
    if (value.length < min) {
      return message || `Must be at least ${min} characters`;
    }
    return null;
  },

  maxLength: (max, message) => (value) => {
    if (!value) return null;
    if (value.length > max) {
      return message || `Must be no more than ${max} characters`;
    }
    return null;
  },

  email: (message = 'Invalid email format') => (value) => {
    if (!value) return null;
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(value) ? null : message;
  },

  pattern: (regex, message) => (value) => {
    if (!value) return null;
    return regex.test(value) ? null : message;
  },

  passwordStrength: (message = 'Password must contain uppercase, lowercase, number, and special character') => (value) => {
    if (!value) return null;
    const hasUpper = /[A-Z]/.test(value);
    const hasLower = /[a-z]/.test(value);
    const hasNumber = /\d/.test(value);
    const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(value);
    
    if (value.length < 8) return 'Password must be at least 8 characters';
    if (!(hasUpper && hasLower && hasNumber)) {
      return 'Password must contain uppercase, lowercase, and number';
    }
    return null;
  },

  match: (otherValue, fieldName, message) => (value) => {
    if (!value) return null;
    return value === otherValue ? null : (message || `Must match ${fieldName}`);
  },

  url: (message = 'Invalid URL format') => (value) => {
    if (!value) return null;
    try {
      new URL(value);
      return null;
    } catch {
      return message;
    }
  },

  phone: (message = 'Invalid phone number format') => (value) => {
    if (!value) return null;
    const phoneRegex = /^[\+]?[\d\s\-\(\)]{10,}$/;
    return phoneRegex.test(value) ? null : message;
  },

  range: (min, max, message) => (value) => {
    if (!value) return null;
    const num = Number(value);
    if (isNaN(num)) return 'Must be a number';
    if (num < min || num > max) {
      return message || `Must be between ${min} and ${max}`;
    }
    return null;
  },

  custom: (validator, message) => (value) => {
    try {
      const result = validator(value);
      return result ? null : message;
    } catch (error) {
      return message || 'Validation failed';
    }
  }
};

// Async validation (for checking availability, etc.)
export const asyncValidationRules = {
  unique: (checkFunction, message = 'This value is already taken') => async (value) => {
    if (!value) return null;
    try {
      const isUnique = await checkFunction(value);
      return isUnique ? null : message;
    } catch (error) {
      return 'Unable to verify uniqueness';
    }
  },

  exists: (checkFunction, message = 'This value does not exist') => async (value) => {
    if (!value) return null;
    try {
      const exists = await checkFunction(value);
      return exists ? null : message;
    } catch (error) {
      return 'Unable to verify existence';
    }
  }
};

// Validation Provider
export const ValidationProvider = ({ children }) => {
  const [fields, setFields] = useState({});
  const [isValidating, setIsValidating] = useState({});
  const [hasErrors, setHasErrors] = useState(false);

  const registerField = useCallback((name, rules = [], asyncRules = [], initialValue = '') => {
    setFields(prev => ({
      ...prev,
      [name]: {
        value: initialValue,
        error: null,
        touched: false,
        validating: false,
        rules,
        asyncRules
      }
    }));
  }, []);

  const unregisterField = useCallback((name) => {
    setFields(prev => {
      const newFields = { ...prev };
      delete newFields[name];
      return newFields;
    });
  }, []);

  const setValue = useCallback((name, value) => {
    setFields(prev => ({
      ...prev,
      [name]: {
        ...prev[name],
        value,
        touched: true
      }
    }));
  }, []);

  const validateField = useCallback(async (name) => {
    const field = fields[name];
    if (!field) return;

    setIsValidating(prev => ({ ...prev, [name]: true }));
    
    // Run sync validations
    let error = null;
    for (const rule of field.rules) {
      error = rule(field.value);
      if (error) break;
    }

    // Run async validations if sync passed
    if (!error && field.asyncRules.length > 0) {
      setFields(prev => ({
        ...prev,
        [name]: { ...prev[name], validating: true }
      }));

      for (const asyncRule of field.asyncRules) {
        try {
          error = await asyncRule(field.value);
          if (error) break;
        } catch (err) {
          error = 'Validation failed';
          break;
        }
      }

      setFields(prev => ({
        ...prev,
        [name]: { ...prev[name], validating: false }
      }));
    }

    setFields(prev => ({
      ...prev,
      [name]: { ...prev[name], error }
    }));

    setIsValidating(prev => ({ ...prev, [name]: false }));
    return !error;
  }, [fields]);

  const validateAll = useCallback(async () => {
    const validationPromises = Object.keys(fields).map(validateField);
    const results = await Promise.all(validationPromises);
    return results.every(Boolean);
  }, [fields, validateField]);

  const getFieldError = useCallback((name) => {
    return fields[name]?.error;
  }, [fields]);

  const isFieldValid = useCallback((name) => {
    const field = fields[name];
    return field && field.touched && !field.error;
  }, [fields]);

  const isFieldInvalid = useCallback((name) => {
    const field = fields[name];
    return field && field.touched && !!field.error;
  }, [fields]);

  const clearErrors = useCallback(() => {
    setFields(prev => {
      const newFields = {};
      Object.keys(prev).forEach(name => {
        newFields[name] = { ...prev[name], error: null };
      });
      return newFields;
    });
  }, []);

  const reset = useCallback(() => {
    setFields({});
    setIsValidating({});
    setHasErrors(false);
  }, []);

  // Update hasErrors when fields change
  useEffect(() => {
    const hasAnyErrors = Object.values(fields).some(field => !!field.error);
    setHasErrors(hasAnyErrors);
  }, [fields]);

  const contextValue = {
    fields,
    isValidating,
    hasErrors,
    registerField,
    unregisterField,
    setValue,
    validateField,
    validateAll,
    getFieldError,
    isFieldValid,
    isFieldInvalid,
    clearErrors,
    reset
  };

  return (
    <ValidationContext.Provider value={contextValue}>
      {children}
    </ValidationContext.Provider>
  );
};

// Enhanced Form Field Component
export const ValidatedField = ({
  name,
  label,
  type = 'text',
  rules = [],
  asyncRules = [],
  helpText,
  placeholder,
  options = [], // for select fields
  children,
  showPasswordToggle = false,
  debounceMs = 300,
  validateOnChange = true,
  validateOnBlur = true,
  ...props
}) => {
  const {
    registerField,
    unregisterField,
    setValue,
    validateField,
    getFieldError,
    isFieldValid,
    isFieldInvalid,
    fields
  } = useValidation();

  const [showPassword, setShowPassword] = useState(false);
  const [validationTimeout, setValidationTimeout] = useState(null);

  const field = fields[name];
  const value = field?.value || '';
  const error = getFieldError(name);
  const isValid = isFieldValid(name);
  const isInvalid = isFieldInvalid(name);
  const isValidating = field?.validating;

  // Register field on mount
  useEffect(() => {
    registerField(name, rules, asyncRules, props.defaultValue || '');
    return () => unregisterField(name);
  }, [name, registerField, unregisterField, props.defaultValue]);

  // Debounced validation
  const debouncedValidate = useCallback(() => {
    if (validationTimeout) {
      clearTimeout(validationTimeout);
    }

    const timeout = setTimeout(() => {
      validateField(name);
    }, debounceMs);

    setValidationTimeout(timeout);
  }, [name, validateField, debounceMs, validationTimeout]);

  const handleChange = (e) => {
    const newValue = e.target.value;
    setValue(name, newValue);

    if (validateOnChange) {
      debouncedValidate();
    }

    if (props.onChange) {
      props.onChange(e);
    }
  };

  const handleBlur = (e) => {
    if (validateOnBlur) {
      validateField(name);
    }

    if (props.onBlur) {
      props.onBlur(e);
    }
  };

  const renderField = () => {
    const commonProps = {
      ...props,
      id: name,
      name,
      value,
      onChange: handleChange,
      onBlur: handleBlur,
      isValid,
      isInvalid,
      placeholder
    };

    switch (type) {
      case 'select':
        return (
          <Form.Select {...commonProps}>
            <option value="">Choose...</option>
            {options.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Form.Select>
        );

      case 'textarea':
        return <Form.Control as="textarea" {...commonProps} />;

      case 'checkbox':
        return (
          <Form.Check
            {...commonProps}
            type="checkbox"
            checked={value}
            onChange={(e) => {
              setValue(name, e.target.checked);
              if (props.onChange) props.onChange(e);
            }}
          />
        );

      case 'radio':
        return (
          <div>
            {options.map((option) => (
              <Form.Check
                key={option.value}
                type="radio"
                id={`${name}-${option.value}`}
                name={name}
                value={option.value}
                label={option.label}
                checked={value === option.value}
                onChange={handleChange}
              />
            ))}
          </div>
        );

      case 'password':
        return (
          <div className="input-group">
            <Form.Control
              {...commonProps}
              type={showPassword ? 'text' : 'password'}
            />
            {showPasswordToggle && (
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                <FontAwesomeIcon icon={showPassword ? faEyeSlash : faEye} />
              </button>
            )}
            {isValidating && (
              <span className="input-group-text">
                <Spinner animation="border" size="sm" />
              </span>
            )}
          </div>
        );

      default:
        return (
          <div className="input-group">
            <Form.Control {...commonProps} type={type} />
            {isValidating && (
              <span className="input-group-text">
                <FontAwesomeIcon icon={faSpinner} spin />
              </span>
            )}
            {isValid && !isValidating && (
              <span className="input-group-text text-success">
                <FontAwesomeIcon icon={faCheckCircle} />
              </span>
            )}
          </div>
        );
    }
  };

  return (
    <Form.Group className="mb-3">
      {label && (
        <Form.Label htmlFor={name}>
          {label}
          {rules.some(rule => rule.name === 'required') && (
            <span className="text-danger ms-1">*</span>
          )}
        </Form.Label>
      )}

      {children || renderField()}

      {error && (
        <Form.Control.Feedback type="invalid" className="d-block">
          <FontAwesomeIcon icon={faExclamationCircle} className="me-1" />
          {error}
        </Form.Control.Feedback>
      )}

      {helpText && !error && (
        <Form.Text className="text-muted">
          <FontAwesomeIcon icon={faInfoCircle} className="me-1" />
          {helpText}
        </Form.Text>
      )}
    </Form.Group>
  );
};

// Form Summary Component
export const ValidationSummary = ({ title = 'Please fix the following errors:' }) => {
  const { fields, hasErrors } = useValidation();

  if (!hasErrors) return null;

  const errors = Object.entries(fields)
    .filter(([_, field]) => field.error)
    .map(([name, field]) => ({ name, error: field.error }));

  return (
    <Alert variant="danger">
      <h6>{title}</h6>
      <ul className="mb-0">
        {errors.map(({ name, error }) => (
          <li key={name}>
            <strong>{name}:</strong> {error}
          </li>
        ))}
      </ul>
    </Alert>
  );
};

// Password Strength Indicator
export const PasswordStrength = ({ password }) => {
  const getStrength = (pwd) => {
    let score = 0;
    if (!pwd) return 0;

    // Length
    if (pwd.length >= 8) score += 25;
    if (pwd.length >= 12) score += 25;

    // Character types
    if (/[a-z]/.test(pwd)) score += 10;
    if (/[A-Z]/.test(pwd)) score += 10;
    if (/[0-9]/.test(pwd)) score += 10;
    if (/[^a-zA-Z0-9]/.test(pwd)) score += 20;

    return Math.min(score, 100);
  };

  const strength = getStrength(password);

  const getStrengthLabel = () => {
    if (strength < 30) return 'Weak';
    if (strength < 60) return 'Fair';
    if (strength < 80) return 'Good';
    return 'Strong';
  };

  const getStrengthColor = () => {
    if (strength < 30) return 'danger';
    if (strength < 60) return 'warning';
    if (strength < 80) return 'info';
    return 'success';
  };

  if (!password) return null;

  return (
    <div className="password-strength mt-2">
      <div className="d-flex justify-content-between align-items-center mb-1">
        <small>Password Strength</small>
        <small className={`text-${getStrengthColor()}`}>{getStrengthLabel()}</small>
      </div>
      <div className="progress" style={{ height: '4px' }}>
        <div
          className={`progress-bar bg-${getStrengthColor()}`}
          role="progressbar"
          style={{ width: `${strength}%` }}
        />
      </div>
    </div>
  );
};

// Validation Hook for forms
export const useFormValidation = (initialValues = {}, validationSchema = {}) => {
  const [values, setValues] = useState(initialValues);
  const [errors, setErrors] = useState({});
  const [touched, setTouched] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validate = useCallback(async (fieldName = null) => {
    const fieldsToValidate = fieldName ? [fieldName] : Object.keys(validationSchema);
    const newErrors = { ...errors };

    for (const field of fieldsToValidate) {
      const rules = validationSchema[field] || [];
      let error = null;

      for (const rule of rules) {
        if (typeof rule === 'function') {
          error = await rule(values[field], values);
        } else if (rule.validator) {
          error = await rule.validator(values[field], values);
        }
        
        if (error) break;
      }

      if (error) {
        newErrors[field] = error;
      } else {
        delete newErrors[field];
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [values, validationSchema, errors]);

  const handleChange = useCallback((name, value) => {
    setValues(prev => ({ ...prev, [name]: value }));
    setTouched(prev => ({ ...prev, [name]: true }));
  }, []);

  const handleSubmit = useCallback(async (onSubmit) => {
    setIsSubmitting(true);
    const isValid = await validate();
    
    if (isValid) {
      try {
        await onSubmit(values);
      } catch (error) {
        console.error('Submit error:', error);
      }
    }
    
    setIsSubmitting(false);
    return isValid;
  }, [validate, values]);

  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
    setIsSubmitting(false);
  }, [initialValues]);

  return {
    values,
    errors,
    touched,
    isSubmitting,
    handleChange,
    handleSubmit,
    validate,
    reset,
    isValid: Object.keys(errors).length === 0
  };
};

export default {
  ValidationProvider,
  ValidatedField,
  ValidationSummary,
  PasswordStrength,
  validationRules,
  asyncValidationRules,
  useValidation,
  useFormValidation
};