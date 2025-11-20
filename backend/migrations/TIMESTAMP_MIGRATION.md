# Timestamp Migration to TIMESTAMPTZ

## Overview

Migration `006_convert_timestamps_to_timestamptz` converts all `TIMESTAMP` columns to `TIMESTAMP WITH TIME ZONE` (timestamptz) to ensure proper UTC storage and eliminate timezone ambiguity.

## Why This Change?

### Problem with TIMESTAMP (without timezone)
- Stores date/time values **without** timezone information
- When Go's `time.Now()` sends a timestamp, PostgreSQL strips the timezone
- Creates ambiguity: is "2025-11-18 18:00:00" in EST, UTC, or another timezone?
- Causes issues when servers are in different timezones than users

### Solution with TIMESTAMP WITH TIME ZONE
- Stores absolute instant in time (internally as UTC)
- PostgreSQL automatically converts incoming timestamps to UTC
- When reading, can convert to any timezone
- Eliminates all timezone ambiguity

## What This Migration Does

1. **Converts existing data**: Uses `AT TIME ZONE 'UTC'` to interpret existing timestamps as UTC
2. **Changes column types**: Alters all timestamp columns to `TIMESTAMP WITH TIME ZONE`
3. **Preserves data**: No data loss, just adds timezone awareness

## Tables Affected

- `users` (created_at, updated_at)
- `graphs` (created_at, updated_at)
- `graph_memberships` (created_at)
- `documents` (created_at, updated_at)
- `password_reset_tokens` (expires_at, created_at)
- `chat_threads` (created_at, updated_at)
- `chat_messages` (created_at)
- `gemini_filesearch_stores` (already using timestamptz)

## Running the Migration

### For New Databases
The migration will run automatically as part of the migration sequence.

### For Existing Databases
```bash
cd backend
go run cmd/migrate/main.go up
```

This will:
1. Convert all timestamp columns to timestamptz
2. Interpret existing timestamps as UTC
3. Update column defaults to use `CURRENT_TIMESTAMP`

## Impact

### Before Migration
```sql
-- Stored as: 2025-11-18 18:00:00 (ambiguous timezone)
-- Could be EST, UTC, or any timezone
```

### After Migration
```sql
-- Stored as: 2025-11-18 23:00:00+00 (explicit UTC)
-- Always unambiguous
```

## Code Changes

### Go Backend
No changes needed! Go's `time.Time` already handles timezones correctly. The PostgreSQL driver will automatically:
- Convert `time.Now()` to UTC when inserting
- Return timestamps with proper timezone when querying

### API Responses
Updated to use `time.RFC3339` format with explicit UTC conversion:
```go
CreatedAt: thread.CreatedAt.UTC().Format(time.RFC3339)
// Produces: "2025-11-18T23:00:00Z"
```

## Rollback

If needed, you can rollback using:
```bash
go run cmd/migrate/main.go down
```

**Warning**: Rolling back will lose timezone information and revert to ambiguous timestamps.

## Best Practices Going Forward

1. **Always use `TIMESTAMP WITH TIME ZONE`** for new timestamp columns
2. **Use `CURRENT_TIMESTAMP`** as default (not `NOW()`)
3. **Format as RFC3339** when sending to frontend: `time.RFC3339`
4. **Convert to UTC** before formatting: `.UTC().Format(time.RFC3339)`

## Verification

After migration, verify timestamps are correct:

```sql
-- Check a sample of timestamps
SELECT id, created_at, updated_at 
FROM chat_threads 
ORDER BY created_at DESC 
LIMIT 5;

-- Should show timestamps like: 2025-11-18 23:27:13+00
```

## References

- PostgreSQL Docs: https://www.postgresql.org/docs/current/datatype-datetime.html
- Go time package: https://pkg.go.dev/time
- RFC3339 format: https://datatracker.ietf.org/doc/html/rfc3339
