'use client';

interface SidebarProps {
  trendingTopics: string[];
  onTopicClick: (topic: string) => void;
}

export default function Sidebar({ trendingTopics, onTopicClick }: SidebarProps) {
  return (
    <aside className="space-y-6">
      {/* Trending Topics */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
        <div className="flex items-center gap-2 mb-4">
          <svg className="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M12.395 2.553a1 1 0 00-1.45-.385c-.345.23-.614.558-.822.88-.214.33-.403.713-.57 1.116-.334.804-.614 1.768-.84 2.734a31.365 31.365 0 00-.613 3.58 2.64 2.64 0 01-.945-1.067c-.328-.68-.398-1.534-.398-2.654A1 1 0 005.05 6.05 6.981 6.981 0 003 11a7 7 0 1011.95-4.95c-.592-.591-.98-.985-1.348-1.467-.363-.476-.724-1.063-1.207-2.03zM12.12 15.12A3 3 0 017 13s.879.5 2.5.5c0-1 .5-4 1.25-4.5.5 1 .786 1.293 1.371 1.879A2.99 2.99 0 0113 13a2.99 2.99 0 01-.879 2.121z" clipRule="evenodd" />
          </svg>
          <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">Trending Topics</h3>
        </div>
        <div className="space-y-2">
          {trendingTopics.length > 0 ? (
            trendingTopics.slice(0, 8).map((topic, index) => (
              <button
                key={topic}
                onClick={() => onTopicClick(topic)}
                className="w-full flex items-center justify-between p-2 rounded-lg 
                           hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors group"
              >
                <div className="flex items-center gap-3">
                  <span className="text-xs font-semibold text-gray-400 w-4">{index + 1}</span>
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300 group-hover:text-blue-600 dark:group-hover:text-blue-400">
                    {topic}
                  </span>
                </div>
                <svg className="w-4 h-4 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-400" 
                     fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            ))
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
              Search for news to see trending topics
            </p>
          )}
        </div>
      </div>

      {/* Market Summary */}
      <div className="bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg shadow-sm p-4 text-white">
        <div className="flex items-center gap-2 mb-3">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <h3 className="text-lg font-bold">Market Summary</h3>
        </div>
        <p className="text-sm text-blue-100 mb-3">
          Track stocks and indices mentioned in your news articles
        </p>
        <div className="text-xs text-blue-200">
          Stock data updates automatically when articles mention tickers
        </div>
      </div>

      {/* Quick Stats */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
        <h3 className="text-sm font-bold text-gray-900 dark:text-gray-100 mb-3">
          Analysis Stats
        </h3>
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <span className="text-xs text-gray-500 dark:text-gray-400">Sentiment Range</span>
            <span className="text-xs font-medium text-gray-700 dark:text-gray-300">-100 to +100</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-xs text-gray-500 dark:text-gray-400">Impact Range</span>
            <span className="text-xs font-medium text-gray-700 dark:text-gray-300">0 to 100</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-xs text-gray-500 dark:text-gray-400">Policy Range</span>
            <span className="text-xs font-medium text-gray-700 dark:text-gray-300">0% to 100%</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
