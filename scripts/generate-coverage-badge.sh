#!/bin/bash
# Coverage Badge Generator for Nexmonyx Go SDK
# Generates SVG badge showing current coverage percentage
# Usage: ./generate-coverage-badge.sh [coverage_file]

set -e

COVERAGE_FILE="${1:-coverage.out}"
BADGE_DIR="${2:-.coverage-badges}"
BADGE_FILE="$BADGE_DIR/coverage-badge.svg"

# Create badge directory
mkdir -p "$BADGE_DIR"

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Error: Coverage file not found: $COVERAGE_FILE"
    exit 1
fi

# Extract coverage percentage
COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

# Determine color based on coverage threshold
# Red: < 40%, Yellow: 40-79%, Green: >= 80%
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    COLOR="#4CAF50"  # Green
    STATUS="excellent"
elif (( $(echo "$COVERAGE >= 60" | bc -l) )); then
    COLOR="#FFC107"  # Yellow
    STATUS="good"
elif (( $(echo "$COVERAGE >= 40" | bc -l) )); then
    COLOR="#FF9800"  # Orange
    STATUS="acceptable"
else
    COLOR="#F44336"  # Red
    STATUS="poor"
fi

# Create SVG badge
cat > "$BADGE_FILE" <<EOF
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="142" height="20" role="img" aria-label="coverage: ${COVERAGE}%">
    <title>coverage: ${COVERAGE}%</title>
    <linearGradient id="s" x2="0" y2="100%">
        <stop offset="0" stop-color="#bbb"/>
        <stop offset="1" stop-color="#999"/>
    </linearGradient>
    <clipPath id="r">
        <rect width="142" height="20" rx="3" fill="#fff"/>
    </clipPath>
    <g clip-path="url(#r)">
        <rect width="107" height="20" fill="#555"/>
        <rect x="107" width="35" height="20" fill="$COLOR"/>
        <rect width="142" height="20" fill="url(#s)"/>
    </g>
    <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
        <text aria-hidden="true" x="545" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="970">coverage</text>
        <text x="545" y="140" transform="scale(.1)" fill="#fff" textLength="970">coverage</text>
        <text aria-hidden="true" x="1235" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="250">${COVERAGE}%</text>
        <text x="1235" y="140" transform="scale(.1)" fill="#fff" textLength="250">${COVERAGE}%</text>
    </g>
</svg>
EOF

echo "✅ Coverage badge generated: $BADGE_FILE"
echo "   Coverage: ${COVERAGE}% ($STATUS)"
echo ""
echo "Add to README.md:"
echo "[![Coverage Badge](.coverage-badges/coverage-badge.svg)](coverage_reports/latest.html)"
echo ""

# Also generate markdown file with badge
cat > "$BADGE_DIR/badge.md" <<'EOF'
# Coverage Badge

![Coverage](coverage-badge.svg)

**Generated:** $(date '+%Y-%m-%d %H:%M:%S')
**Coverage:** ${COVERAGE}%
**View Full Report:** [Coverage Reports](../coverage_reports/latest.html)
EOF

sed -i "s/\${COVERAGE}/$COVERAGE/g" "$BADGE_DIR/badge.md"
sed -i "s/\$(date.*/$(date '+%Y-%m-%d %H:%M:%S')/g" "$BADGE_DIR/badge.md"

echo "✅ Markdown badge reference created: $BADGE_DIR/badge.md"
echo ""

exit 0
