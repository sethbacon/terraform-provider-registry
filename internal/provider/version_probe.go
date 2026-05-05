package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

// probeBackendVersion calls GET /version and adds a warning diagnostic if the
// backend is older than minSupportedBackendVersion or the endpoint is unreachable.
func probeBackendVersion(ctx context.Context, c *client.Client, resp *provider.ConfigureResponse) {
	bv, err := c.GetVersion(ctx)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Backend Version Check Failed",
			fmt.Sprintf(
				"Could not reach GET /version on the registry endpoint: %s. "+
					"Verify the endpoint is correct and that /version is accessible. "+
					"The provider will still attempt to operate normally.",
				err.Error(),
			),
		)
		return
	}

	min, err := parseSemver(minSupportedBackendVersion)
	if err != nil {
		return
	}

	got, err := parseSemver(bv.Version)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Backend Version Unrecognized",
			fmt.Sprintf(
				"The backend reported version %q which could not be parsed as a semantic version. "+
					"Minimum supported backend version is %s.",
				bv.Version, minSupportedBackendVersion,
			),
		)
		return
	}

	if got.less(min) {
		resp.Diagnostics.AddWarning(
			"Unsupported Backend Version",
			fmt.Sprintf(
				"The registry backend is running version %s, but this provider requires at least %s. "+
					"Upgrade the backend to avoid API errors. "+
					"Set version_check = false in the provider block to suppress this warning.",
				bv.Version, minSupportedBackendVersion,
			),
		)
	}
}
