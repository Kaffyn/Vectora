/**
 * Operational Transformation Module - Real-time Collaborative Editing
 * Phase 14: Collaborative Features
 *
 * Provides:
 * - Operational transformation primitives
 * - Insert and delete operations
 * - Operation composition and transformation
 * - Convergence of concurrent edits
 */

/**
 * Operation type
 */
export type OperationType = 'insert' | 'delete' | 'retain';

/**
 * Single operation
 */
export interface Operation {
  type: OperationType;
  position: number;
  content?: string; // for insert
  length?: number; // for delete/retain
  userId?: string;
  timestamp?: number;
}

/**
 * Change set
 */
export interface ChangeSet {
  id: string;
  operations: Operation[];
  userId: string;
  timestamp: number;
  version: number;
}

/**
 * Operational Transformer
 */
export class OperationalTransformer {
  /**
   * Apply operations to text
   */
  static apply(text: string, operations: Operation[]): string {
    let result = '';
    let textIndex = 0;

    for (const op of operations) {
      if (op.type === 'retain') {
        const length = op.length || 0;
        result += text.substring(textIndex, textIndex + length);
        textIndex += length;
      } else if (op.type === 'insert') {
        result += op.content || '';
      } else if (op.type === 'delete') {
        textIndex += op.length || 0;
      }
    }

    // Add remaining text
    result += text.substring(textIndex);

    return result;
  }

  /**
   * Transform operation against another operation
   * Resolves concurrent edits
   */
  static transform(op1: Operation, op2: Operation): Operation {
    // If operations don't overlap, return original
    if (op1.position >= (op2.position + (op2.length || 0))) {
      return op1;
    }

    if (op2.position >= (op1.position + (op1.length || 0))) {
      return op1;
    }

    // Handle overlapping operations
    if (op1.type === 'insert' && op2.type === 'insert') {
      // Both insertions at same position - preserve order by userId
      if (op1.position === op2.position) {
        if ((op1.userId || '') < (op2.userId || '')) {
          return op1;
        } else {
          const newOp = { ...op1 };
          newOp.position += op2.content?.length || 0;
          return newOp;
        }
      }

      // Different positions
      if (op1.position < op2.position) {
        return op1;
      } else {
        const newOp = { ...op1 };
        newOp.position += op2.content?.length || 0;
        return newOp;
      }
    }

    if (op1.type === 'delete' && op2.type === 'insert') {
      if (op1.position > op2.position) {
        const newOp = { ...op1 };
        newOp.position += op2.content?.length || 0;
        return newOp;
      }
      return op1;
    }

    if (op1.type === 'insert' && op2.type === 'delete') {
      if (op1.position >= op2.position) {
        const deletedLength = op2.length || 0;
        const newOp = { ...op1 };
        newOp.position = Math.max(op2.position, op1.position - deletedLength);
        return newOp;
      }
      return op1;
    }

    if (op1.type === 'delete' && op2.type === 'delete') {
      if (op1.position === op2.position) {
        return {
          ...op1,
          length: Math.max(0, (op1.length || 0) - (op2.length || 0))
        };
      }

      if (op1.position < op2.position) {
        return op1;
      } else {
        const newOp = { ...op1 };
        newOp.position = Math.max(
          op2.position,
          op1.position - (op2.length || 0)
        );
        return newOp;
      }
    }

    return op1;
  }

  /**
   * Transform operation array against another operation
   */
  static transformAgainstOperation(operations: Operation[], against: Operation): Operation[] {
    return operations.map(op => this.transform(op, against));
  }

  /**
   * Compose operations into single operation
   */
  static compose(operations: Operation[]): Operation[] {
    if (operations.length === 0) return [];

    const composed: Operation[] = [];

    for (const op of operations) {
      if (composed.length === 0) {
        composed.push(op);
      } else {
        const last = composed[composed.length - 1];

        // Merge consecutive operations
        if (last.type === 'insert' && op.type === 'insert') {
          if (last.position + (last.content?.length || 0) === op.position) {
            last.content = (last.content || '') + (op.content || '');
          } else {
            composed.push(op);
          }
        } else if (last.type === 'delete' && op.type === 'delete') {
          if (last.position === op.position) {
            last.length = (last.length || 0) + (op.length || 0);
          } else {
            composed.push(op);
          }
        } else {
          composed.push(op);
        }
      }
    }

    return composed;
  }

  /**
   * Invert operations
   */
  static invert(operations: Operation[]): Operation[] {
    const inverted: Operation[] = [];
    let offset = 0;

    for (const op of operations) {
      if (op.type === 'insert') {
        const length = op.content?.length || 0;
        inverted.push({
          type: 'delete',
          position: op.position + offset,
          length
        });
        offset -= length;
      } else if (op.type === 'delete') {
        const length = op.length || 0;
        inverted.push({
          type: 'insert',
          position: op.position + offset,
          content: '' // We don't store deleted content
        });
        offset += length;
      }
    }

    return inverted.reverse();
  }

  /**
   * Get operation string representation for debugging
   */
  static toString(operations: Operation[]): string {
    return operations
      .map(op => {
        switch (op.type) {
          case 'insert':
            return `+${op.content}`;
          case 'delete':
            return `-${op.length}`;
          case 'retain':
            return `=${op.length}`;
          default:
            return '';
        }
      })
      .join('');
  }
}

/**
 * Collaborative document manager
 */
export class CollaborativeDocument {
  private content: string = '';
  private version: number = 0;
  private history: ChangeSet[] = [];
  private maxHistorySize: number = 1000;

  constructor(initialContent: string = '') {
    this.content = initialContent;
  }

  /**
   * Apply change set
   */
  applyChangeSet(changeSet: ChangeSet): void {
    if (changeSet.version !== this.version) {
      console.warn(`Version mismatch: expected ${this.version}, got ${changeSet.version}`);
    }

    this.content = OperationalTransformer.apply(this.content, changeSet.operations);
    this.version = changeSet.version + 1;

    this.history.push(changeSet);

    // Keep history size manageable
    if (this.history.length > this.maxHistorySize) {
      this.history = this.history.slice(-this.maxHistorySize);
    }
  }

  /**
   * Create change set from operations
   */
  createChangeSet(operations: Operation[], userId: string): ChangeSet {
    return {
      id: `cs-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      operations,
      userId,
      timestamp: Date.now(),
      version: this.version
    };
  }

  /**
   * Get current content
   */
  getContent(): string {
    return this.content;
  }

  /**
   * Get current version
   */
  getVersion(): number {
    return this.version;
  }

  /**
   * Get history
   */
  getHistory(): ChangeSet[] {
    return [...this.history];
  }

  /**
   * Get changes since version
   */
  getChangesSince(version: number): ChangeSet[] {
    return this.history.filter(cs => cs.version >= version);
  }

  /**
   * Reset to initial state
   */
  reset(content: string = ''): void {
    this.content = content;
    this.version = 0;
    this.history = [];
  }
}

export default OperationalTransformer;
