/**
 * Converts RFC3339 time to human-readable format
 */
export function humanTime(dateString: string): string {
  const date = new Date(dateString);
  if (isNaN(date.getTime())) {
    return dateString;
  }

  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) {
    return 'just now';
  } else if (diffMins < 60) {
    return `${diffMins}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else if (diffDays < 7) {
    return `${diffDays}d ago`;
  } else {
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric', 
      year: 'numeric' 
    });
  }
}

/**
 * Formats a number with commas
 */
export function formatNumber(num: number): string {
  return num.toLocaleString('en-US');
}
