// Enhanced impact score algorithm with word boundary regex
// Original: calculateImpactScore() in main.go and ui-aggregator.go

// Impactful words that indicate importance
const impactfulWords = [
  "major", "significant", "important", "critical", "breaking", "urgent",
  "massive", "huge", "substantial", "considerable", "remarkable",
  "dramatic", "drastic", "severe", "extreme", "exceptional",
  // Additional impactful words from UI version
  "crisis", "breakthrough", "disaster", "economy", "war", "pandemic",
  "reform", "global", "election", "protest", "conflict", "threat",
];

/**
 * Creates a regex pattern with word boundaries for exact word matching
 */
function createWordBoundaryPattern(word: string): RegExp {
  const escaped = word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`, 'gi');
}

/**
 * Calculates impact score based on important words with word boundary matching
 * @param text - The text to analyze
 * @returns Score from 0 to 100
 */
export function calculateImpactScore(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  for (const word of impactfulWords) {
    const pattern = createWordBoundaryPattern(word);
    const matches = textLower.match(pattern);
    if (matches) {
      score += matches.length * 5;
    }
  }

  // Cap the score
  return Math.min(100, score);
}
