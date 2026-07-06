'use client';

import { useState, useEffect, useMemo } from 'react';
import { useNewsStore } from '@/store/newsStore';
import {
  SearchBar,
  ArticleList,
  SortFilter,
  BookmarksModal,
  ThemeToggle,
  StockTicker,
  Sidebar,
  ApiKeyModal
} from '@/components';
import { extractStocksFromArticles, expandTickerQuery, StockMention, TickerExpansion } from '@/lib/stockExtractor';
import { Article } from '@/lib/types';

export default function Home() {
  const {
    articles,
    totalResults,
    query,
    isLoading,
    error,
    bookmarks,
    isDarkMode,
    newsApiKey,
    alphaVantageKey,
    viewMode,
    sortBy,
    setViewMode,
    setArticles,
    setLoading,
    setError,
    toggleBookmark,
    isBookmarked,
    toggleTheme,
    setApiKeys,
    setSortBy,
    getSortedArticles,
  } = useNewsStore();

  const [showBookmarks, setShowBookmarks] = useState(false);
  const [showApiKeys, setShowApiKeys] = useState(false);
  const [tickerExpansion, setTickerExpansion] = useState<TickerExpansion | null>(null);
  const [stockMentionsMap, setStockMentionsMap] = useState<Map<string, StockMention[]>>(new Map());
  const [allStockMentions, setAllStockMentions] = useState<StockMention[]>([]);

  // Apply dark mode class to html element
  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDarkMode]);

  // Extract stock mentions from articles
  useEffect(() => {
    if (articles.length === 0) {
      setStockMentionsMap(new Map());
      setAllStockMentions([]);
      return;
    }

    const mentionsMap = new Map<string, StockMention[]>();
    const allMentionsSet = new Map<string, StockMention>();

    for (const article of articles) {
      const content = `${article.title} ${article.description}`;
      const mentions = extractStocksFromArticles([{ title: article.title, description: article.description }]);
      mentionsMap.set(article.url, mentions);

      for (const mention of mentions) {
        if (!allMentionsSet.has(mention.symbol)) {
          allMentionsSet.set(mention.symbol, mention);
        }
      }
    }

    setStockMentionsMap(mentionsMap);
    setAllStockMentions(Array.from(allMentionsSet.values()));
  }, [articles]);

  // Extract trending topics from query and articles
  const trendingTopics = useMemo(() => {
    if (!query) return [];
    const topics = new Set<string>();
    topics.add(query);
    
    // Add related topics based on article content
    for (const article of articles.slice(0, 10)) {
      const words = article.title.split(' ').filter(w => w.length > 4);
      words.slice(0, 3).forEach(w => topics.add(w.replace(/[^a-zA-Z]/g, '')));
    }
    
    return Array.from(topics).slice(0, 10);
  }, [query, articles]);

  const handleSearch = async (searchQuery: string) => {
    setLoading(true);
    setError(null);

    // If the query is a stock ticker, widen the search to the company name too
    const expansion = expandTickerQuery(searchQuery);
    setTickerExpansion(expansion);
    const apiQuery = expansion ? expansion.expandedQuery : searchQuery;

    try {
      const response = await fetch(`/api/news?q=${encodeURIComponent(apiQuery)}`, {
        headers: newsApiKey ? { 'X-News-Api-Key': newsApiKey } : undefined,
      });
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to fetch news');
      }

      setArticles(data.articles, data.totalResults, searchQuery);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const handleTopicClick = (topic: string) => {
    if (topic && topic !== query) {
      handleSearch(topic);
    }
  };

  const sortedArticles = getSortedArticles();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors duration-200">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white/75 dark:bg-gray-950/75 backdrop-blur-xl border-b border-gray-200/70 dark:border-gray-800/70">
        <div className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-violet-600 rounded-xl flex items-center justify-center shadow-lg shadow-blue-600/20">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
                </svg>
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 tracking-tight">News Aggregator</h1>
                <p className="text-xs text-gray-500 dark:text-gray-400">Smart News Analysis</p>
              </div>
            </div>

            {/* Right side */}
            <div className="flex items-center gap-4">
              <button
                onClick={() => setShowApiKeys(true)}
                className="relative p-2 rounded-lg bg-gray-100 dark:bg-gray-700
                           hover:bg-gray-200 dark:hover:bg-gray-600
                           transition-colors duration-200"
                title={newsApiKey ? 'API keys (your key is active)' : 'Add your API keys'}
              >
                <svg className="w-5 h-5 text-gray-700 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
                {newsApiKey && (
                  <span className="absolute top-1 right-1 w-2 h-2 bg-green-500 rounded-full" />
                )}
              </button>
              <ThemeToggle isDarkMode={isDarkMode} onToggle={toggleTheme} />
            </div>
          </div>
        </div>
      </header>

      {/* Stock Ticker */}
      <StockTicker mentions={allStockMentions} />

      {/* Main Content */}
      <div className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Main Content Area */}
          <div className="lg:col-span-3 space-y-6">
            {/* Search */}
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200/80 dark:border-gray-800 p-4">
              <SearchBar onSearch={handleSearch} isLoading={isLoading} />
              {tickerExpansion && !error && (
                <p className="mt-2 text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1.5">
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                  </svg>
                  Ticker detected — also searching{' '}
                  <span className="font-semibold text-gray-700 dark:text-gray-300">
                    {tickerExpansion.company} ({tickerExpansion.symbol})
                  </span>
                </p>
              )}
            </div>

            {/* Error Message */}
            {error && (
              <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <p className="text-red-600 dark:text-red-400">{error}</p>
              </div>
            )}

            {/* Sort and Filter */}
            <SortFilter
              sortBy={sortBy}
              onSortChange={setSortBy}
              totalResults={totalResults}
              onOpenBookmarks={() => setShowBookmarks(true)}
              viewMode={viewMode}
              onViewModeChange={setViewMode}
            />

            {/* Articles List */}
            {isLoading && articles.length === 0 ? (
              <div className="space-y-5" aria-label="Loading articles">
                {[0, 1, 2].map((i) => (
                  <div
                    key={i}
                    className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200/80 dark:border-gray-800 overflow-hidden"
                  >
                    <div className="flex flex-col lg:flex-row animate-pulse">
                      <div className="lg:w-64 h-48 bg-gray-200 dark:bg-gray-800 flex-shrink-0" />
                      <div className="flex-1 p-5 space-y-3.5">
                        <div className="h-3 w-44 bg-gray-200 dark:bg-gray-800 rounded-full" />
                        <div className="h-5 w-3/4 bg-gray-200 dark:bg-gray-800 rounded-full" />
                        <div className="h-3 w-full bg-gray-200 dark:bg-gray-800 rounded-full" />
                        <div className="h-3 w-2/3 bg-gray-200 dark:bg-gray-800 rounded-full" />
                        <div className="flex gap-2 pt-1">
                          <div className="h-6 w-16 bg-gray-200 dark:bg-gray-800 rounded-full" />
                          <div className="h-6 w-16 bg-gray-200 dark:bg-gray-800 rounded-full" />
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <ArticleList
                articles={sortedArticles}
                isBookmarked={isBookmarked}
                onToggleBookmark={toggleBookmark}
                stockMentionsMap={stockMentionsMap}
                viewMode={viewMode}
              />
            )}
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1">
            <div className="sticky top-24">
              <Sidebar 
                trendingTopics={trendingTopics}
                onTopicClick={handleTopicClick}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Bookmarks Modal */}
      <BookmarksModal
        isOpen={showBookmarks}
        onClose={() => setShowBookmarks(false)}
        bookmarks={bookmarks}
        isBookmarked={isBookmarked}
        onToggleBookmark={toggleBookmark}
      />

      {/* API Keys Modal (mounted on open so inputs pick up saved keys) */}
      {showApiKeys && (
        <ApiKeyModal
          isOpen={showApiKeys}
          onClose={() => setShowApiKeys(false)}
          newsApiKey={newsApiKey}
          alphaVantageKey={alphaVantageKey}
          onSave={setApiKeys}
        />
      )}
    </div>
  );
}
