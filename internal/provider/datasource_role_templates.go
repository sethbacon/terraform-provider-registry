package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &RoleTemplatesDataSource{}

type RoleTemplatesDataSource struct {
	client *client.Client
}

type RoleTemplatesDataSourceModel struct {
	RoleTemplates []RoleTemplateDSItem `tfsdk:"role_templates"`
}

type RoleTemplateDSItem struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Scopes      types.List   `tfsdk:"scopes"`
	IsSystem    types.Bool   `tfsdk:"is_system"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewRoleTemplatesDataSource() datasource.DataSource {
	return &RoleTemplatesDataSource{}
}

func (d *RoleTemplatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_templates"
}

func (d *RoleTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all RBAC role templates.",
		Attributes: map[string]schema.Attribute{
			"role_templates": schema.ListNestedAttribute{
				Description: "List of role templates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "UUID."},
						"name":         schema.StringAttribute{Computed: true, Description: "Role name."},
						"display_name": schema.StringAttribute{Computed: true, Description: "Display name."},
						"description":  schema.StringAttribute{Computed: true, Description: "Description."},
						"scopes": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Permission scopes.",
						},
						"is_system":  schema.BoolAttribute{Computed: true, Description: "Whether built-in."},
						"created_at": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *RoleTemplatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoleTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RoleTemplatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	templates, err := d.client.ListRoleTemplates(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Role Templates", err.Error())
		return
	}

	items := make([]RoleTemplateDSItem, len(templates))
	for i, t := range templates {
		scopeValues := make([]types.String, len(t.Scopes))
		for j, s := range t.Scopes {
			scopeValues[j] = types.StringValue(s)
		}
		scopeList, _ := types.ListValueFrom(ctx, types.StringType, scopeValues)

		item := RoleTemplateDSItem{
			ID:          types.StringValue(t.ID),
			Name:        types.StringValue(t.Name),
			DisplayName: types.StringValue(t.DisplayName),
			Scopes:      scopeList,
			IsSystem:    types.BoolValue(t.IsSystem),
			CreatedAt:   types.StringValue(t.CreatedAt),
			UpdatedAt:   types.StringValue(t.UpdatedAt),
		}
		if t.Description != nil {
			item.Description = types.StringValue(*t.Description)
		} else {
			item.Description = types.StringNull()
		}
		items[i] = item
	}

	config.RoleTemplates = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
