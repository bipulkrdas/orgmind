'use client';

interface ErrorMessageProps {
  message: string;
  onRetry?: () => void;
  showConnectionStatus?: boolean;
}

export function ErrorMessage({ message, onRetry, showConnectionStatus = false }: ErrorMessageProps) {
  return (
    <div 
      className="bg-red-50 border border-red-200 rounded-lg p-4 animate-fadeIn"
      role="alert"
      aria-live="assertive"
    >
      <div className="flex items-start">
        {/* Error Icon */}
        <div className="flex-shrink-0" aria-hidden="true">
          <svg
            className="h-5 w-5 text-red-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>

        {/* Error Content */}
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium text-red-800" id="error-title">
            {showConnectionStatus ? 'Connection Error' : 'Error'}
          </h3>
          <p className="mt-1 text-sm text-red-700" id="error-message">
            {message}
          </p>

          {/* Retry Button */}
          {onRetry && (
            <button
              onClick={onRetry}
              className="mt-3 inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors"
              aria-label="Retry action"
              aria-describedby="error-message"
            >
              <svg
                className="h-4 w-4 mr-1"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
              Retry
            </button>
          )}
        </div>

        {/* Close Button (optional) */}
        <div className="ml-auto pl-3">
          <button
            onClick={onRetry}
            className="inline-flex text-red-400 hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-500 rounded transition-colors"
            aria-label="Dismiss error message"
          >
            <span className="sr-only">Dismiss</span>
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
              <path
                fillRule="evenodd"
                d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                clipRule="evenodd"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
