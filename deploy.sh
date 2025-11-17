#!/bin/bash

# OrgMind Deployment Script for Google Cloud Run
# This script helps deploy the application to Google Cloud Run

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    print_error "gcloud CLI is not installed. Please install it first:"
    echo "https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Get project ID
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
if [ -z "$PROJECT_ID" ]; then
    print_error "No GCP project set. Run: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

print_info "Using GCP Project: $PROJECT_ID"

# Configuration
REGION=${REGION:-"us-central1"}
BACKEND_SERVICE="orgmind-backend"
FRONTEND_SERVICE="orgmind-frontend"

# Parse command line arguments
DEPLOY_BACKEND=true
DEPLOY_FRONTEND=true
SKIP_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --backend-only)
            DEPLOY_FRONTEND=false
            shift
            ;;
        --frontend-only)
            DEPLOY_BACKEND=false
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --help)
            echo "Usage: ./deploy.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --backend-only    Deploy only backend service"
            echo "  --frontend-only   Deploy only frontend service"
            echo "  --skip-build      Skip Docker build (use existing images)"
            echo "  --region REGION   GCP region (default: us-central1)"
            echo "  --help            Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Run './deploy.sh --help' for usage information"
            exit 1
            ;;
    esac
done

# Get backend URL if deploying frontend
if [ "$DEPLOY_FRONTEND" = true ]; then
    print_info "Getting backend URL..."
    BACKEND_URL=$(gcloud run services describe $BACKEND_SERVICE --region=$REGION --format="value(status.url)" 2>/dev/null || echo "")
    
    if [ -z "$BACKEND_URL" ]; then
        print_warning "Backend service not found. Deploy backend first or use --backend-only flag"
        if [ "$DEPLOY_BACKEND" = false ]; then
            exit 1
        fi
    else
        print_info "Backend URL: $BACKEND_URL"
    fi
fi

# Get frontend URL if deploying backend
if [ "$DEPLOY_BACKEND" = true ]; then
    print_info "Getting frontend URL..."
    FRONTEND_URL=$(gcloud run services describe $FRONTEND_SERVICE --region=$REGION --format="value(status.url)" 2>/dev/null || echo "")
    
    if [ -z "$FRONTEND_URL" ]; then
        print_warning "Frontend service not found. This is the first deployment."
        FRONTEND_URL="https://orgmind-frontend-temp.a.run.app"  # Temporary placeholder
    else
        print_info "Frontend URL: $FRONTEND_URL"
    fi
fi

# Check required environment variables
print_info "Checking required secrets..."
REQUIRED_SECRETS=("DATABASE_URL" "JWT_SECRET" "AWS_ACCESS_KEY_ID" "AWS_SECRET_ACCESS_KEY" "ZEP_API_KEY")
MISSING_SECRETS=()

for SECRET in "${REQUIRED_SECRETS[@]}"; do
    if ! gcloud secrets describe $SECRET &>/dev/null; then
        MISSING_SECRETS+=($SECRET)
    fi
done

if [ ${#MISSING_SECRETS[@]} -gt 0 ]; then
    print_error "Missing required secrets in Secret Manager:"
    for SECRET in "${MISSING_SECRETS[@]}"; do
        echo "  - $SECRET"
    done
    echo ""
    echo "Create secrets with:"
    echo "  echo -n 'value' | gcloud secrets create SECRET_NAME --data-file=-"
    exit 1
fi

print_info "All required secrets found"

# Prompt for AWS configuration
read -p "AWS Region (default: us-east-1): " AWS_REGION
AWS_REGION=${AWS_REGION:-"us-east-1"}

read -p "AWS S3 Bucket name: " AWS_S3_BUCKET
if [ -z "$AWS_S3_BUCKET" ]; then
    print_error "AWS S3 Bucket name is required"
    exit 1
fi

# Build and deploy using Cloud Build
print_info "Starting deployment..."

SUBSTITUTIONS="_PROJECT_ID=$PROJECT_ID"
SUBSTITUTIONS="$SUBSTITUTIONS,_REGION=$REGION"
SUBSTITUTIONS="$SUBSTITUTIONS,_AWS_REGION=$AWS_REGION"
SUBSTITUTIONS="$SUBSTITUTIONS,_AWS_S3_BUCKET=$AWS_S3_BUCKET"

if [ -n "$BACKEND_URL" ]; then
    SUBSTITUTIONS="$SUBSTITUTIONS,_BACKEND_URL=$BACKEND_URL"
fi

if [ -n "$FRONTEND_URL" ]; then
    SUBSTITUTIONS="$SUBSTITUTIONS,_FRONTEND_URL=$FRONTEND_URL"
fi

print_info "Submitting build to Cloud Build..."
gcloud builds submit \
    --config=cloudbuild.yaml \
    --substitutions="$SUBSTITUTIONS"

# Get service URLs
if [ "$DEPLOY_BACKEND" = true ]; then
    BACKEND_URL=$(gcloud run services describe $BACKEND_SERVICE --region=$REGION --format="value(status.url)")
    print_info "Backend deployed: $BACKEND_URL"
fi

if [ "$DEPLOY_FRONTEND" = true ]; then
    FRONTEND_URL=$(gcloud run services describe $FRONTEND_SERVICE --region=$REGION --format="value(status.url)")
    print_info "Frontend deployed: $FRONTEND_URL"
fi

# Print next steps
echo ""
print_info "Deployment complete!"
echo ""
echo "Next steps:"
echo "1. Update OAuth provider redirect URIs to: ${FRONTEND_URL}/auth/callback"
echo "2. If this is the first deployment, redeploy to update CORS settings:"
echo "   ./deploy.sh"
echo "3. Test the application:"
echo "   Frontend: $FRONTEND_URL"
echo "   Backend:  $BACKEND_URL"
echo "   Health:   $BACKEND_URL/health"
echo ""
