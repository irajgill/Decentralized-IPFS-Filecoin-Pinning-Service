#!/bin/bash

set -e

# Configuration
DB_URL=${DATABASE_DSN:-"postgres://user:password@localhost:5432/pinning_db?sslmode=disable"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Seeding database with test data...${NC}"

# Generate API key
API_KEY=$(openssl rand -hex 32)

# SQL to insert test user
SQL="
INSERT INTO users (email, api_key, balance) 
VALUES ('test@example.com', '$API_KEY', 1.0)
ON CONFLICT (email) DO UPDATE SET 
  api_key = EXCLUDED.api_key,
  balance = EXCLUDED.balance;
"

# Execute SQL
psql "$DB_URL" -c "$SQL"

echo -e "${GREEN}Test user created successfully!${NC}"
echo -e "${GREEN}Email: test@example.com${NC}"
echo -e "${GREEN}API Key: $API_KEY${NC}"
echo -e "${GREEN}Balance: 1.0 FIL${NC}"
echo ""
echo -e "${YELLOW}You can now test the API with:${NC}"
echo "curl -H 'X-API-Key: $API_KEY' http://localhost:8080/pricing"
