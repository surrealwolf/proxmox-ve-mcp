package proxmox

import (
	"context"
	"fmt"
)

// GetClusterConfig retrieves cluster configuration
func (c *Client) GetClusterConfig(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/config", nil)
	if err != nil {
		return nil, err
	}

	config, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected cluster config format")
	}

	return config, nil
}

// GetClusterNodesStatus retrieves status of all nodes in the cluster
func (c *Client) GetClusterNodesStatus(ctx context.Context) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/status", nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetHAStatus retrieves High Availability status
func (c *Client) GetHAStatus(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/ha/status/current", nil)
	if err != nil {
		return nil, err
	}

	status, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected HA status format")
	}

	return status, nil
}

// EnableHAResource enables High Availability for a resource
func (c *Client) EnableHAResource(ctx context.Context, sid, comment string, state string) (interface{}, error) {
	body := map[string]interface{}{
		"sid": sid,
	}
	if comment != "" {
		body["comment"] = comment
	}
	if state != "" {
		body["state"] = state
	}

	return c.doRequest(ctx, "POST", "cluster/ha/resources", body)
}

// DisableHAResource disables High Availability for a resource
func (c *Client) DisableHAResource(ctx context.Context, sid string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("cluster/ha/resources/%s", sid), nil)
}
