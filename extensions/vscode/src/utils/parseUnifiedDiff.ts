/**
 * Parse unified diff format
 */

export interface DiffLine {
  type: "add" | "remove" | "context" | "header";
  content: string;
  lineNumber?: number;
}

export interface DiffBlock {
  file: string;
  hunks: DiffLine[][];
}

export function parseUnifiedDiff(diffString: string): DiffBlock[] {
  // Stub implementation - just returns empty array
  // Phase 2+ will implement proper diff parsing
  const blocks: DiffBlock[] = [];

  const lines = diffString.split("\n");
  let currentBlock: DiffBlock | null = null;
  let currentHunk: DiffLine[] = [];

  for (const line of lines) {
    if (line.startsWith("---")) {
      if (currentBlock && currentHunk.length > 0) {
        currentBlock.hunks.push(currentHunk);
        currentHunk = [];
      }
      currentBlock = {
        file: line.substring(4).trim(),
        hunks: [],
      };
    } else if (line.startsWith("+++")) {
      // Skip
    } else if (line.startsWith("@@")) {
      if (currentHunk.length > 0 && currentBlock) {
        currentBlock.hunks.push(currentHunk);
      }
      currentHunk = [];
    } else if (line.startsWith("+")) {
      currentHunk.push({ type: "add", content: line.substring(1) });
    } else if (line.startsWith("-")) {
      currentHunk.push({ type: "remove", content: line.substring(1) });
    } else if (line.startsWith(" ")) {
      currentHunk.push({ type: "context", content: line.substring(1) });
    }
  }

  if (currentBlock && currentHunk.length > 0) {
    currentBlock.hunks.push(currentHunk);
    blocks.push(currentBlock);
  }

  return blocks;
}
