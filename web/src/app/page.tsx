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
  Sidebar
} from '@/components';
import { extractStocksFromArticles, StockMention } from '@/lib/stockExtractor';
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
    sortBy,
    setArticles,
    setLoading,
    setError,
    toggleBookmark,
    isBookmarked,
    toggleTheme,
    setSortBy,
    getSortedArticles,
  } = useNewsStore();

  const [showBookmarks, setShowBookmarks] = useState(false);
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

    try {
      const response = await fetch(`/api/news?q=${encodeURIComponent(searchQuery)}`);
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
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 shadow-sm">
        <div className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
                </svg>
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">News Aggregator</h1>
                <p className="text-xs text-gray-500 dark:text-gray-400">Smart News Analysis</p>
              </div>
            </div>

            {/* Right side */}
            <div className="flex items-center gap-4">
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
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
              <SearchBar onSearch={handleSearch} isLoading={isLoading} />
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
            />

            {/* Articles List */}
            {isLoading && articles.length === 0 ? (
              <div className="flex items-center justify-center py-16">
                <div className="text-center">
                  <svg className="animate-spin h-10 w-10 text-blue-600 mx-auto mb-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  <p className="text-gray-600 dark:text-gray-400">Fetching latest news...</p>
                </div>
              </div>
            ) : (
              <ArticleList
                articles={sortedArticles}
                isBookmarked={isBookmarked}
                onToggleBookmark={toggleBookmark}
                stockMentionsMap={stockMentionsMap}
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
    </div>
  );
}
