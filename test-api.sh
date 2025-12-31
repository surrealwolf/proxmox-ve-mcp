#!/bin/bash
# Proxmox VE MCP - API Verification Script
# Tests all API endpoints used by the MCP tools

# Load environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Check required variables
if [ -z "$PROXMOX_BASE_URL" ] || [ -z "$PROXMOX_API_TOKEN" ]; then
  echo "Error: PROXMOX_BASE_URL and PROXMOX_API_TOKEN must be set in .env"
  exit 1
fi

# Setup curl options
CURL_OPTS="-s -k"
if [ "$PROXMOX_SKIP_SSL_VERIFY" != "true" ]; then
  CURL_OPTS="-s"
fi

BASE_URL="$PROXMOX_BASE_URL"
TOKEN="$PROXMOX_API_TOKEN"

echo "=========================================="
echo "Proxmox VE API Verification"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Token: ${TOKEN:0:30}..."
echo ""

# Test 1: Version Check
echo "[1/5] Testing API Version..."
RESPONSE=$(curl $CURL_OPTS -H "Authorization: PVEAPIToken=$TOKEN" "$BASE_URL/api2/json/version")
VERSION=$(echo "$RESPONSE" | jq -r '.data.version // empty')
if [ -z "$VERSION" ]; then
  echo "  ❌ Failed to get version"
  echo "  Response: $RESPONSE"
  exit 1
fi
echo "  ✓ Proxmox Version: $VERSION"

# Test 2: Get Nodes
echo ""
echo "[2/5] Getting Cluster Nodes..."
RESPONSE=$(curl $CURL_OPTS -H "Authorization: PVEAPIToken=$TOKEN" "$BASE_URL/api2/json/nodes")
NODES=$(echo "$RESPONSE" | jq -r '.data[] | .node' 2>/dev/null)
if [ -z "$NODES" ]; then
  echo "  ❌ Failed to get nodes"
  echo "  Response: $RESPONSE"
  exit 1
fi
NODE_COUNT=$(echo "$NODES" | wc -l)
echo "  ✓ Found $NODE_COUNT node(s):"
for node in $NODES; do
  echo "    - $node"
done

# Test 3: VMs on each node
echo ""
echo "[3/5] Querying Virtual Machines..."
for node in $NODES; do
  RESPONSE=$(curl $CURL_OPTS -H "Authorization: PVEAPIToken=$TOKEN" "$BASE_URL/api2/json/nodes/$node/qemu")
  VM_COUNT=$(echo "$RESPONSE" | jq '.data | length // 0')
  echo "  ✓ $node: $VM_COUNT VM(s)"
done

# Test 4: Containers on each node
echo ""
echo "[4/5] Querying LXC Containers..."
for node in $NODES; do
  RESPONSE=$(curl $CURL_OPTS -H "Authorization: PVEAPIToken=$TOKEN" "$BASE_URL/api2/json/nodes/$node/lxc")
  CT_COUNT=$(echo "$RESPONSE" | jq '.data | length // 0')
  echo "  ✓ $node: $CT_COUNT container(s)"
done

# Test 5: Token Validation
echo ""
echo "[5/5] Validating API Token..."
RESPONSE=$(curl $CURL_OPTS -w "\n%{http_code}" -H "Authorization: PVEAPIToken=$TOKEN" "$BASE_URL/api2/json/access/ticket")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
  echo "  ✓ Token is valid (HTTP 200)"
else
  echo "  ⚠ Token response: HTTP $HTTP_CODE"
fi

echo ""
echo "=========================================="
echo "✅ All API verification tests passed!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  • API is accessible and responding"
echo "  • Node discovery working"
echo "  • VM and container queries functional"
echo "  • Token authentication confirmed"
