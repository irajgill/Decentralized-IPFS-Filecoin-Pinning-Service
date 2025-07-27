#!/bin/bash

set -e

# Configuration
DB_URL=${DATABASE_DSN:-"postgres://user:password@localhost:5432/pinning_db?sslmode=disable"}
MIGRATIONS_DIR="./migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 {up|down|force|version}"
    echo "  up      - Apply all up migrations"
    echo "  down    - Apply one down migration"
    echo "  force   - Force set migration version"
    echo "  version - Show current migration version"
    exit 1
}

check_migrate() {
    if ! command -v migrate &> /dev/null; then
        echo -e "${RED}Error: migrate tool not found${NC}"
        echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

run_migration() {
    local action=$1
    local version=${2:-""}
    
    echo -e "${YELLOW}Running migration: $action${NC}"
    
    case $action in
        "up")
            migrate -database "$DB_URL" -path "$MIGRATIONS_DIR" up
            ;;
        "down")
            migrate -database "$DB_URL" -path "$MIGRATIONS_DIR" down 1
            ;;
        "force")
            if [ -z "$version" ]; then
                echo -e "${RED}Error: Version required for force command${NC}"
                exit 1
            fi
            migrate -database "$DB_URL" -path "$MIGRATIONS_DIR" force "$version"
            ;;
        "version")
            migrate -database "$DB_URL" -path "$MIGRATIONS_DIR" version
            ;;
        *)
            usage
            ;;
    esac
}

main() {
    if [ $# -eq 0 ]; then
        usage
    fi
    
    check_migrate
    
    echo -e "${GREEN}Database URL: $DB_URL${NC}"
    echo -e "${GREEN}Migrations directory: $MIGRATIONS_DIR${NC}"
    
    run_migration "$@"
    
    echo -e "${GREEN}Migration completed successfully!${NC}"
}

main "$@"
