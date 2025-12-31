package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/surrealwolf/proxmox-ve-mcp/internal/proxmox"
)

// Server represents the MCP server
type Server struct {
	proxmoxClient *proxmox.Client
	server        *server.MCPServer
	logger        *logrus.Entry
}

// NewServer creates a new MCP server
func NewServer(proxmoxClient *proxmox.Client) *Server {
	s := &Server{
		proxmoxClient: proxmoxClient,
		server:        server.NewMCPServer("proxmox-ve-mcp", "0.1.0"),
		logger:        logrus.WithField("component", "MCPServer"),
	}

	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	tools := []server.ServerTool{}

	// Helper to create tool definitions
	addTool := func(name, desc string, handler server.ToolHandlerFunc, properties map[string]any) {
		tools = append(tools, server.ServerTool{
			Tool: mcp.Tool{
				Name:        name,
				Description: desc,
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: properties,
				},
			},
			Handler: handler,
		})
	}

	// Cluster and Node Management
	addTool("get_nodes", "Get all nodes in the Proxmox cluster", s.getNodes, map[string]any{})
	addTool("get_node_status", "Get detailed status information for a specific node", s.getNodeStatus, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
	})
	addTool("get_cluster_resources", "Get all cluster resources (nodes, VMs, containers)", s.getClusterResources, map[string]any{})
	addTool("get_cluster_status", "Get cluster-wide status information", s.getClusterStatus, map[string]any{})

	// Storage Management
	addTool("get_storage", "Get all storage devices in the cluster", s.getStorage, map[string]any{})
	addTool("get_node_storage", "Get storage devices for a specific node", s.getNodeStorage, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
	})

	// Virtual Machine Management - Query
	addTool("get_vms", "Get all VMs on a specific node", s.getVMs, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
	})
	addTool("get_vm_status", "Get detailed status of a specific VM", s.getVMStatus, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})

	// Virtual Machine Management - Control
	addTool("start_vm", "Start a virtual machine", s.startVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("stop_vm", "Stop a virtual machine (immediate)", s.stopVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("shutdown_vm", "Gracefully shutdown a virtual machine", s.shutdownVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("reboot_vm", "Reboot a virtual machine", s.rebootVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("get_vm_config", "Get full configuration of a virtual machine", s.getVMConfig, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("delete_vm", "Delete a virtual machine", s.deleteVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
		"force":     map[string]any{"type": "boolean", "description": "Force delete even if running (optional)"},
	})
	addTool("suspend_vm", "Suspend (pause) a virtual machine", s.suspendVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("resume_vm", "Resume a suspended virtual machine", s.resumeVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("create_vm", "Create a new virtual machine", s.createVM, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID (must be unique)"},
		"name":      map[string]any{"type": "string", "description": "VM name"},
		"memory":    map[string]any{"type": "integer", "description": "Memory in MB (default: 512)"},
		"cores":     map[string]any{"type": "integer", "description": "CPU cores (default: 1)"},
		"sockets":   map[string]any{"type": "integer", "description": "CPU sockets (default: 1)"},
	})
	addTool("create_vm_advanced", "Create a VM with advanced configuration options", s.createVMAdvanced, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID (must be unique)"},
		"name":      map[string]any{"type": "string", "description": "VM name (optional)"},
		"memory":    map[string]any{"type": "integer", "description": "Memory in MB (optional)"},
		"cores":     map[string]any{"type": "integer", "description": "CPU cores (optional)"},
		"sockets":   map[string]any{"type": "integer", "description": "CPU sockets (optional)"},
		"ide2":      map[string]any{"type": "string", "description": "CD/DVD drive (e.g., /mnt/pve/iso/ubuntu.iso, optional)"},
		"sata0":     map[string]any{"type": "string", "description": "Primary disk storage (e.g., local-lvm:10, optional)"},
		"net0":      map[string]any{"type": "string", "description": "Network configuration (e.g., virtio,bridge=vmbr0, optional)"},
	})
	addTool("clone_vm", "Clone an existing virtual machine", s.cloneVM, map[string]any{
		"node_name":   map[string]any{"type": "string", "description": "Name of the node"},
		"source_vmid": map[string]any{"type": "integer", "description": "Source VM ID to clone from"},
		"new_vmid":    map[string]any{"type": "integer", "description": "New VM ID (must be unique)"},
		"new_name":    map[string]any{"type": "string", "description": "New VM name"},
		"full":        map[string]any{"type": "boolean", "description": "Full clone (default: true) vs linked clone"},
	})
	addTool("update_vm_config", "Update virtual machine configuration (e.g., mark as template)", s.updateVMConfig, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
		"config":    map[string]any{"type": "object", "description": "Configuration to update (e.g., {\"template\": 1} to mark as template)"},
	})
	addTool("get_vm_console", "Get console access information for a VM", s.getVMConsole, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("create_vm_snapshot", "Create a snapshot of a virtual machine", s.createVMSnapshot, map[string]any{
		"node_name":   map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":        map[string]any{"type": "integer", "description": "VM ID"},
		"snap_name":   map[string]any{"type": "string", "description": "Snapshot name"},
		"description": map[string]any{"type": "string", "description": "Snapshot description (optional)"},
	})
	addTool("list_vm_snapshots", "List all snapshots for a virtual machine", s.listVMSnapshots, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("delete_vm_snapshot", "Delete a snapshot from a virtual machine", s.deleteVMSnapshot, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
		"snap_name": map[string]any{"type": "string", "description": "Snapshot name"},
		"force":     map[string]any{"type": "boolean", "description": "Force delete (optional)"},
	})
	addTool("restore_vm_snapshot", "Restore a virtual machine from a snapshot", s.restoreVMSnapshot, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
		"snap_name": map[string]any{"type": "string", "description": "Snapshot name"},
	})
	addTool("get_vm_firewall_rules", "Get firewall rules for a virtual machine", s.getVMFirewallRules, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("migrate_vm", "Migrate a virtual machine to another node", s.migrateVM, map[string]any{
		"node_name":   map[string]any{"type": "string", "description": "Source node name"},
		"vmid":        map[string]any{"type": "integer", "description": "VM ID"},
		"target_node": map[string]any{"type": "string", "description": "Target node name"},
		"online":      map[string]any{"type": "boolean", "description": "Perform live migration (optional)"},
	})

	// Container Management - Query
	addTool("get_containers", "Get all containers on a specific node", s.getContainers, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Name of the node"},
	})
	addTool("get_container_status", "Get detailed status of a specific container", s.getContainerStatus, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})

	// Container Management - Control
	addTool("start_container", "Start an LXC container", s.startContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("stop_container", "Stop an LXC container (immediate)", s.stopContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("shutdown_container", "Gracefully shutdown an LXC container", s.shutdownContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("reboot_container", "Reboot an LXC container", s.rebootContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("get_container_config", "Get full configuration of a container", s.getContainerConfig, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("delete_container", "Delete an LXC container", s.deleteContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"force":        map[string]any{"type": "boolean", "description": "Force delete even if running (optional)"},
	})
	addTool("create_container", "Create a new LXC container", s.createContainer, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID (must be unique)"},
		"hostname":     map[string]any{"type": "string", "description": "Container hostname"},
		"storage":      map[string]any{"type": "string", "description": "Storage device ID"},
		"memory":       map[string]any{"type": "integer", "description": "Memory in MB (default: 512)"},
		"cores":        map[string]any{"type": "integer", "description": "CPU cores (default: 1)"},
		"ostype":       map[string]any{"type": "string", "description": "OS type (default: debian)"},
	})
	addTool("create_container_advanced", "Create a container with advanced configuration options", s.createContainerAdvanced, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID (must be unique)"},
		"hostname":     map[string]any{"type": "string", "description": "Container hostname (optional)"},
		"storage":      map[string]any{"type": "string", "description": "Storage device ID (optional)"},
		"memory":       map[string]any{"type": "integer", "description": "Memory in MB (optional)"},
		"cores":        map[string]any{"type": "integer", "description": "CPU cores (optional)"},
		"ostype":       map[string]any{"type": "string", "description": "OS type (optional)"},
		"net0":         map[string]any{"type": "string", "description": "Network configuration (e.g., name=eth0,bridge=vmbr0, optional)"},
		"rootfs":       map[string]any{"type": "string", "description": "Root filesystem (e.g., local-lvm:10, optional)"},
	})
	addTool("clone_container", "Clone an existing LXC container", s.cloneContainer, map[string]any{
		"node_name":           map[string]any{"type": "string", "description": "Name of the node"},
		"source_container_id": map[string]any{"type": "integer", "description": "Source container ID to clone from"},
		"new_container_id":    map[string]any{"type": "integer", "description": "New container ID (must be unique)"},
		"new_hostname":        map[string]any{"type": "string", "description": "New container hostname"},
		"full":                map[string]any{"type": "boolean", "description": "Full clone (default: true) vs linked clone"},
	})
	addTool("update_container_config", "Update LXC container configuration", s.updateContainerConfig, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"config":       map[string]any{"type": "object", "description": "Configuration to update"},
	})
	addTool("create_container_snapshot", "Create a snapshot of an LXC container", s.createContainerSnapshot, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"snap_name":    map[string]any{"type": "string", "description": "Snapshot name"},
		"description":  map[string]any{"type": "string", "description": "Snapshot description (optional)"},
	})
	addTool("list_container_snapshots", "List all snapshots for an LXC container", s.listContainerSnapshots, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})
	addTool("delete_container_snapshot", "Delete a snapshot from an LXC container", s.deleteContainerSnapshot, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"snap_name":    map[string]any{"type": "string", "description": "Snapshot name"},
		"force":        map[string]any{"type": "boolean", "description": "Force delete (optional)"},
	})
	addTool("restore_container_snapshot", "Restore an LXC container from a snapshot", s.restoreContainerSnapshot, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Name of the node"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"snap_name":    map[string]any{"type": "string", "description": "Snapshot name"},
	})

	// User Management - Query
	addTool("list_users", "List all users in the system", s.listUsers, map[string]any{})
	addTool("get_user", "Get details for a specific user", s.getUser, map[string]any{
		"userid": map[string]any{"type": "string", "description": "User ID (e.g., user@pve)"},
	})
	addTool("list_groups", "List all groups", s.listGroups, map[string]any{})
	addTool("list_roles", "List all available roles and their privileges", s.listRoles, map[string]any{})
	addTool("list_acl", "List all access control list entries", s.listACLs, map[string]any{})

	// User Management - Control
	addTool("create_user", "Create a new user", s.createUser, map[string]any{
		"userid":   map[string]any{"type": "string", "description": "User ID (e.g., user@pve)"},
		"password": map[string]any{"type": "string", "description": "Initial password"},
		"email":    map[string]any{"type": "string", "description": "Email address (optional)"},
		"comment":  map[string]any{"type": "string", "description": "Comment (optional)"},
	})
	addTool("update_user", "Update user properties", s.updateUser, map[string]any{
		"userid":    map[string]any{"type": "string", "description": "User ID"},
		"email":     map[string]any{"type": "string", "description": "Email address (optional)"},
		"comment":   map[string]any{"type": "string", "description": "Comment (optional)"},
		"firstname": map[string]any{"type": "string", "description": "First name (optional)"},
		"lastname":  map[string]any{"type": "string", "description": "Last name (optional)"},
		"enable":    map[string]any{"type": "boolean", "description": "Enable/disable user (optional)"},
		"expire":    map[string]any{"type": "integer", "description": "Expiration Unix timestamp (optional)"},
	})
	addTool("delete_user", "Delete a user", s.deleteUser, map[string]any{
		"userid": map[string]any{"type": "string", "description": "User ID"},
	})
	addTool("change_password", "Change user password", s.changePassword, map[string]any{
		"userid":   map[string]any{"type": "string", "description": "User ID"},
		"password": map[string]any{"type": "string", "description": "New password"},
	})
	addTool("create_group", "Create a new user group", s.createGroup, map[string]any{
		"groupid": map[string]any{"type": "string", "description": "Group ID"},
		"comment": map[string]any{"type": "string", "description": "Comment (optional)"},
	})
	addTool("delete_group", "Delete a user group", s.deleteGroup, map[string]any{
		"groupid": map[string]any{"type": "string", "description": "Group ID"},
	})
	addTool("create_role", "Create a new role with specific privileges", s.createRole, map[string]any{
		"roleid": map[string]any{"type": "string", "description": "Role ID"},
		"privs":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of privileges"},
	})
	addTool("delete_role", "Delete a role", s.deleteRole, map[string]any{
		"roleid": map[string]any{"type": "string", "description": "Role ID"},
	})
	addTool("set_acl", "Create or update an access control list entry", s.setACL, map[string]any{
		"path":      map[string]any{"type": "string", "description": "ACL path (e.g., /vms, /nodes)"},
		"role":      map[string]any{"type": "string", "description": "Role ID"},
		"userid":    map[string]any{"type": "string", "description": "User ID (optional)"},
		"groupid":   map[string]any{"type": "string", "description": "Group ID (optional)"},
		"tokenid":   map[string]any{"type": "string", "description": "Token ID (optional)"},
		"propagate": map[string]any{"type": "boolean", "description": "Propagate permissions down tree (optional)"},
	})
	addTool("create_api_token", "Create a new API token for a user", s.createAPIToken, map[string]any{
		"userid":  map[string]any{"type": "string", "description": "User ID"},
		"tokenid": map[string]any{"type": "string", "description": "Token ID"},
		"expire":  map[string]any{"type": "integer", "description": "Expiration Unix timestamp (optional)"},
		"privsep": map[string]any{"type": "boolean", "description": "Separate privileges (optional)"},
	})
	addTool("delete_api_token", "Delete an API token", s.deleteAPIToken, map[string]any{
		"userid":  map[string]any{"type": "string", "description": "User ID"},
		"tokenid": map[string]any{"type": "string", "description": "Token ID"},
	})

	// Backup & Restore - Query
	addTool("list_backups", "List available backups in storage", s.listBackups, map[string]any{
		"storage": map[string]any{"type": "string", "description": "Storage device ID"},
	})

	// Backup & Restore - Control
	addTool("create_vm_backup", "Create a backup of a virtual machine", s.createVMBackup, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
		"storage":   map[string]any{"type": "string", "description": "Storage device ID"},
		"backup_id": map[string]any{"type": "string", "description": "Backup ID (optional)"},
		"notes":     map[string]any{"type": "string", "description": "Backup notes (optional)"},
	})
	addTool("create_container_backup", "Create a backup of a container", s.createContainerBackup, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Node name"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
		"storage":      map[string]any{"type": "string", "description": "Storage device ID"},
		"backup_id":    map[string]any{"type": "string", "description": "Backup ID (optional)"},
		"notes":        map[string]any{"type": "string", "description": "Backup notes (optional)"},
	})
	addTool("delete_backup", "Delete a backup file", s.deleteBackup, map[string]any{
		"storage":   map[string]any{"type": "string", "description": "Storage device ID"},
		"backup_id": map[string]any{"type": "string", "description": "Backup ID/filename"},
	})
	addTool("restore_vm_backup", "Restore a virtual machine from a backup", s.restoreVMBackup, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
		"backup_id": map[string]any{"type": "string", "description": "Backup ID/filename"},
		"storage":   map[string]any{"type": "string", "description": "Storage device ID"},
	})
	addTool("restore_container_backup", "Restore a container from a backup", s.restoreContainerBackup, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
		"backup_id": map[string]any{"type": "string", "description": "Backup ID/filename"},
		"storage":   map[string]any{"type": "string", "description": "Storage device ID"},
	})

	// Resource Pools - Query
	addTool("list_pools", "List all resource pools in the cluster", s.listPools, map[string]any{})
	addTool("get_pool", "Get details for a specific resource pool", s.getPool, map[string]any{
		"poolid": map[string]any{"type": "string", "description": "Pool ID"},
	})

	// Node Management
	addTool("get_node_tasks", "Get tasks for a specific node", s.getNodeTasks, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
	})
	addTool("get_cluster_tasks", "Get all tasks in the cluster", s.getClusterTasks, map[string]any{})

	// Statistics
	addTool("get_node_stats", "Get performance statistics for a specific node", s.getNodeStats, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
	})
	addTool("get_vm_stats", "Get performance statistics for a specific VM", s.getVMStats, map[string]any{
		"node_name": map[string]any{"type": "string", "description": "Node name"},
		"vmid":      map[string]any{"type": "integer", "description": "VM ID"},
	})
	addTool("get_container_stats", "Get performance statistics for a specific container", s.getContainerStats, map[string]any{
		"node_name":    map[string]any{"type": "string", "description": "Node name"},
		"container_id": map[string]any{"type": "integer", "description": "Container ID"},
	})

	for _, tool := range tools {
		s.server.AddTool(tool.Tool, tool.Handler)
	}
	s.logger.Info("Registered 68 tools")
}

// ServeStdio starts the MCP server with stdio transport
func (s *Server) ServeStdio(ctx context.Context) error {
	s.logger.Info("Starting Proxmox VE MCP Server")
	return server.ServeStdio(s.server)
}

// ServeHTTP starts the MCP server with HTTP transport
func (s *Server) ServeHTTP(addr string, ctx context.Context) error {
	s.logger.Infof("Starting Proxmox VE MCP Server on HTTP at %s", addr)

	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the MCP request
		var requestData map[string]interface{}
		if err := json.Unmarshal(body, &requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log the request
		s.logger.Debugf("HTTP MCP request received: %v", requestData)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status": "MCP HTTP transport is available",
			"info":   "This is an HTTP endpoint. Use stdio transport for full MCP protocol support.",
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	return http.ListenAndServe(addr, nil)
}

// getNodes handles the get_nodes tool
func (s *Server) getNodes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_nodes")

	nodes, err := s.proxmoxClient.GetNodes(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get nodes: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// getNodeStatus handles the get_node_status tool
func (s *Server) getNodeStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_node_status")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	node, err := s.proxmoxClient.GetNode(ctx, nodeName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get node status: %v", err)), nil
	}

	return mcp.NewToolResultJSON(node)
}

// getVMs handles the get_vms tool
func (s *Server) getVMs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vms")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vms, err := s.proxmoxClient.GetVMs(ctx, nodeName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VMs: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"vms":   vms,
		"count": len(vms),
		"node":  nodeName,
	})
}

// getVMStatus handles the get_vm_status tool
func (s *Server) getVMStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vm_status")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	vm, err := s.proxmoxClient.GetVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM status: %v", err)), nil
	}

	return mcp.NewToolResultJSON(vm)
}

// getContainers handles the get_containers tool
func (s *Server) getContainers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_containers")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containers, err := s.proxmoxClient.GetContainers(ctx, nodeName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get containers: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"containers": containers,
		"count":      len(containers),
		"node":       nodeName,
	})
}

// getContainerStatus handles the get_container_status tool
func (s *Server) getContainerStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_container_status")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	container, err := s.proxmoxClient.GetContainer(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get container status: %v", err)), nil
	}

	return mcp.NewToolResultJSON(container)
}

// getClusterResources handles the get_cluster_resources tool
func (s *Server) getClusterResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_cluster_resources")

	resources, err := s.proxmoxClient.GetClusterResources(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get cluster resources: %v", err)), nil
	}

	return mcp.NewToolResultJSON(resources)
}

// getClusterStatus handles the get_cluster_status tool
func (s *Server) getClusterStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_cluster_status")

	status, err := s.proxmoxClient.GetClusterStatus(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get cluster status: %v", err)), nil
	}

	return mcp.NewToolResultJSON(status)
}

// getStorage handles the get_storage tool
func (s *Server) getStorage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_storage")

	storage, err := s.proxmoxClient.GetStorage(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get storage: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"storage": storage,
		"count":   len(storage),
	})
}

// getNodeStorage handles the get_node_storage tool
func (s *Server) getNodeStorage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_node_storage")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	storage, err := s.proxmoxClient.GetNodeStorage(ctx, nodeName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get node storage: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"storage": storage,
		"node":    nodeName,
		"count":   len(storage),
	})
}

// startVM handles the start_vm tool
func (s *Server) startVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: start_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.StartVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "start",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// stopVM handles the stop_vm tool
func (s *Server) stopVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: stop_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.StopVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to stop VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "stop",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// shutdownVM handles the shutdown_vm tool
func (s *Server) shutdownVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: shutdown_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.ShutdownVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to shutdown VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "shutdown",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// rebootVM handles the reboot_vm tool
func (s *Server) rebootVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: reboot_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.RebootVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reboot VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "reboot",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// getVMConfig handles the get_vm_config tool
func (s *Server) getVMConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vm_config")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	config, err := s.proxmoxClient.GetVMConfig(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM config: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"vmid":   vmID,
		"node":   nodeName,
		"config": config,
	})
}

// deleteVM handles the delete_vm tool
func (s *Server) deleteVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	force := request.GetBool("force", false)

	result, err := s.proxmoxClient.DeleteVM(ctx, nodeName, vmID, force)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "delete",
		"vmid":   vmID,
		"node":   nodeName,
		"force":  force,
		"result": result,
	})
}

// suspendVM handles the suspend_vm tool
func (s *Server) suspendVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: suspend_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.SuspendVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to suspend VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "suspend",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// resumeVM handles the resume_vm tool
func (s *Server) resumeVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: resume_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.ResumeVM(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resume VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "resume",
		"vmid":   vmID,
		"node":   nodeName,
		"result": result,
	})
}

// createVM handles the create_vm tool
func (s *Server) createVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	memory := request.GetInt("memory", 512)
	cores := request.GetInt("cores", 1)
	sockets := request.GetInt("sockets", 1)

	result, err := s.proxmoxClient.CreateVMFull(ctx, nodeName, vmID, name, memory, cores, sockets)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "create",
		"vmid":   vmID,
		"name":   name,
		"node":   nodeName,
		"config": map[string]interface{}{
			"memory":  memory,
			"cores":   cores,
			"sockets": sockets,
		},
		"result": result,
	})
}

// createVMAdvanced handles the create_vm_advanced tool with full configuration
func (s *Server) createVMAdvanced(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_vm_advanced")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	// Build configuration from optional parameters
	config := map[string]interface{}{
		"vmid": vmID,
	}

	// Add optional parameters if provided
	if name := request.GetString("name", ""); name != "" {
		config["name"] = name
	}
	if memory := request.GetInt("memory", 0); memory > 0 {
		config["memory"] = memory
	}
	if cores := request.GetInt("cores", 0); cores > 0 {
		config["cores"] = cores
	}
	if sockets := request.GetInt("sockets", 0); sockets > 0 {
		config["sockets"] = sockets
	}
	if ide2 := request.GetString("ide2", ""); ide2 != "" {
		config["ide2"] = ide2
	}
	if sata0 := request.GetString("sata0", ""); sata0 != "" {
		config["sata0"] = sata0
	}
	if net0 := request.GetString("net0", ""); net0 != "" {
		config["net0"] = net0
	}

	result, err := s.proxmoxClient.CreateVM(ctx, nodeName, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "create",
		"vmid":   vmID,
		"node":   nodeName,
		"config": config,
		"result": result,
	})
}

// cloneVM handles the clone_vm tool
func (s *Server) cloneVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: clone_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	sourceVMID := request.GetInt("source_vmid", 0)
	if sourceVMID <= 0 {
		return mcp.NewToolResultError("source_vmid parameter is required and must be a positive integer"), nil
	}

	newVMID := request.GetInt("new_vmid", 0)
	if newVMID <= 0 {
		return mcp.NewToolResultError("new_vmid parameter is required and must be a positive integer"), nil
	}

	newName := request.GetString("new_name", "")
	if newName == "" {
		return mcp.NewToolResultError("new_name parameter is required"), nil
	}

	full := request.GetBool("full", true)

	result, err := s.proxmoxClient.CloneVM(ctx, nodeName, sourceVMID, newVMID, newName, full)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to clone VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":      "clone",
		"source_vmid": sourceVMID,
		"new_vmid":    newVMID,
		"new_name":    newName,
		"node":        nodeName,
		"full_clone":  full,
		"result":      result,
	})
}

// updateVMConfig handles the update_vm_config tool
func (s *Server) updateVMConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: update_vm_config")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	// Get all arguments to extract config
	args := request.GetArguments()
	configValue := args["config"]
	if configValue == nil {
		return mcp.NewToolResultError("config parameter is required"), nil
	}

	// Convert config to map[string]interface{}
	var config map[string]interface{}
	configBytes, err := json.Marshal(configValue)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid config parameter: %v", err)), nil
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse config: %v", err)), nil
	}

	result, err := s.proxmoxClient.UpdateVM(ctx, nodeName, vmID, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update VM config: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action": "update_config",
		"vmid":   vmID,
		"node":   nodeName,
		"config": config,
		"result": result,
	})
}

// getVMConsole handles the get_vm_console tool
func (s *Server) getVMConsole(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vm_console")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.GetVMConsole(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM console: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"vmid":    vmID,
		"node":    nodeName,
		"console": result,
	})
}

// createVMSnapshot handles the create_vm_snapshot tool
func (s *Server) createVMSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_vm_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	description := request.GetString("description", "")

	result, err := s.proxmoxClient.CreateVMSnapshot(ctx, nodeName, vmID, snapName, description)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create VM snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":      "create_snapshot",
		"vmid":        vmID,
		"node":        nodeName,
		"snapshot":    snapName,
		"description": description,
		"result":      result,
	})
}

// listVMSnapshots handles the list_vm_snapshots tool
func (s *Server) listVMSnapshots(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_vm_snapshots")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.ListVMSnapshots(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list VM snapshots: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"vmid":      vmID,
		"node":      nodeName,
		"snapshots": result,
	})
}

// deleteVMSnapshot handles the delete_vm_snapshot tool
func (s *Server) deleteVMSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_vm_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	force := request.GetBool("force", false)

	result, err := s.proxmoxClient.DeleteVMSnapshot(ctx, nodeName, vmID, snapName, force)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete VM snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":   "delete_snapshot",
		"vmid":     vmID,
		"node":     nodeName,
		"snapshot": snapName,
		"force":    force,
		"result":   result,
	})
}

// restoreVMSnapshot handles the restore_vm_snapshot tool
func (s *Server) restoreVMSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: restore_vm_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	result, err := s.proxmoxClient.RestoreVMSnapshot(ctx, nodeName, vmID, snapName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore VM snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":   "restore_snapshot",
		"vmid":     vmID,
		"node":     nodeName,
		"snapshot": snapName,
		"result":   result,
	})
}

// getVMFirewallRules handles the get_vm_firewall_rules tool
func (s *Server) getVMFirewallRules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vm_firewall_rules")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.GetVMFirewallRules(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM firewall rules: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"vmid":  vmID,
		"node":  nodeName,
		"rules": result,
	})
}

// migrateVM handles the migrate_vm tool
func (s *Server) migrateVM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: migrate_vm")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	targetNode := request.GetString("target_node", "")
	if targetNode == "" {
		return mcp.NewToolResultError("target_node parameter is required"), nil
	}

	online := request.GetBool("online", false)

	result, err := s.proxmoxClient.MigrateVM(ctx, nodeName, vmID, targetNode, online)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to migrate VM: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":      "migrate",
		"vmid":        vmID,
		"source_node": nodeName,
		"target_node": targetNode,
		"online":      online,
		"result":      result,
	})
}

// startContainer handles the start_container tool
func (s *Server) startContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: start_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.StartContainer(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "start",
		"container_id": containerID,
		"node":         nodeName,
		"result":       result,
	})
}

// stopContainer handles the stop_container tool
func (s *Server) stopContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: stop_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.StopContainer(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to stop container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "stop",
		"container_id": containerID,
		"node":         nodeName,
		"result":       result,
	})
}

// shutdownContainer handles the shutdown_container tool
func (s *Server) shutdownContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: shutdown_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.ShutdownContainer(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to shutdown container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "shutdown",
		"container_id": containerID,
		"node":         nodeName,
		"result":       result,
	})
}

// rebootContainer handles the reboot_container tool
func (s *Server) rebootContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: reboot_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.RebootContainer(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reboot container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "reboot",
		"container_id": containerID,
		"node":         nodeName,
		"result":       result,
	})
}

// getContainerConfig handles the get_container_config tool
func (s *Server) getContainerConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_container_config")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	config, err := s.proxmoxClient.GetContainerConfig(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get container config: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"container_id": containerID,
		"node":         nodeName,
		"config":       config,
	})
}

// deleteContainer handles the delete_container tool
func (s *Server) deleteContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	force := request.GetBool("force", false)

	result, err := s.proxmoxClient.DeleteContainer(ctx, nodeName, containerID, force)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "delete",
		"container_id": containerID,
		"node":         nodeName,
		"force":        force,
		"result":       result,
	})
}

// createContainer handles the create_container tool
func (s *Server) createContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	hostname := request.GetString("hostname", "")
	if hostname == "" {
		return mcp.NewToolResultError("hostname parameter is required"), nil
	}

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	memory := request.GetInt("memory", 512)
	cores := request.GetInt("cores", 1)
	ostype := request.GetString("ostype", "debian")

	result, err := s.proxmoxClient.CreateContainerFull(ctx, nodeName, containerID, hostname, storage, memory, cores, ostype)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "create",
		"container_id": containerID,
		"hostname":     hostname,
		"node":         nodeName,
		"config": map[string]interface{}{
			"storage": storage,
			"memory":  memory,
			"cores":   cores,
			"ostype":  ostype,
		},
		"result": result,
	})
}

// createContainerAdvanced handles the create_container_advanced tool with full configuration
func (s *Server) createContainerAdvanced(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_container_advanced")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	// Build configuration from optional parameters
	config := map[string]interface{}{
		"vmid": containerID,
	}

	// Add optional parameters if provided
	if hostname := request.GetString("hostname", ""); hostname != "" {
		config["hostname"] = hostname
	}
	if storage := request.GetString("storage", ""); storage != "" {
		config["storage"] = storage
	}
	if memory := request.GetInt("memory", 0); memory > 0 {
		config["memory"] = memory
	}
	if cores := request.GetInt("cores", 0); cores > 0 {
		config["cores"] = cores
	}
	if ostype := request.GetString("ostype", ""); ostype != "" {
		config["ostype"] = ostype
	}
	if net0 := request.GetString("net0", ""); net0 != "" {
		config["net0"] = net0
	}
	if rootfs := request.GetString("rootfs", ""); rootfs != "" {
		config["rootfs"] = rootfs
	}

	result, err := s.proxmoxClient.CreateContainer(ctx, nodeName, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "create",
		"container_id": containerID,
		"node":         nodeName,
		"config":       config,
		"result":       result,
	})
}

// cloneContainer handles the clone_container tool
func (s *Server) cloneContainer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: clone_container")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	sourceContainerID := request.GetInt("source_container_id", 0)
	if sourceContainerID <= 0 {
		return mcp.NewToolResultError("source_container_id parameter is required and must be a positive integer"), nil
	}

	newContainerID := request.GetInt("new_container_id", 0)
	if newContainerID <= 0 {
		return mcp.NewToolResultError("new_container_id parameter is required and must be a positive integer"), nil
	}

	newHostname := request.GetString("new_hostname", "")
	if newHostname == "" {
		return mcp.NewToolResultError("new_hostname parameter is required"), nil
	}

	full := request.GetBool("full", true)

	result, err := s.proxmoxClient.CloneContainer(ctx, nodeName, sourceContainerID, newContainerID, newHostname, full)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to clone container: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":              "clone",
		"source_container_id": sourceContainerID,
		"new_container_id":    newContainerID,
		"new_hostname":        newHostname,
		"node":                nodeName,
		"full_clone":          full,
		"result":              result,
	})
}

// updateContainerConfig handles the update_container_config tool
func (s *Server) updateContainerConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: update_container_config")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	// Get all arguments to extract config
	args := request.GetArguments()
	configValue := args["config"]
	if configValue == nil {
		return mcp.NewToolResultError("config parameter is required"), nil
	}

	// Convert config to map[string]interface{}
	var config map[string]interface{}
	configBytes, err := json.Marshal(configValue)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid config parameter: %v", err)), nil
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse config: %v", err)), nil
	}

	result, err := s.proxmoxClient.UpdateContainer(ctx, nodeName, containerID, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update container config: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "update_config",
		"container_id": containerID,
		"node":         nodeName,
		"config":       config,
		"result":       result,
	})
}

// createContainerSnapshot handles the create_container_snapshot tool
func (s *Server) createContainerSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_container_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	description := request.GetString("description", "")

	result, err := s.proxmoxClient.CreateContainerSnapshot(ctx, nodeName, containerID, snapName, description)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create container snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "create_snapshot",
		"container_id": containerID,
		"node":         nodeName,
		"snapshot":     snapName,
		"description":  description,
		"result":       result,
	})
}

// listContainerSnapshots handles the list_container_snapshots tool
func (s *Server) listContainerSnapshots(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_container_snapshots")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	result, err := s.proxmoxClient.ListContainerSnapshots(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list container snapshots: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"container_id": containerID,
		"node":         nodeName,
		"snapshots":    result,
	})
}

// deleteContainerSnapshot handles the delete_container_snapshot tool
func (s *Server) deleteContainerSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_container_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	force := request.GetBool("force", false)

	result, err := s.proxmoxClient.DeleteContainerSnapshot(ctx, nodeName, containerID, snapName, force)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete container snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "delete_snapshot",
		"container_id": containerID,
		"node":         nodeName,
		"snapshot":     snapName,
		"force":        force,
		"result":       result,
	})
}

// restoreContainerSnapshot handles the restore_container_snapshot tool
func (s *Server) restoreContainerSnapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: restore_container_snapshot")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	snapName := request.GetString("snap_name", "")
	if snapName == "" {
		return mcp.NewToolResultError("snap_name parameter is required"), nil
	}

	result, err := s.proxmoxClient.RestoreContainerSnapshot(ctx, nodeName, containerID, snapName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore container snapshot: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "restore_snapshot",
		"container_id": containerID,
		"node":         nodeName,
		"snapshot":     snapName,
		"result":       result,
	})
}

// ============ USER MANAGEMENT ============

// listUsers handles the list_users tool
func (s *Server) listUsers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_users")

	users, err := s.proxmoxClient.ListUsers(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list users: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// getUser handles the get_user tool
func (s *Server) getUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_user")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	user, err := s.proxmoxClient.GetUser(ctx, userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user: %v", err)), nil
	}

	return mcp.NewToolResultJSON(user)
}

// createUser handles the create_user tool
func (s *Server) createUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_user")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	password := request.GetString("password", "")
	if password == "" {
		return mcp.NewToolResultError("password parameter is required"), nil
	}

	email := request.GetString("email", "")
	comment := request.GetString("comment", "")

	result, err := s.proxmoxClient.CreateUser(ctx, userID, password, email, comment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create user: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "create",
		"userid":  userID,
		"message": "User created successfully",
		"result":  result,
	})
}

// updateUser handles the update_user tool
func (s *Server) updateUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: update_user")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	email := request.GetString("email", "")
	comment := request.GetString("comment", "")
	firstName := request.GetString("firstname", "")
	lastName := request.GetString("lastname", "")
	enable := request.GetBool("enable", true)
	expire := int64(request.GetInt("expire", 0))

	result, err := s.proxmoxClient.UpdateUser(ctx, userID, email, comment, firstName, lastName, enable, expire)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "update",
		"userid":  userID,
		"message": "User updated successfully",
		"result":  result,
	})
}

// deleteUser handles the delete_user tool
func (s *Server) deleteUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_user")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	result, err := s.proxmoxClient.DeleteUser(ctx, userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete user: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "delete",
		"userid":  userID,
		"message": "User deleted successfully",
		"result":  result,
	})
}

// changePassword handles the change_password tool
func (s *Server) changePassword(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: change_password")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	password := request.GetString("password", "")
	if password == "" {
		return mcp.NewToolResultError("password parameter is required"), nil
	}

	result, err := s.proxmoxClient.ChangePassword(ctx, userID, password)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to change password: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "change_password",
		"userid":  userID,
		"message": "Password changed successfully",
		"result":  result,
	})
}

// listGroups handles the list_groups tool
func (s *Server) listGroups(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_groups")

	groups, err := s.proxmoxClient.ListGroups(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list groups: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"groups": groups,
		"count":  len(groups),
	})
}

// createGroup handles the create_group tool
func (s *Server) createGroup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_group")

	groupID := request.GetString("groupid", "")
	if groupID == "" {
		return mcp.NewToolResultError("groupid parameter is required"), nil
	}

	comment := request.GetString("comment", "")

	result, err := s.proxmoxClient.CreateGroup(ctx, groupID, comment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create group: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "create",
		"groupid": groupID,
		"message": "Group created successfully",
		"result":  result,
	})
}

// deleteGroup handles the delete_group tool
func (s *Server) deleteGroup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_group")

	groupID := request.GetString("groupid", "")
	if groupID == "" {
		return mcp.NewToolResultError("groupid parameter is required"), nil
	}

	result, err := s.proxmoxClient.DeleteGroup(ctx, groupID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete group: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "delete",
		"groupid": groupID,
		"message": "Group deleted successfully",
		"result":  result,
	})
}

// listRoles handles the list_roles tool
func (s *Server) listRoles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_roles")

	roles, err := s.proxmoxClient.ListRoles(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list roles: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	})
}

// createRole handles the create_role tool
func (s *Server) createRole(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_role")

	roleID := request.GetString("roleid", "")
	if roleID == "" {
		return mcp.NewToolResultError("roleid parameter is required"), nil
	}

	// Parse privileges - for simplicity, accept a space-separated string or array
	privs := []string{}

	// Try to get as string first (space-separated)
	if privStr := request.GetString("privs", ""); privStr != "" {
		// If it's a string, it might be space-separated
		privsList := splitPrivileges(privStr)
		privs = privsList
	}

	result, err := s.proxmoxClient.CreateRole(ctx, roleID, privs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create role: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "create",
		"roleid":  roleID,
		"message": "Role created successfully",
		"result":  result,
	})
}

// Helper function to split privileges string
func splitPrivileges(privStr string) []string {
	if privStr == "" {
		return []string{}
	}
	// For now, assume space-separated or comma-separated
	parts := make([]string, 0)
	current := ""
	for _, ch := range privStr {
		if ch == ' ' || ch == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// deleteRole handles the delete_role tool
func (s *Server) deleteRole(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_role")

	roleID := request.GetString("roleid", "")
	if roleID == "" {
		return mcp.NewToolResultError("roleid parameter is required"), nil
	}

	result, err := s.proxmoxClient.DeleteRole(ctx, roleID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete role: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "delete",
		"roleid":  roleID,
		"message": "Role deleted successfully",
		"result":  result,
	})
}

// listACLs handles the list_acl tool
func (s *Server) listACLs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_acl")

	acls, err := s.proxmoxClient.ListACLs(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list ACLs: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"acls":  acls,
		"count": len(acls),
	})
}

// setACL handles the set_acl tool
func (s *Server) setACL(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: set_acl")

	path := request.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	role := request.GetString("role", "")
	if role == "" {
		return mcp.NewToolResultError("role parameter is required"), nil
	}

	userID := request.GetString("userid", "")
	groupID := request.GetString("groupid", "")
	tokenID := request.GetString("tokenid", "")
	propagate := request.GetBool("propagate", true)

	if userID == "" && groupID == "" && tokenID == "" {
		return mcp.NewToolResultError("At least one of userid, groupid, or tokenid is required"), nil
	}

	result, err := s.proxmoxClient.SetACL(ctx, path, role, userID, groupID, tokenID, propagate)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set ACL: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":    "set_acl",
		"path":      path,
		"role":      role,
		"propagate": propagate,
		"message":   "ACL set successfully",
		"result":    result,
	})
}

// createAPIToken handles the create_api_token tool
func (s *Server) createAPIToken(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_api_token")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	tokenID := request.GetString("tokenid", "")
	if tokenID == "" {
		return mcp.NewToolResultError("tokenid parameter is required"), nil
	}

	expire := int64(request.GetInt("expire", 0))
	privSep := request.GetBool("privsep", false)

	result, err := s.proxmoxClient.CreateAPIToken(ctx, userID, tokenID, expire, privSep)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create API token: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "create",
		"userid":  userID,
		"tokenid": tokenID,
		"message": "API token created successfully",
		"result":  result,
	})
}

// deleteAPIToken handles the delete_api_token tool
func (s *Server) deleteAPIToken(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_api_token")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	tokenID := request.GetString("tokenid", "")
	if tokenID == "" {
		return mcp.NewToolResultError("tokenid parameter is required"), nil
	}

	result, err := s.proxmoxClient.DeleteAPIToken(ctx, userID, tokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete API token: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "delete",
		"userid":  userID,
		"tokenid": tokenID,
		"message": "API token deleted successfully",
		"result":  result,
	})
}

// ============ BACKUP & RESTORE ============

// listBackups handles the list_backups tool
func (s *Server) listBackups(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_backups")

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	backups, err := s.proxmoxClient.ListBackups(ctx, storage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list backups: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"backups": backups,
		"storage": storage,
		"count":   len(backups),
	})
}

// createVMBackup handles the create_vm_backup tool
func (s *Server) createVMBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_vm_backup")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID <= 0 {
		return mcp.NewToolResultError("vmid parameter is required and must be a positive integer"), nil
	}

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	backupID := request.GetString("backup_id", "")
	notes := request.GetString("notes", "")

	result, err := s.proxmoxClient.CreateVMBackup(ctx, nodeName, vmID, storage, backupID, notes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create VM backup: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":  "backup",
		"vmid":    vmID,
		"node":    nodeName,
		"storage": storage,
		"message": "VM backup started",
		"result":  result,
	})
}

// createContainerBackup handles the create_container_backup tool
func (s *Server) createContainerBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: create_container_backup")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID <= 0 {
		return mcp.NewToolResultError("container_id parameter is required and must be a positive integer"), nil
	}

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	backupID := request.GetString("backup_id", "")
	notes := request.GetString("notes", "")

	result, err := s.proxmoxClient.CreateContainerBackup(ctx, nodeName, containerID, storage, backupID, notes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create container backup: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":       "backup",
		"container_id": containerID,
		"node":         nodeName,
		"storage":      storage,
		"message":      "Container backup started",
		"result":       result,
	})
}

// deleteBackup handles the delete_backup tool
func (s *Server) deleteBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: delete_backup")

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	backupID := request.GetString("backup_id", "")
	if backupID == "" {
		return mcp.NewToolResultError("backup_id parameter is required"), nil
	}

	result, err := s.proxmoxClient.DeleteBackup(ctx, storage, backupID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete backup: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":    "delete",
		"storage":   storage,
		"backup_id": backupID,
		"message":   "Backup deleted successfully",
		"result":    result,
	})
}

// restoreVMBackup handles the restore_vm_backup tool
func (s *Server) restoreVMBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: restore_vm_backup")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	backupID := request.GetString("backup_id", "")
	if backupID == "" {
		return mcp.NewToolResultError("backup_id parameter is required"), nil
	}

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	result, err := s.proxmoxClient.RestoreVMBackup(ctx, nodeName, backupID, storage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore VM backup: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":    "restore",
		"node":      nodeName,
		"backup_id": backupID,
		"storage":   storage,
		"message":   "VM restore started",
		"result":    result,
	})
}

// restoreContainerBackup handles the restore_container_backup tool
func (s *Server) restoreContainerBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: restore_container_backup")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	backupID := request.GetString("backup_id", "")
	if backupID == "" {
		return mcp.NewToolResultError("backup_id parameter is required"), nil
	}

	storage := request.GetString("storage", "")
	if storage == "" {
		return mcp.NewToolResultError("storage parameter is required"), nil
	}

	result, err := s.proxmoxClient.RestoreContainerBackup(ctx, nodeName, backupID, storage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore container backup: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"action":    "restore",
		"node":      nodeName,
		"backup_id": backupID,
		"storage":   storage,
		"message":   "Container restore started",
		"result":    result,
	})
}

// ============ RESOURCE POOLS ============

func (s *Server) listPools(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_pools")

	pools, err := s.proxmoxClient.ListPools(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list pools: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "Pools retrieved successfully",
		"pools":   pools,
	})
}

func (s *Server) getPool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_pool")

	poolID := request.GetString("poolid", "")
	if poolID == "" {
		return mcp.NewToolResultError("poolid parameter is required"), nil
	}

	pool, err := s.proxmoxClient.GetPool(ctx, poolID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get pool: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "Pool retrieved successfully",
		"pool":    pool,
	})
}

// ============ NODE MANAGEMENT - TASKS ============

func (s *Server) getNodeTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_node_tasks")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	tasks, err := s.proxmoxClient.GetNodeTasks(ctx, nodeName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get node tasks: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "Node tasks retrieved successfully",
		"node":    nodeName,
		"tasks":   tasks,
	})
}

func (s *Server) getClusterTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_cluster_tasks")

	tasks, err := s.proxmoxClient.GetClusterTasks(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get cluster tasks: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "Cluster tasks retrieved successfully",
		"tasks":   tasks,
	})
}

// ============ STATISTICS ============

func (s *Server) getNodeStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_node_stats")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	stats, err := s.proxmoxClient.GetNodeStats(ctx, nodeName, "day")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get node statistics: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "Node statistics retrieved successfully",
		"node":    nodeName,
		"stats":   stats,
	})
}

func (s *Server) getVMStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_vm_stats")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	vmID := request.GetInt("vmid", 0)
	if vmID == 0 {
		return mcp.NewToolResultError("vmid parameter is required"), nil
	}

	stats, err := s.proxmoxClient.GetVMStats(ctx, nodeName, vmID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM statistics: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message": "VM statistics retrieved successfully",
		"node":    nodeName,
		"vmid":    vmID,
		"stats":   stats,
	})
}

func (s *Server) getContainerStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: get_container_stats")

	nodeName := request.GetString("node_name", "")
	if nodeName == "" {
		return mcp.NewToolResultError("node_name parameter is required"), nil
	}

	containerID := request.GetInt("container_id", 0)
	if containerID == 0 {
		return mcp.NewToolResultError("container_id parameter is required"), nil
	}

	stats, err := s.proxmoxClient.GetContainerStats(ctx, nodeName, containerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get container statistics: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"message":      "Container statistics retrieved successfully",
		"node":         nodeName,
		"container_id": containerID,
		"stats":        stats,
	})
}
