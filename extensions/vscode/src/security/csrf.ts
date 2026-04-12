/**
 * CSRF Protection Module
 * Phase 12: Enterprise-Grade Security
 *
 * Provides:
 * - CSRF token generation and validation
 * - Double submit cookie pattern
 * - SameSite cookie support
 * - Origin/Referer verification
 * - Token storage and rotation
 */

/**
 * CSRF token metadata
 */
export interface CsrfToken {
  token: string;
  secret: string;
  createdAt: number;
  expiresAt: number;
}

/**
 * CSRF validation options
 */
export interface CsrfOptions {
  headerName?: string;
  tokenLength?: number;
  expirationMinutes?: number;
}

/**
 * CSRFProtection - Core CSRF protection logic
 */
export class CSRFProtection {
  private tokens: Map<string, CsrfToken> = new Map();
  private options: Required<CsrfOptions>;

  constructor(options?: CsrfOptions) {
    this.options = {
      headerName: options?.headerName || 'x-csrf-token',
      tokenLength: options?.tokenLength || 32,
      expirationMinutes: options?.expirationMinutes || 60
    };
  }

  /**
   * Generate new CSRF token
   */
  generateToken(sessionId: string): string {
    const token = this.generateRandomToken();
    const secret = this.generateRandomToken();

    const csrfToken: CsrfToken = {
      token,
      secret,
      createdAt: Date.now(),
      expiresAt: Date.now() + (this.options.expirationMinutes * 60 * 1000)
    };

    this.tokens.set(sessionId, csrfToken);
    return token;
  }

  /**
   * Verify CSRF token
   */
  verifyToken(sessionId: string, token: string, secret: string): boolean {
    const stored = this.tokens.get(sessionId);

    if (!stored) {
      return false;
    }

    // Check if token has expired
    if (stored.expiresAt < Date.now()) {
      this.tokens.delete(sessionId);
      return false;
    }

    // Use timing-safe comparison
    return this.timingSafeEqual(token, stored.token) &&
           this.timingSafeEqual(secret, stored.secret);
  }

  /**
   * Refresh CSRF token
   */
  refreshToken(sessionId: string): string | null {
    const stored = this.tokens.get(sessionId);

    if (!stored) {
      return null;
    }

    // Delete old token and generate new one
    this.tokens.delete(sessionId);
    return this.generateToken(sessionId);
  }

  /**
   * Invalidate token
   */
  invalidateToken(sessionId: string): void {
    this.tokens.delete(sessionId);
  }

  /**
   * Cleanup expired tokens
   */
  cleanup(): void {
    const now = Date.now();
    for (const [sessionId, token] of this.tokens.entries()) {
      if (token.expiresAt < now) {
        this.tokens.delete(sessionId);
      }
    }
  }

  /**
   * Generate random token
   */
  private generateRandomToken(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let token = '';
    for (let i = 0; i < this.options.tokenLength; i++) {
      token += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return token;
  }

  /**
   * Timing-safe string comparison
   */
  private timingSafeEqual(a: string, b: string): boolean {
    if (a.length !== b.length) {
      return false;
    }

    let result = 0;
    for (let i = 0; i < a.length; i++) {
      result |= a.charCodeAt(i) ^ b.charCodeAt(i);
    }

    return result === 0;
  }

  /**
   * Get header name for CSRF token
   */
  getHeaderName(): string {
    return this.options.headerName;
  }
}

/**
 * Double Submit Cookie pattern
 */
export class DoubleSubmitCookie {
  private cookieName: string;
  private secure: boolean;
  private httpOnly: boolean;
  private sameSite: 'Strict' | 'Lax' | 'None';

  constructor(
    cookieName: string = '__CSRF-TOKEN',
    secure: boolean = true,
    sameSite: 'Strict' | 'Lax' | 'None' = 'Lax'
  ) {
    this.cookieName = cookieName;
    this.secure = secure;
    this.httpOnly = false; // Must be false to allow JS access
    this.sameSite = sameSite;
  }

  /**
   * Generate cookie value
   */
  generateCookie(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let value = '';
    for (let i = 0; i < 32; i++) {
      value += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return value;
  }

  /**
   * Get cookie header value
   */
  getCookieHeader(value: string, maxAgeSeconds: number = 3600): string {
    const parts = [
      `${this.cookieName}=${value}`,
      `Max-Age=${maxAgeSeconds}`,
      this.secure ? 'Secure' : '',
      `SameSite=${this.sameSite}`,
      'Path=/'
    ];

    return parts.filter(p => p).join('; ');
  }

  /**
   * Verify double submit pattern
   */
  verify(cookieValue: string, headerValue: string): boolean {
    if (!cookieValue || !headerValue) {
      return false;
    }

    // Values must match exactly
    return cookieValue === headerValue;
  }

  /**
   * Get cookie name
   */
  getCookieName(): string {
    return this.cookieName;
  }
}

/**
 * Origin and Referer validator
 */
export class OriginValidator {
  private allowedOrigins: Set<string>;

  constructor(allowedOrigins: string[] = []) {
    this.allowedOrigins = new Set(allowedOrigins);
  }

  /**
   * Add allowed origin
   */
  addOrigin(origin: string): void {
    this.allowedOrigins.add(origin);
  }

  /**
   * Remove allowed origin
   */
  removeOrigin(origin: string): void {
    this.allowedOrigins.delete(origin);
  }

  /**
   * Validate request origin
   */
  validateOrigin(origin: string | undefined | null): boolean {
    if (!origin) {
      return false;
    }

    try {
      const url = new URL(origin);
      // Only allow http and https
      if (!['http:', 'https:'].includes(url.protocol)) {
        return false;
      }

      return this.allowedOrigins.has(origin) || this.allowedOrigins.has(url.hostname);
    } catch (error) {
      return false;
    }
  }

  /**
   * Validate Referer header
   */
  validateReferer(referer: string | undefined | null): boolean {
    if (!referer) {
      return false;
    }

    try {
      const url = new URL(referer);
      return this.allowedOrigins.has(`${url.protocol}//${url.hostname}`) ||
             this.allowedOrigins.has(url.hostname);
    } catch (error) {
      return false;
    }
  }

  /**
   * Extract origin from referer
   */
  extractOrigin(referer: string): string | null {
    try {
      const url = new URL(referer);
      return `${url.protocol}//${url.hostname}`;
    } catch (error) {
      return null;
    }
  }
}

/**
 * SameSite Cookie manager
 */
export class SameSiteCookieManager {
  private cookies: Map<string, { value: string; sameSite: 'Strict' | 'Lax' | 'None' }> = new Map();

  /**
   * Set SameSite cookie
   */
  setCookie(
    name: string,
    value: string,
    sameSite: 'Strict' | 'Lax' | 'None' = 'Lax',
    options?: {
      secure?: boolean;
      httpOnly?: boolean;
      maxAge?: number;
      domain?: string;
      path?: string;
    }
  ): string {
    this.cookies.set(name, { value, sameSite });

    const parts = [
      `${name}=${value}`,
      `SameSite=${sameSite}`
    ];

    if (options?.secure) {
      parts.push('Secure');
    }

    if (options?.httpOnly) {
      parts.push('HttpOnly');
    }

    if (options?.maxAge) {
      parts.push(`Max-Age=${options.maxAge}`);
    }

    if (options?.domain) {
      parts.push(`Domain=${options.domain}`);
    }

    if (options?.path) {
      parts.push(`Path=${options.path}`);
    }

    return parts.join('; ');
  }

  /**
   * Get cookie value
   */
  getCookie(name: string): string | null {
    return this.cookies.get(name)?.value || null;
  }

  /**
   * Check SameSite attribute
   */
  getSameSite(name: string): 'Strict' | 'Lax' | 'None' | null {
    return this.cookies.get(name)?.sameSite || null;
  }

  /**
   * Clear cookie
   */
  clearCookie(name: string): string {
    this.cookies.delete(name);
    return `${name}=; Max-Age=0; Path=/`;
  }
}

/**
 * CSRF Middleware factory
 */
export class CSRFMiddlewareFactory {
  /**
   * Create CSRF validation middleware
   */
  static createValidator(protection: CSRFProtection): (config: any) => boolean {
    return (config: any) => {
      const { sessionId, token, secret } = config;

      if (!sessionId || !token) {
        return false;
      }

      return protection.verifyToken(sessionId, token, secret || '');
    };
  }

  /**
   * Create CSRF token generator middleware
   */
  static createGenerator(protection: CSRFProtection): (sessionId: string) => string {
    return (sessionId: string) => {
      return protection.generateToken(sessionId);
    };
  }
}

export default CSRFProtection;
