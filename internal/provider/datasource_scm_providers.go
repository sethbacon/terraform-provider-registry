package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &SCMProvidersDataSource{}

type SCMProvidersDataSource struct {
	client *client.Client
}

type SCMProvidersDataSourceModel struct {
	SCMProviders []SCMProviderDSItem `tfsdk:"scm_providers"`
}

type SCMProviderDSItem struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	BaseURL        types.String `tfsdk:"base_url"`
	TenantID       types.String `tfsdk:"tenant_id"`
	ClientID       types.String `tfsdk:"client_id"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewSCMProvidersDataSource() datasource.DataSource {
	return &SCMProvidersDataSource{}
}

func (d *SCMProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scm_providers"
}

func (d *SCMProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all SCM provider integrations.",
		Attributes: map[string]schema.Attribute{
			"scm_providers": schema.ListNestedAttribute{
				Description: "List of SCM providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "UUID."},
						"organization_id": schema.StringAttribute{Computed: true, Description: "Organization the integration is scoped to (null for global)."},
						"name":            schema.StringAttribute{Computed: true, Description: "Display name."},
						"type":            schema.StringAttribute{Computed: true, Description: "SCM type."},
						"base_url":        schema.StringAttribute{Computed: true, Description: "Base URL for self-hosted instances."},
						"tenant_id":       schema.StringAttribute{Computed: true, Description: "Azure DevOps tenant ID."},
						"client_id":       schema.StringAttribute{Computed: true, Description: "OAuth client ID."},
						"is_active":       schema.BoolAttribute{Computed: true, Description: "Whether the integration is enabled."},
						"created_at":      schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at":      schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *SCMProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SCMProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SCMProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scmProviders, err := d.client.ListSCMProviders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing SCM Providers", err.Error())
		return
	}

	items := make([]SCMProviderDSItem, len(scmProviders))
	for i, s := range scmProviders {
		item := SCMProviderDSItem{
			ID:        types.StringValue(s.ID),
			Name:      types.StringValue(s.Name),
			Type:      types.StringValue(s.ProviderType),
			IsActive:  types.BoolValue(s.IsActive),
			CreatedAt: types.StringValue(s.CreatedAt),
			UpdatedAt: types.StringValue(s.UpdatedAt),
		}
		if s.OrganizationID != nil {
			item.OrganizationID = types.StringValue(*s.OrganizationID)
		} else {
			item.OrganizationID = types.StringNull()
		}
		if s.BaseURL != nil {
			item.BaseURL = types.StringValue(*s.BaseURL)
		} else {
			item.BaseURL = types.StringNull()
		}
		if s.TenantID != nil {
			item.TenantID = types.StringValue(*s.TenantID)
		} else {
			item.TenantID = types.StringNull()
		}
		if s.ClientID != "" {
			item.ClientID = types.StringValue(s.ClientID)
		} else {
			item.ClientID = types.StringNull()
		}
		items[i] = item
	}

	config.SCMProviders = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
