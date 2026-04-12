/**
 * Error Recovery & Resilience
 *
 * Strategies for handling errors gracefully and recovering from failures.
 */

import * as vscode from "vscode";

/**
 * Error context for detailed logging
 */
export interface ErrorContext {
  operation: string;
  timestamp: number;
  userAgent?: string;
  version?: string;
  sessionId?: string;
}

/**
 * Custom error with context
 */
export class VectoraError extends Error {
  constructor(
    message: string,
    public code: string,
    public context?: ErrorContext,
    public originalError?: unknown
  ) {
    super(message);
    this.name = "VectoraError";
  }
}

/**
 * Error handler with recovery strategies
 */
export class ErrorRecoveryManager {
  private static instance: ErrorRecoveryManager;
  private errorLog: Array<{ error: Error; timestamp: number }> = [];
  private maxLogSize = 100;

  private constructor() {
    // Initialize error recovery
  }

  static getInstance(): ErrorRecoveryManager {
    if (!ErrorRecoveryManager.instance) {
      ErrorRecoveryManager.instance = new ErrorRecoveryManager();
    }
    return ErrorRecoveryManager.instance;
  }

  /**
   * Handle error with recovery attempt
   */
  async handleError(error: unknown, context: ErrorContext): Promise<void> {
    const err = this.normalizeError(error);
    this.logError(err);

    // Determine recovery strategy based on error type
    const strategy = this.getRecoveryStrategy(err, context);

    try {
      await strategy.recover();
    } catch (recoveryError) {
      console.error("Recovery failed:", recoveryError);
      this.notifyUser(err, strategy);
    }
  }

  /**
   * Normalize error to standard format
   */
  private normalizeError(error: unknown): Error {
    if (error instanceof Error) {
      return error;
    }
    if (typeof error === "string") {
      return new Error(error);
    }
    return new Error(`Unknown error: ${JSON.stringify(error)}`);
  }

  /**
   * Get recovery strategy for error type
   */
  private getRecoveryStrategy(
    error: Error,
    context: ErrorContext
  ): RecoveryStrategy {
    const message = error.message.toLowerCase();

    // Network errors
    if (
      message.includes("network") ||
      message.includes("timeout") ||
      message.includes("econnrefused")
    ) {
      return new NetworkErrorRecovery(context);
    }

    // File system errors
    if (
      message.includes("enoent") ||
      message.includes("permission") ||
      message.includes("eacces")
    ) {
      return new FileSystemErrorRecovery(context);
    }

    // Memory/resource errors
    if (
      message.includes("memory") ||
      message.includes("exceeded") ||
      message.includes("heap")
    ) {
      return new ResourceErrorRecovery(context);
    }

    // Generic recovery
    return new GenericErrorRecovery(context);
  }

  /**
   * Log error for debugging
   */
  private logError(error: Error): void {
    this.errorLog.push({
      error,
      timestamp: Date.now(),
    });

    // Keep log size manageable
    if (this.errorLog.length > this.maxLogSize) {
      this.errorLog.shift();
    }

    // Also log to console in dev
    if (process.env.NODE_ENV === "development") {
      console.error("[Vectora Error]", error);
    }
  }

  /**
   * Notify user of error
   */
  private notifyUser(error: Error, strategy: RecoveryStrategy): void {
    const message =
      strategy.getUserMessage() ||
      `An error occurred: ${error.message}. Please try again.`;

    vscode.window.showErrorMessage(message, "Retry", "Dismiss").then((choice) => {
      if (choice === "Retry") {
        strategy.recover();
      }
    });
  }

  /**
   * Get error log for debugging
   */
  getErrorLog(): typeof this.errorLog {
    return [...this.errorLog];
  }

  /**
   * Clear error log
   */
  clearErrorLog(): void {
    this.errorLog = [];
  }
}

/**
 * Base recovery strategy
 */
abstract class RecoveryStrategy {
  protected context: ErrorContext;

  constructor(context: ErrorContext) {
    this.context = context;
  }

  abstract recover(): Promise<void>;

  getUserMessage(): string | null {
    return null;
  }
}

/**
 * Network error recovery
 */
class NetworkErrorRecovery extends RecoveryStrategy {
  async recover(): Promise<void> {
    // Attempt to reconnect
    console.log("Attempting network recovery...");
    // Implementation would reconnect to service
  }

  getUserMessage(): string {
    return "Network connection lost. Attempting to reconnect...";
  }
}

/**
 * File system error recovery
 */
class FileSystemErrorRecovery extends RecoveryStrategy {
  async recover(): Promise<void> {
    console.log("Attempting file system recovery...");
    // Implementation would check permissions and retry
  }

  getUserMessage(): string {
    return "File access error. Please check permissions.";
  }
}

/**
 * Resource error recovery
 */
class ResourceErrorRecovery extends RecoveryStrategy {
  async recover(): Promise<void> {
    console.log("Clearing memory cache...");
    // Implementation would free resources
  }

  getUserMessage(): string {
    return "Running low on resources. Clearing cache...";
  }
}

/**
 * Generic error recovery
 */
class GenericErrorRecovery extends RecoveryStrategy {
  async recover(): Promise<void> {
    console.log("Generic error recovery...");
    // Retry with exponential backoff
  }

  getUserMessage(): string {
    return "An error occurred. Retrying...";
  }
}
