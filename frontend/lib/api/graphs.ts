import { apiCall } from './client';
import type { Graph, CreateGraphRequest, UpdateGraphRequest, GraphData, GraphMembership, AddMemberRequest } from '../types';

/**
 * List all graphs the authenticated user is a member of
 */
export async function listGraphs(): Promise<Graph[]> {
  const response = await apiCall<{ graphs: Graph[] } | Graph[]>('/api/graphs', {
    method: 'GET',
  });
  
  // Handle both response formats defensively
  if (Array.isArray(response)) {
    return response;
  }
  
  // Handle wrapped response format
  if (response && typeof response === 'object' && 'graphs' in response) {
    return Array.isArray(response.graphs) ? response.graphs : [];
  }
  
  // Fallback to empty array for unexpected formats
  return [];
}

/**
 * Create a new graph
 */
export async function createGraph(request: CreateGraphRequest): Promise<Graph> {
  return apiCall<Graph>('/api/graphs', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Get graph details by ID
 */
export async function getGraph(graphId: string): Promise<Graph> {
  return apiCall<Graph>(`/api/graphs/${graphId}`, {
    method: 'GET',
  });
}

/**
 * Update graph metadata
 */
export async function updateGraph(graphId: string, request: UpdateGraphRequest): Promise<Graph> {
  return apiCall<Graph>(`/api/graphs/${graphId}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  });
}

/**
 * Delete a graph
 */
export async function deleteGraph(graphId: string): Promise<void> {
  return apiCall<void>(`/api/graphs/${graphId}`, {
    method: 'DELETE',
  });
}

/**
 * Add a member to a graph
 */
export async function addMember(graphId: string, request: AddMemberRequest): Promise<void> {
  return apiCall<void>(`/api/graphs/${graphId}/members`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Remove a member from a graph
 */
export async function removeMember(graphId: string, userId: string): Promise<void> {
  return apiCall<void>(`/api/graphs/${graphId}/members/${userId}`, {
    method: 'DELETE',
  });
}

/**
 * List all members of a graph
 */
export async function listMembers(graphId: string): Promise<GraphMembership[]> {
  const response = await apiCall<{ members: GraphMembership[] } | GraphMembership[]>(`/api/graphs/${graphId}/members`, {
    method: 'GET',
  });
  
  // Handle both response formats defensively
  if (Array.isArray(response)) {
    return response;
  }
  
  // Handle wrapped response format
  if (response && typeof response === 'object' && 'members' in response) {
    return Array.isArray(response.members) ? response.members : [];
  }
  
  // Fallback to empty array for unexpected formats
  return [];
}

/**
 * Get graph visualization data with optional query filter
 */
export async function getGraphVisualization(graphId: string, query?: string): Promise<GraphData> {
  const queryParam = query ? `?query=${encodeURIComponent(query)}` : '';
  return apiCall<GraphData>(`/api/graphs/${graphId}/visualization${queryParam}`, {
    method: 'GET',
  });
}
