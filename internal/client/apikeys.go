package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	var resp CreateAPIKeyResponse
	if err := c.Post(ctx, "/api/v1/apikeys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	var resp struct {
		Key APIKey `json:"key"`
	}
	if err := c.Get(ctx, "/api/v1/apikeys/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

func (c *Client) UpdateAPIKey(ctx context.Context, id string, req UpdateAPIKeyRequest) (*APIKey, error) {
	var resp struct {
		Key APIKey `json:"key"`
	}
	if err := c.Put(ctx, "/api/v1/apikeys/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/apikeys/"+id)
}

func (c *Client) ListAPIKeys(ctx context.Context, organizationID string) ([]APIKey, error) {
	path := "/api/v1/apikeys"
	if organizationID != "" {
		path += BuildQuery(map[string]string{"organization_id": organizationID})
	}

	items, err := FetchAllPages(ctx, c, path, "keys")
	if err != nil {
		return nil, err
	}

	keys := make([]APIKey, 0, len(items))
	for _, raw := range items {
		var k APIKey
		if err := json.Unmarshal(raw, &k); err != nil {
			return nil, fmt.Errorf("unmarshaling api key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// RotateAPIKey rotates an API key via POST /api/v1/apikeys/:id/rotate.
// Returns the new key (raw key value is only available in this response).
func (c *Client) RotateAPIKey(ctx context.Context, id string) (*CreateAPIKeyResponse, error) {
	var resp CreateAPIKeyResponse
	if err := c.Post(ctx, "/api/v1/apikeys/"+id+"/rotate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
