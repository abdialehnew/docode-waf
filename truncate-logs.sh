#!/bin/bash

# ============================================
# Truncate Attack Logs Script
# Clears all attack/traffic logs from database
# ============================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}     Attack Logs Truncation Script         ${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Confirm before proceeding
echo -e "${YELLOW}WARNING: This will permanently delete all attack and traffic logs!${NC}"
echo -e "${YELLOW}This action cannot be undone.${NC}"
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo -e "${RED}Operation cancelled.${NC}"
    exit 0
fi

echo ""
echo -e "${BLUE}Truncating logs...${NC}"

# Execute SQL commands via docker compose
docker compose exec -T postgres psql -U waf_user -d docode_waf << 'EOF'
-- Truncate traffic logs (includes attack logs)
TRUNCATE TABLE traffic_logs;

-- Vacuum to reclaim space
VACUUM FULL traffic_logs;

-- Show result
SELECT 'Traffic logs truncated successfully!' as status;
SELECT COUNT(*) as remaining_logs FROM traffic_logs;
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Attack logs truncated successfully!${NC}"
    echo -e "${GREEN}✓ Database space reclaimed.${NC}"
else
    echo ""
    echo -e "${RED}✗ Error truncating logs.${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}Done!${NC}"
