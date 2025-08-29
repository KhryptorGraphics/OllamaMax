import React, { Component } from 'react';
import { Card, Button, Alert, Accordion } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faExclamationTriangle,
  faRedo,
  faHome,
  faBug,
  faClipboard,
  faCode,
  faInfoCircle,
  faEnvelope
} from '@fortawesome/free-solid-svg-icons';

class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
      showDetails: false,
      reportSent: false
    };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI
    return {
      hasError: true,
      errorId: `ERR-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    };
  }

  componentDidCatch(error, errorInfo) {
    // Log the error to console and external services
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    this.setState({
      error,
      errorInfo
    });

    // Send error to monitoring service
    this.logErrorToService(error, errorInfo);
  }

  logErrorToService = async (error, errorInfo) => {
    try {
      // In a real application, you would send this to your error tracking service
      const errorData = {
        id: this.state.errorId,
        message: error?.message,
        stack: error?.stack,
        componentStack: errorInfo?.componentStack,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
        userId: this.props.userId || 'anonymous'
      };

      // Simulate API call to error tracking service
      console.log('Sending error to monitoring service:', errorData);
      
      // You could integrate with services like Sentry, Rollbar, etc.
      // await fetch('/api/errors', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(errorData)
      // });
      
    } catch (err) {
      console.error('Failed to log error to service:', err);
    }
  };

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
      showDetails: false,
      reportSent: false
    });
  };

  copyErrorToClipboard = async () => {
    const errorText = `
Error ID: ${this.state.errorId}
Error: ${this.state.error?.message || 'Unknown error'}
Stack: ${this.state.error?.stack || 'No stack trace'}
Component Stack: ${this.state.errorInfo?.componentStack || 'No component stack'}
Timestamp: ${new Date().toISOString()}
URL: ${window.location.href}
User Agent: ${navigator.userAgent}
    `.trim();

    try {
      await navigator.clipboard.writeText(errorText);
      // Show success feedback
      const button = document.querySelector('.copy-button');
      if (button) {
        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.classList.add('btn-success');
        setTimeout(() => {
          button.textContent = originalText;
          button.classList.remove('btn-success');
        }, 2000);
      }
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  sendErrorReport = async () => {
    try {
      // Simulate sending error report
      this.setState({ reportSent: true });
      
      // In a real application, you would send this via email or support system
      console.log('Error report sent for:', this.state.errorId);
      
    } catch (err) {
      console.error('Failed to send error report:', err);
    }
  };

  render() {
    if (this.state.hasError) {
      const { error, errorInfo, errorId, showDetails, reportSent } = this.state;

      return (
        <div className="error-boundary">
          <style jsx>{`
            .error-boundary {
              min-height: 100vh;
              background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
              display: flex;
              align-items: center;
              justify-content: center;
              padding: 2rem;
            }
            
            .error-container {
              max-width: 600px;
              width: 100%;
            }
            
            .error-icon {
              font-size: 4rem;
              color: var(--warning);
              margin-bottom: 1.5rem;
              text-align: center;
            }
            
            .error-title {
              font-size: 1.75rem;
              font-weight: var(--font-weight-bold);
              color: var(--text-primary);
              text-align: center;
              margin-bottom: 1rem;
            }
            
            .error-description {
              color: var(--text-secondary);
              text-align: center;
              margin-bottom: 2rem;
              line-height: 1.6;
            }
            
            .error-id {
              background: var(--bg-muted);
              border: 1px solid var(--border-primary);
              border-radius: var(--radius-md);
              padding: 0.75rem;
              font-family: var(--font-family-mono);
              font-size: 0.875rem;
              text-align: center;
              margin-bottom: 2rem;
            }
            
            .error-actions {
              display: flex;
              gap: 0.75rem;
              justify-content: center;
              flex-wrap: wrap;
              margin-bottom: 2rem;
            }
            
            .error-details {
              margin-top: 1.5rem;
            }
            
            .error-stack {
              background: var(--neutral-900);
              color: var(--neutral-100);
              padding: 1rem;
              border-radius: var(--radius-md);
              font-family: var(--font-family-mono);
              font-size: 0.8rem;
              overflow-x: auto;
              white-space: pre-wrap;
              word-break: break-all;
            }
            
            .report-section {
              background: var(--bg-subtle);
              border: 1px solid var(--border-primary);
              border-radius: var(--radius-md);
              padding: 1.5rem;
              margin-top: 1.5rem;
            }
            
            .report-success {
              color: var(--success);
              display: flex;
              align-items: center;
              gap: 0.5rem;
            }
            
            @media (max-width: 768px) {
              .error-boundary {
                padding: 1rem;
              }
              
              .error-actions {
                flex-direction: column;
                align-items: stretch;
              }
              
              .error-actions .btn {
                justify-content: center;
              }
            }
          `}</style>

          <div className="error-container">
            <Card className="shadow-lg border-0">
              <Card.Body className="p-4">
                <div className="error-icon">
                  <FontAwesomeIcon icon={faExclamationTriangle} />
                </div>
                
                <h1 className="error-title">Something went wrong</h1>
                
                <p className="error-description">
                  We're sorry, but an unexpected error occurred while loading this part of the application. 
                  Our team has been notified and is working to fix the issue.
                </p>

                <div className="error-id">
                  <strong>Error ID:</strong> {errorId}
                </div>

                <div className="error-actions">
                  <Button variant="primary" onClick={this.handleRetry}>
                    <FontAwesomeIcon icon={faRedo} className="me-2" />
                    Try Again
                  </Button>
                  
                  <Button variant="outline-secondary" onClick={this.handleReload}>
                    <FontAwesomeIcon icon={faRedo} className="me-2" />
                    Reload Page
                  </Button>
                  
                  <Button variant="outline-primary" onClick={this.handleGoHome}>
                    <FontAwesomeIcon icon={faHome} className="me-2" />
                    Go Home
                  </Button>
                </div>

                <Alert variant="info" className="mb-3">
                  <FontAwesomeIcon icon={faInfoCircle} className="me-2" />
                  <strong>What can you do?</strong>
                  <ul className="mb-0 mt-2">
                    <li>Try refreshing the page or clicking "Try Again"</li>
                    <li>Check your internet connection</li>
                    <li>Clear your browser cache and cookies</li>
                    <li>Contact support if the problem persists</li>
                  </ul>
                </Alert>

                <Accordion className="error-details">
                  <Accordion.Item eventKey="0">
                    <Accordion.Header>
                      <FontAwesomeIcon icon={faBug} className="me-2" />
                      Technical Details
                    </Accordion.Header>
                    <Accordion.Body>
                      {error && (
                        <div className="mb-3">
                          <h6>Error Message:</h6>
                          <div className="error-stack">
                            {error.message || 'No error message available'}
                          </div>
                        </div>
                      )}
                      
                      {error?.stack && (
                        <div className="mb-3">
                          <h6>Stack Trace:</h6>
                          <div className="error-stack">
                            {error.stack}
                          </div>
                        </div>
                      )}
                      
                      {errorInfo?.componentStack && (
                        <div className="mb-3">
                          <h6>Component Stack:</h6>
                          <div className="error-stack">
                            {errorInfo.componentStack}
                          </div>
                        </div>
                      )}
                      
                      <div className="mt-3">
                        <Button 
                          variant="outline-secondary" 
                          size="sm"
                          className="copy-button"
                          onClick={this.copyErrorToClipboard}
                        >
                          <FontAwesomeIcon icon={faClipboard} className="me-2" />
                          Copy Error Details
                        </Button>
                      </div>
                    </Accordion.Body>
                  </Accordion.Item>
                </Accordion>

                <div className="report-section">
                  <h6>
                    <FontAwesomeIcon icon={faEnvelope} className="me-2" />
                    Report this Issue
                  </h6>
                  
                  {reportSent ? (
                    <div className="report-success">
                      <FontAwesomeIcon icon={faInfoCircle} />
                      Thank you! Your error report has been sent to our development team.
                    </div>
                  ) : (
                    <div>
                      <p className="mb-3 text-muted small">
                        Help us improve by sending this error report to our development team. 
                        No personal information will be shared.
                      </p>
                      <Button 
                        variant="outline-primary" 
                        size="sm"
                        onClick={this.sendErrorReport}
                      >
                        <FontAwesomeIcon icon={faEnvelope} className="me-2" />
                        Send Error Report
                      </Button>
                    </div>
                  )}
                </div>
              </Card.Body>
            </Card>
          </div>
        </div>
      );
    }

    // If there's no error, render children normally
    return this.props.children;
  }
}

export default ErrorBoundary;