# Database Deployment Guide

Quick reference for deploying OrgMind database schema in different environments.

## Production Deployment (First Time)

For a **brand new production database**:

```bash
# 1. Set your database URL
export DATABASE_URL="postgresql://user:password@host:5432/orgmind?sslmode=require"

# 2. Run consolidated schema
migrate -path ./migrations/consolidated -database "$DATABASE_URL" up

# 3. Verify
migrate -path ./migrations/consolidated -database "$DATABASE_URL" version
```

Expected output: `1` (one migration applied)

## Production Update (Existing Database)

For **updating an existing production database**:

```bash
# 1. Set your database URL
export DATABASE_URL="postgresql://user:password@host:5432/orgmind?sslmode=require"

# 2. Check current version
migrate -path ./migrations -database "$DATABASE_URL" version

# 3. Run incremental migrations
migrate -path ./migrations -database "$DATABASE_URL" up

# 4. Verify new version
migrate -path ./migrations -database "$DATABASE_URL" version
```

## Development Setup

For **local development**:

```bash
# Using Docker Compose (recommended)
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
sleep 5

# Run migrations
export DATABASE_URL="postgresql://orgmind:orgmind@localhost:5432/orgmind?sslmode=disable"
migrate -path ./migrations/consolidated -database "$DATABASE_URL" up
```

## Cloud Run / GCP Deployment

For **Google Cloud Run** with Cloud SQL:

```bash
# 1. Create Cloud SQL instance (if not exists)
gcloud sql instances create orgmind-db \
  --database-version=POSTGRES_14 \
  --tier=db-f1-micro \
  --region=us-central1

# 2. Create database
gcloud sql databases create orgmind --instance=orgmind-db

# 3. Set up Cloud SQL Proxy (for migrations)
cloud_sql_proxy -instances=PROJECT:REGION:orgmind-db=tcp:5432 &

# 4. Run migrations
export DATABASE_URL="postgresql://postgres:PASSWORD@localhost:5432/orgmind?sslmode=disable"
migrate -path ./migrations/consolidated -database "$DATABASE_URL" up

# 5. Deploy application with connection string
gcloud run deploy orgmind-backend \
  --image gcr.io/PROJECT/orgmind-backend \
  --add-cloudsql-instances PROJECT:REGION:orgmind-db \
  --set-env-vars DATABASE_URL="postgresql://postgres:PASSWORD@/orgmind?host=/cloudsql/PROJECT:REGION:orgmind-db"
```

## AWS RDS Deployment

For **AWS RDS PostgreSQL**:

```bash
# 1. Create RDS instance (via AWS Console or CLI)
aws rds create-db-instance \
  --db-instance-identifier orgmind-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username orgmind \
  --master-user-password YOUR_PASSWORD \
  --allocated-storage 20

# 2. Wait for instance to be available
aws rds wait db-instance-available --db-instance-identifier orgmind-db

# 3. Get endpoint
ENDPOINT=$(aws rds describe-db-instances \
  --db-instance-identifier orgmind-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text)

# 4. Run migrations
export DATABASE_URL="postgresql://orgmind:YOUR_PASSWORD@$ENDPOINT:5432/postgres?sslmode=require"
migrate -path ./migrations/consolidated -database "$DATABASE_URL" up
```

## Docker Deployment

For **Docker Compose** production setup:

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: orgmind
      POSTGRES_USER: orgmind
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations/consolidated:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

  backend:
    image: orgmind-backend:latest
    environment:
      DATABASE_URL: postgresql://orgmind:${DB_PASSWORD}@postgres:5432/orgmind?sslmode=disable
    depends_on:
      - postgres

volumes:
  postgres_data:
```

Deploy:
```bash
export DB_PASSWORD="your-secure-password"
docker-compose -f docker-compose.prod.yml up -d
```

## Kubernetes Deployment

For **Kubernetes** with PostgreSQL operator:

```yaml
# migrations-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: orgmind-migrations
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: migrate/migrate:v4.15.2
        args:
          - "-path=/migrations"
          - "-database=$(DATABASE_URL)"
          - "up"
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: orgmind-db-secret
              key: database-url
        volumeMounts:
        - name: migrations
          mountPath: /migrations
      volumes:
      - name: migrations
        configMap:
          name: orgmind-migrations
      restartPolicy: OnFailure
```

Apply:
```bash
kubectl create configmap orgmind-migrations --from-file=./migrations/consolidated/
kubectl apply -f migrations-job.yaml
```

## Rollback Procedures

### Rollback Last Migration

```bash
migrate -path ./migrations -database "$DATABASE_URL" down 1
```

### Rollback to Specific Version

```bash
migrate -path ./migrations -database "$DATABASE_URL" force <version>
```

### Complete Teardown (DANGER!)

```bash
# This will DROP ALL TABLES
migrate -path ./migrations/consolidated -database "$DATABASE_URL" down
```

## Verification Checklist

After deployment, verify:

- [ ] All tables exist: `users`, `graphs`, `graph_memberships`, `documents`, `password_reset_tokens`
- [ ] Indexes are created (check with `\di` in psql)
- [ ] Foreign key constraints are in place
- [ ] UUID extension is enabled
- [ ] Application can connect and query the database

```sql
-- Quick verification queries
SELECT tablename FROM pg_tables WHERE schemaname = 'public';
SELECT indexname FROM pg_indexes WHERE schemaname = 'public';
SELECT conname FROM pg_constraint WHERE contype = 'f';
```

## Troubleshooting

### "password authentication failed"
- Check username/password in DATABASE_URL
- Verify user has CREATE privileges
- Check pg_hba.conf for connection rules

### "database does not exist"
- Create database first: `createdb orgmind`
- Or use psql: `CREATE DATABASE orgmind;`

### "permission denied for schema public"
- Grant privileges: `GRANT ALL ON SCHEMA public TO orgmind;`

### "relation already exists"
- Database already has tables
- Use incremental migrations instead of consolidated
- Or drop existing tables first (DANGER!)

## Support

For issues or questions:
- Check [migrations/README.md](./README.md)
- Review [MIGRATION_GUIDE.md](../MIGRATION_GUIDE.md)
- Check application logs for detailed error messages
