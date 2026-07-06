'use client';

import { Article } from '@/lib/types';
import ArticleCard from './ArticleCard';
import { StockMention } from '@/lib/stockExtractor';

interface ArticleListProps {
  articles: Article[];
  isBookmarked: (url: string) => boolean;
  onToggleBookmark: (article: Article) => void;
  stockMentionsMap: Map<string, StockMention[]>;
  viewMode?: 'list' | 'grid';
}

export default function ArticleList({
  articles,
  isBookmarked,
  onToggleBookmark,
  stockMentionsMap,
  viewMode = 'list'
}: ArticleListProps) {
  if (articles.length === 0) {
    return (
      <div className="text-center py-20 px-6 bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-300 dark:border-gray-700">
        <div className="w-16 h-16 mx-auto mb-5 rounded-2xl bg-gradient-to-br from-blue-600 to-violet-600 flex items-center justify-center shadow-lg shadow-blue-600/20">
          <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
                  d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
          </svg>
        </div>
        <h3 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-2 tracking-tight">
          No articles yet
        </h3>
        <p className="text-gray-500 dark:text-gray-400 max-w-md mx-auto text-sm leading-relaxed">
          Search any topic or stock ticker to start reading news with sentiment analysis,
          stock mentions, and 24h market reaction charts.
        </p>
      </div>
    );
  }

  return (
    <div
      className={
        viewMode === 'grid'
          ? 'grid grid-cols-1 sm:grid-cols-2 gap-5 items-stretch'
          : 'space-y-5'
      }
    >
      {articles.map((article) => (
        <ArticleCard
          key={article.url}
          article={article}
          isBookmarked={isBookmarked(article.url)}
          onToggleBookmark={onToggleBookmark}
          stockMentions={stockMentionsMap.get(article.url) || []}
          layout={viewMode}
        />
      ))}
    </div>
  );
}
