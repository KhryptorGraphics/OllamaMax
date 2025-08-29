import React, { useState, useEffect } from 'react';
import { Card, Button, Form, Modal, Alert, Badge, Offcanvas, Accordion } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faMobile,
  faBars,
  faTimes,
  faHome,
  faChartLine,
  faServer,
  faDatabase,
  faCog,
  faUser,
  faSearch,
  faFilter,
  faSort,
  faExpand,
  faCompress,
  faRefresh,
  faDownload,
  faUpload,
  faShare,
  faBell,
  faEnvelope,
  faHeart,
  faStar,
  faBookmark,
  faEye,
  faEyeSlash,
  faVolumeUp,
  faVolumeOff,
  faSun,
  faMoon,
  faWifi,
  faBattery3,
  faSignal,
  faLocationArrow,
  faCamera,
  faMicrophone,
  faKeyboard,
  faHandPaper
} from '@fortawesome/free-solid-svg-icons';
import LoadingSpinner from './LoadingSpinner';

const MobileInterface = ({
  activeView = 'dashboard',
  onViewChange,
  data = {},
  notifications = [],
  onNotificationDismiss,
  className = ""
}) => {
  const [showSidebar, setShowSidebar] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [orientation, setOrientation] = useState('portrait');
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [touchPosition, setTouchPosition] = useState({ x: 0, y: 0 });
  const [swipeDirection, setSwipeDirection] = useState(null);
  const [vibrationEnabled, setVibrationEnabled] = useState(true);
  const [darkMode, setDarkMode] = useState(false);
  const [voiceEnabled, setVoiceEnabled] = useState(false);
  const [deviceInfo, setDeviceInfo] = useState({
    userAgent: '',
    screenSize: { width: 0, height: 0 },
    pixelRatio: 1,
    touchSupport: false,
    batteryLevel: 100,
    networkType: 'wifi',
    signalStrength: 4
  });
  const [gestureSettings, setGestureSettings] = useState({
    swipeNavigation: true,
    pinchZoom: true,
    doubleTapZoom: true,
    longPress: true,
    shakeToRefresh: true
  });
  const [mobileSettings, setMobileSettings] = useState({
    autoRotate: true,
    hapticFeedback: true,
    reducedData: false,
    offlineMode: false,
    compactMode: true,
    oneHandedMode: false
  });

  const navigationItems = [
    { id: 'dashboard', label: 'Dashboard', icon: faHome },
    { id: 'metrics', label: 'Metrics', icon: faChartLine },
    { id: 'nodes', label: 'Nodes', icon: faServer },
    { id: 'models', label: 'Models', icon: faDatabase },
    { id: 'settings', label: 'Settings', icon: faCog },
    { id: 'profile', label: 'Profile', icon: faUser }
  ];

  // Initialize mobile-specific features
  useEffect(() => {
    detectMobileFeatures();
    setupOrientationListener();
    setupTouchEvents();
    setupNetworkListener();
    setupBatteryListener();
    setupVisibilityListener();

    return () => {
      cleanupMobileFeatures();
    };
  }, []);

  // Handle orientation changes
  useEffect(() => {
    updateLayoutForOrientation(orientation);
  }, [orientation]);

  const detectMobileFeatures = () => {
    const info = {
      userAgent: navigator.userAgent,
      screenSize: {
        width: window.screen.width,
        height: window.screen.height
      },
      pixelRatio: window.devicePixelRatio || 1,
      touchSupport: 'ontouchstart' in window || navigator.maxTouchPoints > 0,
      batteryLevel: 100,
      networkType: getNetworkType(),
      signalStrength: getSignalStrength()
    };
    
    setDeviceInfo(info);
  };

  const setupOrientationListener = () => {
    const handleOrientationChange = () => {
      const orientation = window.screen.orientation?.type || 
        (window.innerHeight > window.innerWidth ? 'portrait' : 'landscape');
      setOrientation(orientation.includes('portrait') ? 'portrait' : 'landscape');
    };

    window.addEventListener('orientationchange', handleOrientationChange);
    window.addEventListener('resize', handleOrientationChange);
    
    // Initial check
    handleOrientationChange();

    return () => {
      window.removeEventListener('orientationchange', handleOrientationChange);
      window.removeEventListener('resize', handleOrientationChange);
    };
  };

  const setupTouchEvents = () => {
    if (!deviceInfo.touchSupport) return;

    let startPos = { x: 0, y: 0 };
    let startTime = 0;
    let lastTap = 0;

    const handleTouchStart = (e) => {
      const touch = e.touches[0];
      startPos = { x: touch.clientX, y: touch.clientY };
      startTime = Date.now();
      setTouchPosition(startPos);
    };

    const handleTouchMove = (e) => {
      if (!gestureSettings.swipeNavigation) return;
      
      const touch = e.touches[0];
      const currentPos = { x: touch.clientX, y: touch.clientY };
      const deltaX = currentPos.x - startPos.x;
      const deltaY = currentPos.y - startPos.y;
      
      // Determine swipe direction
      if (Math.abs(deltaX) > Math.abs(deltaY)) {
        setSwipeDirection(deltaX > 0 ? 'right' : 'left');
      } else {
        setSwipeDirection(deltaY > 0 ? 'down' : 'up');
      }
    };

    const handleTouchEnd = (e) => {
      const endTime = Date.now();
      const touch = e.changedTouches[0];
      const endPos = { x: touch.clientX, y: touch.clientY };
      const deltaTime = endTime - startTime;
      const deltaX = endPos.x - startPos.x;
      const deltaY = endPos.y - startPos.y;
      const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

      // Handle different gestures
      if (deltaTime < 300 && distance < 10) {
        // Tap
        if (endTime - lastTap < 300) {
          // Double tap
          if (gestureSettings.doubleTapZoom) {
            handleDoubleTap(endPos);
          }
        }
        lastTap = endTime;
        
        if (vibrationEnabled && navigator.vibrate) {
          navigator.vibrate(10); // Short haptic feedback
        }
      } else if (deltaTime > 500 && distance < 10) {
        // Long press
        if (gestureSettings.longPress) {
          handleLongPress(endPos);
        }
      } else if (distance > 50) {
        // Swipe
        handleSwipe(swipeDirection, distance, deltaTime);
      }

      setSwipeDirection(null);
    };

    document.addEventListener('touchstart', handleTouchStart, { passive: true });
    document.addEventListener('touchmove', handleTouchMove, { passive: false });
    document.addEventListener('touchend', handleTouchEnd, { passive: true });

    return () => {
      document.removeEventListener('touchstart', handleTouchStart);
      document.removeEventListener('touchmove', handleTouchMove);
      document.removeEventListener('touchend', handleTouchEnd);
    };
  };

  const setupNetworkListener = () => {
    const updateNetworkStatus = () => {
      const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
      if (connection) {
        setDeviceInfo(prev => ({
          ...prev,
          networkType: getNetworkType(connection),
          signalStrength: getSignalStrength(connection)
        }));
      }
    };

    if (navigator.connection) {
      navigator.connection.addEventListener('change', updateNetworkStatus);
    }

    return () => {
      if (navigator.connection) {
        navigator.connection.removeEventListener('change', updateNetworkStatus);
      }
    };
  };

  const setupBatteryListener = () => {
    if (navigator.getBattery) {
      navigator.getBattery().then(battery => {
        const updateBattery = () => {
          setDeviceInfo(prev => ({
            ...prev,
            batteryLevel: Math.round(battery.level * 100)
          }));
        };

        battery.addEventListener('levelchange', updateBattery);
        updateBattery();
      });
    }
  };

  const setupVisibilityListener = () => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // App went to background - reduce activity
        console.log('App backgrounded');
      } else {
        // App came to foreground - resume activity
        console.log('App foregrounded');
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  };

  const cleanupMobileFeatures = () => {
    // Cleanup would be handled by the return functions in useEffect
  };

  const getNetworkType = (connection) => {
    if (!connection) return 'unknown';
    
    const type = connection.effectiveType || connection.type;
    if (type.includes('wifi')) return 'wifi';
    if (type.includes('4g') || type.includes('3g')) return 'cellular';
    return 'unknown';
  };

  const getSignalStrength = (connection) => {
    if (!connection) return 4;
    
    const downlink = connection.downlink || 10;
    if (downlink > 10) return 4;
    if (downlink > 5) return 3;
    if (downlink > 1) return 2;
    return 1;
  };

  const updateLayoutForOrientation = (orientation) => {
    const body = document.body;
    body.classList.remove('portrait', 'landscape');
    body.classList.add(orientation);

    // Adjust layout based on orientation
    if (orientation === 'landscape') {
      // Landscape-specific adjustments
      setMobileSettings(prev => ({ ...prev, compactMode: true }));
    } else {
      // Portrait-specific adjustments
      setMobileSettings(prev => ({ ...prev, compactMode: false }));
    }
  };

  const handleSwipe = (direction, distance, duration) => {
    if (!gestureSettings.swipeNavigation) return;

    const velocity = distance / duration;
    if (velocity < 0.5) return; // Swipe too slow

    switch (direction) {
      case 'left':
        // Navigate to next view
        navigateToNextView();
        break;
      case 'right':
        // Navigate to previous view or open sidebar
        if (distance > 100) {
          setShowSidebar(true);
        } else {
          navigateToPreviousView();
        }
        break;
      case 'up':
        // Refresh or scroll up
        if (gestureSettings.shakeToRefresh) {
          refreshData();
        }
        break;
      case 'down':
        // Show notifications or scroll down
        setShowNotifications(true);
        break;
    }

    // Provide haptic feedback
    if (vibrationEnabled && navigator.vibrate) {
      navigator.vibrate([20, 10, 20]); // Pattern for swipe
    }
  };

  const handleDoubleTap = (position) => {
    // Toggle fullscreen or zoom
    toggleFullscreen();
  };

  const handleLongPress = (position) => {
    // Show context menu or additional options
    if (vibrationEnabled && navigator.vibrate) {
      navigator.vibrate(50); // Longer vibration for long press
    }
  };

  const navigateToNextView = () => {
    const currentIndex = navigationItems.findIndex(item => item.id === activeView);
    const nextIndex = (currentIndex + 1) % navigationItems.length;
    const nextView = navigationItems[nextIndex].id;
    
    if (onViewChange) {
      onViewChange(nextView);
    }
  };

  const navigateToPreviousView = () => {
    const currentIndex = navigationItems.findIndex(item => item.id === activeView);
    const prevIndex = currentIndex > 0 ? currentIndex - 1 : navigationItems.length - 1;
    const prevView = navigationItems[prevIndex].id;
    
    if (onViewChange) {
      onViewChange(prevView);
    }
  };

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => {
        setIsFullscreen(true);
      });
    } else {
      document.exitFullscreen().then(() => {
        setIsFullscreen(false);
      });
    }
  };

  const refreshData = () => {
    // Trigger data refresh
    window.location.reload();
  };

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
    document.documentElement.setAttribute('data-theme', !darkMode ? 'dark' : 'light');
  };

  const StatusBar = () => (
    <div className="mobile-status-bar d-flex justify-content-between align-items-center px-3 py-1 bg-dark text-white" style={{ fontSize: '0.8rem' }}>
      <div className="d-flex align-items-center gap-2">
        <span>{new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
      </div>
      <div className="d-flex align-items-center gap-2">
        <FontAwesome icon={faSignal} />
        <span className="signal-bars">
          {Array.from({ length: deviceInfo.signalStrength }, (_, i) => (
            <span key={i} className="signal-bar bg-white"></span>
          ))}
        </span>
        <FontAwesome icon={deviceInfo.networkType === 'wifi' ? faWifi : faSignal} />
        <span>{deviceInfo.batteryLevel}%</span>
        <FontAwesome icon={faBattery3} />
      </div>
    </div>
  );

  const MobileNavigation = () => (
    <div className="mobile-navigation fixed-bottom bg-white border-top">
      <div className="d-flex">
        {navigationItems.slice(0, 5).map(item => (
          <Button
            key={item.id}
            variant="link"
            className={`flex-fill text-center py-3 ${activeView === item.id ? 'active' : ''}`}
            onClick={() => {
              if (onViewChange) onViewChange(item.id);
              if (vibrationEnabled && navigator.vibrate) {
                navigator.vibrate(10);
              }
            }}
          >
            <div>
              <FontAwesome icon={item.icon} size="lg" />
              <div style={{ fontSize: '0.7rem' }}>{item.label}</div>
            </div>
          </Button>
        ))}
      </div>
    </div>
  );

  const MobileSidebar = () => (
    <Offcanvas 
      show={showSidebar} 
      onHide={() => setShowSidebar(false)} 
      placement="start"
      className="mobile-sidebar"
    >
      <Offcanvas.Header closeButton>
        <Offcanvas.Title>
          <FontAwesome icon={faMobile} className="me-2" />
          OllamaMax Mobile
        </Offcanvas.Title>
      </Offcanvas.Header>
      <Offcanvas.Body>
        <div className="mobile-menu">
          {navigationItems.map(item => (
            <Button
              key={item.id}
              variant={activeView === item.id ? 'primary' : 'outline-primary'}
              className="w-100 mb-2 text-start"
              onClick={() => {
                if (onViewChange) onViewChange(item.id);
                setShowSidebar(false);
                if (vibrationEnabled && navigator.vibrate) {
                  navigator.vibrate(10);
                }
              }}
            >
              <FontAwesome icon={item.icon} className="me-2" />
              {item.label}
            </Button>
          ))}
        </div>
        
        <hr />
        
        <div className="mobile-settings">
          <h6>Quick Settings</h6>
          
          <div className="d-flex justify-content-between align-items-center mb-3">
            <span>Dark Mode</span>
            <Form.Check
              type="switch"
              checked={darkMode}
              onChange={toggleDarkMode}
            />
          </div>
          
          <div className="d-flex justify-content-between align-items-center mb-3">
            <span>Vibration</span>
            <Form.Check
              type="switch"
              checked={vibrationEnabled}
              onChange={(e) => setVibrationEnabled(e.target.checked)}
            />
          </div>
          
          <div className="d-flex justify-content-between align-items-center mb-3">
            <span>Voice Commands</span>
            <Form.Check
              type="switch"
              checked={voiceEnabled}
              onChange={(e) => setVoiceEnabled(e.target.checked)}
            />
          </div>
          
          <div className="d-flex justify-content-between align-items-center mb-3">
            <span>Compact Mode</span>
            <Form.Check
              type="switch"
              checked={mobileSettings.compactMode}
              onChange={(e) => setMobileSettings(prev => ({ ...prev, compactMode: e.target.checked }))}
            />
          </div>
        </div>
        
        <hr />
        
        <div className="device-info">
          <h6>Device Info</h6>
          <small className="text-muted">
            <div>Screen: {deviceInfo.screenSize.width}x{deviceInfo.screenSize.height}</div>
            <div>Orientation: {orientation}</div>
            <div>Touch: {deviceInfo.touchSupport ? 'Supported' : 'Not supported'}</div>
            <div>Network: {deviceInfo.networkType}</div>
            <div>Battery: {deviceInfo.batteryLevel}%</div>
          </small>
        </div>
      </Offcanvas.Body>
    </Offcanvas>
  );

  const NotificationDrawer = () => (
    <Offcanvas 
      show={showNotifications} 
      onHide={() => setShowNotifications(false)} 
      placement="top"
      className="notification-drawer"
    >
      <Offcanvas.Header closeButton>
        <Offcanvas.Title>
          <FontAwesome icon={faBell} className="me-2" />
          Notifications ({notifications.length})
        </Offcanvas.Title>
      </Offcanvas.Header>
      <Offcanvas.Body>
        {notifications.length === 0 ? (
          <div className="text-center py-4 text-muted">
            <FontAwesome icon={faBell} size="3x" className="mb-3" />
            <p>No notifications</p>
          </div>
        ) : (
          notifications.map(notification => (
            <Alert
              key={notification.id}
              variant={notification.type}
              dismissible
              onClose={() => onNotificationDismiss && onNotificationDismiss(notification.id)}
              className="mb-2"
            >
              <div className="d-flex justify-content-between align-items-start">
                <div>
                  <Alert.Heading as="h6">{notification.title}</Alert.Heading>
                  <p className="mb-1">{notification.message}</p>
                  <small className="text-muted">
                    {new Date(notification.timestamp).toLocaleString()}
                  </small>
                </div>
              </div>
            </Alert>
          ))
        )}
      </Offcanvas.Body>
    </Offcanvas>
  );

  const MobileHeader = () => (
    <div className="mobile-header d-flex justify-content-between align-items-center p-3 bg-white border-bottom">
      <Button
        variant="outline-primary"
        size="sm"
        onClick={() => setShowSidebar(true)}
      >
        <FontAwesome icon={faBars} />
      </Button>
      
      <h5 className="mb-0">
        {navigationItems.find(item => item.id === activeView)?.label || 'OllamaMax'}
      </h5>
      
      <div className="d-flex gap-2">
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={() => setShowNotifications(true)}
          className="position-relative"
        >
          <FontAwesome icon={faBell} />
          {notifications.length > 0 && (
            <Badge 
              bg="danger" 
              className="position-absolute top-0 start-100 translate-middle"
              style={{ fontSize: '0.6rem' }}
            >
              {notifications.length}
            </Badge>
          )}
        </Button>
        
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={toggleFullscreen}
        >
          <FontAwesome icon={isFullscreen ? faCompress : faExpand} />
        </Button>
      </div>
    </div>
  );

  return (
    <div className={`mobile-interface ${className} ${orientation} ${mobileSettings.compactMode ? 'compact' : ''}`}>
      {/* Status Bar (iOS/Android style) */}
      <StatusBar />
      
      {/* Mobile Header */}
      <MobileHeader />
      
      {/* Main Content Area */}
      <div 
        className="mobile-content"
        id="main-content"
        style={{ 
          paddingBottom: '80px', // Space for bottom navigation
          height: 'calc(100vh - 140px)', // Adjust for header and nav
          overflowY: 'auto',
          WebkitOverflowScrolling: 'touch' // Smooth scrolling on iOS
        }}
      >
        {/* Content will be rendered by parent component based on activeView */}
        <div className="p-3">
          {/* Placeholder for view-specific content */}
          <Card className="text-center">
            <Card.Body>
              <FontAwesome icon={faMobile} size="3x" className="text-primary mb-3" />
              <h5>Mobile Interface</h5>
              <p className="text-muted">Optimized for touch interaction</p>
              <div className="row text-center">
                <div className="col-4">
                  <div className="mb-2">
                    <FontAwesome icon={faHandPaper} size="2x" className="text-success" />
                  </div>
                  <small>Touch Gestures</small>
                </div>
                <div className="col-4">
                  <div className="mb-2">
                    <FontAwesome icon={faMicrophone} size="2x" className="text-info" />
                  </div>
                  <small>Voice Control</small>
                </div>
                <div className="col-4">
                  <div className="mb-2">
                    <FontAwesome icon={faKeyboard} size="2x" className="text-warning" />
                  </div>
                  <small>Haptic Feedback</small>
                </div>
              </div>
            </Card.Body>
          </Card>
        </div>
      </div>
      
      {/* Bottom Navigation */}
      <MobileNavigation />
      
      {/* Sidebar */}
      <MobileSidebar />
      
      {/* Notification Drawer */}
      <NotificationDrawer />
      
      {/* Gesture Indicator */}
      {swipeDirection && (
        <div 
          className="gesture-indicator position-fixed d-flex align-items-center justify-content-center"
          style={{
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: '60px',
            height: '60px',
            backgroundColor: 'rgba(0,0,0,0.7)',
            color: 'white',
            borderRadius: '50%',
            zIndex: 2000,
            pointerEvents: 'none'
          }}
        >
          <FontAwesome 
            icon={
              swipeDirection === 'left' ? faChevronLeft :
              swipeDirection === 'right' ? faChevronRight :
              swipeDirection === 'up' ? faChevronUp :
              faChevronDown
            } 
            size="lg" 
          />
        </div>
      )}
    </div>
  );
};

// Helper component for FontAwesome (since import might be different)
const FontAwesome = ({ icon, ...props }) => (
  <FontAwesome icon={icon} {...props} />
);

export default MobileInterface;