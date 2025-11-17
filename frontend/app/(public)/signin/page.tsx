'use client';

import { Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import SignInForm from '@/components/auth/SignInForm';
import OAuthButtons from '@/components/auth/OAuthButtons';

function SignInContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const handleSuccess = () => {
    // Check if there's a redirect parameter
    const redirect = searchParams.get('redirect');
    
    // Redirect to the original page or home page after successful signin
    if (redirect && redirect.startsWith('/')) {
      router.push(redirect);
    } else {
      router.push('/home');
    }
  };

  const handleError = (error: Error) => {
    console.error('Sign in error:', error);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <h1 className="text-2xl font-bold text-center mb-6">Sign In</h1>
        
        <SignInForm onSuccess={handleSuccess} onError={handleError} />
        
        <div className="mt-4 text-center">
          <Link
            href="/reset-password"
            className="text-sm text-indigo-600 hover:text-indigo-700"
          >
            Forgot your password?
          </Link>
        </div>
        
        <OAuthButtons onError={handleError} />
        
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600">
            Don&apos;t have an account?{' '}
            <Link href="/signup" className="text-indigo-600 hover:text-indigo-700 font-medium">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

export default function SignInPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
      </div>
    }>
      <SignInContent />
    </Suspense>
  );
}
