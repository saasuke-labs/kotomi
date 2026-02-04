#!/bin/bash
# Kotomi Smoke Test Script
# Tests critical paths to verify deployment health
#
# Usage: ./smoke_test.sh [BASE_URL]
# Example: ./smoke_test.sh http://localhost:8080
# Example: ./smoke_test.sh https://kotomi-prod-xyz.run.app

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${1:-http://localhost:8080}"
TEST_RESULTS=()
FAILED_TESTS=0
PASSED_TESTS=0

echo "=========================================="
echo "Kotomi Smoke Test Suite"
echo "=========================================="
echo "Target: $BASE_URL"
echo "Started: $(date)"
echo "=========================================="
echo ""

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    TEST_RESULTS+=("PASS: $1")
    ((PASSED_TESTS++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    TEST_RESULTS+=("FAIL: $1")
    ((FAILED_TESTS++))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local expected_status="$4"
    local extra_args="${5:-}"
    
    echo -n "Testing $name... "
    
    response=$(curl -s -w "\n%{http_code}" -X "$method" \
        $extra_args \
        "$BASE_URL$endpoint" 2>&1)
    
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        pass "$name (HTTP $status_code)"
    else
        fail "$name (Expected HTTP $expected_status, got $status_code)"
        echo "Response: $body" | head -c 200
        echo ""
    fi
}

# Test 1: Health Check
echo "1. Health Check Endpoint"
test_endpoint "Health Check" "GET" "/healthz" "200"
echo ""

# Test 2: API Endpoints (unauthenticated reads)
echo "2. Public API Endpoints"
test_endpoint "List Sites" "GET" "/api/sites" "200"
test_endpoint "List Sites (with query)" "GET" "/api/sites?limit=10" "200"
echo ""

# Test 3: Admin Panel Access
echo "3. Admin Panel"
test_endpoint "Admin Login Page" "GET" "/admin/login" "200"
test_endpoint "Admin Dashboard (should redirect)" "GET" "/admin/dashboard" "302"
echo ""

# Test 4: API Documentation
echo "4. API Documentation"
test_endpoint "Swagger JSON" "GET" "/docs/swagger.json" "200"
test_endpoint "Swagger UI" "GET" "/docs/index.html" "200"
echo ""

# Test 5: Static Assets
echo "5. Static Assets"
test_endpoint "Static CSS" "GET" "/static/css/admin.css" "200"
test_endpoint "Static JS" "GET" "/static/js/admin.js" "200"
echo ""

# Test 6: CORS Preflight
echo "6. CORS Configuration"
test_endpoint "CORS Preflight" "OPTIONS" "/api/sites" "200" \
    "-H 'Origin: http://example.com' -H 'Access-Control-Request-Method: GET'"
echo ""

# Test 7: Rate Limiting Headers
echo "7. Rate Limiting"
echo -n "Testing rate limit headers... "
response=$(curl -s -i "$BASE_URL/api/sites" 2>&1)
if echo "$response" | grep -q "X-RateLimit-Limit"; then
    pass "Rate limit headers present"
else
    warn "Rate limit headers not found (may not be enabled)"
fi
echo ""

# Test 8: Error Handling
echo "8. Error Handling"
test_endpoint "404 Not Found" "GET" "/nonexistent-path" "404"
test_endpoint "Invalid Site ID" "GET" "/api/sites/invalid-uuid" "400"
echo ""

# Test 9: Authentication Required Endpoints
echo "9. Authentication Required"
test_endpoint "Create Comment (no auth)" "POST" "/api/sites/00000000-0000-0000-0000-000000000000/pages/test/comments" "401" \
    "-H 'Content-Type: application/json' -d '{\"content\":\"test\"}'"
test_endpoint "Create Site (no auth)" "POST" "/api/sites" "401" \
    "-H 'Content-Type: application/json' -d '{\"name\":\"test\"}'"
echo ""

# Test 10: Method Not Allowed
echo "10. HTTP Method Validation"
test_endpoint "Invalid Method" "PATCH" "/api/sites" "405"
echo ""

# Summary
echo "=========================================="
echo "Test Results Summary"
echo "=========================================="
echo "Total Tests: $((PASSED_TESTS + FAILED_TESTS))"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
else
    echo "Failed: 0"
fi
echo "=========================================="
echo ""

# Print individual results
echo "Detailed Results:"
for result in "${TEST_RESULTS[@]}"; do
    if [[ $result == PASS:* ]]; then
        echo -e "${GREEN}✓${NC} ${result#PASS: }"
    else
        echo -e "${RED}✗${NC} ${result#FAIL: }"
    fi
done
echo ""

echo "Completed: $(date)"
echo "=========================================="

# Exit with error if any tests failed
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}SMOKE TESTS FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}ALL SMOKE TESTS PASSED${NC}"
    exit 0
fi
