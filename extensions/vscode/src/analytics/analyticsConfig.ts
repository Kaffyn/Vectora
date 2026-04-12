/**
 * Analytics Configuration
 * Configuration management for analytics system
 */

import { AnalyticsConfig, TrackingConfig } from '../types/analytics';

export class AnalyticsConfigManager {
  private static instance: AnalyticsConfigManager;
  private config: AnalyticsConfig;
  private trackingConfig: TrackingConfig;

  private constructor() {
    this.config = this.loadAnalyticsConfig();
    this.trackingConfig = this.loadTrackingConfig();
  }

  static getInstance(): AnalyticsConfigManager {
    if (!AnalyticsConfigManager.instance) {
      AnalyticsConfigManager.instance = new AnalyticsConfigManager();
    }
    return AnalyticsConfigManager.instance;
  }

  /**
   * Get analytics configuration
   */
  getConfig(): AnalyticsConfig {
    return this.config;
  }

  /**
   * Update analytics configuration
   */
  updateConfig(updates: Partial<AnalyticsConfig>): void {
    this.config = { ...this.config, ...updates };
    this.saveAnalyticsConfig();
  }

  /**
   * Get tracking configuration
   */
  getTrackingConfig(): TrackingConfig {
    return this.trackingConfig;
  }

  /**
   * Update tracking configuration
   */
  updateTrackingConfig(updates: Partial<TrackingConfig>): void {
    this.trackingConfig = { ...this.trackingConfig, ...updates };
    this.saveTrackingConfig();
  }

  /**
   * Enable analytics
   */
  enable(): void {
    this.config.enabled = true;
    this.saveAnalyticsConfig();
  }

  /**
   * Disable analytics
   */
  disable(): void {
    this.config.enabled = false;
    this.saveAnalyticsConfig();
  }

  /**
   * Check if analytics is enabled
   */
  isEnabled(): boolean {
    return this.config.enabled && !this.trackingConfig.doNotTrack;
  }

  /**
   * Set tracking level
   */
  setTrackingLevel(level: 'none' | 'basic' | 'detailed'): void {
    this.config.trackingLevel = level;
    this.saveAnalyticsConfig();
  }

  /**
   * Get tracking level
   */
  getTrackingLevel(): 'none' | 'basic' | 'detailed' {
    return this.config.trackingLevel;
  }

  /**
   * Set privacy level
   */
  setPrivacyLevel(level: 'strict' | 'normal' | 'detailed'): void {
    this.config.privacyLevel = level;
    this.saveAnalyticsConfig();
  }

  /**
   * Check if tracking is allowed
   */
  isTrackingAllowed(): boolean {
    return this.config.trackingLevel !== 'none' && this.isEnabled();
  }

  /**
   * Check if personal data can be tracked
   */
  canTrackPersonalData(): boolean {
    return (
      this.config.privacyLevel !== 'strict' &&
      !this.trackingConfig.anonymizeIp &&
      this.isEnabled()
    );
  }

  /**
   * Get retention days
   */
  getRetentionDays(): number {
    return this.config.retentionDays;
  }

  /**
   * Set retention days
   */
  setRetentionDays(days: number): void {
    if (days < 1 || days > 365) {
      throw new Error('Retention days must be between 1 and 365');
    }
    this.config.retentionDays = days;
    this.saveAnalyticsConfig();
  }

  /**
   * Reset to default configuration
   */
  resetToDefaults(): void {
    this.config = this.getDefaultConfig();
    this.trackingConfig = this.getDefaultTrackingConfig();
    this.saveAnalyticsConfig();
    this.saveTrackingConfig();
  }

  /**
   * Get default configuration
   */
  private getDefaultConfig(): AnalyticsConfig {
    return {
      enabled: true,
      trackingLevel: 'detailed',
      retentionDays: 30,
      sessionIdStrategy: 'session',
      privacyLevel: 'normal',
      batchSize: 50,
      flushIntervalMs: 30000,
      storageBackend: 'both',
      exportFormats: ['json', 'csv'],
    };
  }

  /**
   * Get default tracking configuration
   */
  private getDefaultTrackingConfig(): TrackingConfig {
    return {
      enabled: true,
      anonymizeIp: false,
      cookieConsent: false,
      doNotTrack: this.checkDoNotTrackPreference(),
      dataRetention: 30,
    };
  }

  /**
   * Load analytics config from storage
   */
  private loadAnalyticsConfig(): AnalyticsConfig {
    try {
      const stored = localStorage.getItem('vectora-analytics-config');
      if (stored) {
        return JSON.parse(stored);
      }
    } catch (error) {
      console.warn('Failed to load analytics config:', error);
    }
    return this.getDefaultConfig();
  }

  /**
   * Load tracking config from storage
   */
  private loadTrackingConfig(): TrackingConfig {
    try {
      const stored = localStorage.getItem('vectora-tracking-config');
      if (stored) {
        return JSON.parse(stored);
      }
    } catch (error) {
      console.warn('Failed to load tracking config:', error);
    }
    return this.getDefaultTrackingConfig();
  }

  /**
   * Save analytics config to storage
   */
  private saveAnalyticsConfig(): void {
    try {
      localStorage.setItem('vectora-analytics-config', JSON.stringify(this.config));
    } catch (error) {
      console.warn('Failed to save analytics config:', error);
    }
  }

  /**
   * Save tracking config to storage
   */
  private saveTrackingConfig(): void {
    try {
      localStorage.setItem('vectora-tracking-config', JSON.stringify(this.trackingConfig));
    } catch (error) {
      console.warn('Failed to save tracking config:', error);
    }
  }

  /**
   * Check browser's Do Not Track preference
   */
  private checkDoNotTrackPreference(): boolean {
    if (typeof navigator === 'undefined') return false;
    const dnt = navigator.doNotTrack as string | undefined;
    return dnt === '1' || dnt === 'yes';
  }

  /**
   * Export current configuration
   */
  exportConfig(): Record<string, unknown> {
    return {
      analytics: this.config,
      tracking: this.trackingConfig,
      exportedAt: new Date().toISOString(),
    };
  }

  /**
   * Import configuration
   */
  importConfig(data: Record<string, unknown>): void {
    try {
      if (data.analytics && typeof data.analytics === 'object') {
        this.config = { ...this.config, ...data.analytics };
      }
      if (data.tracking && typeof data.tracking === 'object') {
        this.trackingConfig = { ...this.trackingConfig, ...data.tracking };
      }
      this.saveAnalyticsConfig();
      this.saveTrackingConfig();
    } catch (error) {
      console.error('Failed to import config:', error);
      throw error;
    }
  }

  /**
   * Validate configuration
   */
  validateConfig(): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (this.config.retentionDays < 1 || this.config.retentionDays > 365) {
      errors.push('Retention days must be between 1 and 365');
    }

    if (this.config.batchSize < 10 || this.config.batchSize > 1000) {
      errors.push('Batch size must be between 10 and 1000');
    }

    if (this.config.flushIntervalMs < 5000 || this.config.flushIntervalMs > 300000) {
      errors.push('Flush interval must be between 5 and 300 seconds');
    }

    if (!['localStorage', 'indexeddb', 'both'].includes(this.config.storageBackend)) {
      errors.push('Invalid storage backend');
    }

    return {
      valid: errors.length === 0,
      errors,
    };
  }

  /**
   * Get configuration summary
   */
  getSummary(): Record<string, unknown> {
    return {
      analyticsEnabled: this.config.enabled,
      trackingLevel: this.config.trackingLevel,
      privacyLevel: this.config.privacyLevel,
      retentionDays: this.config.retentionDays,
      storageBackend: this.config.storageBackend,
      trackingAllowed: this.isTrackingAllowed(),
      doNotTrack: this.trackingConfig.doNotTrack,
    };
  }
}
