# Backend Development Guide

Quick reference for running and debugging the OrgMind backend locally.

## Environment Setup

### 1. Create `.env` File

Copy the example file and fill in your values:

```bash
cd backend
cp .env.example .env
```

Edit `backend/.env` with your configuration:

```env
# Server
SERVER_PORT=8080

# Database
DATABASE_URL=postgresql://orgmind:orgmind@localhost:5432/orgmind?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRATION_HOURS=24

# AWS S3 (or MinIO for local development)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=orgmind-documents

# Zep Cloud
ZEP_API_KEY=your-zep-api-key
ZEP_API_URL=https://api.getzep.com/api/v2
```

### 2. Start PostgreSQL

Using Docker Compose (recommended):

```bash
# From project root
docker-compose up -d postgres
```

Or install PostgreSQL locally and create the database:

```bash
createdb orgmind
```

## Running the Backend

### Option 1: Using Go Run (from project root)

```bash
go run backend/cmd/server/main.go
```

### Option 2: Using Go Run (from backend directory)

```bash
cd backend
go run cmd/server/main.go
```

### Option 3: Build and Run

```bash
# From project root
go build -o bin/server backend/cmd/server/main.go
./bin/server

# Or from backend directory
cd backend
go build -o ../bin/server cmd/server/main.go
../bin/server
```

### Option 4: Using Air (Hot Reload)

Install Air:
```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:
```bash
cd backend
air
```

## Debugging

### VSCode Debugging

1. Open the project in VSCode
2. Set breakpoints in your code
3. Press `F5` or go to Run > Start Debugging
4. Select "Launch Backend Server" configuration

The debugger will:
- Start from the project root directory
- Automatically load `backend/.env`
- Stop at your breakpoints
- Show variable values and call stack

### Kiro IDE Debugging

1. Open `backend/cmd/server/main.go`
2. Set breakpoints by clicking in the gutter
3. Click the debug icon or use the debug panel
4. The `.env` file will be automatically loaded

### Manual Debugging with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
cd backend
dlv debug cmd/server/main.go
```

## Environment File Loading

The application automatically searches for `.env` files in these locations (in order):

1. `.env` (current directory)
2. `backend/.env` (from project root)
3. `../.env` (from cmd/server directory)
4. `../../.env` (from internal/config directory)

This means the `.env` file will be found regardless of where you run the command from.

## Common Development Tasks

### Run Migrations

```bash
# From project root
go run backend/cmd/migrate/main.go

# Or using migrate CLI
migrate -path backend/migrations -database $DATABASE_URL up
```

### Run Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/service/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Format Code

```bash
# Format all Go files
go fmt ./...

# Or use gofmt directly
gofmt -w .
```

### Lint Code

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Generate Mocks (if using mockgen)

```bash
# Install mockgen
go install github.com/golang/mock/mockgen@latest

# Generate mocks
go generate ./...
```

## Troubleshooting

### "Failed to load configuration"

- Check that `backend/.env` exists
- Verify all required environment variables are set
- Check for syntax errors in `.env` file

### "Failed to connect to database"

- Ensure PostgreSQL is running: `docker-compose ps`
- Check DATABASE_URL is correct
- Verify database exists: `psql $DATABASE_URL -c "SELECT 1;"`

### "Failed to run migrations"

- Check migrations directory exists: `ls backend/migrations/`
- Verify database connection
- Check migration files are valid SQL

### "Failed to initialize storage service"

- Verify AWS credentials are correct
- Check S3 bucket exists and is accessible
- For local development, consider using MinIO instead of S3

### "Failed to initialize Zep service"

- Verify ZEP_API_KEY is set and valid
- Check ZEP_API_URL is correct
- Test API key: `curl -H "Authorization: Bearer $ZEP_API_KEY" $ZEP_API_URL/health`

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change the port in .env
SERVER_PORT=8081
```

## Project Structure

```
backend/
├── cmd/
│   ├── server/          # Main application entry point
│   └── migrate/         # Migration runner
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database connection and migrations
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── router/          # Route definitions
│   ├── service/         # Business logic
│   └── storage/         # File storage (S3, GCS, MinIO)
├── migrations/          # Database migrations
│   └── consolidated/    # Consolidated schema for fresh deployments
├── .env                 # Environment variables (not in git)
├── .env.example         # Example environment file
└── go.mod              # Go module definition
```

## Best Practices

1. **Always use `.env` for local development** - Never commit secrets to git
2. **Run migrations before starting** - Ensure database schema is up to date
3. **Use Docker Compose for dependencies** - PostgreSQL, MinIO, etc.
4. **Write tests for new features** - Maintain code quality
5. **Use structured logging** - Makes debugging easier
6. **Handle errors gracefully** - Return appropriate HTTP status codes
7. **Validate input** - Never trust user input
8. **Use interfaces** - Makes code testable and maintainable

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Zep Documentation](https://docs.getzep.com/)
- [AWS S3 SDK](https://aws.github.io/aws-sdk-go-v2/docs/)
