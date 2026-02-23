// Enhanced policy probability algorithm with word boundary regex
// Original: calculatePolicyProbability() in ui-aggregator.go

// Keywords that indicate policy relevance
const policyKeywords = [
  "policy", "regulation", "law", "government", "legislation", "bill",
  "congress", "senate", "parliament", "decree", "treaty", "court",
  "ruling", "initiative",
];

/**
 * Creates a regex pattern with word boundaries for exact word matching
 */
function createWordBoundaryPattern(word: string): RegExp {
  const escaped = word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`, 'gi');
}

/**
 * Calculates policy probability score based on policy-related keywords with word boundary matching
 * @param text - The text to analyze
 * @returns Score from 0 to 100 (percentage)
 */
export function calculatePolicyProbability(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  for (const word of policyKeywords) {
    const pattern = createWordBoundaryPattern(word);
    const matches = textLower.match(pattern);
    if (matches) {
      score += matches.length * 10;
    }
  }

  // Cap the score
  return Math.min(100, score);
}
