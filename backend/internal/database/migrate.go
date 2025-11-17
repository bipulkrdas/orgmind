package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db             *DB
	migrationsPath string
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *DB, migrationsPath string) *MigrationRunner {
	return &MigrationRunner{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist
func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := mr.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a map of applied migration versions
func (mr *MigrationRunner) getAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := mr.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// loadMigrations loads all migration files from the migrations directory
func (mr *MigrationRunner) loadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(mr.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Only process .up.sql files
		if !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		// Parse migration version and name from filename
		// Expected format: 001_create_initial_schema.up.sql
		filename := filepath.Base(path)
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			return fmt.Errorf("failed to parse migration version from %s: %w", filename, err)
		}

		name := strings.TrimSuffix(parts[1], ".up.sql")

		// Read up migration
		upSQL, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read up migration %s: %w", path, err)
		}

		// Read down migration
		downPath := strings.Replace(path, ".up.sql", ".down.sql", 1)
		downSQL, err := os.ReadFile(downPath)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", downPath, err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration applies a single migration within a transaction
func (mr *MigrationRunner) applyMigration(migration Migration) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	// Record migration in schema_migrations table
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)",
		migration.Version,
		migration.Name,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	return nil
}

// Up runs all pending migrations
func (mr *MigrationRunner) Up() error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Load all migrations
	migrations, err := mr.loadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Name)
		if err := mr.applyMigration(migration); err != nil {
			return err
		}
		appliedCount++
	}

	if appliedCount == 0 {
		fmt.Println("No pending migrations")
	} else {
		fmt.Printf("Successfully applied %d migration(s)\n", appliedCount)
	}

	return nil
}

// Down rolls back the last migration
func (mr *MigrationRunner) Down() error {
	// Get the last applied migration
	var version int
	var name string
	err := mr.db.QueryRow("SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &name)
	if err == sql.ErrNoRows {
		fmt.Println("No migrations to roll back")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// Load all migrations to find the down SQL
	migrations, err := mr.loadMigrations()
	if err != nil {
		return err
	}

	var targetMigration *Migration
	for _, m := range migrations {
		if m.Version == version {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found in migration files", version)
	}

	fmt.Printf("Rolling back migration %d: %s\n", version, name)

	// Execute rollback in transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute down migration
	if _, err := tx.Exec(targetMigration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute down migration %d: %w", version, err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", version, err)
	}

	fmt.Printf("Successfully rolled back migration %d\n", version)
	return nil
}
