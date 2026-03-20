package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateStorageConfig(ctx context.Context, req CreateStorageConfigRequest) (*StorageConfig, error) {
	var resp struct {
		Config StorageConfig `json:"config"`
	}
	if err := c.Post(ctx, "/api/v1/storage/configs", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Config, nil
}

func (c *Client) GetStorageConfig(ctx context.Context, id string) (*StorageConfig, error) {
	var resp struct {
		Config StorageConfig `json:"config"`
	}
	if err := c.Get(ctx, "/api/v1/storage/configs/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Config, nil
}

func (c *Client) UpdateStorageConfig(ctx context.Context, id string, req UpdateStorageConfigRequest) (*StorageConfig, error) {
	var resp struct {
		Config StorageConfig `json:"config"`
	}
	if err := c.Put(ctx, "/api/v1/storage/configs/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Config, nil
}

func (c *Client) DeleteStorageConfig(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/storage/configs/"+id)
}

func (c *Client) ActivateStorageConfig(ctx context.Context, id string) error {
	return c.Post(ctx, "/api/v1/storage/configs/"+id+"/activate", nil, nil)
}

func (c *Client) ListStorageConfigs(ctx context.Context) ([]StorageConfig, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/storage/configs", "configs")
	if err != nil {
		return nil, err
	}

	configs := make([]StorageConfig, 0, len(items))
	for _, raw := range items {
		var sc StorageConfig
		if err := json.Unmarshal(raw, &sc); err != nil {
			return nil, fmt.Errorf("unmarshaling storage config: %w", err)
		}
		configs = append(configs, sc)
	}
	return configs, nil
}
