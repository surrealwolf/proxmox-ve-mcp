# Proxmox VE MCP - Missing Tools Analysis

## Overview
This document compares the current tool implementations with the complete Proxmox API specification to identify missing tools that could be added to enhance functionality.

## Current Implementation Summary
- **Total Tools**: 62
- **Read-only Tools**: 20 (query/monitoring)
- **Control Tools**: 42 (action/management)
- **Categories**: 6

## Recently Implemented Tools (Phase 2)
- ✅ `update_vm_config` - Update VM configuration including marking as template
- ✅ `update_container_config` - Update container configuration
- ✅ `get_vm_config` - Get full VM configuration
- ✅ `get_container_config` - Get full container configuration

### Current Tools by Category

#### Cluster & Node Management (3 tools)
- `get_nodes` ✓
- `get_node_status` ✓
- `get_cluster_resources` ✓

#### Storage Management (2 tools)
- `get_storage` ✓
- `get_node_storage` ✓

#### Virtual Machine Management (6 tools)
- `get_vms` ✓
- `get_vm_status` ✓
- `start_vm` ✓
- `stop_vm` ✓
- `shutdown_vm` ✓
- `reboot_vm` ✓

#### Container Management (6 tools)
- `get_containers` ✓
- `get_container_status` ✓
- `start_container` ✓
- `stop_container` ✓
- `shutdown_container` ✓
- `reboot_container` ✓

---

## Missing API Functionality

### 1. Access Management & User Management
**Importance**: HIGH  
**Security Impact**: Critical  
**Estimated Complexity**: Medium

#### Missing Tools:
- `list_users` - List all users in the system
- `create_user` - Create a new user
- `delete_user` - Remove a user from the system
- `update_user` - Modify user properties (email, name, groups, etc.)
- `change_password` - Change password for a user
- `list_groups` - List all groups
- `create_group` - Create a new user group
- `delete_group` - Remove a group
- `list_roles` - List all roles and their privileges
- `create_role` - Create a custom role with specific privileges
- `delete_role` - Remove a custom role
- `list_api_tokens` - List API tokens for users
- `create_api_token` - Generate new API token
- `delete_api_token` - Revoke an API token
- `list_acl` - List access control lists
- `set_acl` - Create or modify ACL entries

#### Use Cases:
- Automated user provisioning/deprovisioning
- Permission management and role assignment
- API token lifecycle management
- Audit and compliance operations
- Multi-tenant access control

#### Implementation Priority: **HIGH**

---

### 2. VM Configuration & Metadata
**Importance**: MEDIUM-HIGH  
**Security Impact**: Medium  
**Estimated Complexity**: Medium-High

#### Missing Tools:
- ✅ `get_vm_config` - Get full VM configuration details (IMPLEMENTED)
- ✅ `update_vm_config` - Modify VM configuration (CPU, memory, disks, etc.) (IMPLEMENTED)
- ✅ `get_vm_console` - Get console access information (IMPLEMENTED)
- ✅ `create_vm` - Create a new virtual machine (IMPLEMENTED)
- ✅ `delete_vm` - Delete a virtual machine (IMPLEMENTED)
- ✅ `clone_vm` - Clone an existing VM (IMPLEMENTED)
- ✅ `create_vm_snapshot` - Create a VM snapshot (IMPLEMENTED)
- ✅ `list_vm_snapshots` - List snapshots for a VM (IMPLEMENTED)
- ✅ `delete_vm_snapshot` - Remove a snapshot (IMPLEMENTED)
- ✅ `restore_vm_snapshot` - Restore a VM from snapshot (IMPLEMENTED)
- ✅ `get_vm_firewall_rules` - List firewall rules for a VM (IMPLEMENTED)
- ✅ `migrate_vm` - Migrate VM to another node (IMPLEMENTED)

#### Use Cases:
- Infrastructure as Code (IaC) automation
- VM configuration drift detection
- Capacity planning and scaling
- Disaster recovery and snapshots
- Network policy management
- VM lifecycle automation

#### Implementation Priority: **HIGH**

---

### 3. Container Configuration & Management
**Importance**: MEDIUM-HIGH  
**Security Impact**: Medium  
**Estimated Complexity**: Medium-High

#### Missing Tools:
- ✅ `get_container_config` - Get container configuration details (IMPLEMENTED)
- ✅ `update_container_config` - Modify container settings (IMPLEMENTED)
- ✅ `create_container` - Create a new container (IMPLEMENTED)
- ✅ `delete_container` - Delete a container (IMPLEMENTED)
- ✅ `clone_container` - Clone an existing container (IMPLEMENTED)
- `create_container_snapshot` - Create container snapshot
- `list_container_snapshots` - List container snapshots
- `delete_container_snapshot` - Remove a snapshot
- `restore_container_snapshot` - Restore container from snapshot

#### Use Cases:
- Container deployment automation
- Configuration management
- Container lifecycle operations
- Capacity planning

#### Implementation Priority: **MEDIUM**

---

### 4. Storage & Backup Management
**Importance**: HIGH  
**Security Impact**: High  
**Estimated Complexity**: High

#### Missing Tools:
- `get_storage_info` - Get detailed storage device information
- `create_storage` - Create new storage mount
- `delete_storage` - Remove storage configuration
- `get_storage_content` - List storage contents (ISO, backups, templates)
- `upload_backup` - Upload backup to storage
- `delete_backup` - Remove a backup file
- `get_backup_list` - List available backups
- `create_vm_backup` - Backup a virtual machine
- `create_container_backup` - Backup a container
- `restore_vm_backup` - Restore VM from backup
- `restore_container_backup` - Restore container from backup
- `get_storage_quota` - Get storage quotas

#### Use Cases:
- Backup automation and scheduling
- Disaster recovery automation
- Storage capacity management
- Backup lifecycle management
- Data retention policies

#### Implementation Priority: **CRITICAL**

---

### 5. Task & Background Job Management
**Importance**: MEDIUM  
**Security Impact**: Low  
**Estimated Complexity**: Low

#### Missing Tools:
- `list_tasks` - List background tasks (already has GetTasks in client but not exposed)
- `get_task_status` - Get detailed task status and progress
- `get_task_log` - Get task execution log
- `cancel_task` - Cancel a running task

#### Use Cases:
- Long-running operation monitoring
- Automation workflow tracking
- Error diagnostics
- Cleanup operations

#### Implementation Priority: **MEDIUM**

---

### 6. Node Management & Maintenance
**Importance**: MEDIUM  
**Security Impact**: Medium  
**Estimated Complexity**: Medium

#### Missing Tools:
- `get_node_config` - Get node network/system configuration
- `update_node_config` - Modify node settings
- `reboot_node` - Reboot a node
- `shutdown_node` - Gracefully shutdown a node
- `get_node_cert` - Get SSL certificate info
- `get_node_logs` - Get node system logs
- `get_node_apt_updates` - Check available updates
- `apply_node_updates` - Install system updates
- `get_node_disks` - List physical disks
- `get_node_network` - Get network configuration
- `get_node_dns` - Get DNS configuration

#### Use Cases:
- Maintenance operations
- System monitoring and alerting
- Compliance and audit logging
- Network troubleshooting
- Update management

#### Implementation Priority: **MEDIUM**

---

### 7. Cluster Management
**Importance**: MEDIUM  
**Security Impact**: Medium  
**Estimated Complexity**: High

#### Missing Tools:
- `get_cluster_status` - Get detailed cluster status
- `get_cluster_config` - Get cluster configuration
- `get_cluster_nodes_status` - Get all nodes in cluster and their status
- `add_node_to_cluster` - Add node to cluster
- `remove_node_from_cluster` - Remove node from cluster
- `get_ha_status` - Get HA (High Availability) status
- `enable_ha_resource` - Enable HA for a resource
- `disable_ha_resource` - Disable HA for a resource

#### Use Cases:
- Cluster topology management
- High availability management
- Disaster recovery planning
- Cluster capacity planning

#### Implementation Priority: **MEDIUM**

---

### 8. Firewall & Network Management
**Importance**: MEDIUM  
**Security Impact**: HIGH  
**Estimated Complexity**: Medium-High

#### Missing Tools:
- `get_firewall_rules` - List cluster firewall rules
- `create_firewall_rule` - Add firewall rule
- `delete_firewall_rule` - Remove firewall rule
- `get_security_groups` - List security groups
- `create_security_group` - Create security group
- `get_network_interfaces` - List network interfaces
- `get_vlan_config` - Get VLAN configuration

#### Use Cases:
- Network security hardening
- Traffic policy enforcement
- Compliance requirements
- Network troubleshooting

#### Implementation Priority: **MEDIUM**

---

### 9. Pool Management
**Importance**: MEDIUM  
**Security Impact**: Low  
**Estimated Complexity**: Low

#### Missing Tools:
- `list_pools` - List resource pools (not yet in tools schema)
- `create_pool` - Create new resource pool (not yet in tools schema)
- `update_pool` - Modify resource pool (not yet in tools schema)
- `delete_pool` - Remove resource pool (not yet in tools schema)
- `get_pool_members` - List resources in a pool

#### Use Cases:
- Multi-tenant resource separation
- Permission delegation
- Resource quota enforcement

#### Implementation Priority: **LOW**

---

### 10. Monitoring & Metrics
**Importance**: MEDIUM  
**Security Impact**: Low  
**Estimated Complexity**: Medium

#### Missing Tools:
- `get_vm_stats` - Get VM resource usage statistics
- `get_container_stats` - Get container resource usage statistics
- `get_node_stats` - Get node resource statistics over time
- `get_cluster_stats` - Get cluster-wide statistics
- `get_alerts` - Get active alerts and warnings

#### Use Cases:
- Capacity planning
- Performance analysis
- Trend analysis
- Alerting and monitoring integration

#### Implementation Priority: **MEDIUM**

---

## Priority Implementation Roadmap

### Phase 1: Critical (Immediate)
1. **Backup & Restore Tools** (Storage Management)
   - Backup creation, deletion, listing
   - Restore operations
   - **Impact**: Enables disaster recovery automation

2. **User & Access Management** (Access Control)
   - User CRUD operations
   - Group management
   - API token lifecycle
   - ACL management
   - **Impact**: Enables security and compliance automation

### Phase 2: High Priority (Next)
3. **VM Configuration Management** (VM Management)
   - Full VM CRUD operations
   - Snapshot management
   - Configuration queries and updates
   - **Impact**: Enables IaC and advanced automation

4. **Container Configuration Management** (Container Management)
   - Full container CRUD operations
   - Snapshot management
   - **Impact**: Enables container lifecycle automation

### Phase 3: Medium Priority
5. **Task & Log Management** (Operations)
   - Task monitoring
   - Log retrieval
   - **Impact**: Better operation visibility

6. **Node Maintenance** (Node Management)
   - Node reboot/shutdown
   - Update management
   - Configuration management
   - **Impact**: Enables automated maintenance

7. **Cluster Management** (Cluster Operations)
   - Cluster topology
   - HA management
   - **Impact**: Enables cluster operations automation

### Phase 4: Lower Priority
8. **Firewall & Network Management**
9. **Pool Management**
10. **Monitoring & Metrics**

---

## Implementation Considerations

### Client Library Updates Needed
The Proxmox Go client in `internal/proxmox/client.go` needs to add support for:
- User and group management API endpoints
- Storage and backup operations
- VM and container configuration endpoints
- Task status and logging endpoints
- Node maintenance operations
- Pool management endpoints

### MCP Server Updates
The MCP server in `internal/mcp/server.go` needs to:
- Register all new tools with proper input schemas
- Implement handlers for each new tool
- Add error handling and validation
- Update tool registration count

### Testing Requirements
- Unit tests for each new client method
- Integration tests with actual Proxmox instance
- Error handling tests for edge cases
- Input validation tests

### Documentation Updates
- Update tools-schema.json with new tools
- Add examples to QUICK_REFERENCE.md
- Document permission requirements for each tool
- Add troubleshooting guides

---

## Risk Assessment

### Security Considerations
- **User Management Tools**: Require proper authentication and role-based access control
- **Backup/Restore Tools**: Critical data operations - need audit logging
- **Firewall Tools**: Network security impact - comprehensive testing required
- **Cluster Operations**: Stability impact - thorough testing required

### API Compatibility
- Tools should target Proxmox VE 9.x+ API
- Need graceful handling of older/newer versions
- Version compatibility checks recommended

### Performance
- Backup/restore operations may be long-running - need timeout handling
- Large result sets (many VMs/containers) - pagination may be needed
- Bulk operations - consider rate limiting

---

## Metrics & Success Criteria

- [ ] All Priority 1 tools implemented and tested
- [ ] 90%+ API coverage for Proxmox VE REST API
- [ ] All tools have proper error handling
- [ ] Comprehensive documentation with examples
- [ ] Integration tests pass with real Proxmox instance
- [ ] Tools properly integrated into MCP framework

---

## Related Resources

- Proxmox VE API Viewer: https://pve.proxmox.com/pve-docs/api-viewer/
- Proxmox VE Admin Guide: https://pve.proxmox.com/pve-docs/pve-admin-guide.html
- Current API Spec: [proxmox-api-spec.json](proxmox-api-spec.json)
- Current Tools Schema: [tools-schema.json](tools-schema.json)
