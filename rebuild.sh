#!/bin/bash

# DoCode WAF - Rebuild and Deploy Script
# This script stops, removes, rebuilds and redeploys the WAF and Frontend containers

set -e

echo "=========================================="
echo "  DoCode WAF - Rebuild & Deploy Script"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Stop containers
echo -e "${YELLOW}[1/5] Stopping containers...${NC}"
docker compose stop waf frontend nginx-proxy || true
echo -e "${GREEN}✓ Containers stopped${NC}"
echo ""

# Step 2: Remove containers
echo -e "${YELLOW}[2/5] Removing containers...${NC}"
docker compose rm -f waf frontend nginx-proxy || true
echo -e "${GREEN}✓ Containers removed${NC}"
echo ""

# Step 3: Remove old images
echo -e "${YELLOW}[3/5] Removing old images...${NC}"
docker rmi docode-waf-waf:latest 2>/dev/null || true
docker rmi docode-waf-frontend:latest 2>/dev/null || true
docker rmi docode-waf-nginx-proxy:latest 2>/dev/null || true
echo -e "${GREEN}✓ Old images removed${NC}"
echo ""

# Step 4: Build new images
echo -e "${YELLOW}[4/5] Building new images...${NC}"
docker compose build nginx-proxy waf frontend
echo -e "${GREEN}✓ Images built successfully${NC}"
echo ""

# Step 5: Deploy containers
echo -e "${YELLOW}[5/5] Deploying containers...${NC}"
docker compose up -d nginx-proxy waf frontend
echo -e "${GREEN}✓ Containers deployed${NC}"
echo ""

# Show status
echo "=========================================="
echo -e "${GREEN}  Deployment Complete!${NC}"
echo "=========================================="
echo ""
docker compose ps waf frontend nginx-proxy
echo ""
echo "Access the application at:"
echo "  - Frontend: http://localhost:3000"
echo "  - WAF API:  http://localhost:9090"
echo ""
