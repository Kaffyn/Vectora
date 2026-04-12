/**
 * Security Module Tests
 * Phase 12: Enterprise-Grade Security
 */

import {
  EncryptionService,
  SecureKeyStore,
  TokenGenerator,
  KeyManager
} from '../../security/encryption';

import {
  TokenBucketRateLimiter,
  SlidingWindowRateLimiter,
  AdaptiveRateLimiter,
  DDoSDetector,
  CircuitBreaker,
  RateLimiterManager
} from '../../security/rateLimiter';

import {
  InputValidator,
  InputSanitizer,
  SchemaValidator,
  CSPGenerator,
  SecurityHeadersGenerator
} from '../../security/validation';

import {
  CSRFProtection,
  DoubleSubmitCookie,
  OriginValidator,
  SameSiteCookieManager
} from '../../security/csrf';

import {
  AuditLogger,
  AuditCategory,
  AuditSeverity,
  ComplianceLogger
} from '../../security/auditLogger';

import {
  SecurityManager,
  getSecurityManager
} from '../../security/securityManager';

describe('Security Module Tests', () => {
  // ============================================================================
  // Encryption Tests
  // ============================================================================

  describe('EncryptionService', () => {
    let encryption: EncryptionService;

    beforeEach(async () => {
      encryption = new EncryptionService();
      await encryption.initialize();
    });

    test('should encrypt and decrypt data', () => {
      const plaintext = 'Hello, World!';
      const encrypted = encryption.encrypt(plaintext);

      expect(encrypted.ciphertext).toBeDefined();
      expect(encrypted.iv).toBeDefined();
      expect(encrypted.authTag).toBeDefined();

      const decrypted = encryption.decrypt(encrypted);
      expect(decrypted).toBe(plaintext);
    });

    test('should generate secure random strings', () => {
      const random1 = encryption.generateRandomString(32);
      const random2 = encryption.generateRandomString(32);

      expect(random1).toHaveLength(64); // 32 bytes * 2 hex chars
      expect(random2).toHaveLength(64);
      expect(random1).not.toBe(random2);
    });

    test('should generate and verify HMAC', () => {
      const data = 'sensitive data';
      const hmac = encryption.generateHmac(data);

      expect(encryption.verifyHmac(data, hmac)).toBe(true);
      expect(encryption.verifyHmac('modified data', hmac)).toBe(false);
    });

    test('should hash data', () => {
      const data = 'password';
      const hash = encryption.hash(data);

      expect(hash).toBeDefined();
      expect(hash).toHaveLength(64); // SHA256 hex
      expect(encryption.hash(data)).toBe(hash); // Deterministic
    });
  });

  describe('SecureKeyStore', () => {
    let encryption: EncryptionService;
    let keyStore: SecureKeyStore;

    beforeEach(async () => {
      encryption = new EncryptionService();
      await encryption.initialize();
      keyStore = new SecureKeyStore(encryption);
    });

    test('should store and retrieve keys', async () => {
      const apiKey = 'sk_test_123456789';
      await keyStore.storeKey('test_api_key', apiKey);

      const retrieved = await keyStore.retrieveKey('test_api_key');
      expect(retrieved).toBe(apiKey);
    });

    test('should revoke keys', async () => {
      await keyStore.storeKey('test_key', 'secret_value');
      expect(keyStore.revokeKey('test_key')).toBe(true);

      const retrieved = await keyStore.retrieveKey('test_key');
      expect(retrieved).toBeNull();
    });

    test('should rotate keys', async () => {
      await keyStore.storeKey('rotating_key', 'old_secret');
      const rotated = await keyStore.rotateKey('rotating_key', 'new_secret');

      expect(rotated).toBe(true);
      const retrieved = await keyStore.retrieveKey('rotating_key');
      expect(retrieved).toBe('new_secret');
    });

    test('should list keys without revealing values', async () => {
      await keyStore.storeKey('key1', 'secret1');
      await keyStore.storeKey('key2', 'secret2');

      const list = keyStore.listKeys();
      expect(list).toHaveLength(2);
      expect(list[0]).toHaveProperty('name');
      expect(list[0]).toHaveProperty('status');
      expect(list[0]).not.toHaveProperty('secret');
    });
  });

  describe('TokenGenerator', () => {
    let encryption: EncryptionService;
    let tokenGen: TokenGenerator;

    beforeEach(async () => {
      encryption = new EncryptionService();
      await encryption.initialize();
      tokenGen = new TokenGenerator(encryption);
    });

    test('should generate session tokens', () => {
      const token = tokenGen.generateSessionToken('user123');
      expect(token).toBeDefined();
      expect(token.length).toBeGreaterThan(0);
    });

    test('should generate CSRF tokens', () => {
      const token = tokenGen.generateCsrfToken();
      expect(token).toBeDefined();
      expect(token.length).toBeGreaterThan(0);
    });

    test('should generate API tokens with expiration', () => {
      const { token, expires } = tokenGen.generateApiToken('client1', ['read', 'write']);
      expect(token).toBeDefined();
      expect(expires).toBeGreaterThan(Date.now());
    });
  });

  // ============================================================================
  // Rate Limiting Tests
  // ============================================================================

  describe('TokenBucketRateLimiter', () => {
    test('should allow requests within limit', () => {
      const limiter = new TokenBucketRateLimiter({
        windowMs: 1000,
        maxRequests: 5
      });

      for (let i = 0; i < 5; i++) {
        const status = limiter.isAllowed('user1');
        expect(status.isLimited).toBe(false);
      }

      const status = limiter.isAllowed('user1');
      expect(status.isLimited).toBe(true);
    });

    test('should refill tokens over time', async () => {
      const limiter = new TokenBucketRateLimiter({
        windowMs: 100,
        maxRequests: 1
      });

      limiter.isAllowed('user1'); // Use token
      expect(limiter.isAllowed('user1').isLimited).toBe(true);

      await new Promise(resolve => setTimeout(resolve, 150));

      expect(limiter.isAllowed('user1').isLimited).toBe(false);
    });

    test('should reset limits', () => {
      const limiter = new TokenBucketRateLimiter({
        windowMs: 1000,
        maxRequests: 2
      });

      limiter.isAllowed('user1');
      limiter.isAllowed('user1');

      expect(limiter.isAllowed('user1').isLimited).toBe(true);

      limiter.reset('user1');

      expect(limiter.isAllowed('user1').isLimited).toBe(false);
    });
  });

  describe('SlidingWindowRateLimiter', () => {
    test('should enforce sliding window limits', () => {
      const limiter = new SlidingWindowRateLimiter({
        windowMs: 1000,
        maxRequests: 3
      });

      for (let i = 0; i < 3; i++) {
        expect(limiter.isAllowed('user1').isLimited).toBe(false);
      }

      expect(limiter.isAllowed('user1').isLimited).toBe(true);
    });
  });

  describe('DDoSDetector', () => {
    test('should detect potential DDoS attacks', () => {
      const detector = new DDoSDetector(10, 1000);

      for (let i = 0; i < 10; i++) {
        expect(detector.detect('192.168.1.1')).toBe(false);
      }

      expect(detector.detect('192.168.1.1')).toBe(true);
      expect(detector.isSuspicious('192.168.1.1')).toBe(true);
    });
  });

  describe('CircuitBreaker', () => {
    test('should open after failure threshold', () => {
      const breaker = new CircuitBreaker(3);

      expect(breaker.getState()).toBe('closed');

      for (let i = 0; i < 3; i++) {
        breaker.recordFailure();
      }

      expect(breaker.getState()).toBe('open');
      expect(breaker.isAllowed()).toBe(false);
    });

    test('should transition to half-open after timeout', async () => {
      const breaker = new CircuitBreaker(1, 100);

      breaker.recordFailure();
      expect(breaker.getState()).toBe('open');

      await new Promise(resolve => setTimeout(resolve, 150));

      expect(breaker.getState()).toBe('half-open');
    });
  });

  // ============================================================================
  // Input Validation Tests
  // ============================================================================

  describe('InputValidator', () => {
    test('should validate email addresses', () => {
      expect(InputValidator.validateEmail('test@example.com').valid).toBe(true);
      expect(InputValidator.validateEmail('invalid-email').valid).toBe(false);
    });

    test('should validate URLs', () => {
      expect(InputValidator.validateUrl('https://example.com').valid).toBe(true);
      expect(InputValidator.validateUrl('javascript:alert(1)').valid).toBe(false);
    });

    test('should validate password strength', () => {
      const result = InputValidator.validatePassword('WeakPass1!');
      expect(result.valid).toBe(true);

      const weak = InputValidator.validatePassword('weak');
      expect(weak.valid).toBe(false);
      expect(weak.errors).toBeDefined();
    });

    test('should validate integers', () => {
      expect(InputValidator.validateInteger('42').valid).toBe(true);
      expect(InputValidator.validateInteger('42', 0, 100).valid).toBe(true);
      expect(InputValidator.validateInteger('42', 50, 100).valid).toBe(false);
    });

    test('should validate UUID', () => {
      const uuid = '550e8400-e29b-41d4-a716-446655440000';
      expect(InputValidator.validateUuid(uuid).valid).toBe(true);
      expect(InputValidator.validateUuid('not-a-uuid').valid).toBe(false);
    });
  });

  describe('InputSanitizer', () => {
    test('should escape HTML', () => {
      const escaped = InputSanitizer.escapeHtml('<script>alert("xss")</script>');
      expect(escaped).toBe('&lt;script&gt;alert(&quot;xss&quot;)&lt;&#x2F;script&gt;');
      expect(escaped).not.toContain('<script>');
    });

    test('should strip HTML tags', () => {
      const stripped = InputSanitizer.stripHtmlTags('<p>Hello <b>World</b></p>');
      expect(stripped).toBe('Hello World');
    });

    test('should sanitize URLs', () => {
      const safe = InputSanitizer.sanitizeUrl('https://example.com');
      expect(safe).toBeDefined();

      const dangerous = InputSanitizer.sanitizeUrl('javascript:alert(1)');
      expect(dangerous).toBeNull();
    });

    test('should prevent path traversal', () => {
      const safe = InputSanitizer.sanitizePath('../../../etc/passwd');
      expect(safe).not.toContain('..');
    });
  });

  // ============================================================================
  // CSRF Protection Tests
  // ============================================================================

  describe('CSRFProtection', () => {
    let csrf: CSRFProtection;

    beforeEach(() => {
      csrf = new CSRFProtection();
    });

    test('should generate and verify CSRF tokens', () => {
      const sessionId = 'session123';
      const token = csrf.generateToken(sessionId);

      expect(token).toBeDefined();

      // For simplicity, we store both token and secret
      // In real implementation, secret would be stored separately
      expect(csrf.verifyToken(sessionId, token, token)).toBe(true);
    });

    test('should invalidate tokens', () => {
      const sessionId = 'session123';
      const token = csrf.generateToken(sessionId);

      csrf.invalidateToken(sessionId);

      expect(csrf.verifyToken(sessionId, token, token)).toBe(false);
    });
  });

  describe('OriginValidator', () => {
    test('should validate allowed origins', () => {
      const validator = new OriginValidator(['https://example.com']);

      expect(validator.validateOrigin('https://example.com')).toBe(true);
      expect(validator.validateOrigin('https://evil.com')).toBe(false);
    });

    test('should extract origin from referer', () => {
      const validator = new OriginValidator(['https://example.com']);

      const origin = validator.extractOrigin('https://example.com/page');
      expect(origin).toBe('https://example.com');
    });
  });

  // ============================================================================
  // Audit Logging Tests
  // ============================================================================

  describe('AuditLogger', () => {
    let logger: AuditLogger;

    beforeEach(() => {
      logger = new AuditLogger(1000);
    });

    test('should log security events', () => {
      const id = logger.logSecurityEvent('login_attempt', 'success', AuditSeverity.INFO);

      expect(id).toBeDefined();
      const entry = logger.getLog(id);
      expect(entry).toBeDefined();
      expect(entry?.action).toBe('login_attempt');
    });

    test('should log authentication events', () => {
      const id = logger.logAuth('user_login', 'success', 'user123');

      expect(id).toBeDefined();
      const entry = logger.getLog(id);
      expect(entry?.category).toBe(AuditCategory.AUTHENTICATION);
    });

    test('should query logs', () => {
      logger.logAuth('login', 'success', 'user1');
      logger.logAuth('logout', 'success', 'user1');
      logger.logAuth('login', 'failure', 'user2');

      const results = logger.query({ userId: 'user1' });
      expect(results.length).toBe(2);
      expect(results.every(l => l.userId === 'user1')).toBe(true);
    });

    test('should get statistics', () => {
      logger.logAuth('login', 'success', 'user1');
      logger.logAuth('login', 'failure', 'user2');

      const stats = logger.getStatistics();
      expect(stats.totalEvents).toBeGreaterThan(0);
      expect(stats.failureRate).toBeGreaterThan(0);
    });

    test('should export logs as JSON', () => {
      logger.logAuth('login', 'success', 'user1');

      const json = logger.exportJSON();
      const parsed = JSON.parse(json);
      expect(Array.isArray(parsed)).toBe(true);
    });
  });

  describe('ComplianceLogger', () => {
    let logger: AuditLogger;
    let compliance: ComplianceLogger;

    beforeEach(() => {
      logger = new AuditLogger();
      compliance = new ComplianceLogger(logger);
    });

    test('should log GDPR consent', () => {
      const id = compliance.logConsent('user1', 'marketing', true);
      expect(id).toBeDefined();
    });

    test('should log data deletion', () => {
      const id = compliance.logDataDeletion('user1', ['profile', 'history']);
      expect(id).toBeDefined();
    });

    test('should log data breach', () => {
      const id = compliance.logDataBreach('Unauthorized access detected', 1000);
      expect(id).toBeDefined();
    });
  });

  // ============================================================================
  // SecurityManager Integration Tests
  // ============================================================================

  describe('SecurityManager', () => {
    let manager: SecurityManager;

    beforeEach(async () => {
      manager = await SecurityManager.create({
        encryptionEnabled: true,
        rateLimitingEnabled: true,
        csrfEnabled: true,
        auditLoggingEnabled: true
      });
    });

    test('should encrypt and decrypt data', () => {
      const plaintext = 'sensitive information';
      const encrypted = manager.encryptData(plaintext);

      expect(encrypted).not.toBe(plaintext);

      const decrypted = manager.decryptData(encrypted);
      expect(decrypted).toBe(plaintext);
    });

    test('should enforce rate limits', () => {
      const result = manager.checkRateLimit('/api/chat', 'user1');

      expect(result).toHaveProperty('allowed');
      expect(result).toHaveProperty('remaining');
      expect(result).toHaveProperty('resetTime');
    });

    test('should validate and sanitize input', () => {
      const result = manager.validateAndSanitize('test@example.com', {
        type: 'email'
      });

      expect(result.valid).toBe(true);
    });

    test('should generate CSRF tokens', () => {
      const token = manager.generateCsrfToken('session1');
      expect(token).toBeDefined();
    });

    test('should log security events', () => {
      const id = manager.logAuthEvent('login', 'success', 'user1');
      expect(id).toBeDefined();
    });

    test('should get security headers', () => {
      const headers = manager.getSecurityHeaders();
      expect(headers['Content-Security-Policy']).toBeDefined();
      expect(headers['X-Frame-Options']).toBe('DENY');
    });
  });
});

export {};
