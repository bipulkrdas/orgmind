'use client';

import { Graph, Document } from '@/lib/types';
import { ReactNode } from 'react';

interface GraphDetailProps {
  graph: Graph;
  documents: Document[];
  currentUserId: string;
  onAddDocument: () => void;
  onDocumentClick: (documentId: string) => void;
  onManageMembers?: () => void;
  onViewGraph?: () => void;
  chatComponent?: ReactNode;
}

export function GraphDetail({ 
  graph, 
  documents, 
  currentUserId, 
  onAddDocument, 
  onDocumentClick, 
  onManageMembers, 
  onViewGraph,
  chatComponent 
}: GraphDetailProps) {
  const isCreator = graph.creatorId === currentUserId;

  return (
    <div className="space-y-3">
      {/* Compact Graph Header */}
      <div className="bg-white rounded-lg shadow px-4 py-3">
        <div className="flex items-center justify-between">
          <div className="flex-1 min-w-0 mr-4">
            <h1 className="text-xl font-bold text-gray-900 truncate">{graph.name}</h1>
            <div className="mt-1 flex items-center space-x-4 text-xs text-gray-500">
              <span>{graph.documentCount} docs</span>
              <span>•</span>
              <span>{new Date(graph.createdAt).toLocaleDateString()}</span>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {onViewGraph && (
              <button
                onClick={onViewGraph}
                className="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-indigo-600 hover:bg-indigo-700"
                title="View Knowledge Graph"
              >
                <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                Graph
              </button>
            )}
            {isCreator && onManageMembers && (
              <button
                onClick={onManageMembers}
                className="inline-flex items-center px-2.5 py-1.5 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50"
                title="Manage Members"
              >
                <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
                Members
              </button>
            )}
            <button
              onClick={onAddDocument}
              className="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-blue-600 hover:bg-blue-700"
            >
              <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Doc
            </button>
          </div>
        </div>
      </div>

      {/* Two-Column Layout: Documents (40%) + Chat (60%) on desktop, stacked on mobile */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
        {/* Documents Column - 40% width on desktop (2/5 columns), full width on mobile */}
        <div className="md:col-span-2 order-1 md:order-1">
          <div className="bg-white rounded-lg shadow h-full flex flex-col">
            <div className="px-4 py-2 border-b border-gray-200 flex-shrink-0">
              <h2 className="text-sm font-semibold text-gray-900">Documents</h2>
            </div>
            <div className="p-3 overflow-y-auto flex-1" style={{ maxHeight: 'calc(100vh - 200px)' }}>
              {documents.length === 0 ? (
                <div className="text-center py-8">
                  <svg
                    className="mx-auto h-10 w-10 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                  <h3 className="mt-3 text-sm font-medium text-gray-900">No documents yet</h3>
                  <p className="mt-1 text-xs text-gray-500">
                    Add your first document
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {documents.map((doc) => (
                    <div
                      key={doc.id}
                      onClick={() => onDocumentClick(doc.id)}
                      className="flex items-center justify-between p-2 border border-gray-200 rounded hover:bg-gray-50 cursor-pointer transition-colors"
                    >
                      <div className="flex items-center space-x-2 min-w-0 flex-1">
                        {doc.source === 'editor' ? (
                          <svg
                            className="h-5 w-5 text-blue-500 flex-shrink-0"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                            />
                          </svg>
                        ) : (
                          <svg
                            className="h-5 w-5 text-green-500 flex-shrink-0"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                            />
                          </svg>
                        )}
                        <div className="min-w-0 flex-1">
                          <p className="text-xs font-medium text-gray-900 truncate">
                            {doc.filename || 'Untitled Document'}
                          </p>
                          <p className="text-xs text-gray-500 truncate">
                            {new Date(doc.createdAt).toLocaleDateString()} • {(doc.sizeBytes / 1024).toFixed(1)} KB
                          </p>
                        </div>
                      </div>
                      <svg
                        className="h-4 w-4 text-gray-400 flex-shrink-0 ml-1"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Chat Column - 60% width on desktop (3/5 columns), full width on mobile */}
        <div className="md:col-span-3 order-2 md:order-2">
          <div className="h-full" style={{ height: 'calc(100vh - 200px)', minHeight: '500px' }}>
            {chatComponent}
          </div>
        </div>
      </div>
    </div>
  );
}
