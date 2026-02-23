// Stock and Index extraction utility
// Extracts ticker symbols and index names from article text

// Major US stock indices
const indices = [
  'S&P 500', 'SPX', 'Dow Jones', 'DJIA', 'NASDAQ', 'NDX', 'Russell 2000', 'RUT',
  'FTSE 100', 'DAX', 'CAC 40', 'Nikkei 225', 'Hang Seng', 'Shanghai Composite',
  'JSE All Share', 'JALSH', 'Top 40', 'FTSE JSE',
  'VIX', 'Volatility Index',
];

// Common stock ticker patterns
const tickerPatterns = [
  /\b[A-Z]{1,5}\b/g, // Standard tickers (1-5 uppercase letters)
];

// Known tickers to look for (major companies)
const knownTickers = new Set([
  // Tech
  'AAPL', 'MSFT', 'GOOGL', 'GOOG', 'AMZN', 'META', 'TSLA', 'NVDA', 'AMD', 'INTC',
  'NFLX', 'ORCL', 'CRM', 'ADBE', 'CSCO', 'AVGO', 'QCOM', 'TXN', 'IBM', 'NOW',
  // Finance
  'JPM', 'BAC', 'WFC', 'GS', 'MS', 'C', 'BLK', 'SCHW', 'AXP', 'V', 'MA', 'PYPL',
  // Healthcare
  'JNJ', 'UNH', 'PFE', 'MRK', 'ABBV', 'TMO', 'ABT', 'DHR', 'BMY', 'LLY', 'GILD',
  // Consumer
  'WMT', 'PG', 'KO', 'PEP', 'COST', 'HD', 'MCD', 'NKE', 'SBUX', 'TGT', 'LOW',
  // Energy
  'XOM', 'CVX', 'COP', 'SLB', 'EOG', 'MPC', 'PSX', 'VLO', 'OXY', 'HAL',
  // Industrial
  'CAT', 'BA', 'HON', 'UPS', 'GE', 'MMM', 'LMT', 'RTX', 'DE', 'UNP',
  // Telecom
  'T', 'VZ', 'TMUS', 'CHTR', 'CMCSA', 'DIS', 'CMCSA',
  // Real Estate
  'AMT', 'PLD', 'CCI', 'EQIX', 'SPG', 'PSA', 'WELL', 'DLR',
  // Materials
  'LIN', 'APD', 'SHW', 'ECL', 'FCX', 'NEM', 'DOW', 'DD', 'PPG',
  // South African (JSE)
  'AGL', 'ANG', 'APN', 'ARI', 'BHP', 'BVT', 'CFR', 'CLS', 'CPI', 'DRM',
  'EXX', 'FSR', 'GFI', 'GOLD', 'HCG', 'IMP', 'INL', 'INP', 'KIO', 'LHC',
  'MNP', 'MRM', 'NPN', 'OMU', 'PIK', 'PPC', 'REM', 'RMI', 'SHP', 'SLM',
  'SOL', 'SSW', 'TCG', 'TFG', 'VOD', 'WHL', 'WKP', 'ZMP',
  // ETFs
  'SPY', 'QQQ', 'DIA', 'IWM', 'VTI', 'VOO', 'VEA', 'VWO', 'AGG', 'BND',
  'GLD', 'SLV', 'USO', 'UNG', 'TLT', 'HYG', 'LQD', 'XLF', 'XLE', 'XLK',
]);

// Company name to ticker mapping (for text mentions)
const companyToTicker: Record<string, string> = {
  'Apple': 'AAPL', 'Microsoft': 'MSFT', 'Google': 'GOOGL', 'Amazon': 'AMZN',
  'Meta': 'META', 'Facebook': 'META', 'Tesla': 'TSLA', 'Nvidia': 'NVDA',
  'JPMorgan': 'JPM', 'Bank of America': 'BAC', 'Wells Fargo': 'WFC',
  'Goldman Sachs': 'GS', 'Morgan Stanley': 'MS', 'BlackRock': 'BLK',
  'Johnson & Johnson': 'JNJ', 'Pfizer': 'PFE', 'Merck': 'MRK',
  'Walmart': 'WMT', 'Procter & Gamble': 'PG', 'Coca-Cola': 'KO',
  'ExxonMobil': 'XOM', 'Chevron': 'CVX', 'ConocoPhillips': 'COP',
  'Boeing': 'BA', 'Caterpillar': 'CAT', 'General Electric': 'GE',
  'AT&T': 'T', 'Verizon': 'VZ', 'Comcast': 'CMCSA', 'Disney': 'DIS',
  'Gold Fields': 'GFI', 'Anglo American': 'AGL', 'Naspers': 'NPN',
  'Prosus': 'PRX', 'Richemont': 'CFR', 'FirstRand': 'FSR',
  'Standard Bank': 'SBK', 'Capitec': 'CPI', 'Vodacom': 'VOD',
  'MTN': 'MTN', 'Telkom': 'TKG', 'Eskom': 'ESK',
  'Sasol': 'SOL', 'AngloGold Ashanti': 'ANG', 'Harmony Gold': 'HAR',
};

export interface StockMention {
  symbol: string;
  name: string;
  type: 'stock' | 'index' | 'etf';
  context: string;
}

/**
 * Extracts stock tickers and index mentions from text
 */
export function extractStockMentions(text: string): StockMention[] {
  const mentions: StockMention[] = [];
  const foundSymbols = new Set<string>();
  const textLower = text.toLowerCase();

  // Check for index mentions
  for (const index of indices) {
    if (textLower.includes(index.toLowerCase())) {
      mentions.push({
        symbol: index.toUpperCase(),
        name: index,
        type: 'index',
        context: '',
      });
      foundSymbols.add(index.toUpperCase());
    }
  }

  // Check for company names and map to tickers
  for (const [company, ticker] of Object.entries(companyToTicker)) {
    if (!foundSymbols.has(ticker) && textLower.includes(company.toLowerCase())) {
      mentions.push({
        symbol: ticker,
        name: company,
        type: 'stock',
        context: '',
      });
      foundSymbols.add(ticker);
    }
  }

  // Look for ticker symbols in the text (uppercase 1-5 letters)
  const tickerRegex = /\b[A-Z]{1,5}\b/g;
  const matches = text.match(tickerRegex);
  
  if (matches) {
    for (const match of matches) {
      if (!foundSymbols.has(match) && knownTickers.has(match)) {
        mentions.push({
          symbol: match,
          name: match,
          type: 'stock',
          context: '',
        });
        foundSymbols.add(match);
      }
    }
  }

  // Check for ETF patterns
  const etfPatterns = ['ETF', 'ETN', 'Fund'];
  for (const pattern of etfPatterns) {
    const etfRegex = new RegExp(`\\b([A-Z]{2,4})\\s*${pattern}\\b`, 'gi');
    const etfMatches = text.match(etfRegex);
    if (etfMatches) {
      for (const etfMatch of etfMatches) {
        const symbolMatch = etfMatch.match(/[A-Z]{2,4}/);
        if (symbolMatch && !foundSymbols.has(symbolMatch[0])) {
          mentions.push({
            symbol: symbolMatch[0],
            name: etfMatch,
            type: 'etf',
            context: '',
          });
          foundSymbols.add(symbolMatch[0]);
        }
      }
    }
  }

  return mentions;
}

/**
 * Extracts unique stock symbols from an array of articles
 */
export function extractStocksFromArticles(articles: Array<{ title: string; description: string }>): StockMention[] {
  const allMentions = new Map<string, StockMention>();

  for (const article of articles) {
    const content = `${article.title} ${article.description}`;
    const mentions = extractStockMentions(content);
    
    for (const mention of mentions) {
      if (!allMentions.has(mention.symbol)) {
        allMentions.set(mention.symbol, mention);
      }
    }
  }

  return Array.from(allMentions.values());
}
