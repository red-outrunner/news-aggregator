// Combined analyzer utilities
// Re-exports all scoring functions from a single entry point

import { calculateSentimentScore } from './sentiment';
import { calculateImpactScore } from './impact';
import { calculatePolicyProbability } from './policy';

export { calculateSentimentScore, calculateImpactScore, calculatePolicyProbability };

/**
 * Analyzes text and returns all scores at once
 * @param text - The text to analyze (title + description)
 * @returns Object with all three scores
 */
export function analyzeArticle(text: string): {
  sentimentScore: number;
  impactScore: number;
  policyProbability: number;
} {
  return {
    sentimentScore: calculateSentimentScore(text),
    impactScore: calculateImpactScore(text),
    policyProbability: calculatePolicyProbability(text),
  };
}
