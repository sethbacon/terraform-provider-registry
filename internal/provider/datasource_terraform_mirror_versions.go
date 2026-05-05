package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &TerraformMirrorVersionsDataSource{}

type TerraformMirrorVersionsDataSource struct {
	client *client.Client
}

type TerraformMirrorVersionsDataSourceModel struct {
	MirrorID types.String                      `tfsdk:"mirror_id"`
	Versions []TerraformMirrorVersionItemModel `tfsdk:"versions"`
}

type TerraformMirrorVersionItemModel struct {
	ID       types.String `tfsdk:"id"`
	Version  types.String `tfsdk:"version"`
	Stable   types.Bool   `tfsdk:"stable"`
	SyncedAt types.String `tfsdk:"synced_at"`
}

func NewTerraformMirrorVersionsDataSource() datasource.DataSource {
	return &TerraformMirrorVersionsDataSource{}
}

func (d *TerraformMirrorVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_mirror_versions"
}

func (d *TerraformMirrorVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Terraform/OpenTofu versions mirrored by a given mirror configuration.",
		Attributes: map[string]schema.Attribute{
			"mirror_id": schema.StringAttribute{
				Description: "UUID of the Terraform mirror configuration.",
				Required:    true,
			},
			"versions": schema.ListNestedAttribute{
				Description: "List of mirrored versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal UUID of the version record.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Semantic version string.",
							Computed:    true,
						},
						"stable": schema.BoolAttribute{
							Description: "Whether this is a stable release.",
							Computed:    true,
						},
						"synced_at": schema.StringAttribute{
							Description: "ISO 8601 timestamp when this version was synced.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TerraformMirrorVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerraformMirrorVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TerraformMirrorVersionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	versions, err := d.client.ListTerraformMirrorVersions(ctx, state.MirrorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Terraform Mirror Versions", err.Error())
		return
	}

	items := make([]TerraformMirrorVersionItemModel, 0, len(versions))
	for _, v := range versions {
		item := TerraformMirrorVersionItemModel{
			ID:      types.StringValue(v.ID),
			Version: types.StringValue(v.Version),
			Stable:  types.BoolValue(v.Stable),
		}
		if v.SyncedAt != nil {
			item.SyncedAt = types.StringValue(normalizeTimestamp(*v.SyncedAt))
		} else {
			item.SyncedAt = types.StringValue("")
		}
		items = append(items, item)
	}

	state.Versions = items
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
