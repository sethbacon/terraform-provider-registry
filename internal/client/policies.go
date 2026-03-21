package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*Policy, error) {
	var policy Policy
	if err := c.Post(ctx, "/api/v1/admin/policies", req, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	var policy Policy
	if err := c.Get(ctx, "/api/v1/admin/policies/"+id, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) UpdatePolicy(ctx context.Context, id string, req UpdatePolicyRequest) (*Policy, error) {
	var policy Policy
	if err := c.Put(ctx, "/api/v1/admin/policies/"+id, req, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/policies/"+id)
}

func (c *Client) ListPolicies(ctx context.Context) ([]Policy, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/v1/admin/policies", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseResponseError(resp)
	}

	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return nil, fmt.Errorf("decoding policies: %w", err)
	}
	return policies, nil
}
