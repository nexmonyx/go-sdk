#!/bin/bash

# Test script to verify metrics submission endpoint authentication
# Usage: ./test_metrics.sh <SERVER_UUID> <SERVER_SECRET> [API_URL]

SERVER_UUID="${1:-$NEXMONYX_SERVER_UUID}"
SERVER_SECRET="${2:-$NEXMONYX_SERVER_SECRET}"
API_URL="${3:-https://api-dev.nexmonyx.com}"

if [ -z "$SERVER_UUID" ] || [ -z "$SERVER_SECRET" ]; then
    echo "Usage: $0 <SERVER_UUID> <SERVER_SECRET> [API_URL]"
    echo "Or set NEXMONYX_SERVER_UUID and NEXMONYX_SERVER_SECRET environment variables"
    exit 1
fi

echo "Testing Nexmonyx Metrics API Authentication"
echo "==========================================="
echo "Server UUID: $SERVER_UUID"
echo "API URL: $API_URL"
echo ""

# Minimal metrics payload
METRICS_PAYLOAD=$(cat <<EOF
{
  "server_uuid": "$SERVER_UUID",
  "collected_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "cpu": {
    "usage_percent": 45.5
  },
  "memory": {
    "total_bytes": 8589934592,
    "used_bytes": 4294967296
  }
}
EOF
)

# Test 1: Headers with X- prefix (current SDK format)
echo "Test 1: POST /v2/metrics/comprehensive with X- headers"
echo "------------------------------------------------------"
echo "Headers: X-Server-UUID, X-Server-Secret"
echo ""
echo "Response:"
RESPONSE1=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_URL/v2/metrics/comprehensive" \
  -H "X-Server-UUID: $SERVER_UUID" \
  -H "X-Server-Secret: $SERVER_SECRET" \
  -H "Content-Type: application/json" \
  -d "$METRICS_PAYLOAD" 2>&1)

HTTP_STATUS1=$(echo "$RESPONSE1" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY1=$(echo "$RESPONSE1" | sed '/HTTP_STATUS:/d')

echo "Status Code: $HTTP_STATUS1"
if [ -n "$BODY1" ]; then
    echo "Body: $BODY1"
fi
echo ""

# Test 2: Headers without X- prefix
echo "Test 2: POST /v2/metrics/comprehensive without X- prefix"
echo "--------------------------------------------------------"
echo "Headers: Server-UUID, Server-Secret"
echo ""
echo "Response:"
RESPONSE2=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_URL/v2/metrics/comprehensive" \
  -H "Server-UUID: $SERVER_UUID" \
  -H "Server-Secret: $SERVER_SECRET" \
  -H "Content-Type: application/json" \
  -d "$METRICS_PAYLOAD" 2>&1)

HTTP_STATUS2=$(echo "$RESPONSE2" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY2=$(echo "$RESPONSE2" | sed '/HTTP_STATUS:/d')

echo "Status Code: $HTTP_STATUS2"
if [ -n "$BODY2" ]; then
    echo "Body: $BODY2"
fi
echo ""

# Summary
echo "Summary"
echo "======="
if [ "$HTTP_STATUS1" = "200" ] || [ "$HTTP_STATUS1" = "201" ] || [ "$HTTP_STATUS1" = "204" ]; then
    echo "✅ Headers WITH 'X-' prefix: SUCCESS"
else
    echo "❌ Headers WITH 'X-' prefix: FAILED (Status: $HTTP_STATUS1)"
    if [[ "$BODY1" == *"Missing Server-UUID"* ]]; then
        echo "   ⚠️  Error mentions missing headers WITHOUT X- prefix!"
    fi
fi

if [ "$HTTP_STATUS2" = "200" ] || [ "$HTTP_STATUS2" = "201" ] || [ "$HTTP_STATUS2" = "204" ]; then
    echo "✅ Headers WITHOUT 'X-' prefix: SUCCESS"
else
    echo "❌ Headers WITHOUT 'X-' prefix: FAILED (Status: $HTTP_STATUS2)"
fi

echo ""
echo "Diagnosis:"
if [[ "$BODY1" == *"Missing Server-UUID or Server-Secret headers"* ]]; then
    echo "🔍 The API error message suggests it expects headers WITHOUT the X- prefix,"
    echo "   but the SDK is sending them WITH the X- prefix."
    echo ""
    echo "RECOMMENDED FIX: Update the SDK to use 'Server-UUID' and 'Server-Secret'"
    echo "                 instead of 'X-Server-UUID' and 'X-Server-Secret'"
fi