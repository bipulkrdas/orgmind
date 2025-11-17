# Migration Quick Start

Quick reference for migrating to multi-tenant graph management.

## Prerequisites

- ✅ Database backup completed
- ✅ Schema migrations applied (002_add_graphs_table.up.sql)
- ✅ Environment variables configured (DATABASE_URL, etc.)

## Quick Migration

```bash
cd backend

# 1. Preview what will be migrated (recommended)
./scripts/migrate-existing-documents.sh --dry-run

# 2. Run the migration
./scripts/migrate-existing-documents.sh

# 3. Verify (in psql or your SQL client)
# Check all documents have graphs:
SELECT COUNT(*) FROM documents WHERE graph_id IS NULL;
# Should return 0
```

## Alternative: Direct Go Command

```bash
cd backend

# Dry run
go run cmd/migrate/main.go --migrate-existing-documents --dry-run

# Actual migration
go run cmd/migrate/main.go --migrate-existing-documents
```

## What Gets Created

For each user with documents:
- ✅ One graph named "My Knowledge Graph"
- ✅ Owner membership for the user
- ✅ All documents linked to the graph
- ✅ Document count updated

## Verification Queries

```sql
-- Check migration status
SELECT 
    u.email,
    COUNT(DISTINCT g.id) as graphs,
    COUNT(d.id) as documents
FROM users u
LEFT JOIN graph_memberships gm ON u.id = gm.user_id
LEFT JOIN graphs g ON gm.graph_id = g.id
LEFT JOIN documents d ON d.graph_id = g.id
GROUP BY u.id, u.email;

-- Verify document counts
SELECT 
    g.name,
    g.document_count as stored,
    COUNT(d.id) as actual
FROM graphs g
LEFT JOIN documents d ON d.graph_id = g.id
GROUP BY g.id, g.name, g.document_count;
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "No users found" | Normal - no migration needed |
| Connection error | Check DATABASE_URL |
| Transaction error | Check database logs, retry |
| Partial migration | Safe to re-run |

## Rollback

```sql
-- Remove migrated graphs only
DELETE FROM graphs 
WHERE description = 'Default graph created during migration';
```

## Next Steps

1. Test the application
2. Verify users can see their graphs
3. Check documents appear correctly
4. Optional: Make graph_id NOT NULL
   ```sql
   ALTER TABLE documents ALTER COLUMN graph_id SET NOT NULL;
   ```

## Need Help?

See [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) for detailed documentation.
