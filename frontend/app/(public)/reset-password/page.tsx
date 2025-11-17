'use client';

import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import ResetPasswordForm from '@/components/auth/ResetPasswordForm';
import UpdatePasswordForm from '@/components/auth/UpdatePasswordForm';

function ResetPasswordContent() {
  const searchParams = useSearchParams();
  const token = searchParams?.get('token');

  const handleSuccess = () => {
    console.log('Password reset successful');
  };

  const handleError = (error: Error) => {
    console.error('Password reset error:', error);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        {token ? (
          <>
            <h1 className="text-2xl font-bold text-center mb-2">Set New Password</h1>
            <p className="text-sm text-gray-600 text-center mb-6">
              Enter your new password below
            </p>
            <UpdatePasswordForm
              token={token}
              onSuccess={handleSuccess}
              onError={handleError}
            />
          </>
        ) : (
          <>
            <h1 className="text-2xl font-bold text-center mb-2">Reset Password</h1>
            <p className="text-sm text-gray-600 text-center mb-6">
              Enter your email address and we&apos;ll send you a link to reset your password
            </p>
            <ResetPasswordForm onSuccess={handleSuccess} onError={handleError} />
          </>
        )}
        
        <div className="mt-6 text-center">
          <Link
            href="/signin"
            className="text-sm text-indigo-600 hover:text-indigo-700"
          >
            Back to sign in
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <ResetPasswordContent />
    </Suspense>
  );
}
