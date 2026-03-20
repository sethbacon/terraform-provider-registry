package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &ModulesDataSource{}

type ModulesDataSource struct {
	client *client.Client
}

type ModulesDataSourceModel struct {
	Namespace types.String  `tfsdk:"namespace"`
	Search    types.String  `tfsdk:"search"`
	Modules   []ModuleDSItem `tfsdk:"modules"`
}

type ModuleDSItem struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	System         types.String `tfsdk:"system"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewModulesDataSource() datasource.DataSource {
	return &ModulesDataSource{}
}

func (d *ModulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modules"
}

func (d *ModulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists registry modules with optional namespace and search filtering.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Filter by namespace.",
				Optional:    true,
			},
			"search": schema.StringAttribute{
				Description: "Search string.",
				Optional:    true,
			},
			"modules": schema.ListNestedAttribute{
				Description: "List of modules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "UUID."},
						"organization_id": schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"namespace":       schema.StringAttribute{Computed: true, Description: "Module namespace."},
						"name":            schema.StringAttribute{Computed: true, Description: "Module name."},
						"system":          schema.StringAttribute{Computed: true, Description: "Provider system."},
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

func (d *ModulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ModulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modules, err := d.client.ListModules(ctx, config.Namespace.ValueString(), config.Search.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Modules", err.Error())
		return
	}

	items := make([]ModuleDSItem, len(modules))
	for i, m := range modules {
		item := ModuleDSItem{
			ID:             types.StringValue(m.ID),
			OrganizationID: types.StringValue(m.OrganizationID),
			Namespace:      types.StringValue(m.Namespace),
			Name:           types.StringValue(m.Name),
			System:         types.StringValue(m.System),
			CreatedAt:      types.StringValue(m.CreatedAt),
			UpdatedAt:      types.StringValue(m.UpdatedAt),
		}
		if m.Description != nil {
			item.Description = types.StringValue(*m.Description)
		} else {
			item.Description = types.StringNull()
		}
		if m.Source != nil {
			item.Source = types.StringValue(*m.Source)
		} else {
			item.Source = types.StringNull()
		}
		items[i] = item
	}

	config.Modules = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
