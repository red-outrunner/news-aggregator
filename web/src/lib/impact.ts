// Migrated from Go impact score algorithm
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
 * Calculates impact score based on important words
 * @param text - The text to analyze
 * @returns Score from 0 to 100
 */
export function calculateImpactScore(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  for (const word of impactfulWords) {
    const count = countOccurrences(textLower, word);
    score += count * 5;
  }

  // Cap the score
  return Math.min(100, score);
}

/**
 * Counts occurrences of a substring in a string
 */
function countOccurrences(str: string, substr: string): number {
  let count = 0;
  let pos = 0;
  while ((pos = str.indexOf(substr, pos)) !== -1) {
    count++;
    pos += substr.length;
  }
  return count;
}
