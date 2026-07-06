// Shared text-matching helpers for the analysis modules

/**
 * Creates a regex pattern with word boundaries for exact word/phrase matching
 */
export function createWordBoundaryPattern(phrase: string): RegExp {
  // Escape special regex characters and add word boundaries
  const escaped = phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`, 'gi');
}

export interface CompiledPattern {
  text: string;
  pattern: RegExp;
}

/**
 * Precompiles word-boundary patterns once (at module load) so scoring
 * doesn't rebuild regexes for every article
 */
export function compileWordBoundaryPatterns(phrases: readonly string[]): CompiledPattern[] {
  return phrases.map((text) => ({ text, pattern: createWordBoundaryPattern(text) }));
}
