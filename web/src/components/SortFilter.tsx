'use client';

interface SortFilterProps {
  sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact';
  onSortChange: (sortBy: 'latest' | 'oldest' | 'sentiment' | 'impact') => void;
  totalResults: number;
  onOpenBookmarks: () => void;
  viewMode: 'list' | 'grid';
  onViewModeChange: (viewMode: 'list' | 'grid') => void;
}

export default function SortFilter({
  sortBy,
  onSortChange,
  totalResults,
  onOpenBookmarks,
  viewMode,
  onViewModeChange
}: SortFilterProps) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
      {/* Results count */}
      <div className="text-sm text-gray-500 dark:text-gray-400">
        {totalResults > 0 ? (
          <span>
            <strong className="text-gray-900 dark:text-gray-100 font-semibold">
              {totalResults.toLocaleString()}
            </strong>{' '}
            articles found
          </span>
        ) : (
          <span>Search for news to get started</span>
        )}
      </div>

      {/* Sort and actions */}
      <div className="flex items-center gap-2">
        {/* View mode toggle */}
        <div className="flex rounded-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-0.5">
          <button
            onClick={() => onViewModeChange('list')}
            className={`p-1.5 rounded-full transition-colors duration-200 ${
              viewMode === 'list'
                ? 'bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900'
                : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
            }`}
            title="List view"
            aria-pressed={viewMode === 'list'}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button
            onClick={() => onViewModeChange('grid')}
            className={`p-1.5 rounded-full transition-colors duration-200 ${
              viewMode === 'grid'
                ? 'bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900'
                : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
            }`}
            title="Grid view"
            aria-pressed={viewMode === 'grid'}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
            </svg>
          </button>
        </div>

        {/* Sort dropdown */}
        <select
          value={sortBy}
          onChange={(e) => onSortChange(e.target.value as typeof sortBy)}
          className="px-3.5 py-2 rounded-full border border-gray-200 dark:border-gray-700
                     bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-200
                     text-sm font-medium focus:outline-none focus:ring-2 focus:ring-blue-500/60
                     cursor-pointer transition-colors"
        >
          <option value="latest">Sort: Latest</option>
          <option value="oldest">Sort: Oldest</option>
          <option value="sentiment">Sort: Sentiment</option>
          <option value="impact">Sort: Impact</option>
        </select>

        {/* Bookmarks button */}
        <button
          onClick={onOpenBookmarks}
          className="px-4 py-2 rounded-full border border-gray-200 dark:border-gray-700
                     bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-200
                     hover:border-blue-400 hover:text-blue-600 dark:hover:text-blue-400
                     text-sm font-medium transition-colors duration-200
                     flex items-center gap-1.5"
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
