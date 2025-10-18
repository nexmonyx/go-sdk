# Coverage Automation Documentation

**Last Updated:** 2025-10-17
**Task:** #3012 - Automated Coverage Reporting

---

## Overview

The Nexmonyx Go SDK includes a comprehensive automated coverage reporting system that tracks, monitors, and reports on code coverage metrics. This system consists of multiple integrated tools and workflows.

## Components

### 1. Coverage Audit Script (`scripts/coverage_audit.sh`)

Comprehensive coverage analysis with detailed reporting.

**Features:**
- Runs test suite with coverage measurement
- Calculates service layer coverage statistics
- Validates coverage thresholds
- Generates HTML reports
- Creates detailed audit summaries
- Creates symlinks for easy access to latest reports

**Usage:**
```bash
# Run monthly coverage audit
./scripts/coverage_audit.sh

# Output directory: coverage_reports/
```

**Generated Artifacts:**
- `coverage_TIMESTAMP.out` - Raw coverage data
- `coverage_TIMESTAMP.html` - Interactive HTML report
- `coverage_detailed_TIMESTAMP.txt` - Detailed function coverage
- `test_output_TIMESTAMP.log` - Test execution log
- `audit_summary_TIMESTAMP.md` - Audit summary markdown
- `latest.html` - Symlink to latest HTML report
- `latest_detailed.txt` - Symlink to latest detailed report
- `latest_summary.md` - Symlink to latest audit summary

**Thresholds:**
- Package-wide: ≥ 40%
- Service layer: ≥ 80%

### 2. Coverage Badge Generator (`scripts/generate-coverage-badge.sh`)

Generates SVG badge showing current coverage percentage.

**Features:**
- Creates color-coded SVG badge
- Color scheme based on coverage thresholds:
  - Red: < 40% (poor)
  - Orange: 40-60% (acceptable)
  - Yellow: 60-80% (good)
  - Green: ≥ 80% (excellent)
- Generates markdown reference file
- Badge updates with each run

**Usage:**
```bash
# Generate badge from coverage file
./scripts/generate-coverage-badge.sh [coverage_file]

# Default: coverage.out
./scripts/generate-coverage-badge.sh

# Output directory: .coverage-badges/
```

**Usage in README:**
```markdown
[![Coverage Badge](.coverage-badges/coverage-badge.svg)](coverage_reports/latest.html)
```

**Generated Artifacts:**
- `.coverage-badges/coverage-badge.svg` - SVG badge
- `.coverage-badges/badge.md` - Markdown reference

### 3. Coverage History Tracker (`scripts/track-coverage-history.sh`)

Tracks coverage metrics over time and generates trend reports.

**Features:**
- Appends coverage metrics to CSV history file
- Calculates service layer coverage
- Detects coverage trends (improving/declining/stable)
- Generates trend analysis markdown
- Shows trend visualization

**Usage:**
```bash
# Track coverage history from coverage file
./scripts/track-coverage-history.sh [coverage_file]

# Default: coverage.out
./scripts/track-coverage-history.sh

# Output directory: coverage_reports/
```

**Generated Artifacts:**
- `coverage_history.csv` - CSV with historical data
- `coverage_trends.md` - Trend analysis report

**CSV Format:**
```csv
Date,Total Coverage %,Service Layer %,Commit Hash
2025-10-17,40.3,85.0,4276aa5
```

**Trend Detection:**
- 📈 Improving: Coverage increasing over time
- 📉 Declining: Coverage decreasing over time
- 📊 Stable: No significant change

### 4. Monthly Coverage Audit Workflow (`.github/workflows/coverage-audit.yml`)

Automated workflow that runs monthly coverage audits via GitHub Actions.

**Schedule:**
- Monthly: 2 AM UTC on the 1st of each month
- Trigger: Manual workflow dispatch available

**Workflow Steps:**
1. Run comprehensive coverage audit
2. Generate coverage badge
3. Track coverage history
4. Upload artifacts (90-day retention)
5. Create PR with updated badges/history (automatic)
6. Generate summary report

**Features:**
- Automatic badge generation
- History tracking with trend analysis
- Coverage metrics comparison
- PR creation for badge updates
- Artifact retention for 90 days
- Failure notifications

**Accessing Results:**
- View artifacts: Actions → Coverage Audit workflow → Artifacts
- HTML Report: `coverage_reports/coverage_*.html`
- Trends: `coverage_reports/coverage_trends.md`
- History: `coverage_reports/coverage_history.csv`

---

## Integration

### With CI/CD Pipeline

The existing CI pipeline (`.github/workflows/ci.yml`) already includes:
- Test execution with coverage (`-coverprofile=coverage.out`)
- Coverage threshold validation
- HTML report generation
- Codecov integration
- PR comment with coverage summary

**The new coverage automation extends this by:**
- Monthly comprehensive audits
- Long-term trend tracking
- Badge generation and updates
- Automated badge PRs

### Local Usage

**One-time audit:**
```bash
./scripts/coverage_audit.sh
```

**Generate badge:**
```bash
./scripts/generate-coverage-badge.sh
```

**Track metrics:**
```bash
./scripts/track-coverage-history.sh
```

**Combined workflow:**
```bash
# Run tests
go test -short -coverprofile=coverage.out ./...

# Run audit
./scripts/coverage_audit.sh

# Generate badge
./scripts/generate-coverage-badge.sh coverage_reports/coverage_*.out

# Track history
./scripts/track-coverage-history.sh coverage_reports/coverage_*.out
```

---

## Reports and Artifacts

### Coverage Report Directory

All reports are stored in `coverage_reports/`:

```
coverage_reports/
├── latest.html                          # Latest HTML report (symlink)
├── latest_detailed.txt                  # Latest detailed report (symlink)
├── latest_summary.md                    # Latest audit summary (symlink)
├── coverage_TIMESTAMP.out               # Raw coverage data
├── coverage_TIMESTAMP.html              # Interactive HTML report
├── coverage_detailed_TIMESTAMP.txt      # Function-level coverage
├── test_output_TIMESTAMP.log            # Test output log
├── audit_summary_TIMESTAMP.md           # Audit summary
├── coverage_history.csv                 # Historical data (cumulative)
└── coverage_trends.md                   # Trend analysis report
```

### Badge Directory

Badge artifacts in `.coverage-badges/`:

```
.coverage-badges/
├── coverage-badge.svg                   # Current coverage badge
└── badge.md                             # Markdown reference
```

### Viewing Reports

**HTML Report:**
```bash
# View in browser
open coverage_reports/latest.html        # macOS
xdg-open coverage_reports/latest.html    # Linux
```

**Markdown Reports:**
- Trends: `coverage_reports/coverage_trends.md`
- History: `coverage_reports/coverage_history.csv`
- Audit Summary: `coverage_reports/latest_summary.md`

---

## Coverage Metrics

### Service Layer Coverage (Target: ≥80%)

Tests all user-facing API methods:
- Organizations service
- Servers service
- Users service
- Monitoring service
- Alerts service
- Metrics service
- And all other core services

**Status:** ✅ **85% achieved** (exceeds target)

### Package-Wide Coverage (Target: ≥40%)

Includes all code in the SDK:
- Service implementations
- Models and types
- Helpers and utilities
- Client implementation

**Status:** ✅ **40.3% achieved** (meets target)

### Coverage Exemptions

The following are excluded from coverage requirements:
- Auto-generated getters/setters
- JSON marshaling/unmarshaling
- Simple helper functions (< 3 lines)
- Deprecated code
- Third-party library wrappers

See `TESTING.md` for complete exemption list and justification.

---

## Automation Workflow

### Monthly Audit Process

1. **Trigger** (1st of month at 2 AM UTC)
2. **Tests Run** - Full test suite with coverage
3. **Metrics Calculated** - Service layer and package coverage
4. **Badge Generated** - Color-coded SVG badge
5. **History Tracked** - Metrics added to CSV
6. **Trends Analyzed** - Trend report generated
7. **PR Created** - Automatic PR if changes detected
8. **Artifacts Uploaded** - Results available for 90 days

### Local Audit Process

```
go test → coverage_audit.sh → generate-coverage-badge.sh → track-coverage-history.sh
```

---

## Interpreting Results

### Coverage Badges

**Color Meanings:**
- 🟢 Green (≥80%): Excellent coverage, production-ready
- 🟡 Yellow (60-80%): Good coverage, acceptable for release
- 🟠 Orange (40-60%): Acceptable coverage, consider improvements
- 🔴 Red (<40%): Poor coverage, needs attention

### Trend Reports

**Trend Indicators:**
- 📈 **Improving**: Coverage increased from baseline
- 📉 **Declining**: Coverage decreased from baseline
- 📊 **Stable**: Coverage unchanged or minimal change

### History CSV

**Reading the CSV:**
```csv
Date,Total Coverage %,Service Layer %,Commit Hash
2025-10-17,40.3,85.0,4276aa5
```

- Date: Audit date (YYYY-MM-DD)
- Total Coverage: Package-wide coverage %
- Service Layer: Service API coverage %
- Commit: Git commit hash at time of audit

---

## Maintenance

### Update Badge

Badge is automatically generated during:
1. Monthly workflow runs
2. Manual badge script execution

To manually update:
```bash
go test -short -coverprofile=coverage.out ./...
./scripts/generate-coverage-badge.sh coverage.out
```

### Track New Metrics

History is automatically updated during:
1. Monthly workflow runs
2. Manual history script execution

To manually track:
```bash
./scripts/track-coverage-history.sh coverage.out
```

### Review Trends

View coverage trends:
```bash
cat coverage_reports/coverage_trends.md
```

View raw history:
```bash
cat coverage_reports/coverage_history.csv
```

---

## Troubleshooting

### Badge Not Generating

**Issue:** Script fails to generate badge

**Solution:**
1. Verify coverage file exists: `ls coverage.out`
2. Check script is executable: `chmod +x scripts/generate-coverage-badge.sh`
3. Run with explicit coverage file: `./scripts/generate-coverage-badge.sh coverage.out`

### History Not Tracking

**Issue:** History file empty or not updating

**Solution:**
1. Verify coverage file exists
2. Create directory: `mkdir -p coverage_reports`
3. Run tracker: `./scripts/track-coverage-history.sh coverage.out`

### CSV Parsing Issues

**Issue:** Trend report shows incorrect values

**Solution:**
1. Check CSV format: `head coverage_reports/coverage_history.csv`
2. Verify no special characters in commit hashes
3. Re-run tracker to generate fresh report

---

## References

- **Testing Guide:** `TESTING.md`
- **Coverage Exemptions:** `TESTING.md` → Coverage Exemptions section
- **Handler Testing:** `docs/HANDLER_TESTING_STANDARDS.md`
- **GitHub Actions:** `.github/workflows/coverage-audit.yml`

---

## Future Enhancements

Potential improvements for future versions:

1. **Coverage Reports Web UI**
   - Interactive dashboard showing trends
   - Historical comparisons
   - Service breakdown

2. **Slack Notifications**
   - Monthly audit summaries
   - Coverage drop alerts
   - Achievement celebrations

3. **Coverage Regression Detection**
   - Automatic alerts on coverage drops
   - Thresholds for blocking PRs
   - Minimum coverage enforcement

4. **Per-Service Coverage Tracking**
   - Individual service coverage trends
   - Service-level badges
   - Service comparison reports

5. **Performance Benchmarking**
   - Track test execution time
   - Detect performance regressions
   - Benchmark trend analysis

---

**Maintained by:** Nexmonyx Development Team
**Last Updated:** 2025-10-17
**Next Review:** 2025-12-17 (Quarterly)
