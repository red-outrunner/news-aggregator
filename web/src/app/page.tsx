'use client';

import { useState, useEffect } from 'react';
import { useNewsStore } from '@/store/newsStore';
import { SearchBar, ArticleList, SortFilter, BookmarksModal, ThemeToggle } from '@/components';

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

  // Apply dark mode class to html element
  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDarkMode]);

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

  const sortedArticles = getSortedArticles();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
              News Aggregator
            </h1>
            <ThemeToggle isDarkMode={isDarkMode} onToggle={toggleTheme} />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Search */}
        <div className="mb-8">
          <SearchBar onSearch={handleSearch} isLoading={isLoading} />
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
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
          <div className="flex items-center justify-center py-12">
            <svg className="animate-spin h-8 w-8 text-blue-600" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
          </div>
        ) : (
          <ArticleList
            articles={sortedArticles}
            isBookmarked={isBookmarked}
            onToggleBookmark={toggleBookmark}
          />
        )}
      </main>

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
