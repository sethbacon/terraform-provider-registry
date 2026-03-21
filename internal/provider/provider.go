package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ provider.Provider = &RegistryProvider{}
var _ provider.ProviderWithFunctions = &RegistryProvider{}

// RegistryProvider implements the Terraform Registry provider.
type RegistryProvider struct {
	version string
}

// RegistryProviderModel holds provider-level configuration.
type RegistryProviderModel struct {
	Endpoint   types.String `tfsdk:"endpoint"`
	Token      types.String `tfsdk:"token"`
	Insecure   types.Bool   `tfsdk:"insecure"`
	Timeout    types.Int64  `tfsdk:"timeout"`
	MaxRetries types.Int64  `tfsdk:"max_retries"`
}

// New returns a provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RegistryProvider{version: version}
	}
}

func (p *RegistryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "registry"
	resp.Version = p.version
}

func (p *RegistryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages resources in a self-hosted Terraform Registry Backend instance.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The base URL of the Terraform Registry Backend (e.g., https://registry.example.com). " +
					"Can also be set with the TF_REGISTRY_ENDPOINT environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: "API key or JWT bearer token for authentication. " +
					"Can also be set with the TF_REGISTRY_TOKEN environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Disable TLS certificate verification. Should only be used in development environments.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP request timeout in seconds. Defaults to 30.",
				Optional:    true,
			},
			"max_retries": schema.Int64Attribute{
				Description: "Maximum number of retries for failed requests (429, 5xx). Defaults to 3.",
				Optional:    true,
			},
		},
	}
}

func (p *RegistryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RegistryProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("TF_REGISTRY_ENDPOINT")
	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() {
		endpoint = config.Endpoint.ValueString()
	}

	token := os.Getenv("TF_REGISTRY_TOKEN")
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Registry Endpoint",
			"The provider requires an endpoint URL. Set the endpoint attribute or the TF_REGISTRY_ENDPOINT environment variable.",
		)
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Registry Token",
			"The provider requires an authentication token. Set the token attribute or the TF_REGISTRY_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	opts := []client.Option{}

	if !config.Insecure.IsNull() && !config.Insecure.IsUnknown() && config.Insecure.ValueBool() {
		opts = append(opts, client.WithInsecure(true))
	}

	if !config.Timeout.IsNull() && !config.Timeout.IsUnknown() {
		opts = append(opts, client.WithTimeout(time.Duration(config.Timeout.ValueInt64())*time.Second))
	}

	if !config.MaxRetries.IsNull() && !config.MaxRetries.IsUnknown() {
		opts = append(opts, client.WithMaxRetries(int(config.MaxRetries.ValueInt64())))
	}

	c, err := client.NewClient(endpoint, token, opts...)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Provider Configuration", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *RegistryProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewOrganizationResource,
		NewOrganizationMemberResource,
		NewAPIKeyResource,
		NewRoleTemplateResource,
		NewModuleResource,
		NewProviderRecordResource,
		NewSCMProviderResource,
		NewModuleSCMLinkResource,
		NewMirrorResource,
		NewTerraformMirrorResource,
		NewStorageConfigResource,
		NewPolicyResource,
		NewApprovalRequestResource,
	}
}

func (p *RegistryProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUsersDataSource,
		NewOrganizationsDataSource,
		NewAPIKeysDataSource,
		NewModulesDataSource,
		NewProvidersDataSource,
		NewSCMProvidersDataSource,
		NewMirrorsDataSource,
		NewTerraformMirrorsDataSource,
		NewRoleTemplatesDataSource,
		NewAuditLogsDataSource,
		NewStatsDataSource,
	}
}

func (p *RegistryProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}
