package client

import "context"

// GetOIDCConfig fetches the current OIDC configuration via GET /api/v1/admin/oidc/config.
func (c *Client) GetOIDCConfig(ctx context.Context) (*OIDCConfig, error) {
	var cfg OIDCConfig
	if err := c.Get(ctx, "/api/v1/admin/oidc/config", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetOIDCGroupMappings fetches the current OIDC group mappings via GET /api/v1/admin/oidc/group-mapping.
func (c *Client) GetOIDCGroupMappings(ctx context.Context) ([]OIDCGroupMapping, error) {
	var mappings []OIDCGroupMapping
	if err := c.Get(ctx, "/api/v1/admin/oidc/group-mapping", &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// SetOIDCGroupMappings replaces the entire OIDC group mapping table via PUT /api/v1/admin/oidc/group-mapping.
func (c *Client) SetOIDCGroupMappings(ctx context.Context, mappings []OIDCGroupMapping) ([]OIDCGroupMapping, error) {
	var result []OIDCGroupMapping
	if err := c.Put(ctx, "/api/v1/admin/oidc/group-mapping", mappings, &result); err != nil {
		return nil, err
	}
	return result, nil
}
