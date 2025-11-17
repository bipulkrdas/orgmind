import { apiCall } from './client';
import { setJWTToken } from '../auth/jwt';
import type {
  Credentials,
  SignUpCredentials,
  AuthResponse,
  OAuthProvider,
  ResetPasswordRequest,
  UpdatePasswordRequest,
} from '../types';

/**
 * Sign up a new user with email and password
 */
export async function signUp(
  credentials: SignUpCredentials
): Promise<AuthResponse> {
  const response = await apiCall<AuthResponse>('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify(credentials),
  });
  
  // Store JWT token
  setJWTToken(response.token);
  
  return response;
}

/**
 * Sign in an existing user with email and password
 */
export async function signIn(credentials: Credentials): Promise<AuthResponse> {
  const response = await apiCall<AuthResponse>('/api/auth/signin', {
    method: 'POST',
    body: JSON.stringify(credentials),
  });
  
  // Store JWT token
  setJWTToken(response.token);
  
  return response;
}

/**
 * Initiate OAuth authentication flow
 * Returns the OAuth provider's authorization URL
 */
export async function initiateOAuth(
  provider: OAuthProvider
): Promise<{ url: string }> {
  return apiCall<{ url: string }>(`/api/auth/oauth/${provider}`, {
    method: 'GET',
  });
}

/**
 * Handle OAuth callback after user authorizes
 * Exchanges the authorization code for a JWT token
 */
export async function handleOAuthCallback(
  provider: OAuthProvider,
  code: string
): Promise<AuthResponse> {
  const response = await apiCall<AuthResponse>(
    `/api/auth/oauth/${provider}/callback?code=${encodeURIComponent(code)}`,
    {
      method: 'GET',
    }
  );
  
  // Store JWT token
  setJWTToken(response.token);
  
  return response;
}

/**
 * Request a password reset email
 */
export async function resetPassword(
  request: ResetPasswordRequest
): Promise<{ message: string }> {
  return apiCall<{ message: string }>('/api/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Update password using reset token
 */
export async function updatePassword(
  request: UpdatePasswordRequest
): Promise<{ message: string }> {
  return apiCall<{ message: string }>('/api/auth/update-password', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}
