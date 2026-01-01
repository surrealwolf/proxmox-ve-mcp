package proxmox

import (
	"context"
	"fmt"
)

// GetNodeConfig retrieves node network and system configuration
func (c *Client) GetNodeConfig(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/config", nodeName), nil)
	if err != nil {
		return nil, err
	}

	config, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected node config format")
	}

	return config, nil
}

// UpdateNodeConfig modifies node settings
func (c *Client) UpdateNodeConfig(ctx context.Context, nodeName string, config map[string]interface{}) (interface{}, error) {
	return c.doRequest(ctx, "PUT", fmt.Sprintf("nodes/%s/config", nodeName), config)
}

// RebootNode reboots a node
func (c *Client) RebootNode(ctx context.Context, nodeName string) (interface{}, error) {
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/status/reboot", nodeName), nil)
}

// ShutdownNode gracefully shuts down a node
func (c *Client) ShutdownNode(ctx context.Context, nodeName string) (interface{}, error) {
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/status/shutdown", nodeName), nil)
}

// GetNodeDisks lists physical disks in a node
func (c *Client) GetNodeDisks(ctx context.Context, nodeName string) ([]map[string]interface{}, error) {
	result, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/disks", nodeName), nil)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []map[string]interface{}{}, nil
	}

	disks := []map[string]interface{}{}
	if data, ok := result.(map[string]interface{}); ok {
		if disksList, ok := data["disks"].([]interface{}); ok {
			for _, item := range disksList {
				if disk, ok := item.(map[string]interface{}); ok {
					disks = append(disks, disk)
				}
			}
		}
	}
	return disks, nil
}

// GetNodeCert retrieves SSL certificate information for a node
func (c *Client) GetNodeCert(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/certificates/info", nodeName), nil)
	if err != nil {
		return nil, err
	}

	cert, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected certificate info format")
	}

	return cert, nil
}

// GetNodeLogs retrieves node system logs
func (c *Client) GetNodeLogs(ctx context.Context, nodeName string, lines int) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	if lines > 0 {
		params["lines"] = lines
	}

	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/syslog", nodeName), params)
	if err != nil {
		return nil, err
	}

	logs, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected syslog format")
	}

	return logs, nil
}

// GetNodeAPTUpdates checks available package updates
func (c *Client) GetNodeAPTUpdates(ctx context.Context, nodeName string) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/apt/update", nodeName), nil)
	if err != nil {
		return nil, err
	}

	// Handle both map and array formats
	return data, nil
}

// ApplyNodeUpdates installs system updates on a node
func (c *Client) ApplyNodeUpdates(ctx context.Context, nodeName string) (interface{}, error) {
	return c.doRequest(ctx, "POST", fmt.Sprintf("nodes/%s/apt/update", nodeName), nil)
}

// GetNodeNetwork retrieves detailed network configuration
func (c *Client) GetNodeNetwork(ctx context.Context, nodeName string) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/network", nodeName), nil)
	if err != nil {
		return nil, err
	}

	// API returns array of network interfaces
	if _, ok := data.([]interface{}); ok {
		return data, nil
	}

	// Also accept map format for backwards compatibility
	if _, ok := data.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, fmt.Errorf("unexpected network format: expected array or map")
}

// GetNodeDNS retrieves DNS configuration
func (c *Client) GetNodeDNS(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/dns", nodeName), nil)
	if err != nil {
		return nil, err
	}

	dns, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected DNS format")
	}

	return dns, nil
}
