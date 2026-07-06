// Shared text-matching helpers for the analysis modules

/**
 * Creates a regex pattern with word boundaries for exact word/phrase matching
 */
export function createWordBoundaryPattern(phrase: string): RegExp {
  // Escape special regex characters and add word boundaries
  const escaped = phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`, 'gi');
}
