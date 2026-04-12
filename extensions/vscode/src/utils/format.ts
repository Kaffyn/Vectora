/**
 * Format a large number with thousand separators
 * @param num The number to format
 * @returns Formatted string representation
 */
export function formatLargeNumber(num: number): string {
  if (!Number.isFinite(num)) return "0";

  // Use Intl.NumberFormat for locale-aware formatting
  return new Intl.NumberFormat("en-US").format(num);
}

/**
 * Format a timestamp as "time ago"
 * @param timestamp ISO timestamp or milliseconds
 * @returns Formatted time ago string
 */
export function formatTimeAgo(timestamp: string | number): string {
  let date: Date;

  if (typeof timestamp === "string") {
    date = new Date(timestamp);
  } else {
    date = new Date(timestamp);
  }

  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;

  return date.toLocaleDateString();
}
