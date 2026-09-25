package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
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

// testConfiguredResource wires c into r the way the provider does, so unit
// tests can call a resource's CRUD methods directly against an httptest
// backend. Those tests need neither TF_ACC nor a Terraform binary, which is
// what lets them pin the exact requests a method sends (or does not send).
func testConfiguredResource(t *testing.T, r fwresource.Resource, c *client.Client) fwresource.Resource {
	t.Helper()
	rc, ok := r.(fwresource.ResourceWithConfigure)
	if !ok {
		t.Fatalf("%T does not implement ResourceWithConfigure", r)
	}
	var resp fwresource.ConfigureResponse
	rc.Configure(context.Background(), fwresource.ConfigureRequest{ProviderData: c}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure: %v", resp.Diagnostics)
	}
	return r
}

// testResourceState builds a tfsdk.State for r's schema holding model.
func testResourceState(t *testing.T, r fwresource.Resource, model any) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var sresp fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &sresp)
	if sresp.Diagnostics.HasError() {
		t.Fatalf("Schema: %v", sresp.Diagnostics)
	}
	state := tfsdk.State{Schema: sresp.Schema}
	if diags := state.Set(ctx, model); diags.HasError() {
		t.Fatalf("State.Set: %v", diags)
	}
	return state
}
