package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &PolicyEngineConfigDataSource{}

type PolicyEngineConfigDataSource struct {
	client *client.Client
}

type PolicyEngineConfigDataSourceModel struct {
	BundleURL    types.String `tfsdk:"bundle_url"`
	BundleETag   types.String `tfsdk:"bundle_etag"`
	LastLoadedAt types.String `tfsdk:"last_loaded_at"`
	Status       types.String `tfsdk:"status"`
}

func NewPolicyEngineConfigDataSource() datasource.DataSource {
	return &PolicyEngineConfigDataSource{}
}

func (d *PolicyEngineConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_engine_config"
}

func (d *PolicyEngineConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the rego policy engine bundle configuration.",
		Attributes: map[string]schema.Attribute{
			"bundle_url": schema.StringAttribute{
				Description: "URL from which the rego bundle is loaded.",
				Computed:    true,
			},
			"bundle_etag": schema.StringAttribute{
				Description: "ETag of the currently loaded bundle.",
				Computed:    true,
			},
			"last_loaded_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the bundle was last loaded.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Bundle load status ('ok', 'error', etc.).",
				Computed:    true,
			},
		},
	}
}

func (d *PolicyEngineConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *PolicyEngineConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	cfg, err := d.client.GetPolicyEngineConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Policy Engine Config", err.Error())
		return
	}

	model := PolicyEngineConfigDataSourceModel{
		BundleURL: types.StringValue(cfg.BundleURL),
		Status:    types.StringValue(cfg.Status),
	}
	if cfg.BundleETag != nil {
		model.BundleETag = types.StringValue(*cfg.BundleETag)
	} else {
		model.BundleETag = types.StringValue("")
	}
	if cfg.LastLoadedAt != nil {
		model.LastLoadedAt = types.StringValue(normalizeTimestamp(*cfg.LastLoadedAt))
	} else {
		model.LastLoadedAt = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
