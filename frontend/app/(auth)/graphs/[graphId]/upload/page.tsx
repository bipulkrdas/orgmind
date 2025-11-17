'use client';

import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import FileUpload from '@/components/upload/FileUpload';
import type { Document } from '@/lib/types';

export default function UploadPage() {
  const params = useParams();
  const router = useRouter();
  const graphId = params?.graphId as string;
  const [error, setError] = useState<string | null>(null);

  const handleUploadSuccess = (document: Document) => {
    // Navigate back to graph detail page on success
    router.push(`/graphs/${graphId}`);
  };

  const handleUploadError = (err: Error) => {
    setError(err.message);
  };

  if (!graphId) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Invalid Graph</h1>
          <p className="text-gray-600">Graph ID is missing from the URL</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-6">
          <button
            onClick={() => router.push(`/graphs/${graphId}`)}
            className="text-blue-600 hover:text-blue-700 font-medium flex items-center gap-2"
          >
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 19l-7-7 7-7"
              />
            </svg>
            Back to Graph
          </button>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-6">Upload Document</h1>
          
          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-800 text-sm">{error}</p>
            </div>
          )}

          <FileUpload
            graphId={graphId}
            onUploadSuccess={handleUploadSuccess}
            onUploadError={handleUploadError}
          />
        </div>
      </div>
    </div>
  );
}
