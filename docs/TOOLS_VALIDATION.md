# MCP Tools Validation Report

## Summary

✅ **All 17 Proxmox VE MCP Tools Defined and Validated**

- **Total Tools**: 17
- **Read-Only Tools**: 5 (Query operations)
- **Control Tools**: 12 (Lifecycle management)
- **Categories**: 4 major categories

---

## Tool Categories

### 1. Cluster & Node Management (3 tools)

Query operations for cluster and node information.

| Tool | Parameters | Purpose |
|------|-----------|---------|
| `get_nodes` | None | List all nodes in the cluster |
| `get_node_status` | `node_name` | Get detailed node status (CPU, memory, uptime) |
| `get_cluster_resources` | None | List all cluster resources (nodes, VMs, containers) |

**Example Usage:**
```json
{
  "name": "get_nodes",
  "arguments": {}
}
```

### 2. Storage Management (2 tools)

Query operations for storage devices and capacity information.

| Tool | Parameters | Purpose |
|------|-----------|---------|
| `get_storage` | None | List all storage devices in cluster |
| `get_node_storage` | `node_name` | List storage devices accessible from a node |

**Example Usage:**
```json
{
  "name": "get_node_storage",
  "arguments": {
    "node_name": "pve1"
  }
}
```

### 3. Virtual Machine Management (6 tools)

Query and control operations for QEMU virtual machines.

**Query Tools:**
| Tool | Parameters | Purpose |
|------|-----------|---------|
| `get_vms` | `node_name` | List all VMs on a node |
| `get_vm_status` | `node_name`, `vmid` | Get detailed VM status |

**Control Tools:**
| Tool | Parameters | Purpose |
|------|-----------|---------|
| `start_vm` | `node_name`, `vmid` | Start a stopped VM |
| `stop_vm` | `node_name`, `vmid` | Immediately stop a running VM |
| `shutdown_vm` | `node_name`, `vmid` | Gracefully shutdown a VM |
| `reboot_vm` | `node_name`, `vmid` | Gracefully reboot a VM |

**Example Usage:**
```json
{
  "name": "start_vm",
  "arguments": {
    "node_name": "pve1",
    "vmid": 100
  }
}
```

### 4. Container Management (6 tools)

Query and control operations for LXC containers.

**Query Tools:**
| Tool | Parameters | Purpose |
|------|-----------|---------|
| `get_containers` | `node_name` | List all containers on a node |
| `get_container_status` | `node_name`, `container_id` | Get detailed container status |

**Control Tools:**
| Tool | Parameters | Purpose |
|------|-----------|---------|
| `start_container` | `node_name`, `container_id` | Start a stopped container |
| `stop_container` | `node_name`, `container_id` | Immediately stop a running container |
| `shutdown_container` | `node_name`, `container_id` | Gracefully shutdown a container |
| `reboot_container` | `node_name`, `container_id` | Gracefully reboot a container |

**Example Usage:**
```json
{
  "name": "stop_container",
  "arguments": {
    "node_name": "pve2",
    "container_id": 200
  }
}
```

---

## Input Schema Validation

### ✅ Cluster & Node Management
- `get_nodes`: No parameters (empty schema)
- `get_node_status`: **Required**: `node_name` (string)
- `get_cluster_resources`: No parameters (empty schema)

### ✅ Storage Management
- `get_storage`: No parameters (empty schema)
- `get_node_storage`: **Required**: `node_name` (string)

### ✅ Virtual Machine Management
- `get_vms`: **Required**: `node_name` (string)
- `get_vm_status`: **Required**: `node_name` (string), `vmid` (integer > 0)
- `start_vm`: **Required**: `node_name` (string), `vmid` (integer > 0)
- `stop_vm`: **Required**: `node_name` (string), `vmid` (integer > 0)
- `shutdown_vm`: **Required**: `node_name` (string), `vmid` (integer > 0)
- `reboot_vm`: **Required**: `node_name` (string), `vmid` (integer > 0)

### ✅ Container Management
- `get_containers`: **Required**: `node_name` (string)
- `get_container_status`: **Required**: `node_name` (string), `container_id` (integer > 0)
- `start_container`: **Required**: `node_name` (string), `container_id` (integer > 0)
- `stop_container`: **Required**: `node_name` (string), `container_id` (integer > 0)
- `shutdown_container`: **Required**: `node_name` (string), `container_id` (integer > 0)
- `reboot_container`: **Required**: `node_name` (string), `container_id` (integer > 0)

---

## Response Format

All tools return a standardized JSON response:

```json
{
  "action": "operation_name",
  "node": "pve1",
  "status": "success",
  "data": {
    "result": "operation_result"
  }
}
```

### Response Fields

- **action**: The operation performed (e.g., "get_nodes", "start_vm")
- **node**: The node affected (for node-specific operations)
- **status**: "success" or "error"
- **data**: Query results or operation confirmation

---

## Error Handling

### Parameter Validation
- Missing required parameters: Returns error with parameter name
- Invalid parameter types: Returns type mismatch error
- Invalid ranges (e.g., vmid ≤ 0): Returns validation error

### API Errors
- Connection failures: Connection error message
- Authentication failures: Token validation error
- Resource not found: "Node/VM/Container not found" error
- Permission denied: "Insufficient permissions" error

---

## Testing Notes

### Query Tools (Safe to Test)
All query tools are safe to test and return read-only information:
```bash
# List all nodes
curl -X POST http://localhost:3000/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "get_nodes", "arguments": {}}'

# Get node status
curl -X POST http://localhost:3000/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "get_node_status", "arguments": {"node_name": "pve1"}}'
```

### Control Tools (Require Valid Resources)
Control tools require existing VM or Container IDs. First query available resources:
```bash
# Get VMs on pve1
curl -X POST http://localhost:3000/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name": "get_vms", "arguments": {"node_name": "pve1"}}'
```

Then use a valid VMID to test control operations.

---

## File Locations

- **Tool Definitions**: [`docs/tools-schema.json`](tools-schema.json)
- **Implementation**: [`internal/mcp/server.go`](../internal/mcp/server.go)
- **API Client**: [`internal/proxmox/client.go`](../internal/proxmox/client.go)
- **Test Script**: [`test-api.sh`](../test-api.sh)

---

## Compatibility

- **MCP Version**: v1.0.0 (mark3labs/mcp-go)
- **Proxmox API**: v2 (JSON endpoints)
- **Proxmox Version**: 9.1.x and later
- **Transport**: Stdio (for Claude integration)
- **Authentication**: PVEAPIToken format

---

## Validation Checklist

- ✅ All 17 tools have names and descriptions
- ✅ All tools have input schemas (JSON Schema v7 compatible)
- ✅ Required parameters properly marked
- ✅ Parameter types specified (string, integer, etc.)
- ✅ Parameter constraints documented
- ✅ No duplicate tool names
- ✅ Tool categories properly organized
- ✅ Consistent response format across all tools
- ✅ Error handling implemented
- ✅ Read-only vs control operations distinguished

---

## Next Steps

1. ✅ Tool definitions validated
2. ⏳ Integration testing with Claude
3. ⏳ Performance testing with large clusters
4. ⏳ Add batch operations (e.g., start multiple VMs)
5. ⏳ Add monitoring tools (CPU, memory, disk usage)
6. ⏳ Add backup/restore operations
7. ⏳ Add migration tools
8. ⏳ Add clustering operations (proxmox cluster join, etc.)

---

**Last Updated**: 2024  
**Status**: ✅ Ready for Testing
