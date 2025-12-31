# Proxmox VE MCP

Model Context Protocol (MCP) server for Proxmox Virtual Environment infrastructure management. Control and monitor your Proxmox infrastructure through an AI-powered interface.

**Focused on:** Comprehensive Proxmox infrastructure management including cluster operations, VM/container lifecycle, user access control, and backup/restore operations.

⚠️ **Early Development Warning**: This project is in early development and may contain bugs that could cause unexpected behavior. Use with caution in production environments.

⚠️ **Prompt Injection Risk**: You are responsible for guarding against prompt injection when using these tools. Exercise extreme caution or use MCP tools only on systems and data you trust.

## Features

- **48 management tools** across 6 operational categories
- **User & Access Management**: 17 tools for users, groups, roles, API tokens, and ACLs
- **Backup & Restore Operations**: 6 tools for VM/container backup creation, management, and restoration
- **Cluster Management**: Monitor cluster health and node status
- **Virtual Machine Management**: List, monitor, and manage VMs
- **Container Management**: Manage LXC containers
- **Node Monitoring**: Track resource usage and uptime
- **Stdio Transport**: MCP protocol over standard input/output for seamless integration

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
PROXMOX_API_TOKEN=user@realm!tokenid=token-secret-here
PROXMOX_SKIP_SSL_VERIFY=false
LOG_LEVEL=info
```

### Obtaining API Token

1. Log in to Proxmox Web UI
2. Go to Datacenter → Permissions → API Tokens
3. Create a new API token with appropriate permissions
4. The token format is: `user@realm!tokenid=secret`

### Running the Server

```bash
./bin/proxmox-ve-mcp
```

The server listens on stdio and is ready for MCP protocol messages.

## Available Tools (48 Total)

### User & Access Management (17 tools)
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
- `list_acl_entries` - List all ACL entries (path-based permissions)
- `set_acl` - Configure access control for a specific path
- `list_api_tokens` - List all API tokens in the system
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
- `get_cluster_resources` - Get overview of cluster resources
- `get_storage` - List all storage devices in the cluster
- `get_node_storage` - Get storage devices for a specific node
- [Planned] Additional cluster operations

### Virtual Machine Management (6 tools)
- `get_vms` - List all VMs on a specific node
- `get_vm_status` - Get detailed VM information and status
- `start_vm` - Power on a virtual machine
- `stop_vm` - Power off a virtual machine
- `reboot_vm` - Reboot a virtual machine
- [Planned] VM configuration and creation

### Container Management (6 tools)
- `get_containers` - List all containers on a specific node
- `get_container_status` - Get detailed container information and status
- `start_container` - Start an LXC container
- `stop_container` - Stop an LXC container
- `reboot_container` - Reboot a container
- [Planned] Container configuration and creation

### Monitoring & Resources (1 tool)
- `get_storage` - Query cluster storage resources

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PROXMOX_BASE_URL` | Proxmox server URL with port | Required |
| `PROXMOX_API_TOKEN` | API token from Proxmox | Required |
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
