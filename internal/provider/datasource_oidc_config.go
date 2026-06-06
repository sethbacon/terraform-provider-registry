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
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	IssuerURL      types.String `tfsdk:"issuer_url"`
	ClientID       types.String `tfsdk:"client_id"`
	RedirectURL    types.String `tfsdk:"redirect_url"`
	Scopes         types.List   `tfsdk:"scopes"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	GroupClaimName types.String `tfsdk:"group_claim_name"`
	DefaultRole    types.String `tfsdk:"default_role"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
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
			"id": schema.StringAttribute{
				Description: "UUID of the OIDC configuration.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Display name of the OIDC configuration.",
				Computed:    true,
			},
			"provider_type": schema.StringAttribute{
				Description: "OIDC provider type (e.g. generic, okta, azure-ad).",
				Computed:    true,
			},
			"issuer_url": schema.StringAttribute{
				Description: "OIDC issuer URL.",
				Computed:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "OIDC client ID.",
				Computed:    true,
			},
			"redirect_url": schema.StringAttribute{
				Description: "OAuth2 redirect URL.",
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Requested OIDC scopes.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this OIDC configuration is active.",
				Computed:    true,
			},
			"group_claim_name": schema.StringAttribute{
				Description: "JWT claim name used to extract group membership.",
				Computed:    true,
			},
			"default_role": schema.StringAttribute{
				Description: "Default role assigned when no group mapping matches.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp (RFC3339).",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp (RFC3339).",
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
		ID:             types.StringValue(cfg.ID),
		Name:           types.StringValue(cfg.Name),
		ProviderType:   types.StringValue(cfg.ProviderType),
		IssuerURL:      types.StringValue(cfg.IssuerURL),
		ClientID:       types.StringValue(cfg.ClientID),
		RedirectURL:    types.StringValue(cfg.RedirectURL),
		Scopes:         scopes,
		IsActive:       types.BoolValue(cfg.IsActive),
		GroupClaimName: types.StringValue(cfg.GroupClaimName),
		DefaultRole:    types.StringValue(cfg.DefaultRole),
		CreatedAt:      types.StringValue(cfg.CreatedAt),
		UpdatedAt:      types.StringValue(cfg.UpdatedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
