import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Alert, ProgressBar, Modal, Spinner } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faUser,
  faEnvelope,
  faLock,
  faBuilding,
  faCheck,
  faEye,
  faEyeSlash,
  faArrowLeft,
  faArrowRight,
  faShieldAlt,
  faUsers,
  faServer,
  faCog,
  faCheckCircle,
  faExclamationTriangle,
  faInfoCircle
} from '@fortawesome/free-solid-svg-icons';
import authService from '../services/auth.js';
import '../styles/design-system.css';

const RegistrationFlow = ({ onRegistrationComplete, onCancel, existingUser = null }) => {
  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState({
    // Step 1: Basic Information
    firstName: '',
    lastName: '',
    email: '',
    username: '',
    
    // Step 2: Security
    password: '',
    confirmPassword: '',
    enableTwoFactor: false,
    securityQuestions: [
      { question: '', answer: '' },
      { question: '', answer: '' }
    ],
    
    // Step 3: Organization
    organizationName: '',
    role: 'user',
    department: '',
    teamSize: '1-5',
    
    // Step 4: Preferences
    theme: 'system',
    language: 'en',
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    notifications: {
      email: true,
      browser: true,
      sms: false
    },
    
    // Step 5: Permissions
    permissions: {
      viewNodes: true,
      manageModels: false,
      adminAccess: false,
      apiAccess: true
    }
  });

  const [validation, setValidation] = useState({});
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [verificationSent, setVerificationSent] = useState(false);
  const [availabilityCheck, setAvailabilityCheck] = useState({});

  const steps = [
    {
      id: 1,
      title: 'Basic Information',
      description: 'Tell us about yourself',
      icon: faUser,
      fields: ['firstName', 'lastName', 'email', 'username']
    },
    {
      id: 2,
      title: 'Security Setup',
      description: 'Secure your account',
      icon: faShieldAlt,
      fields: ['password', 'confirmPassword']
    },
    {
      id: 3,
      title: 'Organization',
      description: 'Your workplace details',
      icon: faBuilding,
      fields: ['organizationName', 'role', 'department']
    },
    {
      id: 4,
      title: 'Preferences',
      description: 'Customize your experience',
      icon: faCog,
      fields: ['theme', 'language', 'notifications']
    },
    {
      id: 5,
      title: 'Permissions',
      description: 'Set access levels',
      icon: faUsers,
      fields: ['permissions']
    }
  ];

  const securityQuestionOptions = [
    "What was the name of your first pet?",
    "What city were you born in?",
    "What was your mother's maiden name?",
    "What was the name of your elementary school?",
    "What is your favorite movie?",
    "What was the make of your first car?",
    "What is your favorite food?",
    "In what city did you meet your spouse/significant other?"
  ];

  useEffect(() => {
    // Pre-fill data if editing existing user
    if (existingUser) {
      setFormData(prevData => ({
        ...prevData,
        ...existingUser,
        password: '', // Never pre-fill passwords
        confirmPassword: ''
      }));
    }
  }, [existingUser]);

  // Real-time validation
  const validateField = (name, value) => {
    const errors = {};

    switch (name) {
      case 'firstName':
      case 'lastName':
        if (!value.trim()) {
          errors[name] = 'This field is required';
        } else if (value.length < 2) {
          errors[name] = 'Must be at least 2 characters';
        }
        break;

      case 'email':
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!value) {
          errors[name] = 'Email is required';
        } else if (!emailRegex.test(value)) {
          errors[name] = 'Invalid email format';
        }
        break;

      case 'username':
        const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/;
        if (!value) {
          errors[name] = 'Username is required';
        } else if (!usernameRegex.test(value)) {
          errors[name] = 'Username must be 3-20 characters, letters, numbers, and underscore only';
        }
        break;

      case 'password':
        if (!value) {
          errors[name] = 'Password is required';
        } else if (value.length < 8) {
          errors[name] = 'Password must be at least 8 characters';
        } else if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(value)) {
          errors[name] = 'Password must contain lowercase, uppercase, and number';
        }
        break;

      case 'confirmPassword':
        if (value !== formData.password) {
          errors[name] = 'Passwords do not match';
        }
        break;

      case 'organizationName':
        if (!value.trim()) {
          errors[name] = 'Organization name is required';
        }
        break;

      default:
        break;
    }

    return errors;
  };

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    const fieldValue = type === 'checkbox' ? checked : value;

    setFormData(prev => ({
      ...prev,
      [name]: fieldValue
    }));

    // Real-time validation
    const fieldErrors = validateField(name, fieldValue);
    setValidation(prev => ({
      ...prev,
      [name]: fieldErrors[name] || null
    }));

    // Check availability for username/email
    if ((name === 'username' || name === 'email') && fieldValue && !fieldErrors[name]) {
      checkAvailability(name, fieldValue);
    }
  };

  const handleNestedChange = (path, value) => {
    const keys = path.split('.');
    setFormData(prev => {
      const newData = { ...prev };
      let current = newData;
      for (let i = 0; i < keys.length - 1; i++) {
        current[keys[i]] = { ...current[keys[i]] };
        current = current[keys[i]];
      }
      current[keys[keys.length - 1]] = value;
      return newData;
    });
  };

  const checkAvailability = async (field, value) => {
    setAvailabilityCheck(prev => ({ ...prev, [field]: 'checking' }));
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      const isAvailable = Math.random() > 0.3; // Simulate availability
      
      setAvailabilityCheck(prev => ({ 
        ...prev, 
        [field]: isAvailable ? 'available' : 'unavailable' 
      }));
      
      if (!isAvailable) {
        setValidation(prev => ({
          ...prev,
          [field]: `This ${field} is already taken`
        }));
      }
    } catch (error) {
      setAvailabilityCheck(prev => ({ ...prev, [field]: 'error' }));
    }
  };

  const validateStep = (stepNumber) => {
    const step = steps.find(s => s.id === stepNumber);
    const stepErrors = {};

    step.fields.forEach(field => {
      const errors = validateField(field, formData[field]);
      if (errors[field]) {
        stepErrors[field] = errors[field];
      }
    });

    // Additional validations
    if (stepNumber === 2) {
      if (!formData.confirmPassword) {
        stepErrors.confirmPassword = 'Please confirm your password';
      } else if (formData.password !== formData.confirmPassword) {
        stepErrors.confirmPassword = 'Passwords do not match';
      }
    }

    setValidation(stepErrors);
    return Object.keys(stepErrors).length === 0;
  };

  const handleNext = () => {
    if (validateStep(currentStep)) {
      setCurrentStep(prev => Math.min(prev + 1, steps.length));
    }
  };

  const handleBack = () => {
    setCurrentStep(prev => Math.max(prev - 1, 1));
  };

  const handleSubmit = async () => {
    if (!validateStep(currentStep)) return;

    setIsLoading(true);
    setErrors({});

    try {
      // Prepare registration data
      const registrationData = {
        username: formData.username,
        email: formData.email,
        password: formData.password,
        full_name: `${formData.firstName} ${formData.lastName}`,
        first_name: formData.firstName,
        last_name: formData.lastName,
        role: formData.role || 'user',
        organization: formData.organization,
        department: formData.department,
        phone: formData.phone,
        preferences: {
          theme: formData.preferences?.theme || 'light',
          notifications: formData.preferences?.notifications || true,
          language: formData.preferences?.language || 'en'
        },
        metadata: {
          registration_source: 'web_ui',
          agreed_to_terms: formData.agreeToTerms,
          newsletter_subscription: formData.subscribeNewsletter
        }
      };

      // Call registration API
      const response = await authService.register(registrationData);

      if (response.success) {
        setVerificationSent(true);

        // Call completion callback after a brief delay
        setTimeout(() => {
          onRegistrationComplete({
            ...registrationData,
            id: response.user_id,
            createdAt: new Date().toISOString()
          });
        }, 3000);
      } else {
        throw new Error(response.message || 'Registration failed');
      }

    } catch (error) {
      console.error('Registration error:', error);
      setErrors({
        submit: error.message || 'Registration failed. Please try again.'
      });
    } finally {
      setIsLoading(false);
    }
  };

  const getPasswordStrength = (password) => {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (password.match(/[a-z]/)) strength += 25;
    if (password.match(/[A-Z]/)) strength += 25;
    if (password.match(/[0-9]/)) strength += 25;
    if (password.match(/[^a-zA-Z0-9]/)) strength += 25;
    return Math.min(strength, 100);
  };

  const renderStepIndicator = () => (
    <div className="registration-steps mb-4">
      <div className="steps-container">
        {steps.map((step, index) => (
          <div key={step.id} className="step-item">
            <div className={`step-circle ${currentStep === step.id ? 'active' : ''} ${currentStep > step.id ? 'completed' : ''}`}>
              {currentStep > step.id ? (
                <FontAwesomeIcon icon={faCheck} />
              ) : (
                <FontAwesomeIcon icon={step.icon} />
              )}
            </div>
            <div className="step-content">
              <div className="step-title">{step.title}</div>
              <div className="step-description">{step.description}</div>
            </div>
            {index < steps.length - 1 && <div className="step-connector" />}
          </div>
        ))}
      </div>
      <ProgressBar 
        now={(currentStep / steps.length) * 100} 
        className="mt-3"
        variant="primary"
      />
    </div>
  );

  const renderStep1 = () => (
    <div className="registration-step">
      <h3>Basic Information</h3>
      <p className="text-muted mb-4">Let's start with the basics. This information helps us personalize your experience.</p>
      
      <Form>
        <div className="row">
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="firstName">
                First Name <span className="text-danger">*</span>
              </Form.Label>
              <Form.Control
                id="firstName"
                name="firstName"
                type="text"
                value={formData.firstName}
                onChange={handleInputChange}
                isInvalid={!!validation.firstName}
                placeholder="Enter your first name"
                autoComplete="given-name"
                aria-describedby="firstName-feedback"
              />
              <Form.Control.Feedback type="invalid" id="firstName-feedback">
                {validation.firstName}
              </Form.Control.Feedback>
            </Form.Group>
          </div>
          
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="lastName">
                Last Name <span className="text-danger">*</span>
              </Form.Label>
              <Form.Control
                id="lastName"
                name="lastName"
                type="text"
                value={formData.lastName}
                onChange={handleInputChange}
                isInvalid={!!validation.lastName}
                placeholder="Enter your last name"
                autoComplete="family-name"
                aria-describedby="lastName-feedback"
              />
              <Form.Control.Feedback type="invalid" id="lastName-feedback">
                {validation.lastName}
              </Form.Control.Feedback>
            </Form.Group>
          </div>
        </div>

        <Form.Group className="mb-3">
          <Form.Label htmlFor="email">
            Email Address <span className="text-danger">*</span>
          </Form.Label>
          <div className="input-group">
            <span className="input-group-text">
              <FontAwesomeIcon icon={faEnvelope} />
            </span>
            <Form.Control
              id="email"
              name="email"
              type="email"
              value={formData.email}
              onChange={handleInputChange}
              isInvalid={!!validation.email}
              placeholder="your.email@company.com"
              autoComplete="email"
              aria-describedby="email-feedback email-help"
            />
            {availabilityCheck.email === 'checking' && (
              <span className="input-group-text">
                <Spinner animation="border" size="sm" />
              </span>
            )}
            {availabilityCheck.email === 'available' && (
              <span className="input-group-text text-success">
                <FontAwesomeIcon icon={faCheckCircle} />
              </span>
            )}
          </div>
          <Form.Control.Feedback type="invalid" id="email-feedback">
            {validation.email}
          </Form.Control.Feedback>
          <Form.Text id="email-help">
            We'll use this for account notifications and verification.
          </Form.Text>
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label htmlFor="username">
            Username <span className="text-danger">*</span>
          </Form.Label>
          <div className="input-group">
            <span className="input-group-text">@</span>
            <Form.Control
              id="username"
              name="username"
              type="text"
              value={formData.username}
              onChange={handleInputChange}
              isInvalid={!!validation.username}
              placeholder="choose_username"
              autoComplete="username"
              aria-describedby="username-feedback username-help"
            />
            {availabilityCheck.username === 'checking' && (
              <span className="input-group-text">
                <Spinner animation="border" size="sm" />
              </span>
            )}
            {availabilityCheck.username === 'available' && (
              <span className="input-group-text text-success">
                <FontAwesomeIcon icon={faCheckCircle} />
              </span>
            )}
          </div>
          <Form.Control.Feedback type="invalid" id="username-feedback">
            {validation.username}
          </Form.Control.Feedback>
          <Form.Text id="username-help">
            3-20 characters. Letters, numbers, and underscores only.
          </Form.Text>
        </Form.Group>
      </Form>
    </div>
  );

  const renderStep2 = () => (
    <div className="registration-step">
      <h3>Security Setup</h3>
      <p className="text-muted mb-4">Keep your account secure with a strong password and optional two-factor authentication.</p>
      
      <Form>
        <Form.Group className="mb-3">
          <Form.Label htmlFor="password">
            Password <span className="text-danger">*</span>
          </Form.Label>
          <div className="input-group">
            <span className="input-group-text">
              <FontAwesomeIcon icon={faLock} />
            </span>
            <Form.Control
              id="password"
              name="password"
              type={showPassword ? 'text' : 'password'}
              value={formData.password}
              onChange={handleInputChange}
              isInvalid={!!validation.password}
              placeholder="Create a strong password"
              autoComplete="new-password"
              aria-describedby="password-feedback password-help"
            />
            <Button
              variant="outline-secondary"
              onClick={() => setShowPassword(!showPassword)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              <FontAwesomeIcon icon={showPassword ? faEyeSlash : faEye} />
            </Button>
          </div>
          <Form.Control.Feedback type="invalid" id="password-feedback">
            {validation.password}
          </Form.Control.Feedback>
          {formData.password && (
            <div className="mt-2">
              <div className="password-strength-meter">
                <div className="strength-bar">
                  <div 
                    className="strength-fill"
                    style={{
                      width: `${getPasswordStrength(formData.password)}%`,
                      backgroundColor: getPasswordStrength(formData.password) < 50 ? '#dc3545' : 
                                     getPasswordStrength(formData.password) < 75 ? '#ffc107' : '#28a745'
                    }}
                  />
                </div>
                <small className="form-text">
                  Strength: {getPasswordStrength(formData.password) < 50 ? 'Weak' : 
                           getPasswordStrength(formData.password) < 75 ? 'Medium' : 'Strong'}
                </small>
              </div>
            </div>
          )}
          <Form.Text id="password-help">
            At least 8 characters with uppercase, lowercase, and numbers.
          </Form.Text>
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label htmlFor="confirmPassword">
            Confirm Password <span className="text-danger">*</span>
          </Form.Label>
          <div className="input-group">
            <span className="input-group-text">
              <FontAwesomeIcon icon={faLock} />
            </span>
            <Form.Control
              id="confirmPassword"
              name="confirmPassword"
              type={showConfirmPassword ? 'text' : 'password'}
              value={formData.confirmPassword}
              onChange={handleInputChange}
              isInvalid={!!validation.confirmPassword}
              placeholder="Confirm your password"
              autoComplete="new-password"
              aria-describedby="confirmPassword-feedback"
            />
            <Button
              variant="outline-secondary"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              aria-label={showConfirmPassword ? 'Hide confirm password' : 'Show confirm password'}
            >
              <FontAwesomeIcon icon={showConfirmPassword ? faEyeSlash : faEye} />
            </Button>
          </div>
          <Form.Control.Feedback type="invalid" id="confirmPassword-feedback">
            {validation.confirmPassword}
          </Form.Control.Feedback>
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Check
            type="switch"
            id="enableTwoFactor"
            name="enableTwoFactor"
            label="Enable Two-Factor Authentication (Recommended)"
            checked={formData.enableTwoFactor}
            onChange={handleInputChange}
          />
          <Form.Text>
            Add an extra layer of security to your account with 2FA.
          </Form.Text>
        </Form.Group>
      </Form>
    </div>
  );

  const renderStep3 = () => (
    <div className="registration-step">
      <h3>Organization Details</h3>
      <p className="text-muted mb-4">Tell us about your organization to help us customize your experience.</p>
      
      <Form>
        <Form.Group className="mb-3">
          <Form.Label htmlFor="organizationName">
            Organization Name <span className="text-danger">*</span>
          </Form.Label>
          <div className="input-group">
            <span className="input-group-text">
              <FontAwesomeIcon icon={faBuilding} />
            </span>
            <Form.Control
              id="organizationName"
              name="organizationName"
              type="text"
              value={formData.organizationName}
              onChange={handleInputChange}
              isInvalid={!!validation.organizationName}
              placeholder="Your company or organization name"
              autoComplete="organization"
              aria-describedby="organizationName-feedback"
            />
          </div>
          <Form.Control.Feedback type="invalid" id="organizationName-feedback">
            {validation.organizationName}
          </Form.Control.Feedback>
        </Form.Group>

        <div className="row">
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="role">Your Role</Form.Label>
              <Form.Select
                id="role"
                name="role"
                value={formData.role}
                onChange={handleInputChange}
                aria-describedby="role-help"
              >
                <option value="user">Team Member</option>
                <option value="lead">Team Lead</option>
                <option value="manager">Manager</option>
                <option value="admin">Administrator</option>
                <option value="developer">Developer</option>
                <option value="analyst">Data Analyst</option>
                <option value="researcher">Researcher</option>
                <option value="other">Other</option>
              </Form.Select>
              <Form.Text id="role-help">
                This helps us customize permissions and features.
              </Form.Text>
            </Form.Group>
          </div>
          
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="department">Department (Optional)</Form.Label>
              <Form.Control
                id="department"
                name="department"
                type="text"
                value={formData.department}
                onChange={handleInputChange}
                placeholder="e.g., Engineering, Data Science"
                autoComplete="organization-title"
              />
            </Form.Group>
          </div>
        </div>

        <Form.Group className="mb-3">
          <Form.Label htmlFor="teamSize">Team Size</Form.Label>
          <Form.Select
            id="teamSize"
            name="teamSize"
            value={formData.teamSize}
            onChange={handleInputChange}
          >
            <option value="1-5">1-5 people</option>
            <option value="6-10">6-10 people</option>
            <option value="11-25">11-25 people</option>
            <option value="26-50">26-50 people</option>
            <option value="51-100">51-100 people</option>
            <option value="100+">100+ people</option>
          </Form.Select>
        </Form.Group>
      </Form>
    </div>
  );

  const renderStep4 = () => (
    <div className="registration-step">
      <h3>Preferences</h3>
      <p className="text-muted mb-4">Customize your experience with these preference settings.</p>
      
      <Form>
        <div className="row">
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="theme">Theme Preference</Form.Label>
              <Form.Select
                id="theme"
                name="theme"
                value={formData.theme}
                onChange={handleInputChange}
              >
                <option value="system">System Default</option>
                <option value="light">Light Theme</option>
                <option value="dark">Dark Theme</option>
              </Form.Select>
            </Form.Group>
          </div>
          
          <div className="col-md-6">
            <Form.Group className="mb-3">
              <Form.Label htmlFor="language">Language</Form.Label>
              <Form.Select
                id="language"
                name="language"
                value={formData.language}
                onChange={handleInputChange}
              >
                <option value="en">English</option>
                <option value="es">Español</option>
                <option value="fr">Français</option>
                <option value="de">Deutsch</option>
                <option value="zh">中文</option>
                <option value="ja">日本語</option>
              </Form.Select>
            </Form.Group>
          </div>
        </div>

        <Form.Group className="mb-3">
          <Form.Label htmlFor="timezone">Timezone</Form.Label>
          <Form.Select
            id="timezone"
            name="timezone"
            value={formData.timezone}
            onChange={handleInputChange}
          >
            <option value="America/New_York">Eastern Time (EST/EDT)</option>
            <option value="America/Chicago">Central Time (CST/CDT)</option>
            <option value="America/Denver">Mountain Time (MST/MDT)</option>
            <option value="America/Los_Angeles">Pacific Time (PST/PDT)</option>
            <option value="Europe/London">London (GMT/BST)</option>
            <option value="Europe/Paris">Paris (CET/CEST)</option>
            <option value="Asia/Tokyo">Tokyo (JST)</option>
            <option value="Asia/Shanghai">Shanghai (CST)</option>
            <option value="Australia/Sydney">Sydney (AEST/AEDT)</option>
          </Select>
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label>Notification Preferences</Form.Label>
          <div className="notification-options">
            <Form.Check
              type="switch"
              id="emailNotifications"
              label="Email Notifications"
              checked={formData.notifications.email}
              onChange={(e) => handleNestedChange('notifications.email', e.target.checked)}
            />
            <Form.Text className="d-block mb-2">
              Receive important updates and alerts via email.
            </Form.Text>
            
            <Form.Check
              type="switch"
              id="browserNotifications"
              label="Browser Notifications"
              checked={formData.notifications.browser}
              onChange={(e) => handleNestedChange('notifications.browser', e.target.checked)}
            />
            <Form.Text className="d-block mb-2">
              Get real-time notifications in your browser.
            </Form.Text>
            
            <Form.Check
              type="switch"
              id="smsNotifications"
              label="SMS Notifications (Optional)"
              checked={formData.notifications.sms}
              onChange={(e) => handleNestedChange('notifications.sms', e.target.checked)}
            />
            <Form.Text className="d-block">
              Receive critical alerts via SMS.
            </Form.Text>
          </div>
        </Form.Group>
      </Form>
    </div>
  );

  const renderStep5 = () => (
    <div className="registration-step">
      <h3>Access Permissions</h3>
      <p className="text-muted mb-4">Configure your initial access permissions. These can be modified later by an administrator.</p>
      
      <Alert variant="info" className="mb-4">
        <FontAwesomeIcon icon={faInfoCircle} className="me-2" />
        These permissions can be adjusted later by your organization's administrator.
      </Alert>
      
      <Form>
        <div className="permissions-grid">
          <div className="permission-item">
            <Form.Check
              type="switch"
              id="viewNodes"
              label="View Nodes"
              checked={formData.permissions.viewNodes}
              onChange={(e) => handleNestedChange('permissions.viewNodes', e.target.checked)}
            />
            <Form.Text>View cluster nodes and their status information.</Form.Text>
          </div>
          
          <div className="permission-item">
            <Form.Check
              type="switch"
              id="manageModels"
              label="Manage Models"
              checked={formData.permissions.manageModels}
              onChange={(e) => handleNestedChange('permissions.manageModels', e.target.checked)}
            />
            <Form.Text>Download, deploy, and manage AI models.</Form.Text>
          </div>
          
          <div className="permission-item">
            <Form.Check
              type="switch"
              id="apiAccess"
              label="API Access"
              checked={formData.permissions.apiAccess}
              onChange={(e) => handleNestedChange('permissions.apiAccess', e.target.checked)}
            />
            <Form.Text>Access the REST API for programmatic integration.</Form.Text>
          </div>
          
          <div className="permission-item">
            <Form.Check
              type="switch"
              id="adminAccess"
              label="Administrative Access"
              checked={formData.permissions.adminAccess}
              onChange={(e) => handleNestedChange('permissions.adminAccess', e.target.checked)}
            />
            <Form.Text>Manage users, system settings, and configuration.</Form.Text>
            
            {formData.permissions.adminAccess && (
              <Alert variant="warning" className="mt-2">
                <FontAwesomeIcon icon={faExclamationTriangle} className="me-2" />
                Administrative access requires approval from an existing administrator.
              </Alert>
            )}
          </div>
        </div>
      </Form>
    </div>
  );

  const renderSuccessModal = () => (
    <Modal show={verificationSent} backdrop="static" keyboard={false} centered>
      <Modal.Body className="text-center p-5">
        <div className="success-animation mb-4">
          <FontAwesomeIcon icon={faCheckCircle} size="4x" className="text-success" />
        </div>
        <h4>Registration Successful!</h4>
        <p className="text-muted mb-4">
          We've sent a verification email to <strong>{formData.email}</strong>.
          Please check your inbox and click the verification link to activate your account.
        </p>
        <div className="d-flex justify-content-center">
          <Spinner animation="border" variant="primary" />
        </div>
        <p className="text-muted mt-2 small">
          Redirecting you in a moment...
        </p>
      </Modal.Body>
    </Modal>
  );

  return (
    <div className="registration-flow">
      <style jsx>{`
        .registration-flow {
          max-width: 800px;
          margin: 0 auto;
          padding: 2rem;
        }
        
        .registration-steps {
          margin-bottom: 2rem;
        }
        
        .steps-container {
          display: flex;
          justify-content: space-between;
          margin-bottom: 1rem;
          position: relative;
        }
        
        .step-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          flex: 1;
          text-align: center;
          position: relative;
        }
        
        .step-circle {
          width: 3rem;
          height: 3rem;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--neutral-200);
          color: var(--neutral-600);
          font-size: 1.2rem;
          margin-bottom: 0.5rem;
          transition: all 0.3s ease;
        }
        
        .step-circle.active {
          background: var(--brand-primary);
          color: white;
        }
        
        .step-circle.completed {
          background: var(--success);
          color: white;
        }
        
        .step-content {
          max-width: 120px;
        }
        
        .step-title {
          font-weight: 600;
          font-size: 0.9rem;
          color: var(--text-primary);
        }
        
        .step-description {
          font-size: 0.8rem;
          color: var(--text-tertiary);
        }
        
        .step-connector {
          position: absolute;
          top: 1.5rem;
          left: 50%;
          right: -50%;
          height: 2px;
          background: var(--neutral-300);
          z-index: -1;
        }
        
        .registration-step {
          background: var(--bg-primary);
          border-radius: var(--radius-card);
          padding: 2rem;
          box-shadow: var(--shadow-sm);
          margin-bottom: 2rem;
        }
        
        .registration-step h3 {
          color: var(--text-primary);
          margin-bottom: 0.5rem;
        }
        
        .password-strength-meter {
          margin-top: 0.5rem;
        }
        
        .strength-bar {
          height: 4px;
          background: var(--neutral-200);
          border-radius: 2px;
          overflow: hidden;
          margin-bottom: 0.25rem;
        }
        
        .strength-fill {
          height: 100%;
          transition: all 0.3s ease;
        }
        
        .notification-options {
          padding: 1rem;
          background: var(--bg-subtle);
          border-radius: var(--radius-md);
        }
        
        .permissions-grid {
          display: grid;
          gap: 1.5rem;
        }
        
        .permission-item {
          padding: 1rem;
          background: var(--bg-subtle);
          border-radius: var(--radius-md);
          border: 1px solid var(--border-primary);
        }
        
        .success-animation {
          animation: bounceIn 0.6s ease-out;
        }
        
        @keyframes bounceIn {
          0% { transform: scale(0.3); opacity: 0; }
          50% { transform: scale(1.05); }
          70% { transform: scale(0.9); }
          100% { transform: scale(1); opacity: 1; }
        }
        
        .navigation-buttons {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1.5rem 0;
          border-top: 1px solid var(--border-primary);
        }
        
        @media (max-width: 768px) {
          .registration-flow {
            padding: 1rem;
          }
          
          .steps-container {
            flex-direction: column;
            gap: 1rem;
          }
          
          .step-item {
            flex-direction: row;
            text-align: left;
          }
          
          .step-circle {
            margin-right: 1rem;
            margin-bottom: 0;
            width: 2.5rem;
            height: 2.5rem;
          }
          
          .step-connector {
            display: none;
          }
        }
      `}</style>

      <Card>
        <Card.Body>
          {renderStepIndicator()}
          
          {currentStep === 1 && renderStep1()}
          {currentStep === 2 && renderStep2()}
          {currentStep === 3 && renderStep3()}
          {currentStep === 4 && renderStep4()}
          {currentStep === 5 && renderStep5()}
          
          <div className="navigation-buttons">
            <div>
              {currentStep > 1 && (
                <Button variant="outline-secondary" onClick={handleBack}>
                  <FontAwesomeIcon icon={faArrowLeft} className="me-2" />
                  Back
                </Button>
              )}
            </div>
            
            <div>
              <Button variant="outline-secondary" onClick={onCancel} className="me-3">
                Cancel
              </Button>
              
              {currentStep < steps.length ? (
                <Button variant="primary" onClick={handleNext}>
                  Next
                  <FontAwesomeIcon icon={faArrowRight} className="ms-2" />
                </Button>
              ) : (
                <Button 
                  variant="success" 
                  onClick={handleSubmit}
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <>
                      <Spinner animation="border" size="sm" className="me-2" />
                      Creating Account...
                    </>
                  ) : (
                    <>
                      <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                      Create Account
                    </>
                  )}
                </Button>
              )}
            </div>
          </div>
        </Card.Body>
      </Card>
      
      {renderSuccessModal()}
    </div>
  );
};

export default RegistrationFlow;