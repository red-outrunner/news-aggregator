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
  market?: string;
}

interface StockTickerProps {
  mentions: StockMention[];
}

// Get market/region from symbol
function getMarket(symbol: string): string {
  const jsePrefixes = ['AGL', 'ANG', 'ARI', 'BHP', 'BVT', 'CFR', 'CLS', 'CPI', 'DRM', 'EXX', 'FSR', 'GFI', 'GOLD', 'HAR', 'IMP', 'INL', 'INP', 'KIO', 'LHC', 'MNP', 'MRP', 'NPN', 'OMU', 'PIK', 'PPC', 'REM', 'RMI', 'SHP', 'SLM', 'SOL', 'SSW', 'TCG', 'TFG', 'VOD', 'WHL', 'WKP', 'ZMP', 'MTN', 'TKG', 'SBK', 'NED', 'DSY', 'BID', 'BAW', 'DCP', 'MEH', 'NTC', 'LHC', 'APN', 'ADI', 'MSM', 'PPK', 'JDG', 'LEW', 'MTC', 'GRT', 'RDF', 'HYB', 'NEP', 'VKE', 'ACT', 'CAP', 'DIP', 'EMS', 'FAIR'];
  const asxPrefixes = ['CBA', 'WBC', 'NAB', 'ANZ', 'MQG', 'QBE', 'WES', 'FMG', 'STO', 'WDS', 'COL', 'WOW', 'TLS', 'TCL', 'ALL', 'IAG', 'SUN', 'AMP', 'BEN', 'BOQ', 'CSL', 'RMD', 'COH', 'SHL', 'ANN', 'SGP', 'MGR', 'XRO', 'WTC', 'NEC', 'CPU', 'APT', 'ZIP', 'QAN', 'SYD', 'APA', 'ALX', 'AZJ', 'BXB', 'JHX', 'LLC', 'LYC', 'MIN', 'NCM', 'NST', 'ORG', 'PLS', 'SEK', 'TAH', 'TWE', 'VEA', 'WHC', 'YAL'];
  const lseSuffixes = ['.L', 'HSBA', 'BP', 'SHEL', 'AZN', 'GSK', 'DGE', 'ULVR', 'RIO', 'AAL', 'LSEG', 'LLOY', 'BARC', 'VOD', 'BT.A', 'NG', 'GLEN', 'PRU', 'AHT', 'BA', 'CRH', 'FLTR', 'REL', 'EXPN', 'WPP', 'IAG', 'RR', 'CCH', 'TSCO', 'MRO', 'RTO', 'SMDS', 'BRBY', 'JD', 'OCDO', 'PSON', 'WTB', 'MNDI', 'HWDN', 'BDEV'];
  const hkexNumeric = /^\d{4}$/; // HKEX uses 4-digit codes
  
  if (hkexNumeric.test(symbol)) return 'HKEX';
  if (jsePrefixes.includes(symbol)) return 'JSE';
  if (asxPrefixes.includes(symbol)) return 'ASX';
  if (lseSuffixes.some(s => symbol.includes(s))) return 'LSE';
  if (/^\d{4}$/.test(symbol)) return 'TSE'; // Tokyo
  return 'NYSE/NASDAQ';
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
        const stocksWithMarket = (data.stocks || []).map((stock: StockData) => ({
          ...stock,
          market: getMarket(stock.symbol),
        }));
        setStockData(stocksWithMarket);
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
      <div className="flex items-center gap-6">
        <span className="text-sm font-semibold text-gray-400 uppercase tracking-wider flex-shrink-0">
          Global Markets:
        </span>
        <div className="flex gap-4 overflow-x-auto scrollbar-hide">
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
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-bold">{stock.symbol}</span>
                    <span className="text-xs text-gray-400 bg-gray-700/50 px-1.5 py-0.5 rounded">
                      {stock.market}
                    </span>
                  </div>
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
