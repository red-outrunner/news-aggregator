import { describe, it, expect } from 'vitest';
import {
  extractStockMentions,
  extractStocksFromArticles,
  expandTickerQuery,
} from './stockExtractor';

const symbolsOf = (text: string) => extractStockMentions(text).map((m) => m.symbol);

describe('extractStockMentions', () => {
  it('finds tickers, company names, and indices', () => {
    const mentions = extractStockMentions('Apple and Tesla rally as AAPL beats the S&P 500');
    const symbols = mentions.map((m) => m.symbol);
    expect(symbols).toContain('AAPL');
    expect(symbols).toContain('TSLA');
    expect(symbols).toContain('S&P 500');
    expect(mentions.find((m) => m.symbol === 'S&P 500')?.type).toBe('index');
    expect(mentions.find((m) => m.symbol === 'AAPL')?.name).toBe('Apple');
  });

  it('does not match symbols inside larger words', () => {
    // "eSTImates" and "scRUTiny" used to produce STI and RUT mentions
    expect(symbolsOf('Analysts raised delivery estimates under regulatory scrutiny')).toEqual([]);
    // "exAMPle" used to produce an AMP mention
    expect(symbolsOf('for example, markets were calm')).toEqual([]);
  });

  it('deduplicates a company found by both name and ticker', () => {
    const mentions = extractStockMentions('Tesla stock TSLA jumped');
    expect(mentions.filter((m) => m.symbol === 'TSLA')).toHaveLength(1);
  });

  it('finds known ETF tickers', () => {
    expect(symbolsOf('Investors poured money into GLD this week')).toContain('GLD');
  });

  it('finds JSE and HKEX symbols', () => {
    const symbols = symbolsOf('Naspers led the JSE All Share higher while Tencent gained');
    expect(symbols).toContain('NPN');
    expect(symbols).toContain('0700');
    expect(symbols).toContain('JSE ALL SHARE');
  });
});

describe('extractStocksFromArticles', () => {
  it('deduplicates symbols across articles', () => {
    const mentions = extractStocksFromArticles([
      { title: 'Tesla surges', description: 'TSLA beats estimates' },
      { title: 'Tesla again', description: 'More TSLA news' },
    ]);
    expect(mentions.filter((m) => m.symbol === 'TSLA')).toHaveLength(1);
  });
});

describe('expandTickerQuery', () => {
  it('expands known tickers to company OR ticker', () => {
    expect(expandTickerQuery('TSLA')?.expandedQuery).toBe('"Tesla" OR TSLA');
    expect(expandTickerQuery('tsla')?.expandedQuery).toBe('"Tesla" OR TSLA');
    expect(expandTickerQuery('shp')?.expandedQuery).toBe('"Shoprite" OR SHP');
    expect(expandTickerQuery('0700')?.expandedQuery).toBe('"Tencent" OR 0700');
  });

  it('leaves non-stock topics untouched', () => {
    expect(expandTickerQuery('climate change')).toBeNull();
    expect(expandTickerQuery('elon musk')).toBeNull();
    expect(expandTickerQuery('xyzzy')).toBeNull();
  });

  it('only expands short ambiguous tickers when typed uppercase', () => {
    expect(expandTickerQuery('bp')).toBeNull();
    expect(expandTickerQuery('BP')).not.toBeNull();
    expect(expandTickerQuery('t')).toBeNull();
    expect(expandTickerQuery('T')?.company).toBe('AT&T');
  });
});
