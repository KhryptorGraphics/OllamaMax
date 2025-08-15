/**
 * Authentication Service for React Native
 * 
 * Handles authentication, token management, and biometric authentication.
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Keychain from 'react-native-keychain';
import ReactNativeBiometrics from 'react-native-biometrics';
import DeviceInfo from 'react-native-device-info';

interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role?: string;
  permissions?: string[];
}

interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

interface LoginCredentials {
  email: string;
  password: string;
  rememberMe?: boolean;
}

interface RegisterData {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

class AuthService {
  private baseURL: string;
  private user: User | null = null;
  private tokens: AuthTokens | null = null;
  private biometrics: ReactNativeBiometrics;
  private refreshTimer: NodeJS.Timeout | null = null;

  constructor() {
    this.baseURL = __DEV__ 
      ? 'http://localhost:8080/api/v1/auth'
      : 'https://api.ollamamax.com/v1/auth';
    
    this.biometrics = new ReactNativeBiometrics({
      allowDeviceCredentials: true,
    });
  }

  // Initialize authentication service
  async initialize(): Promise<void> {
    try {
      // Load stored authentication data
      await this.loadStoredAuth();
      
      // Check biometric availability
      await this.checkBiometricAvailability();
      
      // Setup automatic token refresh
      if (this.tokens) {
        this.setupTokenRefresh();
      }
    } catch (error) {
      console.error('Auth service initialization failed:', error);
    }
  }

  // Load stored authentication data
  private async loadStoredAuth(): Promise<void> {
    try {
      // Load user data
      const userData = await AsyncStorage.getItem('user');
      if (userData) {
        this.user = JSON.parse(userData);
      }

      // Load tokens from secure storage
      const credentials = await Keychain.getInternetCredentials('ollama_tokens');
      if (credentials && credentials.password) {
        this.tokens = JSON.parse(credentials.password);
        
        // Check if tokens are still valid
        if (this.tokens && this.tokens.expiresAt < Date.now()) {
          await this.refreshAccessToken();
        }
      }
    } catch (error) {
      console.error('Failed to load stored auth:', error);
      await this.clearAuthData();
    }
  }

  // Check biometric availability
  private async checkBiometricAvailability(): Promise<void> {
    try {
      const { available, biometryType } = await this.biometrics.isSensorAvailable();
      console.log('Biometric availability:', { available, biometryType });
    } catch (error) {
      console.error('Biometric check failed:', error);
    }
  }

  // Login with email and password
  async login(credentials: LoginCredentials): Promise<User> {
    try {
      const deviceInfo = await this.getDeviceInfo();
      
      const response = await fetch(`${this.baseURL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...credentials,
          deviceInfo,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Login failed');
      }

      const data = await response.json();
      
      // Store authentication data
      await this.storeAuthData(data.user, {
        accessToken: data.token,
        refreshToken: data.refreshToken,
        expiresAt: Date.now() + (data.expiresIn * 1000),
      });

      // Setup biometric authentication if requested
      if (credentials.rememberMe) {
        await this.setupBiometricAuth(credentials);
      }

      return data.user;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  }

  // Register new user
  async register(userData: RegisterData): Promise<User> {
    try {
      const deviceInfo = await this.getDeviceInfo();
      
      const response = await fetch(`${this.baseURL}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...userData,
          deviceInfo,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Registration failed');
      }

      const data = await response.json();
      
      // Store authentication data if auto-login is enabled
      if (data.token) {
        await this.storeAuthData(data.user, {
          accessToken: data.token,
          refreshToken: data.refreshToken,
          expiresAt: Date.now() + (data.expiresIn * 1000),
        });
      }

      return data.user;
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  }

  // Biometric authentication
  async authenticateWithBiometrics(): Promise<User | null> {
    try {
      const { available } = await this.biometrics.isSensorAvailable();
      if (!available) {
        throw new Error('Biometric authentication not available');
      }

      // Check if biometric credentials are stored
      const credentials = await Keychain.getInternetCredentials('ollama_biometric');
      if (!credentials || !credentials.password) {
        throw new Error('No biometric credentials stored');
      }

      // Prompt for biometric authentication
      const { success } = await this.biometrics.simplePrompt({
        promptMessage: 'Authenticate to access OllamaMax',
        cancelButtonText: 'Cancel',
      });

      if (!success) {
        throw new Error('Biometric authentication failed');
      }

      // Decrypt and use stored credentials
      const storedCredentials = JSON.parse(credentials.password);
      return await this.login(storedCredentials);
    } catch (error) {
      console.error('Biometric authentication failed:', error);
      throw error;
    }
  }

  // Setup biometric authentication
  private async setupBiometricAuth(credentials: LoginCredentials): Promise<void> {
    try {
      const { available } = await this.biometrics.isSensorAvailable();
      if (!available) return;

      // Store credentials securely for biometric auth
      await Keychain.setInternetCredentials(
        'ollama_biometric',
        credentials.email,
        JSON.stringify({
          email: credentials.email,
          password: credentials.password,
        })
      );
    } catch (error) {
      console.error('Failed to setup biometric auth:', error);
    }
  }

  // Logout
  async logout(): Promise<void> {
    try {
      // Notify server
      if (this.tokens?.accessToken) {
        await fetch(`${this.baseURL}/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${this.tokens.accessToken}`,
            'Content-Type': 'application/json',
          },
        }).catch(() => {
          // Ignore errors during logout
        });
      }
    } finally {
      // Always clear local data
      await this.clearAuthData();
    }
  }

  // Refresh access token
  async refreshAccessToken(): Promise<boolean> {
    try {
      if (!this.tokens?.refreshToken) {
        throw new Error('No refresh token available');
      }

      const response = await fetch(`${this.baseURL}/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refreshToken: this.tokens.refreshToken,
        }),
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      
      // Update tokens
      this.tokens = {
        accessToken: data.token,
        refreshToken: data.refreshToken || this.tokens.refreshToken,
        expiresAt: Date.now() + (data.expiresIn * 1000),
      };

      // Store updated tokens
      await this.storeTokens(this.tokens);
      
      // Setup next refresh
      this.setupTokenRefresh();

      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      await this.clearAuthData();
      return false;
    }
  }

  // Check if token needs refresh
  async refreshTokenIfNeeded(): Promise<void> {
    if (!this.tokens) return;

    // Refresh if token expires in the next 5 minutes
    const fiveMinutes = 5 * 60 * 1000;
    if (this.tokens.expiresAt - Date.now() < fiveMinutes) {
      await this.refreshAccessToken();
    }
  }

  // Setup automatic token refresh
  private setupTokenRefresh(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    if (!this.tokens) return;

    // Schedule refresh 5 minutes before expiry
    const refreshTime = this.tokens.expiresAt - Date.now() - (5 * 60 * 1000);
    
    if (refreshTime > 0) {
      this.refreshTimer = setTimeout(async () => {
        await this.refreshAccessToken();
      }, refreshTime);
    }
  }

  // Store authentication data
  private async storeAuthData(user: User, tokens: AuthTokens): Promise<void> {
    this.user = user;
    this.tokens = tokens;

    // Store user data in AsyncStorage
    await AsyncStorage.setItem('user', JSON.stringify(user));
    
    // Store tokens in secure storage
    await this.storeTokens(tokens);
    
    // Setup token refresh
    this.setupTokenRefresh();
  }

  // Store tokens securely
  private async storeTokens(tokens: AuthTokens): Promise<void> {
    await Keychain.setInternetCredentials(
      'ollama_tokens',
      'tokens',
      JSON.stringify(tokens)
    );
  }

  // Clear authentication data
  private async clearAuthData(): Promise<void> {
    this.user = null;
    this.tokens = null;

    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }

    // Clear stored data
    await AsyncStorage.removeItem('user');
    await Keychain.resetInternetCredentials('ollama_tokens');
    await Keychain.resetInternetCredentials('ollama_biometric');
  }

  // Get device information
  private async getDeviceInfo(): Promise<object> {
    return {
      deviceId: await DeviceInfo.getUniqueId(),
      platform: DeviceInfo.getSystemName(),
      version: DeviceInfo.getSystemVersion(),
      appVersion: DeviceInfo.getVersion(),
      buildNumber: DeviceInfo.getBuildNumber(),
      model: DeviceInfo.getModel(),
      brand: DeviceInfo.getBrand(),
    };
  }

  // Make authenticated API request
  async makeAuthenticatedRequest(endpoint: string, options: RequestInit = {}): Promise<Response> {
    if (!this.tokens?.accessToken) {
      throw new Error('No access token available');
    }

    const response = await fetch(`${this.baseURL.replace('/auth', '')}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.tokens.accessToken}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    // Handle token expiration
    if (response.status === 401) {
      const refreshed = await this.refreshAccessToken();
      if (refreshed && this.tokens?.accessToken) {
        // Retry request with new token
        return fetch(`${this.baseURL.replace('/auth', '')}${endpoint}`, {
          ...options,
          headers: {
            'Authorization': `Bearer ${this.tokens.accessToken}`,
            'Content-Type': 'application/json',
            ...options.headers,
          },
        });
      } else {
        throw new Error('Authentication required');
      }
    }

    return response;
  }

  // Getters
  get currentUser(): User | null {
    return this.user;
  }

  get isAuthenticated(): boolean {
    return !!(this.user && this.tokens && this.tokens.expiresAt > Date.now());
  }

  get accessToken(): string | null {
    return this.tokens?.accessToken || null;
  }

  // Check if biometric authentication is available and set up
  async isBiometricAuthAvailable(): Promise<boolean> {
    try {
      const { available } = await this.biometrics.isSensorAvailable();
      if (!available) return false;

      const credentials = await Keychain.getInternetCredentials('ollama_biometric');
      return !!(credentials && credentials.password);
    } catch {
      return false;
    }
  }
}

// Export singleton instance
export const authService = new AuthService();
export default authService;
