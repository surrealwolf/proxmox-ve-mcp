package proxmox

import (
	"context"
	"fmt"
)

// GetStorageInfo retrieves detailed information about a specific storage device
func (c *Client) GetStorageInfo(ctx context.Context, storage string) (map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("storage/%s", storage), nil)
	if err != nil {
		return nil, err
	}

	info, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected storage info format")
	}

	return info, nil
}

// GetStorageContent lists the contents of a storage device (ISOs, backups, templates, etc.)
func (c *Client) GetStorageContent(ctx context.Context, storage string) ([]map[string]interface{}, error) {
	result, err := c.doRequest(ctx, "GET", fmt.Sprintf("storage/%s/content", storage), nil)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []map[string]interface{}{}, nil
	}

	content := []map[string]interface{}{}
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if content_item, ok := item.(map[string]interface{}); ok {
				content = append(content, content_item)
			}
		}
	}
	return content, nil
}

// CreateStorage creates a new storage mount
func (c *Client) CreateStorage(ctx context.Context, storage, storageType, content string, config map[string]interface{}) (interface{}, error) {
	body := map[string]interface{}{
		"storage": storage,
		"type":    storageType,
		"content": content,
	}
	// Merge additional config
	for k, v := range config {
		body[k] = v
	}

	return c.doRequest(ctx, "POST", "storage", body)
}

// DeleteStorage removes a storage configuration
func (c *Client) DeleteStorage(ctx context.Context, storage string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("storage/%s", storage), nil)
}

// UpdateStorage modifies storage configuration
func (c *Client) UpdateStorage(ctx context.Context, storage string, config map[string]interface{}) (interface{}, error) {
	body := config
	body["storage"] = storage
	return c.doRequest(ctx, "PUT", fmt.Sprintf("storage/%s", storage), body)
}
