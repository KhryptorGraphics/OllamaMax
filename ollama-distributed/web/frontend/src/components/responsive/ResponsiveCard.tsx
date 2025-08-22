import React, { useState, useRef, useEffect } from 'react';
import { ChevronRight, MoreVertical, ExternalLink } from 'lucide-react';

interface ResponsiveCardProps {
  title: string;
  subtitle?: string;
  children?: React.ReactNode;
  actions?: React.ReactNode;
  onClick?: () => void;
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  className?: string;
  variant?: 'default' | 'elevated' | 'outlined' | 'filled';
  size?: 'small' | 'medium' | 'large';
  isLoading?: boolean;
  badge?: string | number;
  image?: string;
  isClickable?: boolean;
}

export const ResponsiveCard: React.FC<ResponsiveCardProps> = ({
  title,
  subtitle,
  children,
  actions,
  onClick,
  onSwipeLeft,
  onSwipeRight,
  className = '',
  variant = 'default',
  size = 'medium',
  isLoading = false,
  badge,
  image,
  isClickable = false
}) => {
  const [isPressed, setIsPressed] = useState(false);
  const [swipeOffset, setSwipeOffset] = useState(0);
  const [showActions, setShowActions] = useState(false);
  
  const cardRef = useRef<HTMLDivElement>(null);
  const touchStartRef = useRef({ x: 0, y: 0, time: 0 });
  const touchCurrentRef = useRef({ x: 0, y: 0 });
  const isDraggingRef = useRef(false);

  // Handle touch events for swipe gestures
  const handleTouchStart = (e: React.TouchEvent) => {
    if (e.touches.length !== 1) return;
    
    const touch = e.touches[0];
    touchStartRef.current = {
      x: touch.clientX,
      y: touch.clientY,
      time: Date.now()
    };
    
    touchCurrentRef.current = {
      x: touch.clientX,
      y: touch.clientY
    };
    
    setIsPressed(true);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (e.touches.length !== 1) return;
    
    const touch = e.touches[0];
    touchCurrentRef.current = {
      x: touch.clientX,
      y: touch.clientY
    };
    
    const deltaX = touch.clientX - touchStartRef.current.x;
    const deltaY = touch.clientY - touchStartRef.current.y;
    
    // Start dragging if horizontal movement is greater than vertical
    if (!isDraggingRef.current && Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 10) {
      isDraggingRef.current = true;
      e.preventDefault();
    }
    
    if (isDraggingRef.current) {
      setSwipeOffset(deltaX);
    }
  };

  const handleTouchEnd = () => {
    const deltaX = touchCurrentRef.current.x - touchStartRef.current.x;
    const deltaY = touchCurrentRef.current.y - touchStartRef.current.y;
    const deltaTime = Date.now() - touchStartRef.current.time;
    
    setIsPressed(false);
    setSwipeOffset(0);
    
    if (isDraggingRef.current) {
      // Handle swipe gestures
      if (Math.abs(deltaX) > 80) {
        if (deltaX > 0 && onSwipeRight) {
          onSwipeRight();
        } else if (deltaX < 0 && onSwipeLeft) {
          onSwipeLeft();
        }
      }
    } else {
      // Handle tap/click
      if (Math.abs(deltaX) < 10 && Math.abs(deltaY) < 10 && deltaTime < 300) {
        if (onClick) {
          onClick();
        }
      }
    }
    
    isDraggingRef.current = false;
  };

  // Keyboard navigation
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick?.();
    }
  };

  // Card variant styles
  const variantStyles = {
    default: 'bg-white border border-gray-200 shadow-sm',
    elevated: 'bg-white shadow-md border border-gray-100',
    outlined: 'bg-white border-2 border-gray-300',
    filled: 'bg-gray-50 border border-gray-200'
  };

  // Card size styles
  const sizeStyles = {
    small: 'p-3',
    medium: 'p-4 sm:p-5',
    large: 'p-5 sm:p-6'
  };

  const baseClasses = `
    relative rounded-lg transition-all duration-200 overflow-hidden
    ${variantStyles[variant]}
    ${sizeStyles[size]}
    ${isClickable ? 'cursor-pointer hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2' : ''}
    ${isPressed ? 'scale-98 shadow-sm' : ''}
    ${className}
  `;

  return (
    <div
      ref={cardRef}
      className={baseClasses}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onKeyDown={handleKeyDown}
      tabIndex={isClickable ? 0 : -1}
      role={isClickable ? 'button' : undefined}
      style={{
        transform: `translateX(${swipeOffset}px)`,
        transition: isDraggingRef.current ? 'none' : 'transform 0.2s ease-out'
      }}
    >
      {/* Loading Overlay */}
      {isLoading && (
        <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-10">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            <span className="text-sm text-gray-600">Loading...</span>
          </div>
        </div>
      )}

      {/* Card Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-start space-x-3 flex-1 min-w-0">
          {/* Image */}
          {image && (
            <div className="flex-shrink-0">
              <img
                src={image}
                alt=""
                className="w-12 h-12 sm:w-16 sm:h-16 rounded-lg object-cover"
                loading="lazy"
              />
            </div>
          )}
          
          {/* Title and Subtitle */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2">
              <h3 className="text-lg font-semibold text-gray-900 truncate">
                {title}
              </h3>
              {badge && (
                <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {badge}
                </span>
              )}
            </div>
            {subtitle && (
              <p className="text-sm text-gray-600 mt-1 line-clamp-2">
                {subtitle}
              </p>
            )}
          </div>
        </div>

        {/* Actions Menu */}
        <div className="flex items-center space-x-2 ml-4">
          {actions && (
            <div className="relative">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setShowActions(!showActions);
                }}
                className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                aria-label="More actions"
              >
                <MoreVertical className="w-4 h-4" />
              </button>
              
              {showActions && (
                <div className="absolute right-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-20 min-w-48">
                  {actions}
                </div>
              )}
            </div>
          )}
          
          {isClickable && (
            <ChevronRight className="w-5 h-5 text-gray-400" />
          )}
        </div>
      </div>

      {/* Card Content */}
      {children && (
        <div className="mb-4">
          {children}
        </div>
      )}

      {/* Swipe Indicators */}
      {(onSwipeLeft || onSwipeRight) && (
        <div className="absolute inset-y-0 left-0 right-0 pointer-events-none">
          {/* Left swipe indicator */}
          {onSwipeLeft && swipeOffset < -50 && (
            <div className="absolute right-4 top-1/2 transform -translate-y-1/2 bg-red-500 text-white p-2 rounded-full">
              <ExternalLink className="w-4 h-4" />
            </div>
          )}
          
          {/* Right swipe indicator */}
          {onSwipeRight && swipeOffset > 50 && (
            <div className="absolute left-4 top-1/2 transform -translate-y-1/2 bg-green-500 text-white p-2 rounded-full">
              <ExternalLink className="w-4 h-4" />
            </div>
          )}
        </div>
      )}

      {/* Focus Ring for Accessibility */}
      <div className="absolute inset-0 rounded-lg ring-2 ring-blue-500 ring-opacity-0 transition-opacity focus-within:ring-opacity-100 pointer-events-none"></div>
    </div>
  );
};

// Action menu item component
interface CardActionProps {
  icon?: React.ComponentType<any>;
  label: string;
  onClick: () => void;
  variant?: 'default' | 'danger';
  disabled?: boolean;
}

export const CardAction: React.FC<CardActionProps> = ({
  icon: Icon,
  label,
  onClick,
  variant = 'default',
  disabled = false
}) => {
  const variantClasses = {
    default: 'text-gray-700 hover:bg-gray-50',
    danger: 'text-red-600 hover:bg-red-50'
  };

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`
        w-full flex items-center px-4 py-3 text-sm transition-colors text-left
        ${variantClasses[variant]}
        ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:text-gray-900'}
        first:rounded-t-lg last:rounded-b-lg
      `}
    >
      {Icon && <Icon className="w-4 h-4 mr-3" />}
      {label}
    </button>
  );
};