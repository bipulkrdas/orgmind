'use client';

import { useEffect, useState } from 'react';
import { GraphVisualizer } from '@/components/graph';

interface GraphVisualizerModalProps {
  isOpen: boolean;
  onClose: () => void;
  graphId: string;
}

export default function GraphVisualizerModal({
  isOpen,
  onClose,
  graphId,
}: GraphVisualizerModalProps) {
  const [searchQuery, setSearchQuery] = useState('all');
  const [currentQuery, setCurrentQuery] = useState('all');

  // Handle ESC key to close modal
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentQuery(searchQuery.trim() || 'all');
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={onClose}
      />

      {/* Modal Container */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div
          className="relative w-full max-w-7xl transform overflow-hidden rounded-2xl bg-white shadow-2xl transition-all"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
            <div>
              <h3 className="text-xl font-semibold text-gray-900">
                Knowledge Graph Visualization
              </h3>
              <p className="mt-1 text-sm text-gray-600">
                Interactive visualization of connections between documents and concepts
              </p>
            </div>
            <button
              type="button"
              className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
              onClick={onClose}
            >
              <span className="sr-only">Close</span>
              <svg
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth="1.5"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Search Bar */}
          <div className="border-b border-gray-200 bg-white px-6 py-4">
            <form onSubmit={handleSearch} className="flex gap-3 items-center">
              <div className="flex-1">
                <label htmlFor="graph-search" className="sr-only">
                  Search graph
                </label>
                <input
                  type="text"
                  id="graph-search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search for entities, concepts, or relationships... (e.g., 'person', 'organization', 'all')"
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 px-4 py-2.5 text-sm"
                />
              </div>
              <button
                type="submit"
                className="inline-flex items-center rounded-md bg-indigo-600 px-6 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 transition-colors whitespace-nowrap"
              >
                <svg
                  className="h-4 w-4 mr-2"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth="2"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
                  />
                </svg>
                Search
              </button>
            </form>
            <p className="mt-2 text-xs text-gray-500">
              Enter a search term to filter the graph, or use &quot;all&quot; to see everything
            </p>
          </div>

          {/* Graph Visualizer Content */}
          <div className="p-6">
            <div className="rounded-lg border border-gray-200 bg-gray-50">
              <GraphVisualizer
                graphId={graphId}
                query={currentQuery}
                width={1200}
                height={700}
              />
            </div>
          </div>

          {/* Footer with instructions */}
          <div className="border-t border-gray-200 bg-gray-50 px-6 py-4">
            <div className="flex items-start space-x-2 text-sm text-gray-600">
              <svg
                className="h-5 w-5 text-gray-400 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <div>
                <p className="font-medium text-gray-700">
                  Interaction Tips:
                </p>
                <ul className="mt-1 space-y-1 text-gray-600">
                  <li>• Drag nodes to rearrange the graph</li>
                  <li>• Scroll to zoom in/out</li>
                  <li>• Click nodes to see details</li>
                  <li>• Hover over connections to see relationships</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
