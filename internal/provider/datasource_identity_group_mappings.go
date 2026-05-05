package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &IdentityGroupMappingsDataSource{}

type IdentityGroupMappingsDataSource struct {
	client *client.Client
}

type IdentityGroupMappingItemModel struct {
	Group          types.String `tfsdk:"group"`
	OrganizationID types.String `tfsdk:"organization_id"`
	RoleTemplateID types.String `tfsdk:"role_template_id"`
}

type IdentityGroupMappingsDataSourceModel struct {
	Mappings []IdentityGroupMappingItemModel `tfsdk:"mappings"`
}

func NewIdentityGroupMappingsDataSource() datasource.DataSource {
	return &IdentityGroupMappingsDataSource{}
}

func (d *IdentityGroupMappingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_group_mappings"
}

func (d *IdentityGroupMappingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the SAML/LDAP group → organization role mappings from the backend runtime config (read-only).",
		Attributes: map[string]schema.Attribute{
			"mappings": schema.ListNestedAttribute{
				Description: "List of identity group → role mappings.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Description: "SAML/LDAP group identifier.",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "UUID of the target organization.",
							Computed:    true,
						},
						"role_template_id": schema.StringAttribute{
							Description: "UUID of the role template assigned to members of this group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *IdentityGroupMappingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityGroupMappingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	mappings, err := d.client.GetIdentityGroupMappings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Identity Group Mappings", err.Error())
		return
	}

	items := make([]IdentityGroupMappingItemModel, 0, len(mappings))
	for _, m := range mappings {
		items = append(items, IdentityGroupMappingItemModel{
			Group:          types.StringValue(m.Group),
			OrganizationID: types.StringValue(m.OrganizationID),
			RoleTemplateID: types.StringValue(m.RoleTemplateID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, IdentityGroupMappingsDataSourceModel{Mappings: items})...)
}
