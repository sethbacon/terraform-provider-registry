package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateMirror(ctx context.Context, req CreateMirrorRequest) (*Mirror, error) {
	var mirror Mirror
	if err := c.Post(ctx, "/api/v1/admin/mirrors", req, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) GetMirror(ctx context.Context, id string) (*Mirror, error) {
	var mirror Mirror
	if err := c.Get(ctx, "/api/v1/admin/mirrors/"+id, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) UpdateMirror(ctx context.Context, id string, req UpdateMirrorRequest) (*Mirror, error) {
	var mirror Mirror
	if err := c.Put(ctx, "/api/v1/admin/mirrors/"+id, req, &mirror); err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (c *Client) DeleteMirror(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/mirrors/"+id)
}

func (c *Client) ListMirrors(ctx context.Context) ([]Mirror, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/admin/mirrors", "mirrors")
	if err != nil {
		return nil, err
	}

	mirrors := make([]Mirror, 0, len(items))
	for _, raw := range items {
		var m Mirror
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("unmarshaling mirror: %w", err)
		}
		mirrors = append(mirrors, m)
	}
	return mirrors, nil
}
