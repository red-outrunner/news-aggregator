import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const query = searchParams.get('q');
  const apiKey = process.env.NEWS_API_KEY;

  console.log('API Request - Query:', query, 'Has API Key:', !!apiKey);

  if (!query) {
    return NextResponse.json(
      { error: 'Missing query parameter', message: 'Please provide a search query' },
      { status: 400 }
    );
  }

  if (!apiKey) {
    console.error('NEWS_API_KEY is not set in environment variables');
    return NextResponse.json(
      { 
        error: 'Missing NEWS_API_KEY', 
        message: 'Please add your NewsAPI key to the .env.local file. See .env.local.example for format.' 
      },
      { status: 500 }
    );
  }

  try {
    const url = new URL('https://newsapi.org/v2/everything');
    url.searchParams.set('q', query);
    url.searchParams.set('sortBy', 'publishedAt');
    url.searchParams.set('language', 'en');
    url.searchParams.set('pageSize', '18');
    url.searchParams.set('apiKey', apiKey);

    const page = searchParams.get('page');
    if (page) {
      url.searchParams.set('page', page);
    }

    const fromDate = searchParams.get('fromDate');
    if (fromDate) {
      url.searchParams.set('from', fromDate);
    }

    const toDate = searchParams.get('toDate');
    if (toDate) {
      url.searchParams.set('to', toDate);
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch news', message: data.message || data.status || 'Unknown error' },
        { status: response.status }
      );
    }

    if (data.status !== 'ok') {
      return NextResponse.json(
        { error: 'API error', message: data.message || 'Unknown API error' },
        { status: 500 }
      );
    }

    // Calculate scores for each article server-side
    const articlesWithScores = data.articles.map((article: any) => {
      const content = `${article.title} ${article.description}`;
      
      // === ENHANCED SENTIMENT ANALYSIS with Negation Detection ===
      let sentimentScore = 0;
      const textLower = content.toLowerCase();
      
      const positivePhrases = [
        "strong results", "exceeded expectations", "record high", "beats estimates",
        "outperforms market", "positive outlook", "upbeat forecast", "robust growth",
        "solid performance", "impressive gains", "significant improvement",
      ];
      
      const negativePhrases = [
        "fell short", "missed expectations", "record low", "disappointing results",
        "underperforms market", "negative outlook", "bleak forecast", "steep decline",
        "poor performance", "significant losses", "sharp drop", "market crash",
      ];

      const positiveKeywords = new Set([
        "good", "great", "excellent", "positive", "success", "improve", "benefit", "effective", "strong", "happy", "joy", "love", "optimistic", "favorable", "promising", "encouraging",
        "grow", "growth", "expansion", "expand", "increase", "surge", "rise", "upward", "upturn", "boom", "accelerate", "augment", "boost", "rally", "recover", "recovery",
        "achieve", "achieved", "outperform", "exceed", "beat", "record", "profitable", "profit", "gains", "earnings", "revenue", "dividend", "surplus",
        "innovative", "innovation", "breakthrough", "advance", "launch", "new", "develop", "upgrade", "leading", "cutting-edge",
        "bullish", "optimism", "confidence", "stable", "stability", "support", "demand", "hot", "high", "robust",
        "acquire", "acquisition", "merger", "partnership", "agreement", "approve", "approved", "endorse", "confirm",
      ]);
      
      const negativeKeywords = new Set([
        "bad", "poor", "terrible", "negative", "fail", "failure", "weak", "adverse", "sad", "angry", "fear", "pessimistic", "unfavorable", "discouraging",
        "decline", "decrease", "drop", "fall", "slump", "downturn", "recession", "contraction", "reduce", "cut", "loss", "losses", "deficit", "shrink", "erode", "weaken",
        "crisis", "disaster", "risk", "warn", "warning", "threat", "problem", "issue", "concern", "challenge", "obstacle", "difficulty", "uncertainty", "volatile", "volatility",
        "underperform", "miss", "shortfall", "struggle", "stagnate", "delay", "halt",
        "bearish", "pessimism", "doubt", "skepticism", "unstable", "instability", "pressure", "low", "oversupply", "bubble",
        "investigation", "lawsuit", "penalty", "fine", "sanction", "ban", "fraud", "scandal", "recall", "dispute", "reject", "denied", "downgrade",
      ]);

      const negationWords = [
        "not", "no", "never", "neither", "nobody", "nothing", "nowhere",
        "lack", "lacking", "lacks", "lacked",
        "without", "hardly", "barely", "scarcely",
        "fail", "failed", "fails", "failing",
        "unlikely", "impossible", "cannot", "can't", "won't", "wouldn't", "couldn't", "shouldn't",
      ];
      const NEGATION_WINDOW = 3;

      // Helper: check if word at index is negated
      const isNegated = (words: string[], idx: number): boolean => {
        const start = Math.max(0, idx - NEGATION_WINDOW);
        for (let i = start; i < idx; i++) {
          if (negationWords.includes(words[i])) return true;
        }
        return false;
      };

      // Helper: create regex with word boundaries
      const createPattern = (phrase: string) => {
        const escaped = phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        return new RegExp(`\\b${escaped}\\b`, 'gi');
      };

      // Check positive phrases with word boundaries and negation
      for (const phrase of positivePhrases) {
        const pattern = createPattern(phrase);
        const matches = textLower.match(pattern);
        if (matches) {
          const phraseIdx = textLower.indexOf(phrase);
          const wordsBefore = textLower.slice(0, phraseIdx).split(/\s+/).filter(w => w.length > 0);
          const recentWords = wordsBefore.slice(-NEGATION_WINDOW);
          const isPhraseNegated = recentWords.some(w => negationWords.includes(w));
          sentimentScore += matches.length * 15 * (isPhraseNegated ? -1 : 1);
        }
      }

      // Check negative phrases with word boundaries and negation
      for (const phrase of negativePhrases) {
        const pattern = createPattern(phrase);
        const matches = textLower.match(pattern);
        if (matches) {
          const phraseIdx = textLower.indexOf(phrase);
          const wordsBefore = textLower.slice(0, phraseIdx).split(/\s+/).filter(w => w.length > 0);
          const recentWords = wordsBefore.slice(-NEGATION_WINDOW);
          const isPhraseNegated = recentWords.some(w => negationWords.includes(w));
          sentimentScore -= matches.length * 15 * (isPhraseNegated ? -1 : 1);
        }
      }

      // Check individual words with negation detection
      const words = textLower.split(/[^a-z0-9]+/).filter(w => w.length > 0);
      for (let i = 0; i < words.length; i++) {
        const negated = isNegated(words, i);
        if (positiveKeywords.has(words[i])) {
          sentimentScore += 10 * (negated ? -1 : 1);
        }
        if (negativeKeywords.has(words[i])) {
          sentimentScore -= 10 * (negated ? -1 : 1);
        }
      }

      // Cap sentiment score
      sentimentScore = Math.max(-100, Math.min(100, sentimentScore));

      // Impact score with word boundaries (uses createPattern from above)
      const impactfulWords = [
        "major", "significant", "important", "critical", "breaking", "urgent",
        "massive", "huge", "substantial", "considerable", "remarkable",
        "dramatic", "drastic", "severe", "extreme", "exceptional",
        "crisis", "breakthrough", "disaster", "economy", "war", "pandemic",
        "reform", "global", "election", "protest", "conflict", "threat",
      ];

      let impactScore = 0;
      for (const word of impactfulWords) {
        const pattern = createPattern(word);
        const matches = textLower.match(pattern);
        if (matches) {
          impactScore += matches.length * 5;
        }
      }
      impactScore = Math.min(100, impactScore);

      // Policy probability with word boundaries (uses createPattern from above)
      const policyKeywords = [
        "policy", "regulation", "law", "government", "legislation", "bill",
        "congress", "senate", "parliament", "decree", "treaty", "court",
        "ruling", "initiative",
      ];

      let policyProbability = 0;
      for (const word of policyKeywords) {
        const pattern = createPattern(word);
        const matches = textLower.match(pattern);
        if (matches) {
          policyProbability += matches.length * 10;
        }
      }
      policyProbability = Math.min(100, policyProbability);

      return {
        ...article,
        sentimentScore,
        impactScore,
        policyProbability,
      };
    });

    return NextResponse.json({
      status: 'ok',
      totalResults: data.totalResults,
      articles: articlesWithScores,
    });
  } catch (error) {
    console.error('Error fetching news:', error);
    return NextResponse.json(
      { error: 'Internal server error', message: 'Failed to fetch news' },
      { status: 500 }
    );
  }
}
