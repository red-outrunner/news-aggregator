'use client';

import { Article } from '@/lib/types';
import { humanTime } from '@/lib/utils';
import { StockMention } from '@/lib/stockExtractor';

interface ArticleCardProps {
  article: Article;
  isBookmarked: boolean;
  onToggleBookmark: (article: Article) => void;
  stockMentions: StockMention[];
}

export default function ArticleCard({ 
  article, 
  isBookmarked, 
  onToggleBookmark,
  stockMentions 
}: ArticleCardProps) {
  const sentimentColor = 
    (article.sentimentScore || 0) > 0 ? 'text-green-600 dark:text-green-400' :
    (article.sentimentScore || 0) < 0 ? 'text-red-600 dark:text-red-400' :
    'text-gray-600 dark:text-gray-400';

  const sentimentLabel = 
    (article.sentimentScore || 0) > 0 ? 'Bullish' :
    (article.sentimentScore || 0) < 0 ? 'Bearish' :
    'Neutral';

  const sentimentBg = 
    (article.sentimentScore || 0) > 0 ? 'bg-green-100 dark:bg-green-900/30' :
    (article.sentimentScore || 0) < 0 ? 'bg-red-100 dark:bg-red-900/30' :
    'bg-gray-100 dark:bg-gray-700';

  return (
    <article className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 
                        hover:shadow-md transition-all duration-200 overflow-hidden">
      <div className="flex flex-col lg:flex-row">
        {/* Image */}
        {article.urlToImage && (
          <div className="lg:w-64 h-48 lg:h-auto flex-shrink-0 relative overflow-hidden">
            <img
              src={article.urlToImage}
              alt={article.title}
              className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
            <div className="absolute top-2 left-2">
              <span className={`${sentimentBg} ${sentimentColor} text-xs font-semibold px-2 py-1 rounded-full`}>
                {sentimentLabel}
              </span>
            </div>
          </div>
        )}
        
        {/* Content */}
        <div className="flex-1 p-5">
          {/* Meta row */}
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-3">
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {humanTime(article.publishedAt)}
              </span>
              <span className="text-gray-300 dark:text-gray-600">•</span>
              <div className="flex items-center gap-2">
                <span className="text-xs font-medium text-purple-600 dark:text-purple-400">
                  Impact: {article.impactScore || 0}
                </span>
                <span className="text-xs text-gray-300 dark:text-gray-600">|</span>
                <span className="text-xs font-medium text-orange-600 dark:text-orange-400">
                  Policy: {article.policyProbability || 0}%
                </span>
              </div>
            </div>

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
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                </svg>
              ) : (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                </svg>
              )}
            </button>
          </div>

          {/* Title */}
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
            <a href={article.url} target="_blank" rel="noopener noreferrer">
              {article.title}
            </a>
          </h2>

          {/* Description */}
          {article.description && (
            <p className="text-gray-600 dark:text-gray-400 text-sm mb-4 line-clamp-3 leading-relaxed">
              {article.description}
            </p>
          )}

          {/* Stock Mentions */}
          {stockMentions.length > 0 && (
            <div className="mb-4 pb-4 border-b border-gray-100 dark:border-gray-700">
              <div className="flex items-center gap-2 mb-2">
                <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
                <span className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Stocks Mentioned
                </span>
              </div>
              <div className="flex flex-wrap gap-2">
                {stockMentions.map((stock) => (
                  <span
                    key={stock.symbol}
                    className="inline-flex items-center gap-1 bg-gray-100 dark:bg-gray-700 
                               text-gray-700 dark:text-gray-300 text-xs font-medium px-2 py-1 rounded"
                  >
                    <span className="font-bold">{stock.symbol}</span>
                    <span className="text-gray-500 dark:text-gray-400 text-xs">{stock.type}</span>
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex items-center justify-between">
            <a
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-blue-600 dark:text-blue-400 
                         hover:text-blue-700 dark:hover:text-blue-300 text-sm font-medium 
                         transition-colors duration-200"
            >
              Read Full Article
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </a>

            <span className="text-xs text-gray-400 dark:text-gray-500">
              Source: {new URL(article.url).hostname}
            </span>
          </div>
        </div>
      </div>
    </article>
  );
}
