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

/**
 * Calculates sentiment score based on keywords and phrases
 * @param text - The text to analyze
 * @returns Score from -100 to 100
 */
export function calculateSentimentScore(text: string): number {
  let score = 0;
  const textLower = text.toLowerCase();

  // Check for positive phrases first
  for (const phrase of positivePhrases) {
    const count = countOccurrences(textLower, phrase);
    score += count * 15;
  }

  // Check for negative phrases
  for (const phrase of negativePhrases) {
    const count = countOccurrences(textLower, phrase);
    score -= count * 15;
  }

  // Check individual words
  const words = textLower.split(/[^a-z0-9]+/).filter(w => w.length > 0);

  for (const word of words) {
    if (positiveKeywords.has(word)) {
      score += 10;
    }
    if (negativeKeywords.has(word)) {
      score -= 10;
    }
  }

  // Cap the score
  if (score > 100) score = 100;
  if (score < -100) score = -100;

  return score;
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
