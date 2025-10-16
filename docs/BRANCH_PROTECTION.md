# Branch Protection Rules

This document describes the recommended branch protection rules for the Nexmonyx Go SDK repository to ensure code quality, security, and stability.

## Overview

Branch protection rules enforce best practices by requiring code reviews, status checks, and other safeguards before changes can be merged to protected branches.

## Recommended Configuration

### Protected Branches

The following branches should be protected:
- `main` - Production-ready code
- `master` - Legacy default branch (if used)

### Branch Protection Settings

#### 1. Require Pull Request Reviews

**Requirement**: ✅ **REQUIRED**

- **Minimum reviewers**: 1
- **Dismiss stale reviews**: Yes
- **Require review from code owners**: No (optional)
- **Restrict who can dismiss reviews**: No

**Rationale**: All changes must be reviewed before merge to catch issues early.

#### 2. Require Status Checks

**Requirement**: ✅ **REQUIRED**

Status checks that must pass before merging:

| Check Name | Status | Critical |
|-----------|--------|----------|
| `test-and-build` | Required | Yes |
| `integration-tests-mock` | Required | Yes |
| `security-scan` | Required | Yes |

**Additional Settings**:
- **Require branches to be up to date**: Yes
- **Require conversations to be resolved**: Yes (optional)

**Rationale**: Automated testing, security scanning, and integration tests validate code quality before merge.

#### 3. Require Linear History

**Requirement**: ⚠️ **OPTIONAL**

- **Enforce**: No (allows merge commits)

**Rationale**: Allow flexibility in merge strategies while maintaining clear history.

#### 4. Require Code Owner Reviews

**Requirement**: ❌ **NOT REQUIRED** (for initial setup)

Can be enabled in the future by creating a `CODEOWNERS` file.

#### 5. Include Administrators

**Requirement**: ✅ **RECOMMENDED**

- **Include administrators**: Yes

**Rationale**: Ensures admins follow the same review process.

#### 6. Allow Force Pushes

**Requirement**: ✅ **DISABLE**

- **Allow force pushes**: None (disabled)
- **Allow deletions**: No

**Rationale**: Prevents accidental history rewrites that could lose commits.

## Setup Instructions

### Method 1: GitHub Web UI

1. Go to repository **Settings** → **Branches**
2. Click **Add rule** under "Branch protection rules"
3. Enter branch name pattern: `main`
4. Configure settings as described above:
   - ✅ Require a pull request before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators
   - ❌ Allow force pushes (disable)
   - ❌ Allow deletions (disable)
5. Click **Create**
6. Repeat for `master` branch if applicable

### Method 2: GitHub CLI

Use these commands to configure branch protection via GitHub CLI:

```bash
# Protect main branch with required checks
gh repo rule create \
  --branch main \
  --require-pull-request-reviews \
  --require-status-checks \
  --required-status-checks test-and-build integration-tests-mock security-scan \
  --require-up-to-date-before-merge \
  --include-administrators \
  --dismiss-stale-reviews
```

Or use the API:

```bash
curl -X PUT \
  -H "Authorization: token YOUR_GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/nexmonyx/go-sdk/branches/main/protection \
  -d '{
    "required_status_checks": {
      "strict": true,
      "contexts": ["test-and-build", "integration-tests-mock", "security-scan"]
    },
    "required_pull_request_reviews": {
      "dismissal_restrictions": {},
      "dismiss_stale_reviews": true,
      "require_code_owner_reviews": false,
      "required_approving_review_count": 1
    },
    "enforce_admins": true,
    "allow_force_pushes": false,
    "allow_deletions": false,
    "required_linear_history": false
  }'
```

### Method 3: Infrastructure as Code (Terraform)

```hcl
resource "github_branch_protection" "main" {
  repository_id = github_repository.sdk.node_id

  pattern          = "main"
  enforce_admins   = true
  allows_deletions = false
  allows_force_pushes = false

  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    restrict_dismissals             = false
    required_approving_review_count = 1
  }

  required_status_checks {
    strict   = true
    contexts = ["test-and-build", "integration-tests-mock", "security-scan"]
  }
}
```

## Required Status Checks Explained

### `test-and-build`
- **Job**: `test-and-build` in `.github/workflows/ci.yml`
- **Purpose**: Runs unit tests, builds SDK, checks coverage
- **Failure**: Indicates test failure or build issue
- **Action**: Review test logs and fix failures before retry

### `integration-tests-mock`
- **Job**: `integration-tests-mock` in `.github/workflows/integration-tests.yml`
- **Purpose**: Runs integration tests against mock API
- **Failure**: Indicates integration test failure
- **Action**: Review test logs and fix integration issues

### `security-scan`
- **Job**: `security-scan` in `.github/workflows/security-nightly.yml` (or on-demand)
- **Purpose**: Runs gosec security scanner and vulnerability checks
- **Failure**: Indicates security issue
- **Action**: Review security findings and address vulnerabilities

## Workflow When Branch Protection Is Enabled

```
Developer                    GitHub
    |                           |
    +-- Create PR ------------>  |
    |                           |
    |                    <-- Checks start
    |                           |
    |                    <-- test-and-build running
    |                    <-- integration-tests running
    |                    <-- security checks running
    |                           |
    |      (if failed)   <-- Checks fail (red X)
    +-- Fix issues ------------>  |
    |                           |
    |      (if passed)   <-- All checks pass (green ✓)
    |                           |
    |                    <-- Request review
    +-- Request review -------->  |
    |                           |
    |  <-- Reviewer reviews  --> Reviewer
    |                           |
    |                    <-- Review approved
    |                           |
    +-- Merge PR -------------->  |
    |                           |
    |                    <-- Merged to main
```

## Best Practices

### For Developers

1. **Branch Early**: Create a branch before starting work
2. **Small PRs**: Keep PRs focused and easy to review
3. **Meaningful Commits**: Use clear, descriptive commit messages
4. **Test Locally**: Run tests locally before pushing
5. **Respond to Reviews**: Address review feedback promptly

### For Reviewers

1. **Check Tests**: Verify all status checks are passing
2. **Review Code**: Look for bugs, style issues, security concerns
3. **Test Changes**: Consider testing the changes locally
4. **Constructive Feedback**: Provide helpful, respectful feedback
5. **Approve When Ready**: Approve only when confident

### For Maintainers

1. **Monitor PRs**: Keep PR queue current
2. **Enforce Rules**: Don't bypass branch protection
3. **Update Rules**: Adjust as needed based on team feedback
4. **Document Changes**: Update this guide if rules change

## Bypassing Branch Protection

⚠️ **Use with extreme caution!**

Branch protection can be temporarily dismissed in emergencies:

```bash
# Using GitHub CLI (requires admin access)
gh pr merge <PR_NUMBER> --admin --merge
```

**When to use**: Critical production hotfix that requires immediate deployment

**Requirements**:
- Admin access to repository
- Document reason in PR description
- Include post-mortem analysis

## Troubleshooting

### "Commit cannot be merged" error

**Cause**: Branch protection requires all status checks to pass

**Solution**:
1. Wait for all status checks to complete
2. If failing, review logs and fix issues
3. Push fixes to branch
4. Wait for checks to re-run

### "This branch is out of date" error

**Cause**: `Require branches to be up to date` setting enabled

**Solution**:
```bash
git pull origin main
git push origin your-branch
```

Or use GitHub UI: Click "Update branch" button on PR

### Status check not appearing

**Cause**: Workflow may not be configured for pull requests

**Solution**: Verify workflow triggers include `pull_request`

```yaml
on:
  push:
    branches: [main]
  pull_request:            # <-- Add this
    branches: [main]
```

## Monitoring and Metrics

Use GitHub dashboards to monitor branch protection:

- **PR Merge Rate**: Measure how many PRs are being merged
- **Review Time**: Track average time from PR open to merge
- **Status Check Failures**: Monitor which checks fail most often
- **Protected Branch Activity**: See commits to protected branches

## Related Documentation

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule)
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development workflow
- [CI/CD Workflows](../.github/workflows/) - Automated checks
- [TESTING.md](./TESTING.md) - Testing standards

## Questions?

- 📖 See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines
- 🐛 Report issues on GitHub Issues
- 💬 Discuss in GitHub Discussions
