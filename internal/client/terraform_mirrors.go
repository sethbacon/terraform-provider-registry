package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateTerraformMirror(ctx context.Context, req CreateTerraformMirrorRequest) (*TerraformMirror, error) {
	var mirror TerraformMirror
	if err := c.Post(ctx, "/api/v1/admin/terraform-mirrors", req, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) GetTerraformMirror(ctx context.Context, id string) (*TerraformMirror, error) {
	var mirror TerraformMirror
	if err := c.Get(ctx, "/api/v1/admin/terraform-mirrors/"+id, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) UpdateTerraformMirror(ctx context.Context, id string, req UpdateTerraformMirrorRequest) (*TerraformMirror, error) {
	var mirror TerraformMirror
	if err := c.Put(ctx, "/api/v1/admin/terraform-mirrors/"+id, req, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) DeleteTerraformMirror(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/terraform-mirrors/"+id)
}

func (c *Client) ListTerraformMirrors(ctx context.Context) ([]TerraformMirror, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/admin/terraform-mirrors", "configs")
	if err != nil {
		return nil, err
	}

	mirrors := make([]TerraformMirror, 0, len(items))
	for _, raw := range items {
		var m TerraformMirror
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("unmarshaling terraform mirror: %w", err)
		}
		mirrors = append(mirrors, m)
	}
	return mirrors, nil
}
