import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Article } from '@/lib/types';

interface NewsState {
  // Articles
  articles: Article[];
  totalResults: number;
  query: string;
  isLoading: boolean;
  error: string | null;

  // Bookmarks (persisted)
  bookmarks: Article[];

  // Theme (persisted)
  isDarkMode: boolean;

  // User-supplied API keys (persisted, stored only in this browser)
  newsApiKey: string;
  alphaVantageKey: string;

  // Layout (persisted)
  viewMode: 'list' | 'grid';

  // Sorting
  sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact';

  // Actions
  setArticles: (articles: Article[], totalResults: number, query: string) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Bookmarks
  toggleBookmark: (article: Article) => void;
  isBookmarked: (url: string) => boolean;
  removeBookmark: (url: string) => void;
  
  // Theme
  toggleTheme: () => void;

  // API keys
  setApiKeys: (newsApiKey: string, alphaVantageKey: string) => void;

  // Layout
  setViewMode: (viewMode: 'list' | 'grid') => void;

  // Sorting
  setSortBy: (sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact') => void;
  getSortedArticles: () => Article[];
}

export const useNewsStore = create<NewsState>()(
  persist(
    (set, get) => ({
      // Initial state
      articles: [],
      totalResults: 0,
      query: '',
      isLoading: false,
      error: null,
      bookmarks: [],
      isDarkMode: false,
      newsApiKey: '',
      alphaVantageKey: '',
      viewMode: 'list',
      sortBy: 'latest',

      // Actions
      setArticles: (articles, totalResults, query) =>
        set({ articles, totalResults, query, error: null }),

      setLoading: (loading) => set({ isLoading: loading }),

      setError: (error) => set({ error }),

      // Bookmarks
      toggleBookmark: (article) => {
        const bookmarks = get().bookmarks;
        const exists = bookmarks.some((bm) => bm.url === article.url);
        
        if (exists) {
          set({ bookmarks: bookmarks.filter((bm) => bm.url !== article.url) });
        } else {
          set({ bookmarks: [...bookmarks, article] });
        }
      },

      isBookmarked: (url) => {
        return get().bookmarks.some((bm) => bm.url === url);
      },

      removeBookmark: (url) => {
        set({ bookmarks: get().bookmarks.filter((bm) => bm.url !== url) });
      },

      // Theme
      toggleTheme: () => {
        set({ isDarkMode: !get().isDarkMode });
      },

      // API keys
      setApiKeys: (newsApiKey, alphaVantageKey) => {
        set({ newsApiKey: newsApiKey.trim(), alphaVantageKey: alphaVantageKey.trim() });
      },

      // Layout
      setViewMode: (viewMode) => set({ viewMode }),

      // Sorting
      setSortBy: (sortBy) => set({ sortBy }),

      getSortedArticles: () => {
        const articles = [...get().articles];
        const sortBy = get().sortBy;

        switch (sortBy) {
          case 'latest':
            articles.sort((a, b) => 
              new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime()
            );
            break;
          case 'oldest':
            articles.sort((a, b) => 
              new Date(a.publishedAt).getTime() - new Date(b.publishedAt).getTime()
            );
            break;
          case 'sentiment':
            articles.sort((a, b) => 
              (b.sentimentScore || 0) - (a.sentimentScore || 0)
            );
            break;
          case 'impact':
            articles.sort((a, b) => 
              (b.impactScore || 0) - (a.impactScore || 0)
            );
            break;
        }

        return articles;
      },
    }),
    {
      name: 'news-aggregator-storage',
      partialize: (state) => ({
        bookmarks: state.bookmarks,
        isDarkMode: state.isDarkMode,
        newsApiKey: state.newsApiKey,
        alphaVantageKey: state.alphaVantageKey,
        viewMode: state.viewMode,
      }),
    }
  )
);
