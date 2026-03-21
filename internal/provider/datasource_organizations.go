package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &OrganizationsDataSource{}

type OrganizationsDataSource struct {
	client *client.Client
}

type OrganizationsDataSourceModel struct {
	Search        types.String        `tfsdk:"search"`
	Organizations []OrganizationModel `tfsdk:"organizations"`
}

type OrganizationModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewOrganizationsDataSource() datasource.DataSource {
	return &OrganizationsDataSource{}
}

func (d *OrganizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

func (d *OrganizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all registry organizations.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Description: "Optional search string to filter organizations.",
				Optional:    true,
			},
			"organizations": schema.ListNestedAttribute{
				Description: "List of organizations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "UUID of the organization."},
						"name":         schema.StringAttribute{Computed: true, Description: "URL-safe namespace name."},
						"display_name": schema.StringAttribute{Computed: true, Description: "Human-readable display name."},
						"created_at":   schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at":   schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *OrganizationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OrganizationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := d.client.ListOrganizations(ctx, config.Search.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Organizations", err.Error())
		return
	}

	models := make([]OrganizationModel, len(orgs))
	for i, o := range orgs {
		models[i] = OrganizationModel{
			ID:          types.StringValue(o.ID),
			Name:        types.StringValue(o.Name),
			DisplayName: types.StringValue(o.DisplayName),
			CreatedAt:   types.StringValue(o.CreatedAt),
			UpdatedAt:   types.StringValue(o.UpdatedAt),
		}
	}

	config.Organizations = models
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
