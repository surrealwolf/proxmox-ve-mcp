# Proxmox VE MCP - Phase 1 Tools Quick Reference

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

