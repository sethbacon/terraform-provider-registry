package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
	"github.com/terraform-registry/terraform-provider-registry/internal/provider"
)

// devAdminEmail is the account DEV_MODE's dev-login endpoint always signs
// in as (backend internal/api/admin/dev.go). It is not configurable on the
// backend side, so bootstrapAcceptanceToken names the same account when it
// creates the platform administrator through the setup API.
const devAdminEmail = "admin@dev.local"

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

// TestMain skips all acceptance tests unless TF_ACC=1 is set. When
// TF_REGISTRY_TOKEN is not supplied, it bootstraps one against the backend
// named by TF_REGISTRY_ENDPOINT — see bootstrapAcceptanceToken.
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

	if os.Getenv("TF_REGISTRY_TOKEN") == "" {
		token, err := bootstrapAcceptanceToken(endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TF_REGISTRY_TOKEN not set and bootstrap failed: %v\n", err)
			os.Exit(1)
		}
		if err := os.Setenv("TF_REGISTRY_TOKEN", token); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set TF_REGISTRY_TOKEN: %v\n", err)
			os.Exit(1)
		}
	}

	resource.TestMain(m)
}

// bootstrapAcceptanceToken obtains a bearer token for a backend that was not
// handed one directly via TF_REGISTRY_TOKEN.
//
// It reads the backend's one-time setup-wizard token from the file named by
// TFR_ACC_SETUP_TOKEN_FILE (written by the backend itself — see
// deployments/docker-compose.test.yml's SETUP_TOKEN_FILE and README.md),
// uses it to configure the DEV_MODE dev-admin account as this deployment's
// platform administrator through the backend's own setup API (POST
// /api/v1/setup/admin — no direct database writes), and signs in as that
// account through the DEV_MODE dev-login endpoint. This only works against
// a fresh backend running with DEV_MODE=true that has not completed initial
// setup, which is what the included test stack starts.
//
// To run the acceptance suite against any other registry, set
// TF_REGISTRY_TOKEN to an existing admin API key or JWT instead; this
// function is then never called.
func bootstrapAcceptanceToken(endpoint string) (string, error) {
	tokenFile := os.Getenv("TFR_ACC_SETUP_TOKEN_FILE")
	if tokenFile == "" {
		return "", fmt.Errorf("neither TF_REGISTRY_TOKEN nor TFR_ACC_SETUP_TOKEN_FILE is set; " +
			"set TF_REGISTRY_TOKEN to an admin API key or JWT for an already-configured registry, " +
			"or TFR_ACC_SETUP_TOKEN_FILE to bootstrap a fresh DEV_MODE backend (see README.md)")
	}

	setupToken, err := readSetupToken(tokenFile)
	if err != nil {
		return "", err
	}

	if err := configureAcceptanceAdmin(endpoint, setupToken); err != nil {
		return "", err
	}

	return devLogin(endpoint)
}

// readSetupToken polls for the backend's one-time setup token at path. The
// backend writes this file after it can serve requests, so a short poll
// covers the race between the container starting and the file appearing —
// the same race the caller's own health check already waits out for the
// HTTP port.
func readSetupToken(path string) (string, error) {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(b))) > 0 {
			return strings.TrimSpace(string(b)), nil
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return "", fmt.Errorf("reading setup token from %s (waited 30s): %w", path, lastErr)
}

// configureAcceptanceAdmin bootstraps devAdminEmail as this deployment's
// platform administrator by calling the backend's setup wizard API
// (POST /api/v1/setup/admin) directly, authenticated with the one-time
// setup token. This is the same call the setup wizard's UI makes; it never
// touches the database itself.
func configureAcceptanceAdmin(endpoint, setupToken string) error {
	reqBody, err := json.Marshal(map[string]string{"email": devAdminEmail})
	if err != nil {
		return fmt.Errorf("marshal setup/admin request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint+"/api/v1/setup/admin", bytes.NewReader(reqBody)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("build setup/admin request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "SetupToken "+setupToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s/api/v1/setup/admin: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setup/admin returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// devLogin signs in as devAdminEmail through DEV_MODE's dev-login endpoint
// and returns the session JWT. DevLoginHandler delivers the JWT only via an
// httpOnly cookie (the JSON body carries no token field), so it is read off
// the response's Set-Cookie header; the same JWT value authenticates as a
// Bearer token, which is how the provider's client sends it.
func devLogin(endpoint string) (string, error) {
	url := endpoint + "/api/v1/dev/login"
	resp, err := http.Post(url, "application/json", nil) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("dev-login returned %d: %s", resp.StatusCode, body)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "tfr_auth_token" && cookie.Value != "" {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("dev-login response carried no tfr_auth_token session cookie")
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
