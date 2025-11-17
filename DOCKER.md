# Docker Deployment Guide

This guide explains how to build and run OrgMind using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- (Optional) Zep Cloud API key

## Quick Start with Docker Compose

### 1. Set Environment Variables

Create a `.env` file in the project root:

```bash
# Zep Cloud API Key (required)
ZEP_API_KEY=your-zep-api-key-here
```

### 2. Start All Services

```bash
# Start all services (PostgreSQL, MinIO, Backend, Frontend)
docker-compose up -d

# View logs
docker-compose logs -f

# Check service status
docker-compose ps
```

### 3. Access Services

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **PostgreSQL**: localhost:5432 (orgmind_user/orgmind_password)

### 4. Create MinIO Bucket

```bash
# Access MinIO container
docker exec -it orgmind-minio sh

# Create bucket using mc (MinIO Client)
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/orgmind-documents
mc anonymous set download local/orgmind-documents
exit
```

Or use the MinIO Console at http://localhost:9001

### 5. Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (deletes data)
docker-compose down -v
```

## Building Individual Services

### Backend

```bash
# Build backend image
docker build -t orgmind-backend:latest -f backend/Dockerfile ./backend

# Run backend container
docker run -d \
  --name orgmind-backend \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/orgmind" \
  -e JWT_SECRET="your-secret-key" \
  -e AWS_REGION="us-east-1" \
  -e AWS_ACCESS_KEY_ID="your-key" \
  -e AWS_SECRET_ACCESS_KEY="your-secret" \
  -e AWS_S3_BUCKET="your-bucket" \
  -e ZEP_API_KEY="your-zep-key" \
  -e CORS_ALLOWED_ORIGINS="http://localhost:3000" \
  orgmind-backend:latest

# View logs
docker logs -f orgmind-backend

# Stop container
docker stop orgmind-backend
docker rm orgmind-backend
```

### Frontend

```bash
# Build frontend image with build args
docker build -t orgmind-frontend:latest \
  --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080 \
  --build-arg NEXT_PUBLIC_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback \
  -f frontend/Dockerfile \
  ./frontend

# Run frontend container
docker run -d \
  --name orgmind-frontend \
  -p 3000:3000 \
  orgmind-frontend:latest

# View logs
docker logs -f orgmind-frontend

# Stop container
docker stop orgmind-frontend
docker rm orgmind-frontend
```

## Multi-Stage Build Benefits

Both Dockerfiles use multi-stage builds for:

1. **Smaller Images**: Only runtime dependencies in final image
2. **Security**: No build tools in production image
3. **Speed**: Cached layers speed up rebuilds
4. **Separation**: Build and runtime environments isolated

### Backend Image Size

- Builder stage: ~500MB (Go compiler + dependencies)
- Final image: ~20MB (Alpine + binary)

### Frontend Image Size

- Builder stage: ~1GB (Node + dependencies + build)
- Final image: ~150MB (Node runtime + built app)

## Health Checks

Both services include health checks:

### Backend Health Check

```bash
curl http://localhost:8080/health
# Response: {"status":"healthy","service":"orgmind-backend"}
```

### Frontend Health Check

```bash
curl http://localhost:3000/api/health
# Response: {"status":"healthy","service":"orgmind-frontend","timestamp":"..."}
```

## Environment Variables

### Backend Required Variables

```bash
DATABASE_URL=postgresql://user:pass@host:5432/orgmind
JWT_SECRET=your-secret-key
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
AWS_S3_BUCKET=your-bucket
ZEP_API_KEY=your-zep-key
```

### Frontend Build Arguments

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
NEXT_PUBLIC_APP_NAME=OrgMind
NEXT_PUBLIC_ENVIRONMENT=production
```

## Production Deployment

### Using Docker Compose in Production

1. **Update docker-compose.yml** for production:

```yaml
services:
  backend:
    environment:
      ENVIRONMENT: production
      LOG_LEVEL: info
      LOG_FORMAT: json
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
      # ... other secrets from environment
    restart: unless-stopped
    
  frontend:
    restart: unless-stopped
```

2. **Use external database and storage**:
   - Remove postgres and minio services
   - Use managed PostgreSQL (AWS RDS, Google Cloud SQL)
   - Use managed storage (AWS S3, Google Cloud Storage)

3. **Set production environment variables**:

```bash
export DATABASE_URL="postgresql://..."
export JWT_SECRET="..."
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
export ZEP_API_KEY="..."

docker-compose up -d
```

### Using Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Create secrets
echo "your-jwt-secret" | docker secret create jwt_secret -
echo "your-db-url" | docker secret create database_url -

# Deploy stack
docker stack deploy -c docker-compose.yml orgmind

# Check services
docker service ls

# View logs
docker service logs orgmind_backend
docker service logs orgmind_frontend

# Scale services
docker service scale orgmind_backend=3
docker service scale orgmind_frontend=3

# Remove stack
docker stack rm orgmind
```

### Using Kubernetes

Convert Docker Compose to Kubernetes:

```bash
# Install kompose
curl -L https://github.com/kubernetes/kompose/releases/download/v1.31.2/kompose-linux-amd64 -o kompose
chmod +x kompose
sudo mv kompose /usr/local/bin/

# Convert to Kubernetes manifests
kompose convert -f docker-compose.yml

# Apply to cluster
kubectl apply -f .

# Check status
kubectl get pods
kubectl get services

# View logs
kubectl logs -f deployment/orgmind-backend
kubectl logs -f deployment/orgmind-frontend
```

## Troubleshooting

### Backend Container Exits Immediately

```bash
# Check logs
docker logs orgmind-backend

# Common issues:
# - Missing required environment variables
# - Database connection failed
# - Invalid AWS credentials
```

### Frontend Build Fails

```bash
# Check build logs
docker build -t orgmind-frontend:latest -f frontend/Dockerfile ./frontend

# Common issues:
# - Missing build arguments
# - npm install failures
# - Build errors in Next.js
```

### Cannot Connect to Backend from Frontend

```bash
# Check network
docker network ls
docker network inspect orgmind_default

# Ensure services are on same network
# Use service names (backend, frontend) not localhost
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
docker exec -it orgmind-postgres psql -U orgmind_user -d orgmind

# Check connection string format
# postgresql://user:pass@host:5432/database?sslmode=disable
```

### MinIO Connection Issues

```bash
# Check MinIO is running
docker ps | grep minio

# Access MinIO console
# http://localhost:9001

# Check bucket exists
docker exec -it orgmind-minio mc ls local/
```

## Development Workflow

### 1. Local Development with Hot Reload

For development, use local dev servers instead of Docker:

```bash
# Terminal 1: Backend
cd backend
go run cmd/server/main.go

# Terminal 2: Frontend
cd frontend
npm run dev

# Terminal 3: Services only
docker-compose up postgres minio
```

### 2. Test Docker Build Locally

```bash
# Build and test backend
docker build -t orgmind-backend:test -f backend/Dockerfile ./backend
docker run --rm orgmind-backend:test ./server --version

# Build and test frontend
docker build -t orgmind-frontend:test \
  --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -f frontend/Dockerfile ./frontend
```

### 3. Debug Running Container

```bash
# Execute shell in running container
docker exec -it orgmind-backend sh
docker exec -it orgmind-frontend sh

# Check environment variables
docker exec orgmind-backend env

# Check processes
docker exec orgmind-backend ps aux

# Check network connectivity
docker exec orgmind-backend wget -O- http://postgres:5432
```

## Performance Optimization

### 1. Use BuildKit

```bash
# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1

# Build with BuildKit
docker build -t orgmind-backend:latest -f backend/Dockerfile ./backend
```

### 2. Layer Caching

```bash
# Build with cache from registry
docker build \
  --cache-from orgmind-backend:latest \
  -t orgmind-backend:latest \
  -f backend/Dockerfile \
  ./backend
```

### 3. Multi-Platform Builds

```bash
# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t orgmind-backend:latest \
  -f backend/Dockerfile \
  ./backend
```

## Security Best Practices

1. **Non-root User**: Both images run as non-root user
2. **Minimal Base**: Use Alpine Linux for smaller attack surface
3. **No Secrets in Image**: Use environment variables or secrets
4. **Scan Images**: Regularly scan for vulnerabilities

```bash
# Scan images for vulnerabilities
docker scan orgmind-backend:latest
docker scan orgmind-frontend:latest
```

5. **Read-only Filesystem**: Run containers with read-only root

```bash
docker run --read-only --tmpfs /tmp orgmind-backend:latest
```

## Monitoring

### Container Stats

```bash
# View resource usage
docker stats

# View specific container
docker stats orgmind-backend
```

### Export Logs

```bash
# Export logs to file
docker logs orgmind-backend > backend.log 2>&1
docker logs orgmind-frontend > frontend.log 2>&1
```

### Health Check Status

```bash
# Check health status
docker inspect --format='{{.State.Health.Status}}' orgmind-backend
docker inspect --format='{{.State.Health.Status}}' orgmind-frontend
```

## Cleanup

```bash
# Remove all OrgMind containers
docker ps -a | grep orgmind | awk '{print $1}' | xargs docker rm -f

# Remove all OrgMind images
docker images | grep orgmind | awk '{print $3}' | xargs docker rmi -f

# Remove unused volumes
docker volume prune

# Remove unused networks
docker network prune

# Complete cleanup
docker system prune -a --volumes
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
