package proxmox

import (
	"context"
	"fmt"
)

// Pool represents a Proxmox resource pool
type Pool struct {
	PoolID  string   `json:"poolid"`
	Comment string   `json:"comment,omitempty"`
	Members []string `json:"members,omitempty"`
	Guests  []string `json:"guests,omitempty"`
	Storage []string `json:"storage,omitempty"`
}

// ListPools lists all resource pools
func (c *Client) ListPools(ctx context.Context) ([]Pool, error) {
	data, err := c.doRequest(ctx, "GET", "pools", nil)
	if err != nil {
		return nil, err
	}

	pools := []Pool{}
	if err := c.unmarshalData(data, &pools); err != nil {
		return nil, err
	}

	return pools, nil
}

// GetPool retrieves details of a specific pool
func (c *Client) GetPool(ctx context.Context, poolID string) (*Pool, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("pools/%s", poolID), nil)
	if err != nil {
		return nil, err
	}

	pool := &Pool{}
	if err := c.unmarshalData(data, pool); err != nil {
		return nil, err
	}

	return pool, nil
}

// CreatePool creates a new resource pool
func (c *Client) CreatePool(ctx context.Context, poolID, comment string, members []string) (interface{}, error) {
	body := map[string]interface{}{
		"poolid": poolID,
	}
	if comment != "" {
		body["comment"] = comment
	}
	if len(members) > 0 {
		body["members"] = members
	}

	return c.doRequest(ctx, "POST", "pools", body)
}

// UpdatePool modifies an existing resource pool
func (c *Client) UpdatePool(ctx context.Context, poolID, comment string, members []string, delete bool) (interface{}, error) {
	body := map[string]interface{}{
		"poolid": poolID,
	}
	if comment != "" {
		body["comment"] = comment
	}
	if len(members) > 0 {
		body["members"] = members
	}
	if delete {
		body["delete"] = 1
	}

	return c.doRequest(ctx, "PUT", fmt.Sprintf("pools/%s", poolID), body)
}

// DeletePool removes a resource pool
func (c *Client) DeletePool(ctx context.Context, poolID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("pools/%s", poolID), nil)
}

// GetPoolMembers retrieves all resources in a pool
func (c *Client) GetPoolMembers(ctx context.Context, poolID string) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("pools/%s", poolID), nil)
	if err != nil {
		return nil, err
	}

	members, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected pool members format")
	}

	return members, nil
}
