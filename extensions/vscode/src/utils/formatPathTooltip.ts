/**
 * Format file paths for tooltip display
 */

export interface PathTooltipOptions {
  maxLength?: number;
  showDirectory?: boolean;
}

export function formatPathTooltip(filePath: string, options?: PathTooltipOptions): string {
  const maxLength = options?.maxLength ?? 60;
  const showDirectory = options?.showDirectory ?? true;

  if (!filePath) return "";

  // If path is short enough, return as-is
  if (filePath.length <= maxLength) {
    return filePath;
  }

  // Truncate with ellipsis in the middle
  const start = Math.floor((maxLength - 3) / 2);
  const end = filePath.length - Math.ceil((maxLength - 3) / 2);

  return `${filePath.substring(0, start)}...${filePath.substring(end)}`;
}
