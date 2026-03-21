package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CreateRoleTemplate(ctx context.Context, req CreateRoleTemplateRequest) (*RoleTemplate, error) {
	var rt RoleTemplate
	if err := c.Post(ctx, "/api/v1/admin/role-templates", req, &rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (c *Client) GetRoleTemplate(ctx context.Context, id string) (*RoleTemplate, error) {
	var rt RoleTemplate
	if err := c.Get(ctx, "/api/v1/admin/role-templates/"+id, &rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (c *Client) UpdateRoleTemplate(ctx context.Context, id string, req UpdateRoleTemplateRequest) (*RoleTemplate, error) {
	var rt RoleTemplate
	if err := c.Put(ctx, "/api/v1/admin/role-templates/"+id, req, &rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (c *Client) DeleteRoleTemplate(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/role-templates/"+id)
}

func (c *Client) ListRoleTemplates(ctx context.Context) ([]RoleTemplate, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/v1/admin/role-templates", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseResponseError(resp)
	}

	var templates []RoleTemplate
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, fmt.Errorf("decoding role templates: %w", err)
	}
	return templates, nil
}
