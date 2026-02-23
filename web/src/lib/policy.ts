// Migrated from Go policy probability algorithm
// Original: calculatePolicyProbability() in ui-aggregator.go

// Keywords that indicate policy relevance
const policyKeywords = [
  "policy", "regulation", "law", "government", "legislation", "bill",
  "congress", "senate", "parliament", "decree", "treaty", "court",
  "ruling", "initiative",
];

/**
 * Calculates policy probability score based on policy-related keywords
 * @param text - The text to analyze
 * @returns Score from 0 to 100 (percentage)
 */
export function calculatePolicyProbability(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  for (const word of policyKeywords) {
    const count = countOccurrences(textLower, word);
    score += count * 10;
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
