package proxmox

import (
	"context"
	"fmt"
)

// ListTasks lists all background tasks
func (c *Client) ListTasks(ctx context.Context) ([]Task, error) {
	data, err := c.doRequest(ctx, "GET", "cluster/tasks", nil)
	if err != nil {
		return nil, err
	}

	tasks := []Task{}
	if err := c.unmarshalData(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTaskStatus retrieves detailed status and progress of a task
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (map[string]interface{}, error) {
	// TaskID format is typically "UPID:node:pid:start_time:type:user"
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("cluster/tasks/%s", taskID), nil)
	if err != nil {
		return nil, err
	}

	status, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected task status format")
	}

	return status, nil
}

// GetTaskLog retrieves the log output for a task
func (c *Client) GetTaskLog(ctx context.Context, taskID string, start, limit int) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	if start > 0 {
		params["start"] = start
	}
	if limit > 0 {
		params["limit"] = limit
	}

	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("cluster/tasks/%s/log", taskID), params)
	if err != nil {
		return nil, err
	}

	log, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected task log format")
	}

	return log, nil
}

// CancelTask cancels a running task
func (c *Client) CancelTask(ctx context.Context, taskID string) (interface{}, error) {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("cluster/tasks/%s", taskID), nil)
}

// GetNodeTasks gets all tasks for a specific node
func (c *Client) GetNodeTasks(ctx context.Context, nodeName string) ([]Task, error) {
	tasks, err := c.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Filter tasks for the specific node
	var nodeTasks []Task
	for _, task := range tasks {
		if task.Node == nodeName {
			nodeTasks = append(nodeTasks, task)
		}
	}

	return nodeTasks, nil
}

// GetClusterTasks gets all cluster tasks
func (c *Client) GetClusterTasks(ctx context.Context) ([]Task, error) {
	return c.ListTasks(ctx)
}
