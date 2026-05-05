package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &TerraformMirrorVersionDataSource{}

type TerraformMirrorVersionDataSource struct {
	client *client.Client
}

type TerraformMirrorVersionPlatformModel struct {
	OS   types.String `tfsdk:"os"`
	Arch types.String `tfsdk:"arch"`
	URL  types.String `tfsdk:"url"`
}

type TerraformMirrorVersionDataSourceModel struct {
	MirrorID  types.String                          `tfsdk:"mirror_id"`
	Version   types.String                          `tfsdk:"version"`
	ID        types.String                          `tfsdk:"id"`
	Stable    types.Bool                            `tfsdk:"stable"`
	SyncedAt  types.String                          `tfsdk:"synced_at"`
	Platforms []TerraformMirrorVersionPlatformModel `tfsdk:"platforms"`
}

func NewTerraformMirrorVersionDataSource() datasource.DataSource {
	return &TerraformMirrorVersionDataSource{}
}

func (d *TerraformMirrorVersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_mirror_version"
}

func (d *TerraformMirrorVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads details for a single mirrored Terraform/OpenTofu version including available platforms.",
		Attributes: map[string]schema.Attribute{
			"mirror_id": schema.StringAttribute{
				Description: "UUID of the Terraform mirror configuration.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Semantic version string to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Internal UUID of the version record.",
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
			"platforms": schema.ListNestedAttribute{
				Description: "Available platform builds for this version.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"os": schema.StringAttribute{
							Description: "Operating system (e.g. 'linux', 'darwin').",
							Computed:    true,
						},
						"arch": schema.StringAttribute{
							Description: "Architecture (e.g. 'amd64', 'arm64').",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "Download URL for this platform binary.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TerraformMirrorVersionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerraformMirrorVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TerraformMirrorVersionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := d.client.GetTerraformMirrorVersion(ctx, state.MirrorID.ValueString(), state.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Terraform Mirror Version", err.Error())
		return
	}

	state.ID = types.StringValue(v.ID)
	state.Stable = types.BoolValue(v.Stable)
	if v.SyncedAt != nil {
		state.SyncedAt = types.StringValue(normalizeTimestamp(*v.SyncedAt))
	} else {
		state.SyncedAt = types.StringValue("")
	}

	platforms := make([]TerraformMirrorVersionPlatformModel, 0, len(v.Platforms))
	for _, p := range v.Platforms {
		platforms = append(platforms, TerraformMirrorVersionPlatformModel{
			OS:   types.StringValue(p.OS),
			Arch: types.StringValue(p.Arch),
			URL:  types.StringValue(p.URL),
		})
	}
	state.Platforms = platforms

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
