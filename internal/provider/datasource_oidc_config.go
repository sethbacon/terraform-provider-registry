package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &OIDCConfigDataSource{}

type OIDCConfigDataSource struct {
	client *client.Client
}

type OIDCConfigDataSourceModel struct {
	Issuer        types.String `tfsdk:"issuer"`
	ClientID      types.String `tfsdk:"client_id"`
	Scopes        types.List   `tfsdk:"scopes"`
	GroupsClaim   types.String `tfsdk:"groups_claim"`
	UsernameClaim types.String `tfsdk:"username_claim"`
}

func NewOIDCConfigDataSource() datasource.DataSource {
	return &OIDCConfigDataSource{}
}

func (d *OIDCConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_config"
}

func (d *OIDCConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the backend OIDC configuration (read-only). The client secret is never exposed.",
		Attributes: map[string]schema.Attribute{
			"issuer": schema.StringAttribute{
				Description: "OIDC issuer URL.",
				Computed:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "OIDC client ID.",
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Requested OIDC scopes.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"groups_claim": schema.StringAttribute{
				Description: "JWT claim used to extract group membership.",
				Computed:    true,
			},
			"username_claim": schema.StringAttribute{
				Description: "JWT claim used as the username.",
				Computed:    true,
			},
		},
	}
}

func (d *OIDCConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OIDCConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	cfg, err := d.client.GetOIDCConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OIDC Config", err.Error())
		return
	}

	scopes, diags := types.ListValueFrom(ctx, types.StringType, cfg.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := OIDCConfigDataSourceModel{
		Issuer:        types.StringValue(cfg.Issuer),
		ClientID:      types.StringValue(cfg.ClientID),
		Scopes:        scopes,
		GroupsClaim:   types.StringValue(cfg.GroupsClaim),
		UsernameClaim: types.StringValue(cfg.UsernameClaim),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
