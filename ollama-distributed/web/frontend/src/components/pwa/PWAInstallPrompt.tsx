import React, { useState, useEffect } from 'react';
import { Download, X, Smartphone, Monitor, Share, Star } from 'lucide-react';
import { usePWA } from '../../hooks/usePWA';

interface PWAInstallPromptProps {
  onInstall?: () => void;
  onDismiss?: () => void;
  className?: string;
}

export const PWAInstallPrompt: React.FC<PWAInstallPromptProps> = ({
  onInstall,
  onDismiss,
  className = ''
}) => {
  const { 
    canInstall, 
    isInstallable, 
    isStandalone,
    installApp, 
    dismissInstallPrompt 
  } = usePWA();
  
  const [isVisible, setIsVisible] = useState(false);
  const [isInstalling, setIsInstalling] = useState(false);
  const [showFeatures, setShowFeatures] = useState(false);

  // Show prompt conditions
  useEffect(() => {
    const shouldShow = canInstall && isInstallable && !isStandalone;
    setIsVisible(shouldShow);
  }, [canInstall, isInstallable, isStandalone]);

  // Handle install
  const handleInstall = async () => {
    setIsInstalling(true);
    
    try {
      const success = await installApp();
      if (success) {
        onInstall?.();
        setIsVisible(false);
      }
    } catch (error) {
      console.error('Install failed:', error);
    } finally {
      setIsInstalling(false);
    }
  };

  // Handle dismiss
  const handleDismiss = () => {
    dismissInstallPrompt();
    onDismiss?.();
    setIsVisible(false);
  };

  // Features of the PWA
  const features = [
    {
      icon: Smartphone,
      title: 'Works Offline',
      description: 'Access your dashboard even without internet'
    },
    {
      icon: Monitor,
      title: 'Native Experience',
      description: 'Runs like a native app on your device'
    },
    {
      icon: Share,
      title: 'Quick Access',
      description: 'Add to home screen for instant access'
    },
    {
      icon: Star,
      title: 'Always Updated',
      description: 'Automatically stays up to date'
    }
  ];

  if (!isVisible) {
    return null;
  }

  return (
    <>
      {/* Mobile Bottom Sheet */}
      <div className="md:hidden fixed inset-x-0 bottom-0 z-50 bg-white border-t-2 border-gray-200 shadow-lg transform transition-transform duration-300">
        <div className="px-4 py-4">
          <div className="flex items-start space-x-3">
            <div className="w-12 h-12 bg-blue-600 rounded-xl flex items-center justify-center flex-shrink-0">
              <span className="text-white text-lg font-bold">O</span>
            </div>
            
            <div className="flex-1 min-w-0">
              <h3 className="text-lg font-semibold text-gray-900">
                Install OllamaMax
              </h3>
              <p className="text-sm text-gray-600 mt-1">
                Get the full app experience with offline access and native performance.
              </p>
              
              {showFeatures && (
                <div className="mt-3 grid grid-cols-2 gap-3">
                  {features.map((feature, index) => {
                    const Icon = feature.icon;
                    return (
                      <div key={index} className="flex items-center space-x-2">
                        <Icon className="w-4 h-4 text-blue-600 flex-shrink-0" />
                        <div>
                          <p className="text-xs font-medium text-gray-900">{feature.title}</p>
                          <p className="text-xs text-gray-600">{feature.description}</p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
            
            <button
              onClick={handleDismiss}
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
              aria-label="Dismiss install prompt"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
          
          <div className="flex items-center space-x-3 mt-4">
            <button
              onClick={handleInstall}
              disabled={isInstalling}
              className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white px-4 py-3 rounded-lg font-medium flex items-center justify-center space-x-2 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>{isInstalling ? 'Installing...' : 'Install App'}</span>
            </button>
            
            <button
              onClick={() => setShowFeatures(!showFeatures)}
              className="px-4 py-3 text-blue-600 hover:text-blue-700 font-medium transition-colors"
            >
              {showFeatures ? 'Less' : 'More'}
            </button>
          </div>
        </div>
      </div>

      {/* Desktop Banner */}
      <div className={`hidden md:block fixed top-0 left-0 right-0 z-50 bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg ${className}`}>
        <div className="max-w-7xl mx-auto px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="w-10 h-10 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
                <Download className="w-5 h-5" />
              </div>
              
              <div>
                <h3 className="font-semibold">Install OllamaMax App</h3>
                <p className="text-sm text-blue-100">
                  Get native performance, offline access, and push notifications
                </p>
              </div>
            </div>
            
            <div className="flex items-center space-x-3">
              <button
                onClick={() => setShowFeatures(!showFeatures)}
                className="px-4 py-2 bg-white bg-opacity-20 hover:bg-opacity-30 rounded-lg text-sm font-medium transition-colors"
              >
                {showFeatures ? 'Hide Features' : 'View Features'}
              </button>
              
              <button
                onClick={handleInstall}
                disabled={isInstalling}
                className="bg-white text-blue-600 hover:bg-gray-100 disabled:bg-gray-200 px-4 py-2 rounded-lg font-medium flex items-center space-x-2 transition-colors"
              >
                <Download className="w-4 h-4" />
                <span>{isInstalling ? 'Installing...' : 'Install'}</span>
              </button>
              
              <button
                onClick={handleDismiss}
                className="p-2 text-white hover:bg-white hover:bg-opacity-20 rounded-lg transition-colors"
                aria-label="Dismiss install prompt"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>
          
          {/* Features Grid */}
          {showFeatures && (
            <div className="mt-4 pt-4 border-t border-white border-opacity-20">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {features.map((feature, index) => {
                  const Icon = feature.icon;
                  return (
                    <div key={index} className="flex items-start space-x-3">
                      <div className="w-8 h-8 bg-white bg-opacity-20 rounded-lg flex items-center justify-center flex-shrink-0">
                        <Icon className="w-4 h-4" />
                      </div>
                      <div>
                        <p className="font-medium text-sm">{feature.title}</p>
                        <p className="text-xs text-blue-100">{feature.description}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Floating Action Button for tablets */}
      <div className="hidden sm:block md:hidden fixed bottom-6 right-6 z-50">
        <button
          onClick={handleInstall}
          disabled={isInstalling}
          className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white w-14 h-14 rounded-full shadow-lg flex items-center justify-center transition-all duration-200 hover:scale-110"
          aria-label="Install OllamaMax app"
        >
          <Download className="w-6 h-6" />
        </button>
      </div>
    </>
  );
};