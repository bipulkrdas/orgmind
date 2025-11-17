'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { getDocument, getDocumentContent, updateDocument, deleteDocument } from '@/lib/api/documents';
import { LexicalEditor } from '@/components/editor';
import type { Document } from '@/lib/types';

export default function DocumentViewPage() {
  const params = useParams();
  const router = useRouter();
  const graphId = params?.graphId as string;
  const documentId = params?.documentId as string;

  const [document, setDocument] = useState<Document | null>(null);
  const [content, setContent] = useState<string>('');
  const [lexicalState, setLexicalState] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    if (!documentId) return;

    const fetchDocument = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch document metadata
        const doc = await getDocument(documentId);
        setDocument(doc);

        // For editor documents, fetch content
        if (doc.source === 'editor') {
          const docContent = await getDocumentContent(documentId);
          setContent(docContent.plainText || '');
          setLexicalState(docContent.lexicalState);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load document');
      } finally {
        setLoading(false);
      }
    };

    fetchDocument();
  }, [documentId]);

  const handleSave = async (plainText: string, newLexicalState: string) => {
    if (!documentId) return;

    try {
      const updatedDoc = await updateDocument(documentId, plainText, newLexicalState);
      setDocument(updatedDoc);
      setContent(plainText);
      setLexicalState(newLexicalState);
      setIsEditing(false);
    } catch (err) {
      throw new Error(err instanceof Error ? err.message : 'Failed to update document');
    }
  };

  const handleDelete = async () => {
    if (!documentId || !graphId) return;

    const confirmed = window.confirm('Are you sure you want to delete this document? This action cannot be undone.');
    if (!confirmed) return;

    try {
      setIsDeleting(true);
      await deleteDocument(documentId);
      router.push(`/graphs/${graphId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete document');
      setIsDeleting(false);
    }
  };

  const handleDownload = () => {
    if (!document || !content) return;

    const blob = new Blob([content], { type: document.contentType || 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = window.document.createElement('a');
    a.href = url;
    a.download = document.filename || 'document.txt';
    window.document.body.appendChild(a);
    a.click();
    window.document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600">Loading document...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen">
        <div className="text-red-600 mb-4">{error}</div>
        <button
          onClick={() => router.push(`/graphs/${graphId}`)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Back to Graph
        </button>
      </div>
    );
  }

  if (!document) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600">Document not found</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-6xl">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <button
            onClick={() => router.push(`/graphs/${graphId}`)}
            className="text-blue-600 hover:text-blue-700 flex items-center"
          >
            <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Graph
          </button>
          <div className="flex gap-2">
            {document.source === 'editor' && !isEditing && (
              <button
                onClick={() => setIsEditing(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Edit
              </button>
            )}
            {document.source === 'editor' && isEditing && (
              <button
                onClick={() => setIsEditing(false)}
                className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700"
              >
                Cancel
              </button>
            )}
            <button
              onClick={handleDelete}
              disabled={isDeleting}
              className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:bg-gray-400"
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </button>
          </div>
        </div>

        <h1 className="text-3xl font-bold text-gray-900">
          {document.filename || 'Untitled Document'}
        </h1>
        <div className="flex items-center gap-4 mt-2 text-sm text-gray-600">
          <span className="capitalize">{document.source} document</span>
          <span>•</span>
          <span>{new Date(document.createdAt).toLocaleDateString()}</span>
          <span>•</span>
          <span>{(document.sizeBytes / 1024).toFixed(2)} KB</span>
          {document.status && (
            <>
              <span>•</span>
              <span className={`capitalize ${
                document.status === 'completed' ? 'text-green-600' :
                document.status === 'processing' ? 'text-yellow-600' :
                'text-red-600'
              }`}>
                {document.status}
              </span>
            </>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        {document.source === 'editor' ? (
          isEditing ? (
            <LexicalEditor
              onSubmit={handleSave}
              initialContent={content}
              initialLexicalState={lexicalState}
            />
          ) : (
            <div className="prose max-w-none">
              <div className="whitespace-pre-wrap">{content}</div>
            </div>
          )
        ) : (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-md">
              <h2 className="text-lg font-semibold mb-3">File Information</h2>
              <dl className="space-y-2">
                <div className="flex">
                  <dt className="font-medium text-gray-700 w-32">Filename:</dt>
                  <dd className="text-gray-900">{document.filename}</dd>
                </div>
                <div className="flex">
                  <dt className="font-medium text-gray-700 w-32">Type:</dt>
                  <dd className="text-gray-900">{document.contentType}</dd>
                </div>
                <div className="flex">
                  <dt className="font-medium text-gray-700 w-32">Size:</dt>
                  <dd className="text-gray-900">{(document.sizeBytes / 1024).toFixed(2)} KB</dd>
                </div>
                <div className="flex">
                  <dt className="font-medium text-gray-700 w-32">Uploaded:</dt>
                  <dd className="text-gray-900">{new Date(document.createdAt).toLocaleString()}</dd>
                </div>
              </dl>
            </div>
            <button
              onClick={handleDownload}
              className="w-full px-4 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 flex items-center justify-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download File
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
