# Consolidated Database Schema

This directory contains a consolidated initial schema for **fresh deployments** of OrgMind.

## When to Use This

Use the consolidated schema when:
- Setting up a **new production environment** from scratch
- Creating a **new development database**
- Running **integration tests** that need a clean database
- Deploying to a **new region or environment**

## Files

- `001_initial_schema.up.sql` - Complete database schema with all tables
- `001_initial_schema.down.sql` - Teardown script to drop all tables

## What's Included

The consolidated schema includes:

1. **Users Table** - User accounts with email/password and OAuth support
2. **Graphs Table** - Knowledge graphs with Zep integration
3. **Graph Memberships Table** - Many-to-many relationship for graph access control
4. **Documents Table** - Document storage with graph associations
5. **Password Reset Tokens Table** - Secure password reset functionality

## Usage

### Using golang-migrate CLI

```bash
# Apply the consolidated schema
migrate -path ./migrations/consolidated -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" up

# Rollback (drops all tables)
migrate -path ./migrations/consolidated -database "postgresql://user:pass@localhost:5432/orgmind?sslmode=disable" down
```

### Using Docker Compose

Update your `docker-compose.yml` to mount this directory:

```yaml
volumes:
  - ./backend/migrations/consolidated:/migrations
```

### Using the Application

The Go application can run migrations automatically on startup. Update the migration path in your configuration to point to this directory for fresh deployments.

## Incremental Migrations

If you have an **existing database** that was created with the original migrations, continue using the incremental migrations in the parent directory (`../001_*.sql`, `../002_*.sql`, etc.).

Do **NOT** run this consolidated schema on an existing database - it will fail due to existing tables.

## Schema Overview

```
users (1) ──────┬─────> (N) documents
                │
                └─────> (N) graph_memberships (N) <───── (1) graphs
                                                              │
                                                              └─────> (N) documents
```

### Key Relationships

- Users can create multiple graphs (creator_id)
- Users can be members of multiple graphs (graph_memberships)
- Documents belong to a user and optionally to a graph
- Graph memberships support roles: owner, editor, viewer, member

## Notes

- The `graph_id` column in the `documents` table is nullable to support documents created before graph assignment
- All foreign keys use `ON DELETE CASCADE` for automatic cleanup
- Indexes are optimized for common query patterns (user lookups, graph memberships, document listings)
- UUID primary keys are generated using PostgreSQL's `pgcrypto` extension
