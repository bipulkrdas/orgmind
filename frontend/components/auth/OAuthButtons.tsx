'use client';

import { useState } from 'react';
import { initiateOAuth } from '@/lib/api/auth';
import type { OAuthProvider } from '@/lib/types';

interface OAuthButtonsProps {
  onError?: (error: Error) => void;
}

export default function OAuthButtons({ onError }: OAuthButtonsProps) {
  const [loadingProvider, setLoadingProvider] = useState<OAuthProvider | null>(null);

  const handleOAuthClick = async (provider: OAuthProvider) => {
    setLoadingProvider(provider);

    try {
      const { url } = await initiateOAuth(provider);
      // Redirect to OAuth provider's authorization page
      window.location.href = url;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'OAuth initiation failed';
      onError?.(error instanceof Error ? error : new Error(errorMessage));
      setLoadingProvider(null);
    }
  };

  const providers: Array<{ id: OAuthProvider; name: string; icon: string }> = [
    { id: 'google', name: 'Google', icon: '🔍' },
    { id: 'okta', name: 'Okta', icon: '🔐' },
    { id: 'office365', name: 'Office 365', icon: '📧' },
  ];

  return (
    <div className="space-y-3">
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-gray-300" />
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-gray-500">Or continue with</span>
        </div>
      </div>

      <div className="space-y-2">
        {providers.map((provider) => (
          <button
            key={provider.id}
            type="button"
            onClick={() => handleOAuthClick(provider.id)}
            disabled={loadingProvider !== null}
            className="w-full flex items-center justify-center gap-3 px-4 py-2 border border-gray-300 rounded-md bg-white text-gray-700 font-medium hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="text-xl">{provider.icon}</span>
            <span>
              {loadingProvider === provider.id
                ? `Connecting to ${provider.name}...`
                : `Sign in with ${provider.name}`}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
