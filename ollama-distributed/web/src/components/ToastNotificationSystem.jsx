import React, { useState, useEffect, createContext, useContext, useCallback } from 'react';
import { Toast, ToastContainer } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faCheckCircle,
  faExclamationTriangle,
  faInfoCircle,
  faTimes,
  faExclamationCircle,
  faBell,
  faRocket,
  faHeart,
  faStar,
  faGift
} from '@fortawesome/free-solid-svg-icons';

// Toast Context for global state management
const ToastContext = createContext();

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

// Toast Provider Component
export const ToastProvider = ({ children, maxToasts = 5 }) => {
  const [toasts, setToasts] = useState([]);
  const [nextId, setNextId] = useState(1);

  const addToast = useCallback((toast) => {
    const id = nextId;
    setNextId(prev => prev + 1);

    const newToast = {
      id,
      title: '',
      message: '',
      type: 'info',
      duration: 5000,
      persistent: false,
      position: 'top-end',
      showProgress: true,
      actions: [],
      ...toast,
      timestamp: Date.now()
    };

    setToasts(prev => {
      const updated = [newToast, ...prev];
      // Limit the number of toasts
      return updated.slice(0, maxToasts);
    });

    // Auto-remove if not persistent
    if (!newToast.persistent && newToast.duration > 0) {
      setTimeout(() => {
        removeToast(id);
      }, newToast.duration);
    }

    return id;
  }, [nextId, maxToasts]);

  const removeToast = useCallback((id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  }, []);

  const clearAllToasts = useCallback(() => {
    setToasts([]);
  }, []);

  const updateToast = useCallback((id, updates) => {
    setToasts(prev => prev.map(toast => 
      toast.id === id ? { ...toast, ...updates } : toast
    ));
  }, []);

  // Convenience methods
  const success = useCallback((message, options = {}) => {
    return addToast({
      type: 'success',
      message,
      title: 'Success',
      ...options
    });
  }, [addToast]);

  const error = useCallback((message, options = {}) => {
    return addToast({
      type: 'error',
      message,
      title: 'Error',
      persistent: true,
      ...options
    });
  }, [addToast]);

  const warning = useCallback((message, options = {}) => {
    return addToast({
      type: 'warning',
      message,
      title: 'Warning',
      duration: 7000,
      ...options
    });
  }, [addToast]);

  const info = useCallback((message, options = {}) => {
    return addToast({
      type: 'info',
      message,
      title: 'Information',
      ...options
    });
  }, [addToast]);

  const promise = useCallback((promiseOrFunction, options = {}) => {
    const {
      loading = 'Loading...',
      success: successMessage = 'Operation completed successfully',
      error: errorMessage = 'Operation failed',
      ...toastOptions
    } = options;

    // Show loading toast
    const loadingToastId = addToast({
      type: 'loading',
      message: loading,
      title: 'Loading',
      persistent: true,
      showProgress: false,
      ...toastOptions
    });

    const handlePromise = async (promise) => {
      try {
        const result = await promise;
        
        // Update to success
        updateToast(loadingToastId, {
          type: 'success',
          message: typeof successMessage === 'function' ? successMessage(result) : successMessage,
          title: 'Success',
          persistent: false,
          duration: 5000
        });

        return result;
      } catch (err) {
        // Update to error
        updateToast(loadingToastId, {
          type: 'error',
          message: typeof errorMessage === 'function' ? errorMessage(err) : errorMessage,
          title: 'Error',
          persistent: true
        });

        throw err;
      }
    };

    // Handle both promises and functions that return promises
    const promise = typeof promiseOrFunction === 'function' ? promiseOrFunction() : promiseOrFunction;
    
    return handlePromise(promise);
  }, [addToast, updateToast]);

  const contextValue = {
    toasts,
    addToast,
    removeToast,
    clearAllToasts,
    updateToast,
    success,
    error,
    warning,
    info,
    promise
  };

  return (
    <ToastContext.Provider value={contextValue}>
      {children}
      <ToastNotificationSystem toasts={toasts} onRemove={removeToast} />
    </ToastContext.Provider>
  );
};

// Individual Toast Component
const ToastNotification = ({ toast, onRemove }) => {
  const [progress, setProgress] = useState(100);
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    if (toast.persistent || !toast.showProgress || toast.duration <= 0) return;

    const startTime = Date.now();
    const interval = setInterval(() => {
      const elapsed = Date.now() - startTime;
      const remaining = Math.max(0, ((toast.duration - elapsed) / toast.duration) * 100);
      setProgress(remaining);

      if (remaining <= 0) {
        clearInterval(interval);
      }
    }, 50);

    return () => clearInterval(interval);
  }, [toast.persistent, toast.showProgress, toast.duration]);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(() => onRemove(toast.id), 300);
  };

  const getIcon = () => {
    switch (toast.type) {
      case 'success':
        return faCheckCircle;
      case 'error':
        return faExclamationCircle;
      case 'warning':
        return faExclamationTriangle;
      case 'info':
        return faInfoCircle;
      case 'loading':
        return faBell;
      default:
        return faInfoCircle;
    }
  };

  const getVariant = () => {
    switch (toast.type) {
      case 'success':
        return 'success';
      case 'error':
        return 'danger';
      case 'warning':
        return 'warning';
      case 'info':
        return 'info';
      case 'loading':
        return 'primary';
      default:
        return 'info';
    }
  };

  return (
    <Toast
      show={isVisible}
      onClose={handleClose}
      className={`toast-notification toast-${toast.type}`}
      bg={getVariant()}
    >
      <style jsx>{`
        .toast-notification {
          min-width: 300px;
          max-width: 400px;
          position: relative;
          overflow: hidden;
        }
        
        .toast-header {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          font-weight: var(--font-weight-medium);
        }
        
        .toast-icon {
          font-size: 1.1rem;
        }
        
        .toast-body {
          padding: 0.75rem 1rem;
          line-height: 1.4;
        }
        
        .toast-actions {
          display: flex;
          gap: 0.5rem;
          margin-top: 0.75rem;
        }
        
        .toast-progress {
          position: absolute;
          bottom: 0;
          left: 0;
          height: 3px;
          background: rgba(255, 255, 255, 0.3);
          transition: width 0.1s linear;
        }
        
        .toast-loading .toast-icon {
          animation: pulse 1.5s ease-in-out infinite;
        }
        
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
        
        .toast-success {
          border-left: 4px solid var(--success);
        }
        
        .toast-error {
          border-left: 4px solid var(--error);
        }
        
        .toast-warning {
          border-left: 4px solid var(--warning);
        }
        
        .toast-info {
          border-left: 4px solid var(--info);
        }
        
        .toast-loading {
          border-left: 4px solid var(--brand-primary);
        }
      `}</style>

      <Toast.Header>
        <FontAwesomeIcon icon={getIcon()} className="toast-icon" />
        <strong className="me-auto">{toast.title}</strong>
        {!toast.persistent && (
          <small className="text-muted">
            {Math.ceil((toast.duration - (Date.now() - toast.timestamp)) / 1000)}s
          </small>
        )}
      </Toast.Header>
      
      <Toast.Body>
        {toast.message}
        
        {toast.actions && toast.actions.length > 0 && (
          <div className="toast-actions">
            {toast.actions.map((action, index) => (
              <button
                key={index}
                type="button"
                className={`btn btn-sm ${action.variant || 'btn-outline-light'}`}
                onClick={action.onClick}
              >
                {action.label}
              </button>
            ))}
          </div>
        )}
      </Toast.Body>
      
      {toast.showProgress && !toast.persistent && toast.duration > 0 && (
        <div
          className="toast-progress"
          style={{ width: `${progress}%` }}
        />
      )}
    </Toast>
  );
};

// Main Toast System Component
const ToastNotificationSystem = ({ toasts, onRemove }) => {
  // Group toasts by position
  const toastsByPosition = toasts.reduce((acc, toast) => {
    const position = toast.position || 'top-end';
    if (!acc[position]) acc[position] = [];
    acc[position].push(toast);
    return acc;
  }, {});

  return (
    <>
      <style jsx global>{`
        .toast-container {
          z-index: var(--z-toast);
        }
        
        .toast-container .toast {
          margin-bottom: 0.5rem;
          animation: slideIn 0.3s ease-out;
        }
        
        @keyframes slideIn {
          from {
            opacity: 0;
            transform: translateX(100%);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }
        
        @keyframes slideOut {
          from {
            opacity: 1;
            transform: translateX(0);
          }
          to {
            opacity: 0;
            transform: translateX(100%);
          }
        }
        
        .toast.hiding {
          animation: slideOut 0.3s ease-in;
        }
        
        /* Position-specific styles */
        .toast-container.position-top-start {
          top: 1rem;
          left: 1rem;
        }
        
        .toast-container.position-top-center {
          top: 1rem;
          left: 50%;
          transform: translateX(-50%);
        }
        
        .toast-container.position-top-end {
          top: 1rem;
          right: 1rem;
        }
        
        .toast-container.position-middle-start {
          top: 50%;
          left: 1rem;
          transform: translateY(-50%);
        }
        
        .toast-container.position-middle-center {
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
        }
        
        .toast-container.position-middle-end {
          top: 50%;
          right: 1rem;
          transform: translateY(-50%);
        }
        
        .toast-container.position-bottom-start {
          bottom: 1rem;
          left: 1rem;
        }
        
        .toast-container.position-bottom-center {
          bottom: 1rem;
          left: 50%;
          transform: translateX(-50%);
        }
        
        .toast-container.position-bottom-end {
          bottom: 1rem;
          right: 1rem;
        }
        
        @media (max-width: 768px) {
          .toast-container {
            left: 1rem !important;
            right: 1rem !important;
            transform: none !important;
          }
          
          .toast-notification {
            min-width: auto;
            width: 100%;
          }
        }
      `}</style>

      {Object.entries(toastsByPosition).map(([position, positionToasts]) => (
        <ToastContainer
          key={position}
          position={position}
          className={`position-fixed position-${position}`}
        >
          {positionToasts.map((toast) => (
            <ToastNotification
              key={toast.id}
              toast={toast}
              onRemove={onRemove}
            />
          ))}
        </ToastContainer>
      ))}
    </>
  );
};

// Utility hook for common toast patterns
export const useToastUtils = () => {
  const { success, error, warning, info, promise } = useToast();

  return {
    // API operation patterns
    apiSuccess: (message = 'Operation completed successfully') => 
      success(message, { title: 'Success' }),
    
    apiError: (error, fallbackMessage = 'Something went wrong') => {
      const message = error?.response?.data?.message || error?.message || fallbackMessage;
      return error(message, { 
        title: 'Error',
        actions: [
          {
            label: 'Retry',
            variant: 'btn-outline-light',
            onClick: () => window.location.reload()
          }
        ]
      });
    },

    // Form validation
    validationError: (message = 'Please check your input and try again') =>
      warning(message, { title: 'Validation Error' }),

    // Network status
    offline: () => warning('You appear to be offline. Some features may not be available.', {
      title: 'Connection Issue',
      persistent: true
    }),

    online: () => success('Connection restored', { title: 'Back Online' }),

    // File operations
    fileUploadSuccess: (filename) => 
      success(`${filename} uploaded successfully`, { title: 'Upload Complete' }),
    
    fileUploadError: (filename, error) => 
      error(`Failed to upload ${filename}: ${error}`, { title: 'Upload Failed' }),

    // User actions
    saveSuccess: () => success('Changes saved successfully', { title: 'Saved' }),
    
    deleteSuccess: (item = 'Item') => 
      success(`${item} deleted successfully`, { title: 'Deleted' }),
    
    // System notifications
    maintenance: (message = 'System maintenance in progress') =>
      warning(message, { 
        title: 'Maintenance', 
        persistent: true,
        actions: [
          {
            label: 'Learn More',
            variant: 'btn-outline-light',
            onClick: () => window.open('/maintenance', '_blank')
          }
        ]
      }),

    // Welcome/celebration
    welcome: (username) => success(`Welcome back, ${username}!`, {
      title: '👋 Welcome',
      duration: 3000
    }),

    achievement: (message) => success(message, {
      title: '🎉 Achievement Unlocked',
      duration: 6000
    }),

    // Loading operations
    loadingOperation: (operation, options = {}) => 
      promise(operation, {
        loading: 'Processing...',
        success: 'Operation completed successfully',
        error: 'Operation failed',
        ...options
      })
  };
};

export default ToastNotificationSystem;