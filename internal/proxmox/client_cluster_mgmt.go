package proxmox

import (
	"context"
	"fmt"
)

// GetClusterConfig retrieves cluster configuration
func (c *Client) GetClusterConfig(ctx context.Context) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/config", nil)
	if err != nil {
		return nil, err
	}

	// Handle both map and array formats
	return data, nil
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
func (c *Client) GetHAStatus(ctx context.Context) (interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/ha/status/current", nil)
	if err != nil {
		return nil, err
	}

	// Handle both map and array formats
	return data, nil
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

// AddNodeToCluster adds a node to the cluster
func (c *Client) AddNodeToCluster(ctx context.Context, nodeName, clusterName, clusterNetwork string) (interface{}, error) {
	body := map[string]interface{}{
		"nodeid": nodeName,
	}
	if clusterName != "" {
		body["clustername"] = clusterName
	}
	if clusterNetwork != "" {
		body["clusternetwork"] = clusterNetwork
	}

	return c.doRequest(ctx, "POST", "cluster/nodes", body)
}

// RemoveNodeFromCluster removes a node from the cluster
func (c *Client) RemoveNodeFromCluster(ctx context.Context, nodeName string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("cluster/nodes/%s", nodeName), nil)
}
