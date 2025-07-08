#!/bin/bash

# Test script to verify which header format the Nexmonyx API expects
# Usage: ./test_headers.sh <SERVER_UUID> <SERVER_SECRET> [API_URL]

SERVER_UUID="${1:-$NEXMONYX_SERVER_UUID}"
SERVER_SECRET="${2:-$NEXMONYX_SERVER_SECRET}"
API_URL="${3:-https://api-dev.nexmonyx.com}"

if [ -z "$SERVER_UUID" ] || [ -z "$SERVER_SECRET" ]; then
    echo "Usage: $0 <SERVER_UUID> <SERVER_SECRET> [API_URL]"
    echo "Or set NEXMONYX_SERVER_UUID and NEXMONYX_SERVER_SECRET environment variables"
    exit 1
fi

echo "Testing Nexmonyx API Authentication Headers"
echo "==========================================="
echo "Server UUID: $SERVER_UUID"
echo "API URL: $API_URL"
echo ""

# Test 1: Headers with X- prefix (current SDK format)
echo "Test 1: Headers WITH 'X-' prefix (current SDK format)"
echo "-----------------------------------------------------"
echo "curl -X POST \"$API_URL/v1/servers/heartbeat\" \\"
echo "  -H \"X-Server-UUID: $SERVER_UUID\" \\"
echo "  -H \"X-Server-Secret: [REDACTED]\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{}'"
echo ""
echo "Response:"
RESPONSE1=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_URL/v1/servers/heartbeat" \
  -H "X-Server-UUID: $SERVER_UUID" \
  -H "X-Server-Secret: $SERVER_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}' 2>&1)

HTTP_STATUS1=$(echo "$RESPONSE1" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY1=$(echo "$RESPONSE1" | sed '/HTTP_STATUS:/d')

echo "Status Code: $HTTP_STATUS1"
if [ -n "$BODY1" ]; then
    echo "Body: $BODY1"
fi
echo ""

# Test 2: Headers without X- prefix
echo "Test 2: Headers WITHOUT 'X-' prefix"
echo "-----------------------------------"
echo "curl -X POST \"$API_URL/v1/servers/heartbeat\" \\"
echo "  -H \"Server-UUID: $SERVER_UUID\" \\"
echo "  -H \"Server-Secret: [REDACTED]\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{}'"
echo ""
echo "Response:"
RESPONSE2=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_URL/v1/servers/heartbeat" \
  -H "Server-UUID: $SERVER_UUID" \
  -H "Server-Secret: $SERVER_SECRET" \
  -H "Content-Type: application/json" \
  -d '{}' 2>&1)

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
if [ "$HTTP_STATUS1" = "200" ] || [ "$HTTP_STATUS1" = "204" ]; then
    echo "✅ Headers WITH 'X-' prefix: SUCCESS"
else
    echo "❌ Headers WITH 'X-' prefix: FAILED (Status: $HTTP_STATUS1)"
fi

if [ "$HTTP_STATUS2" = "200" ] || [ "$HTTP_STATUS2" = "204" ]; then
    echo "✅ Headers WITHOUT 'X-' prefix: SUCCESS"
else
    echo "❌ Headers WITHOUT 'X-' prefix: FAILED (Status: $HTTP_STATUS2)"
fi

echo ""
echo "Conclusion:"
if ([ "$HTTP_STATUS1" = "200" ] || [ "$HTTP_STATUS1" = "204" ]) && ! ([ "$HTTP_STATUS2" = "200" ] || [ "$HTTP_STATUS2" = "204" ]); then
    echo "API expects headers WITH 'X-' prefix (current SDK format is correct)"
elif ! ([ "$HTTP_STATUS1" = "200" ] || [ "$HTTP_STATUS1" = "204" ]) && ([ "$HTTP_STATUS2" = "200" ] || [ "$HTTP_STATUS2" = "204" ]); then
    echo "API expects headers WITHOUT 'X-' prefix - SDK needs to be updated!"
elif ([ "$HTTP_STATUS1" = "200" ] || [ "$HTTP_STATUS1" = "204" ]) && ([ "$HTTP_STATUS2" = "200" ] || [ "$HTTP_STATUS2" = "204" ]); then
    echo "API accepts both header formats"
else
    echo "Neither header format worked - authentication issue may be elsewhere"
    echo ""
    echo "Possible issues:"
    echo "1. Invalid server credentials"
    echo "2. Server not registered or inactive"
    echo "3. API endpoint not accessible"
    echo "4. Different authentication mechanism required"
fi