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

	for _, tool := range tools {
		s.server.AddTool(tool.Tool, tool.Handler)
	}
	s.logger.Info("Registered 21 tools")
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
