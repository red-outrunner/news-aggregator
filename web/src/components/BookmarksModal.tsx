'use client';

import { Article } from '@/lib/types';
import { humanTime } from '@/lib/utils';

interface BookmarksModalProps {
  isOpen: boolean;
  onClose: () => void;
  bookmarks: Article[];
  isBookmarked: (url: string) => boolean;
  onToggleBookmark: (article: Article) => void;
}

export default function BookmarksModal({ 
  isOpen, 
  onClose, 
  bookmarks,
  isBookmarked,
  onToggleBookmark 
}: BookmarksModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl 
                        w-full max-w-2xl max-h-[80vh] overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
              Bookmarked Articles
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg 
                         transition-colors duration-200"
            >
              <svg className="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Content */}
          <div className="p-4 overflow-y-auto max-h-[60vh]">
            {bookmarks.length === 0 ? (
              <p className="text-center text-gray-500 dark:text-gray-400 py-8">
                No bookmarked articles yet. Click the bookmark icon on articles to save them here.
              </p>
            ) : (
              <div className="space-y-4">
                {bookmarks.map((article) => (
                  <div
                    key={article.url}
                    className="p-4 bg-gray-50 dark:bg-gray-700 rounded-lg"
                  >
                    <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">
                      {article.title}
                    </h3>
                    {article.description && (
                      <p className="text-sm text-gray-600 dark:text-gray-400 mb-2 line-clamp-2">
                        {article.description}
                      </p>
                    )}
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-gray-500 dark:text-gray-500">
                        {humanTime(article.publishedAt)}
                      </span>
                      <div className="flex gap-2">
                        <button
                          onClick={() => onToggleBookmark(article)}
                          className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white 
                                     text-sm font-medium rounded transition-colors duration-200"
                        >
                          Remove
                        </button>
                        <a
                          href={article.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white 
                                     text-sm font-medium rounded transition-colors duration-200"
                        >
                          Read
                        </a>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
