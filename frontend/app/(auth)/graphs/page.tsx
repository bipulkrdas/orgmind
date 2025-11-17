'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { listGraphs } from '@/lib/api/graphs';
import { getUserIdFromToken } from '@/lib/auth/jwt';
import type { Graph } from '@/lib/types';
import {
  GraphsList,
  CreateGraphModal,
  EditGraphModal,
  DeleteGraphDialog,
} from '@/components/graphs';

export default function GraphsPage() {
  const router = useRouter();
  const [graphs, setGraphs] = useState<Graph[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  // Modal states
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [selectedGraph, setSelectedGraph] = useState<Graph | null>(null);

  // Fetch current user ID
  useEffect(() => {
    const userId = getUserIdFromToken();
    setCurrentUserId(userId);
  }, []);

  // Fetch graphs on mount
  useEffect(() => {
    fetchGraphs();
  }, []);

  const fetchGraphs = async () => {
    setIsLoading(true);
    setError('');

    try {
      const data = await listGraphs();
      setGraphs(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load graphs';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateGraph = () => {
    setIsCreateModalOpen(true);
  };

  const handleEditGraph = (graph: Graph) => {
    setSelectedGraph(graph);
    setIsEditModalOpen(true);
  };

  const handleDeleteGraph = (graphId: string) => {
    const graph = graphs.find((g) => g.id === graphId);
    if (graph) {
      setSelectedGraph(graph);
      setIsDeleteDialogOpen(true);
    }
  };

  const handleGraphClick = (graphId: string) => {
    router.push(`/graphs/${graphId}`);
  };

  const handleModalSuccess = () => {
    // Refresh graphs list after create/edit/delete
    fetchGraphs();
  };

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-red-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading graphs</h3>
              <p className="mt-1 text-sm text-red-700">{error}</p>
              <button
                onClick={fetchGraphs}
                className="mt-3 text-sm font-medium text-red-800 hover:text-red-900 underline"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <GraphsList
        graphs={graphs}
        isLoading={isLoading}
        onCreateGraph={handleCreateGraph}
        onEditGraph={handleEditGraph}
        onDeleteGraph={handleDeleteGraph}
        onGraphClick={handleGraphClick}
        currentUserId={currentUserId || undefined}
      />

      <CreateGraphModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSuccess={handleModalSuccess}
      />

      <EditGraphModal
        isOpen={isEditModalOpen}
        graph={selectedGraph}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedGraph(null);
        }}
        onSuccess={handleModalSuccess}
      />

      <DeleteGraphDialog
        isOpen={isDeleteDialogOpen}
        graph={selectedGraph}
        onClose={() => {
          setIsDeleteDialogOpen(false);
          setSelectedGraph(null);
        }}
        onSuccess={handleModalSuccess}
      />
    </>
  );
}
