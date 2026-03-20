package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &ProvidersDataSource{}

type ProvidersDataSource struct {
	client *client.Client
}

type ProvidersDataSourceModel struct {
	Namespace types.String      `tfsdk:"namespace"`
	Search    types.String      `tfsdk:"search"`
	Providers []ProviderDSItem  `tfsdk:"providers"`
}

type ProviderDSItem struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Namespace      types.String `tfsdk:"namespace"`
	Type           types.String `tfsdk:"type"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewProvidersDataSource() datasource.DataSource {
	return &ProvidersDataSource{}
}

func (d *ProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_providers"
}

func (d *ProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists registry provider records.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Filter by namespace.",
				Optional:    true,
			},
			"search": schema.StringAttribute{
				Description: "Search string.",
				Optional:    true,
			},
			"providers": schema.ListNestedAttribute{
				Description: "List of provider records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "UUID."},
						"organization_id": schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"namespace":       schema.StringAttribute{Computed: true, Description: "Provider namespace."},
						"type":            schema.StringAttribute{Computed: true, Description: "Provider type."},
						"description":     schema.StringAttribute{Computed: true, Description: "Description."},
						"source":          schema.StringAttribute{Computed: true, Description: "Source repository URL."},
						"created_at":      schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at":      schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *ProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providers, err := d.client.ListProviderRecords(ctx, config.Namespace.ValueString(), config.Search.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Providers", err.Error())
		return
	}

	items := make([]ProviderDSItem, len(providers))
	for i, p := range providers {
		item := ProviderDSItem{
			ID:             types.StringValue(p.ID),
			OrganizationID: types.StringValue(p.OrganizationID),
			Namespace:      types.StringValue(p.Namespace),
			Type:           types.StringValue(p.Type),
			CreatedAt:      types.StringValue(p.CreatedAt),
			UpdatedAt:      types.StringValue(p.UpdatedAt),
		}
		if p.Description != nil {
			item.Description = types.StringValue(*p.Description)
		} else {
			item.Description = types.StringNull()
		}
		if p.Source != nil {
			item.Source = types.StringValue(*p.Source)
		} else {
			item.Source = types.StringNull()
		}
		items[i] = item
	}

	config.Providers = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
