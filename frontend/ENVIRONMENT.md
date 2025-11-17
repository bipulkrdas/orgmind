# Frontend Environment Configuration Guide

This guide explains how to configure environment variables for the OrgMind frontend.

## Quick Start

1. Copy the example file:
   ```bash
   cp .env.local.example .env.local
   ```

2. Edit `.env.local` and fill in your actual values

3. Never commit `.env.local` to version control (it's in `.gitignore`)

4. Restart the development server after changing environment variables

## Required Variables

These variables MUST be set for the application to run:

### API Configuration
- `NEXT_PUBLIC_API_URL`: Backend API base URL
  - Development: `http://localhost:8080`
  - Production: `https://api.orgmind.com`
  - **Important**: Do not include trailing slash

### OAuth Configuration
- `NEXT_PUBLIC_OAUTH_REDIRECT_URL`: OAuth callback URL
  - Development: `http://localhost:3000/auth/callback`
  - Production: `https://orgmind.com/auth/callback`
  - **Important**: Must match backend `OAUTH_REDIRECT_URL`

## Understanding NEXT_PUBLIC_ Prefix

Variables prefixed with `NEXT_PUBLIC_` are exposed to the browser. This is required for:
- API calls from client components
- OAuth redirect URLs
- Feature flags used in client-side code

Variables without this prefix are only available during build time and server-side rendering.

## Optional Variables

### Feature Flags

Control which features are enabled:

```bash
NEXT_PUBLIC_ENABLE_GOOGLE_AUTH=true
NEXT_PUBLIC_ENABLE_OKTA_AUTH=true
NEXT_PUBLIC_ENABLE_OFFICE365_AUTH=true
NEXT_PUBLIC_ENABLE_FILE_UPLOAD=true
NEXT_PUBLIC_ENABLE_EDITOR=true
NEXT_PUBLIC_ENABLE_GRAPH_VISUALIZATION=true
```

### Upload Configuration

```bash
NEXT_PUBLIC_MAX_UPLOAD_SIZE_MB=50
NEXT_PUBLIC_ALLOWED_FILE_EXTENSIONS=.txt,.pdf,.doc,.docx
```

**Note**: These should match or be more restrictive than backend settings.

### Editor Configuration

```bash
NEXT_PUBLIC_EDITOR_AUTO_SAVE=true
NEXT_PUBLIC_EDITOR_AUTO_SAVE_INTERVAL=5000
```

### Graph Visualization

```bash
NEXT_PUBLIC_GRAPH_LAYOUT=force-directed
NEXT_PUBLIC_GRAPH_ANIMATION=true
NEXT_PUBLIC_GRAPH_NODE_SIZE_MIN=5
NEXT_PUBLIC_GRAPH_NODE_SIZE_MAX=20
```

### Analytics (Optional)

```bash
NEXT_PUBLIC_GA_MEASUREMENT_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SENTRY_DSN=https://xxxxx@sentry.io/xxxxx
```

## Environment-Specific Configuration

### Development (.env.local)
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
NEXT_PUBLIC_ENVIRONMENT=development
NEXT_PUBLIC_DEBUG_MODE=true
NEXT_PUBLIC_SHOW_DEV_INDICATORS=true
```

### Production (.env.production)
```bash
NEXT_PUBLIC_API_URL=https://api.orgmind.com
NEXT_PUBLIC_OAUTH_REDIRECT_URL=https://orgmind.com/auth/callback
NEXT_PUBLIC_ENVIRONMENT=production
NEXT_PUBLIC_DEBUG_MODE=false
NEXT_PUBLIC_SHOW_DEV_INDICATORS=false
```

## Environment File Priority

Next.js loads environment files in this order (later files override earlier ones):

1. `.env` - All environments
2. `.env.local` - All environments (ignored by git)
3. `.env.development` - Development only
4. `.env.development.local` - Development only (ignored by git)
5. `.env.production` - Production only
6. `.env.production.local` - Production only (ignored by git)

**Recommendation**: Use `.env.local` for local development overrides.

## Security Best Practices

1. **Never commit `.env.local`** to version control
2. **Never expose sensitive data** in `NEXT_PUBLIC_` variables
3. **Use HTTPS** in production for API URLs
4. **Validate environment variables** on application startup
5. **Use different OAuth credentials** for each environment

## Common Issues

### API Calls Failing

**Problem**: API calls return CORS errors or connection refused

**Solutions**:
- Verify `NEXT_PUBLIC_API_URL` is correct
- Check backend is running on the specified port
- Verify backend CORS configuration allows frontend origin
- Ensure no trailing slash in API URL

### OAuth Redirect Issues

**Problem**: OAuth callback fails or redirects to wrong URL

**Solutions**:
- Verify `NEXT_PUBLIC_OAUTH_REDIRECT_URL` matches OAuth provider configuration
- Check URL is accessible from the internet (for production)
- Ensure backend `OAUTH_REDIRECT_URL` matches frontend setting
- Verify OAuth provider has correct redirect URI configured

### Environment Variables Not Updating

**Problem**: Changes to `.env.local` not taking effect

**Solutions**:
- Restart the Next.js development server
- Clear `.next` cache: `rm -rf .next`
- Verify variable name has `NEXT_PUBLIC_` prefix if used in client components
- Check for typos in variable names

### File Upload Failures

**Problem**: File uploads fail with size or type errors

**Solutions**:
- Verify `NEXT_PUBLIC_MAX_UPLOAD_SIZE_MB` matches backend setting
- Check `NEXT_PUBLIC_ALLOWED_FILE_EXTENSIONS` includes the file type
- Ensure backend has sufficient upload limits configured
- Check network timeout settings for large files

## Development Tips

### Using Multiple Environments

Create separate environment files:

```bash
.env.local              # Local development
.env.staging.local      # Staging environment
.env.production.local   # Production environment
```

Load specific environment:
```bash
# Development (default)
npm run dev

# Production build
npm run build
npm start
```

### Debugging Environment Variables

Add this to a page to see all public environment variables:

```typescript
console.log('Environment:', {
  apiUrl: process.env.NEXT_PUBLIC_API_URL,
  environment: process.env.NEXT_PUBLIC_ENVIRONMENT,
  // Add other variables you want to check
});
```

**Warning**: Remove debug code before committing!

### Type Safety for Environment Variables

Create a type-safe environment configuration:

```typescript
// lib/env.ts
export const env = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL!,
  oauthRedirectUrl: process.env.NEXT_PUBLIC_OAUTH_REDIRECT_URL!,
  environment: process.env.NEXT_PUBLIC_ENVIRONMENT || 'development',
  // Add other variables
};

// Validate required variables
if (!env.apiUrl) {
  throw new Error('NEXT_PUBLIC_API_URL is required');
}
```

## Deployment

### Vercel

Set environment variables in Vercel dashboard:
1. Go to Project Settings > Environment Variables
2. Add each variable with appropriate environment (Production/Preview/Development)
3. Redeploy to apply changes

### Docker

Pass environment variables via docker-compose:

```yaml
services:
  frontend:
    build: ./frontend
    environment:
      - NEXT_PUBLIC_API_URL=${API_URL}
      - NEXT_PUBLIC_OAUTH_REDIRECT_URL=${OAUTH_REDIRECT_URL}
```

Or use `.env` file:

```bash
docker-compose --env-file .env.production up
```

## Support

For additional help, refer to:
- [Frontend README](./README.md)
- [Next.js Environment Variables Documentation](https://nextjs.org/docs/basic-features/environment-variables)
- [Design Document](../.kiro/specs/document-processing-platform/design.md)
