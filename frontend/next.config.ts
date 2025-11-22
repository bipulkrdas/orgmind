import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  /* config options here */
  
  // Enable standalone output for Docker deployment
  output: 'standalone',
  async rewrites() {
    // This proxy is used ONLY in local development.
    // In production, the frontend will call the backend API directly using the NEXT_PUBLIC_API_URL.
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*', // Proxy to the local Go backend
      },
    ];
  },
  
  // Optimize images for production
  images: {
    unoptimized: false,
    remotePatterns: [],
  },
  
  // Disable telemetry in production
  ...(process.env.NODE_ENV === 'production' && {
    productionBrowserSourceMaps: false,
  }),
};

export default nextConfig;
