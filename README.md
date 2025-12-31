# Proxmox VE MCP

Model Context Protocol (MCP) server for Proxmox Virtual Environment infrastructure management. Control and monitor your Proxmox infrastructure through an AI-powered interface.

**Focused on:** Comprehensive Proxmox infrastructure management including cluster operations, VM/container lifecycle, user access control, and backup/restore operations.

⚠️ **Early Development Warning**: This project is in early development and may contain bugs that could cause unexpected behavior. Use with caution in production environments.

⚠️ **Prompt Injection Risk**: You are responsible for guarding against prompt injection when using these tools. Exercise extreme caution or use MCP tools only on systems and data you trust.

## Features

- **69 management tools** across 6 operational categories
- **User & Access Management**: 16 tools for users, groups, roles, and ACLs
- **Backup & Restore Operations**: 6 tools for VM/container backup creation, management, and restoration
- **VM Creation & Cloning**: 4 tools for creating, cloning, and configuring virtual machines
- **VM Snapshots & Backups**: 5 tools for creating, listing, restoring, and deleting VM snapshots
- **VM Migration**: Tools for live and offline VM migration to other nodes
- **Container Creation & Cloning**: 6 tools for container management and lifecycle
- **Advanced Cluster Management**: 6 tools for detailed cluster and status operations
- **VM Configuration Management**: Update VM configs, mark as template, manage settings
- **Container Configuration Management**: Update container configs, manage settings
- **Cluster Management**: Monitor cluster health and node status
- **Virtual Machine Management**: List, monitor, and manage VMs
- **Container Management**: Manage LXC containers
- **Node Monitoring**: Track resource usage and uptime
- **Stdio Transport**: MCP protocol over standard input/output for seamless integration
- **HTTP Transport**: Optional HTTP API for remote connections and integration

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/surrealwolf/proxmox-ve-mcp.git
cd proxmox-ve-mcp
go build -o bin/proxmox-ve-mcp ./cmd
```

### Configuration

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Then edit `.env` with your Proxmox credentials:

```bash
PROXMOX_BASE_URL=https://your-proxmox-server.com:8006
PROXMOX_API_USER=root@pam
PROXMOX_API_TOKEN_ID=proxmox_mcp_token
PROXMOX_API_TOKEN_SECRET=your-token-secret-here
PROXMOX_SKIP_SSL_VERIFY=false
LOG_LEVEL=info
```

### Obtaining API Token

1. Log in to Proxmox Web UI
2. Go to Datacenter → Permissions → API Tokens
3. Create a new API token with appropriate permissions
4. Note the username (e.g., `root@pam`) for `PROXMOX_API_USER`
5. Note the token ID (e.g., `proxmox_mcp_token`) for `PROXMOX_API_TOKEN_ID`
6. Note the token secret (the generated password) for `PROXMOX_API_TOKEN_SECRET`
   - The full token is combined as: `user@realm!tokenid=secret`
   - `PROXMOX_API_USER`: The user part (e.g., `root@pam`)
   - `PROXMOX_API_TOKEN_ID`: The token ID part (e.g., `proxmox_mcp_token`)
   - `PROXMOX_API_TOKEN_SECRET`: The secret part only (no special characters)

### Running the Server

**Stdio Transport (Default):**
```bash
./bin/proxmox-ve-mcp
```

**HTTP Transport:**
```bash
MCP_TRANSPORT=http MCP_HTTP_ADDR=:8000 ./bin/proxmox-ve-mcp
```

Then access the endpoints:
```bash
# Health check
curl http://localhost:8000/health

# MCP endpoint
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -d '{"method": "tools/list"}'
```

**Environment Variables:**
- `MCP_TRANSPORT`: Set to `"http"` for HTTP transport (default: `"stdio"`)
- `MCP_HTTP_ADDR`: HTTP server address (default: `:8000`)

## Available Tools (69 Total)

### User & Access Management (16 tools)
- `list_users` - List all users in the system
- `get_user` - Get detailed information about a specific user
- `create_user` - Create a new user account
- `update_user` - Update user properties (email, comment, groups)
- `delete_user` - Remove a user from the system
- `change_password` - Change a user's password
- `list_groups` - List all groups in the system
- `create_group` - Create a new group
- `delete_group` - Remove a group from the system
- `list_roles` - List all roles and their privileges
- `create_role` - Create a custom role with specific privileges
- `delete_role` - Remove a custom role
- `list_acl` - List all ACL entries (path-based permissions)
- `set_acl` - Configure access control for a specific path
- `create_api_token` - Create a new API token for authentication
- `delete_api_token` - Revoke an API token

### Backup & Restore Operations (6 tools)
- `list_backups` - List all available backups
- `create_vm_backup` - Create a backup of a virtual machine
- `create_container_backup` - Create a backup of an LXC container
- `delete_backup` - Remove a backup
- `restore_vm_backup` - Restore a VM from backup
- `restore_container_backup` - Restore a container from backup

### Cluster & Node Management (6 tools)
- `get_nodes` - List all nodes in the Proxmox cluster
- `get_node_status` - Get detailed status for a specific node
- `get_cluster_resources` - Get overview of cluster resources (nodes, VMs, containers, storage)
- `get_cluster_status` - Get cluster-wide status information
- `get_storage` - List all storage devices in the cluster
- `get_node_storage` - Get storage devices for a specific node

### Virtual Machine Management (21 tools)
- `get_vms` - List all VMs on a specific node
- `get_vm_status` - Get detailed VM information and status
- `get_vm_config` - Get full configuration of a virtual machine
- `start_vm` - Power on a virtual machine
- `stop_vm` - Power off a virtual machine
- `shutdown_vm` - Gracefully shutdown a virtual machine
- `reboot_vm` - Reboot a virtual machine
- `delete_vm` - Delete a virtual machine (with optional force)
- `suspend_vm` - Suspend (pause) a virtual machine
- `resume_vm` - Resume a suspended virtual machine
- `create_vm` - Create a new virtual machine with basic configuration
- `create_vm_advanced` - Create a VM with advanced configuration options
- `clone_vm` - Clone an existing virtual machine
- `update_vm_config` - Update VM configuration (mark as template, adjust resources, etc.)
- `get_vm_console` - Get console access information for a VM
- `create_vm_snapshot` - Create a snapshot of a virtual machine
- `list_vm_snapshots` - List all snapshots for a virtual machine
- `delete_vm_snapshot` - Delete a snapshot from a virtual machine
- `restore_vm_snapshot` - Restore a virtual machine from a snapshot
- `get_vm_firewall_rules` - Get firewall rules for a virtual machine
- `migrate_vm` - Migrate a virtual machine to another node

### Container Management (13 tools)
- `get_containers` - List all containers on a specific node
- `get_container_status` - Get detailed container information and status
- `get_container_config` - Get full configuration of a container
- `start_container` - Start an LXC container
- `stop_container` - Stop an LXC container
- `shutdown_container` - Gracefully shutdown an LXC container
- `reboot_container` - Reboot a container
- `delete_container` - Delete an LXC container (with optional force)
- `create_container` - Create a new LXC container with basic configuration
- `create_container_advanced` - Create a container with advanced configuration options
- `clone_container` - Clone an existing LXC container
- `update_container_config` - Update container configuration

## Skills & Capabilities

This MCP implements the following domain-specific skills:

1. **Cluster Management** - Monitor and manage cluster nodes and resources
2. **Virtual Machine Management** - Create and manage virtual machines
3. **Container Management** - Create and manage LXC containers
4. **Storage Management** - Manage and monitor storage infrastructure
5. **Monitoring & Analytics** - Monitor performance and health metrics
6. **Disaster Recovery** - Implement backup and recovery strategies

See [.github/skills](.github/skills) for detailed skill documentation.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PROXMOX_BASE_URL` | Proxmox server URL with port | Required |
| `PROXMOX_API_USER` | Proxmox API user (e.g., root@pam) | Required |
| `PROXMOX_API_TOKEN_ID` | Proxmox API token ID | Required |
| `PROXMOX_API_TOKEN_SECRET` | Proxmox API token secret | Required |
| `PROXMOX_SKIP_SSL_VERIFY` | Skip SSL certificate verification | false |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | info |

## API Reference

For detailed Proxmox API documentation and implementation details:
- **Proxmox API Docs**: https://pve.proxmox.com/pve-docs/api-viewer/index.html
- **Phase 1 Implementation Guide**: See [docs/PHASE1_IMPLEMENTATION.md](docs/PHASE1_IMPLEMENTATION.md)
- **Tools Quick Reference**: See [docs/TOOLS_QUICK_REFERENCE.md](docs/TOOLS_QUICK_REFERENCE.md)
- **API Specification**: See [docs/proxmox-api-spec.json](docs/proxmox-api-spec.json)
- **Gap Analysis & Roadmap**: See [docs/MISSING_TOOLS_ANALYSIS.md](docs/MISSING_TOOLS_ANALYSIS.md)

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Docker

```bash
make docker-build
make docker-run
```

## License

MIT License - See LICENSE file for details

## Support

For issues and questions:
- Check the [Proxmox API Documentation](https://pve.proxmox.com/pve-docs/api-viewer/index.html)
- Review implementation examples in `internal/`

---

**Built with Claude Haiku 4.5** - Crafted by AI to extend your infrastructure possibilities. 🤖✨
