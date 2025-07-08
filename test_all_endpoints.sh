#!/bin/bash

# Comprehensive test script for all agent endpoints
# Usage: ./test_all_endpoints.sh <SERVER_UUID> <SERVER_SECRET> [API_URL]

SERVER_UUID="${1:-$NEXMONYX_SERVER_UUID}"
SERVER_SECRET="${2:-$NEXMONYX_SERVER_SECRET}"
API_URL="${3:-https://api-dev.nexmonyx.com}"

if [ -z "$SERVER_UUID" ] || [ -z "$SERVER_SECRET" ]; then
    echo "Usage: $0 <SERVER_UUID> <SERVER_SECRET> [API_URL]"
    echo "Or set NEXMONYX_SERVER_UUID and NEXMONYX_SERVER_SECRET environment variables"
    exit 1
fi

echo "Testing All Nexmonyx Agent Endpoints"
echo "===================================="
echo "Server UUID: $SERVER_UUID"
echo "API URL: $API_URL"
echo ""

# Function to test an endpoint with both header formats
test_endpoint() {
    local METHOD=$1
    local ENDPOINT=$2
    local BODY=$3
    local DESC=$4
    
    echo "----------------------------------------"
    echo "$DESC"
    echo "Endpoint: $METHOD $ENDPOINT"
    echo ""
    
    # Test with X- prefix
    echo "Test 1: WITH X- prefix (X-Server-UUID, X-Server-Secret):"
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $METHOD "$API_URL$ENDPOINT" \
        -H "X-Server-UUID: $SERVER_UUID" \
        -H "X-Server-Secret: $SERVER_SECRET" \
        -H "Content-Type: application/json" \
        -d "$BODY" 2>&1)
    
    HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    BODY_RESP=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" = "200" ] || [ "$HTTP_STATUS" = "201" ] || [ "$HTTP_STATUS" = "204" ]; then
        echo "  ✅ SUCCESS (Status: $HTTP_STATUS)"
    else
        echo "  ❌ FAILED (Status: $HTTP_STATUS)"
        if [ -n "$BODY_RESP" ]; then
            echo "  Response: $BODY_RESP" | head -n 1
        fi
    fi
    
    # Test without X- prefix
    echo ""
    echo "Test 2: WITHOUT X- prefix (Server-UUID, Server-Secret):"
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $METHOD "$API_URL$ENDPOINT" \
        -H "Server-UUID: $SERVER_UUID" \
        -H "Server-Secret: $SERVER_SECRET" \
        -H "Content-Type: application/json" \
        -d "$BODY" 2>&1)
    
    HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    BODY_RESP=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" = "200" ] || [ "$HTTP_STATUS" = "201" ] || [ "$HTTP_STATUS" = "204" ]; then
        echo "  ✅ SUCCESS (Status: $HTTP_STATUS)"
    else
        echo "  ❌ FAILED (Status: $HTTP_STATUS)"
        if [ -n "$BODY_RESP" ]; then
            echo "  Response: $BODY_RESP" | head -n 1
        fi
    fi
    echo ""
}

# Test 1: Server Heartbeat
test_endpoint "POST" "/v1/servers/heartbeat" '{}' "1. Server Heartbeat"

# Test 2: Server Details Update
UPDATE_PAYLOAD='{"hostname":"test-server","os":"Linux","kernel":"6.14.0"}'
test_endpoint "PUT" "/v1/servers/$SERVER_UUID/details" "$UPDATE_PAYLOAD" "2. Update Server Details"

# Test 3: Metrics v1
METRICS_V1_PAYLOAD=$(cat <<EOF
{
  "server_uuid": "$SERVER_UUID",
  "collected_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "cpu": {"usage_percent": 45.5},
  "memory": {"total_bytes": 8589934592, "used_bytes": 4294967296}
}
EOF
)
test_endpoint "POST" "/v1/metrics/comprehensive" "$METRICS_V1_PAYLOAD" "3. Submit Metrics (v1)"

# Test 4: Metrics v2
METRICS_V2_PAYLOAD=$(cat <<EOF
{
  "server_uuid": "$SERVER_UUID",
  "collected_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "cpu": {"usage_percent": 45.5},
  "memory": {"total_bytes": 8589934592, "used_bytes": 4294967296}
}
EOF
)
test_endpoint "POST" "/v2/metrics/comprehensive" "$METRICS_V2_PAYLOAD" "4. Submit Metrics (v2)"

# Test 5: Hardware Inventory
HW_PAYLOAD=$(cat <<EOF
{
  "server_uuid": "$SERVER_UUID",
  "collected_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "system": {
    "manufacturer": "Dell Inc.",
    "product": "PowerEdge R740",
    "serial_number": "ABC123"
  }
}
EOF
)
test_endpoint "POST" "/v2/hardware/inventory" "$HW_PAYLOAD" "5. Submit Hardware Inventory"

# Summary
echo "========================================"
echo "SUMMARY"
echo "========================================"
echo ""
echo "Based on the test results above, you can determine:"
echo "1. Which header format each endpoint expects"
echo "2. Whether there's consistency across endpoints"
echo "3. What changes need to be made to the SDK"
echo ""
echo "If some endpoints work with X- and others without,"
echo "the API has inconsistent authentication handling that"
echo "needs to be addressed on the server side."