#!/bin/bash
#
# Cleanup integration test resources
#
# This script removes all Docker containers, volumes, and temporary files
# created during integration testing.
#
# Usage:
#   ./scripts/cleanup-test-resources.sh [--keep-volumes] [--force]
#
# Options:
#   --keep-volumes  Don't remove Docker volumes
#   --force         Don't ask for confirmation
#   --help          Show this help message
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
KEEP_VOLUMES=false
FORCE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --keep-volumes)
            KEEP_VOLUMES=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --help)
            grep "^#" "$0" | sed 's/^# //' | sed 's/^#//'
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Integration Test Resource Cleanup${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if docker-compose.yml exists
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Error: docker-compose.yml not found${NC}"
    echo -e "${YELLOW}Please run this script from the repository root${NC}"
    exit 1
fi

echo -e "${BLUE}This will clean up:${NC}"
echo "  • Docker containers (nexmonyx-*)"
echo "  • Docker networks (nexmonyx-test-network)"
if [ "$KEEP_VOLUMES" = false ]; then
    echo "  • Docker volumes"
fi
echo "  • Temporary test files"
echo "  • Coverage reports"
echo ""

# Confirm unless --force is set
if [ "$FORCE" = false ]; then
    read -p "$(echo -e ${YELLOW}Continue? [y/N]:${NC} )" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Cleanup cancelled${NC}"
        exit 0
    fi
fi

echo ""
echo -e "${BLUE}Starting cleanup...${NC}"
echo ""

# Step 1: Stop and remove Docker containers
echo -e "${BLUE}Step 1/5: Removing Docker containers...${NC}"

if command -v docker &> /dev/null; then
    # Get running containers
    CONTAINERS=$(docker ps -q --filter "name=nexmonyx" 2>/dev/null || true)
    if [ -n "$CONTAINERS" ]; then
        echo "Stopping containers: $CONTAINERS"
        docker stop $CONTAINERS 2>/dev/null || true
        docker rm $CONTAINERS 2>/dev/null || true
        echo -e "${GREEN}✓${NC} Containers removed"
    else
        echo -e "${GREEN}✓${NC} No running containers found"
    fi
else
    echo -e "${YELLOW}⚠  Docker not installed, skipping${NC}"
fi

# Step 2: Remove Docker volumes
if [ "$KEEP_VOLUMES" = false ]; then
    echo ""
    echo -e "${BLUE}Step 2/5: Removing Docker volumes...${NC}"

    if command -v docker &> /dev/null; then
        VOLUMES=$(docker volume ls -q --filter "name=nexmonyx" 2>/dev/null || true)
        if [ -n "$VOLUMES" ]; then
            echo "Removing volumes: $VOLUMES"
            docker volume rm $VOLUMES 2>/dev/null || true
            echo -e "${GREEN}✓${NC} Volumes removed"
        else
            echo -e "${GREEN}✓${NC} No test volumes found"
        fi
    fi
else
    echo ""
    echo -e "${BLUE}Step 2/5: Skipping volume cleanup (--keep-volumes)${NC}"
fi

# Step 3: Remove Docker networks
echo ""
echo -e "${BLUE}Step 3/5: Removing Docker networks...${NC}"

if command -v docker &> /dev/null; then
    if docker network ls --format "{{.Name}}" 2>/dev/null | grep -q "nexmonyx"; then
        docker network rm nexmonyx-test-network 2>/dev/null || true
        echo -e "${GREEN}✓${NC} Networks removed"
    else
        echo -e "${GREEN}✓${NC} No test networks found"
    fi
fi

# Step 4: Remove temporary test files
echo ""
echo -e "${BLUE}Step 4/5: Removing temporary files...${NC}"

TEMP_FILES=(
    "integration-coverage.out"
    "integration-coverage-summary.md"
    "coverage.out"
    "coverage.txt"
    "coverage.html"
    "coverage-summary.md"
    "security-summary.md"
    "gosec-results.json"
    ".test-env-created"
)

FOUND=0
for file in "${TEMP_FILES[@]}"; do
    if [ -f "$file" ]; then
        rm -f "$file"
        echo "  Removed: $file"
        FOUND=1
    fi
done

if [ $FOUND -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No temporary files found"
else
    echo -e "${GREEN}✓${NC} Temporary files removed"
fi

# Step 5: Reset test data (optional)
echo ""
echo -e "${BLUE}Step 5/5: Cleanup Summary${NC}"

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✓ Cleanup completed successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

echo -e "${BLUE}Verification:${NC}"
if command -v docker &> /dev/null; then
    CONTAINER_COUNT=$(docker ps -a -q --filter "name=nexmonyx" 2>/dev/null | wc -l)
    echo "  Remaining containers: $CONTAINER_COUNT"

    NETWORK_COUNT=$(docker network ls --format "{{.Name}}" 2>/dev/null | grep -c "nexmonyx" || true)
    echo "  Remaining networks: $NETWORK_COUNT"

    VOLUME_COUNT=$(docker volume ls -q --filter "name=nexmonyx" 2>/dev/null | wc -l)
    echo "  Remaining volumes: $VOLUME_COUNT"
fi

echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "  1. Run setup again: ./scripts/setup-test-env.sh"
echo "  2. Start services: docker-compose up -d"
echo "  3. Run tests: INTEGRATION_TESTS=true go test -v ./tests/integration/..."
echo ""
