'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { Article } from '@/lib/types';
import { humanTime, safeHostname } from '@/lib/utils';
import { StockMention } from '@/lib/stockExtractor';

// three.js only loads when a reaction chart is first opened
const MarketReaction3D = dynamic(() => import('./MarketReaction3D'), { ssr: false });

interface ArticleCardProps {
  article: Article;
  isBookmarked: boolean;
  onToggleBookmark: (article: Article) => void;
  stockMentions: StockMention[];
  layout?: 'list' | 'grid';
}

export default function ArticleCard({
  article,
  isBookmarked,
  onToggleBookmark,
  stockMentions,
  layout = 'list'
}: ArticleCardProps) {
  const [showReaction, setShowReaction] = useState(false);
  const [imageBroken, setImageBroken] = useState(false);
  const chartableMentions = stockMentions.filter((m) => m.type === 'stock' || m.type === 'etf');

  const score = article.sentimentScore || 0;
  const sentiment =
    score > 0
      ? { label: 'Bullish', badge: 'bg-emerald-500/90 text-white' }
      : score < 0
      ? { label: 'Bearish', badge: 'bg-red-500/90 text-white' }
      : { label: 'Neutral', badge: 'bg-gray-500/80 text-white' };

  const isGrid = layout === 'grid';
  const showImage = article.urlToImage && !imageBroken;

  const image = showImage && (
    <div
      className={`relative overflow-hidden flex-shrink-0 ${
        isGrid ? 'h-44 w-full' : 'lg:w-64 h-48 lg:h-auto'
      }`}
    >
      <img
        src={article.urlToImage}
        alt={article.title}
        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
        onError={() => setImageBroken(true)}
      />
      <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none" />
      <span
        className={`absolute top-3 left-3 ${sentiment.badge} backdrop-blur-sm text-[11px] font-semibold px-2.5 py-1 rounded-full shadow-sm`}
      >
        {sentiment.label}
      </span>
    </div>
  );

  return (
    <article
      className={`group bg-white dark:bg-gray-900 rounded-2xl border border-gray-200/80 dark:border-gray-800
                  shadow-sm hover:shadow-xl hover:-translate-y-0.5 hover:border-gray-300 dark:hover:border-gray-700
                  transition-all duration-300 overflow-hidden ${isGrid ? 'flex flex-col h-full' : ''}`}
    >
      <div className={isGrid ? 'flex flex-col h-full' : 'flex flex-col lg:flex-row'}>
        {image}

        {/* Content */}
        <div className={`flex-1 p-5 ${isGrid ? 'flex flex-col' : ''}`}>
          {/* Meta row */}
          <div className="flex items-center justify-between mb-2.5">
            <div className="flex items-center gap-2 flex-wrap min-w-0">
              {!showImage && (
                <span className={`${sentiment.badge} text-[11px] font-semibold px-2 py-0.5 rounded-full`}>
                  {sentiment.label}
                </span>
              )}
              <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                {humanTime(article.publishedAt)}
              </span>
              <span className="text-[11px] font-medium px-2 py-0.5 rounded-full bg-violet-50 dark:bg-violet-500/10 text-violet-600 dark:text-violet-400 whitespace-nowrap">
                Impact {article.impactScore || 0}
              </span>
              <span className="text-[11px] font-medium px-2 py-0.5 rounded-full bg-amber-50 dark:bg-amber-500/10 text-amber-600 dark:text-amber-400 whitespace-nowrap">
                Policy {article.policyProbability || 0}%
              </span>
            </div>

            {/* Bookmark Button */}
            <button
              onClick={() => onToggleBookmark(article)}
              className={`p-2 rounded-full transition-all duration-200 flex-shrink-0 ml-2 ${
                isBookmarked
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'
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
          <h2
            className={`font-bold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2 tracking-tight
                        group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors ${
                          isGrid ? 'text-base leading-snug' : 'text-xl'
                        }`}
          >
            <a href={article.url} target="_blank" rel="noopener noreferrer">
              {article.title}
            </a>
          </h2>

          {/* Description */}
          {article.description && (
            <p
              className={`text-gray-600 dark:text-gray-400 text-sm mb-4 leading-relaxed ${
                isGrid ? 'line-clamp-2' : 'line-clamp-3'
              }`}
            >
              {article.description}
            </p>
          )}

          {/* Stock Mentions */}
          {stockMentions.length > 0 && (
            <div className={`mb-4 ${isGrid ? 'mt-auto' : ''}`}>
              <div className="flex flex-wrap items-center gap-1.5">
                {stockMentions.slice(0, isGrid ? 3 : 8).map((stock) => (
                  <span
                    key={stock.symbol}
                    className="inline-flex items-center gap-1 bg-gray-100 dark:bg-gray-800
                               text-gray-700 dark:text-gray-300 text-[11px] font-semibold px-2 py-1 rounded-full"
                  >
                    {stock.symbol}
                    <span className="font-normal text-gray-400 dark:text-gray-500">{stock.type}</span>
                  </span>
                ))}
                {chartableMentions.length > 0 && (
                  <button
                    onClick={() => setShowReaction(true)}
                    className="inline-flex items-center gap-1.5 text-[11px] font-semibold px-2.5 py-1 rounded-full
                               bg-gradient-to-r from-blue-600 to-violet-600 hover:from-blue-500 hover:to-violet-500
                               text-white shadow-sm transition-all duration-200"
                    title="Interactive 3D chart of how these stocks moved in the 24h after this article"
                  >
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                        d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                    24h Reaction
                  </button>
                )}
              </div>
            </div>
          )}

          {/* Action Row */}
          <div className={`flex items-center justify-between ${isGrid && stockMentions.length === 0 ? 'mt-auto' : ''}`}>
            <a
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 text-blue-600 dark:text-blue-400
                         hover:gap-2.5 text-sm font-semibold transition-all duration-200"
            >
              Read article
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
              </svg>
            </a>

            <span className="text-xs text-gray-400 dark:text-gray-500 truncate ml-3">
              {safeHostname(article.url)}
            </span>
          </div>
        </div>
      </div>

      {/* 24h market reaction chart (mounted on open) */}
      {showReaction && (
        <MarketReaction3D
          isOpen={showReaction}
          onClose={() => setShowReaction(false)}
          articleTitle={article.title}
          publishedAt={article.publishedAt}
          mentions={chartableMentions}
        />
      )}
    </article>
  );
}
