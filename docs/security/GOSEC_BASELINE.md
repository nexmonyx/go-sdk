# Gosec Security Baseline

## Overview

This document describes the security baseline for the Nexmonyx Go SDK, including tracked security issues, remediation plans, and policies for maintaining security standards.

## Table of Contents

- [Current Security Status](#current-security-status)
- [Baseline Policy](#baseline-policy)
- [Tracked Issues](#tracked-issues)
- [Remediation Plan](#remediation-plan)
- [Running Security Scans](#running-security-scans)
- [Pre-Push Validation](#pre-push-validation)
- [CI/CD Integration](#cicd-integration)
- [Contributing Guidelines](#contributing-guidelines)
- [Emergency Procedures](#emergency-procedures)

## Current Security Status

**Last Updated**: 2025-10-14

### Issue Summary

| Category | Count | Status |
|----------|-------|--------|
| **Total Issues** | 153 | Tracked in baseline |
| G401 (Weak Crypto) | 2 | 🔴 Priority: HIGH |
| G104 (Unhandled Errors) | 151 | 🟡 Priority: Medium-Low |

### Severity Breakdown

- **High**: 0 (excluding baseline)
- **Medium**: 2 (G401 - weak cryptography)
- **Low**: 151 (G104 - unhandled errors)

## Baseline Policy

### Core Principles

1. **Zero Tolerance for New Issues**: All new high/critical security issues must be fixed before merging
2. **Baseline Tracking**: Existing issues are tracked in remediation tasks (not blocking)
3. **Continuous Improvement**: Security debt must trend downward over time
4. **G401 Never Allowed**: Weak cryptography (MD5, SHA1, DES, RC4) is ALWAYS blocked

### What Gets Blocked

The following will **block** a commit/push:

- ❌ **Any new G401 issues** (weak cryptographic primitives)
- ❌ **Any new HIGH severity issues**
- ❌ **Any new CRITICAL issues**
- ⚠️ **New G104 issues** (warning only in normal mode, blocked in strict mode)

### What Is Allowed (Baseline)

The following are **allowed** because they're tracked in remediation tasks:

- ✅ Existing G401 issues (2 total) - **Task #2266**
- ✅ Existing G104 issues (151 total) - **Tasks #2267-#2271**

## Tracked Issues

### G401: Use of Weak Cryptographic Primitives

**Severity**: MEDIUM
**Confidence**: HIGH
**Priority**: 🔴 HIGH
**Task**: [#2266](../../TASKS.md#2266)

**Affected Files**:
- `pkg/utils/sql_migrations.go:134` - MD5 used for migration checksums
- `pkg/migrations/sql_runner.go:148` - MD5 used for migration checksums

**Issue**: MD5 is cryptographically weak and should not be used for integrity checks.

**Remediation**: Replace `md5.Sum()` with `sha256.Sum256()` for secure checksums.

**Estimated Effort**: 2 hours

### G104: Errors Not Checked

**Severity**: LOW
**Confidence**: HIGH
**Priority**: 🟡 MEDIUM to LOW (varies by context)

G104 issues are categorized by operational area:

#### Category 1: WebSocket Operations
**Task**: [#2267](../../TASKS.md#2267)
**Priority**: MEDIUM
**Estimated Effort**: 4 hours

**Affected Files**:
- `pkg/api/audit/command_real_time_monitoring.go` (~10 instances)
- `pkg/api/agents/monitoring_proxy/websocket_proxy.go` (~10 instances)

**Impact**: Silent connection failures, lost messages, undetected disconnections

#### Category 2: Database Operations
**Task**: [#2268](../../TASKS.md#2268)
**Priority**: MEDIUM
**Estimated Effort**: 4 hours

**Affected Files**:
- `pkg/alerts/controller/alert_controller.go`
- Various database connection close operations (~30 files)

**Impact**: Resource leaks, unclosed connections, transaction failures

#### Category 3: HTTP Response Operations
**Task**: [#2269](../../TASKS.md#2269)
**Priority**: LOW
**Estimated Effort**: 3 hours

**Affected Files**:
- `pkg/agent/health/health.go` (~25 instances)
- Various HTTP write operations

**Impact**: Partial responses, silent write failures

#### Category 4: Command Execution and Audit
**Task**: [#2270](../../TASKS.md#2270)
**Priority**: 🔴 HIGH
**Estimated Effort**: 3 hours

**Affected Files**:
- `pkg/api/agents/commands/helpers.go` (~15 instances)

**Impact**: Missing audit trail, compliance violations, lost security events

#### Category 5: SSH and Network Operations
**Task**: [#2271](../../TASKS.md#2271)
**Priority**: MEDIUM
**Estimated Effort**: 3 hours

**Affected Files**:
- `pkg/agent/probes/executors.go` (~10 instances)

**Impact**: Resource leaks, connection pool exhaustion, timeout issues

## Remediation Plan

### Priority Order

1. **Task #2266** (G401) - Replace MD5 with SHA256 - **2 hours** 🔴
2. **Task #2270** (G104) - Fix audit logging errors - **3 hours** 🔴
3. **Task #2267** (G104) - Fix WebSocket errors - **4 hours**
4. **Task #2268** (G104) - Fix database errors - **4 hours**
5. **Task #2271** (G104) - Fix SSH/network errors - **3 hours**
6. **Task #2269** (G104) - Fix HTTP response errors - **3 hours**

**Total Estimated Effort**: 19 hours

### Timeline

- **Sprint 1**: Tasks #2266, #2270 (High priority) - 5 hours
- **Sprint 2**: Tasks #2267, #2268 (Medium priority) - 8 hours
- **Sprint 3**: Tasks #2271, #2269 (Remaining) - 6 hours

**Target**: Zero security issues within 3 sprints

## Running Security Scans

### Local Security Scan

Run gosec with baseline validation:

```bash
# Standard scan
./scripts/check-gosec.sh

# Strict mode (blocks on any new issue)
./scripts/check-gosec.sh --strict

# JSON output
./scripts/check-gosec.sh --format=json
```

### Manual gosec Scan

Run gosec directly:

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run scan with JSON output
gosec -fmt=json -out=gosec-results.json -exclude-generated ./...

# Run scan with text output
gosec -exclude-generated ./...
```

### Analyzing Results

View detailed results:

```bash
# View JSON results
cat gosec-results.json | jq '.Issues[] | {rule: .rule_id, file: .file, line: .line, severity: .severity}'

# Count by severity
cat gosec-results.json | jq '[.Issues[] | .severity] | group_by(.) | map({severity: .[0], count: length})'

# Filter high severity only
cat gosec-results.json | jq '.Issues[] | select(.severity == "HIGH")'
```

## Pre-Push Validation

### Automatic Validation

A pre-push git hook automatically runs security scans before allowing pushes.

**Location**: `.git/hooks/pre-push`

### Setup Pre-Push Hook

The hook is installed automatically during repository setup. To reinstall:

```bash
# The hook is already created at .git/hooks/pre-push
# Just ensure it's executable
chmod +x .git/hooks/pre-push
```

### Hook Behavior

**On Security Issues**:
1. Runs gosec scan
2. Validates against baseline
3. **Blocks push** if new high/critical issues found
4. Shows detailed error report
5. Provides remediation guidance

**On Success**:
- Push proceeds normally
- Shows security summary

### Bypassing the Hook

⚠️ **Use with extreme caution**

```bash
# Emergency bypass
git push --no-verify
```

**Requirements for bypass**:
1. ✅ Clear justification (critical hotfix, security incident)
2. ✅ Create immediate remediation task
3. ✅ Notify security team
4. ✅ Document in commit message

## CI/CD Integration

### GitHub Actions Workflow

Security scanning is integrated into the CI/CD pipeline at `.github/workflows/ci.yml`.

### Workflow Steps

1. **Install Tools**: gosec, staticcheck
2. **Run Security Scan**: Execute `./scripts/check-gosec.sh --ci`
3. **Generate Reports**: Create security summary
4. **Upload Artifacts**: Store gosec results (30-day retention)
5. **PR Comments**: Post security summary on pull requests
6. **Fail Build**: Block merge if new high/critical issues found

### Viewing Results

**In GitHub Actions**:
1. Go to Actions tab
2. Select your workflow run
3. View "Security scanning with baseline" step
4. Download "security-scan-results" artifact

**In Pull Requests**:
- Automated comment with security summary
- Links to detailed reports
- Status check (required to pass)

## Contributing Guidelines

### For Contributors

When contributing code:

1. ✅ **Run security scan locally** before committing
2. ✅ **Fix any new high/critical issues** immediately
3. ✅ **Handle all errors properly** (avoid new G104 issues)
4. ✅ **Never use weak cryptography** (G401)
5. ✅ **Document security decisions** in code comments

### Best Practices

#### Error Handling

```go
// ❌ BAD: Ignored error
conn.Close()

// ✅ GOOD: Explicit ignore with justification
_ = conn.Close() // Connection already closed, error not critical

// ✅ BEST: Handle the error
if err := conn.Close(); err != nil {
    log.Warn("Failed to close connection", "error", err)
}
```

#### Cryptographic Operations

```go
// ❌ NEVER: Weak cryptography
checksum := md5.Sum(data)

// ✅ ALWAYS: Strong cryptography
checksum := sha256.Sum256(data)
```

### Code Review Checklist

Reviewers should verify:

- [ ] Security scan passes in CI/CD
- [ ] No new high/critical issues introduced
- [ ] All errors handled appropriately
- [ ] Strong cryptography used where applicable
- [ ] Security best practices followed

## Emergency Procedures

### Hotfix Process

For critical production issues requiring immediate deployment:

1. **Create hotfix branch**: `git checkout -b hotfix/critical-issue`
2. **Make minimal changes**: Only fix the critical issue
3. **Run security scan**: `./scripts/check-gosec.sh`
4. **Document bypass** (if needed):
   ```bash
   # In commit message
   SECURITY: Using --no-verify due to [INCIDENT-123]
   Justification: Production down, customer impact
   Remediation: Task #XXXX created for security fixes
   ```
5. **Push with bypass**: `git push --no-verify` (if required)
6. **Immediate follow-up**: Create remediation task within 24 hours
7. **Security review**: Post-deployment security audit required

### Incident Response

If security vulnerability discovered:

1. **Report immediately**: Contact security team
2. **Create private issue**: Don't disclose publicly
3. **Assess impact**: Determine affected versions
4. **Develop fix**: Create patch for vulnerability
5. **Test thoroughly**: Verify fix with security scan
6. **Deploy urgently**: Follow hotfix process
7. **Disclose responsibly**: After fix is deployed

## Metrics and Tracking

### Security Debt Metrics

Track progress monthly:

| Month | Total Issues | High | Medium | Low | Trend |
|-------|--------------|------|--------|-----|-------|
| Oct 2025 | 153 | 0 | 2 | 151 | 📊 Baseline |
| Nov 2025 | TBD | TBD | TBD | TBD | TBD |

### Success Criteria

- ✅ **Zero new high/critical issues** in main branch
- ✅ **Downward trend** in total issue count
- ✅ **All G401 issues resolved** by end of Sprint 1
- ✅ **All G104 high-priority issues resolved** by end of Sprint 2
- 🎯 **Zero total issues** by end of Sprint 3

## Resources

### Documentation

- [gosec Official Documentation](https://github.com/securego/gosec)
- [OWASP Go Secure Coding Practices](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [Go Security Best Practices](https://golang.org/doc/security/best-practices)

### Tools

- [gosec](https://github.com/securego/gosec) - Go Security Checker
- [staticcheck](https://staticcheck.io/) - Go Static Analyzer
- [gosec-baseline](https://github.com/securego/gosec#baseline) - Baseline Management

### Internal Links

- [Contributing Guidelines](../../CONTRIBUTING.md)
- [Security Policy](../../SECURITY.md)
- [Task Tracking](../../TASKS.md)

## Questions?

For questions about security baseline or policies:

- **Security Team**: security@nexmonyx.com
- **Documentation Issues**: Open GitHub issue
- **Urgent Security Concerns**: See [Emergency Procedures](#emergency-procedures)

---

**Last Updated**: 2025-10-14
**Document Version**: 1.0
**Maintained By**: Security Team
