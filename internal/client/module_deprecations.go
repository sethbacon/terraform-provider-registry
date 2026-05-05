package client

import (
	"context"
	"fmt"
)

// DeprecateModule sets module-level deprecation via POST /api/v1/modules/{ns}/{name}/{sys}/deprecate.
func (c *Client) DeprecateModule(ctx context.Context, namespace, name, system string, req DeprecateModuleRequest) (*Module, error) {
	var mod Module
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s/deprecate", namespace, name, system)
	if err := c.Post(ctx, path, req, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

// UndeprecateModule removes module-level deprecation via DELETE /api/v1/modules/{ns}/{name}/{sys}/deprecate.
func (c *Client) UndeprecateModule(ctx context.Context, namespace, name, system string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/v1/modules/%s/%s/%s/deprecate", namespace, name, system))
}

// GetModuleVersion fetches a single module version for inspecting deprecation state.
func (c *Client) GetModuleVersion(ctx context.Context, namespace, name, system, version string) (*ModuleVersion, error) {
	var mv ModuleVersion
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s/%s", namespace, name, system, version)
	if err := c.Get(ctx, path, &mv); err != nil {
		return nil, err
	}
	return &mv, nil
}

// DeprecateModuleVersion sets version-level deprecation via POST /api/v1/modules/{ns}/{name}/{sys}/versions/{ver}/deprecate.
func (c *Client) DeprecateModuleVersion(ctx context.Context, namespace, name, system, version string, req DeprecateModuleRequest) (*ModuleVersion, error) {
	var mv ModuleVersion
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s/versions/%s/deprecate", namespace, name, system, version)
	if err := c.Post(ctx, path, req, &mv); err != nil {
		return nil, err
	}
	return &mv, nil
}

// UndeprecateModuleVersion removes version-level deprecation via DELETE /api/v1/modules/{ns}/{name}/{sys}/versions/{ver}/deprecate.
func (c *Client) UndeprecateModuleVersion(ctx context.Context, namespace, name, system, version string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/v1/modules/%s/%s/%s/versions/%s/deprecate", namespace, name, system, version))
}
