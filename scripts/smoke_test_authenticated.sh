#!/bin/bash
# Kotomi Authenticated Smoke Test Script
# Tests complete user flow including JWT authentication
#
# Usage: ./smoke_test_authenticated.sh [BASE_URL] [JWT_SECRET]
# Example: ./smoke_test_authenticated.sh http://localhost:8080 "your-jwt-secret"

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${1:-http://localhost:8080}"
JWT_SECRET="${2:-test-secret-key-min-32-chars-long}"
TEST_RESULTS=()
FAILED_TESTS=0
PASSED_TESTS=0

# Test data
SITE_ID=""
PAGE_SLUG="test-page"
COMMENT_ID=""

echo "=========================================="
echo "Kotomi Authenticated Smoke Test Suite"
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

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Generate JWT token
generate_jwt() {
    local user_id="$1"
    local user_email="$2"
    local user_name="$3"
    
    # JWT Header
    header='{"alg":"HS256","typ":"JWT"}'
    
    # JWT Payload
    now=$(date +%s)
    exp=$((now + 3600))
    
    payload=$(cat <<EOF
{
  "sub": "$user_id",
  "iss": "kotomi-smoke-test",
  "aud": "kotomi-test",
  "iat": $now,
  "exp": $exp,
  "kotomi_user": {
    "id": "$user_id",
    "email": "$user_email",
    "name": "$user_name"
  }
}
EOF
)
    
    # Base64url encode
    header_b64=$(echo -n "$header" | openssl base64 -e -A | tr '+/' '-_' | tr -d '=')
    payload_b64=$(echo -n "$payload" | openssl base64 -e -A | tr '+/' '-_' | tr -d '=')
    
    # Create signature
    signature=$(echo -n "${header_b64}.${payload_b64}" | \
        openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | \
        openssl base64 -e -A | tr '+/' '-_' | tr -d '=')
    
    # Return complete JWT
    echo "${header_b64}.${payload_b64}.${signature}"
}

# Generate test JWT
info "Generating test JWT token..."
JWT_TOKEN=$(generate_jwt "test-user-123" "testuser@example.com" "Test User")
if [ -z "$JWT_TOKEN" ]; then
    fail "Failed to generate JWT token"
    exit 1
fi
info "JWT Token generated successfully"
echo ""

# Test 1: Create Site via Admin (simulated - would need Auth0 token)
echo "1. Site Management"
info "Note: Site creation requires Auth0 authentication (admin panel)"
info "For smoke test, we'll list existing sites and use the first one"

response=$(curl -s "$BASE_URL/api/sites")
SITE_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -n 1 | cut -d'"' -f4)

if [ -z "$SITE_ID" ]; then
    warn "No existing sites found. Some tests will be skipped."
    warn "Please create a site via admin panel first for full smoke test."
else
    pass "Found existing site: $SITE_ID"
fi
echo ""

# Test 2: Get Site Details
if [ -n "$SITE_ID" ]; then
    echo "2. Site Details"
    echo -n "Testing get site details... "
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/sites/$SITE_ID")
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ]; then
        pass "Get site details (HTTP 200)"
    else
        fail "Get site details (HTTP $status_code)"
    fi
    echo ""
fi

# Test 3: Post Comment with JWT
if [ -n "$SITE_ID" ]; then
    echo "3. Post Comment with Authentication"
    echo -n "Testing post comment... "
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"content":"This is a smoke test comment from authenticated user","page_url":"https://example.com/test"}' \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/comments")
    
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "201" ]; then
        COMMENT_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -n 1 | cut -d'"' -f4)
        pass "Post comment (HTTP 201)"
        info "Comment ID: $COMMENT_ID"
    elif [ "$status_code" = "401" ]; then
        warn "Comment creation failed - likely JWT config not set up"
        info "Please configure JWT auth for site: $SITE_ID"
    else
        fail "Post comment (HTTP $status_code)"
        echo "Response: $body" | head -c 200
    fi
    echo ""
fi

# Test 4: List Comments on Page
if [ -n "$SITE_ID" ]; then
    echo "4. List Comments"
    echo -n "Testing list comments... "
    
    response=$(curl -s -w "\n%{http_code}" \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/comments")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ]; then
        pass "List comments (HTTP 200)"
    else
        fail "List comments (HTTP $status_code)"
    fi
    echo ""
fi

# Test 5: Get Comment Details
if [ -n "$SITE_ID" ] && [ -n "$COMMENT_ID" ]; then
    echo "5. Get Comment Details"
    echo -n "Testing get comment... "
    
    response=$(curl -s -w "\n%{http_code}" \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/comments/$COMMENT_ID")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ]; then
        pass "Get comment details (HTTP 200)"
    else
        fail "Get comment details (HTTP $status_code)"
    fi
    echo ""
fi

# Test 6: Add Page Reaction
if [ -n "$SITE_ID" ]; then
    echo "6. Page Reactions"
    echo -n "Testing add page reaction... "
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"reaction_type":"👍"}' \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/reactions")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "201" ]; then
        pass "Add page reaction (HTTP 201)"
    elif [ "$status_code" = "401" ]; then
        warn "Reaction failed - JWT config may not be set up"
    else
        fail "Add page reaction (HTTP $status_code)"
    fi
    echo ""
fi

# Test 7: Get Page Reactions
if [ -n "$SITE_ID" ]; then
    echo "7. List Page Reactions"
    echo -n "Testing list page reactions... "
    
    response=$(curl -s -w "\n%{http_code}" \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/reactions")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ]; then
        pass "List page reactions (HTTP 200)"
    else
        fail "List page reactions (HTTP $status_code)"
    fi
    echo ""
fi

# Test 8: Remove Page Reaction
if [ -n "$SITE_ID" ]; then
    echo "8. Remove Page Reaction"
    echo -n "Testing remove page reaction... "
    
    response=$(curl -s -w "\n%{http_code}" \
        -X DELETE \
        -H "Authorization: Bearer $JWT_TOKEN" \
        "$BASE_URL/api/sites/$SITE_ID/pages/$PAGE_SLUG/reactions/%F0%9F%91%8D")
    
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ] || [ "$status_code" = "204" ]; then
        pass "Remove page reaction (HTTP $status_code)"
    elif [ "$status_code" = "401" ]; then
        warn "Remove reaction failed - JWT config may not be set up"
    else
        fail "Remove page reaction (HTTP $status_code)"
    fi
    echo ""
fi

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

# Additional notes
if [ -z "$SITE_ID" ]; then
    echo ""
    echo "=========================================="
    echo "Setup Required for Full Test Coverage"
    echo "=========================================="
    echo "1. Access admin panel: $BASE_URL/admin/dashboard"
    echo "2. Create a site via the admin panel"
    echo "3. Configure JWT authentication for the site"
    echo "4. Re-run this smoke test"
    echo "=========================================="
fi

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
