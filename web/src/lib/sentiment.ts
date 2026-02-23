// Enhanced sentiment analysis with negation detection and word boundaries
// Improvements over original Go implementation:
// 1. Negation detection (not, never, no, lack)
// 2. Regex with word boundaries for accurate phrase matching
// 3. Modular design for future NLP library integration

// Positive keywords for sentiment analysis
const positiveKeywords = new Set([
  // General Positive
  "good", "great", "excellent", "positive", "success", "improve", "benefit", "effective", "strong", "happy", "joy", "love", "optimistic", "favorable", "promising", "encouraging",
  // Growth & Expansion
  "grow", "growth", "expansion", "expand", "increase", "surge", "rise", "upward", "upturn", "boom", "accelerate", "augment", "boost", "rally", "recover", "recovery",
  // Achievement & Performance
  "achieve", "achieved", "outperform", "exceed", "beat", "record", "profitable", "profit", "gains", "earnings", "revenue", "dividend", "surplus",
  // Innovation & Advancement
  "innovative", "innovation", "breakthrough", "advance", "launch", "new", "develop", "upgrade", "leading", "cutting-edge",
  // Market Sentiment & Confidence
  "bullish", "optimism", "confidence", "stable", "stability", "support", "demand", "hot", "high", "robust",
  // Deals & Approvals
  "acquire", "acquisition", "merger", "partnership", "agreement", "approve", "approved", "endorse", "confirm",
]);

// Negative keywords for sentiment analysis
const negativeKeywords = new Set([
  // General Negative
  "bad", "poor", "terrible", "negative", "fail", "failure", "weak", "adverse", "sad", "angry", "fear", "pessimistic", "unfavorable", "discouraging",
  // Decline & Contraction
  "decline", "decrease", "drop", "fall", "slump", "downturn", "recession", "contraction", "reduce", "cut", "loss", "losses", "deficit", "shrink", "erode", "weaken",
  // Problems & Risks
  "crisis", "disaster", "risk", "warn", "warning", "threat", "problem", "issue", "concern", "challenge", "obstacle", "difficulty", "uncertainty", "volatile", "volatility",
  // Poor Performance
  "underperform", "miss", "shortfall", "struggle", "stagnate", "delay", "halt",
  // Market Sentiment & Lack of Confidence
  "bearish", "pessimism", "doubt", "skepticism", "unstable", "instability", "pressure", "low", "oversupply", "bubble",
  // Legal & Regulatory Issues
  "investigation", "lawsuit", "penalty", "fine", "sanction", "ban", "fraud", "scandal", "recall", "dispute", "reject", "denied", "downgrade",
]);

// Positive phrases for more accurate sentiment analysis
const positivePhrases = [
  "strong results", "exceeded expectations", "record high", "beats estimates",
  "outperforms market", "positive outlook", "upbeat forecast", "robust growth",
  "solid performance", "impressive gains", "significant improvement",
];

// Negative phrases for more accurate sentiment analysis
const negativePhrases = [
  "fell short", "missed expectations", "record low", "disappointing results",
  "underperforms market", "negative outlook", "bleak forecast", "steep decline",
  "poor performance", "significant losses", "sharp drop", "market crash",
];

// Negation words that reverse sentiment
const negationWords = [
  "not", "no", "never", "neither", "nobody", "nothing", "nowhere",
  "lack", "lacking", "lacks", "lacked",
  "without", "hardly", "barely", "scarcely",
  "fail", "failed", "fails", "failing",
  "unlikely", "impossible", "cannot", "can't", "won't", "wouldn't", "couldn't", "shouldn't",
];

// Window size for negation detection (number of words to look back)
const NEGATION_WINDOW = 3;

/**
 * Creates a regex pattern with word boundaries for exact phrase matching
 */
function createWordBoundaryPattern(phrase: string): RegExp {
  // Escape special regex characters and add word boundaries
  const escaped = phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`, 'gi');
}

/**
 * Checks if a word is negated by looking at preceding words
 */
function isNegated(words: string[], currentIndex: number): boolean {
  const start = Math.max(0, currentIndex - NEGATION_WINDOW);
  for (let i = start; i < currentIndex; i++) {
    if (negationWords.includes(words[i])) {
      return true;
    }
  }
  return false;
}

/**
 * Calculates sentiment score based on keywords and phrases with negation detection
 * @param text - The text to analyze
 * @returns Score from -100 to 100
 */
export function calculateSentimentScore(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  // Split into words for negation detection
  const words = textLower.split(/[^a-z0-9]+/).filter(w => w.length > 0);

  // Check for positive phrases first (with word boundaries)
  for (const phrase of positivePhrases) {
    const pattern = createWordBoundaryPattern(phrase);
    const matches = textLower.match(pattern);
    if (matches) {
      // Check if phrase is negated by looking at word before
      const phraseIndex = textLower.indexOf(phrase);
      const wordsBeforePhrase = textLower.slice(0, phraseIndex).split(/\s+/).filter(w => w.length > 0);
      const recentWords = wordsBeforePhrase.slice(-NEGATION_WINDOW);
      const isPhraseNegated = recentWords.some(w => negationWords.includes(w));
      
      score += matches.length * 15 * (isPhraseNegated ? -1 : 1);
    }
  }

  // Check for negative phrases (with word boundaries)
  for (const phrase of negativePhrases) {
    const pattern = createWordBoundaryPattern(phrase);
    const matches = textLower.match(pattern);
    if (matches) {
      const phraseIndex = textLower.indexOf(phrase);
      const wordsBeforePhrase = textLower.slice(0, phraseIndex).split(/\s+/).filter(w => w.length > 0);
      const recentWords = wordsBeforePhrase.slice(-NEGATION_WINDOW);
      const isPhraseNegated = recentWords.some(w => negationWords.includes(w));
      
      // Negating a negative = positive
      score -= matches.length * 15 * (isPhraseNegated ? -1 : 1);
    }
  }

  // Check individual words with negation detection
  for (let i = 0; i < words.length; i++) {
    const word = words[i];
    const negated = isNegated(words, i);
    
    if (positiveKeywords.has(word)) {
      score += 10 * (negated ? -1 : 1);
    }
    if (negativeKeywords.has(word)) {
      score -= 10 * (negated ? -1 : 1);
    }
  }

  // Cap the score
  if (score > 100) score = 100;
  if (score < -100) score = -100;

  return score;
}

/**
 * Analyzes sentiment and returns detailed breakdown (for debugging/advanced use)
 */
export function analyzeSentimentDetailed(text: string): {
  score: number;
  positiveMatches: string[];
  negativeMatches: string[];
  negatedPhrases: string[];
} {
  const textLower = text.toLowerCase();
  const words = textLower.split(/[^a-z0-9]+/).filter(w => w.length > 0);
  let score = 0;
  const positiveMatches: string[] = [];
  const negativeMatches: string[] = [];
  const negatedPhrases: string[] = [];

  // Check phrases
  for (const phrase of positivePhrases) {
    const pattern = createWordBoundaryPattern(phrase);
    const matches = textLower.match(pattern);
    if (matches) {
      const phraseIndex = textLower.indexOf(phrase);
      const wordsBeforePhrase = textLower.slice(0, phraseIndex).split(/\s+/).filter(w => w.length > 0);
      const recentWords = wordsBeforePhrase.slice(-NEGATION_WINDOW);
      const isNegated = recentWords.some(w => negationWords.includes(w));
      
      score += matches.length * 15 * (isNegated ? -1 : 1);
      if (isNegated) {
        negatedPhrases.push(phrase);
      } else {
        positiveMatches.push(phrase);
      }
    }
  }

  for (const phrase of negativePhrases) {
    const pattern = createWordBoundaryPattern(phrase);
    const matches = textLower.match(pattern);
    if (matches) {
      const phraseIndex = textLower.indexOf(phrase);
      const wordsBeforePhrase = textLower.slice(0, phraseIndex).split(/\s+/).filter(w => w.length > 0);
      const recentWords = wordsBeforePhrase.slice(-NEGATION_WINDOW);
      const isNegated = recentWords.some(w => negationWords.includes(w));
      
      score -= matches.length * 15 * (isNegated ? -1 : 1);
      if (isNegated) {
        negatedPhrases.push(phrase);
      } else {
        negativeMatches.push(phrase);
      }
    }
  }

  // Check words
  for (let i = 0; i < words.length; i++) {
    const word = words[i];
    const negated = isNegated(words, i);
    
    if (positiveKeywords.has(word)) {
      score += 10 * (negated ? -1 : 1);
      if (!negated) positiveMatches.push(word);
      else negatedPhrases.push(`not ${word}`);
    }
    if (negativeKeywords.has(word)) {
      score -= 10 * (negated ? -1 : 1);
      if (!negated) negativeMatches.push(word);
      else negatedPhrases.push(`not ${word}`);
    }
  }

  score = Math.max(-100, Math.min(100, score));

  return { score, positiveMatches, negativeMatches, negatedPhrases };
}
