import React, { useState } from 'react';
import { RefreshCw, X, Download, AlertCircle } from 'lucide-react';
import { usePWA } from '../../hooks/usePWA';

interface PWAUpdateNotificationProps {
  onUpdate?: () => void;
  onDismiss?: () => void;
  className?: string;
}

export const PWAUpdateNotification: React.FC<PWAUpdateNotificationProps> = ({
  onUpdate,
  onDismiss,
  className = ''
}) => {
  const { updateAvailable, updateApp } = usePWA();
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  // Handle update
  const handleUpdate = async () => {
    setIsUpdating(true);
    
    try {
      await updateApp();
      onUpdate?.();
    } catch (error) {
      console.error('Update failed:', error);
    } finally {
      setIsUpdating(false);
    }
  };

  // Handle dismiss
  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  if (!updateAvailable || isDismissed) {
    return null;
  }

  return (
    <>
      {/* Mobile Notification */}
      <div className="md:hidden fixed top-0 left-0 right-0 z-50 bg-green-600 text-white shadow-lg">
        <div className="px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-white bg-opacity-20 rounded-full flex items-center justify-center">
                <RefreshCw className="w-4 h-4" />
              </div>
              <div>
                <p className="font-medium text-sm">Update Available</p>
                <p className="text-xs text-green-100">
                  Tap to install the latest version
                </p>
              </div>
            </div>
            
            <div className="flex items-center space-x-2">
              <button
                onClick={handleUpdate}
                disabled={isUpdating}
                className="bg-white text-green-600 hover:bg-green-50 disabled:bg-gray-200 px-3 py-1.5 rounded text-xs font-medium transition-colors"
              >
                {isUpdating ? 'Updating...' : 'Update'}
              </button>
              
              <button
                onClick={handleDismiss}
                className="p-1.5 text-white hover:bg-white hover:bg-opacity-20 rounded transition-colors"
                aria-label="Dismiss update notification"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Desktop Notification */}
      <div className={`hidden md:block fixed top-4 right-4 z-50 bg-white border border-green-200 rounded-lg shadow-lg max-w-sm ${className}`}>
        <div className="p-4">
          <div className="flex items-start space-x-3">
            <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0">
              <RefreshCw className="w-5 h-5 text-green-600" />
            </div>
            
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-semibold text-gray-900">
                Update Available
              </h3>
              <p className="text-sm text-gray-600 mt-1">
                A new version of OllamaMax is ready to install. 
                Update now for the latest features and improvements.
              </p>
              
              <div className="flex items-center space-x-3 mt-4">
                <button
                  onClick={handleUpdate}
                  disabled={isUpdating}
                  className="bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white px-4 py-2 rounded-lg text-sm font-medium flex items-center space-x-2 transition-colors"
                >
                  <Download className="w-4 h-4" />
                  <span>{isUpdating ? 'Updating...' : 'Update Now'}</span>
                </button>
                
                <button
                  onClick={handleDismiss}
                  className="text-gray-500 hover:text-gray-700 text-sm font-medium transition-colors"
                >
                  Later
                </button>
              </div>
            </div>
            
            <button
              onClick={handleDismiss}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              aria-label="Dismiss update notification"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
      </div>

      {/* Critical Update Modal (for major security updates) */}
      {updateAvailable && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
            <div className="p-6">
              <div className="flex items-center space-x-3">
                <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
                  <AlertCircle className="w-6 h-6 text-orange-600" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">
                    Critical Update Required
                  </h3>
                  <p className="text-sm text-gray-600">
                    This update includes important security fixes
                  </p>
                </div>
              </div>
              
              <div className="mt-4">
                <p className="text-sm text-gray-700">
                  A critical security update is available. We recommend updating 
                  immediately to ensure your data remains secure.
                </p>
                
                <div className="mt-6 flex items-center space-x-3">
                  <button
                    onClick={handleUpdate}
                    disabled={isUpdating}
                    className="flex-1 bg-orange-600 hover:bg-orange-700 disabled:bg-orange-400 text-white px-4 py-3 rounded-lg font-medium flex items-center justify-center space-x-2 transition-colors"
                  >
                    <Download className="w-4 h-4" />
                    <span>{isUpdating ? 'Updating...' : 'Update Now'}</span>
                  </button>
                  
                  <button
                    onClick={handleDismiss}
                    className="px-4 py-3 border border-gray-300 hover:bg-gray-50 rounded-lg font-medium text-gray-700 transition-colors"
                  >
                    Remind Later
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Update Progress Overlay */}
      {isUpdating && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
          <div className="bg-white rounded-lg shadow-xl p-6 max-w-sm w-full">
            <div className="text-center">
              <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                <RefreshCw className="w-8 h-8 text-blue-600 animate-spin" />
              </div>
              
              <h3 className="text-lg font-semibold text-gray-900 mt-4">
                Updating OllamaMax
              </h3>
              
              <p className="text-sm text-gray-600 mt-2">
                Please wait while we install the latest version. 
                The app will restart automatically.
              </p>
              
              <div className="mt-4">
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div className="bg-blue-600 h-2 rounded-full animate-pulse" style={{ width: '70%' }}></div>
                </div>
                <p className="text-xs text-gray-500 mt-2">
                  Installing updates...
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};