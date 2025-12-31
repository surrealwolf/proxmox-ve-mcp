#!/bin/bash
# MCP Tools Test Reference Guide
# This script demonstrates how to test Proxmox VE MCP tools

echo "=========================================="
echo "Proxmox VE MCP Tools - Testing Guide"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROXMOX_URL="${PROXMOX_BASE_URL:-https://pve1.dataknife.net:8006}"
API_TOKEN="${PROXMOX_API_TOKEN}"
SKIP_SSL="${PROXMOX_SKIP_SSL_VERIFY:-true}"

# Test if token is configured
if [ -z "$API_TOKEN" ]; then
    echo -e "${YELLOW}⚠ Warning: PROXMOX_API_TOKEN not set${NC}"
    echo "Please set environment variables or create .env file"
    echo ""
fi

echo "Configuration:"
echo "  Proxmox URL: $PROXMOX_URL"
echo "  Skip SSL Verify: $SKIP_SSL"
echo ""

# Helper function to make API calls
call_tool() {
    local tool_name="$1"
    local json_args="$2"
    
    echo -e "${BLUE}Testing: $tool_name${NC}"
    
    if [ -z "$json_args" ]; then
        json_args="{}"
    fi
    
    echo "Arguments: $json_args"
    echo ""
}

echo "=========================================="
echo "1. Query Tools (Read-Only)"
echo "=========================================="
echo ""

# Cluster & Node Management
echo -e "${GREEN}Cluster & Node Management:${NC}"
echo ""

call_tool "get_nodes" ""
echo "Description: List all nodes in the Proxmox cluster"
echo "Expected Response: Array of node objects with status"
echo ""

call_tool "get_node_status" '{"node_name": "pve1"}'
echo "Description: Get detailed node status (CPU, memory, uptime)"
echo "Expected Response: Node status with metrics"
echo "Parameters: node_name (required, string)"
echo ""

call_tool "get_cluster_resources" ""
echo "Description: List all cluster resources (nodes, VMs, containers)"
echo "Expected Response: Array of all cluster resources"
echo ""

# Storage Management
echo -e "${GREEN}Storage Management:${NC}"
echo ""

call_tool "get_storage" ""
echo "Description: List all storage devices in cluster"
echo "Expected Response: Array of storage devices with capacity"
echo ""

call_tool "get_node_storage" '{"node_name": "pve1"}'
echo "Description: Get storage devices accessible from a node"
echo "Expected Response: Storage devices visible from node"
echo "Parameters: node_name (required, string)"
echo ""

# VM Query Tools
echo -e "${GREEN}Virtual Machine Query:${NC}"
echo ""

call_tool "get_vms" '{"node_name": "pve1"}'
echo "Description: List all VMs on a specific node"
echo "Expected Response: Array of VM objects with basic info"
echo "Parameters: node_name (required, string)"
echo ""

call_tool "get_vm_status" '{"node_name": "pve1", "vmid": 100}'
echo "Description: Get detailed VM status and metrics"
echo "Expected Response: VM object with full status details"
echo "Parameters: node_name (required, string), vmid (required, integer)"
echo ""

# Container Query Tools
echo -e "${GREEN}Container Query:${NC}"
echo ""

call_tool "get_containers" '{"node_name": "pve1"}'
echo "Description: List all containers on a specific node"
echo "Expected Response: Array of container objects"
echo "Parameters: node_name (required, string)"
echo ""

call_tool "get_container_status" '{"node_name": "pve1", "container_id": 200}'
echo "Description: Get detailed container status and metrics"
echo "Expected Response: Container object with full status"
echo "Parameters: node_name (required, string), container_id (required, integer)"
echo ""

echo "=========================================="
echo "2. Control Tools (Requires Valid IDs)"
echo "=========================================="
echo ""

# VM Control Tools
echo -e "${GREEN}Virtual Machine Control:${NC}"
echo ""

call_tool "start_vm" '{"node_name": "pve1", "vmid": 100}'
echo "Description: Start a stopped VM"
echo "Parameters: node_name (required, string), vmid (required, integer > 0)"
echo ""

call_tool "stop_vm" '{"node_name": "pve1", "vmid": 100}'
echo "Description: Immediately stop a running VM (hard stop)"
echo "Parameters: node_name (required, string), vmid (required, integer > 0)"
echo ""

call_tool "shutdown_vm" '{"node_name": "pve1", "vmid": 100}'
echo "Description: Gracefully shutdown a VM (allows guest OS to shut down)"
echo "Parameters: node_name (required, string), vmid (required, integer > 0)"
echo ""

call_tool "reboot_vm" '{"node_name": "pve1", "vmid": 100}'
echo "Description: Gracefully reboot a VM"
echo "Parameters: node_name (required, string), vmid (required, integer > 0)"
echo ""

# Container Control Tools
echo -e "${GREEN}Container Control:${NC}"
echo ""

call_tool "start_container" '{"node_name": "pve1", "container_id": 200}'
echo "Description: Start a stopped container"
echo "Parameters: node_name (required, string), container_id (required, integer > 0)"
echo ""

call_tool "stop_container" '{"node_name": "pve1", "container_id": 200}'
echo "Description: Immediately stop a running container (hard stop)"
echo "Parameters: node_name (required, string), container_id (required, integer > 0)"
echo ""

call_tool "shutdown_container" '{"node_name": "pve1", "container_id": 200}'
echo "Description: Gracefully shutdown a container"
echo "Parameters: node_name (required, string), container_id (required, integer > 0)"
echo ""

call_tool "reboot_container" '{"node_name": "pve1", "container_id": 200}'
echo "Description: Gracefully reboot a container"
echo "Parameters: node_name (required, string), container_id (required, integer > 0)"
echo ""

echo "=========================================="
echo "3. Testing with curl"
echo "=========================================="
echo ""

echo -e "${YELLOW}Note: MCP uses stdio transport (not HTTP REST)${NC}"
echo ""
echo "To test via stdin (if server is running):"
echo ""

cat << 'EOF'
# Create MCP request JSON (tool_call message)
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_nodes",
    "arguments": {}
  }
}

# For tools with parameters:
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_vms",
    "arguments": {
      "node_name": "pve1"
    }
  }
}
EOF

echo ""
echo "=========================================="
echo "4. Test Sequence (Recommended)"
echo "=========================================="
echo ""

cat << 'EOF'
Step 1: Start the MCP server
  $ ./bin/proxmox-ve-mcp

Step 2: Query available nodes
  Tool: get_nodes
  Expected: See which nodes are available (e.g., pve1, pve2)

Step 3: Get node status
  Tool: get_node_status
  Arguments: {"node_name": "pve1"}
  Expected: CPU, memory, uptime information

Step 4: List VMs/Containers
  Tool: get_vms or get_containers
  Arguments: {"node_name": "pve1"}
  Expected: See if any VMs or containers exist

Step 5: Test control operations (if VMs/containers exist)
  Tool: start_vm, stop_vm, etc.
  Arguments: {"node_name": "pve1", "vmid": <valid_id>}
  Expected: Operation confirmation

Step 6: Query resource status after control
  Tool: get_vm_status or get_container_status
  Arguments: {"node_name": "pve1", "vmid": <id>}
  Expected: Updated status reflecting the action
EOF

echo ""
echo "=========================================="
echo "5. Error Handling"
echo "=========================================="
echo ""

cat << 'EOF'
Common Error Scenarios:

1. Missing Required Parameter:
   Error: "missing required parameter: node_name"
   Fix: Include all required parameters in arguments

2. Invalid Node Name:
   Error: "node not found: xyz"
   Fix: Use actual node names from get_nodes

3. Invalid VM/Container ID:
   Error: "vm not found: 999"
   Fix: Use actual IDs from get_vms or get_containers

4. Invalid Parameter Type:
   Error: "invalid type for vmid: expected integer"
   Fix: Use correct JSON type (integer not string for vmid)

5. Permission Denied:
   Error: "insufficient permissions"
   Fix: Check Proxmox token permissions
EOF

echo ""
echo "=========================================="
echo "For complete documentation, see:"
echo "  - TOOLS_VALIDATION.md"
echo "  - docs/tools-schema.json"
echo "=========================================="
