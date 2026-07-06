// Article represents a news article
export interface Article {
  title: string;
  description: string;
  url: string;
  urlToImage: string;
  publishedAt: string;
  sentimentScore?: number;
  impactScore?: number;
  policyProbability?: number;
}

// Stock quote data returned by /api/stock
export interface StockData {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  dayHigh: number;
  dayLow: number;
  drawdown: number;
  previousClose: number;
  market?: string;
}
