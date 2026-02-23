'use client';

interface SortFilterProps {
  sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact';
  onSortChange: (sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact') => void;
  totalResults: number;
  onOpenBookmarks: () => void;
}

export default function SortFilter({ 
  sortBy, 
  onSortChange, 
  totalResults,
  onOpenBookmarks 
}: SortFilterProps) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-4 mb-4">
      {/* Results count */}
      <div className="text-sm text-gray-600 dark:text-gray-400">
        {totalResults > 0 ? (
          <span>Found <strong className="text-gray-900 dark:text-gray-100">{totalResults.toLocaleString()}</strong> articles</span>
        ) : (
          <span>Search for news to get started</span>
        )}
      </div>

      {/* Sort and actions */}
      <div className="flex items-center gap-2">
        {/* Sort dropdown */}
        <select
          value={sortBy}
          onChange={(e) => onSortChange(e.target.value as typeof sortBy)}
          className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 
                     bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100
                     text-sm font-medium focus:outline-none focus:ring-2 focus:ring-blue-500
                     cursor-pointer"
        >
          <option value="latest">Sort: Latest</option>
          <option value="oldest">Sort: Oldest</option>
          <option value="sentiment">Sort: Sentiment</option>
          <option value="impact">Sort: Impact</option>
        </select>

        {/* Bookmarks button */}
        <button
          onClick={onOpenBookmarks}
          className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white 
                     text-sm font-medium rounded-lg transition-colors duration-200
                     flex items-center gap-2"
        >
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
          </svg>
          Bookmarks
        </button>
      </div>
    </div>
  );
}
