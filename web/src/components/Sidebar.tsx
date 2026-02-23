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

      {/* Global Markets Summary */}
      <div className="bg-gradient-to-br from-blue-600 to-purple-700 rounded-lg shadow-sm p-4 text-white">
        <div className="flex items-center gap-2 mb-3">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 className="text-lg font-bold">Global Markets</h3>
        </div>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇺🇸 US (NYSE/NASDAQ)</span>
            <span className="text-xs text-blue-200">Open 9:30 AM EST</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇬🇧 UK (LSE)</span>
            <span className="text-xs text-blue-200">Open 8:00 AM GMT</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇪🇺 EU (XETRA/Euronext)</span>
            <span className="text-xs text-blue-200">Open 9:00 AM CET</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇿🇦 SA (JSE)</span>
            <span className="text-xs text-blue-200">Open 9:00 AM SAST</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇦🇺 AU (ASX)</span>
            <span className="text-xs text-blue-200">Open 10:00 AM AEST</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇭🇰 HK (HKEX)</span>
            <span className="text-xs text-blue-200">Open 9:30 AM HKT</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-blue-100">🇯🇵 JP (TSE)</span>
            <span className="text-xs text-blue-200">Open 9:00 AM JST</span>
          </div>
        </div>
        <div className="mt-3 pt-3 border-t border-white/20">
          <p className="text-xs text-blue-100">
            Stock data updates automatically when articles mention tickers
          </p>
        </div>
      </div>

      {/* Analysis Stats */}
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
          <div className="pt-2 border-t border-gray-200 dark:border-gray-700">
            <div className="flex justify-between items-center">
              <span className="text-xs text-gray-500 dark:text-gray-400">Markets Tracked</span>
              <span className="text-xs font-medium text-blue-600 dark:text-blue-400">8 Regions</span>
            </div>
          </div>
        </div>
      </div>
    </aside>
  );
}
