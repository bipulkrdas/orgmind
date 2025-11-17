'use client';

import type { Graph } from '@/lib/types';
import GraphCard from './GraphCard';

interface GraphsListProps {
  graphs: Graph[];
  isLoading: boolean;
  onCreateGraph: () => void;
  onEditGraph: (graph: Graph) => void;
  onDeleteGraph: (graphId: string) => void;
  onGraphClick: (graphId: string) => void;
  currentUserId?: string;
}

export default function GraphsList({
  graphs,
  isLoading,
  onCreateGraph,
  onEditGraph,
  onDeleteGraph,
  onGraphClick,
  currentUserId,
}: GraphsListProps) {
  // Loading state
  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Loading graphs...</p>
          </div>
        </div>
      </div>
    );
  }

  // Empty state - defensive check for non-array values
  if (!graphs || !Array.isArray(graphs) || graphs.length === 0) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Knowledge Graphs</h1>
          <button
            onClick={onCreateGraph}
            className="px-4 py-2 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition"
          >
            Create Graph
          </button>
        </div>
        <div className="flex flex-col items-center justify-center min-h-[400px] bg-white rounded-lg border-2 border-dashed border-gray-300">
          <svg
            className="w-16 h-16 text-gray-400 mb-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No graphs yet</h3>
          <p className="text-gray-500 mb-6 text-center max-w-sm">
            Create your first knowledge graph to start organizing and connecting your documents.
          </p>
          <button
            onClick={onCreateGraph}
            className="px-6 py-3 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition"
          >
            Create Your First Graph
          </button>
        </div>
      </div>
    );
  }

  // Graphs grid
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Knowledge Graphs</h1>
          <p className="mt-2 text-gray-600">
            {graphs.length} {graphs.length === 1 ? 'graph' : 'graphs'}
          </p>
        </div>
        <button
          onClick={onCreateGraph}
          className="px-4 py-2 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition"
        >
          Create Graph
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Array.isArray(graphs) && graphs.map((graph) => (
          <GraphCard
            key={graph.id}
            graph={graph}
            isCreator={currentUserId === graph.creatorId}
            onEdit={() => onEditGraph(graph)}
            onDelete={() => onDeleteGraph(graph.id)}
            onClick={() => onGraphClick(graph.id)}
          />
        ))}
      </div>
    </div>
  );
}
