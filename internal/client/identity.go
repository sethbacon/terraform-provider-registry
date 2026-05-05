package client

import "context"

// GetIdentityGroupMappings fetches SAML/LDAP group → role mappings via GET /api/v1/admin/identity/group-mappings.
func (c *Client) GetIdentityGroupMappings(ctx context.Context) ([]IdentityGroupMapping, error) {
	var mappings []IdentityGroupMapping
	if err := c.Get(ctx, "/api/v1/admin/identity/group-mappings", &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// GetMTLSConfig fetches the mTLS configuration via GET /api/v1/admin/mtls/config.
func (c *Client) GetMTLSConfig(ctx context.Context) (*MTLSConfig, error) {
	var cfg MTLSConfig
	if err := c.Get(ctx, "/api/v1/admin/mtls/config", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
