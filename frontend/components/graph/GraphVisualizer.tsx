'use client';

import { useEffect, useRef, useState } from 'react';
import Graph from 'graphology';
import Sigma from 'sigma';
import { circular } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import { getGraphVisualization } from '@/lib/api/graphs';
import type { GraphData, GraphNode as GraphNodeType, GraphEdge as GraphEdgeType } from '@/lib/types';

interface GraphVisualizerProps {
  graphId: string;
  query?: string;
  width?: number;
  height?: number;
}

interface TooltipState {
  visible: boolean;
  x: number;
  y: number;
  type: 'node' | 'edge';
  data: GraphNodeType | GraphEdgeType | null;
}

export default function GraphVisualizer({ 
  graphId,
  query = 'all',
  width = 800, 
  height = 600 
}: GraphVisualizerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const sigmaRef = useRef<Sigma | null>(null);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  const [graphData, setGraphData] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tooltip, setTooltip] = useState<TooltipState>({
    visible: false,
    x: 0,
    y: 0,
    type: 'node',
    data: null,
  });

  // Fetch graph visualization data
  useEffect(() => {
    const fetchGraphData = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await getGraphVisualization(graphId, query);
        setGraphData(data);
      } catch (err: any) {
        console.error('Failed to fetch graph visualization:', err);
        setError(err.message || 'Failed to load graph visualization');
      } finally {
        setLoading(false);
      }
    };

    fetchGraphData();
  }, [graphId, query]);

  // Render graph with sigma.js
  useEffect(() => {
    if (!containerRef.current || !graphData || !graphData.nodes?.length) return;

    // Create a new graph
    const graph = new Graph();

    // Add nodes with full metadata
    graphData.nodes.forEach(node => {
      graph.addNode(node.id, {
        label: node.name,
        size: node.size,
        color: node.color,
        x: Math.random(),
        y: Math.random(),
        // Store full node data for tooltips/details
        summary: node.summary,
        labels: node.labels,
        score: node.score,
        relevance: node.relevance,
        attributes: node.attributes,
      });
    });

    // Add edges with full metadata
    graphData.edges.forEach(edge => {
      try {
        graph.addEdge(edge.source, edge.target, {
          label: edge.name,
          size: 4, // Increased from 2 to make edges easier to hover
          color: '#3b82f6', // Blue color for better visibility
          // Store full edge data for tooltips/details
          fact: edge.fact,
          validAt: edge.validAt,
          invalidAt: edge.invalidAt,
          episodes: edge.episodes,
          sourceNodeName: edge.sourceNodeName,
          targetNodeName: edge.targetNodeName,
        });
      } catch (error) {
        console.warn(`Failed to add edge from ${edge.source} to ${edge.target}:`, error);
      }
    });

    // Apply circular layout first
    circular.assign(graph);

    // Apply force-directed layout
    const settings = forceAtlas2.inferSettings(graph);
    forceAtlas2.assign(graph, {
      iterations: 50,
      settings: {
        ...settings,
        gravity: 1,
        scalingRatio: 10,
      },
    });

    // Initialize Sigma
    const sigma = new Sigma(graph, containerRef.current, {
      renderEdgeLabels: true,
      defaultNodeColor: '#999',
      defaultEdgeColor: '#3b82f6', // Blue color for edges
      labelSize: 14,
      labelWeight: 'bold',
      enableEdgeEvents: true,
      // Make edges easier to hover with better styling
      edgeLabelSize: 12,
      edgeLabelWeight: 'normal',
      edgeLabelColor: { color: '#1e40af' }, // Darker blue for edge labels
    });

    sigmaRef.current = sigma;

    // Node click handler
    sigma.on('clickNode', ({ node }) => {
      setSelectedNode(node);
      
      // Highlight selected node and its neighbors
      const neighbors = new Set(graph.neighbors(node));
      
      graph.forEachNode((n, attributes) => {
        if (n === node) {
          graph.setNodeAttribute(n, 'highlighted', true);
          graph.setNodeAttribute(n, 'size', attributes.size * 1.5);
        } else if (neighbors.has(n)) {
          graph.setNodeAttribute(n, 'highlighted', true);
        } else {
          graph.setNodeAttribute(n, 'highlighted', false);
          graph.setNodeAttribute(n, 'color', '#ddd');
        }
      });

      graph.forEachEdge((edge, _attributes, source, target) => {
        if (source === node || target === node) {
          graph.setEdgeAttribute(edge, 'color', '#1e40af'); // Darker blue for highlighted edges
          graph.setEdgeAttribute(edge, 'size', 6);
        } else {
          graph.setEdgeAttribute(edge, 'color', '#93c5fd'); // Light blue for non-highlighted edges
          graph.setEdgeAttribute(edge, 'size', 2);
        }
      });

      sigma.refresh();
    });

    // Node hover handlers with tooltip
    sigma.on('enterNode', ({ node, event }) => {
      setHoveredNode(node);
      containerRef.current!.style.cursor = 'pointer';
      
      // Show tooltip at node position
      const nodeData = graphData.nodes.find(n => n.id === node);
      if (nodeData && containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        setTooltip({
          visible: true,
          x: event.x - rect.left,
          y: event.y - rect.top,
          type: 'node',
          data: nodeData,
        });
      }
    });

    sigma.on('leaveNode', () => {
      setHoveredNode(null);
      containerRef.current!.style.cursor = 'default';
      setTooltip(prev => ({ ...prev, visible: false }));
    });

    // Edge hover handlers with tooltip
    sigma.on('enterEdge', ({ edge, event }) => {
      containerRef.current!.style.cursor = 'pointer';
      
      // Find edge data
      const edgeData = graphData.edges.find(e => {
        const edgeKey = graph.edge(e.source, e.target);
        return edgeKey === edge;
      });
      
      if (edgeData && containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        setTooltip({
          visible: true,
          x: event.x - rect.left,
          y: event.y - rect.top,
          type: 'edge',
          data: edgeData,
        });
      }
    });

    sigma.on('leaveEdge', () => {
      containerRef.current!.style.cursor = 'default';
      setTooltip(prev => ({ ...prev, visible: false }));
    });

    // Click on stage to deselect
    sigma.on('clickStage', () => {
      setSelectedNode(null);
      
      // Reset all nodes and edges to original state
      graphData.nodes.forEach(node => {
        graph.setNodeAttribute(node.id, 'highlighted', false);
        graph.setNodeAttribute(node.id, 'size', node.size);
        graph.setNodeAttribute(node.id, 'color', node.color);
      });

      graphData.edges.forEach(edge => {
        try {
          graph.setEdgeAttribute(edge.source + '-' + edge.target, 'color', '#3b82f6');
          graph.setEdgeAttribute(edge.source + '-' + edge.target, 'size', 4);
        } catch (error) {
          // Edge might not exist
        }
      });

      sigma.refresh();
    });

    // Cleanup
    return () => {
      sigma.kill();
      sigmaRef.current = null;
    };
  }, [graphData]);

  // Loading state
  if (loading) {
    return (
      <div 
        className="flex flex-col items-center justify-center bg-gray-50 rounded-lg border border-gray-300"
        style={{ width, height }}
      >
        <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
        <p className="text-gray-600">Loading graph visualization...</p>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div 
        className="flex flex-col items-center justify-center bg-red-50 rounded-lg border border-red-300"
        style={{ width, height }}
      >
        <svg
          className="h-12 w-12 text-red-400 mb-4"
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
        <p className="text-red-600 font-medium">Failed to load graph</p>
        <p className="text-red-500 text-sm mt-1">{error}</p>
      </div>
    );
  }

  // Empty state
  if (!graphData || !graphData.nodes?.length) {
    return (
      <div 
        className="flex flex-col items-center justify-center bg-gray-50 rounded-lg border-2 border-dashed border-gray-300"
        style={{ width, height }}
      >
        <svg
          className="h-12 w-12 text-gray-400 mb-4"
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
        <p className="text-gray-600 font-medium">No graph data available</p>
        <p className="text-gray-500 text-sm mt-1">Add documents to this graph to see connections</p>
      </div>
    );
  }

  // Zoom controls
  const handleZoomIn = () => {
    if (sigmaRef.current) {
      const camera = sigmaRef.current.getCamera();
      camera.animatedZoom({ duration: 200 });
    }
  };

  const handleZoomOut = () => {
    if (sigmaRef.current) {
      const camera = sigmaRef.current.getCamera();
      camera.animatedUnzoom({ duration: 200 });
    }
  };

  const handleResetView = () => {
    if (sigmaRef.current) {
      const camera = sigmaRef.current.getCamera();
      camera.animatedReset({ duration: 200 });
    }
  };

  return (
    <div className="relative">
      <div
        ref={containerRef}
        style={{ width, height }}
        className="border border-gray-300 rounded-lg bg-white relative"
      >
        {/* Tooltip */}
        {tooltip.visible && tooltip.data && (
          <div
            className="absolute z-50 pointer-events-none"
            style={{
              left: tooltip.x + 10,
              top: tooltip.y + 10,
              maxWidth: '300px',
            }}
          >
            <div className="bg-white border-2 border-gray-300 rounded-lg shadow-xl p-3">
              {tooltip.type === 'node' && 'summary' in tooltip.data && (
                <div className="space-y-2">
                  <div>
                    <p className="font-bold text-gray-900 text-sm">{tooltip.data.name}</p>
                    {tooltip.data.summary && (
                      <p className="text-gray-700 text-xs mt-1">{tooltip.data.summary}</p>
                    )}
                  </div>
                  {tooltip.data.labels && tooltip.data.labels.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {tooltip.data.labels.map((label: string, idx: number) => (
                        <span
                          key={idx}
                          className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                        >
                          {label}
                        </span>
                      ))}
                    </div>
                  )}
                  {tooltip.data.score !== undefined && (
                    <p className="text-xs text-gray-500">Score: {tooltip.data.score.toFixed(2)}</p>
                  )}
                </div>
              )}
              {tooltip.type === 'edge' && 'fact' in tooltip.data && (
                <div className="space-y-1">
                  <p className="font-bold text-gray-900 text-sm">{tooltip.data.name}</p>
                  <p className="text-gray-700 text-xs">{tooltip.data.fact}</p>
                  {tooltip.data.validAt && (
                    <p className="text-xs text-gray-500">
                      Valid: {tooltip.data.validAt}
                      {tooltip.data.invalidAt && ` - ${tooltip.data.invalidAt}`}
                    </p>
                  )}
                  {tooltip.data.sourceNodeName && tooltip.data.targetNodeName && (
                    <p className="text-xs text-gray-600 mt-1">
                      {tooltip.data.sourceNodeName} → {tooltip.data.targetNodeName}
                    </p>
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Zoom Controls */}
      <div className="absolute top-4 right-4 flex flex-col gap-2 bg-white rounded-lg shadow-lg border border-gray-300 p-1">
        <button
          onClick={handleZoomIn}
          className="p-2 hover:bg-gray-100 rounded transition-colors"
          title="Zoom In"
        >
          <svg className="w-5 h-5 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
        <button
          onClick={handleZoomOut}
          className="p-2 hover:bg-gray-100 rounded transition-colors"
          title="Zoom Out"
        >
          <svg className="w-5 h-5 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
          </svg>
        </button>
        <button
          onClick={handleResetView}
          className="p-2 hover:bg-gray-100 rounded transition-colors border-t border-gray-200"
          title="Reset View"
        >
          <svg className="w-5 h-5 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
      
      {/* Controls info */}
      <div className="mt-4 p-4 bg-gray-50 rounded-lg">
        <h3 className="font-semibold mb-2">Controls:</h3>
        <ul className="text-sm text-gray-600 space-y-1">
          <li>• Click and drag to pan</li>
          <li>• Scroll to zoom in/out (or use +/- buttons)</li>
          <li>• Hover over nodes/edges to see details</li>
          <li>• Click a node to highlight it and its connections</li>
          <li>• Click the background to deselect</li>
        </ul>
        
        {selectedNode && (
          <div className="mt-3 pt-3 border-t border-gray-200">
            <div className="text-sm space-y-2">
              <p className="font-semibold text-gray-900">Selected Node:</p>
              {(() => {
                const node = graphData?.nodes.find(n => n.id === selectedNode);
                if (!node) return <p className="text-gray-600">{selectedNode}</p>;
                return (
                  <div className="space-y-1">
                    <p className="font-medium text-gray-900">{node.name}</p>
                    {node.summary && (
                      <p className="text-gray-600 text-xs">{node.summary}</p>
                    )}
                    {node.labels && node.labels.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {node.labels.map((label, idx) => (
                          <span
                            key={idx}
                            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                          >
                            {label}
                          </span>
                        ))}
                      </div>
                    )}
                    {node.score !== undefined && (
                      <p className="text-xs text-gray-500">Score: {node.score.toFixed(2)}</p>
                    )}
                  </div>
                );
              })()}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
