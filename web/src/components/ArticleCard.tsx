'use client';

import { Article } from '@/lib/types';
import { humanTime } from '@/lib/utils';

interface ArticleCardProps {
  article: Article;
  isBookmarked: boolean;
  onToggleBookmark: (article: Article) => void;
}

export default function ArticleCard({ article, isBookmarked, onToggleBookmark }: ArticleCardProps) {
  const sentimentColor = 
    (article.sentimentScore || 0) > 0 ? 'text-green-600 dark:text-green-400' :
    (article.sentimentScore || 0) < 0 ? 'text-red-600 dark:text-red-400' :
    'text-gray-600 dark:text-gray-400';

  const sentimentLabel = 
    (article.sentimentScore || 0) > 0 ? 'Positive' :
    (article.sentimentScore || 0) < 0 ? 'Negative' :
    'Neutral';

  return (
    <article className="bg-white dark:bg-gray-800 rounded-xl shadow-md overflow-hidden 
                        hover:shadow-lg transition-shadow duration-200">
      <div className="flex flex-col md:flex-row">
        {/* Image */}
        {article.urlToImage && (
          <div className="md:w-48 h-48 md:h-auto flex-shrink-0">
            <img
              src={article.urlToImage}
              alt={article.title}
              className="w-full h-full object-cover"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
        )}
        
        {/* Content */}
        <div className="flex-1 p-4 md:p-5">
          {/* Title */}
          <h2 className="text-lg md:text-xl font-bold text-gray-900 dark:text-gray-100 
                         mb-2 line-clamp-2">
            {article.title}
          </h2>

          {/* Description */}
          {article.description && (
            <p className="text-gray-600 dark:text-gray-400 text-sm mb-3 line-clamp-3">
              {article.description}
            </p>
          )}

          {/* Scores */}
          <div className="flex flex-wrap gap-3 mb-3">
            <span className={`text-sm font-medium ${sentimentColor}`}>
              Sentiment: {sentimentLabel} ({article.sentimentScore || 0})
            </span>
            <span className="text-sm font-medium text-purple-600 dark:text-purple-400">
              Impact: {article.impactScore || 0}
            </span>
            <span className="text-sm font-medium text-orange-600 dark:text-orange-400">
              Policy: {article.policyProbability || 0}%
            </span>
          </div>

          {/* Meta */}
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-500 dark:text-gray-500">
              {humanTime(article.publishedAt)}
            </span>
            
            <div className="flex gap-2">
              {/* Bookmark Button */}
              <button
                onClick={() => onToggleBookmark(article)}
                className={`p-2 rounded-lg transition-colors duration-200 ${
                  isBookmarked
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
                title={isBookmarked ? 'Remove bookmark' : 'Add bookmark'}
              >
                {isBookmarked ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                )}
              </button>

              {/* Read Full Article Link */}
              <a
                href={article.url}
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm 
                           font-medium rounded-lg transition-colors duration-200"
              >
                Read More
              </a>
            </div>
          </div>
        </div>
      </div>
    </article>
  );
}
