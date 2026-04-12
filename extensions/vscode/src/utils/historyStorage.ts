/**
 * History Storage - Persistent Chat History Management
 *
 * Handles saving, loading, and managing chat session history
 * stored locally in the user's home directory.
 */

import * as fs from "fs";
import * as path from "path";
import * as os from "os";
import type { ChatSession, ChatMessage } from "@/types/vectora";

const HISTORY_DIR = path.join(os.homedir(), ".vectora", "history");
const MAX_SESSIONS = 50;
const SESSION_FILE_PATTERN = /^session-[\w\-]+\.json$/;

/**
 * Ensure history directory exists
 */
function ensureHistoryDir(): void {
  if (!fs.existsSync(HISTORY_DIR)) {
    fs.mkdirSync(HISTORY_DIR, { recursive: true });
  }
}

/**
 * Save a chat session to disk
 */
export function saveSession(session: ChatSession): void {
  try {
    ensureHistoryDir();
    const filename = `session-${session.id}.json`;
    const filepath = path.join(HISTORY_DIR, filename);

    const data = JSON.stringify(session, null, 2);
    fs.writeFileSync(filepath, data, "utf-8");
  } catch (err) {
    console.error("Failed to save session:", err);
  }
}

/**
 * Load a specific session by ID
 */
export function loadSession(sessionId: string): ChatSession | null {
  try {
    ensureHistoryDir();
    const filename = `session-${sessionId}.json`;
    const filepath = path.join(HISTORY_DIR, filename);

    if (!fs.existsSync(filepath)) {
      return null;
    }

    const data = fs.readFileSync(filepath, "utf-8");
    return JSON.parse(data) as ChatSession;
  } catch (err) {
    console.error("Failed to load session:", err);
    return null;
  }
}

/**
 * List all saved sessions with metadata
 */
export function listSessions(): Array<{
  id: string;
  title: string;
  provider: string;
  model: string;
  messageCount: number;
  lastModified: number;
}> {
  try {
    ensureHistoryDir();
    const files = fs.readdirSync(HISTORY_DIR).filter((f) => SESSION_FILE_PATTERN.test(f));

    const sessions = files
      .map((filename) => {
        try {
          const filepath = path.join(HISTORY_DIR, filename);
          const stat = fs.statSync(filepath);
          const data = fs.readFileSync(filepath, "utf-8");
          const session = JSON.parse(data) as ChatSession;

          return {
            id: session.id,
            title: session.title,
            provider: session.provider,
            model: session.model,
            messageCount: session.messages.length,
            lastModified: stat.mtimeMs,
          };
        } catch {
          return null;
        }
      })
      .filter((s) => s !== null) as Array<{
        id: string;
        title: string;
        provider: string;
        model: string;
        messageCount: number;
        lastModified: number;
      }>;

    // Sort by last modified (newest first) and limit
    return sessions.sort((a, b) => b.lastModified - a.lastModified).slice(0, MAX_SESSIONS);
  } catch (err) {
    console.error("Failed to list sessions:", err);
    return [];
  }
}

/**
 * Search sessions by title or content
 */
export function searchSessions(query: string): ChatSession[] {
  try {
    ensureHistoryDir();
    const files = fs.readdirSync(HISTORY_DIR).filter((f) => SESSION_FILE_PATTERN.test(f));
    const lowerQuery = query.toLowerCase();

    return files
      .map((filename) => {
        try {
          const filepath = path.join(HISTORY_DIR, filename);
          const data = fs.readFileSync(filepath, "utf-8");
          return JSON.parse(data) as ChatSession;
        } catch {
          return null;
        }
      })
      .filter((session) => {
        if (!session) return false;

        // Search in title
        if (session.title.toLowerCase().includes(lowerQuery)) return true;

        // Search in message content
        return session.messages.some((msg) =>
          msg.content.toLowerCase().includes(lowerQuery)
        );
      }) as ChatSession[];
  } catch (err) {
    console.error("Failed to search sessions:", err);
    return [];
  }
}

/**
 * Delete a session
 */
export function deleteSession(sessionId: string): boolean {
  try {
    const filename = `session-${sessionId}.json`;
    const filepath = path.join(HISTORY_DIR, filename);

    if (fs.existsSync(filepath)) {
      fs.unlinkSync(filepath);
      return true;
    }
    return false;
  } catch (err) {
    console.error("Failed to delete session:", err);
    return false;
  }
}

/**
 * Clear all history
 */
export function clearHistory(): void {
  try {
    ensureHistoryDir();
    const files = fs.readdirSync(HISTORY_DIR).filter((f) => SESSION_FILE_PATTERN.test(f));

    for (const file of files) {
      try {
        fs.unlinkSync(path.join(HISTORY_DIR, file));
      } catch {
        // Ignore individual deletion errors
      }
    }
  } catch (err) {
    console.error("Failed to clear history:", err);
  }
}

/**
 * Export sessions for backup
 */
export function exportSessions(outputPath: string): boolean {
  try {
    ensureHistoryDir();
    const sessions = listSessions().map((meta) => loadSession(meta.id)).filter((s) => s !== null);

    const data = JSON.stringify(sessions, null, 2);
    fs.writeFileSync(outputPath, data, "utf-8");
    return true;
  } catch (err) {
    console.error("Failed to export sessions:", err);
    return false;
  }
}

/**
 * Import sessions from backup
 */
export function importSessions(inputPath: string): boolean {
  try {
    if (!fs.existsSync(inputPath)) {
      return false;
    }

    const data = fs.readFileSync(inputPath, "utf-8");
    const sessions = JSON.parse(data) as ChatSession[];

    ensureHistoryDir();
    for (const session of sessions) {
      saveSession(session);
    }
    return true;
  } catch (err) {
    console.error("Failed to import sessions:", err);
    return false;
  }
}

/**
 * Get storage size in bytes
 */
export function getStorageSize(): number {
  try {
    ensureHistoryDir();
    let totalSize = 0;

    const files = fs.readdirSync(HISTORY_DIR);
    for (const file of files) {
      const filepath = path.join(HISTORY_DIR, file);
      const stat = fs.statSync(filepath);
      totalSize += stat.size;
    }

    return totalSize;
  } catch (err) {
    console.error("Failed to get storage size:", err);
    return 0;
  }
}
