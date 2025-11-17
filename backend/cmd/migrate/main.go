package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/config"
	"github.com/bipulkrdas/orgmind/backend/internal/database"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Parse command line flags
	migrateExisting := flag.Bool("migrate-existing-documents", false, "Migrate existing documents to default graphs")
	dryRun := flag.Bool("dry-run", false, "Show what would be migrated without making changes")
	flag.Parse()

	if !*migrateExisting {
		fmt.Println("Usage: go run cmd/migrate/main.go --migrate-existing-documents [--dry-run]")
		fmt.Println("\nThis script creates default graphs for users with existing documents.")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	dbCfg := database.Config{
		DatabaseURL:     cfg.DatabaseURL,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		RetryAttempts:   3,
		RetryDelay:      2 * time.Second,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Initialize repositories
	graphRepo := repository.NewGraphRepository(db.DB)
	docRepo := repository.NewDocumentRepository(db.DB)

	// Initialize Zep service
	zepSvc, err := service.NewZepService(cfg.ZepAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Zep service: %v", err)
	}

	// Run migration
	ctx := context.Background()
	if *dryRun {
		fmt.Println("\n=== DRY RUN MODE - No changes will be made ===")
		fmt.Println()
		if err := dryRunMigration(ctx, db.DB, graphRepo, docRepo); err != nil {
			log.Fatalf("Dry run failed: %v", err)
		}
	} else {
		fmt.Println("\n=== STARTING MIGRATION ===")
		fmt.Println()
		if err := migrateExistingDocuments(ctx, db.DB, graphRepo, docRepo, zepSvc); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("\n=== MIGRATION COMPLETED SUCCESSFULLY ===")
	}
}

// dryRunMigration shows what would be migrated without making changes
func dryRunMigration(ctx context.Context, db *sqlx.DB, graphRepo repository.GraphRepository, docRepo repository.DocumentRepository) error {
	users, err := findUsersWithoutGraphs(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to find users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found that need migration.")
		return nil
	}

	fmt.Printf("Found %d user(s) that need migration:\n\n", len(users))

	for i, user := range users {
		docCount, err := countUserDocuments(ctx, db, user.ID)
		if err != nil {
			log.Printf("Warning: Failed to count documents for user %s: %v", user.ID, err)
			continue
		}

		fmt.Printf("%d. User: %s (ID: %s)\n", i+1, user.Email, user.ID)
		fmt.Printf("   Documents: %d\n", docCount)
		fmt.Printf("   Would create: Default graph with %d document(s)\n\n", docCount)
	}

	return nil
}

// migrateExistingDocuments creates default graphs for users with documents
func migrateExistingDocuments(ctx context.Context, db *sqlx.DB, graphRepo repository.GraphRepository, docRepo repository.DocumentRepository, zepSvc service.ZepService) error {
	// Find all users with documents but no graphs
	users, err := findUsersWithoutGraphs(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to find users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found that need migration.")
		return nil
	}

	fmt.Printf("Found %d user(s) to migrate\n\n", len(users))

	successCount := 0
	failureCount := 0

	for i, user := range users {
		fmt.Printf("[%d/%d] Processing user: %s (ID: %s)\n", i+1, len(users), user.Email, user.ID)

		if err := migrateUser(ctx, db, graphRepo, docRepo, zepSvc, user); err != nil {
			log.Printf("ERROR: Failed to migrate user %s: %v\n", user.Email, err)
			failureCount++
			continue
		}

		fmt.Printf("✓ Successfully migrated user: %s\n\n", user.Email)
		successCount++
	}

	fmt.Printf("\nMigration Summary:\n")
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed:  %d\n", failureCount)
	fmt.Printf("  Total:   %d\n", len(users))

	if failureCount > 0 {
		return fmt.Errorf("migration completed with %d failure(s)", failureCount)
	}

	return nil
}

// migrateUser creates a default graph for a single user and migrates their documents
func migrateUser(ctx context.Context, db *sqlx.DB, graphRepo repository.GraphRepository, docRepo repository.DocumentRepository, zepSvc service.ZepService, user *models.User) error {
	// Start a transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate graph ID
	graphID := uuid.New().String()
	zepGraphID := fmt.Sprintf("graph-%s", graphID)

	// Note: Current Zep SDK doesn't have explicit CreateGraph method
	// The graph is implicitly created when we add data to it
	// For future compatibility, we store the zepGraphID
	fmt.Printf("  Creating graph with ID: %s\n", graphID)

	// Create graph record
	now := time.Now()
	graph := &models.Graph{
		ID:            graphID,
		CreatorID:     user.ID,
		ZepGraphID:    zepGraphID,
		Name:          "My Knowledge Graph",
		Description:   stringPtr("Default graph created during migration"),
		DocumentCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Insert graph using transaction
	query := `
		INSERT INTO graphs (
			id, creator_id, zep_graph_id, name, description, 
			document_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`
	_, err = tx.ExecContext(ctx, query,
		graph.ID, graph.CreatorID, graph.ZepGraphID, graph.Name, graph.Description,
		graph.DocumentCount, graph.CreatedAt, graph.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create graph: %w", err)
	}

	fmt.Printf("  ✓ Graph created\n")

	// Create owner membership
	membershipID := uuid.New().String()
	membership := &models.GraphMembership{
		ID:        membershipID,
		GraphID:   graphID,
		UserID:    user.ID,
		Role:      "owner",
		CreatedAt: now,
	}

	membershipQuery := `
		INSERT INTO graph_memberships (
			id, graph_id, user_id, role, created_at
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`
	_, err = tx.ExecContext(ctx, membershipQuery,
		membership.ID, membership.GraphID, membership.UserID, membership.Role, membership.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	fmt.Printf("  ✓ Owner membership created\n")

	// Update all user's documents to reference the new graph
	updateDocsQuery := `
		UPDATE documents
		SET graph_id = $1, updated_at = $2
		WHERE user_id = $3 AND graph_id IS NULL
	`
	result, err := tx.ExecContext(ctx, updateDocsQuery, graphID, now, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update documents: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	fmt.Printf("  ✓ Updated %d document(s)\n", rowsAffected)

	// Update document count for the graph
	if rowsAffected > 0 {
		updateCountQuery := `
			UPDATE graphs
			SET document_count = $1, updated_at = $2
			WHERE id = $3
		`
		_, err = tx.ExecContext(ctx, updateCountQuery, rowsAffected, now, graphID)
		if err != nil {
			return fmt.Errorf("failed to update document count: %w", err)
		}

		fmt.Printf("  ✓ Document count updated to %d\n", rowsAffected)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// findUsersWithoutGraphs finds all users who have documents but no graphs
func findUsersWithoutGraphs(ctx context.Context, db *sqlx.DB) ([]*models.User, error) {
	query := `
		SELECT DISTINCT 
			u.id, u.email, u.password_hash, u.first_name, u.last_name,
			u.oauth_provider, u.oauth_id, u.created_at, u.updated_at
		FROM users u
		INNER JOIN documents d ON u.id = d.user_id
		WHERE d.graph_id IS NULL
		ORDER BY u.created_at ASC
	`

	var users []*models.User
	err := db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	return users, nil
}

// countUserDocuments counts the number of documents for a user without a graph
func countUserDocuments(ctx context.Context, db *sqlx.DB, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM documents
		WHERE user_id = $1 AND graph_id IS NULL
	`

	var count int
	err := db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
