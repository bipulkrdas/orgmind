# Environment Configuration Guide

This guide explains how to configure environment variables for the OrgMind backend.

## Quick Start

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and fill in your actual values

3. Never commit `.env` to version control (it's in `.gitignore`)

## Required Variables

These variables MUST be set for the application to run:

### Database
- `DATABASE_URL`: PostgreSQL connection string
  - Format: `postgresql://username:password@host:port/database?sslmode=disable`
  - Production: Use `sslmode=require`

### JWT Authentication
- `JWT_SECRET`: Secret key for signing JWT tokens
  - Generate with: `openssl rand -base64 32`
  - Minimum 32 characters recommended

### AWS S3 Storage
- `AWS_REGION`: AWS region (e.g., `us-east-1`)
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `AWS_S3_BUCKET`: S3 bucket name

### Zep Cloud
- `ZEP_API_KEY`: Zep Cloud API key
  - Obtain from: https://www.getzep.com/

## Optional Variables

### OAuth Providers

Only configure the OAuth providers you want to enable:

#### Google OAuth
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- Obtain from: https://console.cloud.google.com/apis/credentials

#### Okta OAuth
- `OKTA_DOMAIN`
- `OKTA_CLIENT_ID`
- `OKTA_CLIENT_SECRET`
- Obtain from Okta admin console

#### Office365 OAuth
- `OFFICE365_CLIENT_ID`
- `OFFICE365_CLIENT_SECRET`
- Obtain from: https://portal.azure.com/ > App registrations

### Email (Password Reset)
- `SMTP_HOST`: SMTP server host
- `SMTP_PORT`: SMTP server port (default: 587)
- `SMTP_USERNAME`: SMTP username
- `SMTP_PASSWORD`: SMTP password
- `SMTP_FROM_EMAIL`: From email address
- `SMTP_FROM_NAME`: From name (default: OrgMind)

## Environment-Specific Configuration

### Development
```bash
ENVIRONMENT=development
SERVER_PORT=8080
DATABASE_URL=postgresql://user:password@localhost:5432/orgmind?sslmode=disable
OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
CORS_ALLOWED_ORIGINS=http://localhost:3000
LOG_LEVEL=debug
```

### Production
```bash
ENVIRONMENT=production
SERVER_PORT=8080
DATABASE_URL=postgresql://user:password@prod-db:5432/orgmind?sslmode=require
OAUTH_REDIRECT_URL=https://orgmind.com/auth/callback
CORS_ALLOWED_ORIGINS=https://orgmind.com
LOG_LEVEL=info
LOG_FORMAT=json
RATE_LIMIT_ENABLED=true
```

## Security Best Practices

1. **Never commit `.env` files** to version control
2. **Use strong secrets** for JWT_SECRET (minimum 32 characters)
3. **Enable SSL** for database connections in production (`sslmode=require`)
4. **Rotate credentials** regularly
5. **Use IAM roles** instead of access keys when running on AWS
6. **Enable rate limiting** in production
7. **Use environment-specific values** for OAuth redirect URLs

## Validation

The application validates required environment variables on startup. If any required variables are missing, the application will fail to start with a descriptive error message.

## Alternative Storage Providers

### MinIO (S3-Compatible)
```bash
AWS_S3_ENDPOINT=http://localhost:9000
AWS_S3_PATH_STYLE=true
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_S3_BUCKET=orgmind-documents
```

### Google Cloud Storage
Configure using S3-compatible API with appropriate credentials.

## Troubleshooting

### Database Connection Issues
- Verify `DATABASE_URL` format
- Check database is running and accessible
- Verify credentials are correct
- Check SSL mode matches database configuration

### S3 Upload Failures
- Verify AWS credentials are correct
- Check bucket exists and has proper permissions
- Verify IAM user has `s3:PutObject` permission
- Check bucket region matches `AWS_REGION`

### Zep API Errors
- Verify `ZEP_API_KEY` is valid
- Check Zep Cloud service status
- Verify network connectivity to Zep API
- Check retry configuration if experiencing transient failures

### OAuth Issues
- Verify redirect URLs match OAuth provider configuration
- Check client ID and secret are correct
- Verify OAuth provider is properly configured
- Check callback URL is accessible from the internet (for production)

## Support

For additional help, refer to:
- [Backend README](./README.md)
- [Design Document](../.kiro/specs/document-processing-platform/design.md)
- [Requirements Document](../.kiro/specs/document-processing-platform/requirements.md)
