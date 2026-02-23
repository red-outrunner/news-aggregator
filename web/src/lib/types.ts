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

// NewsAPI response structure
export interface NewsResponse {
  status: string;
  totalResults: number;
  articles: Article[];
}

// API request/response types
export interface NewsRequest {
  query: string;
  page?: number;
  fromDate?: string;
  toDate?: string;
}

export interface NewsResult {
  articles: Article[];
  totalResults: number;
}
