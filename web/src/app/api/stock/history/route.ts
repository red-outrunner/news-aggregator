import { NextRequest, NextResponse } from 'next/server';

export interface HistoryPoint {
  t: number; // epoch ms
  price: number;
}

export interface SymbolHistory {
  symbol: string;
  points: HistoryPoint[];
  publishPrice: number | null;
  changePercent: number | null;
}

/**
 * Fetches intraday prices for one symbol from Yahoo Finance between two timestamps
 */
async function fetchHistory(symbol: string, fromMs: number, toMs: number, interval: string): Promise<HistoryPoint[]> {
  const url =
    `https://query1.finance.yahoo.com/v8/finance/chart/${encodeURIComponent(symbol)}` +
    `?period1=${Math.floor(fromMs / 1000)}&period2=${Math.floor(toMs / 1000)}&interval=${interval}`;

  const response = await fetch(url);
  if (!response.ok) {
    return [];
  }

  const data = await response.json();
  const result = data.chart?.result?.[0];
  const timestamps: number[] = result?.timestamp || [];
  const closes: (number | null)[] = result?.indicators?.quote?.[0]?.close || [];

  const points: HistoryPoint[] = [];
  for (let i = 0; i < timestamps.length; i++) {
    const price = closes[i];
    if (price != null && Number.isFinite(price)) {
      points.push({ t: timestamps[i] * 1000, price });
    }
  }
  return points;
}

/**
 * GET /api/stock/history?symbols=AAPL,TSLA&from=2026-07-05T14:00:00Z
 * Returns per-symbol prices for the 24 hours following `from`
 * (how each stock performed since the article was published)
 */
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const symbolsParam = searchParams.get('symbols');
  const fromParam = searchParams.get('from');

  if (!symbolsParam || !fromParam) {
    return NextResponse.json(
      { error: 'Missing parameters', message: 'Provide symbols and from (article publish time)' },
      { status: 400 }
    );
  }

  const fromMs = Date.parse(fromParam);
  if (Number.isNaN(fromMs)) {
    return NextResponse.json(
      { error: 'Invalid from parameter', message: 'from must be an ISO date' },
      { status: 400 }
    );
  }

  const toMs = Math.min(fromMs + 24 * 60 * 60 * 1000, Date.now());
  if (toMs <= fromMs) {
    return NextResponse.json(
      { error: 'Invalid window', message: 'Article publish time is in the future' },
      { status: 400 }
    );
  }

  const symbols = symbolsParam
    .split(',')
    .map((s) => s.trim().toUpperCase())
    .filter(Boolean)
    .slice(0, 4); // keep the scene and upstream load bounded

  const histories: SymbolHistory[] = await Promise.all(
    symbols.map(async (symbol) => {
      try {
        // 15m bars for recent articles; Yahoo only serves them ~60 days back,
        // so fall back to hourly bars for older publish times
        let points = await fetchHistory(symbol, fromMs, toMs, '15m');
        if (points.length < 2) {
          points = await fetchHistory(symbol, fromMs, toMs, '60m');
        }

        const publishPrice = points.length > 0 ? points[0].price : null;
        const lastPrice = points.length > 0 ? points[points.length - 1].price : null;
        const changePercent =
          publishPrice != null && lastPrice != null && publishPrice > 0
            ? ((lastPrice - publishPrice) / publishPrice) * 100
            : null;

        return { symbol, points, publishPrice, changePercent };
      } catch (error) {
        console.error(`Error fetching history for ${symbol}:`, error);
        return { symbol, points: [], publishPrice: null, changePercent: null };
      }
    })
  );

  return NextResponse.json({
    from: new Date(fromMs).toISOString(),
    to: new Date(toMs).toISOString(),
    histories,
  });
}
