/**
 * Collaboration Module - Index (Barrel Export)
 * Phase 14: Collaborative Features
 *
 * Exports all collaboration utilities and services
 */

// WebSocket Management
export {
  WebSocketManager,
  MultiWebSocketManager,
  type WebSocketMessage,
  type WebSocketConfig,
  type ConnectionState
} from './websocketManager';

// Presence Tracking
export {
  PresenceManager,
  PresenceColorManager,
  type UserPresence,
  type CursorPosition,
  type SelectionRange
} from './presence';

// Operational Transformation
export {
  OperationalTransformer,
  CollaborativeDocument,
  type Operation,
  type ChangeSet,
  type OperationType
} from './operationalTransform';

// Activity Tracking
export {
  ActivityFeedManager,
  type Activity,
  type ActivityStats,
  type ActivityType
} from './activityFeed';

// Collaboration Manager (Main Orchestrator)
export {
  CollaborationManager,
  type CollaborationConfig,
  type CollaborativeWorkspace
} from './collaborationManager';

export default {
  WebSocketManager,
  MultiWebSocketManager,
  PresenceManager,
  PresenceColorManager,
  OperationalTransformer,
  CollaborativeDocument,
  ActivityFeedManager,
  CollaborationManager
};
