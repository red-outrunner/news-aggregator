import { NextRequest, NextResponse } from 'next/server';

interface StockData {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  dayHigh: number;
  dayLow: number;
  drawdown: number;
  previousClose: number;
}

/**
 * Fetches stock data from Alpha Vantage API
 * Free tier: 25 requests/day, 5 requests/minute
 */
async function fetchStockDataAlphaVantage(symbol: string): Promise<StockData | null> {
  const apiKey = process.env.ALPHA_VANTAGE_API_KEY;
  
  if (!apiKey) {
    return null;
  }

  try {
    // Get daily data
    const response = await fetch(
      `https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=${symbol}&apikey=${apiKey}`
    );
    
    if (!response.ok) {
      return null;
    }

    const data = await response.json();
    const quote = data['Global Quote'];

    if (!quote || Object.keys(quote).length === 0) {
      return null;
    }

    const price = parseFloat(quote['05. price']) || 0;
    const change = parseFloat(quote['09. change']) || 0;
    const changePercentStr = quote['10. change percent']?.toString() || '';
    const changePercent = parseFloat(changePercentStr.replace('%', '')) || 0;
    const previousClose = parseFloat(quote['08. previous close']) || 0;
    const dayHigh = parseFloat(quote['03. high']) || price;
    const dayLow = parseFloat(quote['04. low']) || price;

    // Calculate drawdown from day high
    const drawdown = dayHigh > 0 ? ((dayHigh - price) / dayHigh) * 100 : 0;

    return {
      symbol,
      name: symbol,
      price,
      change,
      changePercent,
      dayHigh,
      dayLow,
      drawdown,
      previousClose,
    };
  } catch (error) {
    console.error(`Error fetching stock data for ${symbol}:`, error);
    return null;
  }
}

/**
 * Fetches stock data from Yahoo Finance via proxy (alternative)
 * Note: This is for demonstration - in production use a proper API
 */
async function fetchStockDataYahoo(symbol: string): Promise<StockData | null> {
  try {
    // Using a public API wrapper for Yahoo Finance
    const response = await fetch(
      `https://query1.finance.yahoo.com/v8/finance/chart/${symbol}?interval=1d&range=1d`
    );

    if (!response.ok) {
      return null;
    }

    const data = await response.json();
    const result = data.chart.result?.[0];

    if (!result?.meta) {
      return null;
    }

    const meta = result.meta;
    const price = meta.regularMarketPrice || 0;
    const previousClose = meta.previousClose || 0;
    const change = price - previousClose;
    const changePercent = previousClose > 0 ? (change / previousClose) * 100 : 0;
    const dayHigh = meta.regularMarketDayHigh || price;
    const dayLow = meta.regularMarketDayLow || price;
    const drawdown = dayHigh > 0 ? ((dayHigh - price) / dayHigh) * 100 : 0;

    return {
      symbol,
      name: meta.symbol || symbol,
      price,
      change,
      changePercent,
      dayHigh,
      dayLow,
      drawdown,
      previousClose,
    };
  } catch (error) {
    console.error(`Error fetching stock data for ${symbol}:`, error);
    return null;
  }
}

/**
 * GET /api/stock?symbols=AAPL,GOOGL,TSLA
 * Returns stock data for multiple symbols
 */
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const symbolsParam = searchParams.get('symbols');

  if (!symbolsParam) {
    return NextResponse.json(
      { error: 'Missing symbols parameter' },
      { status: 400 }
    );
  }

  const symbols = symbolsParam.split(',').map(s => s.trim().toUpperCase());
  const stockData: StockData[] = [];

  // Fetch data for each symbol (with rate limiting consideration)
  for (const symbol of symbols) {
    // Try Alpha Vantage first
    let data = await fetchStockDataAlphaVantage(symbol);
    
    // Fallback to Yahoo if Alpha Vantage fails or no API key
    if (!data) {
      data = await fetchStockDataYahoo(symbol);
    }

    if (data) {
      stockData.push(data);
    }

    // Rate limiting: wait between requests
    if (symbols.indexOf(symbol) < symbols.length - 1) {
      await new Promise(resolve => setTimeout(resolve, 200));
    }
  }

  return NextResponse.json({
    stocks: stockData,
    timestamp: new Date().toISOString(),
  });
}
