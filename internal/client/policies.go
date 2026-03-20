package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*Policy, error) {
	var resp struct {
		Policy Policy `json:"policy"`
	}
	if err := c.Post(ctx, "/api/v1/admin/policies", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Policy, nil
}

func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	var resp struct {
		Policy Policy `json:"policy"`
	}
	if err := c.Get(ctx, "/api/v1/admin/policies/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Policy, nil
}

func (c *Client) UpdatePolicy(ctx context.Context, id string, req UpdatePolicyRequest) (*Policy, error) {
	var resp struct {
		Policy Policy `json:"policy"`
	}
	if err := c.Put(ctx, "/api/v1/admin/policies/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Policy, nil
}

func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/policies/"+id)
}

func (c *Client) ListPolicies(ctx context.Context) ([]Policy, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/admin/policies", "policies")
	if err != nil {
		return nil, err
	}

	policies := make([]Policy, 0, len(items))
	for _, raw := range items {
		var p Policy
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("unmarshaling policy: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}
