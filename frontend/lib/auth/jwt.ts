const JWT_TOKEN_KEY = 'orgmind_jwt_token';
const JWT_COOKIE_NAME = 'orgmind_token';

/**
 * Store JWT token in localStorage and cookie
 */
export function setJWTToken(token: string): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(JWT_TOKEN_KEY, token);
    
    // Also store in cookie for server-side middleware access
    // Set cookie with secure flags
    const maxAge = 24 * 60 * 60; // 24 hours in seconds
    document.cookie = `${JWT_COOKIE_NAME}=${token}; path=/; max-age=${maxAge}; SameSite=Lax`;
  }
}

/**
 * Retrieve JWT token from localStorage
 */
export function getJWTToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(JWT_TOKEN_KEY);
  }
  return null;
}

/**
 * Clear JWT token from localStorage and cookie
 */
export function clearJWTToken(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(JWT_TOKEN_KEY);
    
    // Clear cookie by setting it to expire immediately
    document.cookie = `${JWT_COOKIE_NAME}=; path=/; max-age=0; SameSite=Lax`;
  }
}

/**
 * Decode JWT token payload without verification
 * Returns null if token is invalid
 */
function decodeJWTPayload(token: string): any | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }
    
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded);
  } catch (error) {
    return null;
  }
}

/**
 * Check if JWT token is expired
 * Returns true if token is expired or invalid
 */
export function isTokenExpired(token: string): boolean {
  const payload = decodeJWTPayload(token);
  
  if (!payload || !payload.exp) {
    return true;
  }
  
  // exp is in seconds, Date.now() is in milliseconds
  const expirationTime = payload.exp * 1000;
  const currentTime = Date.now();
  
  return currentTime >= expirationTime;
}

/**
 * Check if current stored token is valid and not expired
 */
export function hasValidToken(): boolean {
  const token = getJWTToken();
  
  if (!token) {
    return false;
  }
  
  return !isTokenExpired(token);
}

/**
 * Get user ID from JWT token
 * Returns null if token is invalid or doesn't contain user ID
 */
export function getUserIdFromToken(): string | null {
  const token = getJWTToken();
  
  if (!token) {
    return null;
  }
  
  const payload = decodeJWTPayload(token);
  
  if (!payload || !payload.userId) {
    return null;
  }
  
  return payload.userId;
}
