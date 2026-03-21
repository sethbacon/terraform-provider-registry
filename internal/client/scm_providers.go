package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CreateSCMProvider(ctx context.Context, req CreateSCMProviderRequest) (*SCMProvider, error) {
	var scm SCMProvider
	if err := c.Post(ctx, "/api/v1/scm-providers", req, &scm); err != nil {
		return nil, err
	}
	return &scm, nil
}

func (c *Client) GetSCMProvider(ctx context.Context, id string) (*SCMProvider, error) {
	var scm SCMProvider
	if err := c.Get(ctx, "/api/v1/scm-providers/"+id, &scm); err != nil {
		return nil, err
	}
	return &scm, nil
}

func (c *Client) UpdateSCMProvider(ctx context.Context, id string, req UpdateSCMProviderRequest) (*SCMProvider, error) {
	var scm SCMProvider
	if err := c.Put(ctx, "/api/v1/scm-providers/"+id, req, &scm); err != nil {
		return nil, err
	}
	return &scm, nil
}

func (c *Client) DeleteSCMProvider(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/scm-providers/"+id)
}

func (c *Client) ListSCMProviders(ctx context.Context) ([]SCMProvider, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/v1/scm-providers", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseResponseError(resp)
	}

	var providers []SCMProvider
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return nil, fmt.Errorf("decoding scm providers: %w", err)
	}
	return providers, nil
}
