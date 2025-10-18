#!/bin/bash
# Coverage History Tracking for Nexmonyx Go SDK
# Tracks coverage metrics over time and generates trend reports
# Usage: ./track-coverage-history.sh [coverage_file]

set -e

COVERAGE_FILE="${1:-coverage.out}"
HISTORY_FILE="coverage_reports/coverage_history.csv"
HISTORY_DIR="coverage_reports"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
DATE_SHORT=$(date '+%Y-%m-%d')

# Create history directory
mkdir -p "$HISTORY_DIR"

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Error: Coverage file not found: $COVERAGE_FILE"
    exit 1
fi

# Extract coverage metrics
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

# Extract service layer coverage (approximate)
SERVICE_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | \
    grep -E "(servers|alerts|health|monitoring|metrics|billing|users|organizations|incidents|api_keys|vms|tags|clusters|providers|probes)\.go:" | \
    grep -v "_test" | \
    awk '{sum += $NF; count++} END {if (count > 0) printf "%.1f", sum/count; else print "N/A"}')

# Initialize CSV with header if it doesn't exist
if [ ! -f "$HISTORY_FILE" ]; then
    cat > "$HISTORY_FILE" <<EOF
Date,Total Coverage %,Service Layer %,Commit Hash
EOF
fi

# Get current commit hash
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "N/A")

# Append current metrics
echo "$DATE_SHORT,$TOTAL_COVERAGE,$SERVICE_COVERAGE,$COMMIT_HASH" >> "$HISTORY_FILE"

echo "✅ Coverage history updated"
echo "   Total Coverage: ${TOTAL_COVERAGE}%"
echo "   Service Layer: ${SERVICE_COVERAGE}%"
echo "   Timestamp: $TIMESTAMP"
echo "   Commit: $COMMIT_HASH"
echo ""

# Generate trend report
echo "📊 Coverage Trend (Last 10 entries):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
tail -11 "$HISTORY_FILE" | head -10 | awk -F',' '{
    printf "%-12s | Total: %6s | Service: %6s | %s\n", $1, $2"%", $3"%", $4
}'
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Generate markdown trend report
HISTORY_COUNT=$(tail -n +2 "$HISTORY_FILE" | wc -l)
FIRST_COVERAGE=$(tail -n +2 "$HISTORY_FILE" | head -1 | cut -d',' -f2)
LATEST_COVERAGE=$(tail -1 "$HISTORY_FILE" | cut -d',' -f2)

if [ "$FIRST_COVERAGE" != "$LATEST_COVERAGE" ]; then
    COVERAGE_CHANGE=$(echo "$LATEST_COVERAGE - $FIRST_COVERAGE" | bc)
    if (( $(echo "$COVERAGE_CHANGE > 0" | bc -l) )); then
        TREND="📈 Improving"
        DIRECTION="+"
    else
        TREND="📉 Declining"
        DIRECTION=""
    fi
    TREND_LINE="**Trend:** $TREND ($DIRECTION$COVERAGE_CHANGE%) over $HISTORY_COUNT measurements"
else
    TREND_LINE="**Trend:** 📊 Stable at ${LATEST_COVERAGE}%"
fi

# Create trend report markdown
cat > "$HISTORY_DIR/coverage_trends.md" <<EOF
# Coverage Trend Report

**Generated:** $TIMESTAMP

## Summary

- **Current Coverage:** ${TOTAL_COVERAGE}%
- **Service Layer:** ${SERVICE_COVERAGE}%
- $TREND_LINE

## Recent History

| Date | Total Coverage | Service Layer | Commit |
|------|----------------|---------------|--------|
EOF

# Add last 10 entries to report (skip header)
tail -n +2 "$HISTORY_FILE" | tail -10 | while read line; do
    if [ ! -z "$line" ]; then
        IFS=',' read -r date total service commit <<< "$line"
        echo "| $date | ${total}% | ${service}% | \`$commit\` |" >> "$HISTORY_DIR/coverage_trends.md"
    fi
done

cat >> "$HISTORY_DIR/coverage_trends.md" <<'EOF'

## Thresholds

- ✅ **Service Layer Target:** ≥80% (EXCELLENT)
- ✅ **Package Target:** ≥40% (GOOD)

## Recommendations

- Monitor trend for significant changes
- Investigate sudden drops in coverage
- Celebrate improvements!
- Run `./scripts/coverage_audit.sh` for full details

## History File

Raw data available in `coverage_history.csv`
EOF

echo "✅ Trend report generated: $HISTORY_DIR/coverage_trends.md"
echo ""

exit 0
