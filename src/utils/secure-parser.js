/**
 * Secure Parser Utilities
 * Provides safe alternatives to eval() and other dangerous parsing methods
 */

/**
 * Safe JSON parser with validation and size limits
 */
export class SecureJSONParser {
  constructor(options = {}) {
    this.maxSize = options.maxSize || 1024 * 1024; // 1MB default
    this.maxDepth = options.maxDepth || 10;
    this.allowedTypes = options.allowedTypes || ['string', 'number', 'boolean', 'object', 'null'];
  }

  /**
   * Safely parse JSON with validation
   */
  parse(jsonString, reviver = null) {
    // Input validation
    if (typeof jsonString !== 'string') {
      throw new Error('Input must be a string');
    }

    if (jsonString.length > this.maxSize) {
      throw new Error(`JSON string exceeds maximum size of ${this.maxSize} bytes`);
    }

    // Check for potential injection patterns
    this.validateJSONString(jsonString);

    try {
      const parsed = JSON.parse(jsonString, reviver);
      this.validateParsedObject(parsed, 0);
      return parsed;
    } catch (error) {
      throw new Error(`JSON parsing failed: ${error.message}`);
    }
  }

  /**
   * Validate JSON string for dangerous patterns
   */
  validateJSONString(jsonString) {
    // Check for dangerous patterns that might indicate injection attempts
    const dangerousPatterns = [
      /eval\s*\(/g,
      /function\s*\(/g,
      /__proto__/g,
      /constructor/g,
      /prototype/g,
      /javascript:/g,
      /data:.*base64/g,
      /<script/gi,
      /on\w+\s*=/gi
    ];

    for (const pattern of dangerousPatterns) {
      if (pattern.test(jsonString)) {
        throw new Error(`Potentially dangerous pattern detected in JSON: ${pattern.source}`);
      }
    }
  }

  /**
   * Validate parsed object structure and depth
   */
  validateParsedObject(obj, depth) {
    if (depth > this.maxDepth) {
      throw new Error(`Object depth exceeds maximum of ${this.maxDepth}`);
    }

    const objType = obj === null ? 'null' : typeof obj;
    
    if (!this.allowedTypes.includes(objType)) {
      throw new Error(`Type '${objType}' is not allowed`);
    }

    if (objType === 'object' && obj !== null) {
      // Check for prototype pollution attempts
      if (obj.hasOwnProperty('__proto__') || 
          obj.hasOwnProperty('constructor') || 
          obj.hasOwnProperty('prototype')) {
        throw new Error('Object contains dangerous properties');
      }

      // Recursively validate nested objects
      if (Array.isArray(obj)) {
        for (let i = 0; i < obj.length; i++) {
          this.validateParsedObject(obj[i], depth + 1);
        }
      } else {
        for (const key in obj) {
          if (obj.hasOwnProperty(key)) {
            this.validateParsedObject(obj[key], depth + 1);
          }
        }
      }
    }
  }

  /**
   * Safely stringify object with size limits
   */
  stringify(obj, replacer = null, space = null) {
    try {
      // First pass to check size
      const testString = JSON.stringify(obj, replacer);
      if (testString.length > this.maxSize) {
        throw new Error(`Stringified object exceeds maximum size of ${this.maxSize} bytes`);
      }

      return JSON.stringify(obj, replacer, space);
    } catch (error) {
      throw new Error(`JSON stringification failed: ${error.message}`);
    }
  }
}

/**
 * Safe expression evaluator for simple mathematical expressions
 * Alternative to eval() for mathematical calculations
 */
export class SafeExpressionEvaluator {
  constructor() {
    this.allowedOperators = ['+', '-', '*', '/', '(', ')', ' ', '.'];
    this.allowedFunctions = ['Math.abs', 'Math.ceil', 'Math.floor', 'Math.round', 'Math.max', 'Math.min'];
  }

  /**
   * Safely evaluate mathematical expressions
   */
  evaluate(expression) {
    if (typeof expression !== 'string') {
      throw new Error('Expression must be a string');
    }

    // Remove whitespace
    const cleaned = expression.trim();

    if (!cleaned) {
      throw new Error('Expression cannot be empty');
    }

    // Validate expression contains only safe characters
    this.validateExpression(cleaned);

    try {
      // Use Function constructor instead of eval (slightly safer)
      // Only allows return statements with mathematical expressions
      const func = new Function('Math', `"use strict"; return (${cleaned});`);
      const result = func(Math);

      if (typeof result !== 'number' || !isFinite(result)) {
        throw new Error('Expression must evaluate to a finite number');
      }

      return result;
    } catch (error) {
      throw new Error(`Expression evaluation failed: ${error.message}`);
    }
  }

  /**
   * Validate expression for safety
   */
  validateExpression(expression) {
    // Check for dangerous patterns
    const dangerousPatterns = [
      /[a-zA-Z_$][a-zA-Z0-9_$]*\s*\(/g, // Function calls (except Math.*)
      /\w+\s*=/g, // Variable assignments
      /[{}]/g, // Object literals or blocks
      /[[\]]/g, // Array access
      /;/g, // Statement separators
      /eval|function|constructor|prototype|__proto__|import|require|process|global|window|document/gi
    ];

    for (const pattern of dangerousPatterns) {
      if (pattern.test(expression)) {
        // Allow Math functions
        const mathFunctionPattern = /^Math\.(abs|ceil|floor|round|max|min)\s*\(/;
        if (!mathFunctionPattern.test(expression.match(pattern)?.[0] || '')) {
          throw new Error(`Dangerous pattern detected in expression: ${pattern.source}`);
        }
      }
    }

    // Validate only allowed characters
    const allowedPattern = /^[0-9+\-*/().\s,Math.abcdefghijklmnopqrstuvwxyz]*$/i;
    if (!allowedPattern.test(expression)) {
      throw new Error('Expression contains invalid characters');
    }

    // Additional validation for Math functions
    const mathFunctions = expression.match(/Math\.\w+/g) || [];
    for (const func of mathFunctions) {
      if (!this.allowedFunctions.includes(func)) {
        throw new Error(`Math function ${func} is not allowed`);
      }
    }
  }
}

/**
 * Safe template processor (alternative to template literals with eval)
 */
export class SafeTemplateProcessor {
  constructor(options = {}) {
    this.maxTemplateSize = options.maxTemplateSize || 64 * 1024; // 64KB
    this.maxSubstitutions = options.maxSubstitutions || 100;
    this.allowedHelpers = options.allowedHelpers || {};
  }

  /**
   * Process template with safe variable substitution
   */
  process(template, variables = {}) {
    if (typeof template !== 'string') {
      throw new Error('Template must be a string');
    }

    if (template.length > this.maxTemplateSize) {
      throw new Error(`Template exceeds maximum size of ${this.maxTemplateSize} bytes`);
    }

    // Validate template for dangerous patterns
    this.validateTemplate(template);

    // Find all variable placeholders
    const placeholderPattern = /\{\{\s*([a-zA-Z_$][a-zA-Z0-9_$]*(?:\.[a-zA-Z_$][a-zA-Z0-9_$]*)*)\s*\}\}/g;
    const placeholders = [];
    let match;

    while ((match = placeholderPattern.exec(template)) !== null) {
      placeholders.push({
        full: match[0],
        variable: match[1],
        index: match.index
      });
    }

    if (placeholders.length > this.maxSubstitutions) {
      throw new Error(`Template exceeds maximum substitutions of ${this.maxSubstitutions}`);
    }

    // Process substitutions
    let result = template;
    for (const placeholder of placeholders.reverse()) { // Reverse to maintain indices
      const value = this.getVariableValue(placeholder.variable, variables);
      const safeValue = this.sanitizeValue(value);
      result = result.slice(0, placeholder.index) + safeValue + result.slice(placeholder.index + placeholder.full.length);
    }

    return result;
  }

  /**
   * Validate template for dangerous patterns
   */
  validateTemplate(template) {
    const dangerousPatterns = [
      /eval\s*\(/g,
      /function\s*\(/g,
      /javascript:/g,
      /<script/gi,
      /on\w+\s*=/gi,
      /\{\{\{.*?\}\}\}/g // Triple braces might indicate unsafe raw insertion
    ];

    for (const pattern of dangerousPatterns) {
      if (pattern.test(template)) {
        throw new Error(`Dangerous pattern detected in template: ${pattern.source}`);
      }
    }
  }

  /**
   * Get variable value with safe property access
   */
  getVariableValue(path, variables) {
    const keys = path.split('.');
    let current = variables;

    for (const key of keys) {
      if (current === null || current === undefined) {
        return '';
      }

      if (typeof current !== 'object') {
        return '';
      }

      // Prevent prototype pollution
      if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
        return '';
      }

      current = current[key];
    }

    return current;
  }

  /**
   * Sanitize value for safe output
   */
  sanitizeValue(value) {
    if (value === null || value === undefined) {
      return '';
    }

    const stringValue = String(value);

    // HTML escape by default
    return stringValue
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/\//g, '&#x2F;');
  }
}

/**
 * Safe configuration parser
 */
export class SafeConfigParser {
  /**
   * Parse configuration from various formats
   */
  static parseConfig(content, format = 'json') {
    const parser = new SecureJSONParser({
      maxSize: 512 * 1024, // 512KB for config files
      maxDepth: 15
    });

    switch (format.toLowerCase()) {
      case 'json':
        return parser.parse(content);
        
      case 'env':
        return this.parseEnvFormat(content);
        
      default:
        throw new Error(`Unsupported configuration format: ${format}`);
    }
  }

  /**
   * Parse .env format safely
   */
  static parseEnvFormat(content) {
    const lines = content.split('\n');
    const config = {};

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      
      // Skip comments and empty lines
      if (!line || line.startsWith('#')) {
        continue;
      }

      const eqIndex = line.indexOf('=');
      if (eqIndex === -1) {
        continue;
      }

      const key = line.slice(0, eqIndex).trim();
      let value = line.slice(eqIndex + 1).trim();

      // Remove quotes if present
      if ((value.startsWith('"') && value.endsWith('"')) ||
          (value.startsWith("'") && value.endsWith("'"))) {
        value = value.slice(1, -1);
      }

      // Validate key
      if (!/^[A-Z_][A-Z0-9_]*$/i.test(key)) {
        throw new Error(`Invalid environment variable name: ${key}`);
      }

      config[key] = value;
    }

    return config;
  }
}

// Export instances with default configurations
export const secureJSON = new SecureJSONParser();
export const safeEvaluator = new SafeExpressionEvaluator();
export const templateProcessor = new SafeTemplateProcessor();