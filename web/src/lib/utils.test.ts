import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { humanTime, formatNumber, safeHostname } from './utils';
import { getMarket } from './markets';

describe('humanTime', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-06T12:00:00Z'));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('formats recent times relatively', () => {
    expect(humanTime('2026-07-06T11:59:40Z')).toBe('just now');
    expect(humanTime('2026-07-06T11:55:00Z')).toBe('5m ago');
    expect(humanTime('2026-07-06T09:00:00Z')).toBe('3h ago');
    expect(humanTime('2026-07-04T12:00:00Z')).toBe('2d ago');
  });

  it('falls back to a date for older articles', () => {
    expect(humanTime('2026-06-06T12:00:00Z')).toBe('Jun 6, 2026');
  });

  it('returns the input when unparseable', () => {
    expect(humanTime('not-a-date')).toBe('not-a-date');
  });
});

describe('formatNumber', () => {
  it('adds thousands separators', () => {
    expect(formatNumber(1234567)).toBe('1,234,567');
  });
});

describe('safeHostname', () => {
  it('extracts the hostname without www', () => {
    expect(safeHostname('https://www.reuters.com/markets/story')).toBe('reuters.com');
    expect(safeHostname('https://news.ycombinator.com/item')).toBe('news.ycombinator.com');
  });

  it('never throws on malformed URLs', () => {
    expect(safeHostname('not a url')).toBe('not a url');
  });
});

describe('getMarket', () => {
  it('maps symbols to their exchanges', () => {
    expect(getMarket('NPN')).toBe('JSE');
    expect(getMarket('CBA')).toBe('ASX');
    expect(getMarket('HSBA')).toBe('LSE');
    expect(getMarket('AAPL')).toBe('NYSE/NASDAQ');
  });

  it('labels ambiguous 4-digit codes as HKEX/TSE', () => {
    expect(getMarket('0700')).toBe('HKEX/TSE');
    expect(getMarket('7203')).toBe('HKEX/TSE');
  });
});
