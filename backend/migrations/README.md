# Database Migrations

This directory contains database migration files for OrgMind.

## Directory Structure

```
migrations/
├── README.md                              # This file
├── consolidated/                          # For fresh deployments
│   ├── README.md
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── 001_create_initial_schema.up.sql      # Incremental migrations
├── 001_create_initial_schema.down.sql
├── 002_add_graphs_table.up.sql
└── 002_add_graphs_table.down.sql
```

## Which Migrations Should I Use?

### Fresh Deployment (New Database)

If you're setting up a **new production environment** or **new database**, use the **consolidated schema**:

```bash
migrate -path ./migrations/consolidated -database $DATABASE_URL up
```

See [`consolidated/README.md`](./consolidated/README.md) for details.

### Existing Database (Incremental Updates)

If you have an **existing database** that needs updates, use the **incremental migrations** in this directory:

```bash
migrate -path ./migrations -database $DATABASE_URL up
```

## Migration Files

### Incremental Migrations (For Existing Databases)

1. **001_create_initial_schema** - Initial tables (users, documents, password_reset_tokens)
2. **002_add_graphs_table** - Multi-tenant graph management (graphs, graph_memberships, updates to documents)

### Consolidated Schema (For Fresh Deployments)

- **001_initial_schema** - Complete schema with all tables in one migration

## Running Migrations

### Using golang-migrate CLI

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

Run migrations:
```bash
# Up (apply migrations)
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" up

# Down (rollback migrations)
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" down

# Check version
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" version
```

### Using Docker

```bash
docker run --rm -v $(pwd)/migrations:/migrations --network host \
  migrate/migrate:v4.15.2 \
  -path=/migrations \
  -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" \
  up
```

### Using the Application

The Go application can run migrations automatically on startup. See `backend/internal/database/migrate.go`.

## Creating New Migrations

When adding new features that require database changes:

1. Create a new migration pair with the next number:
   ```bash
   touch migrations/003_your_feature_name.up.sql
   touch migrations/003_your_feature_name.down.sql
   ```

2. Write the up migration (schema changes)
3. Write the down migration (rollback changes)
4. Test both up and down migrations
5. Update the consolidated schema if this is a breaking change

## Best Practices

- **Always test migrations** on a development database first
- **Write reversible migrations** - every `.up.sql` should have a corresponding `.down.sql`
- **Use transactions** where possible (most DDL statements in PostgreSQL are transactional)
- **Avoid data loss** - be careful with DROP, TRUNCATE, and DELETE operations
- **Document breaking changes** in commit messages and release notes
- **Keep migrations small** - one logical change per migration
- **Never modify existing migrations** that have been deployed to production

## Troubleshooting

### Migration fails with "Dirty database version"

This happens when a migration fails partway through. Fix it:

```bash
# Force set version to the last successful migration
migrate -path ./migrations -database $DATABASE_URL force <version>

# Then try again
migrate -path ./migrations -database $DATABASE_URL up
```

### "relation already exists" error

You're trying to run migrations on a database that already has some tables. Either:
- Use the incremental migrations starting from the next version
- Drop all tables and start fresh with the consolidated schema

### Connection refused

Check your database connection string and ensure PostgreSQL is running:
```bash
psql $DATABASE_URL -c "SELECT version();"
```

## Schema Documentation

For detailed schema documentation, see:
- [Consolidated Schema README](./consolidated/README.md)
- [Migration Guide](../MIGRATION_GUIDE.md)
