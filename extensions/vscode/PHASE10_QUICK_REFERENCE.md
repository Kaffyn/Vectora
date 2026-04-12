# Phase 10: Analytics & Monitoring - Quick Reference

## 🚀 Quick Start (5 minutes)

### 1. Initialize Analytics
```typescript
import { Analytics } from './analytics/analytics';

const analytics = Analytics.getInstance();
```

### 2. Track an Event
```typescript
// Chat event
analytics.trackChat({
  messageLength: 250,
  provider: 'claude',
  responseTime: 1500
});

// RAG event
analytics.trackRAG({
  query: 'how to...',
  resultCount: 5,
  executionTime: 200
});

// Custom event
analytics.trackEvent('ui.action', {
  action: 'button_click',
  component: 'settings'
});
```

### 3. Use in React
```typescript
import { useAnalytics } from './hooks/useAnalytics';

function MyComponent() {
  const { trackEvent, trackChat } = useAnalytics();

  return (
    <button onClick={() => trackEvent('ui.action', {})}>
      Click me
    </button>
  );
}
```

### 4. View Dashboard Data
```typescript
import { useDashboard } from './hooks/useDashboard';

function Dashboard() {
  const { data, loadDashboardData } = useDashboard();

  useEffect(() => {
    loadDashboardData();
  }, []);

  if (!data) return <div>Loading...</div>;

  return <div>Total Chats: {data.overview.totalChats}</div>;
}
```

## 📊 Key Classes

### Analytics
```typescript
// Singleton instance
const analytics = Analytics.getInstance();

// Methods
analytics.trackEvent(type, metadata)
analytics.trackChat(data)
analytics.trackRAG(data)
analytics.trackProvider(data)
analytics.getDashboardData(startTime, endTime)
analytics.exportData(format)
analytics.getMetrics(startTime, endTime)
analytics.getActiveAlerts()
analytics.createAlert(ruleId, severity, title, message)
```

### EventTracker
```typescript
const tracker = new EventTracker(sessionId);

tracker.createEvent(type, metadata, sessionId)
tracker.createChatEvent(type, data)
tracker.createRAGEvent(type, data)
tracker.createProviderEvent(type, data)
tracker.logSystemError(error, context)
tracker.logMemoryWarning(usagePercent, usageMB)
tracker.logRateLimitWarning(provider, remaining, limit)
```

### MetricsCollector
```typescript
const collector = new MetricsCollector();

collector.aggregateMetrics(events)
collector.calculateUsageMetrics(events)
collector.getTopFeatures(events, limit)
collector.getCostBreakdown(events)
collector.getProviderPerformance(events)
collector.detectAnomalies(events)
```

### PerformanceMonitor
```typescript
const monitor = new PerformanceMonitor(sessionId, eventTracker);

const markId = monitor.startMeasurement('operation');
const duration = monitor.endMeasurement(markId);
monitor.recordPoint(responseTime, memoryUsage, cpuUsage)
monitor.getAverageResponseTime()
monitor.getPeakMemoryUsage()
monitor.detectAnomalies()
```

### MemoryMonitor
```typescript
const memMonitor = new MemoryMonitor(sessionId, eventTracker);

memMonitor.startAutoMonitoring(5000)
memMonitor.takeSnapshot()
memMonitor.getCurrentUsageMB()
memMonitor.detectMemoryLeak()
memMonitor.getHealthStatus()
memMonitor.stopAutoMonitoring()
```

### HealthCheck
```typescript
const health = new HealthCheck(sessionId, eventTracker);

const result = await health.performHealthCheck()
health.recordProviderSuccess(provider, responseTime)
health.recordProviderError(provider, error)
health.updateProviderRateLimit(provider, remaining, limit)
health.startMonitoring(30000)
```

### AlertManager
```typescript
const alerts = new AlertManager();

alerts.createAlert(ruleId, severity, title, message)
alerts.addRule(rule)
alerts.enableRule(ruleId)
alerts.disableRule(ruleId)
alerts.getActiveAlerts()
alerts.subscribe(listener)
alerts.createDefaultRules()
```

### AnalyticsConfigManager
```typescript
const config = AnalyticsConfigManager.getInstance();

config.isTrackingAllowed()
config.setTrackingLevel('detailed')
config.setPrivacyLevel('normal')
config.getConfig()
config.updateConfig(updates)
config.exportConfig()
config.importConfig(data)
```

## 🎯 Event Types

```typescript
// Chat (3)
'chat.send'           // Message sent
'chat.receive'        // Response received
'chat.error'          // Chat error

// RAG (3)
'rag.search'          // Search performed
'rag.retrieve'        // Documents retrieved
'rag.rank'            // Results ranked

// Provider (3)
'provider.switch'     // Provider switched
'provider.fallback'   // Fallback used
'provider.error'      // Provider error

// Cache (2)
'cache.hit'           // Cache hit
'cache.miss'          // Cache miss

// System (3)
'system.startup'      // App startup
'system.error'        // System error
'system.warning'      // System warning

// Resources (3)
'memory.warning'      // Memory warning
'rate.limit.warning'  // Rate limit warning
'token.usage'         // Token usage logged

// UI (1)
'ui.action'           // User action
```

## 📈 Metrics

```typescript
interface PerformanceMetrics {
  responseTime: number;        // ms
  tokenInput: number;          // input tokens
  tokenOutput: number;         // output tokens
  costUSD: number;             // cost in USD
  cacheHitRate: number;        // 0-1
  errorRate: number;           // 0-1
  memoryUsageMB: number;       // MB
  cpuUsagePercent: number;     // %
}
```

## ⚙️ Configuration

```typescript
interface AnalyticsConfig {
  enabled: boolean;
  trackingLevel: 'none' | 'basic' | 'detailed';
  retentionDays: number;                    // 1-365
  sessionIdStrategy: 'device' | 'user' | 'session';
  privacyLevel: 'strict' | 'normal' | 'detailed';
  batchSize: number;                        // 10-1000
  flushIntervalMs: number;                  // 5000-300000
  storageBackend: 'localStorage' | 'indexeddb' | 'both';
  exportFormats: ('json' | 'csv')[];
}
```

## 🎨 Formatting Utilities

```typescript
import { MetricsFormatter } from './utils/analytics/metricsFormatter';

MetricsFormatter.formatResponseTime(1234)      // "1.23s"
MetricsFormatter.formatMemory(1048576)         // "1.0 MB"
MetricsFormatter.formatPercent(0.75)           // "75.0%"
MetricsFormatter.formatCost(0.005)             // "$0.0050"
MetricsFormatter.formatTokens(1250)            // "1.2K"
MetricsFormatter.formatDuration(65000)         // "1m 5s"
MetricsFormatter.formatTimestamp(Date.now())   // "4/12/2026, 10:30:45 AM"
MetricsFormatter.formatHealthStatus('healthy') // "✓ Healthy"
MetricsFormatter.formatAlertSeverity('error')  // "❌ Error"
```

## 📂 File Locations

```
src/
├── analytics/
│   ├── analytics.ts                ← Main engine
│   ├── eventTracker.ts             ← Event tracking
│   ├── metricsCollector.ts         ← Metrics
│   ├── analyticsConfig.ts          ← Configuration
│   └── index.ts                    ← Exports
├── monitoring/
│   ├── performanceMonitor.ts       ← Performance
│   ├── memoryMonitor.ts            ← Memory
│   ├── healthCheck.ts              ← Health
│   ├── alertManager.ts             ← Alerts
│   └── index.ts                    ← Exports
├── hooks/
│   ├── useAnalytics.ts             ← Analytics hook
│   ├── usePerformanceMetrics.ts    ← Performance hook
│   └── useDashboard.ts             ← Dashboard hook
├── types/
│   └── analytics.ts                ← Type definitions
└── utils/analytics/
    └── metricsFormatter.ts         ← Formatting
```

## 🔗 Dependencies

- React (for hooks)
- uuid (for ID generation)
- TypeScript (for types)
- IndexedDB (browser API)
- LocalStorage (browser API)
- Performance API (browser API)

## 📚 Documentation Files

- `PHASE10_ANALYTICS_GUIDE.md` - Complete guide (3500+ lines)
- `PHASE10_SUMMARY.md` - Implementation summary
- `PHASE10_IMPLEMENTATION_INDEX.md` - Quick index
- `PHASE10_QUICK_REFERENCE.md` - This file

## 🧪 Common Patterns

### Pattern 1: Track Chat Message
```typescript
const startTime = performance.now();
try {
  const response = await chat.send(message);
  const duration = performance.now() - startTime;

  analytics.trackChat({
    messageLength: message.length,
    tokenCount: estimateTokens(message),
    provider: 'claude',
    responseTime: duration
  });
} catch (error) {
  analytics.trackChat({
    messageLength: message.length,
    error: error.message
  });
}
```

### Pattern 2: Monitor Performance
```typescript
const { startMeasurement, endMeasurement } = usePerformanceMetrics(sessionId);

const markId = startMeasurement('search');
const results = await search(query);
const duration = endMeasurement(markId);

console.log(`Search took ${duration}ms`);
```

### Pattern 3: Check Memory Health
```typescript
const { getMemoryHealth, detectMemoryLeak } = usePerformanceMetrics(sessionId);

const health = getMemoryHealth();
if (health.status !== 'healthy') {
  console.warn(`Memory: ${health.message}`);
}

const leak = detectMemoryLeak();
if (leak.detected) {
  console.warn(`⚠️ ${leak.recommendation}`);
}
```

### Pattern 4: Dashboard Display
```typescript
const { data, loadDashboardData, startAutoRefresh } = useDashboard();

useEffect(() => {
  loadDashboardData();
  startAutoRefresh(30000);
}, []);

if (!data) return <Loading />;

return (
  <div>
    <h2>Analytics</h2>
    <p>Chats: {data.overview.totalChats}</p>
    <p>Errors: {data.overview.totalErrors}</p>
    <p>Alerts: {data.alerts.length}</p>
  </div>
);
```

## ⚠️ Common Issues

| Issue | Solution |
|-------|----------|
| Tracking not working | Check `config.isTrackingEnabled()` |
| High memory usage | Call `memoryMonitor.clearMemory()` |
| Storage quota exceeded | Export and delete old data |
| Alerts not firing | Verify rule is enabled |
| Memory leak detected | Check for unclosed resources |

## 🔍 Debug Commands

```typescript
// Check analytics status
const analytics = Analytics.getInstance();
console.log(analytics.getSessionInfo());

// Check configuration
const config = AnalyticsConfigManager.getInstance();
console.log(config.getSummary());

// Get active alerts
const alerts = analytics.getActiveAlerts();
console.log(alerts);

// Check memory
const health = memoryMonitor.getHealthStatus();
console.log(health);

// Export data
const exported = await analytics.exportData('json');
console.log(exported);
```

## 🚨 Alert Severity Levels

| Level | Color | Use Case |
|-------|-------|----------|
| info | 🔵 Blue | Informational messages |
| warning | 🟡 Amber | Non-critical issues |
| error | 🔴 Red | Critical issues |
| critical | 🔴 Dark Red | System failure |

## 💾 Data Storage

| Storage | Duration | Size | Use |
|---------|----------|------|-----|
| Memory | Session | <50MB | Real-time |
| LocalStorage | 24h | ~100KB | Quick access |
| IndexedDB | 30d | ~30MB | Historical |

---

**Last Updated**: April 2026
**Status**: Production-Ready
**Support**: See PHASE10_ANALYTICS_GUIDE.md
