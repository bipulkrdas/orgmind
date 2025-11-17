# Migration Guide: Multi-Tenant Graph Management

This guide explains how to migrate from the single-graph-per-user system to the multi-tenant graph management system.

## Overview

The multi-tenant graph management feature introduces:
- Multiple graphs per user
- Graph memberships (many-to-many relationship between users and graphs)
- Graph-level document organization
- Creator-based permissions

## Migration Steps

### Step 1: Backup Your Database

**IMPORTANT**: Always backup your database before running migrations.

```bash
# PostgreSQL backup example
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Step 2: Run Database Schema Migrations

Apply the database schema changes:

```bash
cd backend

# Using golang-migrate
migrate -path ./migrations -database $DATABASE_URL up

# Or using the built-in migration tool
go run cmd/server/main.go --migrate
```

This will:
- Create the `graphs` table
- Create the `graph_memberships` table
- Add `graph_id` column to `documents` table (nullable initially)

### Step 3: Run Data Migration

The data migration creates default graphs for existing users with documents.

#### Option A: Using the Shell Script (Recommended)

```bash
cd backend

# Dry run first (preview changes)
./scripts/migrate-existing-documents.sh --dry-run

# Run the actual migration
./scripts/migrate-existing-documents.sh

# Or skip confirmation prompt
./scripts/migrate-existing-documents.sh --yes
```

#### Option B: Using Go Directly

```bash
cd backend

# Dry run
go run cmd/migrate/main.go --migrate-existing-documents --dry-run

# Actual migration
go run cmd/migrate/main.go --migrate-existing-documents
```

### Step 4: Verify the Migration

After the migration completes, verify the results:

```sql
-- Check that all documents have a graph_id
SELECT COUNT(*) as orphaned_documents 
FROM documents 
WHERE graph_id IS NULL;
-- Should return 0

-- Check graph counts
SELECT 
    u.email,
    COUNT(DISTINCT g.id) as graph_count,
    COUNT(d.id) as document_count
FROM users u
LEFT JOIN graph_memberships gm ON u.id = gm.user_id
LEFT JOIN graphs g ON gm.graph_id = g.id
LEFT JOIN documents d ON d.graph_id = g.id
GROUP BY u.id, u.email
ORDER BY u.email;

-- Verify document counts match
SELECT 
    g.name,
    g.document_count as stored_count,
    COUNT(d.id) as actual_count
FROM graphs g
LEFT JOIN documents d ON d.graph_id = g.id
GROUP BY g.id, g.name, g.document_count
HAVING g.document_count != COUNT(d.id);
-- Should return no rows
```

### Step 5: Test the Application

1. Start the backend server:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. Start the frontend:
   ```bash
   cd frontend
   npm run dev
   ```

3. Test the following:
   - Users can see their migrated graphs
   - Documents appear in the correct graphs
   - Document counts are accurate
   - Users can create new graphs
   - Users can add documents to graphs

### Step 6: Make graph_id NOT NULL (Optional)

Once you've verified everything works, you can make the `graph_id` column required:

```sql
-- Make graph_id NOT NULL
ALTER TABLE documents ALTER COLUMN graph_id SET NOT NULL;
```

**Note**: Only do this after verifying all documents have been migrated successfully.

## What the Migration Does

For each user with documents but no graphs:

1. **Creates a Graph**
   - Name: "My Knowledge Graph"
   - Description: "Default graph created during migration"
   - Generates a unique UUID for the graph
   - Creates a Zep graph ID in the format `graph-{uuid}`

2. **Creates Owner Membership**
   - Associates the user with the graph as "owner"
   - Allows the user to manage the graph

3. **Updates Documents**
   - Sets `graph_id` for all user's documents without a graph
   - Updates the `updated_at` timestamp

4. **Updates Document Count**
   - Sets the graph's `document_count` to match actual documents

## Migration Safety Features

The migration script includes several safety features:

1. **Dry Run Mode**: Preview changes before applying
2. **Transactions**: Each user migration is atomic (all or nothing)
3. **Error Isolation**: Failure for one user doesn't affect others
4. **Detailed Logging**: Clear progress and error messages
5. **Summary Report**: Shows exactly what was migrated

## Troubleshooting

### No Users Found

If you see "No users found that need migration":
- All users already have graphs, OR
- No users have documents yet, OR
- The migration has already been run successfully

This is normal and not an error.

### Transaction Errors

If you encounter transaction errors:
1. Check database logs for details
2. Ensure no other processes are modifying the data
3. Verify database has sufficient resources
4. Try running the migration again (it's safe to retry)

### Partial Migration

If the migration fails partway through:
- Successfully migrated users will have graphs
- Failed users will not have graphs (transaction rollback)
- You can safely re-run the migration
- Only users without graphs will be processed

### Connection Issues

If you can't connect to the database:
1. Verify `DATABASE_URL` is set correctly
2. Check the database is running
3. Verify network connectivity
4. Check firewall rules

## Rollback

If you need to rollback the migration:

### Rollback Data Migration Only

```sql
-- Delete all graphs created during migration
DELETE FROM graphs 
WHERE description = 'Default graph created during migration';

-- This will cascade delete:
-- - Graph memberships
-- - Document associations (sets graph_id to NULL)
```

### Rollback Schema Migration

```bash
cd backend

# Rollback one migration
migrate -path ./migrations -database $DATABASE_URL down 1
```

This will:
- Drop the `graph_memberships` table
- Drop the `graphs` table  
- Remove the `graph_id` column from `documents`

**Warning**: This will delete all graph data, including user-created graphs.

## Post-Migration Considerations

### User Experience

After migration, users will:
- See a "My Knowledge Graph" in their graphs list
- Find all their existing documents in this graph
- Be able to create additional graphs
- Be able to organize documents across multiple graphs

### Performance

The migration:
- Uses transactions for data consistency
- Processes users sequentially
- Should complete quickly for most databases
- May take longer for databases with many users

### Monitoring

After migration, monitor:
- Application logs for any graph-related errors
- Database performance (new indexes are created)
- User feedback on the new graph management UI

## Support

If you encounter issues:

1. Check the migration logs for error messages
2. Verify database state using the SQL queries above
3. Review the troubleshooting section
4. Check application logs for runtime errors

## Additional Resources

- [Migration Script README](cmd/migrate/README.md) - Detailed script documentation
- [Database Migrations](migrations/) - SQL migration files
- [Design Document](../.kiro/specs/multi-tenant-graph-management/design.md) - Architecture details
- [Requirements Document](../.kiro/specs/multi-tenant-graph-management/requirements.md) - Feature requirements
