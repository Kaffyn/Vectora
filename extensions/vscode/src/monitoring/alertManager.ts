/**
 * Alert Manager
 * Manages and routes alerts
 */

import { v4 as uuidv4 } from 'uuid';
import {
  Alert,
  AlertRule,
  AlertSeverity,
  AlertCondition,
  AlertAction,
} from '../types/analytics';

export interface AlertListener {
  onAlert(alert: Alert): void;
}

export interface AlertChannelConfig {
  enabled: boolean;
  severityFilter?: AlertSeverity[];
  quietHours?: { start: number; end: number }; // 0-23 hours
}

export class AlertManager {
  private rules: Map<string, AlertRule> = new Map();
  private alerts: Map<string, Alert> = new Map();
  private listeners: Set<AlertListener> = new Set();
  private alertHistory: Alert[] = [];
  private cooldowns: Map<string, number> = new Map();
  private channelConfig: Map<string, AlertChannelConfig> = new Map();

  /**
   * Create an alert from a rule
   */
  createAlert(
    ruleId: string,
    severity: AlertSeverity,
    title: string,
    message: string,
  ): Alert | null {
    const rule = this.rules.get(ruleId);
    if (!rule || !rule.enabled) {
      return null;
    }

    // Check cooldown
    if (this.isInCooldown(ruleId)) {
      return null;
    }

    const alert: Alert = {
      id: uuidv4(),
      ruleId,
      severity,
      title,
      message,
      timestamp: Date.now(),
      resolved: false,
    };

    this.alerts.set(alert.id, alert);
    this.alertHistory.push(alert);

    // Trim history to last 1000 alerts
    if (this.alertHistory.length > 1000) {
      this.alertHistory = this.alertHistory.slice(-1000);
    }

    // Set cooldown
    const cooldownMs = rule.action.type === 'notification' ? 60000 : 0; // 1 minute
    this.setCooldown(ruleId, cooldownMs);

    // Notify listeners
    this.notifyListeners(alert);

    return alert;
  }

  /**
   * Resolve an alert
   */
  resolveAlert(alertId: string): void {
    const alert = this.alerts.get(alertId);
    if (alert) {
      alert.resolved = true;
      alert.resolvedAt = Date.now();
    }
  }

  /**
   * Add alert rule
   */
  addRule(rule: AlertRule): void {
    this.rules.set(rule.id, rule);
  }

  /**
   * Remove alert rule
   */
  removeRule(ruleId: string): void {
    this.rules.delete(ruleId);
  }

  /**
   * Get all rules
   */
  getRules(): AlertRule[] {
    return Array.from(this.rules.values());
  }

  /**
   * Enable rule
   */
  enableRule(ruleId: string): void {
    const rule = this.rules.get(ruleId);
    if (rule) {
      rule.enabled = true;
    }
  }

  /**
   * Disable rule
   */
  disableRule(ruleId: string): void {
    const rule = this.rules.get(ruleId);
    if (rule) {
      rule.enabled = false;
    }
  }

  /**
   * Get active alerts
   */
  getActiveAlerts(): Alert[] {
    return Array.from(this.alerts.values()).filter((a) => !a.resolved);
  }

  /**
   * Get alerts by severity
   */
  getAlertsBySeverity(severity: AlertSeverity): Alert[] {
    return Array.from(this.alerts.values()).filter((a) => a.severity === severity);
  }

  /**
   * Get alert history
   */
  getHistory(limit: number = 100): Alert[] {
    return this.alertHistory.slice(-limit);
  }

  /**
   * Clear old alerts
   */
  clearOldAlerts(beforeTime: number): void {
    for (const [id, alert] of this.alerts) {
      if (alert.timestamp < beforeTime) {
        this.alerts.delete(id);
      }
    }
  }

  /**
   * Subscribe to alerts
   */
  subscribe(listener: AlertListener): () => void {
    this.listeners.add(listener);

    // Return unsubscribe function
    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Configure alert channel
   */
  configureChannel(
    channel: string,
    config: AlertChannelConfig,
  ): void {
    this.channelConfig.set(channel, config);
  }

  /**
   * Check if channel should emit alert
   */
  shouldEmitToChannel(alert: Alert, channel: string): boolean {
    const config = this.channelConfig.get(channel);
    if (!config || !config.enabled) {
      return false;
    }

    // Check severity filter
    if (
      config.severityFilter &&
      !config.severityFilter.includes(alert.severity)
    ) {
      return false;
    }

    // Check quiet hours
    if (config.quietHours) {
      const now = new Date();
      const currentHour = now.getHours();
      const { start, end } = config.quietHours;

      if (start < end) {
        if (currentHour >= start && currentHour < end) {
          return false;
        }
      } else {
        // Quiet hours span midnight
        if (currentHour >= start || currentHour < end) {
          return false;
        }
      }
    }

    return true;
  }

  /**
   * Get alert statistics
   */
  getStatistics(): {
    totalAlerts: number;
    activeAlerts: number;
    alertsBySevertiy: Record<AlertSeverity, number>;
    mostCommonRules: Array<{ ruleId: string; count: number }>;
  } {
    const alertsBySevertiy: Record<AlertSeverity, number> = {
      info: 0,
      warning: 0,
      error: 0,
      critical: 0,
    };

    const ruleCount: Record<string, number> = {};

    for (const alert of this.alertHistory) {
      alertsBySevertiy[alert.severity]++;
      ruleCount[alert.ruleId] = (ruleCount[alert.ruleId] || 0) + 1;
    }

    const mostCommonRules = Object.entries(ruleCount)
      .map(([ruleId, count]) => ({ ruleId, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5);

    return {
      totalAlerts: this.alertHistory.length,
      activeAlerts: this.getActiveAlerts().length,
      alertsBySevertiy,
      mostCommonRules,
    };
  }

  /**
   * Create rules from conditions
   */
  createRuleFromCondition(
    id: string,
    name: string,
    condition: AlertCondition,
    action: AlertAction,
  ): AlertRule {
    const rule: AlertRule = {
      id,
      name,
      condition,
      action,
      enabled: true,
      createdAt: Date.now(),
    };

    this.addRule(rule);
    return rule;
  }

  /**
   * Batch create rules
   */
  createDefaultRules(): void {
    const rules: AlertRule[] = [
      {
        id: 'memory-warning',
        name: 'High Memory Usage',
        condition: { type: 'high_memory', threshold: 500 },
        action: {
          type: 'notification',
          title: 'Memory Warning',
          message: 'Memory usage exceeds 500MB',
        },
        enabled: true,
        createdAt: Date.now(),
      },
      {
        id: 'error-rate',
        name: 'High Error Rate',
        condition: { type: 'high_error_rate', threshold: 0.1 },
        action: {
          type: 'notification',
          title: 'Error Rate Warning',
          message: 'Error rate exceeds 10%',
        },
        enabled: true,
        createdAt: Date.now(),
      },
      {
        id: 'rate-limit',
        name: 'Rate Limit Warning',
        condition: { type: 'rate_limit', threshold: 10 },
        action: {
          type: 'notification',
          title: 'Rate Limit',
          message: 'Approaching provider rate limit',
        },
        enabled: true,
        createdAt: Date.now(),
      },
      {
        id: 'storage-quota',
        name: 'Storage Quota Warning',
        condition: { type: 'storage_quota', threshold: 80 },
        action: {
          type: 'notification',
          title: 'Storage Warning',
          message: 'Storage usage exceeds 80%',
        },
        enabled: true,
        createdAt: Date.now(),
      },
    ];

    for (const rule of rules) {
      this.addRule(rule);
    }
  }

  /**
   * Reset alerts
   */
  reset(): void {
    this.alerts.clear();
    this.alertHistory = [];
    this.cooldowns.clear();
  }

  // Private methods

  private notifyListeners(alert: Alert): void {
    for (const listener of this.listeners) {
      try {
        listener.onAlert(alert);
      } catch (error) {
        console.error('Error notifying listener:', error);
      }
    }
  }

  private isInCooldown(ruleId: string): boolean {
    const cooldownEnd = this.cooldowns.get(ruleId);
    if (!cooldownEnd) return false;
    return Date.now() < cooldownEnd;
  }

  private setCooldown(ruleId: string, durationMs: number): void {
    if (durationMs > 0) {
      this.cooldowns.set(ruleId, Date.now() + durationMs);
    }
  }
}
