'use client';

import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { handleOAuthCallback } from '@/lib/api/auth';
import type { OAuthProvider } from '@/lib/types';

function OAuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string>('');
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    const processCallback = async () => {
      try {
        // Extract OAuth code and provider from URL parameters
        const code = searchParams?.get('code');
        const state = searchParams?.get('state');
        const errorParam = searchParams?.get('error');

        // Check for OAuth errors
        if (errorParam) {
          setError(`Authentication failed: ${errorParam}`);
          setIsProcessing(false);
          return;
        }

        // Validate required parameters
        if (!code) {
          setError('Missing authorization code');
          setIsProcessing(false);
          return;
        }

        if (!state) {
          setError('Missing state parameter');
          setIsProcessing(false);
          return;
        }

        // Extract provider from state (format: "provider:random")
        const provider = state.split(':')[0] as OAuthProvider;
        
        if (!['google', 'okta', 'office365'].includes(provider)) {
          setError('Invalid OAuth provider');
          setIsProcessing(false);
          return;
        }

        // Call OAuth callback API to complete authentication
        await handleOAuthCallback(provider, code);

        // Redirect to home page after successful authentication
        router.push('/home');
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Authentication failed';
        setError(errorMessage);
        setIsProcessing(false);
      }
    };

    processCallback();
  }, [router, searchParams]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        {isProcessing ? (
          <>
            <div className="flex justify-center mb-4">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
            </div>
            <h1 className="text-2xl font-bold text-center mb-2">
              Processing Authentication...
            </h1>
            <p className="text-gray-600 text-center">
              Please wait while we complete your sign in
            </p>
          </>
        ) : (
          <>
            <div className="text-center mb-4">
              <span className="text-5xl">❌</span>
            </div>
            <h1 className="text-2xl font-bold text-center mb-2 text-red-600">
              Authentication Failed
            </h1>
            <p className="text-gray-600 text-center mb-6">{error}</p>
            <div className="flex gap-4">
              <button
                onClick={() => router.push('/signin')}
                className="flex-1 py-2 px-4 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
              >
                Back to Sign In
              </button>
              <button
                onClick={() => router.push('/')}
                className="flex-1 py-2 px-4 bg-gray-200 text-gray-700 font-medium rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
              >
                Go Home
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default function OAuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <OAuthCallbackContent />
    </Suspense>
  );
}
