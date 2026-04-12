/**
 * Security Module - Index (Barrel Export)
 * Phase 12: Enterprise-Grade Security
 *
 * Exports all security utilities and services
 */

// Encryption & Key Management
export {
  EncryptionService,
  SecureKeyStore,
  TokenGenerator,
  KeyManager,
  getEncryptionService,
  type EncryptionConfig,
  type EncryptedData,
  type KeyMetadata
} from './encryption';

// Rate Limiting
export {
  TokenBucketRateLimiter,
  SlidingWindowRateLimiter,
  AdaptiveRateLimiter,
  DDoSDetector,
  CircuitBreaker,
  RateLimiterManager,
  type RateLimitConfig,
  type RateLimitStatus
} from './rateLimiter';

// Input Validation & Sanitization
export {
  InputValidator,
  InputSanitizer,
  SchemaValidator,
  CSPGenerator,
  SecurityHeadersGenerator,
  type ValidationResult,
  type SanitizationOptions
} from './validation';

// CSRF Protection
export {
  CSRFProtection,
  DoubleSubmitCookie,
  OriginValidator,
  SameSiteCookieManager,
  CSRFMiddlewareFactory,
  type CsrfToken,
  type CsrfOptions
} from './csrf';

// Audit Logging
export {
  AuditLogger,
  ComplianceLogger,
  AuditSeverity,
  AuditCategory,
  type AuditLogEntry,
  type AuditLogQuery
} from './auditLogger';

// Convenience function to initialize security module
export async function initializeSecurityModule() {
  const encryption = await EncryptionService.prototype.constructor.prototype.getEncryptionService?.() ||
    new EncryptionService();
  await encryption.initialize();

  return {
    encryption,
    keyStore: new SecureKeyStore(encryption),
    tokenGenerator: new TokenGenerator(encryption),
    keyManager: new KeyManager(),
    csrf: new CSRFProtection(),
    auditLogger: new AuditLogger()
  };
}

export default {
  EncryptionService,
  SecureKeyStore,
  TokenGenerator,
  KeyManager,
  TokenBucketRateLimiter,
  SlidingWindowRateLimiter,
  AdaptiveRateLimiter,
  DDoSDetector,
  CircuitBreaker,
  RateLimiterManager,
  InputValidator,
  InputSanitizer,
  SchemaValidator,
  CSPGenerator,
  SecurityHeadersGenerator,
  CSRFProtection,
  DoubleSubmitCookie,
  OriginValidator,
  SameSiteCookieManager,
  CSRFMiddlewareFactory,
  AuditLogger,
  ComplianceLogger,
  AuditSeverity,
  AuditCategory
};
