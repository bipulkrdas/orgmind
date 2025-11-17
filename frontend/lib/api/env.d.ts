// Type declaration for Node.js process.env in Next.js
declare namespace NodeJS {
  interface ProcessEnv {
    NEXT_PUBLIC_API_URL?: string;
    NEXT_PUBLIC_OAUTH_REDIRECT_URL?: string;
  }
}
