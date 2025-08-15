/**
 * OllamaMax Mobile App
 * 
 * React Native application for iOS and Android platforms.
 */

import React, { useEffect, useState } from 'react';
import {
  StatusBar,
  useColorScheme,
  Alert,
  AppState,
  AppStateStatus,
} from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import SplashScreen from 'react-native-splash-screen';
import NetInfo from '@react-native-community/netinfo';
import DeviceInfo from 'react-native-device-info';
import { enableScreens } from 'react-native-screens';
import 'react-native-gesture-handler';

// Enable screens for better performance
enableScreens();

// Providers and Navigation
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider } from './contexts/AuthContext';
import { NetworkProvider } from './contexts/NetworkContext';
import { NotificationProvider } from './contexts/NotificationContext';
import AppNavigator from './navigation/AppNavigator';

// Services
import { authService } from './services/AuthService';
import { notificationService } from './services/NotificationService';
import { analyticsService } from './services/AnalyticsService';
import { crashReportingService } from './services/CrashReportingService';

// Utils
import { setupGlobalErrorHandler } from './utils/errorHandler';
import { Colors } from './theme/colors';

const App: React.FC = () => {
  const isDarkMode = useColorScheme() === 'dark';
  const [isAppReady, setIsAppReady] = useState(false);
  const [appState, setAppState] = useState<AppStateStatus>(AppState.currentState);

  // Initialize app
  useEffect(() => {
    initializeApp();
  }, []);

  // Handle app state changes
  useEffect(() => {
    const subscription = AppState.addEventListener('change', handleAppStateChange);
    return () => subscription?.remove();
  }, []);

  const initializeApp = async () => {
    try {
      // Setup global error handling
      setupGlobalErrorHandler();

      // Initialize crash reporting
      await crashReportingService.initialize();

      // Initialize analytics
      await analyticsService.initialize();

      // Initialize authentication
      await authService.initialize();

      // Initialize notifications
      await notificationService.initialize();

      // Setup network monitoring
      setupNetworkMonitoring();

      // Log app launch
      await analyticsService.logEvent('app_launch', {
        platform: DeviceInfo.getSystemName(),
        version: DeviceInfo.getVersion(),
        build: DeviceInfo.getBuildNumber(),
      });

      setIsAppReady(true);
      
      // Hide splash screen
      SplashScreen.hide();
    } catch (error) {
      console.error('App initialization failed:', error);
      crashReportingService.recordError(error as Error);
      
      // Show error alert
      Alert.alert(
        'Initialization Error',
        'Failed to initialize the app. Please restart the application.',
        [
          {
            text: 'Retry',
            onPress: initializeApp,
          },
        ]
      );
    }
  };

  const setupNetworkMonitoring = () => {
    NetInfo.addEventListener(state => {
      console.log('Network state changed:', state);
      
      if (!state.isConnected) {
        // Handle offline state
        notificationService.showLocalNotification({
          title: 'Connection Lost',
          body: 'You are now offline. Some features may be limited.',
        });
      } else if (state.isConnected && !state.isInternetReachable) {
        // Handle limited connectivity
        notificationService.showLocalNotification({
          title: 'Limited Connectivity',
          body: 'Internet connection is limited.',
        });
      }
    });
  };

  const handleAppStateChange = (nextAppState: AppStateStatus) => {
    if (appState.match(/inactive|background/) && nextAppState === 'active') {
      // App has come to the foreground
      console.log('App has come to the foreground');
      
      // Refresh authentication if needed
      authService.refreshTokenIfNeeded();
      
      // Log app foreground
      analyticsService.logEvent('app_foreground');
    } else if (nextAppState.match(/inactive|background/)) {
      // App has gone to the background
      console.log('App has gone to the background');
      
      // Log app background
      analyticsService.logEvent('app_background');
    }

    setAppState(nextAppState);
  };

  if (!isAppReady) {
    // Return null while splash screen is showing
    return null;
  }

  return (
    <SafeAreaProvider>
      <ThemeProvider>
        <NetworkProvider>
          <AuthProvider>
            <NotificationProvider>
              <NavigationContainer
                theme={{
                  dark: isDarkMode,
                  colors: {
                    primary: Colors.primary,
                    background: isDarkMode ? Colors.dark.background : Colors.light.background,
                    card: isDarkMode ? Colors.dark.surface : Colors.light.surface,
                    text: isDarkMode ? Colors.dark.text : Colors.light.text,
                    border: isDarkMode ? Colors.dark.border : Colors.light.border,
                    notification: Colors.primary,
                  },
                }}
                onReady={() => {
                  console.log('Navigation ready');
                  analyticsService.logEvent('navigation_ready');
                }}
                onStateChange={(state) => {
                  // Log navigation state changes for analytics
                  const currentRoute = state?.routes[state.index];
                  if (currentRoute) {
                    analyticsService.logScreenView(currentRoute.name);
                  }
                }}
              >
                <StatusBar
                  barStyle={isDarkMode ? 'light-content' : 'dark-content'}
                  backgroundColor={isDarkMode ? Colors.dark.background : Colors.light.background}
                  translucent={false}
                />
                <AppNavigator />
              </NavigationContainer>
            </NotificationProvider>
          </AuthProvider>
        </NetworkProvider>
      </ThemeProvider>
    </SafeAreaProvider>
  );
};

export default App;
