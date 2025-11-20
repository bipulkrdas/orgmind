'use client';

import { useEffect, useState } from 'react';

/**
 * NetworkStatus Component
 * 
 * Displays a banner when the user goes offline and dismisses when back online.
 * Uses the browser's online/offline events for real-time network status detection.
 */
export function NetworkStatus() {
  const [isOnline, setIsOnline] = useState(true);
  const [showOfflineBanner, setShowOfflineBanner] = useState(false);

  useEffect(() => {
    // Set initial state
    setIsOnline(navigator.onLine);

    // Handle online event
    const handleOnline = () => {
      setIsOnline(true);
      setShowOfflineBanner(false);
    };

    // Handle offline event
    const handleOffline = () => {
      setIsOnline(false);
      setShowOfflineBanner(true);
    };

    // Add event listeners
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Cleanup
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  // Don't render anything if online
  if (isOnline && !showOfflineBanner) {
    return null;
  }

  return (
    <div
      className="fixed top-0 left-0 right-0 z-50 animate-slideDown"
      role="alert"
      aria-live="assertive"
    >
      {!isOnline ? (
        // Offline banner
        <div className="bg-yellow-50 border-b border-yellow-200 px-4 py-3">
          <div className="flex items-center justify-center space-x-2">
            <svg
              className="h-5 w-5 text-yellow-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <p className="text-sm font-medium text-yellow-800">
              You are currently offline. Some features may not work.
            </p>
          </div>
        </div>
      ) : (
        // Back online banner (auto-dismisses)
        <div className="bg-green-50 border-b border-green-200 px-4 py-3">
          <div className="flex items-center justify-center space-x-2">
            <svg
              className="h-5 w-5 text-green-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p className="text-sm font-medium text-green-800">
              You are back online
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
