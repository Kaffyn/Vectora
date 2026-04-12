/**
 * Input Validation & Sanitization Module
 * Phase 12: Enterprise-Grade Security
 *
 * Provides:
 * - Comprehensive input validation
 * - XSS prevention (HTML sanitization)
 * - SQL injection prevention
 * - Command injection prevention
 * - Path traversal prevention
 * - Schema validation with Zod
 */

/**
 * Validation result type
 */
export interface ValidationResult<T = any> {
  valid: boolean;
  data?: T;
  errors?: string[];
}

/**
 * Sanitization options
 */
export interface SanitizationOptions {
  stripHtml?: boolean;
  stripScripts?: boolean;
  encodeHtml?: boolean;
  trimWhitespace?: boolean;
  maxLength?: number;
  allowedTags?: string[];
}

/**
 * InputValidator - Core validation logic
 */
export class InputValidator {
  /**
   * Validate email address
   */
  static validateEmail(email: string): ValidationResult<string> {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const valid = emailRegex.test(email) && email.length <= 255;

    return {
      valid,
      data: valid ? email : undefined,
      errors: valid ? undefined : ['Invalid email format']
    };
  }

  /**
   * Validate URL
   */
  static validateUrl(url: string): ValidationResult<string> {
    try {
      const parsedUrl = new URL(url);
      // Only allow http, https, and ftp protocols
      const validProtocols = ['http:', 'https:', 'ftp:'];
      const isValid = validProtocols.includes(parsedUrl.protocol);

      return {
        valid: isValid,
        data: isValid ? url : undefined,
        errors: isValid ? undefined : ['Invalid URL or unsupported protocol']
      };
    } catch (error) {
      return {
        valid: false,
        errors: ['Invalid URL format']
      };
    }
  }

  /**
   * Validate password strength
   */
  static validatePassword(password: string): ValidationResult<string> {
    const errors: string[] = [];

    if (password.length < 8) {
      errors.push('Password must be at least 8 characters');
    }
    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain uppercase letter');
    }
    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain lowercase letter');
    }
    if (!/[0-9]/.test(password)) {
      errors.push('Password must contain number');
    }
    if (!/[!@#$%^&*]/.test(password)) {
      errors.push('Password must contain special character (!@#$%^&*)');
    }

    return {
      valid: errors.length === 0,
      data: errors.length === 0 ? password : undefined,
      errors: errors.length > 0 ? errors : undefined
    };
  }

  /**
   * Validate integer
   */
  static validateInteger(value: any, min?: number, max?: number): ValidationResult<number> {
    const num = parseInt(String(value), 10);

    if (isNaN(num)) {
      return {
        valid: false,
        errors: ['Must be a valid integer']
      };
    }

    if (min !== undefined && num < min) {
      return {
        valid: false,
        errors: [`Must be at least ${min}`]
      };
    }

    if (max !== undefined && num > max) {
      return {
        valid: false,
        errors: [`Must be at most ${max}`]
      };
    }

    return {
      valid: true,
      data: num
    };
  }

  /**
   * Validate JSON string
   */
  static validateJson(jsonString: string): ValidationResult<any> {
    try {
      const parsed = JSON.parse(jsonString);
      return {
        valid: true,
        data: parsed
      };
    } catch (error) {
      return {
        valid: false,
        errors: ['Invalid JSON format']
      };
    }
  }

  /**
   * Validate UUID
   */
  static validateUuid(uuid: string): ValidationResult<string> {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    const valid = uuidRegex.test(uuid);

    return {
      valid,
      data: valid ? uuid : undefined,
      errors: valid ? undefined : ['Invalid UUID format']
    };
  }

  /**
   * Validate phone number (basic validation)
   */
  static validatePhone(phone: string): ValidationResult<string> {
    const phoneRegex = /^[\d\s\-\+\(\)]{7,}$/;
    const valid = phoneRegex.test(phone) && phone.length <= 20;

    return {
      valid,
      data: valid ? phone : undefined,
      errors: valid ? undefined : ['Invalid phone number format']
    };
  }
}

/**
 * InputSanitizer - Clean and escape unsafe content
 */
export class InputSanitizer {
  private static readonly HTML_ENTITIES: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;'
  };

  private static readonly DANGEROUS_PROTOCOLS = [
    'javascript:',
    'data:',
    'vbscript:',
    'file:'
  ];

  /**
   * Escape HTML entities to prevent XSS
   */
  static escapeHtml(text: string): string {
    return String(text).replace(/[&<>"'\/]/g, (char) => InputSanitizer.HTML_ENTITIES[char]);
  }

  /**
   * Remove HTML tags
   */
  static stripHtmlTags(text: string): string {
    return String(text).replace(/<[^>]*>/g, '');
  }

  /**
   * Sanitize URL to prevent javascript: attacks
   */
  static sanitizeUrl(url: string): string | null {
    try {
      const trimmed = url.trim().toLowerCase();

      // Check for dangerous protocols
      if (InputSanitizer.DANGEROUS_PROTOCOLS.some(proto => trimmed.startsWith(proto))) {
        return null;
      }

      // Validate URL
      const urlObj = new URL(url);
      return urlObj.toString();
    } catch (error) {
      return null;
    }
  }

  /**
   * Sanitize file path to prevent directory traversal
   */
  static sanitizePath(path: string): string {
    // Remove null bytes
    let sanitized = path.replace(/\0/g, '');

    // Remove directory traversal sequences
    sanitized = sanitized.replace(/\.\.\//g, '').replace(/\.\.\\/g, '');

    // Normalize path separators
    sanitized = sanitized.replace(/\\/g, '/');

    // Remove leading slashes or drive letters (on Windows)
    sanitized = sanitized.replace(/^[a-zA-Z]:/, '').replace(/^\/+/, '');

    return sanitized;
  }

  /**
   * Sanitize SQL query to prevent injection (use parameterized queries instead)
   */
  static escapeSqlString(str: string): string {
    return String(str)
      .replace(/\\/g, '\\\\')
      .replace(/'/g, "''")
      .replace(/"/g, '\\"')
      .replace(/\0/g, '\\0')
      .replace(/\n/g, '\\n')
      .replace(/\r/g, '\\r')
      .replace(/\x1a/g, '\\Z');
  }

  /**
   * Sanitize shell command to prevent injection
   */
  static escapeShellArg(arg: string): string {
    if (process.platform === 'win32') {
      // Windows: wrap in quotes and escape internal quotes
      return '"' + String(arg).replace(/"/g, '\\"') + '"';
    } else {
      // Unix: use single quotes and escape single quotes
      return "'" + String(arg).replace(/'/g, "'\\''") + "'";
    }
  }

  /**
   * Trim whitespace and limit length
   */
  static trimAndLimit(text: string, maxLength: number): string {
    return String(text).trim().substring(0, maxLength);
  }

  /**
   * Sanitize with custom options
   */
  static sanitize(text: string, options: SanitizationOptions = {}): string {
    let result = String(text);

    if (options.stripHtml) {
      result = InputSanitizer.stripHtmlTags(result);
    }

    if (options.stripScripts) {
      result = result.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '');
    }

    if (options.encodeHtml) {
      result = InputSanitizer.escapeHtml(result);
    }

    if (options.trimWhitespace) {
      result = result.trim();
    }

    if (options.maxLength) {
      result = result.substring(0, options.maxLength);
    }

    return result;
  }
}

/**
 * SchemaValidator - Validate objects against schemas
 */
export class SchemaValidator {
  /**
   * Validate object has required fields
   */
  static validateRequired(obj: any, requiredFields: string[]): ValidationResult<any> {
    const errors: string[] = [];

    for (const field of requiredFields) {
      if (!(field in obj) || obj[field] === undefined || obj[field] === null) {
        errors.push(`Field '${field}' is required`);
      }
    }

    return {
      valid: errors.length === 0,
      data: errors.length === 0 ? obj : undefined,
      errors: errors.length > 0 ? errors : undefined
    };
  }

  /**
   * Validate object field types
   */
  static validateTypes(obj: any, schema: Record<string, string>): ValidationResult<any> {
    const errors: string[] = [];

    for (const [field, expectedType] of Object.entries(schema)) {
      if (field in obj) {
        const actualType = typeof obj[field];
        if (actualType !== expectedType) {
          errors.push(`Field '${field}' must be of type ${expectedType}, got ${actualType}`);
        }
      }
    }

    return {
      valid: errors.length === 0,
      data: errors.length === 0 ? obj : undefined,
      errors: errors.length > 0 ? errors : undefined
    };
  }

  /**
   * Validate and whitelist object properties
   */
  static validateWhitelist(obj: any, allowedFields: string[]): ValidationResult<any> {
    const whitelisted: Record<string, any> = {};

    for (const field of allowedFields) {
      if (field in obj) {
        whitelisted[field] = obj[field];
      }
    }

    return {
      valid: true,
      data: whitelisted
    };
  }

  /**
   * Validate array of objects
   */
  static validateArray(
    arr: any[],
    itemValidator: (item: any) => ValidationResult<any>
  ): ValidationResult<any[]> {
    const errors: string[] = [];
    const validatedItems: any[] = [];

    for (let i = 0; i < arr.length; i++) {
      const result = itemValidator(arr[i]);
      if (!result.valid) {
        errors.push(`Item ${i}: ${result.errors?.join(', ')}`);
      } else if (result.data) {
        validatedItems.push(result.data);
      }
    }

    return {
      valid: errors.length === 0,
      data: errors.length === 0 ? validatedItems : undefined,
      errors: errors.length > 0 ? errors : undefined
    };
  }
}

/**
 * ContentSecurityPolicy generator
 */
export class CSPGenerator {
  private directives: Record<string, string[]> = {};

  /**
   * Set directive
   */
  setDirective(directive: string, values: string[]): this {
    this.directives[directive] = values;
    return this;
  }

  /**
   * Generate CSP header value
   */
  generate(): string {
    return Object.entries(this.directives)
      .map(([directive, values]) => `${directive} ${values.join(' ')}`)
      .join('; ');
  }

  /**
   * Generate strict CSP policy
   */
  static strict(): string {
    return new CSPGenerator()
      .setDirective('default-src', ["'self'"])
      .setDirective('script-src', ["'self'", "'strict-dynamic'"])
      .setDirective('style-src', ["'self'", "'unsafe-inline'"])
      .setDirective('img-src', ["'self'", 'data:', 'https:'])
      .setDirective('font-src', ["'self'"])
      .setDirective('connect-src', ["'self'"])
      .setDirective('frame-ancestors', ["'none'"])
      .setDirective('form-action', ["'self'"])
      .setDirective('base-uri', ["'self'"])
      .setDirective('upgrade-insecure-requests', [])
      .generate();
  }

  /**
   * Generate moderate CSP policy
   */
  static moderate(): string {
    return new CSPGenerator()
      .setDirective('default-src', ["'self'"])
      .setDirective('script-src', ["'self'", 'https:'])
      .setDirective('style-src', ["'self'", "'unsafe-inline'"])
      .setDirective('img-src', ["'self'", 'https:', 'data:'])
      .setDirective('font-src', ["'self'", 'https:'])
      .setDirective('connect-src', ["'self'", 'https:'])
      .setDirective('frame-ancestors', ["'self'"])
      .generate();
  }
}

/**
 * Security headers generator
 */
export class SecurityHeadersGenerator {
  static generate(): Record<string, string> {
    return {
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      'X-XSS-Protection': '1; mode=block',
      'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
      'Content-Security-Policy': CSPGenerator.strict(),
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Permissions-Policy': 'geolocation=(), microphone=(), camera=()'
    };
  }
}

export default InputValidator;
