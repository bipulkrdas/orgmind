# Google Cloud Run Quick Start Guide

This is a condensed guide to get OrgMind running on Google Cloud Run quickly.

## Prerequisites

- Google Cloud account with billing enabled
- `gcloud` CLI installed and configured
- Zep Cloud API key

## 5-Minute Setup

### 1. Enable Required APIs

```bash
gcloud services enable \
  cloudbuild.googleapis.com \
  run.googleapis.com \
  containerregistry.googleapis.com \
  secretmanager.googleapis.com \
  sqladmin.googleapis.com
```

### 2. Create Database

```bash
# Create PostgreSQL instance (takes ~5 minutes)
gcloud sql instances create orgmind-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --root-password=CHANGE_THIS_PASSWORD

# Create database and user
gcloud sql databases create orgmind --instance=orgmind-db
gcloud sql users create orgmind_user \
  --instance=orgmind-db \
  --password=CHANGE_THIS_PASSWORD
```

### 3. Store Secrets

```bash
# Get database connection name
DB_CONN=$(gcloud sql instances describe orgmind-db --format="value(connectionName)")

# Create secrets
echo -n "postgresql://orgmind_user:CHANGE_THIS_PASSWORD@/orgmind?host=/cloudsql/$DB_CONN" | \
  gcloud secrets create DATABASE_URL --data-file=-

openssl rand -base64 32 | gcloud secrets create JWT_SECRET --data-file=-

echo -n "YOUR_AWS_ACCESS_KEY" | gcloud secrets create AWS_ACCESS_KEY_ID --data-file=-
echo -n "YOUR_AWS_SECRET_KEY" | gcloud secrets create AWS_SECRET_ACCESS_KEY --data-file=-
echo -n "YOUR_ZEP_API_KEY" | gcloud secrets create ZEP_API_KEY --data-file=-

# Grant access to secrets
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")
for SECRET in DATABASE_URL JWT_SECRET AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY ZEP_API_KEY; do
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
done
```

### 4. Grant Cloud Build Permissions

```bash
PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
CLOUD_BUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/storage.admin"
```

### 5. Update Configuration

Edit `cloudbuild.yaml` and update these values:

```yaml
substitutions:
  _PROJECT_ID: 'your-project-id'  # Your GCP project ID
  _REGION: 'us-central1'
  _AWS_REGION: 'us-east-1'        # Your AWS region
  _AWS_S3_BUCKET: 'your-bucket'   # Your S3 bucket name
```

### 6. First Deployment

```bash
# Deploy (this will take 5-10 minutes)
./deploy.sh
```

Or manually:

```bash
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_PROJECT_ID=$(gcloud config get-value project),_REGION=us-central1,_AWS_REGION=us-east-1,_AWS_S3_BUCKET=your-bucket
```

### 7. Get Service URLs

```bash
# Backend URL
gcloud run services describe orgmind-backend --region=us-central1 --format="value(status.url)"

# Frontend URL
gcloud run services describe orgmind-frontend --region=us-central1 --format="value(status.url)"
```

### 8. Update OAuth Providers

Configure OAuth redirect URIs in your OAuth providers:
- Google: https://console.cloud.google.com/apis/credentials
- Okta: Okta Admin Console
- Office365: https://portal.azure.com/

Set redirect URI to: `https://YOUR_FRONTEND_URL/auth/callback`

### 9. Redeploy with Correct URLs

Update `cloudbuild.yaml` with the actual service URLs:

```yaml
substitutions:
  _BACKEND_URL: 'https://orgmind-backend-xxxxx-uc.a.run.app'
  _FRONTEND_URL: 'https://orgmind-frontend-xxxxx-uc.a.run.app'
```

Then redeploy:

```bash
./deploy.sh
```

## Done! 🎉

Your application is now running on Google Cloud Run:
- Frontend: https://orgmind-frontend-xxxxx-uc.a.run.app
- Backend: https://orgmind-backend-xxxxx-uc.a.run.app

## Common Issues

### "Permission denied" errors
```bash
# Make deploy script executable
chmod +x deploy.sh
```

### "Secret not found" errors
```bash
# Verify secrets exist
gcloud secrets list

# Recreate missing secret
echo -n "value" | gcloud secrets create SECRET_NAME --data-file=-
```

### "Service not found" errors
```bash
# Check if services are deployed
gcloud run services list --region=us-central1
```

### Database connection errors
```bash
# Verify database is running
gcloud sql instances list

# Check connection string in secret
gcloud secrets versions access latest --secret=DATABASE_URL
```

## Next Steps

1. **Set up custom domain**: See [DEPLOYMENT.md](./DEPLOYMENT.md#6-configure-custom-domain-optional)
2. **Enable monitoring**: Set up Cloud Monitoring and Logging
3. **Configure CI/CD**: Set up Cloud Build triggers for automatic deployment
4. **Optimize costs**: Adjust min/max instances based on traffic
5. **Add CDN**: Enable Cloud CDN for better performance

## Cost Estimate

With default configuration (scale-to-zero enabled):

- **Cloud Run**: ~$0-10/month (depends on traffic)
- **Cloud SQL**: ~$10/month (db-f1-micro)
- **Cloud Build**: 120 free builds/day
- **Container Registry**: ~$0.10/GB/month
- **Secret Manager**: $0.06 per 10,000 accesses

**Total**: ~$10-20/month for low traffic

## Support

For detailed documentation:
- [Full Deployment Guide](./DEPLOYMENT.md)
- [Docker Guide](./DOCKER.md)
- [Environment Setup](./ENVIRONMENT_SETUP.md)
- [Google Cloud Run Docs](https://cloud.google.com/run/docs)
