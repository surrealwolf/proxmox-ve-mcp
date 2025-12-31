package mcp

import (
	"context"
	"fmt"

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

	// User Management - Query
	addTool("list_users", "List all users in the system", s.listUsers, map[string]any{})
	addTool("get_user", "Get details for a specific user", s.getUser, map[string]any{
		"userid": map[string]any{"type": "string", "description": "User ID (e.g., user@pve)"},
	})
	addTool("list_groups", "List all groups", s.listGroups, map[string]any{})
	addTool("list_roles", "List all available roles and their privileges", s.listRoles, map[string]any{})
	addTool("list_acl", "List all access control list entries", s.listACLs, map[string]any{})
	addTool("list_api_tokens", "List API tokens for a specific user", s.listAPITokens, map[string]any{
		"userid": map[string]any{"type": "string", "description": "User ID"},
	})

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

	for _, tool := range tools {
		s.server.AddTool(tool.Tool, tool.Handler)
	}
	s.logger.Info("Registered 48 tools")
}

// ServeStdio starts the MCP server with stdio transport
func (s *Server) ServeStdio(ctx context.Context) error {
	s.logger.Info("Starting Proxmox VE MCP Server")
	return server.ServeStdio(s.server)
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

// listAPITokens handles the list_api_tokens tool
func (s *Server) listAPITokens(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Tool called: list_api_tokens")

	userID := request.GetString("userid", "")
	if userID == "" {
		return mcp.NewToolResultError("userid parameter is required"), nil
	}

	tokens, err := s.proxmoxClient.ListAPITokens(ctx, userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list API tokens: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"tokens": tokens,
		"userid": userID,
		"count":  len(tokens),
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
