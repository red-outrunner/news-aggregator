import { describe, it, expect } from 'vitest';
import { calculateSentimentScore } from './sentiment';
import { calculateImpactScore } from './impact';
import { calculatePolicyProbability } from './policy';
import { analyzeArticle } from './index';

// Exact scores locked from the shipped implementation — refactors must not
// change scoring behavior. If a test fails after an algorithm change, that
// change is intentional only if these numbers are updated deliberately.
describe('calculateSentimentScore', () => {
  it('scores positive finance language', () => {
    expect(calculateSentimentScore('Strong results as profits surge to record high')).toBe(70);
  });

  it('flips positive words that are negated', () => {
    expect(
      calculateSentimentScore('Company did not achieve growth and missed expectations this quarter')
    ).toBe(-35);
  });

  it('scores negative crisis language', () => {
    expect(
      calculateSentimentScore(
        'Market crash fears trigger massive losses amid global crisis and recession warnings'
      )
    ).toBe(-45);
  });

  it('returns zero for neutral text', () => {
    expect(calculateSentimentScore('The weather was mild and unremarkable today')).toBe(0);
  });

  it('respects word boundaries (no matches inside larger words)', () => {
    expect(calculateSentimentScore('Goodness gracious, the highlands are lovely')).toBe(0);
  });

  it('caps at +100 and -100', () => {
    expect(calculateSentimentScore('great '.repeat(20))).toBe(100);
    expect(calculateSentimentScore('crisis '.repeat(20))).toBe(-100);
  });
});

describe('calculateImpactScore', () => {
  it('counts impactful words', () => {
    expect(
      calculateImpactScore(
        'Market crash fears trigger massive losses amid global crisis and recession warnings'
      )
    ).toBe(15);
  });

  it('returns zero without impactful words and caps at 100', () => {
    expect(calculateImpactScore('The weather was mild and unremarkable today')).toBe(0);
    expect(calculateImpactScore('crisis '.repeat(25))).toBe(100);
  });
});

describe('calculatePolicyProbability', () => {
  it('counts policy keywords', () => {
    expect(
      calculatePolicyProbability('New regulation bill passes senate as government tightens policy')
    ).toBe(50);
  });

  it('caps at 100', () => {
    expect(calculatePolicyProbability('policy '.repeat(12))).toBe(100);
  });
});

describe('analyzeArticle', () => {
  it('combines all three scores', () => {
    expect(
      analyzeArticle('New regulation bill passes senate as government tightens policy')
    ).toEqual({ sentimentScore: 10, impactScore: 0, policyProbability: 50 });
  });
});
