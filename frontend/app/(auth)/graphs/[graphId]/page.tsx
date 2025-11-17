'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { GraphDetail, AddDocumentModal, MembersModal, GraphVisualizerModal } from '@/components/graphs';
import { getGraph } from '@/lib/api/graphs';
import { listGraphDocuments } from '@/lib/api/documents';
import { getUserIdFromToken } from '@/lib/auth/jwt';
import type { Graph, Document } from '@/lib/types';

export default function GraphDetailPage() {
  const params = useParams();
  const router = useRouter();
  const graphId = params?.graphId as string;

  const [graph, setGraph] = useState<Graph | null>(null);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddDocumentModal, setShowAddDocumentModal] = useState(false);
  const [showMembersModal, setShowMembersModal] = useState(false);
  const [showVisualizerModal, setShowVisualizerModal] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string>('');

  useEffect(() => {
    // Get current user ID from JWT token
    const userId = getUserIdFromToken();
    if (userId) {
      setCurrentUserId(userId);
    }
  }, []);

  useEffect(() => {
    if (!graphId) {
      setError('Graph ID is missing');
      setLoading(false);
      return;
    }

    const fetchGraphDetails = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch graph details and documents
        // GraphVisualizer now handles its own data fetching
        
        // WORKAROUND: Neon DB has issues with concurrent prepared statements
        // causing "bind message supplies X parameters, but prepared statement requires Y" errors
        // Using sequential requests instead of Promise.all until backend connection pooling is fixed
        
        // Original parallel approach (commented out due to Neon DB concurrency issues):
        // const [graphDetails, graphDocs] = await Promise.all([
        //   getGraph(graphId),
        //   listGraphDocuments(graphId),
        // ]);
        
        // Sequential workaround:
        const graphDetails = await getGraph(graphId);
        const graphDocs = await listGraphDocuments(graphId);

        setGraph(graphDetails);
        setDocuments(graphDocs);
      } catch (err: any) {
        console.error('Failed to fetch graph details:', err);
        setError(err.message || 'Failed to load graph details');
      } finally {
        setLoading(false);
      }
    };

    fetchGraphDetails();
  }, [graphId]);

  const handleAddDocument = () => {
    setShowAddDocumentModal(true);
  };

  const handleDocumentClick = (documentId: string) => {
    router.push(`/graphs/${graphId}/documents/${documentId}`);
  };

  const handleManageMembers = () => {
    setShowMembersModal(true);
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <div className="h-8 w-64 bg-gray-200 rounded animate-pulse"></div>
          <button
            onClick={() => router.push('/graphs')}
            className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            <svg className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back to Graphs
          </button>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mb-4"></div>
            <p className="text-gray-600">Loading graph details...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !graph) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Graph Details</h1>
          <button
            onClick={() => router.push('/graphs')}
            className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            <svg className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back to Graphs
          </button>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <div className="text-center py-12">
            <svg
              className="mx-auto h-12 w-12 text-red-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <h3 className="mt-4 text-lg font-medium text-gray-900">Error Loading Graph</h3>
            <p className="mt-2 text-sm text-gray-500">{error || 'Graph not found'}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header with Back Button */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => router.push('/graphs')}
            className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            <svg className="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back
          </button>
        </div>
      </div>

      {/* Graph Detail Component */}
      <GraphDetail
        graph={graph}
        documents={documents}
        currentUserId={currentUserId}
        onAddDocument={handleAddDocument}
        onDocumentClick={handleDocumentClick}
        onManageMembers={handleManageMembers}
        onViewGraph={() => setShowVisualizerModal(true)}
      />

      {/* Add Document Modal */}
      <AddDocumentModal
        isOpen={showAddDocumentModal}
        onClose={() => setShowAddDocumentModal(false)}
        graphId={graphId}
      />

      {/* Members Modal */}
      <MembersModal
        isOpen={showMembersModal}
        onClose={() => setShowMembersModal(false)}
        graphId={graphId}
        isCreator={graph.creatorId === currentUserId}
      />

      {/* Graph Visualizer Modal */}
      <GraphVisualizerModal
        isOpen={showVisualizerModal}
        onClose={() => setShowVisualizerModal(false)}
        graphId={graphId}
      />
    </div>
  );
}
