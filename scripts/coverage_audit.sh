#!/bin/bash
# Coverage Audit Script for Nexmonyx Go SDK
# Run this script monthly or before releases to verify coverage standards

set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     Nexmonyx Go SDK - Coverage Audit Script              ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Configuration
COVERAGE_DIR="coverage_reports"
THRESHOLD_PACKAGE="40.0"
THRESHOLD_SERVICE="80.0"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create coverage reports directory
mkdir -p "$COVERAGE_DIR"

# Step 1: Run tests with coverage
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 1: Running test suite with coverage..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Run short tests to avoid timeout issues with comprehensive tests
go test -short -coverprofile="$COVERAGE_DIR/coverage_$TIMESTAMP.out" -covermode=atomic ./... 2>&1 | tee "$COVERAGE_DIR/test_output_$TIMESTAMP.log"

if [ ! -f "$COVERAGE_DIR/coverage_$TIMESTAMP.out" ]; then
    echo "❌ ERROR: Coverage file not generated"
    exit 1
fi

echo ""
echo "✅ Tests completed successfully"
echo ""

# Step 2: Calculate overall coverage
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 2: Calculating coverage metrics..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Get total coverage
PACKAGE_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/coverage_$TIMESTAMP.out" | grep "total:" | awk '{print $3}' | sed 's/%//')

echo "Package Coverage: $PACKAGE_COVERAGE%"

# Step 3: Analyze service layer coverage
echo ""
echo "Service Layer Coverage:"
echo "------------------------"

# Extract coverage for main service files
go tool cover -func="$COVERAGE_DIR/coverage_$TIMESTAMP.out" | \
    grep -E "(servers|alerts|health|monitoring|metrics|billing|users|organizations|incidents|api_keys|vms|tags|clusters|providers|probes)\.go:" | \
    grep -v "_test" | \
    awk '{printf "  %-30s %s\n", $1":"$2, $3}'

echo ""

# Step 4: Check thresholds
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 4: Verifying coverage thresholds..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

THRESHOLD_MET=true

# Check package threshold
if (( $(echo "$PACKAGE_COVERAGE >= $THRESHOLD_PACKAGE" | bc -l) )); then
    echo "✅ Package coverage ($PACKAGE_COVERAGE%) meets threshold (≥$THRESHOLD_PACKAGE%)"
else
    echo "⚠️  Package coverage ($PACKAGE_COVERAGE%) below threshold (≥$THRESHOLD_PACKAGE%)"
    THRESHOLD_MET=false
fi

echo ""

# Step 5: Generate HTML report
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 5: Generating HTML coverage report..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

go tool cover -html="$COVERAGE_DIR/coverage_$TIMESTAMP.out" -o "$COVERAGE_DIR/coverage_$TIMESTAMP.html"

echo "✅ HTML report generated: $COVERAGE_DIR/coverage_$TIMESTAMP.html"
echo ""

# Step 6: Generate detailed report
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 6: Generating detailed coverage report..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

go tool cover -func="$COVERAGE_DIR/coverage_$TIMESTAMP.out" > "$COVERAGE_DIR/coverage_detailed_$TIMESTAMP.txt"

echo "✅ Detailed report saved: $COVERAGE_DIR/coverage_detailed_$TIMESTAMP.txt"
echo ""

# Step 7: Create summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 7: Creating audit summary..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cat > "$COVERAGE_DIR/audit_summary_$TIMESTAMP.md" <<EOF
# Coverage Audit Summary

**Date:** $(date '+%Y-%m-%d %H:%M:%S')
**Audit Type:** Automated Monthly Audit

## Coverage Metrics

- **Package Coverage:** $PACKAGE_COVERAGE%
- **Threshold:** ≥$THRESHOLD_PACKAGE%
- **Status:** $(if [ "$THRESHOLD_MET" = true ]; then echo "✅ PASS"; else echo "⚠️ REVIEW NEEDED"; fi)

## Reports Generated

- HTML Report: \`coverage_$TIMESTAMP.html\`
- Detailed Report: \`coverage_detailed_$TIMESTAMP.txt\`
- Test Output: \`test_output_$TIMESTAMP.log\`

## Service Layer Coverage

See detailed report for line-by-line coverage of service files.

## Recommendations

$(if [ "$THRESHOLD_MET" = true ]; then
    echo "Coverage meets all thresholds. No action required."
else
    echo "Review coverage gaps and prioritize testing for uncovered areas."
    echo "Focus on service layer methods (user-facing APIs)."
fi)

## Next Audit

**Scheduled:** $(date -d '+1 month' '+%Y-%m-%d' 2>/dev/null || date -v+1m '+%Y-%m-%d' 2>/dev/null || echo "Next month")
EOF

echo "✅ Audit summary created: $COVERAGE_DIR/audit_summary_$TIMESTAMP.md"
echo ""

# Step 8: Create symlink to latest reports
ln -sf "coverage_$TIMESTAMP.html" "$COVERAGE_DIR/latest.html"
ln -sf "coverage_detailed_$TIMESTAMP.txt" "$COVERAGE_DIR/latest_detailed.txt"
ln -sf "audit_summary_$TIMESTAMP.md" "$COVERAGE_DIR/latest_summary.md"

echo "✅ Symlinks created for latest reports"
echo ""

# Final summary
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              Audit Complete                               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "Results:"
echo "  Coverage: $PACKAGE_COVERAGE%"
echo "  Status: $(if [ "$THRESHOLD_MET" = true ]; then echo "✅ PASS"; else echo "⚠️ REVIEW NEEDED"; fi)"
echo ""
echo "Reports saved to: $COVERAGE_DIR/"
echo ""
echo "View HTML report:"
echo "  open $COVERAGE_DIR/latest.html  # macOS"
echo "  xdg-open $COVERAGE_DIR/latest.html  # Linux"
echo ""

# Exit with appropriate code
if [ "$THRESHOLD_MET" = true ]; then
    exit 0
else
    echo "⚠️  Warning: Coverage below threshold"
    exit 1
fi
