#!/bin/bash
#
# Coverage Threshold Enforcement Script
#
# This script checks test coverage against thresholds defined in .coveragerc
# It's used in CI/CD to ensure coverage never regresses.
#
# Usage: ./scripts/check-coverage.sh [coverage.out]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
COVERAGE_FILE="${1:-coverage.out}"
CONFIG_FILE=".coveragerc"
EXIT_CODE=0

# Function to print colored output
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${BLUE}ℹ${NC} $1"; }

# Function to read config value
read_config() {
    local section=$1
    local key=$2
    local default=$3

    if [ -f "$CONFIG_FILE" ]; then
        # Try to read from config file (simple grep-based parser)
        value=$(grep -A 20 "^\[$section\]" "$CONFIG_FILE" 2>/dev/null | grep "^$key" | cut -d'=' -f2 | tr -d ' ' | head -1)
        if [ -n "$value" ]; then
            echo "$value"
            return
        fi
    fi

    echo "$default"
}

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    print_error "Coverage file not found: $COVERAGE_FILE"
    echo "Run: go test -coverprofile=coverage.out ./..."
    exit 1
fi

# Read thresholds from config
OVERALL_MIN=$(read_config "thresholds" "overall_minimum" "80.0")
PACKAGE_MIN=$(read_config "thresholds" "package_minimum" "70.0")
CRITICAL_MIN=$(read_config "thresholds" "critical_minimum" "90.0")
CRITICAL_PACKAGES=$(read_config "thresholds" "critical_packages" "client.go,errors.go,models.go")

print_info "Coverage Threshold Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
print_info "Thresholds:"
echo "  Overall minimum: ${OVERALL_MIN}%"
echo "  Package minimum: ${PACKAGE_MIN}%"
echo "  Critical files minimum: ${CRITICAL_MIN}%"
echo ""

# Check overall coverage
print_info "Checking overall coverage..."
OVERALL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

if [ -z "$OVERALL_COVERAGE" ]; then
    print_error "Could not parse overall coverage"
    exit 1
fi

echo "  Overall coverage: ${OVERALL_COVERAGE}%"

if (( $(echo "$OVERALL_COVERAGE < $OVERALL_MIN" | bc -l) )); then
    print_error "Overall coverage ${OVERALL_COVERAGE}% is below threshold ${OVERALL_MIN}%"
    EXIT_CODE=1
else
    print_success "Overall coverage ${OVERALL_COVERAGE}% meets threshold ${OVERALL_MIN}%"
fi
echo ""

# Check per-package coverage
print_info "Checking per-package coverage..."

# Get unique packages
PACKAGES=$(go tool cover -func="$COVERAGE_FILE" | awk '{print $1}' | grep -v "^total" | sed 's/:[0-9]*:.*$//' | sort -u)

FAILED_PACKAGES=()
PASSED_PACKAGES=0

for package in $PACKAGES; do
    # Calculate coverage for this package
    PACKAGE_LINES=$(go tool cover -func="$COVERAGE_FILE" | grep "^$package:" | wc -l)

    if [ "$PACKAGE_LINES" -eq 0 ]; then
        continue
    fi

    # Get coverage percentage for package
    PACKAGE_COV=$(go tool cover -func="$COVERAGE_FILE" | grep "^$package:" | awk '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}')

    if [ -z "$PACKAGE_COV" ] || [ "$PACKAGE_COV" == "0" ]; then
        continue
    fi

    # Check if this is a critical package
    IS_CRITICAL=false
    PACKAGE_NAME=$(basename "$package")
    if echo "$CRITICAL_PACKAGES" | grep -q "$PACKAGE_NAME"; then
        IS_CRITICAL=true
        THRESHOLD=$CRITICAL_MIN
    else
        THRESHOLD=$PACKAGE_MIN
    fi

    # Format package name for display
    DISPLAY_NAME=$(echo "$package" | sed 's|.*/go-sdk/||')

    if (( $(echo "$PACKAGE_COV < $THRESHOLD" | bc -l) )); then
        if [ "$IS_CRITICAL" = true ]; then
            print_error "Critical file $DISPLAY_NAME: ${PACKAGE_COV}% < ${THRESHOLD}% (critical)"
        else
            print_error "Package $DISPLAY_NAME: ${PACKAGE_COV}% < ${THRESHOLD}%"
        fi
        FAILED_PACKAGES+=("$DISPLAY_NAME")
        EXIT_CODE=1
    else
        PASSED_PACKAGES=$((PASSED_PACKAGES + 1))
    fi
done

if [ ${#FAILED_PACKAGES[@]} -eq 0 ]; then
    print_success "All packages meet coverage thresholds ($PASSED_PACKAGES packages checked)"
else
    echo ""
    print_error "${#FAILED_PACKAGES[@]} package(s) below threshold:"
    for pkg in "${FAILED_PACKAGES[@]}"; do
        echo "  - $pkg"
    done
fi
echo ""

# Check critical files specifically
print_info "Checking critical files coverage..."
IFS=',' read -ra CRITICAL_FILES <<< "$CRITICAL_PACKAGES"

CRITICAL_FAILED=()
for file in "${CRITICAL_FILES[@]}"; do
    file=$(echo "$file" | tr -d ' ')

    # Get coverage for this specific file
    FILE_COV=$(go tool cover -func="$COVERAGE_FILE" | grep "/$file:" | awk '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}')

    if [ -z "$FILE_COV" ] || [ "$FILE_COV" == "0" ]; then
        print_warning "Critical file $file: No coverage data found"
        continue
    fi

    if (( $(echo "$FILE_COV < $CRITICAL_MIN" | bc -l) )); then
        print_error "Critical file $file: ${FILE_COV}% < ${CRITICAL_MIN}%"
        CRITICAL_FAILED+=("$file")
        EXIT_CODE=1
    else
        print_success "Critical file $file: ${FILE_COV}% ≥ ${CRITICAL_MIN}%"
    fi
done
echo ""

# Final summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $EXIT_CODE -eq 0 ]; then
    print_success "Coverage check PASSED"
    echo ""
    echo "Summary:"
    echo "  Overall: ${OVERALL_COVERAGE}% (threshold: ${OVERALL_MIN}%)"
    echo "  Packages checked: $PASSED_PACKAGES"
    echo "  All thresholds met ✓"
else
    print_error "Coverage check FAILED"
    echo ""
    echo "Summary:"
    echo "  Overall: ${OVERALL_COVERAGE}% (threshold: ${OVERALL_MIN}%)"
    if [ ${#FAILED_PACKAGES[@]} -gt 0 ]; then
        echo "  Failed packages: ${#FAILED_PACKAGES[@]}"
    fi
    if [ ${#CRITICAL_FAILED[@]} -gt 0 ]; then
        echo "  Failed critical files: ${#CRITICAL_FAILED[@]}"
    fi
    echo ""
    print_error "Please improve test coverage to meet thresholds"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit $EXIT_CODE
