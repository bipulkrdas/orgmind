# Google Cloud Run Deployment Guide

This guide explains how to deploy OrgMind to Google Cloud Run using Cloud Build.

## Prerequisites

1. **Google Cloud Project**
   - Active GCP project with billing enabled
   - Project ID ready

2. **Required APIs** (enable these in GCP Console)
   ```bash
   gcloud services enable cloudbuild.googleapis.com
   gcloud services enable run.googleapis.com
   gcloud services enable containerregistry.googleapis.com
   gcloud services enable secretmanager.googleapis.com
   gcloud services enable sqladmin.googleapis.com
   ```

3. **Google Cloud SDK**
   ```bash
   # Install gcloud CLI
   # https://cloud.google.com/sdk/docs/install
   
   # Login and set project
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

4. **Required Services**
   - Cloud SQL PostgreSQL instance
   - Cloud Storage bucket (or AWS S3)
   - Zep Cloud account

## Setup Steps

### 1. Create Cloud SQL PostgreSQL Instance

```bash
# Create PostgreSQL instance
gcloud sql instances create orgmind-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --root-password=YOUR_SECURE_PASSWORD

# Create database
gcloud sql databases create orgmind --instance=orgmind-db

# Create user
gcloud sql users create orgmind_user \
  --instance=orgmind-db \
  --password=YOUR_USER_PASSWORD

# Get connection name
gcloud sql instances describe orgmind-db --format="value(connectionName)"
# Output: PROJECT_ID:REGION:INSTANCE_NAME
```

### 2. Store Secrets in Secret Manager

```bash
# Database URL
echo -n "postgresql://orgmind_user:YOUR_USER_PASSWORD@/orgmind?host=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME" | \
  gcloud secrets create DATABASE_URL --data-file=-

# JWT Secret (generate strong secret)
openssl rand -base64 32 | gcloud secrets create JWT_SECRET --data-file=-

# AWS Credentials
echo -n "YOUR_AWS_ACCESS_KEY_ID" | gcloud secrets create AWS_ACCESS_KEY_ID --data-file=-
echo -n "YOUR_AWS_SECRET_ACCESS_KEY" | gcloud secrets create AWS_SECRET_ACCESS_KEY --data-file=-

# Zep API Key
echo -n "YOUR_ZEP_API_KEY" | gcloud secrets create ZEP_API_KEY --data-file=-

# Grant Cloud Run access to secrets
PROJECT_NUMBER=$(gcloud projects describe YOUR_PROJECT_ID --format="value(projectNumber)")
gcloud secrets add-iam-policy-binding DATABASE_URL \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Repeat for all secrets
for SECRET in JWT_SECRET AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY ZEP_API_KEY; do
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
done
```

### 3. Configure Cloud Build

Update `cloudbuild.yaml` substitutions:

```yaml
substitutions:
  _PROJECT_ID: 'your-project-id'
  _REGION: 'us-central1'
  _AWS_REGION: 'us-east-1'
  _AWS_S3_BUCKET: 'your-bucket-name'
  _BACKEND_URL: 'https://orgmind-backend-xxxxx-uc.a.run.app'
  _FRONTEND_URL: 'https://orgmind-frontend-xxxxx-uc.a.run.app'
```

**Note**: You'll need to deploy once to get the Cloud Run URLs, then update and redeploy.

### 4. Grant Cloud Build Permissions

```bash
# Get Cloud Build service account
PROJECT_NUMBER=$(gcloud projects describe YOUR_PROJECT_ID --format="value(projectNumber)")
CLOUD_BUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

# Grant Cloud Run Admin role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/run.admin"

# Grant Service Account User role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/iam.serviceAccountUser"

# Grant Storage Admin role (for Container Registry)
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/storage.admin"
```

### 5. Deploy Application

#### Manual Deployment

```bash
# Deploy from local machine
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_PROJECT_ID=YOUR_PROJECT_ID,_REGION=us-central1,_AWS_REGION=us-east-1,_AWS_S3_BUCKET=your-bucket,_BACKEND_URL=https://your-backend-url,_FRONTEND_URL=https://your-frontend-url
```

#### Automated Deployment with Cloud Build Triggers

1. **Connect Repository**
   ```bash
   # Connect GitHub/GitLab/Bitbucket repository
   # Go to: Cloud Console > Cloud Build > Triggers > Connect Repository
   ```

2. **Create Build Trigger**
   ```bash
   gcloud builds triggers create github \
     --name="orgmind-deploy" \
     --repo-name="YOUR_REPO_NAME" \
     --repo-owner="YOUR_GITHUB_USERNAME" \
     --branch-pattern="^main$" \
     --build-config="cloudbuild.yaml" \
     --substitutions=_PROJECT_ID=YOUR_PROJECT_ID,_REGION=us-central1,_AWS_REGION=us-east-1,_AWS_S3_BUCKET=your-bucket
   ```

3. **Set Trigger Substitutions**
   - Go to Cloud Console > Cloud Build > Triggers
   - Edit trigger
   - Add substitution variables
   - Save

### 6. Configure OAuth Providers

After deployment, update OAuth provider settings with Cloud Run URLs:

#### Google OAuth
- Console: https://console.cloud.google.com/apis/credentials
- Authorized redirect URIs: `https://YOUR_FRONTEND_URL/auth/callback`

#### Okta
- Okta Admin Console > Applications
- Sign-in redirect URIs: `https://YOUR_FRONTEND_URL/auth/callback`

#### Office365
- Azure Portal > App registrations
- Redirect URIs: `https://YOUR_FRONTEND_URL/auth/callback`

### 7. Configure Custom Domain (Optional)

```bash
# Map custom domain to Cloud Run service
gcloud run domain-mappings create \
  --service=orgmind-frontend \
  --domain=app.yourdomain.com \
  --region=us-central1

gcloud run domain-mappings create \
  --service=orgmind-backend \
  --domain=api.yourdomain.com \
  --region=us-central1

# Follow DNS configuration instructions provided by the command
```

## Initial Deployment Process

Since frontend needs backend URL and vice versa for CORS, follow this process:

### First Deployment

1. **Deploy backend with temporary CORS**
   ```bash
   # Edit cloudbuild.yaml temporarily
   # Set CORS_ALLOWED_ORIGINS=*
   gcloud builds submit --config=cloudbuild.yaml
   ```

2. **Get backend URL**
   ```bash
   gcloud run services describe orgmind-backend --region=us-central1 --format="value(status.url)"
   ```

3. **Deploy frontend with backend URL**
   ```bash
   # Update cloudbuild.yaml with backend URL
   gcloud builds submit --config=cloudbuild.yaml
   ```

4. **Get frontend URL**
   ```bash
   gcloud run services describe orgmind-frontend --region=us-central1 --format="value(status.url)"
   ```

5. **Redeploy backend with correct CORS**
   ```bash
   # Update cloudbuild.yaml with frontend URL for CORS
   gcloud builds submit --config=cloudbuild.yaml
   ```

### Subsequent Deployments

After initial setup, just run:
```bash
gcloud builds submit --config=cloudbuild.yaml
```

Or push to main branch if using Cloud Build triggers.

## Environment Variables

### Backend Environment Variables

Set in Cloud Run service:

**From Secrets:**
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing secret
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `ZEP_API_KEY` - Zep Cloud API key

**Direct Environment Variables:**
- `ENVIRONMENT=production`
- `SERVER_PORT=8080`
- `AWS_REGION=us-east-1`
- `AWS_S3_BUCKET=your-bucket`
- `CORS_ALLOWED_ORIGINS=https://your-frontend-url`
- `OAUTH_REDIRECT_URL=https://your-frontend-url/auth/callback`
- `LOG_LEVEL=info`
- `LOG_FORMAT=json`

### Frontend Environment Variables

Set as build arguments in Dockerfile:
- `NEXT_PUBLIC_API_URL` - Backend URL
- `NEXT_PUBLIC_OAUTH_REDIRECT_URL` - OAuth callback URL
- `NEXT_PUBLIC_APP_NAME=OrgMind`
- `NEXT_PUBLIC_ENVIRONMENT=production`

## Monitoring and Logging

### View Logs

```bash
# Backend logs
gcloud run services logs read orgmind-backend --region=us-central1

# Frontend logs
gcloud run services logs read orgmind-frontend --region=us-central1

# Cloud Build logs
gcloud builds log BUILD_ID
```

### View Metrics

```bash
# Service metrics
gcloud run services describe orgmind-backend --region=us-central1

# Or use Cloud Console > Cloud Run > Service > Metrics
```

## Scaling Configuration

Cloud Run auto-scales based on traffic. Adjust in `cloudbuild.yaml`:

```yaml
--min-instances=0        # Minimum instances (0 = scale to zero)
--max-instances=10       # Maximum instances
--concurrency=80         # Requests per instance
--memory=512Mi           # Memory per instance
--cpu=1                  # CPU per instance
--timeout=300            # Request timeout (seconds)
```

For production with consistent traffic:
```yaml
--min-instances=1        # Keep at least 1 instance warm
--max-instances=100      # Allow more scaling
--memory=1Gi             # More memory
--cpu=2                  # More CPU
```

## Cost Optimization

1. **Scale to Zero**: Set `--min-instances=0` for low-traffic services
2. **Right-size Resources**: Start with minimal resources, increase if needed
3. **Use Cloud SQL Proxy**: Reduces connection overhead
4. **Enable CDN**: Use Cloud CDN for static assets
5. **Monitor Usage**: Use Cloud Monitoring to track costs

## Troubleshooting

### Build Fails

```bash
# Check build logs
gcloud builds log BUILD_ID

# Common issues:
# - Missing API permissions
# - Invalid substitution variables
# - Docker build errors
```

### Deployment Fails

```bash
# Check service status
gcloud run services describe orgmind-backend --region=us-central1

# Common issues:
# - Missing secrets
# - Invalid environment variables
# - Port configuration mismatch
```

### Service Not Accessible

```bash
# Check service URL
gcloud run services describe orgmind-backend --region=us-central1 --format="value(status.url)"

# Check IAM permissions
gcloud run services get-iam-policy orgmind-backend --region=us-central1

# Make service public
gcloud run services add-iam-policy-binding orgmind-backend \
  --region=us-central1 \
  --member="allUsers" \
  --role="roles/run.invoker"
```

### Database Connection Issues

```bash
# Test Cloud SQL connection
gcloud sql connect orgmind-db --user=orgmind_user

# Check Cloud SQL proxy configuration
# Ensure DATABASE_URL uses correct format:
# postgresql://user:pass@/dbname?host=/cloudsql/PROJECT:REGION:INSTANCE
```

### CORS Errors

```bash
# Update CORS configuration
gcloud run services update orgmind-backend \
  --region=us-central1 \
  --update-env-vars=CORS_ALLOWED_ORIGINS=https://your-frontend-url
```

## Security Best Practices

1. **Use Secret Manager** for all sensitive data
2. **Enable Cloud Armor** for DDoS protection
3. **Use VPC Connector** for private Cloud SQL access
4. **Enable Cloud Audit Logs** for compliance
5. **Implement Cloud IAM** with least privilege
6. **Use HTTPS only** (Cloud Run provides this by default)
7. **Rotate secrets regularly**
8. **Enable Binary Authorization** for image verification

## Rollback

```bash
# List revisions
gcloud run revisions list --service=orgmind-backend --region=us-central1

# Rollback to previous revision
gcloud run services update-traffic orgmind-backend \
  --region=us-central1 \
  --to-revisions=REVISION_NAME=100
```

## CI/CD Pipeline

For automated deployments:

1. **Push to main branch** triggers Cloud Build
2. **Cloud Build** builds and pushes Docker images
3. **Cloud Build** deploys to Cloud Run
4. **Cloud Run** performs rolling update
5. **Health checks** verify deployment
6. **Traffic** gradually shifts to new revision

## Support

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Cloud Build Documentation](https://cloud.google.com/build/docs)
- [Cloud SQL Documentation](https://cloud.google.com/sql/docs)
- [Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)

## Quick Reference

```bash
# Deploy
gcloud builds submit --config=cloudbuild.yaml

# View logs
gcloud run services logs read orgmind-backend --region=us-central1 --limit=50

# Update environment variable
gcloud run services update orgmind-backend --region=us-central1 --update-env-vars=KEY=VALUE

# Scale service
gcloud run services update orgmind-backend --region=us-central1 --min-instances=1 --max-instances=10

# Delete service
gcloud run services delete orgmind-backend --region=us-central1
```
