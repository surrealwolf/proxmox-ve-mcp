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

// NodeStatus represents detailed node status
type NodeStatus struct {
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	Uptime    int64   `json:"uptime,omitempty"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Memory    int64   `json:"memory,omitempty"`
	MaxMemory int64   `json:"maxmemory,omitempty"`
	Disk      int64   `json:"disk,omitempty"`
	MaxDisk   int64   `json:"maxdisk,omitempty"`
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
	req.Header.Set("Content-Type", "application/json")

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
func (c *Client) GetNode(ctx context.Context, nodeName string) (*Node, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/status", nodeName), nil)
	if err != nil {
		return nil, err
	}

	var node Node
	if err := json.Unmarshal(marshalJSON(data), &node); err != nil {
		return nil, fmt.Errorf("failed to parse node: %w", err)
	}

	return &node, nil
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

// marshalJSON is a helper to convert interface{} to JSON bytes
func marshalJSON(data interface{}) []byte {
	jsonData, _ := json.Marshal(data)
	return jsonData
}
