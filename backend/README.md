# OrgMind Backend

Go backend service for the OrgMind document processing platform.

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP request handlers
│   ├── middleware/      # HTTP middleware (auth, logging, etc.)
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── router/          # Route definitions
│   ├── service/         # Business logic
│   └── storage/         # Storage implementations (S3, etc.)
├── pkg/
│   └── utils/           # Shared utilities
├── migrations/          # Database migration files
├── .env.example         # Example environment variables
└── go.mod               # Go module definition
```

## Setup

1. Copy `.env.example` to `.env` and configure your environment variables:
   ```bash
   cp .env.example .env
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o bin/server ./cmd/server
   ```

4. Run the application:
   ```bash
   ./bin/server
   ```

## Dependencies

- **gin-gonic/gin**: HTTP web framework
- **jmoiron/sqlx**: SQL extensions for Go
- **Masterminds/squirrel**: SQL query builder
- **lib/pq**: PostgreSQL driver
- **getzep/zep-go/v3**: Zep Cloud SDK
- **aws-sdk-go-v2**: AWS SDK for S3 storage
- **golang-jwt/jwt/v5**: JWT token handling
- **joho/godotenv**: Environment variable loading

## Environment Variables

See `.env.example` for all required and optional environment variables.

### Core Required Variables
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing
- `ZEP_API_KEY`: Zep Cloud API key for knowledge graphs
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`: AWS credentials for S3 storage

### AI Chat Feature (Optional)
For AI-powered chat functionality, configure Google Gemini:
- `GEMINI_API_KEY`: Google Gemini API key
- `GEMINI_PROJECT_ID`: Google Cloud project ID
- `GEMINI_LOCATION`: Google Cloud region (default: `us-central1`)

See [GEMINI_SETUP.md](GEMINI_SETUP.md) for detailed setup instructions.

## Database Migrations

### Schema Migrations

Apply database schema changes:
```bash
# Using golang-migrate
migrate -path ./migrations -database $DATABASE_URL up

# Or using the built-in migration tool
go run cmd/server/main.go --migrate
```

### Data Migration (Multi-Tenant Graphs)

If upgrading from a previous version, migrate existing documents to the new graph system:

```bash
# Quick start
./scripts/migrate-existing-documents.sh

# Or see detailed guide
cat MIGRATION_QUICKSTART.md
```

For complete migration documentation, see [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md).

## API Documentation

### Chat API

For detailed documentation on the AI-powered chat endpoints, see [CHAT_API.md](CHAT_API.md).

The Chat API provides:
- Thread management for organizing conversations
- Real-time AI responses via Server-Sent Events (SSE)
- Integration with Google Gemini File Search
- Message history with pagination
- Rate limiting and access control

Quick reference:
- `POST /api/graphs/:graphId/chat/threads` - Create chat thread
- `GET /api/graphs/:graphId/chat/threads/:threadId/messages` - Get messages
- `POST /api/graphs/:graphId/chat/threads/:threadId/messages` - Send message
- `GET /api/graphs/:graphId/chat/stream` - Stream AI response (SSE)

## Development

Run in development mode:
```bash
go run cmd/server/main.go
```

Run tests:
```bash
go test ./...
```
