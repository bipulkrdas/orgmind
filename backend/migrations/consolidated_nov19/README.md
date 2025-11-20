# OrgMind Database Schema - November 2025 Release

This directory contains the **final consolidated schema** for OrgMind, ready for production deployment.

## Overview

This is a single-migration schema that includes all features and tables. Use this for:
- ✅ **New production deployments**
- ✅ **Fresh development environments**
- ✅ **Integration testing**
- ✅ **New regional deployments**

## What's Included

### Core Tables
1. **users** - User accounts with email/password and OAuth support
2. **graphs** - Knowledge graphs with Zep integration
3. **graph_memberships** - Multi-tenant access control
4. **documents** - Document storage with graph associations

### Feature Tables
5. **password_reset_tokens** - Secure password reset functionality
6. **chat_threads** - AI chat conversation threads
7. **chat_messages** - Chat message history
8. **gemini_filesearch_stores** - Gemini File Search integration

## Key Features

### Timezone-Aware Timestamps
All timestamp columns use `TIMESTAMP WITH TIME ZONE`:
- Automatically stores in UTC
- Eliminates timezone ambiguity
- Works correctly with distributed servers
- No manual timezone conversion needed

### Gemini Integration
- `graphs.gemini_store_id` - Links graphs to Gemini File Search stores
- `documents.gemini_file_id` - Tracks uploaded files in Gemini
- `gemini_filesearch_stores` - Persists store information

### Multi-Tenant Architecture
- Graph-based isolation
- Role-based access control (owner, editor, viewer, member)
- Cascade deletes for data consistency

## Usage

### Using golang-migrate

```bash
# Set your database URL
export DATABASE_URL="postgresql://user:pass@host:5432/orgmind?sslmode=require"

# Apply schema
migrate -path ./migrations/consolidated_nov19 -database "$DATABASE_URL" up

# Verify
migrate -path ./migrations/consolidated_nov19 -database "$DATABASE_URL" version
```

Expected output: `1` (one migration applied)

### Using Docker

```bash
docker run --rm \
  -v $(pwd)/migrations/consolidated_nov19:/migrations \
  --network host \
  migrate/migrate:v4.15.2 \
  -path=/migrations \
  -database "$DATABASE_URL" \
  up
```

### Using the Go Application

The application can run migrations automatically on startup:

```go
// In your main.go or database initialization
import "github.com/bipulkrdas/orgmind/backend/internal/database"

func main() {
    // Run migrations
    if err := database.RunMigrations("./migrations/consolidated_nov19"); err != nil {
        log.Fatal(err)
    }
}
```

## Schema Diagram

```
users (1) ──────┬─────> (N) documents
                │
                ├─────> (N) graph_memberships (N) <───── (1) graphs ──┬──> (N) documents
                │                                                       │
                └─────> (N) chat_threads (N) <──────────────────────────┘
                                │
                                └─────> (N) chat_messages
```

## Database Requirements

- **PostgreSQL**: 12.0 or higher
- **Extensions**: pgcrypto (for UUID generation)
- **Timezone**: Server should be configured with UTC or proper timezone handling
- **Encoding**: UTF-8

## Environment Variables

```bash
# Required
DATABASE_URL="postgresql://user:password@host:5432/orgmind?sslmode=require"

# Optional (for connection pooling)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

## Verification

After running migrations, verify the schema:

```sql
-- Check all tables exist
SELECT tablename FROM pg_tables 
WHERE schemaname = 'public' 
ORDER BY tablename;

-- Expected tables:
-- chat_messages
-- chat_threads
-- documents
-- gemini_filesearch_stores
-- graph_memberships
-- graphs
-- password_reset_tokens
-- users

-- Verify indexes
SELECT indexname FROM pg_indexes 
WHERE schemaname = 'public' 
ORDER BY indexname;

-- Check foreign keys
SELECT conname, conrelid::regclass AS table_name, 
       confrelid::regclass AS referenced_table
FROM pg_constraint 
WHERE contype = 'f' AND connamespace = 'public'::regnamespace;
```

## Rollback

To completely remove the schema:

```bash
migrate -path ./migrations/consolidated_nov19 -database "$DATABASE_URL" down
```

**⚠️ WARNING**: This will delete ALL data!

## Migration from Old Schema

If you have an existing database with the old schema (without timezone-aware timestamps), you'll need to:

1. **Backup your data** first!
2. Use the incremental migrations in `../migrations/` directory
3. Run migration `006_convert_timestamps_to_timestamptz.up.sql`

Do **NOT** run this consolidated schema on an existing database.

## Differences from Previous Versions

### November 2025 Release
- ✅ All timestamps use `TIMESTAMP WITH TIME ZONE`
- ✅ Gemini File Search integration columns
- ✅ Chat functionality (threads and messages)
- ✅ Document error tracking
- ✅ Optimized indexes for common queries

### Previous Versions
- ❌ Used `TIMESTAMP` (without timezone)
- ❌ Required manual timezone conversion
- ❌ Missing Gemini integration
- ❌ No chat functionality

## Support

For issues or questions:
- Check the main [migrations README](../README.md)
- Review [MIGRATION_GUIDE.md](../MIGRATION_GUIDE.md)
- See [TIMESTAMP_MIGRATION.md](../TIMESTAMP_MIGRATION.md) for timezone details

## License

Copyright © 2025 OrgMind. All rights reserved.
