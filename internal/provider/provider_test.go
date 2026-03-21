package provider_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/terraform-registry/terraform-provider-registry/internal/provider"
)

// testAccProtoV6ProviderFactories wires the provider under test into the acceptance test framework.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"registry": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccProviderConfig returns a provider configuration block using env vars set by TestMain.
func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "registry" {
  endpoint = %q
  token    = %q
  insecure = true
}
`, os.Getenv("TF_REGISTRY_ENDPOINT"), os.Getenv("TF_REGISTRY_TOKEN"))
}

// TestMain skips all acceptance tests unless TF_ACC=1 is set, and optionally
// obtains a dev-mode JWT token when only TF_REGISTRY_ENDPOINT is provided.
func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC") == "" {
		// Unit-safe: skip all acceptance tests when not in acc mode.
		os.Exit(m.Run())
	}

	endpoint := os.Getenv("TF_REGISTRY_ENDPOINT")
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "TF_REGISTRY_ENDPOINT must be set for acceptance tests")
		os.Exit(1)
	}

	// If a token was not supplied, attempt to fetch one via the dev-mode login endpoint.
	if os.Getenv("TF_REGISTRY_TOKEN") == "" {
		token, err := fetchDevToken(endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TF_REGISTRY_TOKEN not set and dev-login failed: %v\n", err)
			os.Exit(1)
		}
		if err := os.Setenv("TF_REGISTRY_TOKEN", token); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set TF_REGISTRY_TOKEN: %v\n", err)
			os.Exit(1)
		}
	}

	resource.TestMain(m)
}

// fetchDevToken calls POST /api/v1/dev/login and returns the token from the response.
// This requires DEV_MODE=true on the backend.
func fetchDevToken(endpoint string) (string, error) {
	url := endpoint + "/api/v1/dev/login"
	resp, err := http.Post(url, "application/json", nil) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dev-login returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse dev-login response: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("dev-login returned empty token")
	}
	return result.Token, nil
}
