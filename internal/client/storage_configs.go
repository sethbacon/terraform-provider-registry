package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CreateStorageConfig(ctx context.Context, req CreateStorageConfigRequest) (*StorageConfig, error) {
	var sc StorageConfig
	if err := c.Post(ctx, "/api/v1/storage/configs", req, &sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

func (c *Client) GetStorageConfig(ctx context.Context, id string) (*StorageConfig, error) {
	var sc StorageConfig
	if err := c.Get(ctx, "/api/v1/storage/configs/"+id, &sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

func (c *Client) UpdateStorageConfig(ctx context.Context, id string, req UpdateStorageConfigRequest) (*StorageConfig, error) {
	var sc StorageConfig
	if err := c.Put(ctx, "/api/v1/storage/configs/"+id, req, &sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

func (c *Client) DeleteStorageConfig(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/storage/configs/"+id)
}

func (c *Client) ActivateStorageConfig(ctx context.Context, id string) error {
	return c.Post(ctx, "/api/v1/storage/configs/"+id+"/activate", nil, nil)
}

func (c *Client) ListStorageConfigs(ctx context.Context) ([]StorageConfig, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/v1/storage/configs", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseResponseError(resp)
	}

	var configs []StorageConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, fmt.Errorf("decoding storage configs: %w", err)
	}
	return configs, nil
}
