package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateProviderRecord(ctx context.Context, req CreateProviderRecordRequest) (*ProviderRecord, error) {
	var resp struct {
		Provider ProviderRecord `json:"provider"`
	}
	if err := c.Post(ctx, "/api/v1/admin/providers/create", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Provider, nil
}

func (c *Client) GetProviderRecord(ctx context.Context, namespace, providerType string) (*ProviderRecord, error) {
	var resp struct {
		Provider ProviderRecord `json:"provider"`
	}
	path := fmt.Sprintf("/api/v1/providers/%s/%s", namespace, providerType)
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp.Provider, nil
}

func (c *Client) GetProviderRecordByID(ctx context.Context, id string) (*ProviderRecord, error) {
	var resp struct {
		Provider ProviderRecord `json:"provider"`
	}
	if err := c.Get(ctx, "/api/v1/admin/providers/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Provider, nil
}

func (c *Client) UpdateProviderRecord(ctx context.Context, namespace, providerType string, req UpdateProviderRecordRequest) (*ProviderRecord, error) {
	var resp struct {
		Provider ProviderRecord `json:"provider"`
	}
	path := fmt.Sprintf("/api/v1/providers/%s/%s", namespace, providerType)
	if err := c.Put(ctx, path, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Provider, nil
}

func (c *Client) DeleteProviderRecord(ctx context.Context, namespace, providerType string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/v1/providers/%s/%s", namespace, providerType))
}

func (c *Client) ListProviderRecords(ctx context.Context, namespace, search string) ([]ProviderRecord, error) {
	path := "/api/v1/providers/search"
	params := map[string]string{}
	if namespace != "" {
		params["namespace"] = namespace
	}
	if search != "" {
		params["q"] = search
	}
	if len(params) > 0 {
		path += BuildQuery(params)
	}

	items, err := FetchAllPages(ctx, c, path, "providers")
	if err != nil {
		return nil, err
	}

	providers := make([]ProviderRecord, 0, len(items))
	for _, raw := range items {
		var p ProviderRecord
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("unmarshaling provider: %w", err)
		}
		providers = append(providers, p)
	}
	return providers, nil
}
