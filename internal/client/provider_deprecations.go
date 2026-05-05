package client

import (
	"context"
	"fmt"
)

// GetProviderVersion fetches a single provider version for inspecting deprecation state.
func (c *Client) GetProviderVersion(ctx context.Context, namespace, providerType, version string) (*ProviderVersion, error) {
	var pv ProviderVersion
	path := fmt.Sprintf("/api/v1/providers/%s/%s/versions/%s", namespace, providerType, version)
	if err := c.Get(ctx, path, &pv); err != nil {
		return nil, err
	}
	return &pv, nil
}

// DeprecateProviderVersion sets version-level deprecation via POST /api/v1/providers/{ns}/{type}/versions/{ver}/deprecate.
func (c *Client) DeprecateProviderVersion(ctx context.Context, namespace, providerType, version string, req DeprecateProviderVersionRequest) (*ProviderVersion, error) {
	var pv ProviderVersion
	path := fmt.Sprintf("/api/v1/providers/%s/%s/versions/%s/deprecate", namespace, providerType, version)
	if err := c.Post(ctx, path, req, &pv); err != nil {
		return nil, err
	}
	return &pv, nil
}

// UndeprecateProviderVersion removes version-level deprecation via DELETE /api/v1/providers/{ns}/{type}/versions/{ver}/deprecate.
func (c *Client) UndeprecateProviderVersion(ctx context.Context, namespace, providerType, version string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/v1/providers/%s/%s/versions/%s/deprecate", namespace, providerType, version))
}
