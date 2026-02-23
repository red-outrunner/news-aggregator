'use client';

import { Article } from '@/lib/types';
import ArticleCard from './ArticleCard';

interface ArticleListProps {
  articles: Article[];
  isBookmarked: (url: string) => boolean;
  onToggleBookmark: (article: Article) => void;
}

export default function ArticleList({ 
  articles, 
  isBookmarked, 
  onToggleBookmark 
}: ArticleListProps) {
  if (articles.length === 0) {
    return (
      <div className="text-center py-12">
        <svg className="w-16 h-16 mx-auto text-gray-400 dark:text-gray-600 mb-4" 
             fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} 
                d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
        </svg>
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
          No articles yet
        </h3>
        <p className="text-gray-600 dark:text-gray-400">
          Search for a topic to start reading news
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {articles.map((article) => (
        <ArticleCard
          key={article.url}
          article={article}
          isBookmarked={isBookmarked(article.url)}
          onToggleBookmark={onToggleBookmark}
        />
      ))}
    </div>
  );
}
