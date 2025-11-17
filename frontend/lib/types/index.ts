// User types
export interface User {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  oauthProvider?: string;
  oauthId?: string;
  createdAt: string;
  updatedAt: string;
}

// Document types
export interface Document {
  id: string;
  userId: string;
  graphId?: string;
  filename?: string;
  contentType?: string;
  storageKey: string;
  sizeBytes: number;
  source: 'editor' | 'upload';
  status: 'processing' | 'completed' | 'failed';
  createdAt: string;
  updatedAt: string;
}

// Graph management types
export interface Graph {
  id: string;
  creatorId: string;
  zepGraphId: string;
  name: string;
  description?: string;
  documentCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface GraphMembership {
  id: string;
  graphId: string;
  userId: string;
  role: 'owner' | 'editor' | 'viewer' | 'member';
  createdAt: string;
}

export interface CreateGraphRequest {
  name: string;
  description?: string;
}

export interface UpdateGraphRequest {
  name?: string;
  description?: string;
}

export interface AddMemberRequest {
  userId: string;
  role?: 'member' | 'editor' | 'viewer';
}

// Graph visualization types with full Zep metadata
export interface GraphNode {
  id: string;
  name: string;
  summary: string;
  labels?: string[];
  score?: number;
  relevance?: number;
  attributes?: Record<string, any>;
  createdAt?: string;
  // Visualization properties
  size: number;
  color: string;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  name: string;
  fact: string;
  validAt?: string;
  invalidAt?: string;
  episodes?: string[];
  sourceNodeName?: string;
  targetNodeName?: string;
  sourceNodeSummary?: string;
  targetNodeSummary?: string;
  attributes?: Record<string, any>;
  createdAt?: string;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

// Authentication types
export interface Credentials {
  email: string;
  password: string;
}

export interface SignUpCredentials extends Credentials {
  firstName?: string;
  lastName?: string;
}

export type OAuthProvider = 'google' | 'okta' | 'office365';

// Password reset types
export interface ResetPasswordRequest {
  email: string;
}

export interface UpdatePasswordRequest {
  token: string;
  newPassword: string;
}

// API response types
export interface AuthResponse {
  token: string;
  user: User;
}

export interface ErrorResponse {
  code: string;
  message: string;
  details?: any;
}
