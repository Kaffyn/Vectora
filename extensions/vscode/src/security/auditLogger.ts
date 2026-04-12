/**
 * Audit Logging Module
 * Phase 12: Enterprise-Grade Security
 *
 * Provides:
 * - Comprehensive security event logging
 * - Activity tracking
 * - Compliance logging (GDPR, SOC 2)
 * - Log rotation and archiving
 * - Log querying and analysis
 */

/**
 * Audit event severity levels
 */
export enum AuditSeverity {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical'
}

/**
 * Audit event categories
 */
export enum AuditCategory {
  AUTHENTICATION = 'authentication',
  AUTHORIZATION = 'authorization',
  DATA_ACCESS = 'data_access',
  DATA_MODIFICATION = 'data_modification',
  SECURITY = 'security',
  CONFIGURATION = 'configuration',
  SYSTEM = 'system'
}

/**
 * Audit log entry
 */
export interface AuditLogEntry {
  id: string;
  timestamp: number;
  severity: AuditSeverity;
  category: AuditCategory;
  action: string;
  userId?: string;
  sessionId?: string;
  ipAddress?: string;
  userAgent?: string;
  resource?: string;
  resourceId?: string;
  status: 'success' | 'failure';
  statusCode?: number;
  details?: Record<string, any>;
  error?: string;
}

/**
 * Audit logger query options
 */
export interface AuditLogQuery {
  startTime?: number;
  endTime?: number;
  userId?: string;
  category?: AuditCategory;
  severity?: AuditSeverity;
  action?: string;
  status?: 'success' | 'failure';
  limit?: number;
  offset?: number;
}

/**
 * AuditLogger - Core audit logging
 */
export class AuditLogger {
  private logs: AuditLogEntry[] = [];
  private maxLogs: number;
  private logFile?: string;
  private eventHandlers: Map<string, Function[]> = new Map();

  constructor(maxLogs: number = 10000, logFile?: string) {
    this.maxLogs = maxLogs;
    this.logFile = logFile;

    // Cleanup old logs periodically
    setInterval(() => this.cleanup(), 60 * 60 * 1000); // Every hour
  }

  /**
   * Log security event
   */
  log(entry: Omit<AuditLogEntry, 'id' | 'timestamp'>): string {
    const logEntry: AuditLogEntry = {
      ...entry,
      id: this.generateId(),
      timestamp: Date.now()
    };

    this.logs.push(logEntry);

    // Keep logs under max size
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }

    // Emit event for real-time handlers
    this.emit(`event:${entry.category}`, logEntry);
    if (entry.severity === AuditSeverity.CRITICAL) {
      this.emit('event:critical', logEntry);
    }

    return logEntry.id;
  }

  /**
   * Log authentication event
   */
  logAuth(
    action: string,
    status: 'success' | 'failure',
    userId?: string,
    details?: Record<string, any>
  ): string {
    return this.log({
      category: AuditCategory.AUTHENTICATION,
      action,
      status,
      userId,
      details,
      severity: status === 'success' ? AuditSeverity.INFO : AuditSeverity.WARNING
    });
  }

  /**
   * Log authorization event
   */
  logAuthz(
    action: string,
    resource: string,
    status: 'success' | 'failure',
    userId?: string,
    details?: Record<string, any>
  ): string {
    return this.log({
      category: AuditCategory.AUTHORIZATION,
      action,
      resource,
      status,
      userId,
      details,
      severity: status === 'success' ? AuditSeverity.INFO : AuditSeverity.WARNING
    });
  }

  /**
   * Log data access event
   */
  logDataAccess(
    resource: string,
    resourceId: string,
    userId?: string,
    details?: Record<string, any>
  ): string {
    return this.log({
      category: AuditCategory.DATA_ACCESS,
      action: 'read',
      resource,
      resourceId,
      status: 'success',
      userId,
      details,
      severity: AuditSeverity.INFO
    });
  }

  /**
   * Log data modification event
   */
  logDataModification(
    action: 'create' | 'update' | 'delete',
    resource: string,
    resourceId: string,
    userId?: string,
    changes?: Record<string, any>,
    details?: Record<string, any>
  ): string {
    return this.log({
      category: AuditCategory.DATA_MODIFICATION,
      action,
      resource,
      resourceId,
      status: 'success',
      userId,
      details: { ...details, changes },
      severity: action === 'delete' ? AuditSeverity.WARNING : AuditSeverity.INFO
    });
  }

  /**
   * Log security event
   */
  logSecurityEvent(
    action: string,
    status: 'success' | 'failure',
    severity: AuditSeverity,
    details?: Record<string, any>,
    userId?: string,
    error?: string
  ): string {
    return this.log({
      category: AuditCategory.SECURITY,
      action,
      status,
      severity,
      details,
      userId,
      error
    });
  }

  /**
   * Query audit logs
   */
  query(options: AuditLogQuery = {}): AuditLogEntry[] {
    let filtered = this.logs;

    if (options.startTime) {
      filtered = filtered.filter(l => l.timestamp >= options.startTime!);
    }

    if (options.endTime) {
      filtered = filtered.filter(l => l.timestamp <= options.endTime!);
    }

    if (options.userId) {
      filtered = filtered.filter(l => l.userId === options.userId);
    }

    if (options.category) {
      filtered = filtered.filter(l => l.category === options.category);
    }

    if (options.severity) {
      filtered = filtered.filter(l => l.severity === options.severity);
    }

    if (options.action) {
      filtered = filtered.filter(l => l.action === options.action);
    }

    if (options.status) {
      filtered = filtered.filter(l => l.status === options.status);
    }

    // Sort by timestamp descending
    filtered = filtered.sort((a, b) => b.timestamp - a.timestamp);

    // Apply pagination
    const offset = options.offset || 0;
    const limit = options.limit || 100;

    return filtered.slice(offset, offset + limit);
  }

  /**
   * Get log entry by ID
   */
  getLog(id: string): AuditLogEntry | null {
    return this.logs.find(l => l.id === id) || null;
  }

  /**
   * Get logs for user
   */
  getUserLogs(userId: string, limit: number = 100): AuditLogEntry[] {
    return this.logs
      .filter(l => l.userId === userId)
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, limit);
  }

  /**
   * Get security incidents
   */
  getIncidents(limit: number = 100): AuditLogEntry[] {
    return this.logs
      .filter(l =>
        l.severity === AuditSeverity.CRITICAL ||
        (l.status === 'failure' && l.category === AuditCategory.SECURITY)
      )
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, limit);
  }

  /**
   * Get statistics
   */
  getStatistics(timeRangeMs: number = 24 * 60 * 60 * 1000): Record<string, any> {
    const now = Date.now();
    const startTime = now - timeRangeMs;

    const recentLogs = this.logs.filter(l => l.timestamp >= startTime);

    const stats = {
      totalEvents: recentLogs.length,
      byCategory: {} as Record<string, number>,
      bySeverity: {} as Record<string, number>,
      byStatus: { success: 0, failure: 0 },
      failureRate: 0
    };

    for (const log of recentLogs) {
      stats.byCategory[log.category] = (stats.byCategory[log.category] || 0) + 1;
      stats.bySeverity[log.severity] = (stats.bySeverity[log.severity] || 0) + 1;
      stats.byStatus[log.status]++;
    }

    stats.failureRate = recentLogs.length > 0
      ? (stats.byStatus.failure / recentLogs.length) * 100
      : 0;

    return stats;
  }

  /**
   * Export logs as JSON
   */
  exportJSON(options: AuditLogQuery = {}): string {
    const logs = this.query(options);
    return JSON.stringify(logs, null, 2);
  }

  /**
   * Export logs as CSV
   */
  exportCSV(options: AuditLogQuery = {}): string {
    const logs = this.query(options);

    if (logs.length === 0) {
      return '';
    }

    const headers = Object.keys(logs[0]).join(',');
    const rows = logs.map(log =>
      Object.values(log).map(v =>
        typeof v === 'string' ? `"${v.replace(/"/g, '""')}"` : v
      ).join(',')
    );

    return [headers, ...rows].join('\n');
  }

  /**
   * Register event handler
   */
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  /**
   * Unregister event handler
   */
  off(event: string, handler: Function): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Emit event
   */
  private emit(event: string, data: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in event handler for ${event}:`, error);
        }
      }
    }
  }

  /**
   * Cleanup old logs
   */
  private cleanup(): void {
    const oneWeekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000);
    this.logs = this.logs.filter(l => l.timestamp > oneWeekAgo);
  }

  /**
   * Generate unique ID
   */
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Clear all logs
   */
  clear(): void {
    this.logs = [];
  }

  /**
   * Get log count
   */
  getCount(): number {
    return this.logs.length;
  }
}

/**
 * ComplianceLogger - GDPR and SOC 2 compliance logging
 */
export class ComplianceLogger {
  private logger: AuditLogger;

  constructor(logger: AuditLogger) {
    this.logger = logger;
  }

  /**
   * Log data processing activity (GDPR)
   */
  logDataProcessing(
    purpose: string,
    dataType: string,
    userId?: string,
    scope?: string[]
  ): string {
    return this.logger.log({
      category: AuditCategory.DATA_ACCESS,
      action: 'data_processing',
      status: 'success',
      severity: AuditSeverity.INFO,
      userId,
      details: {
        purpose,
        dataType,
        scope
      }
    });
  }

  /**
   * Log consent event
   */
  logConsent(
    userId: string,
    consentType: string,
    granted: boolean,
    timestamp?: number
  ): string {
    return this.logger.log({
      category: AuditCategory.SECURITY,
      action: 'consent',
      status: 'success',
      severity: AuditSeverity.INFO,
      userId,
      details: {
        consentType,
        granted,
        timestamp: timestamp || Date.now()
      }
    });
  }

  /**
   * Log data deletion (right to be forgotten)
   */
  logDataDeletion(userId: string, dataTypes: string[]): string {
    return this.logger.log({
      category: AuditCategory.DATA_MODIFICATION,
      action: 'delete',
      status: 'success',
      severity: AuditSeverity.WARNING,
      userId,
      details: {
        purpose: 'gdpr_deletion',
        dataTypes
      }
    });
  }

  /**
   * Log access request (GDPR subject access request)
   */
  logAccessRequest(userId: string, requestId: string): string {
    return this.logger.log({
      category: AuditCategory.DATA_ACCESS,
      action: 'subject_access_request',
      status: 'success',
      severity: AuditSeverity.WARNING,
      userId,
      details: {
        requestId,
        timestamp: Date.now()
      }
    });
  }

  /**
   * Log data breach notification
   */
  logDataBreach(description: string, affectedUsers: number): string {
    return this.logger.log({
      category: AuditCategory.SECURITY,
      action: 'data_breach',
      status: 'failure',
      severity: AuditSeverity.CRITICAL,
      details: {
        description,
        affectedUsers,
        timestamp: Date.now()
      }
    });
  }
}

export default AuditLogger;
