/**
 * Security Manager - Unified Security Orchestration
 * Phase 12: Enterprise-Grade Security
 *
 * Integrates all security modules:
 * - Encryption
 * - Rate limiting
 * - Input validation
 * - CSRF protection
 * - Audit logging
 * - Compliance monitoring
 */

import {
  EncryptionService,
  SecureKeyStore,
  TokenGenerator,
  KeyManager
} from './encryption';

import {
  TokenBucketRateLimiter,
  AdaptiveRateLimiter,
  DDoSDetector,
  CircuitBreaker,
  RateLimiterManager
} from './rateLimiter';

import {
  InputValidator,
  InputSanitizer,
  SchemaValidator,
  SecurityHeadersGenerator
} from './validation';

import {
  CSRFProtection,
  OriginValidator,
  SameSiteCookieManager
} from './csrf';

import {
  AuditLogger,
  ComplianceLogger,
  AuditSeverity,
  AuditCategory
} from './auditLogger';

/**
 * Security configuration
 */
export interface SecurityConfig {
  // Encryption
  encryptionEnabled?: boolean;

  // Rate limiting
  rateLimitingEnabled?: boolean;
  defaultRateLimit?: number; // requests per minute
  apiRateLimit?: number;
  webSocketRateLimit?: number;

  // CSRF
  csrfEnabled?: boolean;
  allowedOrigins?: string[];

  // Audit logging
  auditLoggingEnabled?: boolean;
  maxAuditLogs?: number;

  // DDoS detection
  ddosDetectionEnabled?: boolean;
  ddosThreshold?: number;
}

/**
 * SecurityManager - Unified security orchestration
 */
export class SecurityManager {
  private encryption: EncryptionService;
  private keyStore: SecureKeyStore;
  private tokenGenerator: TokenGenerator;
  private keyManager: KeyManager;

  private rateLimiterManager: RateLimiterManager;
  private adaptiveRateLimiter: AdaptiveRateLimiter;
  private ddosDetector: DDoSDetector;
  private circuitBreaker: CircuitBreaker;

  private inputValidator = InputValidator;
  private inputSanitizer = InputSanitizer;
  private schemaValidator = SchemaValidator;

  private csrf: CSRFProtection;
  private originValidator: OriginValidator;
  private sameSiteCookies: SameSiteCookieManager;

  private auditLogger: AuditLogger;
  private complianceLogger: ComplianceLogger;

  private config: Required<SecurityConfig>;

  private constructor(
    encryption: EncryptionService,
    config: SecurityConfig = {}
  ) {
    this.encryption = encryption;
    this.keyStore = new SecureKeyStore(encryption);
    this.tokenGenerator = new TokenGenerator(encryption);
    this.keyManager = new KeyManager();

    this.rateLimiterManager = new RateLimiterManager({
      windowMs: 60000, // 1 minute
      maxRequests: config.defaultRateLimit || 100
    });
    this.adaptiveRateLimiter = new AdaptiveRateLimiter({
      windowMs: 60000,
      maxRequests: config.defaultRateLimit || 100
    });
    this.ddosDetector = new DDoSDetector(config.ddosThreshold || 100);
    this.circuitBreaker = new CircuitBreaker();

    this.csrf = new CSRFProtection();
    this.originValidator = new OriginValidator(config.allowedOrigins || []);
    this.sameSiteCookies = new SameSiteCookieManager();

    this.auditLogger = new AuditLogger(config.maxAuditLogs || 10000);
    this.complianceLogger = new ComplianceLogger(this.auditLogger);

    this.config = {
      encryptionEnabled: config.encryptionEnabled ?? true,
      rateLimitingEnabled: config.rateLimitingEnabled ?? true,
      defaultRateLimit: config.defaultRateLimit ?? 100,
      apiRateLimit: config.apiRateLimit ?? 1000,
      webSocketRateLimit: config.webSocketRateLimit ?? 500,
      csrfEnabled: config.csrfEnabled ?? true,
      allowedOrigins: config.allowedOrigins ?? [],
      auditLoggingEnabled: config.auditLoggingEnabled ?? true,
      maxAuditLogs: config.maxAuditLogs ?? 10000,
      ddosDetectionEnabled: config.ddosDetectionEnabled ?? true,
      ddosThreshold: config.ddosThreshold ?? 100
    };
  }

  /**
   * Create and initialize SecurityManager
   */
  static async create(config?: SecurityConfig): Promise<SecurityManager> {
    const encryption = new EncryptionService();
    await encryption.initialize();
    return new SecurityManager(encryption, config);
  }

  // ============================================================================
  // Encryption & Key Management
  // ============================================================================

  /**
   * Encrypt sensitive data
   */
  encryptData(plaintext: string): string {
    if (!this.config.encryptionEnabled) {
      return plaintext;
    }

    const encrypted = this.encryption.encrypt(plaintext);
    return JSON.stringify(encrypted);
  }

  /**
   * Decrypt sensitive data
   */
  decryptData(encryptedString: string): string | null {
    if (!this.config.encryptionEnabled) {
      return encryptedString;
    }

    try {
      const encrypted = JSON.parse(encryptedString);
      return this.encryption.decrypt(encrypted);
    } catch (error) {
      return null;
    }
  }

  /**
   * Store API key securely
   */
  async storeApiKey(name: string, value: string, expiresInDays: number = 365): Promise<void> {
    return this.keyStore.storeKey(name, value, expiresInDays);
  }

  /**
   * Retrieve API key
   */
  async retrieveApiKey(name: string): Promise<string | null> {
    return this.keyStore.retrieveKey(name);
  }

  /**
   * Generate session token
   */
  generateSessionToken(userId: string, expiresInMinutes: number = 60): string {
    return this.tokenGenerator.generateSessionToken(userId, expiresInMinutes);
  }

  /**
   * Generate API token
   */
  generateApiToken(clientId: string, scope: string[]): { token: string; expires: number } {
    return this.tokenGenerator.generateApiToken(clientId, scope);
  }

  // ============================================================================
  // Rate Limiting
  // ============================================================================

  /**
   * Check if request is rate limited
   */
  checkRateLimit(endpoint: string, userId: string): { allowed: boolean; remaining: number; resetTime: number } {
    if (!this.config.rateLimitingEnabled) {
      return { allowed: true, remaining: 999, resetTime: 0 };
    }

    const status = this.rateLimiterManager.isAllowed(endpoint, userId);

    return {
      allowed: !status.isLimited,
      remaining: status.remaining,
      resetTime: status.resetTime
    };
  }

  /**
   * Check if IP is under DDoS attack
   */
  checkDDoS(ip: string): boolean {
    if (!this.config.ddosDetectionEnabled) {
      return false;
    }

    return this.ddosDetector.detect(ip) || this.ddosDetector.isSuspicious(ip);
  }

  /**
   * Get rate limit status
   */
  getRateLimitStatus(endpoint: string, userId: string) {
    const limiter = this.rateLimiterManager.getLimiter(endpoint);
    if (limiter) {
      return limiter.getStatus(userId);
    }
    return null;
  }

  // ============================================================================
  // Input Validation & Sanitization
  // ============================================================================

  /**
   * Validate and sanitize input
   */
  validateAndSanitize(input: string, options: {
    type?: 'email' | 'url' | 'password' | 'text' | 'path';
    stripHtml?: boolean;
    maxLength?: number;
  } = {}): { valid: boolean; value: string | null; errors?: string[] } {
    let validated: { valid: boolean; data?: any; errors?: string[] };

    switch (options.type) {
      case 'email':
        validated = InputValidator.validateEmail(input);
        break;
      case 'url':
        validated = InputValidator.validateUrl(input);
        break;
      case 'password':
        validated = InputValidator.validatePassword(input);
        break;
      case 'path':
        const sanitized = InputSanitizer.sanitizePath(input);
        validated = { valid: true, data: sanitized };
        break;
      default:
        validated = { valid: true, data: input };
    }

    if (!validated.valid) {
      return {
        valid: false,
        value: null,
        errors: validated.errors
      };
    }

    let value = validated.data || input;

    if (options.stripHtml) {
      value = InputSanitizer.stripHtmlTags(value);
    }

    if (options.maxLength) {
      value = InputSanitizer.trimAndLimit(value, options.maxLength);
    }

    return {
      valid: true,
      value
    };
  }

  /**
   * Sanitize HTML to prevent XSS
   */
  sanitizeHtml(html: string): string {
    return InputSanitizer.escapeHtml(html);
  }

  /**
   * Get security headers
   */
  getSecurityHeaders(): Record<string, string> {
    return SecurityHeadersGenerator.generate();
  }

  // ============================================================================
  // CSRF Protection
  // ============================================================================

  /**
   * Generate CSRF token for session
   */
  generateCsrfToken(sessionId: string): string {
    if (!this.config.csrfEnabled) {
      return '';
    }

    return this.csrf.generateToken(sessionId);
  }

  /**
   * Verify CSRF token
   */
  verifyCsrfToken(sessionId: string, token: string, secret: string): boolean {
    if (!this.config.csrfEnabled) {
      return true;
    }

    return this.csrf.verifyToken(sessionId, token, secret);
  }

  /**
   * Validate request origin
   */
  validateOrigin(origin: string | undefined): boolean {
    return this.originValidator.validateOrigin(origin);
  }

  /**
   * Validate request referer
   */
  validateReferer(referer: string | undefined): boolean {
    return this.originValidator.validateReferer(referer);
  }

  // ============================================================================
  // Audit Logging
  // ============================================================================

  /**
   * Log security event
   */
  logSecurityEvent(
    action: string,
    status: 'success' | 'failure',
    severity: AuditSeverity,
    details?: Record<string, any>,
    userId?: string
  ): string {
    if (!this.config.auditLoggingEnabled) {
      return '';
    }

    return this.auditLogger.logSecurityEvent(action, status, severity, details, userId);
  }

  /**
   * Log authentication event
   */
  logAuthEvent(
    action: string,
    status: 'success' | 'failure',
    userId?: string,
    details?: Record<string, any>
  ): string {
    if (!this.config.auditLoggingEnabled) {
      return '';
    }

    return this.auditLogger.logAuth(action, status, userId, details);
  }

  /**
   * Log data access
   */
  logDataAccess(resource: string, resourceId: string, userId?: string): string {
    if (!this.config.auditLoggingEnabled) {
      return '';
    }

    return this.auditLogger.logDataAccess(resource, resourceId, userId);
  }

  /**
   * Log data modification
   */
  logDataModification(
    action: 'create' | 'update' | 'delete',
    resource: string,
    resourceId: string,
    userId?: string,
    changes?: Record<string, any>
  ): string {
    if (!this.config.auditLoggingEnabled) {
      return '';
    }

    return this.auditLogger.logDataModification(action, resource, resourceId, userId, changes);
  }

  /**
   * Get audit logs
   */
  getAuditLogs(options?: any) {
    return this.auditLogger.query(options);
  }

  /**
   * Get security incidents
   */
  getSecurityIncidents(limit?: number) {
    return this.auditLogger.getIncidents(limit);
  }

  /**
   * Get audit statistics
   */
  getAuditStatistics(timeRangeMs?: number) {
    return this.auditLogger.getStatistics(timeRangeMs);
  }

  // ============================================================================
  // Compliance
  // ============================================================================

  /**
   * Log GDPR consent
   */
  logConsent(userId: string, consentType: string, granted: boolean): string {
    return this.complianceLogger.logConsent(userId, consentType, granted);
  }

  /**
   * Log GDPR data deletion
   */
  logDataDeletion(userId: string, dataTypes: string[]): string {
    return this.complianceLogger.logDataDeletion(userId, dataTypes);
  }

  /**
   * Log data breach
   */
  logDataBreach(description: string, affectedUsers: number): string {
    return this.complianceLogger.logDataBreach(description, affectedUsers);
  }

  // ============================================================================
  // Getters
  // ============================================================================

  /**
   * Get encryption service
   */
  getEncryption(): EncryptionService {
    return this.encryption;
  }

  /**
   * Get audit logger
   */
  getAuditLogger(): AuditLogger {
    return this.auditLogger;
  }

  /**
   * Get compliance logger
   */
  getComplianceLogger(): ComplianceLogger {
    return this.complianceLogger;
  }

  /**
   * Get circuit breaker
   */
  getCircuitBreaker(): CircuitBreaker {
    return this.circuitBreaker;
  }

  /**
   * Get configuration
   */
  getConfig(): Required<SecurityConfig> {
    return this.config;
  }
}

// Singleton instance
let securityManagerInstance: SecurityManager | null = null;

/**
 * Get or create security manager instance
 */
export async function getSecurityManager(config?: SecurityConfig): Promise<SecurityManager> {
  if (!securityManagerInstance) {
    securityManagerInstance = await SecurityManager.create(config);
  }
  return securityManagerInstance;
}

export default SecurityManager;
