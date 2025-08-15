/**
 * Internationalization Service
 * 
 * Provides multi-language support, locale management, and text formatting utilities.
 */

class I18nService {
  constructor() {
    this.currentLocale = 'en-US';
    this.fallbackLocale = 'en-US';
    this.translations = new Map();
    this.formatters = new Map();
    this.rtlLanguages = new Set(['ar', 'he', 'fa', 'ur']);
    
    this.init();
  }

  // Initialize i18n service
  async init() {
    // Detect user's preferred language
    this.detectLocale();
    
    // Load default translations
    await this.loadTranslations(this.currentLocale);
    
    // Setup formatters
    this.setupFormatters();
    
    // Apply locale settings
    this.applyLocale();
  }

  // Detect user's preferred locale
  detectLocale() {
    // Check localStorage first
    const savedLocale = localStorage.getItem('ollama-locale');
    if (savedLocale && this.isValidLocale(savedLocale)) {
      this.currentLocale = savedLocale;
      return;
    }
    
    // Check browser language
    const browserLanguages = navigator.languages || [navigator.language];
    for (const lang of browserLanguages) {
      if (this.isValidLocale(lang)) {
        this.currentLocale = lang;
        return;
      }
      
      // Try language without region
      const langCode = lang.split('-')[0];
      if (this.isValidLocale(langCode)) {
        this.currentLocale = langCode;
        return;
      }
    }
  }

  // Check if locale is valid/supported
  isValidLocale(locale) {
    const supportedLocales = [
      'en-US', 'en-GB', 'en',
      'es-ES', 'es-MX', 'es',
      'fr-FR', 'fr-CA', 'fr',
      'de-DE', 'de',
      'it-IT', 'it',
      'pt-BR', 'pt-PT', 'pt',
      'ru-RU', 'ru',
      'zh-CN', 'zh-TW', 'zh',
      'ja-JP', 'ja',
      'ko-KR', 'ko',
      'ar-SA', 'ar',
      'he-IL', 'he'
    ];
    
    return supportedLocales.includes(locale);
  }

  // Load translations for a locale
  async loadTranslations(locale) {
    try {
      // In a real app, this would fetch from an API or import translation files
      const translations = await this.fetchTranslations(locale);
      this.translations.set(locale, translations);
      
      console.log(`[i18n] Loaded translations for ${locale}`);
    } catch (error) {
      console.error(`[i18n] Failed to load translations for ${locale}:`, error);
      
      // Load fallback if not already loaded
      if (locale !== this.fallbackLocale && !this.translations.has(this.fallbackLocale)) {
        await this.loadTranslations(this.fallbackLocale);
      }
    }
  }

  // Fetch translations (mock implementation)
  async fetchTranslations(locale) {
    // Mock translations - in real app, this would be loaded from files or API
    const mockTranslations = {
      'en-US': {
        // Common
        'common.loading': 'Loading...',
        'common.error': 'Error',
        'common.success': 'Success',
        'common.warning': 'Warning',
        'common.info': 'Information',
        'common.cancel': 'Cancel',
        'common.confirm': 'Confirm',
        'common.save': 'Save',
        'common.delete': 'Delete',
        'common.edit': 'Edit',
        'common.close': 'Close',
        'common.back': 'Back',
        'common.next': 'Next',
        'common.previous': 'Previous',
        'common.search': 'Search',
        'common.filter': 'Filter',
        'common.sort': 'Sort',
        'common.refresh': 'Refresh',
        
        // Authentication
        'auth.login': 'Sign In',
        'auth.logout': 'Sign Out',
        'auth.register': 'Sign Up',
        'auth.email': 'Email Address',
        'auth.password': 'Password',
        'auth.confirmPassword': 'Confirm Password',
        'auth.firstName': 'First Name',
        'auth.lastName': 'Last Name',
        'auth.forgotPassword': 'Forgot Password?',
        'auth.rememberMe': 'Remember me',
        'auth.createAccount': 'Create Account',
        'auth.alreadyHaveAccount': 'Already have an account?',
        'auth.dontHaveAccount': "Don't have an account?",
        
        // Dashboard
        'dashboard.title': 'Dashboard',
        'dashboard.welcome': 'Welcome back, {name}!',
        'dashboard.clusterStatus': 'Cluster Status',
        'dashboard.totalRequests': 'Total Requests',
        'dashboard.responseTime': 'Response Time',
        'dashboard.errorRate': 'Error Rate',
        'dashboard.uptime': 'Uptime',
        'dashboard.nodes': 'Nodes',
        'dashboard.models': 'Models',
        'dashboard.metrics': 'Metrics',
        'dashboard.settings': 'Settings',
        
        // Accessibility
        'a11y.skipToMain': 'Skip to main content',
        'a11y.openMenu': 'Open menu',
        'a11y.closeMenu': 'Close menu',
        'a11y.toggleTheme': 'Toggle theme',
        'a11y.loading': 'Loading content',
        'a11y.error': 'Error occurred',
        'a11y.success': 'Action completed successfully',
        
        // Time formats
        'time.now': 'now',
        'time.minuteAgo': '{count} minute ago',
        'time.minutesAgo': '{count} minutes ago',
        'time.hourAgo': '{count} hour ago',
        'time.hoursAgo': '{count} hours ago',
        'time.dayAgo': '{count} day ago',
        'time.daysAgo': '{count} days ago'
      },
      
      'es-ES': {
        'common.loading': 'Cargando...',
        'common.error': 'Error',
        'common.success': 'Éxito',
        'common.warning': 'Advertencia',
        'common.info': 'Información',
        'common.cancel': 'Cancelar',
        'common.confirm': 'Confirmar',
        'common.save': 'Guardar',
        'common.delete': 'Eliminar',
        'common.edit': 'Editar',
        'common.close': 'Cerrar',
        'common.back': 'Atrás',
        'common.next': 'Siguiente',
        'common.previous': 'Anterior',
        'common.search': 'Buscar',
        'common.filter': 'Filtrar',
        'common.sort': 'Ordenar',
        'common.refresh': 'Actualizar',
        
        'auth.login': 'Iniciar Sesión',
        'auth.logout': 'Cerrar Sesión',
        'auth.register': 'Registrarse',
        'auth.email': 'Correo Electrónico',
        'auth.password': 'Contraseña',
        'auth.confirmPassword': 'Confirmar Contraseña',
        'auth.firstName': 'Nombre',
        'auth.lastName': 'Apellido',
        'auth.forgotPassword': '¿Olvidaste tu contraseña?',
        'auth.rememberMe': 'Recordarme',
        'auth.createAccount': 'Crear Cuenta',
        'auth.alreadyHaveAccount': '¿Ya tienes una cuenta?',
        'auth.dontHaveAccount': '¿No tienes una cuenta?',
        
        'dashboard.title': 'Panel de Control',
        'dashboard.welcome': '¡Bienvenido de vuelta, {name}!',
        'dashboard.clusterStatus': 'Estado del Clúster',
        'dashboard.totalRequests': 'Solicitudes Totales',
        'dashboard.responseTime': 'Tiempo de Respuesta',
        'dashboard.errorRate': 'Tasa de Error',
        'dashboard.uptime': 'Tiempo Activo',
        'dashboard.nodes': 'Nodos',
        'dashboard.models': 'Modelos',
        'dashboard.metrics': 'Métricas',
        'dashboard.settings': 'Configuración'
      }
    };
    
    return mockTranslations[locale] || mockTranslations[this.fallbackLocale];
  }

  // Setup number and date formatters
  setupFormatters() {
    // Number formatter
    this.formatters.set('number', new Intl.NumberFormat(this.currentLocale));
    
    // Currency formatter
    this.formatters.set('currency', new Intl.NumberFormat(this.currentLocale, {
      style: 'currency',
      currency: 'USD' // This could be configurable
    }));
    
    // Percentage formatter
    this.formatters.set('percentage', new Intl.NumberFormat(this.currentLocale, {
      style: 'percent',
      minimumFractionDigits: 1,
      maximumFractionDigits: 2
    }));
    
    // Date formatters
    this.formatters.set('date', new Intl.DateTimeFormat(this.currentLocale));
    this.formatters.set('time', new Intl.DateTimeFormat(this.currentLocale, {
      hour: '2-digit',
      minute: '2-digit'
    }));
    this.formatters.set('datetime', new Intl.DateTimeFormat(this.currentLocale, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }));
    
    // Relative time formatter
    if (Intl.RelativeTimeFormat) {
      this.formatters.set('relative', new Intl.RelativeTimeFormat(this.currentLocale, {
        numeric: 'auto'
      }));
    }
  }

  // Apply locale settings to document
  applyLocale() {
    // Set document language
    document.documentElement.lang = this.currentLocale;
    
    // Set text direction
    const langCode = this.currentLocale.split('-')[0];
    document.documentElement.dir = this.rtlLanguages.has(langCode) ? 'rtl' : 'ltr';
    
    // Update formatters
    this.setupFormatters();
  }

  // Translate text
  t(key, params = {}) {
    const translations = this.translations.get(this.currentLocale) || 
                        this.translations.get(this.fallbackLocale) || {};
    
    let text = translations[key] || key;
    
    // Replace parameters
    Object.keys(params).forEach(param => {
      const placeholder = `{${param}}`;
      text = text.replace(new RegExp(placeholder, 'g'), params[param]);
    });
    
    return text;
  }

  // Pluralization
  plural(key, count, params = {}) {
    const pluralRules = new Intl.PluralRule(this.currentLocale);
    const rule = pluralRules.select(count);
    
    // Try specific plural form first
    const pluralKey = `${key}.${rule}`;
    if (this.hasTranslation(pluralKey)) {
      return this.t(pluralKey, { count, ...params });
    }
    
    // Fallback to base key
    return this.t(key, { count, ...params });
  }

  // Check if translation exists
  hasTranslation(key) {
    const translations = this.translations.get(this.currentLocale) || 
                        this.translations.get(this.fallbackLocale) || {};
    return key in translations;
  }

  // Format number
  formatNumber(number, options = {}) {
    const formatter = this.formatters.get('number');
    return formatter.format(number);
  }

  // Format currency
  formatCurrency(amount, currency = 'USD') {
    const formatter = new Intl.NumberFormat(this.currentLocale, {
      style: 'currency',
      currency
    });
    return formatter.format(amount);
  }

  // Format percentage
  formatPercentage(value) {
    const formatter = this.formatters.get('percentage');
    return formatter.format(value);
  }

  // Format date
  formatDate(date, options = {}) {
    const formatter = new Intl.DateTimeFormat(this.currentLocale, options);
    return formatter.format(new Date(date));
  }

  // Format relative time
  formatRelativeTime(date) {
    const now = new Date();
    const target = new Date(date);
    const diffMs = target.getTime() - now.getTime();
    const diffSeconds = Math.round(diffMs / 1000);
    const diffMinutes = Math.round(diffSeconds / 60);
    const diffHours = Math.round(diffMinutes / 60);
    const diffDays = Math.round(diffHours / 24);
    
    const formatter = this.formatters.get('relative');
    if (!formatter) {
      // Fallback for browsers without RelativeTimeFormat
      if (Math.abs(diffMinutes) < 1) return this.t('time.now');
      if (Math.abs(diffMinutes) < 60) {
        return diffMinutes === 1 ? 
          this.t('time.minuteAgo', { count: 1 }) :
          this.t('time.minutesAgo', { count: Math.abs(diffMinutes) });
      }
      if (Math.abs(diffHours) < 24) {
        return diffHours === 1 ?
          this.t('time.hourAgo', { count: 1 }) :
          this.t('time.hoursAgo', { count: Math.abs(diffHours) });
      }
      return diffDays === 1 ?
        this.t('time.dayAgo', { count: 1 }) :
        this.t('time.daysAgo', { count: Math.abs(diffDays) });
    }
    
    if (Math.abs(diffDays) >= 1) {
      return formatter.format(diffDays, 'day');
    }
    if (Math.abs(diffHours) >= 1) {
      return formatter.format(diffHours, 'hour');
    }
    if (Math.abs(diffMinutes) >= 1) {
      return formatter.format(diffMinutes, 'minute');
    }
    return formatter.format(diffSeconds, 'second');
  }

  // Change locale
  async changeLocale(locale) {
    if (!this.isValidLocale(locale)) {
      console.warn(`[i18n] Invalid locale: ${locale}`);
      return false;
    }
    
    this.currentLocale = locale;
    localStorage.setItem('ollama-locale', locale);
    
    // Load translations if not already loaded
    if (!this.translations.has(locale)) {
      await this.loadTranslations(locale);
    }
    
    // Apply locale settings
    this.applyLocale();
    
    // Notify listeners
    this.onLocaleChange(locale);
    
    return true;
  }

  // Get current locale info
  getLocaleInfo() {
    const langCode = this.currentLocale.split('-')[0];
    
    return {
      locale: this.currentLocale,
      language: langCode,
      isRTL: this.rtlLanguages.has(langCode),
      hasTranslations: this.translations.has(this.currentLocale),
      supportedLocales: this.getSupportedLocales()
    };
  }

  // Get supported locales
  getSupportedLocales() {
    return [
      { code: 'en-US', name: 'English (US)', nativeName: 'English (US)' },
      { code: 'en-GB', name: 'English (UK)', nativeName: 'English (UK)' },
      { code: 'es-ES', name: 'Spanish (Spain)', nativeName: 'Español (España)' },
      { code: 'es-MX', name: 'Spanish (Mexico)', nativeName: 'Español (México)' },
      { code: 'fr-FR', name: 'French (France)', nativeName: 'Français (France)' },
      { code: 'de-DE', name: 'German (Germany)', nativeName: 'Deutsch (Deutschland)' },
      { code: 'it-IT', name: 'Italian (Italy)', nativeName: 'Italiano (Italia)' },
      { code: 'pt-BR', name: 'Portuguese (Brazil)', nativeName: 'Português (Brasil)' },
      { code: 'ru-RU', name: 'Russian (Russia)', nativeName: 'Русский (Россия)' },
      { code: 'zh-CN', name: 'Chinese (Simplified)', nativeName: '中文 (简体)' },
      { code: 'ja-JP', name: 'Japanese (Japan)', nativeName: '日本語 (日本)' },
      { code: 'ko-KR', name: 'Korean (Korea)', nativeName: '한국어 (대한민국)' },
      { code: 'ar-SA', name: 'Arabic (Saudi Arabia)', nativeName: 'العربية (السعودية)' }
    ];
  }

  // Event handler (to be overridden)
  onLocaleChange(locale) {
    console.log(`[i18n] Locale changed to: ${locale}`);
  }

  // Cleanup
  destroy() {
    this.translations.clear();
    this.formatters.clear();
  }
}

// Create singleton instance
const i18nService = new I18nService();

export default i18nService;
