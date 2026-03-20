package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateRoleTemplate(ctx context.Context, req CreateRoleTemplateRequest) (*RoleTemplate, error) {
	var resp struct {
		RoleTemplate RoleTemplate `json:"role_template"`
	}
	if err := c.Post(ctx, "/api/v1/admin/role-templates", req, &resp); err != nil {
		return nil, err
	}
	return &resp.RoleTemplate, nil
}

func (c *Client) GetRoleTemplate(ctx context.Context, id string) (*RoleTemplate, error) {
	var resp struct {
		RoleTemplate RoleTemplate `json:"role_template"`
	}
	if err := c.Get(ctx, "/api/v1/admin/role-templates/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.RoleTemplate, nil
}

func (c *Client) UpdateRoleTemplate(ctx context.Context, id string, req UpdateRoleTemplateRequest) (*RoleTemplate, error) {
	var resp struct {
		RoleTemplate RoleTemplate `json:"role_template"`
	}
	if err := c.Put(ctx, "/api/v1/admin/role-templates/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.RoleTemplate, nil
}

func (c *Client) DeleteRoleTemplate(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/role-templates/"+id)
}

func (c *Client) ListRoleTemplates(ctx context.Context) ([]RoleTemplate, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/admin/role-templates", "role_templates")
	if err != nil {
		return nil, err
	}

	templates := make([]RoleTemplate, 0, len(items))
	for _, raw := range items {
		var t RoleTemplate
		if err := json.Unmarshal(raw, &t); err != nil {
			return nil, fmt.Errorf("unmarshaling role template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, nil
}
