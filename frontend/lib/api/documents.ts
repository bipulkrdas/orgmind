import { apiCall } from './client';
import type { Document } from '../types';

/**
 * Submit content from the editor
 */
export async function submitEditorContent(content: string, lexicalState: string, graphId: string): Promise<Document> {
  return apiCall<Document>('/api/documents/editor', {
    method: 'POST',
    body: JSON.stringify({ content, lexicalState, graphId }),
  });
}

/**
 * Upload a file
 */
export async function uploadFile(file: File, graphId: string): Promise<Document> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('graphId', graphId);
  
  return apiCall<Document>('/api/documents/upload', {
    method: 'POST',
    body: formData,
  });
}

/**
 * List all documents for the authenticated user
 */
export async function listDocuments(): Promise<Document[]> {
  const response = await apiCall<{ documents: Document[] } | Document[]>('/api/documents', {
    method: 'GET',
  });
  
  // Handle both response formats defensively
  if (Array.isArray(response)) {
    return response;
  }
  
  // Handle wrapped response format
  if (response && typeof response === 'object' && 'documents' in response) {
    return Array.isArray(response.documents) ? response.documents : [];
  }
  
  // Fallback to empty array for unexpected formats
  return [];
}

/**
 * List documents for a specific graph
 */
export async function listGraphDocuments(graphId: string): Promise<Document[]> {
  const response = await apiCall<{ documents: Document[] } | Document[]>(`/api/graphs/${graphId}/documents`, {
    method: 'GET',
  });
  
  // Handle both response formats defensively
  if (Array.isArray(response)) {
    return response;
  }
  
  // Handle wrapped response format
  if (response && typeof response === 'object' && 'documents' in response) {
    return Array.isArray(response.documents) ? response.documents : [];
  }
  
  // Fallback to empty array for unexpected formats
  return [];
}

/**
 * Get a specific document by ID
 */
export async function getDocument(documentId: string): Promise<Document> {
  return apiCall<Document>(`/api/documents/${documentId}`, {
    method: 'GET',
  });
}

/**
 * Update document content
 */
export async function updateDocument(documentId: string, content: string, lexicalState: string): Promise<Document> {
  return apiCall<Document>(`/api/documents/${documentId}`, {
    method: 'PUT',
    body: JSON.stringify({ content, lexicalState }),
  });
}

/**
 * Get document content (returns both plain text and Lexical state)
 */
export async function getDocumentContent(documentId: string): Promise<{
  plainText?: string;
  lexicalState?: string;
  metadata?: Record<string, any>;
}> {
  const response = await apiCall<{
    plainText?: string;
    lexicalState?: string;
    metadata?: Record<string, any>;
    content?: string; // Legacy format support
  }>(`/api/documents/${documentId}/content`, {
    method: 'GET',
  });
  
  // Handle new format (with plainText and lexicalState)
  if (response.plainText !== undefined || response.lexicalState !== undefined) {
    return {
      plainText: response.plainText,
      lexicalState: response.lexicalState,
      metadata: response.metadata,
    };
  }
  
  // Handle legacy format (just content string)
  if (response.content !== undefined) {
    return {
      plainText: response.content,
      metadata: { version: 'legacy', type: 'plain-text' },
    };
  }
  
  // Fallback to empty
  return {
    plainText: '',
    metadata: { version: 'unknown' },
  };
}

/**
 * Delete a document
 */
export async function deleteDocument(documentId: string): Promise<void> {
  return apiCall<void>(`/api/documents/${documentId}`, {
    method: 'DELETE',
  });
}
