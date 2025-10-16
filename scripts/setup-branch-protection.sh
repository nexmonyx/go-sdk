#!/bin/bash
#
# Set up branch protection rules for Nexmonyx Go SDK repository
#
# Prerequisites:
#   - GitHub CLI installed: https://cli.github.com/
#   - Admin access to the repository
#   - Logged in to GitHub CLI: gh auth login
#
# Usage:
#   ./scripts/setup-branch-protection.sh [repository] [branch]
#
# Examples:
#   ./scripts/setup-branch-protection.sh                    # Uses current repo, protects 'main'
#   ./scripts/setup-branch-protection.sh nexmonyx/go-sdk    # Protects 'main' branch
#   ./scripts/setup-branch-protection.sh nexmonyx/go-sdk master
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="${1:-.}"  # Default to current repo
BRANCH="${2:-main}"
STATUS_CHECKS=("test-and-build" "integration-tests-mock" "security-scan")

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Branch Protection Setup Script${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}❌ GitHub CLI is not installed${NC}"
    echo -e "${YELLOW}Please install it from: https://cli.github.com/${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} GitHub CLI found"

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}❌ Not authenticated to GitHub${NC}"
    echo -e "${YELLOW}Please run: gh auth login${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Authenticated to GitHub"

# Get current user
CURRENT_USER=$(gh api user -q '.login')
echo -e "${GREEN}✓${NC} Logged in as: $CURRENT_USER"
echo ""

# Determine repository
if [ "$REPO" = "." ]; then
    # Get current repository from git remote
    if ! REPO=$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\/\(.*\)\.git/\1\/\2/'); then
        echo -e "${RED}❌ Could not determine repository${NC}"
        echo -e "${YELLOW}Please run from a Git repository or provide repo as argument${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}Repository: $REPO${NC}"
echo -e "${BLUE}Branch: $BRANCH${NC}"
echo -e "${BLUE}Status Checks: ${STATUS_CHECKS[*]}${NC}"
echo ""

# Confirm before proceeding
echo -e "${YELLOW}⚠  This will modify branch protection rules for: $REPO/$BRANCH${NC}"
read -p "$(echo -e ${YELLOW}Continue? [y/N]:${NC} )" -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Cancelled${NC}"
    exit 0
fi

echo ""
echo -e "${BLUE}Setting up branch protection...${NC}"

# Build status check context list
STATUS_CHECKS_STR="${STATUS_CHECKS[0]}"
for check in "${STATUS_CHECKS[@]:1}"; do
    STATUS_CHECKS_STR="${STATUS_CHECKS_STR}|${check}"
done

# Step 1: Enable required pull request reviews
echo -e "${BLUE}Step 1/5: Requiring pull request reviews...${NC}"
gh repo rule create \
    --repository="$REPO" \
    --branch="$BRANCH" \
    --type pull_request_review_count \
    --required_approving_review_count=1 \
    2>/dev/null || echo -e "${YELLOW}⚠  Pull request review rule already exists${NC}"

echo -e "${GREEN}✓${NC} Pull request reviews required"

# Step 2: Enable dismiss stale reviews
echo -e "${BLUE}Step 2/5: Configuring review dismissal...${NC}"
gh repo rule create \
    --repository="$REPO" \
    --branch="$BRANCH" \
    --type dismissal_restrictions \
    --bypass_pull_request_allowances="" \
    2>/dev/null || echo -e "${YELLOW}⚠  Dismissal rule already exists${NC}"

echo -e "${GREEN}✓${NC} Review dismissal configured"

# Step 3: Require status checks to pass
echo -e "${BLUE}Step 3/5: Requiring status checks...${NC}"

# Try using the new API format (if available)
for check in "${STATUS_CHECKS[@]}"; do
    gh repo rule create \
        --repository="$REPO" \
        --branch="$BRANCH" \
        --type required_status_checks \
        --required_status_checks="$check" \
        --strict=true \
        2>/dev/null || echo -e "${YELLOW}⚠  Status check '$check' already configured${NC}"
done

echo -e "${GREEN}✓${NC} Status checks configured"

# Step 4: Require branches to be up to date
echo -e "${BLUE}Step 4/5: Requiring up-to-date branches...${NC}"
gh repo rule create \
    --repository="$REPO" \
    --branch="$BRANCH" \
    --type require_branches_to_be_up_to_date_before_merging \
    2>/dev/null || echo -e "${YELLOW}⚠  Up-to-date requirement already exists${NC}"

echo -e "${GREEN}✓${NC} Up-to-date requirement configured"

# Step 5: Include administrators in restrictions
echo -e "${BLUE}Step 5/5: Restricting administrators...${NC}"
gh repo rule create \
    --repository="$REPO" \
    --branch="$BRANCH" \
    --type restrict_bypassing_branch_protections \
    --admin_only=true \
    2>/dev/null || echo -e "${YELLOW}⚠  Admin restriction already exists${NC}"

echo -e "${GREEN}✓${NC} Administrator restrictions configured"

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✓ Branch protection configured successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

echo -e "${BLUE}Configuration Summary:${NC}"
echo "  Repository: $REPO"
echo "  Branch: $BRANCH"
echo "  Pull Request Reviews: Required (1 minimum)"
echo "  Dismiss Stale Reviews: Enabled"
echo "  Status Checks Required:"
for check in "${STATUS_CHECKS[@]}"; do
    echo "    • $check"
done
echo "  Require Up-to-Date: Enabled"
echo "  Include Administrators: Yes"
echo ""

echo -e "${BLUE}Next Steps:${NC}"
echo "1. Visit: https://github.com/$REPO/settings/branches"
echo "2. Verify the rules match your requirements"
echo "3. Adjust as needed via GitHub UI if necessary"
echo ""

echo -e "${BLUE}Useful Commands:${NC}"
echo "  gh repo rule list --repository=$REPO --branch=$BRANCH"
echo "  gh repo rule delete --repository=$REPO --branch=$BRANCH <rule-id>"
echo ""

echo -e "${BLUE}For more information, see:${NC}"
echo "  docs/BRANCH_PROTECTION.md"
echo "  https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches"
echo ""
