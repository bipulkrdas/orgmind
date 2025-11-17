package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB wraps the sqlx database connection
type DB struct {
	*sqlx.DB
}

// Config holds database connection configuration
type Config struct {
	DatabaseURL    string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
}

// Connect establishes a connection to the database with retry logic
func Connect(cfg Config) (*DB, error) {
	var db *sqlx.DB
	var err error

	// Set defaults
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.RetryAttempts == 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}

	// Retry connection with exponential backoff
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		db, err = sqlx.Connect("postgres", cfg.DatabaseURL)
		if err == nil {
			break
		}

		if attempt < cfg.RetryAttempts {
			waitTime := cfg.RetryDelay * time.Duration(attempt)
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...\n",
				attempt, cfg.RetryAttempts, err, waitTime)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.RetryAttempts, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database")

	return &DB{DB: db}, nil
}

// HealthCheck verifies the database connection is alive
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}
