import { getJWTToken, clearJWTToken } from '../auth/jwt';
import type { ErrorResponse } from '../types';

// Get API base URL
// In production, use the backend service URL from environment variable
// In development, use localhost
/*
function getApiBaseUrl(): string {
  // If explicitly set via environment variable (build time), use it
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  
  // In browser, check if we're in production
  if (typeof window !== 'undefined') {
    const hostname = window.location.hostname;
    
    // If running on Cloud Run or production domain
    if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
      // Replace 'orgmind-frontend' with 'orgmind-backend' in the hostname
      // This works for Cloud Run URLs like: orgmind-frontend-xxx.run.app -> orgmind-backend-xxx.run.app
      const backendHostname = hostname.replace('orgmind-frontend', 'orgmind-backend');
      return `${window.location.protocol}//${backendHostname}`;
    }
  }
  
  // Default to localhost for development
  return 'http://localhost:8080';
}
  */

export const getApiBaseUrl = (): string => {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  // Production: Use the full URL from environment variables.
  // The NEXT_PUBLIC_API_URL should be the base URL of the backend (e.g., https://api.example.com)
  if (process.env.NODE_ENV === 'production' && apiUrl) {
    return `${apiUrl}`;
  }

  // Development: Use a relative path for the Next.js proxy.
  return ``;
};

export const API_BASE_URL = getApiBaseUrl();

/**
 * Custom API Error class for structured error handling
 */
export class APIError extends Error {
  constructor(
    public statusCode: number,
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'APIError';
  }
}

/**
 * Generic API call function with automatic JWT injection, timeout, and error handling
 */
export async function apiCall<T>(
  endpoint: string,
  options?: RequestInit & { timeout?: number }
): Promise<T> {
  const token = getJWTToken();
  const timeout = options?.timeout || 30000; // Default 30 second timeout
  
  const headers: Record<string, string> = {
    ...(options?.headers as Record<string, string>),
  };
  
  // Add JWT token if available
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  // Add Content-Type for JSON requests (unless it's FormData)
  if (options?.body && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }
  
  // Create abort controller for timeout
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
      signal: controller.signal,
    });
    
    clearTimeout(timeoutId);

    // Handle 401 Unauthorized - clear token and redirect to signin
    if (response.status === 401) {
      clearJWTToken();
      
      // Only redirect if we're in the browser
      if (typeof window !== 'undefined') {
        window.location.href = '/signin';
      }
      
      throw new APIError(
        401,
        'UNAUTHORIZED',
        'Authentication required. Please sign in.'
      );
    }

    // Handle other error responses
    if (!response.ok) {
      let errorData: ErrorResponse;
      
      try {
        errorData = await response.json();
      } catch {
        // If response is not JSON, create a generic error
        throw new APIError(
          response.status,
          'UNKNOWN_ERROR',
          `Request failed with status ${response.status}`
        );
      }
      
      throw new APIError(
        response.status,
        errorData.code || 'UNKNOWN_ERROR',
        errorData.message || 'An error occurred',
        errorData.details
      );
    }

    // Handle successful responses
    // Check if response has content
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return response.json();
    }
    
    // Return empty object for responses without content
    return {} as T;
  } catch (error) {
    clearTimeout(timeoutId);
    
    // Re-throw APIError instances
    if (error instanceof APIError) {
      throw error;
    }
    
    // Handle abort/timeout errors
    if (error instanceof Error && error.name === 'AbortError') {
      throw new APIError(
        0,
        'TIMEOUT_ERROR',
        'Request timeout. The server is taking too long to respond.'
      );
    }
    
    // Handle network errors (no internet, DNS failure, etc.)
    if (error instanceof TypeError) {
      // Check if it's likely an offline error
      const isOffline = typeof navigator !== 'undefined' && !navigator.onLine;
      throw new APIError(
        0,
        'NETWORK_ERROR',
        isOffline 
          ? 'You appear to be offline. Please check your internet connection.'
          : 'Unable to connect to the server. Please check your connection.'
      );
    }
    
    // Handle other unexpected errors
    throw new APIError(
      500,
      'UNEXPECTED_ERROR',
      error instanceof Error ? error.message : 'An unexpected error occurred'
    );
  }
}
