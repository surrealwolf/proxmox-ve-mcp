package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Client handles communication with Proxmox VE API
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
	logger     *logrus.Entry
}

// APIResponse represents a standard Proxmox API response
type APIResponse struct {
	Data interface{} `json:"data"`
}

// Node represents a Proxmox node
type Node struct {
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	Uptime    int64   `json:"uptime"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Memory    int64   `json:"memory,omitempty"`
	MaxMemory int64   `json:"maxmemory,omitempty"`
	Disk      int64   `json:"disk,omitempty"`
	MaxDisk   int64   `json:"maxdisk,omitempty"`
}

// VM represents a virtual machine
type VM struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Node    string `json:"node"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	CPU     int    `json:"cpus,omitempty"`
	Memory  int64  `json:"memory,omitempty"`
	MaxDisk int64  `json:"maxdisk,omitempty"`
	Uptime  int64  `json:"uptime,omitempty"`
	PID     int    `json:"pid,omitempty"`
}

// Container represents an LXC container
type Container struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Node    string `json:"node"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	CPU     int    `json:"cpus,omitempty"`
	Memory  int64  `json:"memory,omitempty"`
	MaxDisk int64  `json:"maxdisk,omitempty"`
	Uptime  int64  `json:"uptime,omitempty"`
}

// Storage represents a storage device
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Enabled int    `json:"enabled,omitempty"`
	Used    int64  `json:"used,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

// MemoryInfo represents memory statistics
type MemoryInfo struct {
	Used      int64 `json:"used,omitempty"`
	Available int64 `json:"available,omitempty"`
	Total     int64 `json:"total,omitempty"`
	Free      int64 `json:"free,omitempty"`
}

// SwapInfo represents swap statistics
type SwapInfo struct {
	Used  int64 `json:"used,omitempty"`
	Total int64 `json:"total,omitempty"`
	Free  int64 `json:"free,omitempty"`
}

// RootfsInfo represents rootfs statistics
type RootfsInfo struct {
	Used  int64 `json:"used,omitempty"`
	Total int64 `json:"total,omitempty"`
	Free  int64 `json:"free,omitempty"`
	Avail int64 `json:"avail,omitempty"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Cores   int    `json:"cores,omitempty"`
	CPUs    int    `json:"cpus,omitempty"`
	Sockets int    `json:"sockets,omitempty"`
	MHz     string `json:"mhz,omitempty"`
	Model   string `json:"model,omitempty"`
}

// NodeStatus represents detailed node status from /nodes/{node}/status API
type NodeStatus struct {
	Uptime     int64      `json:"uptime,omitempty"`
	CPU        float64    `json:"cpu,omitempty"`
	Idle       float64    `json:"idle,omitempty"`
	Wait       float64    `json:"wait,omitempty"`
	Memory     MemoryInfo `json:"memory,omitempty"`
	Swap       SwapInfo   `json:"swap,omitempty"`
	Rootfs     RootfsInfo `json:"rootfs,omitempty"`
	CPUInfo    CPUInfo    `json:"cpuinfo,omitempty"`
	PVEVersion string     `json:"pveversion,omitempty"`
	KVersion   string     `json:"kversion,omitempty"`
	LoadAvg    []string   `json:"loadavg,omitempty"`
}

// Task represents a background task
type Task struct {
	ID        string `json:"id"`
	Node      string `json:"node"`
	PID       int    `json:"pid,omitempty"`
	PPID      int    `json:"ppid,omitempty"`
	Starttime int64  `json:"starttime,omitempty"`
	Type      string `json:"type,omitempty"`
	User      string `json:"user,omitempty"`
	Status    string `json:"status,omitempty"`
}

// Cluster represents cluster information
type Cluster struct {
	Name       string `json:"name"`
	NodesCount int    `json:"nodes,omitempty"`
	Version    string `json:"version,omitempty"`
}

// NewClient creates a new Proxmox VE API client
func NewClient(baseURL, apiToken string, skipSSLVerify bool) *Client {
	var tlsConfig *tls.Config
	if skipSSLVerify {
		// Disable SSL verification for self-signed certificates
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		httpClient: httpClient,
		logger:     logrus.WithField("component", "ProxmoxClient"),
	}
}

// doRequest performs an HTTP request to the Proxmox API
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (interface{}, error) {
	url := fmt.Sprintf("%s/api2/json/%s", c.baseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s", c.apiToken))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Data, nil
}

// GetNodes retrieves a list of all nodes in the cluster
func (c *Client) GetNodes(ctx context.Context) ([]Node, error) {
	data, err := c.doRequest(ctx, "GET", "nodes", nil)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	if err := json.Unmarshal(marshalJSON(data), &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	return nodes, nil
}

// GetNode retrieves information about a specific node
func (c *Client) GetNode(ctx context.Context, nodeName string) (*NodeStatus, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/status", nodeName), nil)
	if err != nil {
		return nil, err
	}

	var nodeStatus NodeStatus
	if err := json.Unmarshal(marshalJSON(data), &nodeStatus); err != nil {
		return nil, fmt.Errorf("failed to parse node: %w", err)
	}

	return &nodeStatus, nil
}

// GetClusterStatus retrieves cluster-wide status information
func (c *Client) GetClusterStatus(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}

	var status map[string]interface{}
	if jsonMap, ok := data.(map[string]interface{}); ok {
		status = jsonMap
	} else {
		return nil, fmt.Errorf("invalid cluster status response format")
	}

	return status, nil
}

// GetVMs retrieves a list of VMs on a specific node
func (c *Client) GetVMs(ctx context.Context, nodeName string) ([]VM, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/qemu", nodeName), nil)
	if err != nil {
		return nil, err
	}

	var vms []VM
	if err := json.Unmarshal(marshalJSON(data), &vms); err != nil {
		return nil, fmt.Errorf("failed to parse VMs: %w", err)
	}

	return vms, nil
}

// GetVM retrieves information about a specific VM
func (c *Client) GetVM(ctx context.Context, nodeName string, vmID int) (*VM, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/qemu/%d/status/current", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	var vm VM
	if err := json.Unmarshal(marshalJSON(data), &vm); err != nil {
		return nil, fmt.Errorf("failed to parse VM: %w", err)
	}

	return &vm, nil
}

// GetContainers retrieves a list of containers on a specific node
func (c *Client) GetContainers(ctx context.Context, nodeName string) ([]Container, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/lxc", nodeName), nil)
	if err != nil {
		return nil, err
	}

	var containers []Container
	if err := json.Unmarshal(marshalJSON(data), &containers); err != nil {
		return nil, fmt.Errorf("failed to parse containers: %w", err)
	}

	return containers, nil
}

// GetContainer retrieves information about a specific container
func (c *Client) GetContainer(ctx context.Context, nodeName string, containerID int) (*Container, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/lxc/%d/status/current", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	var container Container
	if err := json.Unmarshal(marshalJSON(data), &container); err != nil {
		return nil, fmt.Errorf("failed to parse container: %w", err)
	}

	return &container, nil
}

// GetStorage retrieves storage information
func (c *Client) GetStorage(ctx context.Context) ([]Storage, error) {
	data, err := c.doRequest(ctx, "GET", "storage", nil)
	if err != nil {
		return nil, err
	}

	var storage []Storage
	if err := json.Unmarshal(marshalJSON(data), &storage); err != nil {
		return nil, fmt.Errorf("failed to parse storage: %w", err)
	}

	return storage, nil
}

// GetNodeStorage retrieves storage for a specific node
func (c *Client) GetNodeStorage(ctx context.Context, nodeName string) ([]Storage, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/storage", nodeName), nil)
	if err != nil {
		return nil, err
	}

	var storage []Storage
	if err := json.Unmarshal(marshalJSON(data), &storage); err != nil {
		return nil, fmt.Errorf("failed to parse node storage: %w", err)
	}

	return storage, nil
}

// GetTasks retrieves cluster tasks
func (c *Client) GetTasks(ctx context.Context) ([]Task, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/resources", nil)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(marshalJSON(data), &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	return tasks, nil
}

// GetClusterResources retrieves all cluster resources
func (c *Client) GetClusterResources(ctx context.Context) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/resources", nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StartVM starts a virtual machine
func (c *Client) StartVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/start", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StopVM stops a virtual machine
func (c *Client) StopVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/stop", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// RebootVM reboots a virtual machine
func (c *Client) RebootVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/reboot", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ShutdownVM shuts down a virtual machine gracefully
func (c *Client) ShutdownVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/shutdown", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StartContainer starts an LXC container
func (c *Client) StartContainer(ctx context.Context, nodeName string, containerID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/status/start", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StopContainer stops an LXC container
func (c *Client) StopContainer(ctx context.Context, nodeName string, containerID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/status/stop", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ShutdownContainer shuts down an LXC container gracefully
func (c *Client) ShutdownContainer(ctx context.Context, nodeName string, containerID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/status/shutdown", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// RebootContainer reboots an LXC container
func (c *Client) RebootContainer(ctx context.Context, nodeName string, containerID int) (interface{}, error) {
	data, err := c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/status/reboot", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetContainerConfig returns the configuration of a specific container
func (c *Client) GetContainerConfig(ctx context.Context, nodeName string, containerID int) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/lxc/%d/config", nodeName, containerID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get container config: %w", err)
	}

	var config map[string]interface{}
	if jsonMap, ok := data.(map[string]interface{}); ok {
		config = jsonMap
	} else {
		return nil, fmt.Errorf("invalid container config response format")
	}

	return config, nil
}

// DeleteContainer removes an LXC container
func (c *Client) DeleteContainer(ctx context.Context, nodeName string, containerID int, force bool) (interface{}, error) {
	body := map[string]interface{}{}
	if force {
		body["force"] = 1
	}

	return c.doRequest(ctx, "DELETE", fmt.Sprintf("nodes/%s/lxc/%d", nodeName, containerID), body)
}

// CreateContainer creates a new LXC container with the specified configuration
func (c *Client) CreateContainer(ctx context.Context, nodeName string, config map[string]interface{}) (interface{}, error) {
	// Ensure vmid is present in the config
	if _, ok := config["vmid"]; !ok {
		return nil, fmt.Errorf("vmid is required in container configuration")
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc", nodeName), config)
}

// CreateContainerFull creates a new container with detailed configuration
// Params: vmid, hostname, storage, memory, cores, ostype (e.g., "debian"), osversion
func (c *Client) CreateContainerFull(ctx context.Context, nodeName string, containerID int, hostname string, storage string, memory int, cores int, ostype string) (interface{}, error) {
	config := map[string]interface{}{
		"vmid":      containerID,
		"hostname":  hostname,
		"storage":   storage,
		"memory":    memory,
		"cores":     cores,
		"ostype":    ostype,
	}
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc", nodeName), config)
}

// CloneContainer clones an existing LXC container
func (c *Client) CloneContainer(ctx context.Context, nodeName string, sourceContainerID int, newContainerID int, newHostname string, full bool) (interface{}, error) {
	config := map[string]interface{}{
		"vmid":     newContainerID,
		"hostname": newHostname,
		"full":     full,
	}
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/clone", nodeName, sourceContainerID), config)
}

// GetVMConfig returns the configuration of a specific VM
func (c *Client) GetVMConfig(ctx context.Context, nodeName string, vmID int) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/qemu/%d/config", nodeName, vmID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM config: %w", err)
	}

	var config map[string]interface{}
	if jsonMap, ok := data.(map[string]interface{}); ok {
		config = jsonMap
	} else {
		return nil, fmt.Errorf("invalid VM config response format")
	}

	return config, nil
}

// DeleteVM removes a virtual machine
func (c *Client) DeleteVM(ctx context.Context, nodeName string, vmID int, force bool) (interface{}, error) {
	body := map[string]interface{}{}
	if force {
		body["force"] = 1
	}

	return c.doRequest(ctx, "DELETE", fmt.Sprintf("nodes/%s/qemu/%d", nodeName, vmID), body)
}

// SuspendVM pauses a running VM without shutdown
func (c *Client) SuspendVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/suspend", nodeName, vmID), nil)
}

// ResumeVM resumes a suspended VM
func (c *Client) ResumeVM(ctx context.Context, nodeName string, vmID int) (interface{}, error) {
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/status/resume", nodeName, vmID), nil)
}

// CreateVM creates a new virtual machine with the specified configuration
func (c *Client) CreateVM(ctx context.Context, nodeName string, config map[string]interface{}) (interface{}, error) {
	// Ensure vmid is present in the config
	if _, ok := config["vmid"]; !ok {
		return nil, fmt.Errorf("vmid is required in VM configuration")
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu", nodeName), config)
}

// CreateVMFull creates a new VM with detailed configuration
// Params: vmid, name, memory, cores, sockets, node, ide2 (ISO), sata0 (disk storage), net0 (network)
func (c *Client) CreateVMFull(ctx context.Context, nodeName string, vmID int, name string, memory int, cores int, sockets int) (interface{}, error) {
	config := map[string]interface{}{
		"vmid":    vmID,
		"name":    name,
		"memory":  memory,
		"cores":   cores,
		"sockets": sockets,
	}
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu", nodeName), config)
}

// CloneVM clones an existing virtual machine
func (c *Client) CloneVM(ctx context.Context, nodeName string, sourceVMID int, newVMID int, newName string, full bool) (interface{}, error) {
	config := map[string]interface{}{
		"vmid":   newVMID,
		"name":   newName,
		"full":   full,
	}
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/clone", nodeName, sourceVMID), config)
}

// ============ USER & ACCESS MANAGEMENT ============

// User represents a Proxmox user
type User struct {
	UserID    string   `json:"userid"`
	Enable    int      `json:"enable,omitempty"`
	Expire    int64    `json:"expire,omitempty"`
	FirstName string   `json:"firstname,omitempty"`
	LastName  string   `json:"lastname,omitempty"`
	Email     string   `json:"email,omitempty"`
	Comment   string   `json:"comment,omitempty"`
	Groups    []string `json:"groups,omitempty"`
}

// Group represents a Proxmox group
type Group struct {
	GroupID string `json:"groupid"`
	Comment string `json:"comment,omitempty"`
}

// Role represents a Proxmox role
type Role struct {
	RoleID  string `json:"roleid"`
	Privs   string `json:"privs,omitempty"`
	Special int    `json:"special,omitempty"`
}

// APIToken represents a Proxmox API token
type APIToken struct {
	TokenID string `json:"tokenid"`
	Expire  int64  `json:"expire,omitempty"`
	PrivSep int    `json:"privsep,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ACLEntry represents an access control list entry
type ACLEntry struct {
	Path      string `json:"path"`
	Role      string `json:"role"`
	User      string `json:"user,omitempty"`
	Group     string `json:"group,omitempty"`
	Token     string `json:"token,omitempty"`
	Propagate int    `json:"propagate,omitempty"`
}

// ListUsers returns all users
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	data, err := c.doRequest(ctx, "GET", "access/users", nil)
	if err != nil {
		return nil, err
	}

	users := []User{}
	if err := c.unmarshalData(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUser returns a specific user
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("access/users/%s", userID), nil)
	if err != nil {
		return nil, err
	}

	user := &User{}
	if err := c.unmarshalData(data, user); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUser creates a new user
func (c *Client) CreateUser(ctx context.Context, userID, password, email, comment string) (interface{}, error) {
	body := map[string]interface{}{
		"userid":   userID,
		"password": password,
	}
	if email != "" {
		body["email"] = email
	}
	if comment != "" {
		body["comment"] = comment
	}

	return c.doRequest(ctx, "POST", "access/users", body)
}

// UpdateUser updates user properties
func (c *Client) UpdateUser(ctx context.Context, userID, email, comment, firstName, lastName string, enable bool, expire int64) (interface{}, error) {
	body := map[string]interface{}{
		"userid": userID,
	}
	if email != "" {
		body["email"] = email
	}
	if comment != "" {
		body["comment"] = comment
	}
	if firstName != "" {
		body["firstname"] = firstName
	}
	if lastName != "" {
		body["lastname"] = lastName
	}
	body["enable"] = boolToInt(enable)
	if expire > 0 {
		body["expire"] = expire
	}

	return c.doRequest(ctx, "PUT", fmt.Sprintf("access/users/%s", userID), body)
}

// DeleteUser removes a user
func (c *Client) DeleteUser(ctx context.Context, userID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("access/users/%s", userID), nil)
}

// ChangePassword changes a user's password
func (c *Client) ChangePassword(ctx context.Context, userID, password string) (interface{}, error) {
	body := map[string]interface{}{
		"userid":   userID,
		"password": password,
	}
	return c.doRequest(ctx, "PUT", "access/password", body)
}

// ListGroups returns all groups
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	data, err := c.doRequest(ctx, "GET", "access/groups", nil)
	if err != nil {
		return nil, err
	}

	groups := []Group{}
	if err := c.unmarshalData(data, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, groupID, comment string) (interface{}, error) {
	body := map[string]interface{}{
		"groupid": groupID,
	}
	if comment != "" {
		body["comment"] = comment
	}

	return c.doRequest(ctx, "POST", "access/groups", body)
}

// DeleteGroup removes a group
func (c *Client) DeleteGroup(ctx context.Context, groupID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("access/groups/%s", groupID), nil)
}

// ListRoles returns all roles
func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	data, err := c.doRequest(ctx, "GET", "access/roles", nil)
	if err != nil {
		return nil, err
	}

	roles := []Role{}
	if err := c.unmarshalData(data, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// CreateRole creates a new role with specified privileges
func (c *Client) CreateRole(ctx context.Context, roleID string, privs []string) (interface{}, error) {
	// Convert privileges array to comma-separated string
	privsStr := ""
	if len(privs) > 0 {
		for i, priv := range privs {
			if i > 0 {
				privsStr += ","
			}
			privsStr += priv
		}
	}

	body := map[string]interface{}{
		"roleid": roleID,
		"privs":  privsStr,
	}

	return c.doRequest(ctx, "POST", "access/roles", body)
}

// DeleteRole removes a role
func (c *Client) DeleteRole(ctx context.Context, roleID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("access/roles/%s", roleID), nil)
}

// ListACLs returns all ACL entries
func (c *Client) ListACLs(ctx context.Context) ([]ACLEntry, error) {
	data, err := c.doRequest(ctx, "GET", "access/acl", nil)
	if err != nil {
		return nil, err
	}

	acls := []ACLEntry{}
	if err := c.unmarshalData(data, &acls); err != nil {
		return nil, err
	}

	return acls, nil
}

// SetACL creates or updates an ACL entry
func (c *Client) SetACL(ctx context.Context, path, role, userID, groupID, tokenID string, propagate bool) (interface{}, error) {
	body := map[string]interface{}{
		"path": path,
		"role": role,
	}
	if userID != "" {
		body["user"] = userID
	}
	if groupID != "" {
		body["group"] = groupID
	}
	if tokenID != "" {
		body["token"] = tokenID
	}
	body["propagate"] = boolToInt(propagate)

	return c.doRequest(ctx, "PUT", "access/acl", body)
}

// ListAPITokens returns API tokens for a user
func (c *Client) ListAPITokens(ctx context.Context, userID string) ([]APIToken, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("access/users/%s/tokens", userID), nil)
	if err != nil {
		return nil, err
	}

	tokens := []APIToken{}
	if err := c.unmarshalData(data, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// CreateAPIToken creates a new API token for a user
func (c *Client) CreateAPIToken(ctx context.Context, userID, tokenID string, expire int64, privSep bool) (interface{}, error) {
	body := map[string]interface{}{
		"tokenid": tokenID,
	}
	if expire > 0 {
		body["expire"] = expire
	}
	body["privsep"] = boolToInt(privSep)

	return c.doRequest(ctx, "POST", fmt.Sprintf("access/users/%s/tokens/%s", userID, tokenID), body)
}

// DeleteAPIToken removes an API token
func (c *Client) DeleteAPIToken(ctx context.Context, userID, tokenID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("access/users/%s/tokens/%s", userID, tokenID), nil)
}

// ============ BACKUP & RESTORE ============

// Backup represents a backup file
type Backup struct {
	BackupID  string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	VMID      int    `json:"vmid,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Notes     string `json:"notes,omitempty"`
	CTime     int64  `json:"ctime,omitempty"`
	Content   string `json:"content,omitempty"`
	Verified  int    `json:"verified,omitempty"`
	Encrypted int    `json:"encrypted,omitempty"`
	Nodes     string `json:"nodes,omitempty"`
}

// CreateVMBackup creates a backup of a virtual machine
func (c *Client) CreateVMBackup(ctx context.Context, nodeName string, vmID int, storage, backupID, notes string) (interface{}, error) {
	body := map[string]interface{}{
		"storage": storage,
		"vmid":    vmID,
	}
	if backupID != "" {
		body["id"] = backupID
	}
	if notes != "" {
		body["notes"] = notes
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu/%d/backup", nodeName, vmID), body)
}

// CreateContainerBackup creates a backup of a container
func (c *Client) CreateContainerBackup(ctx context.Context, nodeName string, containerID int, storage, backupID, notes string) (interface{}, error) {
	body := map[string]interface{}{
		"storage": storage,
		"vmid":    containerID,
	}
	if backupID != "" {
		body["id"] = backupID
	}
	if notes != "" {
		body["notes"] = notes
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc/%d/backup", nodeName, containerID), body)
}

// ListBackups returns available backups in storage across all nodes
func (c *Client) ListBackups(ctx context.Context, storage string) ([]Backup, error) {
	// Get all nodes first
	nodes, err := c.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	var allBackups []Backup

	// Try to get backups from each node's storage
	for _, node := range nodes {
		data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/storage/%s/content", node.Node, storage), nil)
		if err != nil {
			// Log error but continue with other nodes
			c.logger.Warnf("Failed to list backups from node %s: %v", node.Node, err)
			continue
		}

		backups := []Backup{}
		if err := c.unmarshalData(data, &backups); err != nil {
			c.logger.Warnf("Failed to unmarshal backups from node %s: %v", node.Node, err)
			continue
		}

		allBackups = append(allBackups, backups...)
	}

	return allBackups, nil
}

// DeleteBackup removes a backup file from a specific node's storage
func (c *Client) DeleteBackup(ctx context.Context, storage, backupID string) (interface{}, error) {
	// Get all nodes to find which one has the backup
	nodes, err := c.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	var lastErr error

	// Try to delete backup from each node
	for _, node := range nodes {
		result, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("nodes/%s/storage/%s/content/%s", node.Node, storage, backupID), nil)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("backup not found on any node")
}

// RestoreVMBackup restores a VM from a backup
func (c *Client) RestoreVMBackup(ctx context.Context, nodeName string, backupID, storage string) (interface{}, error) {
	body := map[string]interface{}{
		"archive": backupID,
		"storage": storage,
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/qemu", nodeName), body)
}

// RestoreContainerBackup restores a container from a backup
func (c *Client) RestoreContainerBackup(ctx context.Context, nodeName string, backupID, storage string) (interface{}, error) {
	body := map[string]interface{}{
		"archive": backupID,
		"storage": storage,
	}

	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/lxc", nodeName), body)
}

// Helper functions

// boolToInt converts boolean to int (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// unmarshalData converts interface{} to specific type
func (c *Client) unmarshalData(data interface{}, result interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, result)
}

// marshalJSON is a helper to convert interface{} to JSON bytes
func marshalJSON(data interface{}) []byte {
	jsonData, _ := json.Marshal(data)
	return jsonData
}
