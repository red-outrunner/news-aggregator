'use client';

import { useState } from 'react';

interface ApiKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  newsApiKey: string;
  alphaVantageKey: string;
  onSave: (newsApiKey: string, alphaVantageKey: string) => void;
}

export default function ApiKeyModal({
  isOpen,
  onClose,
  newsApiKey,
  alphaVantageKey,
  onSave,
}: ApiKeyModalProps) {
  const [newsKeyInput, setNewsKeyInput] = useState(newsApiKey);
  const [alphaKeyInput, setAlphaKeyInput] = useState(alphaVantageKey);
  const [showKeys, setShowKeys] = useState(false);

  if (!isOpen) return null;

  const handleSave = () => {
    onSave(newsKeyInput, alphaKeyInput);
    onClose();
  };

  const handleClear = () => {
    setNewsKeyInput('');
    setAlphaKeyInput('');
    onSave('', '');
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl
                        w-full max-w-lg overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
              API Keys
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
          <div className="p-4 space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Use your own API keys to fetch news and stock data. Keys are stored
              only in this browser and sent only to this site&apos;s own API.
            </p>

            <div>
              <label htmlFor="news-api-key" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                NewsAPI key
              </label>
              <input
                id="news-api-key"
                type={showKeys ? 'text' : 'password'}
                value={newsKeyInput}
                onChange={(e) => setNewsKeyInput(e.target.value)}
                placeholder="Your NewsAPI key"
                autoComplete="off"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600
                           bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100
                           placeholder-gray-400 dark:placeholder-gray-500
                           focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Free key at{' '}
                <a
                  href="https://newsapi.org/register"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  newsapi.org/register
                </a>
              </p>
            </div>

            <div>
              <label htmlFor="alpha-vantage-key" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Alpha Vantage key <span className="text-gray-400 dark:text-gray-500 font-normal">(optional, for stock quotes)</span>
              </label>
              <input
                id="alpha-vantage-key"
                type={showKeys ? 'text' : 'password'}
                value={alphaKeyInput}
                onChange={(e) => setAlphaKeyInput(e.target.value)}
                placeholder="Your Alpha Vantage key"
                autoComplete="off"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600
                           bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100
                           placeholder-gray-400 dark:placeholder-gray-500
                           focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Free key at{' '}
                <a
                  href="https://www.alphavantage.co/support/#api-key"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  alphavantage.co
                </a>
                {' '}&mdash; without it the ticker falls back to Yahoo Finance
              </p>
            </div>

            <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer">
              <input
                type="checkbox"
                checked={showKeys}
                onChange={(e) => setShowKeys(e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              Show keys
            </label>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={handleClear}
              className="px-4 py-2 text-sm font-medium text-red-600 dark:text-red-400
                         hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg
                         transition-colors duration-200"
            >
              Clear keys
            </button>
            <button
              onClick={handleSave}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white
                         text-sm font-medium rounded-lg transition-colors duration-200"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
