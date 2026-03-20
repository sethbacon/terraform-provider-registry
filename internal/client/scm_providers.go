package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateSCMProvider(ctx context.Context, req CreateSCMProviderRequest) (*SCMProvider, error) {
	var resp struct {
		SCMProvider SCMProvider `json:"scm_provider"`
	}
	if err := c.Post(ctx, "/api/v1/scm-providers", req, &resp); err != nil {
		return nil, err
	}
	return &resp.SCMProvider, nil
}

func (c *Client) GetSCMProvider(ctx context.Context, id string) (*SCMProvider, error) {
	var resp struct {
		SCMProvider SCMProvider `json:"scm_provider"`
	}
	if err := c.Get(ctx, "/api/v1/scm-providers/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.SCMProvider, nil
}

func (c *Client) UpdateSCMProvider(ctx context.Context, id string, req UpdateSCMProviderRequest) (*SCMProvider, error) {
	var resp struct {
		SCMProvider SCMProvider `json:"scm_provider"`
	}
	if err := c.Put(ctx, "/api/v1/scm-providers/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.SCMProvider, nil
}

func (c *Client) DeleteSCMProvider(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/scm-providers/"+id)
}

func (c *Client) ListSCMProviders(ctx context.Context) ([]SCMProvider, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/scm-providers", "scm_providers")
	if err != nil {
		return nil, err
	}

	scmProviders := make([]SCMProvider, 0, len(items))
	for _, raw := range items {
		var s SCMProvider
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("unmarshaling scm provider: %w", err)
		}
		scmProviders = append(scmProviders, s)
	}
	return scmProviders, nil
}
