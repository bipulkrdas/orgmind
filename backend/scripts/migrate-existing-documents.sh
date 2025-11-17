#!/bin/bash

# Script to migrate existing documents to default graphs
# This should be run after the database migrations have been applied

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== OrgMind Data Migration Script ===${NC}"
echo ""
echo "This script will create default graphs for users with existing documents."
echo ""

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}ERROR: DATABASE_URL environment variable is not set${NC}"
    echo "Please set DATABASE_URL and try again."
    exit 1
fi

# Parse command line arguments
DRY_RUN=false
SKIP_CONFIRMATION=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --yes|-y)
            SKIP_CONFIRMATION=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --dry-run        Show what would be migrated without making changes"
            echo "  --yes, -y        Skip confirmation prompt"
            echo "  --help, -h       Show this help message"
            echo ""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run dry run first if not already in dry-run mode
if [ "$DRY_RUN" = false ] && [ "$SKIP_CONFIRMATION" = false ]; then
    echo -e "${YELLOW}Running dry run first to preview changes...${NC}"
    echo ""
    go run cmd/migrate/main.go --migrate-existing-documents --dry-run
    echo ""
    
    # Ask for confirmation
    echo -e "${YELLOW}Do you want to proceed with the migration? (yes/no)${NC}"
    read -r response
    
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${RED}Migration cancelled${NC}"
        exit 0
    fi
    echo ""
fi

# Run the migration
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}Running in DRY RUN mode (no changes will be made)${NC}"
    echo ""
    go run cmd/migrate/main.go --migrate-existing-documents --dry-run
else
    echo -e "${GREEN}Running migration...${NC}"
    echo ""
    go run cmd/migrate/main.go --migrate-existing-documents
fi

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Migration completed successfully${NC}"
    
    if [ "$DRY_RUN" = false ]; then
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Verify the migration by checking the database"
        echo "2. Test the application to ensure graphs are working correctly"
        echo "3. Consider making graph_id NOT NULL in documents table:"
        echo "   ALTER TABLE documents ALTER COLUMN graph_id SET NOT NULL;"
    fi
else
    echo ""
    echo -e "${RED}✗ Migration failed${NC}"
    echo "Please check the error messages above and try again."
    exit 1
fi
