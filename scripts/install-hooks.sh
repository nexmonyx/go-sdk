#!/bin/bash
#
# Install pre-commit hooks for Nexmonyx Go SDK
#
# This script installs and configures pre-commit hooks to ensure code quality
# and consistency before commits are made.
#
# Usage:
#   ./scripts/install-hooks.sh
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Nexmonyx Go SDK - Pre-commit Hook Installer${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if we're in the root of the repository
if [ ! -f ".pre-commit-config.yaml" ]; then
    echo -e "${RED}❌ Error: .pre-commit-config.yaml not found${NC}"
    echo -e "${YELLOW}Please run this script from the repository root${NC}"
    exit 1
fi

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Error: Python 3 is not installed${NC}"
    echo -e "${YELLOW}Please install Python 3 to use pre-commit hooks${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Python 3 found: $(python3 --version)"

# Check if pip is installed
if ! command -v pip3 &> /dev/null; then
    echo -e "${RED}❌ Error: pip3 is not installed${NC}"
    echo -e "${YELLOW}Please install pip3 to continue${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} pip3 found"

# Install or upgrade pre-commit
echo ""
echo -e "${BLUE}Installing pre-commit framework...${NC}"
if pip3 install --user --upgrade pre-commit; then
    echo -e "${GREEN}✓${NC} pre-commit installed/upgraded successfully"
else
    echo -e "${RED}❌ Failed to install pre-commit${NC}"
    exit 1
fi

# Check if Go tools are installed
echo ""
echo -e "${BLUE}Checking required Go tools...${NC}"

MISSING_TOOLS=()

if ! command -v gosec &> /dev/null; then
    MISSING_TOOLS+=("gosec")
fi

if ! command -v golangci-lint &> /dev/null; then
    MISSING_TOOLS+=("golangci-lint")
fi

if ! command -v goimports &> /dev/null; then
    MISSING_TOOLS+=("goimports")
fi

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "${YELLOW}⚠ Missing Go tools: ${MISSING_TOOLS[*]}${NC}"
    echo ""
    echo -e "${BLUE}Installing missing Go tools...${NC}"

    for tool in "${MISSING_TOOLS[@]}"; do
        case $tool in
            gosec)
                go install github.com/securego/gosec/v2/cmd/gosec@latest
                ;;
            golangci-lint)
                go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
                ;;
            goimports)
                go install golang.org/x/tools/cmd/goimports@latest
                ;;
        esac
        echo -e "${GREEN}✓${NC} Installed $tool"
    done
else
    echo -e "${GREEN}✓${NC} All required Go tools are installed"
fi

# Install pre-commit hooks
echo ""
echo -e "${BLUE}Installing pre-commit hooks...${NC}"
if pre-commit install; then
    echo -e "${GREEN}✓${NC} Pre-commit hooks installed"
else
    echo -e "${RED}❌ Failed to install pre-commit hooks${NC}"
    exit 1
fi

# Install commit-msg hook
if pre-commit install --hook-type commit-msg; then
    echo -e "${GREEN}✓${NC} Commit-msg hook installed"
else
    echo -e "${YELLOW}⚠ Failed to install commit-msg hook (non-critical)${NC}"
fi

# Run pre-commit on all files (optional)
echo ""
read -p "$(echo -e ${YELLOW}Do you want to run pre-commit on all files now? [y/N]:${NC} )" -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Running pre-commit on all files...${NC}"
    if pre-commit run --all-files; then
        echo -e "${GREEN}✓${NC} Pre-commit checks passed on all files"
    else
        echo -e "${YELLOW}⚠ Some pre-commit checks failed - please review and fix${NC}"
        echo -e "${YELLOW}  You can run 'pre-commit run --all-files' again later${NC}"
    fi
fi

# Success message
echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✓ Pre-commit hooks installed successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "${BLUE}Hooks will now run automatically before each commit.${NC}"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  ${YELLOW}pre-commit run --all-files${NC}  - Run hooks on all files"
echo -e "  ${YELLOW}pre-commit run${NC}              - Run hooks on staged files"
echo -e "  ${YELLOW}git commit --no-verify${NC}      - Skip hooks (use sparingly!)"
echo -e "  ${YELLOW}pre-commit autoupdate${NC}       - Update hook versions"
echo ""
echo -e "${BLUE}For more information, see:${NC}"
echo -e "  ${YELLOW}https://pre-commit.com${NC}"
echo ""
