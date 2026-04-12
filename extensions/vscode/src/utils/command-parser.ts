/**
 * Command parser utilities for tool execution
 * TODO: Implement proper command parsing
 */

export interface ParsedCommand {
  command: string;
  args: string[];
  raw: string;
}

export interface CommandPattern {
  pattern: RegExp;
  type: string;
  example: string;
}

/**
 * Parse a command line - returns array of commands split by pipes and chains
 */
export function parseCommand(commandLine: string): string[] {
  if (!commandLine) return [];

  // Split by pipes and chains, but keep the separators
  const commands = commandLine
    .split(/(?:&&|\|\||;)/)
    .map((cmd) => cmd.trim())
    .filter((cmd) => cmd.length > 0);

  return commands;
}

/**
 * Parse a single command into command and arguments
 */
export function parseSingleCommand(commandLine: string): ParsedCommand {
  const parts = commandLine.trim().split(/\s+/);
  const command = parts[0] || "";
  const args = parts.slice(1);

  return {
    command,
    args,
    raw: commandLine,
  };
}

export function escapeShellArg(arg: string): string {
  // Simple shell escaping
  if (/[^a-zA-Z0-9_.\/-]/.test(arg)) {
    return `"${arg.replace(/"/g, '\\"')}"`;
  }
  return arg;
}

export function extractPatternsFromCommand(command: string): CommandPattern[] {
  // Stub implementation - extract patterns from command
  // Phase 2+ will implement proper pattern extraction
  const patterns: CommandPattern[] = [];

  // Common patterns
  if (command.includes("&&")) {
    patterns.push({
      pattern: /&&/g,
      type: "chain",
      example: "command1 && command2",
    });
  }

  if (command.includes("|")) {
    patterns.push({
      pattern: /\|/g,
      type: "pipe",
      example: "command1 | command2",
    });
  }

  if (command.includes(">")) {
    patterns.push({
      pattern: />/g,
      type: "redirect",
      example: "command > output.txt",
    });
  }

  return patterns;
}

/**
 * Safe JSON parse - returns default value if parse fails
 */
export function safeJsonParse<T = any>(json: string, defaultValue: T): T {
  try {
    return JSON.parse(json);
  } catch (error) {
    return defaultValue;
  }
}
