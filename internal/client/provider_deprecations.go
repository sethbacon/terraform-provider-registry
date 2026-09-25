package client

import (
	"context"
	"fmt"
	"net/http"
)

// GetProviderVersion fetches a single provider version for inspecting deprecation state.
//
// The backend has no `GET /api/v1/providers/{ns}/{type}/versions/{version}`
// route: that path only carries DELETE (remove the version), and the
// deprecation state is written through `.../versions/{version}/deprecate`.
// Requesting it returned 404 on every refresh, which the
// registry_provider_version_deprecation Read took to mean "gone", so the
// resource dropped out of state and was re-created on every plan.
//
// Like GetModuleVersion, this therefore derives the version-level state from
// the provider-detail response (`GET /api/v1/providers/{ns}/{type}`), which
// returns the full, unpaginated versions list with `deprecated`,
// `deprecated_at` and `deprecation_message` on each entry.
//
// Errors:
//   - provider not found: the `*APIError` (404) from the underlying GET is propagated.
//   - provider found but version not in list: returns a synthetic `*APIError` with
//     status 404 so that `IsNotFound` recognises it.
func (c *Client) GetProviderVersion(ctx context.Context, namespace, providerType, version string) (*ProviderVersion, error) {
	var detail providerDetailResponse
	path := fmt.Sprintf("/api/v1/providers/%s/%s", namespace, providerType)
	if err := c.Get(ctx, path, &detail); err != nil {
		return nil, err
	}

	for i := range detail.Versions {
		if detail.Versions[i].Version == version {
			pv := detail.Versions[i]
			return &pv, nil
		}
	}

	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("provider version %s/%s@%s not found", namespace, providerType, version),
	}
}

// providerDetailResponse mirrors the backend `admin.ProviderDetailResponse`
// payload returned by `GET /api/v1/providers/{ns}/{type}`. Only the fields we
// actually consume are declared; unknown fields (platforms, protocols, signed,
// ...) are ignored by the decoder.
type providerDetailResponse struct {
	Versions []ProviderVersion `json:"versions"`
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
