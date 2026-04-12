/**
 * Highlight diff syntax
 */

export interface HighlightedDiffLine {
  type: "add" | "remove" | "context" | "header";
  content: string;
  highlighted: string;
}

export interface HighlightedHunk {
  header: string;
  lines: HighlightedDiffLine[];
}

export function highlightDiff(diffContent: string): HighlightedDiffLine[] {
  // Stub implementation - just returns the content as-is
  // Phase 2+ will implement proper diff highlighting
  const lines = diffContent.split("\n");
  return lines.map((line) => {
    let type: "add" | "remove" | "context" | "header" = "context";

    if (line.startsWith("@@")) {
      type = "header";
    } else if (line.startsWith("+")) {
      type = "add";
    } else if (line.startsWith("-")) {
      type = "remove";
    }

    return {
      type,
      content: line,
      highlighted: line,
    };
  });
}

export function highlightHunks(diffContent: string): HighlightedHunk[] {
  // Stub implementation - parses diff into hunks
  // Phase 2+ will implement proper diff parsing and highlighting
  const lines = diffContent.split("\n");
  const hunks: HighlightedHunk[] = [];
  let currentHunk: HighlightedDiffLine[] = [];
  let currentHeader = "";

  for (const line of lines) {
    if (line.startsWith("@@")) {
      if (currentHunk.length > 0) {
        hunks.push({
          header: currentHeader,
          lines: currentHunk,
        });
        currentHunk = [];
      }
      currentHeader = line;
    } else if (line !== "") {
      let type: "add" | "remove" | "context" | "header" = "context";

      if (line.startsWith("+")) {
        type = "add";
      } else if (line.startsWith("-")) {
        type = "remove";
      }

      currentHunk.push({
        type,
        content: line,
        highlighted: line,
      });
    }
  }

  if (currentHunk.length > 0) {
    hunks.push({
      header: currentHeader,
      lines: currentHunk,
    });
  }

  return hunks;
}
