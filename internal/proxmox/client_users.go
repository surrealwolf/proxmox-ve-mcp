package proxmox

import (
	"context"
	"fmt"
)

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
