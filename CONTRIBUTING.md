# Contributing to Nexmonyx Go SDK

Thank you for your interest in contributing to the Nexmonyx Go SDK! This document provides guidelines and requirements for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing Requirements](#testing-requirements)
- [Code Quality Standards](#code-quality-standards)
  - [Security Standards](#security-standards)
- [Pull Request Process](#pull-request-process)
- [Commit Message Guidelines](#commit-message-guidelines)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- Access to the Nexmonyx API (for integration tests)

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-sdk.git
   cd go-sdk
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/nexmonyx/go-sdk.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```

## Development Workflow

1. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Run tests and checks:
   ```bash
   # Format code
   go fmt ./...

   # Run linters
   go vet ./...
   staticcheck ./...

   # Run tests with coverage
   go test -v -race -coverprofile=coverage.out ./...

   # Check coverage thresholds
   ./scripts/check-coverage.sh coverage.out
   ```

4. Commit your changes (see [Commit Message Guidelines](#commit-message-guidelines))

5. Push to your fork and create a pull request

## Testing Requirements

We maintain high test coverage standards to ensure code quality and reliability. **All contributions must meet these requirements:**

### Coverage Thresholds

The following coverage thresholds are **enforced by CI/CD** and will cause builds to fail if not met:

#### Overall Project Coverage
- **Minimum: 80%**
- The entire SDK must maintain at least 80% test coverage
- This threshold is checked on every commit and PR

#### Per-Package Coverage
- **Minimum: 70%**
- Each package must have at least 70% coverage
- New packages must meet this threshold before merging

#### Critical Files Coverage
- **Minimum: 90%**
- Core files require higher coverage: `client.go`, `errors.go`, `models.go`, `response.go`
- These files are critical to SDK functionality

#### New Code in Pull Requests
- **Minimum: 85%**
- Any new or modified code in PRs must have at least 85% coverage
- PRs that decrease overall coverage will be rejected

### Writing Tests

All new functionality must include comprehensive tests:

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Table-Driven Tests**: Use for testing multiple scenarios
3. **Error Cases**: Test all error paths and edge cases
4. **Mock Services**: Use `httptest` for API mocking

Example test structure:
```go
func TestServiceMethod(t *testing.T) {
    tests := []struct {
        name       string
        input      InputType
        mockStatus int
        mockBody   interface{}
        wantErr    bool
        checkFunc  func(*testing.T, *Result)
    }{
        {
            name:  "successful operation",
            input: validInput,
            mockStatus: http.StatusOK,
            mockBody: validResponse,
            wantErr: false,
            checkFunc: func(t *testing.T, result *Result) {
                assert.Equal(t, expected, result)
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage thresholds
./scripts/check-coverage.sh coverage.out

# Run with race detection
go test -v -race ./...
```

### Coverage Configuration

Coverage requirements are defined in `.coveragerc`. To request an exemption for a specific package:

1. Add justification to `.coveragerc` under `[exemptions]`
2. Document why the package cannot meet the threshold
3. Include in your PR description

## Code Quality Standards

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `go vet` and `staticcheck` before committing
- Keep functions focused and single-purpose
- Add comments for exported functions and complex logic

### Documentation

- All exported functions, types, and methods must have doc comments
- Follow Go doc comment conventions
- Update README.md if adding new features
- Include code examples for new functionality

### Error Handling

- Return errors, don't panic (except for truly unrecoverable situations)
- Wrap errors with context using `fmt.Errorf` with `%w`
- Use SDK-specific error types when appropriate

### Security Standards

We maintain strict security standards to protect our users and their data. **All contributions must pass security scanning before merging.**

#### Security Scanning

The SDK uses [gosec](https://github.com/securego/gosec) to identify security vulnerabilities in Go code. Security scans run:

1. **Locally**: Via pre-push git hook (automatic)
2. **CI/CD**: On every commit and pull request (required)

#### Pre-Push Security Hook

A git hook automatically runs security scans before allowing pushes:

**Setup** (if not already installed):
```bash
# Hook is located at .git/hooks/pre-push
# Ensure it's executable
chmod +x .git/hooks/pre-push
```

**What it checks**:
- 🔴 New weak cryptography usage (G401) - **ALWAYS BLOCKED**
- 🔴 New HIGH severity issues - **BLOCKED**
- 🟡 New unhandled errors (G104) - **WARNING**

**Running manually**:
```bash
# Standard security scan
./scripts/check-gosec.sh

# Strict mode (blocks on any new issue)
./scripts/check-gosec.sh --strict
```

#### Security Baseline

The SDK maintains a **security baseline** tracking existing issues:

- **Total tracked issues**: 15 (being actively remediated)
- **G104 (Unhandled Errors)**: 5 issues → Low priority
- **G115 (Integer Overflow)**: 10 issues → Medium priority
- ✅ **All G401 (weak crypto) issues already fixed!**

**Policy**: New high/critical issues are **blocked**. Existing baseline issues are **allowed** while tracked for remediation.

**Full documentation**: See [docs/security/GOSEC_BASELINE.md](docs/security/GOSEC_BASELINE.md)

#### Security Best Practices

When contributing code, follow these security guidelines:

**Error Handling** (Avoid G104):
```go
// ❌ BAD: Ignored error
conn.Close()

// ✅ ACCEPTABLE: Explicit ignore with justification
_ = conn.Close() // Connection cleanup, error not critical

// ✅ BEST: Handle the error
if err := conn.Close(); err != nil {
    log.Warn("Failed to close connection", "error", err)
}
```

**Cryptographic Operations** (Avoid G401):
```go
// ❌ NEVER: Weak cryptography
import "crypto/md5"
checksum := md5.Sum(data)

// ✅ ALWAYS: Strong cryptography
import "crypto/sha256"
checksum := sha256.Sum256(data)
```

**Sensitive Data**:
- Never hardcode credentials (G101)
- Use environment variables or secure vaults
- Don't log sensitive information

**SQL Queries**:
- Use parameterized queries (avoid G201, G202)
- Never concatenate user input into SQL

#### Emergency Bypass

⚠️ **Use only for critical production incidents**

```bash
# Bypass pre-push hook
git push --no-verify
```

**Requirements for bypass**:
1. Clear justification in commit message
2. Create remediation task immediately (within 24 hours)
3. Notify security team
4. Document in PR description

Example commit message:
```
fix(critical): emergency hotfix for production outage

SECURITY: Using --no-verify due to INCIDENT-123
Justification: Production down, 1000+ users affected
Remediation: Task #XXXX created for security review
```

## Pull Request Process

### Before Submitting

1. ✅ All tests pass
2. ✅ Coverage thresholds met (checked automatically)
3. ✅ **Security scan passes** (`./scripts/check-gosec.sh`)
4. ✅ Code is formatted (`go fmt ./...`)
5. ✅ No lint errors (`go vet ./...`, `staticcheck ./...`)
6. ✅ Commit messages follow guidelines
7. ✅ Documentation is updated
8. ✅ Branch is up to date with `main`

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated (if applicable)
- [ ] All tests passing
- [ ] Coverage thresholds met

## Coverage Report
- Overall coverage: X.X%
- New code coverage: X.X%
- Files affected: X

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
```

### Review Process

1. **Automated Checks**: CI/CD runs tests, coverage checks, and linters
2. **Coverage Enforcement**: Build fails if coverage drops below thresholds
3. **Code Review**: Maintainers review code quality and design
4. **Approval**: At least one maintainer approval required
5. **Merge**: Squash and merge to `main`

### CI/CD Checks

The following checks must pass:
- ✅ All unit tests
- ✅ **Security scan passes (no new high/critical issues)**
- ✅ Coverage ≥ 80% (overall)
- ✅ Coverage ≥ 70% (per package)
- ✅ Coverage ≥ 90% (critical files)
- ✅ Coverage ≥ 85% (new code)
- ✅ No linting errors
- ✅ Code formatted correctly
- ✅ Build succeeds

**Note**: The coverage check will fail the build if thresholds are not met. This is intentional to maintain code quality.

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `style`: Code style changes (formatting)
- `chore`: Build process or auxiliary tool changes
- `perf`: Performance improvements

### Examples

```
feat(client): add retry logic for failed requests

Implement exponential backoff retry strategy for transient failures.
Retries up to 3 times with increasing delays.

Closes #123
```

```
test(servers): improve coverage for heartbeat methods

Add comprehensive tests for SendHeartbeat and UpdateHeartbeat.
Coverage increased from 75% to 92%.
```

```
fix(auth): handle expired token refresh correctly

Previously, expired tokens caused panic. Now properly refreshes
or returns appropriate error.

Fixes #456
```

## Questions or Issues?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the project's license.

---

Thank you for contributing to making the Nexmonyx Go SDK better! 🚀
