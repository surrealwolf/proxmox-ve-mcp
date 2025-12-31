# Proxmox VE MCP - Complete Tools Reference

**Total Tools**: 68  
**Last Updated**: Phase 3 Implementation (December 2025)  
**Status**: ✅ All tools implemented and tested

---

## 📊 Tools by Category (68 Total)

### Cluster & Node Management (6 tools)
- `get_nodes` - Get all nodes in the cluster
- `get_node_status` - Get detailed status for a specific node
- `get_cluster_resources` - Get all cluster resources
- `get_cluster_status` - Get cluster-wide status
- `get_node_tasks` - Get tasks for a specific node
- `get_cluster_tasks` - Get all tasks in the cluster

### Storage Management (4 tools)
- `get_storage` - Get all storage devices
- `get_node_storage` - Get storage devices for a specific node
- `list_backups` - List available backups in storage
- [Additional storage operations available]

### Virtual Machine Management (23 tools)
#### Query/Monitoring (8)
- `get_vms` - Get all VMs on a node
- `get_vm_status` - Get detailed status of a VM
- `get_vm_config` - Get full configuration of a VM
- `get_vm_console` - Get console access information
- `get_vm_firewall_rules` - Get firewall rules for a VM
- `list_vm_snapshots` - List snapshots for a VM
- `get_vm_stats` - Get performance statistics for a VM

#### Control/Management (15)
- `start_vm` - Start a VM
- `stop_vm` - Stop a VM (immediate)
- `shutdown_vm` - Gracefully shutdown a VM
- `reboot_vm` - Reboot a VM
- `suspend_vm` - Suspend (pause) a VM
- `resume_vm` - Resume a suspended VM
- `create_vm` - Create a new VM
- `create_vm_advanced` - Create VM with advanced options
- `clone_vm` - Clone an existing VM
- `delete_vm` - Delete a VM
- `update_vm_config` - Update VM configuration
- `migrate_vm` - Migrate VM to another node
- `create_vm_snapshot` - Create a VM snapshot
- `delete_vm_snapshot` - Delete a snapshot
- `restore_vm_snapshot` - Restore from snapshot

### Container Management (23 tools)
#### Query/Monitoring (8)
- `get_containers` - Get all containers on a node
- `get_container_status` - Get detailed status of a container
- `get_container_config` - Get full configuration of a container
- `list_container_snapshots` - List snapshots for a container
- `get_container_stats` - Get performance statistics for a container

#### Control/Management (15)
- `start_container` - Start an LXC container
- `stop_container` - Stop a container (immediate)
- `shutdown_container` - Gracefully shutdown a container
- `reboot_container` - Reboot a container
- `create_container` - Create a new container
- `create_container_advanced` - Create container with advanced options
- `clone_container` - Clone an existing container
- `delete_container` - Delete a container
- `update_container_config` - Update container configuration
- `create_container_snapshot` - Create a container snapshot
- `delete_container_snapshot` - Delete a snapshot
- `restore_container_snapshot` - Restore from snapshot

### Backup & Restore (6 tools)
- `create_vm_backup` - Create a VM backup
- `create_container_backup` - Create a container backup
- `delete_backup` - Delete a backup file
- `restore_vm_backup` - Restore VM from backup
- `restore_container_backup` - Restore container from backup

### User & Access Management (15 tools)
#### Query (7)
- `list_users` - List all users
- `get_user` - Get user details
- `list_groups` - List all groups
- `list_roles` - List all roles
- `list_acl` - List all ACL entries

#### Control (8)
- `create_user` - Create a new user
- `update_user` - Update user properties
- `delete_user` - Delete a user
- `change_password` - Change user password
- `create_group` - Create a new group
- `delete_group` - Delete a group
- `create_role` - Create a custom role
- `delete_role` - Delete a role
- `create_api_token` - Create an API token
- `delete_api_token` - Delete an API token
- `set_acl` - Create/update ACL entries

### Resource Pool Management (2 tools)
- `list_pools` - List all resource pools
- `get_pool` - Get details for a specific pool

### Performance & Statistics (3 tools)
- `get_node_stats` - Get node performance statistics
- `get_vm_stats` - Get VM performance statistics
- `get_container_stats` - Get container performance statistics

---

## User Management Tools

### List & Query Operations

#### `list_users`
List all users in the system.
```
No parameters required
Returns: Array of User objects with userid, enable, expire, email, firstname, lastname, comment, groups
```

#### `get_user`
Get detailed information for a specific user.
```
Parameters:
  - userid (required): User ID, e.g., "user@pve"

Returns: User object with full details
```

#### `list_groups`
List all user groups in the system.
```
No parameters required
Returns: Array of Group objects
```

#### `list_roles`
List all available roles and their privileges.
```
No parameters required
Returns: Array of Role objects with roleid and privileges
```

#### `list_acl`
List all access control list entries.
```
No parameters required
Returns: Array of ACL entries showing path, role, user/group/token assignments
```

### Create Operations

#### `create_user`
Create a new user in the system.
```
Parameters:
  - userid (required): User ID, e.g., "newuser@pve"
  - password (required): Initial password
  - email (optional): Email address
  - comment (optional): User comment/description

Returns: Creation result with confirmation
```

#### `create_group`
Create a new user group.
```
Parameters:
  - groupid (required): Group ID
  - comment (optional): Group description

Returns: Creation result
```

#### `create_role`
Create a custom role with specific privileges.
```
Parameters:
  - roleid (required): Role ID (cannot start with "PVE")
  - privs (required): Space-separated privilege list
    Examples: "VM.PowerMgmt VM.Console"
              "Sys.Audit Sys.Modify"

Returns: Creation result
```

#### `create_api_token`
Generate a new API token for a user.
```
Parameters:
  - userid (required): User ID
  - tokenid (required): Token identifier
  - expire (optional): Unix timestamp for expiration
  - privsep (optional): true for privilege separation

Returns: Token creation result with generated token value
```

### Modify Operations

#### `update_user`
Update user properties.
```
Parameters:
  - userid (required): User ID
  - email (optional): Update email
  - comment (optional): Update comment
  - firstname (optional): First name
  - lastname (optional): Last name
  - enable (optional): true/false to enable/disable user
  - expire (optional): Unix timestamp for account expiration

Returns: Update confirmation
```

#### `change_password`
Change a user's password.
```
Parameters:
  - userid (required): User ID
  - password (required): New password

Returns: Password change confirmation
```

#### `set_acl`
Create or update an access control list entry.
```
Parameters:
  - path (required): ACL path (e.g., "/", "/nodes", "/vms")
  - role (required): Role ID to assign
  - userid (optional): User to assign role to
  - groupid (optional): Group to assign role to
  - tokenid (optional): API token to assign role to
  - propagate (optional): true/false to propagate permissions

Returns: ACL update confirmation
```

### Delete Operations

#### `delete_user`
Delete a user from the system.
```
Parameters:
  - userid (required): User ID to delete

Returns: Deletion confirmation
```

#### `delete_group`
Delete a user group.
```
Parameters:
  - groupid (required): Group ID to delete

Returns: Deletion confirmation
```

#### `delete_role`
Delete a custom role.
```
Parameters:
  - roleid (required): Role ID (custom roles only)

Returns: Deletion confirmation
```

#### `delete_api_token`
Delete/revoke an API token.
```
Parameters:
  - userid (required): User who owns the token
  - tokenid (required): Token ID to delete

Returns: Deletion confirmation
```

---

## Resource Pool Management

#### `list_pools`
List all resource pools in the cluster.
```
No parameters required
Returns: Array of Pool objects with poolid, comment, members, guests, storage
```

#### `get_pool`
Get details for a specific resource pool.
```
Parameters:
  - poolid (required): Pool ID

Returns: Pool object with full details
```

---

## Task & Statistics Monitoring

#### `get_node_tasks`
Get all tasks running on a specific node.
```
Parameters:
  - node_name (required): Node name

Returns: Array of Task objects for that node
```

#### `get_cluster_tasks`
Get all tasks running in the cluster.
```
No parameters required
Returns: Array of all Task objects
```

#### `get_node_stats`
Get performance statistics for a specific node.
```
Parameters:
  - node_name (required): Node name

Returns: Performance metrics (CPU, memory, disk usage, network)
```

#### `get_vm_stats`
Get performance statistics for a specific VM.
```
Parameters:
  - node_name (required): Node name
  - vmid (required): VM ID

Returns: VM resource usage statistics
```

#### `get_container_stats`
Get performance statistics for a specific container.
```
Parameters:
  - node_name (required): Node name
  - container_id (required): Container ID

Returns: Container resource usage statistics
```

---

## Notes & Best Practices

1. **User ID Format**: Always use format `username@realm`, e.g., `john@pve`
2. **API Tokens**: Token value is only shown once during creation - save it immediately
3. **Privilege Separation**: Use `privsep=true` for API tokens to limit token scope
4. **ACL Propagation**: Default is `propagate=true` - permissions apply to child resources
5. **Password Changes**: Some realms (LDAP, AD) don't support password changes via API
6. **Backup Locations**: Ensure storage has sufficient free space before creating backups
7. **Restore Operations**: Long-running - monitor task status via task monitoring tools
8. **Role Naming**: Custom roles cannot start with "PVE" prefix (reserved for built-in roles)
9. **Statistics**: Node stats use "day" timeframe by default
10. **Task Monitoring**: Use `get_node_tasks` and `get_cluster_tasks` to monitor operation progress

---

## Built-in Roles Reference

- **PVEAdmin** - Most tasks except system settings and permissions
- **PVEAuditor** - Read-only access
- **PVEVMAdmin** - Full VM administration
- **PVEVMUser** - Limited VM user (view, console, backups)
- **PVEDatastoreAdmin** - Storage and templates
- **PVEDatastoreUser** - Storage access (limited)
- **PVEPoolAdmin** - Manage pools
- **PVEPoolUser** - Use pools
- **PVEUserAdmin** - User management



### List & Query Operations

#### `list_users`
List all users in the system.
```
No parameters required
Returns: Array of User objects with userid, enable, expire, email, firstname, lastname, comment, groups
```

#### `get_user`
Get detailed information for a specific user.
```
Parameters:
  - userid (required): User ID, e.g., "user@pve"

Returns: User object with full details
```

#### `list_groups`
List all user groups in the system.
```
No parameters required
Returns: Array of Group objects
```

#### `list_roles`
List all available roles and their privileges.
```
No parameters required
Returns: Array of Role objects with roleid and privileges
```

#### `list_acl`
List all access control list entries.
```
No parameters required
Returns: Array of ACL entries showing path, role, user/group/token assignments
```

#### `list_api_tokens`
List API tokens for a specific user.
```
Parameters:
  - userid (required): User ID

Returns: Array of APIToken objects
```

### Create Operations

#### `create_user`
Create a new user in the system.
```
Parameters:
  - userid (required): User ID, e.g., "newuser@pve"
  - password (required): Initial password
  - email (optional): Email address
  - comment (optional): User comment/description

Returns: Creation result with confirmation
```

#### `create_group`
Create a new user group.
```
Parameters:
  - groupid (required): Group ID
  - comment (optional): Group description

Returns: Creation result
```

#### `create_role`
Create a custom role with specific privileges.
```
Parameters:
  - roleid (required): Role ID (cannot start with "PVE")
  - privs (required): Space-separated privilege list
    Examples: "VM.PowerMgmt VM.Console"
              "Sys.Audit Sys.Modify"

Available privileges:
  - Group.Allocate
  - Mapping.Audit, Mapping.Modify, Mapping.Use
  - Permissions.Modify
  - Pool.Allocate, Pool.Audit
  - Realm.AllocateUser, Realm.Allocate
  - SDN.Allocate, SDN.Audit, SDN.Use
  - Sys.Audit, Sys.Console, Sys.Incoming, Sys.Modify, Sys.PowerMgmt, Sys.Syslog
  - User.Modify
  - VM.Allocate, VM.Audit, VM.Backup, VM.Clone
  - VM.Config.CDROM, VM.Config.CPU, VM.Config.Cloudinit, VM.Config.Disk, VM.Config.HWType, VM.Config.Memory, VM.Config.Network, VM.Config.Options
  - VM.Console, VM.Migrate, VM.PowerMgmt, VM.Replicate, VM.Snapshot, VM.Snapshot.Rollback
  - VM.GuestAgent.Audit, VM.GuestAgent.FileRead, VM.GuestAgent.FileSystemMgmt, VM.GuestAgent.FileWrite, VM.GuestAgent.Unrestricted
  - Datastore.Allocate, Datastore.AllocateSpace, Datastore.AllocateTemplate, Datastore.Audit

Returns: Creation result
```

#### `create_api_token`
Generate a new API token for a user.
```
Parameters:
  - userid (required): User ID
  - tokenid (required): Token identifier
  - expire (optional): Unix timestamp for expiration
  - privsep (optional): true for privilege separation (token uses subset of user perms)

Returns: Token creation result with generated token value (only shown once!)
```

### Modify Operations

#### `update_user`
Update user properties.
```
Parameters:
  - userid (required): User ID
  - email (optional): Update email
  - comment (optional): Update comment
  - firstname (optional): First name
  - lastname (optional): Last name
  - enable (optional): true/false to enable/disable user
  - expire (optional): Unix timestamp for account expiration

Returns: Update confirmation
```

#### `change_password`
Change a user's password.
```
Parameters:
  - userid (required): User ID
  - password (required): New password

Returns: Password change confirmation
```

#### `set_acl`
Create or update an access control list entry.
```
Parameters:
  - path (required): ACL path:
    "/" - Full cluster access
    "/nodes" - All nodes
    "/nodes/{node}" - Specific node
    "/vms" - All VMs
    "/vms/{vmid}" - Specific VM
    "/storage" - All storage
    "/storage/{storage}" - Specific storage
    "/pool/{poolid}" - Specific pool
    "/access" - Access management
    "/access/groups" - Group administration
    
  - role (required): Role ID (e.g., "PVEAdmin", "PVEVMAdmin", "PVEAuditor")
  
  - userid (optional): User ID for assignment
  - groupid (optional): Group ID for assignment
  - tokenid (optional): Token ID for assignment
  - propagate (optional): true/false to propagate permissions down tree

Note: At least one of userid, groupid, or tokenid must be specified

Returns: ACL update confirmation
```

### Delete Operations

#### `delete_user`
Remove a user from the system.
```
Parameters:
  - userid (required): User ID

Returns: Deletion confirmation
```

#### `delete_group`
Remove a user group.
```
Parameters:
  - groupid (required): Group ID

Returns: Deletion confirmation
```

#### `delete_role`
Delete a custom role.
```
Parameters:
  - roleid (required): Role ID (cannot delete built-in roles)

Returns: Deletion confirmation
```

#### `delete_api_token`
Revoke an API token.
```
Parameters:
  - userid (required): User ID
  - tokenid (required): Token ID

Returns: Deletion confirmation
```

---

## Backup & Restore Tools

### List Operations

#### `list_backups`
List all available backups in a storage device.
```
Parameters:
  - storage (required): Storage device ID

Returns: Array of Backup objects with:
  - id: Backup identifier
  - name: Backup name
  - vmid: Associated VM ID (if VM backup)
  - size: Backup size in bytes
  - ctime: Creation timestamp
  - verified: Verification status
  - encrypted: Encryption status
```

### Create Backups

#### `create_vm_backup`
Create a backup of a virtual machine.
```
Parameters:
  - node_name (required): Node where VM is running
  - vmid (required): VM ID to backup
  - storage (required): Storage device for backup
  - backup_id (optional): Custom backup identifier
  - notes (optional): Backup notes/description

Returns: Backup task result with task ID for monitoring
```

#### `create_container_backup`
Create a backup of an LXC container.
```
Parameters:
  - node_name (required): Node where container is running
  - container_id (required): Container ID to backup
  - storage (required): Storage device for backup
  - backup_id (optional): Custom backup identifier
  - notes (optional): Backup notes/description

Returns: Backup task result with task ID for monitoring
```

### Restore from Backups

#### `restore_vm_backup`
Restore a virtual machine from a backup.
```
Parameters:
  - node_name (required): Target node for restoration
  - backup_id (required): Backup file ID/name
  - storage (required): Storage where backup is located

Returns: Restore task result with task ID

Note: This creates a new VM. Original VM must be deleted first to restore with same ID.
```

#### `restore_container_backup`
Restore a container from a backup.
```
Parameters:
  - node_name (required): Target node for restoration
  - backup_id (required): Backup file ID/name
  - storage (required): Storage where backup is located

Returns: Restore task result with task ID

Note: This creates a new container. Original container must be deleted first to restore with same ID.
```

### Delete Backups

#### `delete_backup`
Delete a backup file from storage.
```
Parameters:
  - storage (required): Storage device ID
  - backup_id (required): Backup file ID/name

Returns: Deletion confirmation

Warning: This permanently deletes the backup file. Cannot be recovered.
```

---

## Usage Examples

### Create a New User with Limited VM Admin Role
```
1. create_role: Create "vm_admin" role with privs "VM.PowerMgmt VM.Backup VM.Console"
2. create_user: Create user "john@pve" with initial password
3. set_acl: Assign user john@pve the vm_admin role on path /vms
```

### Backup and Restore Workflow
```
1. list_backups: Check available backups in "local" storage
2. create_vm_backup: Backup VM 100 to "local" storage
3. (Later) list_backups: Find the backup
4. restore_vm_backup: Restore to same or different node
```

### API Token for Monitoring
```
1. create_user: Create "monitor@pve" user
2. create_api_token: Generate token "monitoring" with privsep enabled
3. set_acl: Grant PVEAuditor role to token on /
   (Now the token has read-only access across cluster)
```

---

## Common ACL Paths

| Path | Scope | Use Case |
|------|-------|----------|
| `/` | Entire cluster | Full access |
| `/nodes` | All nodes | Node management |
| `/nodes/{node}` | Specific node | Single node access |
| `/vms` | All VMs | VM management |
| `/vms/{vmid}` | Specific VM | Single VM admin |
| `/storage` | All storage | Storage management |
| `/storage/{storage}` | Specific storage | Single storage access |
| `/pool/{poolid}` | Resource pool | Pool-based access |
| `/access/groups` | Groups | Group administration |

---

## Built-in Roles Reference

- **Administrator** - Full access to everything
- **PVEAdmin** - Most tasks except system settings and permissions
- **PVEAuditor** - Read-only access
- **PVEVMAdmin** - Full VM administration
- **PVEVMUser** - Limited VM user (view, console, backups)
- **PVEDatastoreAdmin** - Storage and templates
- **PVEDatastoreUser** - Storage access (limited)
- **PVEPoolAdmin** - Manage pools
- **PVEPoolUser** - Use pools
- **PVEUserAdmin** - User management

---

## Notes & Best Practices

1. **User ID Format**: Always use format `username@realm`, e.g., `john@pve`
2. **API Tokens**: Token value is only shown once during creation - save it immediately
3. **Privilege Separation**: Use `privsep=true` for API tokens to limit token scope
4. **ACL Propagation**: Default is `propagate=true` - permissions apply to child resources
5. **Password Changes**: Some realms (LDAP, AD) don't support password changes via API
6. **Backup Locations**: Ensure storage has sufficient free space before creating backups
7. **Restore Operations**: Long-running - monitor task status via task monitoring tools
8. **Role Naming**: Custom roles cannot start with "PVE" prefix (reserved for built-in roles)

