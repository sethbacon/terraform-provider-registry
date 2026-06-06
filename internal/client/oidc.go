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

// GetOIDCGroupMappings returns the group mappings from the active OIDC config.
// The backend has no separate GET for group mappings; they are embedded in the config response.
func (c *Client) GetOIDCGroupMappings(ctx context.Context) ([]OIDCGroupMapping, error) {
	cfg, err := c.GetOIDCConfig(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.GroupMappings, nil
}

// SetOIDCGroupMappings replaces the OIDC group mapping configuration via PUT /api/v1/admin/oidc/group-mapping.
func (c *Client) SetOIDCGroupMappings(ctx context.Context, input OIDCGroupMappingInput) (*OIDCConfig, error) {
	var result OIDCConfig
	if err := c.Put(ctx, "/api/v1/admin/oidc/group-mapping", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
