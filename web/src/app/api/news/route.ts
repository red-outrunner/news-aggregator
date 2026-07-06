import { NextRequest, NextResponse } from 'next/server';
import { analyzeArticle } from '@/lib/analysis';
import { Article } from '@/lib/types';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const query = searchParams.get('q');
  // User-supplied key (from site settings) takes priority over the server's key
  const apiKey = request.headers.get('x-news-api-key') || process.env.NEWS_API_KEY;

  if (!query) {
    return NextResponse.json(
      { error: 'Missing query parameter', message: 'Please provide a search query' },
      { status: 400 }
    );
  }

  if (!apiKey) {
    return NextResponse.json(
      {
        error: 'Missing NewsAPI key',
        message: 'Add your NewsAPI key via the key button in the header (free at newsapi.org/register), or set NEWS_API_KEY in .env.local.',
      },
      { status: 401 }
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
    const articlesWithScores = data.articles.map((article: Article) => ({
      ...article,
      ...analyzeArticle(`${article.title} ${article.description}`),
    }));

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
