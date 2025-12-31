package proxmox

import (
	"context"
	"fmt"
)

// GetVMStats retrieves VM resource usage statistics
func (c *Client) GetVMStats(ctx context.Context, nodeName string, vmID int) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/qemu/%d/status/current", nodeName, vmID), nil)
	if err != nil {
		return nil, err
	}

	stats, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected VM stats format")
	}

	return stats, nil
}

// GetContainerStats retrieves container resource usage statistics
func (c *Client) GetContainerStats(ctx context.Context, nodeName string, containerID int) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/lxc/%d/status/current", nodeName, containerID), nil)
	if err != nil {
		return nil, err
	}

	stats, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected container stats format")
	}

	return stats, nil
}

// GetNodeStats retrieves node resource statistics over time
func (c *Client) GetNodeStats(ctx context.Context, nodeName string, timeframe string) (interface{}, error) {
	// timeframe can be "hour", "day", "week", "month", "year"
	params := map[string]interface{}{}
	if timeframe != "" {
		params["timeframe"] = timeframe
	}

	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("nodes/%s/rrddata", nodeName), params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetClusterStats retrieves cluster-wide resource statistics
func (c *Client) GetClusterStats(ctx context.Context, timeframe string) (interface{}, error) {
	// timeframe can be "hour", "day", "week", "month", "year"
	params := map[string]interface{}{}
	if timeframe != "" {
		params["timeframe"] = timeframe
	}

	data, err := c.doRequest(ctx, "GET", "cluster/resources", params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetAlerts retrieves active alerts and warnings from the cluster
func (c *Client) GetAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/health", nil)
	if err != nil {
		return nil, err
	}

	alerts := []map[string]interface{}{}
	if err := c.unmarshalData(data, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}
