# OrgMind Environment Setup Guide

This guide provides a complete overview of environment configuration for the OrgMind platform.

## Overview

OrgMind uses environment variables for all configuration to ensure:
- Security: No credentials in source code
- Flexibility: Easy deployment across environments
- Portability: Same codebase for dev/staging/production

## Quick Setup

### Backend Setup

1. Navigate to backend directory:
   ```bash
   cd backend
   ```

2. Copy environment template:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` and configure required variables (see [Backend Environment Guide](./backend/ENVIRONMENT.md))

4. Start the backend:
   ```bash
   go run cmd/server/main.go
   ```

### Frontend Setup

1. Navigate to frontend directory:
   ```bash
   cd frontend
   ```

2. Copy environment template:
   ```bash
   cp .env.local.example .env.local
   ```

3. Edit `.env.local` and configure required variables (see [Frontend Environment Guide](./frontend/ENVIRONMENT.md))

4. Start the frontend:
   ```bash
   npm run dev
   ```

## Minimum Required Configuration

To run OrgMind locally with minimal setup:

### Backend (.env)
```bash
# Server
SERVER_PORT=8080

# Database (requires PostgreSQL running)
DATABASE_URL=postgresql://user:password@localhost:5432/orgmind?sslmode=disable

# JWT (generate with: openssl rand -base64 32)
JWT_SECRET=your-generated-secret-key-here

# AWS S3 (requires AWS account or MinIO)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=orgmind-documents

# Zep Cloud (requires Zep account)
ZEP_API_KEY=your-zep-api-key

# OAuth Redirect
OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### Frontend (.env.local)
```bash
# API
NEXT_PUBLIC_API_URL=http://localhost:8080

# OAuth
NEXT_PUBLIC_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
```

## Optional Features

### Email/Password Authentication
Already enabled by default. No additional configuration needed beyond database.

### Google OAuth
```bash
# Backend
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

Setup: https://console.cloud.google.com/apis/credentials

### Okta OAuth
```bash
# Backend
OKTA_DOMAIN=your-company.okta.com
OKTA_CLIENT_ID=your-okta-client-id
OKTA_CLIENT_SECRET=your-okta-client-secret
```

Setup: Okta Admin Console

### Office365 OAuth
```bash
# Backend
OFFICE365_CLIENT_ID=your-office365-client-id
OFFICE365_CLIENT_SECRET=your-office365-client-secret
```

Setup: https://portal.azure.com/ > App registrations

### Password Reset Email
```bash
# Backend
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM_EMAIL=noreply@orgmind.com
```

### AI-Powered Chat (Google Gemini)
```bash
# Backend
GEMINI_API_KEY=your-gemini-api-key
GEMINI_PROJECT_ID=your-project-id
GEMINI_LOCATION=us-central1
GEMINI_STORE_NAME=OrgMind Documents  # Optional: Display name for shared File Search store
```

Setup: https://aistudio.google.com/app/apikey

**Note:** OrgMind uses a single shared File Search store with metadata-based filtering for graph isolation. The store is created automatically at startup.

For detailed setup instructions, see [backend/GEMINI_SETUP.md](./backend/GEMINI_SETUP.md)

## Environment-Specific Configurations

### Development
- Use `localhost` URLs
- Enable debug logging
- Disable SSL for database (local only)
- Use test OAuth credentials

### Staging
- Use staging domain URLs
- Enable info logging
- Enable SSL for database
- Use staging OAuth credentials
- Enable rate limiting

### Production
- Use production domain URLs
- Enable error-only logging
- Require SSL for database
- Use production OAuth credentials
- Enable all security features
- Enable monitoring and analytics

## Security Checklist

- [ ] `.env` and `.env.local` are in `.gitignore`
- [ ] JWT_SECRET is strong (32+ characters)
- [ ] Database uses SSL in production
- [ ] OAuth redirect URLs match provider configuration
- [ ] AWS credentials have minimal required permissions
- [ ] CORS origins are restricted to your domains
- [ ] Rate limiting is enabled in production
- [ ] SMTP credentials are secure
- [ ] All secrets are rotated regularly

## Validation

Both backend and frontend validate required environment variables on startup:

### Backend Validation
The Go application checks for required variables and exits with descriptive errors if any are missing.

### Frontend Validation
Next.js will show build errors if required `NEXT_PUBLIC_` variables are missing.

## Troubleshooting

### Backend Won't Start
1. Check all required variables are set in `.env`
2. Verify database is running and accessible
3. Test database connection string
4. Check AWS credentials are valid
5. Verify Zep API key is active

### Frontend Won't Connect to Backend
1. Verify `NEXT_PUBLIC_API_URL` matches backend port
2. Check backend is running
3. Verify CORS configuration allows frontend origin
4. Check for typos in environment variable names

### OAuth Not Working
1. Verify redirect URLs match in all three places:
   - OAuth provider configuration
   - Backend `OAUTH_REDIRECT_URL`
   - Frontend `NEXT_PUBLIC_OAUTH_REDIRECT_URL`
2. Check client ID and secret are correct
3. Verify OAuth provider is properly configured
4. Check callback URL is accessible

### File Uploads Failing
1. Verify AWS credentials are correct
2. Check S3 bucket exists and has proper permissions
3. Verify upload size limits match between frontend and backend
4. Check network connectivity to S3

## Additional Resources

- [Backend Environment Guide](./backend/ENVIRONMENT.md) - Detailed backend configuration
- [Frontend Environment Guide](./frontend/ENVIRONMENT.md) - Detailed frontend configuration
- [Backend README](./backend/README.md) - Backend setup and development
- [Frontend README](./frontend/README.md) - Frontend setup and development
- [Design Document](./.kiro/specs/document-processing-platform/design.md) - Architecture details
- [Requirements Document](./.kiro/specs/document-processing-platform/requirements.md) - System requirements

## Getting Help

If you encounter issues:

1. Check the troubleshooting sections in this guide
2. Review the detailed environment guides for backend and frontend
3. Verify all required services are running (PostgreSQL, backend, frontend)
4. Check application logs for specific error messages
5. Ensure environment variables are properly formatted

## Example: Complete Local Setup

```bash
# 1. Start PostgreSQL (using Docker)
docker run --name orgmind-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=orgmind -p 5432:5432 -d postgres

# 2. Start MinIO (S3-compatible storage)
docker run --name orgmind-minio -p 9000:9000 -p 9001:9001 -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin -d minio/minio server /data --console-address ":9001"

# 3. Configure backend
cd backend
cp .env.example .env
# Edit .env with your values

# 4. Start backend
go run cmd/server/main.go

# 5. Configure frontend (in new terminal)
cd frontend
cp .env.local.example .env.local
# Edit .env.local with your values

# 6. Start frontend
npm install
npm run dev

# 7. Access application
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# MinIO Console: http://localhost:9001
```

## Production Deployment

OrgMind supports multiple deployment options:

### Google Cloud Run (Recommended for Serverless)

See [DEPLOYMENT.md](./DEPLOYMENT.md) for complete Google Cloud Run deployment guide.

**Quick Deploy:**
```bash
./deploy.sh
```

**Benefits:**
- Automatic scaling (including scale-to-zero)
- Pay only for actual usage
- Managed infrastructure
- Built-in HTTPS and CDN
- Easy CI/CD with Cloud Build

### Docker / Docker Compose

See [DOCKER.md](./DOCKER.md) for complete Docker deployment guide.

**Quick Start:**
```bash
docker-compose up -d
```

**Benefits:**
- Run anywhere (local, VPS, cloud)
- Full control over infrastructure
- Easy local testing
- Portable across environments

### Other Platforms

OrgMind can be deployed to:
- **AWS**: ECS, Fargate, or Elastic Beanstalk
- **Azure**: Container Instances or App Service
- **Kubernetes**: Any managed or self-hosted cluster
- **Vercel/Netlify**: Frontend only (backend separately)

### Production Checklist

1. **Use managed services:**
   - Database: AWS RDS PostgreSQL or Google Cloud SQL
   - Storage: AWS S3 or Google Cloud Storage
   - Secrets: AWS Secrets Manager or Google Secret Manager

2. **Set environment variables** in your deployment platform

3. **Enable security features:**
   - SSL/TLS for all connections
   - Rate limiting
   - CORS restrictions
   - Secure OAuth redirect URLs
   - Non-root container users

4. **Monitor and log:**
   - Application logs
   - Error tracking (Sentry)
   - Performance monitoring
   - API usage metrics

5. **Regular maintenance:**
   - Rotate credentials
   - Update dependencies
   - Review security settings
   - Monitor resource usage
