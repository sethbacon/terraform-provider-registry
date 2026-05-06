package client

import (
	"context"
	"fmt"
	"net/http"
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
//
// The backend has never exposed a dedicated `/api/v1/modules/{ns}/{name}/{sys}/{version}`
// endpoint — every version-scoped route lives under `.../versions/{version}/...`. This
// function therefore derives the version-level state from the module-detail response
// (`GET /api/v1/modules/{ns}/{name}/{sys}`), which already returns the full versions
// list. Doing it this way keeps the provider working against every backend release
// shipped to date, regardless of whether a future dedicated endpoint is added.
//
// Errors:
//   - module not found: the `*APIError` (404) from the underlying GET is propagated.
//   - module found but version not in list: returns a synthetic `*APIError` with
//     status 404 so that `IsNotFound` recognises it.
func (c *Client) GetModuleVersion(ctx context.Context, namespace, name, system, version string) (*ModuleVersion, error) {
	var detail moduleDetailResponse
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s", namespace, name, system)
	if err := c.Get(ctx, path, &detail); err != nil {
		return nil, err
	}

	for i := range detail.Versions {
		if detail.Versions[i].Version == version {
			mv := detail.Versions[i]
			return &mv, nil
		}
	}

	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("module version %s/%s/%s@%s not found", namespace, name, system, version),
	}
}

// moduleDetailResponse mirrors the backend `admin.ModuleDetailResponse` payload
// returned by `GET /api/v1/modules/{ns}/{name}/{sys}`. Only the fields we
// actually consume are declared; unknown fields are ignored by the decoder.
type moduleDetailResponse struct {
	Versions []ModuleVersion `json:"versions"`
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
