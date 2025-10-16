#!/bin/bash
#
# Setup integration test environment for Nexmonyx Go SDK
#
# This script configures the local development environment for integration testing:
# - Creates .env file from template
# - Validates Docker installation
# - Checks Go installation
# - Sets up local test credentials
# - Verifies environment variables
#
# Usage:
#   ./scripts/setup-test-env.sh [--docker-only] [--skip-docker]
#
# Options:
#   --docker-only     Only set up Docker, skip Go checks
#   --skip-docker     Skip Docker checks
#   --help            Show this help message
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
DOCKER_ONLY=false
SKIP_DOCKER=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --docker-only)
            DOCKER_ONLY=true
            shift
            ;;
        --skip-docker)
            SKIP_DOCKER=true
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
echo -e "${BLUE}  Integration Test Environment Setup${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if we're in the root directory
if [ ! -f "go.mod" ] || [ ! -f ".env.example" ]; then
    echo -e "${RED}❌ Error: .env.example not found${NC}"
    echo -e "${YELLOW}Please run this script from the repository root directory${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Running from repository root"

# Step 1: Check Docker (unless --skip-docker is set)
if [ "$SKIP_DOCKER" = false ]; then
    echo ""
    echo -e "${BLUE}Step 1/5: Checking Docker installation...${NC}"

    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed${NC}"
        echo -e "${YELLOW}Please install Docker from: https://docs.docker.com/install/${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Docker found: $(docker --version)"

    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed${NC}"
        echo -e "${YELLOW}Please install Docker Compose${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Docker Compose found: $(docker-compose --version)"

    # Check if Docker daemon is running
    if ! docker ps &> /dev/null; then
        echo -e "${RED}❌ Docker daemon is not running${NC}"
        echo -e "${YELLOW}Please start Docker and try again${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Docker daemon is running"
fi

# Step 2: Check Go (unless --docker-only is set)
if [ "$DOCKER_ONLY" = false ]; then
    echo ""
    echo -e "${BLUE}Step 2/5: Checking Go installation...${NC}"

    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        echo -e "${YELLOW}Please install Go 1.24 or later from: https://golang.org/dl/${NC}"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go found: $GO_VERSION"

    # Check if Git is available (needed for Go modules)
    if ! command -v git &> /dev/null; then
        echo -e "${RED}❌ Git is not installed${NC}"
        echo -e "${YELLOW}Please install Git${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Git found: $(git --version)"
fi

# Step 3: Create .env file
echo ""
echo -e "${BLUE}Step 3/5: Creating .env file...${NC}"

if [ -f ".env" ]; then
    echo -e "${YELLOW}⚠  .env file already exists${NC}"
    read -p "$(echo -e ${YELLOW}Overwrite? [y/N]:${NC} )" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Keeping existing .env file${NC}"
    else
        cp .env.example .env
        echo -e "${GREEN}✓${NC} .env file created"
    fi
else
    cp .env.example .env
    echo -e "${GREEN}✓${NC} .env file created"
fi

# Step 4: Validate .env file
echo ""
echo -e "${BLUE}Step 4/5: Validating environment configuration...${NC}"

if grep -q "^INTEGRATION_TESTS=" .env; then
    echo -e "${GREEN}✓${NC} INTEGRATION_TESTS configured"
else
    echo -e "${YELLOW}⚠  INTEGRATION_TESTS not configured${NC}"
fi

if grep -q "^INTEGRATION_TEST_MODE=" .env; then
    MODE=$(grep "^INTEGRATION_TEST_MODE=" .env | cut -d= -f2 | tr -d ' ')
    echo -e "${GREEN}✓${NC} INTEGRATION_TEST_MODE set to: $MODE"
else
    echo -e "${YELLOW}⚠  INTEGRATION_TEST_MODE not configured${NC}"
fi

if grep -q "^INTEGRATION_TEST_API_URL=" .env; then
    echo -e "${GREEN}✓${NC} INTEGRATION_TEST_API_URL configured"
else
    echo -e "${YELLOW}⚠  INTEGRATION_TEST_API_URL not configured${NC}"
fi

# Step 5: Summary and next steps
echo ""
echo -e "${BLUE}Step 5/5: Setup Complete${NC}"

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✓ Integration test environment configured!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

echo -e "${BLUE}Next Steps:${NC}"
echo ""
echo "1. Review and customize .env file:"
echo -e "   ${YELLOW}cat .env${NC}"
echo ""
echo "2. Start mock API server:"
echo -e "   ${YELLOW}docker-compose up -d${NC}"
echo ""
echo "3. Verify services are running:"
echo -e "   ${YELLOW}docker-compose ps${NC}"
echo ""
echo "4. Run integration tests:"
echo -e "   ${YELLOW}source .env && go test -v ./tests/integration/...${NC}"
echo ""
echo "5. Stop services when done:"
echo -e "   ${YELLOW}docker-compose down${NC}"
echo ""

echo -e "${BLUE}Useful Commands:${NC}"
echo "  View logs:     docker-compose logs -f mock-api"
echo "  Shell access:  docker-compose exec mock-api sh"
echo "  API health:    curl http://localhost:8080/health"
echo ""

echo -e "${BLUE}For more information, see:${NC}"
echo "  docs/INTEGRATION_TESTING.md"
echo "  tests/integration/README.md"
echo "  tests/integration/docker/README.md"
echo ""
