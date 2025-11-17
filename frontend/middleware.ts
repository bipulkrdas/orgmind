import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Minimal Next.js middleware for OrgMind
 * 
 * NOTE: This middleware does NOT handle authentication/authorization.
 * 
 * Authentication is handled by:
 * 1. Client-side: React components check for token and redirect if needed
 * 2. Backend: Go server validates JWT signature on every API request
 * 
 * Why no auth in middleware?
 * - Next.js middleware can't securely verify JWT signatures (would need backend secret)
 * - Client-side checks provide UX (redirect to login)
 * - Backend validation provides security (verify signature, check permissions)
 * - Middleware auth would be redundant and create false sense of security
 * 
 * This middleware is kept minimal and can be extended for:
 * - Request logging
 * - Rate limiting (if needed at edge)
 * - A/B testing
 * - Geolocation-based routing
 */
export function middleware(request: NextRequest) {
  // Pass through all requests
  // Client-side code handles auth redirects
  // Backend validates all API requests
  return NextResponse.next();
}

// Disable middleware by default - enable only if needed for specific features
export const config = {
  matcher: [],  // Empty matcher = middleware disabled
};
