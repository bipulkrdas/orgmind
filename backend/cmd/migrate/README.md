# Data Migration Script

This directory contains the data migration script for migrating existing users and documents to the multi-tenant graph management system.

## Overview

The migration script creates default graphs for users who have existing documents but no graphs. This is necessary when upgrading from the single-graph-per-user system to the multi-tenant graph management system.

## What the Migration Does

For each user with documents but no graphs, the script:

1. Creates a new graph in the database with the name "My Knowledge Graph"
2. Stores the graph metadata including `creator_id`, `zep_graph_id`, name, and description
3. Creates an owner membership record associating the user with the graph
4. Updates all existing documents to reference the new graph
5. Updates the document count for the graph

## Prerequisites

- Database migrations must be run first (002_add_graphs_table.up.sql)
- The `graph_id` column must exist in the `documents` table
- Environment variables must be configured (DATABASE_URL, etc.)

## Usage

### Dry Run (Recommended First)

Before running the actual migration, perform a dry run to see what would be migrated:

```bash
cd backend
go run cmd/migrate/main.go --migrate-existing-documents --dry-run
```

This will show:
- How many users need migration
- Each user's email and ID
- How many documents each user has
- What would be created (without making any changes)

### Run Migration

Once you've reviewed the dry run output, run the actual migration:

```bash
cd backend
go run cmd/migrate/main.go --migrate-existing-documents
```

The script will:
- Process each user sequentially
- Use database transactions to ensure data consistency
- Show progress for each user
- Display a summary at the end

### Example Output

```
Connected to database successfully

=== STARTING MIGRATION ===

Found 3 user(s) to migrate

[1/3] Processing user: alice@example.com (ID: 123e4567-e89b-12d3-a456-426614174000)
  Creating graph with ID: 987fcdeb-51a2-43f7-8765-123456789abc
  ✓ Graph created
  ✓ Owner membership created
  ✓ Updated 5 document(s)
  ✓ Document count updated to 5
✓ Successfully migrated user: alice@example.com

[2/3] Processing user: bob@example.com (ID: 234e5678-e89b-12d3-a456-426614174001)
  Creating graph with ID: 876fedcb-51a2-43f7-8765-123456789def
  ✓ Graph created
  ✓ Owner membership created
  ✓ Updated 3 document(s)
  ✓ Document count updated to 3
✓ Successfully migrated user: bob@example.com

[3/3] Processing user: carol@example.com (ID: 345e6789-e89b-12d3-a456-426614174002)
  Creating graph with ID: 765fedcb-51a2-43f7-8765-123456789ghi
  ✓ Graph created
  ✓ Owner membership created
  ✓ Updated 2 document(s)
  ✓ Document count updated to 2
✓ Successfully migrated user: carol@example.com

Migration Summary:
  Success: 3
  Failed:  0
  Total:   3

=== MIGRATION COMPLETED SUCCESSFULLY ===
```

## Error Handling

The migration script:
- Uses database transactions to ensure atomicity
- Rolls back changes if any step fails for a user
- Continues processing other users if one fails
- Displays a summary showing successes and failures

If a user migration fails:
- The transaction is rolled back (no partial state)
- An error message is logged
- The script continues with the next user

## Post-Migration

After successful migration:

1. Verify the migration:
   ```sql
   -- Check that all documents have a graph_id
   SELECT COUNT(*) FROM documents WHERE graph_id IS NULL;
   -- Should return 0
   
   -- Check that all users with documents have graphs
   SELECT u.email, COUNT(g.id) as graph_count
   FROM users u
   LEFT JOIN graph_memberships gm ON u.id = gm.user_id
   LEFT JOIN graphs g ON gm.graph_id = g.id
   GROUP BY u.id, u.email;
   ```

2. Make `graph_id` NOT NULL (if desired):
   ```sql
   ALTER TABLE documents ALTER COLUMN graph_id SET NOT NULL;
   ```

3. Test the application to ensure:
   - Users can see their migrated graphs
   - Documents appear in the correct graphs
   - Document counts are accurate

## Rollback

If you need to rollback the migration:

1. Run the down migration to remove the graphs table:
   ```bash
   migrate -path ./migrations -database $DATABASE_URL down 1
   ```

2. This will:
   - Drop the `graph_memberships` table
   - Drop the `graphs` table
   - Remove the `graph_id` column from `documents`

Note: This will delete all graph data, so only do this if necessary.

## Troubleshooting

### "No users found that need migration"

This means either:
- All users already have graphs
- No users have documents yet
- The migration has already been run

### "Failed to connect to database"

Check that:
- DATABASE_URL environment variable is set correctly
- The database is running and accessible
- Network connectivity is working

### Transaction errors

If you see transaction errors:
- Check database logs for more details
- Ensure no other processes are modifying the same data
- Verify database has sufficient resources

## Safety Features

The migration script includes several safety features:

1. **Dry Run Mode**: Preview changes before applying them
2. **Transactions**: Each user migration is atomic (all or nothing)
3. **Error Isolation**: Failure for one user doesn't affect others
4. **Detailed Logging**: Clear progress and error messages
5. **Summary Report**: Shows exactly what was migrated

## Development

To modify the migration script:

1. Edit `backend/cmd/migrate/main.go`
2. Test with `--dry-run` flag first
3. Test on a development database before production
4. Consider adding additional validation or checks

## Notes

- The script is idempotent for users without graphs
- Users who already have graphs are skipped automatically
- The default graph name is "My Knowledge Graph"
- The default description is "Default graph created during migration"
- Graph IDs are generated as UUIDs
- Zep graph IDs follow the format `graph-{uuid}`
