/**
 * Text highlighting utility
 * TODO: Implement proper text highlighting with search term matching
 */

export interface HighlightOptions {
  searchTerm?: string;
  className?: string;
}

export function highlightSearchTerms(text: string, searchTerm?: string): string {
  if (!searchTerm) {
    return text;
  }

  // Simple highlight implementation
  const regex = new RegExp(`(${searchTerm})`, "gi");
  return text.replace(regex, `<mark>$1</mark>`);
}

export interface FzfMatch {
  positions: Set<number>;
  score: number;
}

export function highlightFzfMatch(text: string, match?: FzfMatch): string {
  if (!match || !match.positions || match.positions.size === 0) {
    return text;
  }

  // Build the highlighted text by inserting marks around matched characters
  let result = "";
  let inMark = false;
  const sortedPositions = Array.from(match.positions).sort((a, b) => a - b);
  let positionIndex = 0;

  for (let i = 0; i < text.length; i++) {
    const shouldMark = sortedPositions[positionIndex] === i;

    if (shouldMark && !inMark) {
      result += "<mark>";
      inMark = true;
    } else if (!shouldMark && inMark) {
      result += "</mark>";
      inMark = false;
    }

    result += text[i];

    if (shouldMark) {
      positionIndex++;
    }
  }

  if (inMark) {
    result += "</mark>";
  }

  return result;
}
