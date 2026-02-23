'use client';

import { useEffect, useState } from 'react';
import { StockMention } from '@/lib/stockExtractor';

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

interface StockTickerProps {
  mentions: StockMention[];
}

export default function StockTicker({ mentions }: StockTickerProps) {
  const [stockData, setStockData] = useState<StockData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (mentions.length === 0) {
      setLoading(false);
      return;
    }

    const fetchStockData = async () => {
      setLoading(true);
      const symbols = mentions.map(m => m.symbol).join(',');
      
      try {
        const response = await fetch(`/api/stock?symbols=${symbols}`);
        const data = await response.json();
        setStockData(data.stocks || []);
      } catch (error) {
        console.error('Error fetching stock data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStockData();
  }, [mentions]);

  if (mentions.length === 0 || (!loading && stockData.length === 0)) {
    return null;
  }

  return (
    <div className="bg-gradient-to-r from-gray-900 to-gray-800 text-white py-3 px-4 overflow-hidden">
      <div className="flex items-center gap-6 animate-scroll">
        <span className="text-sm font-semibold text-gray-400 uppercase tracking-wider flex-shrink-0">
          Market Mentioned:
        </span>
        <div className="flex gap-6 overflow-x-auto scrollbar-hide">
          {loading ? (
            <div className="flex gap-2">
              <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
              <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
              <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
            </div>
          ) : (
            stockData.map((stock) => (
              <div
                key={stock.symbol}
                className="flex items-center gap-3 bg-white/10 backdrop-blur-sm rounded-lg px-4 py-2 flex-shrink-0 hover:bg-white/15 transition-colors"
              >
                <div className="text-left">
                  <div className="text-sm font-bold">{stock.symbol}</div>
                  <div className="text-xs text-gray-400">${stock.price.toFixed(2)}</div>
                </div>
                <div className="text-right">
                  <div className={`text-sm font-semibold ${stock.changePercent >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                    {stock.changePercent >= 0 ? '+' : ''}{stock.changePercent.toFixed(2)}%
                  </div>
                  <div className="text-xs text-gray-400">
                    DD: {stock.drawdown.toFixed(2)}%
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
