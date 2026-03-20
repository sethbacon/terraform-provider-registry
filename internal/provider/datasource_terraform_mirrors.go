package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &TerraformMirrorsDataSource{}

type TerraformMirrorsDataSource struct {
	client *client.Client
}

type TerraformMirrorsDataSourceModel struct {
	TerraformMirrors []TerraformMirrorDSItem `tfsdk:"terraform_mirrors"`
}

type TerraformMirrorDSItem struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Tool              types.String `tfsdk:"tool"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	UpstreamURL       types.String `tfsdk:"upstream_url"`
	GPGVerify         types.Bool   `tfsdk:"gpg_verify"`
	StableOnly        types.Bool   `tfsdk:"stable_only"`
	SyncIntervalHours types.Int64  `tfsdk:"sync_interval_hours"`
	LastSyncAt        types.String `tfsdk:"last_sync_at"`
	LastSyncStatus    types.String `tfsdk:"last_sync_status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewTerraformMirrorsDataSource() datasource.DataSource {
	return &TerraformMirrorsDataSource{}
}

func (d *TerraformMirrorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_mirrors"
}

func (d *TerraformMirrorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Terraform/OpenTofu binary mirror configurations.",
		Attributes: map[string]schema.Attribute{
			"terraform_mirrors": schema.ListNestedAttribute{
				Description: "List of Terraform binary mirror configurations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.StringAttribute{Computed: true, Description: "UUID."},
						"name":                schema.StringAttribute{Computed: true, Description: "Mirror name."},
						"description":         schema.StringAttribute{Computed: true, Description: "Description."},
						"tool":                schema.StringAttribute{Computed: true, Description: "Tool type."},
						"enabled":             schema.BoolAttribute{Computed: true, Description: "Whether syncing is enabled."},
						"upstream_url":        schema.StringAttribute{Computed: true, Description: "Upstream URL."},
						"gpg_verify":          schema.BoolAttribute{Computed: true, Description: "GPG verification enabled."},
						"stable_only":         schema.BoolAttribute{Computed: true, Description: "Only stable releases."},
						"sync_interval_hours": schema.Int64Attribute{Computed: true, Description: "Sync interval in hours."},
						"last_sync_at":        schema.StringAttribute{Computed: true, Description: "Last sync timestamp."},
						"last_sync_status":    schema.StringAttribute{Computed: true, Description: "Last sync status."},
						"created_at":          schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at":          schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *TerraformMirrorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerraformMirrorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TerraformMirrorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mirrors, err := d.client.ListTerraformMirrors(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Terraform Mirrors", err.Error())
		return
	}

	items := make([]TerraformMirrorDSItem, len(mirrors))
	for i, m := range mirrors {
		item := TerraformMirrorDSItem{
			ID:                types.StringValue(m.ID),
			Name:              types.StringValue(m.Name),
			Tool:              types.StringValue(m.Tool),
			Enabled:           types.BoolValue(m.Enabled),
			UpstreamURL:       types.StringValue(m.UpstreamURL),
			GPGVerify:         types.BoolValue(m.GPGVerify),
			StableOnly:        types.BoolValue(m.StableOnly),
			SyncIntervalHours: types.Int64Value(int64(m.SyncIntervalHours)),
			CreatedAt:         types.StringValue(m.CreatedAt),
			UpdatedAt:         types.StringValue(m.UpdatedAt),
		}
		if m.Description != nil {
			item.Description = types.StringValue(*m.Description)
		} else {
			item.Description = types.StringNull()
		}
		if m.LastSyncAt != nil {
			item.LastSyncAt = types.StringValue(*m.LastSyncAt)
		} else {
			item.LastSyncAt = types.StringNull()
		}
		if m.LastSyncStatus != nil {
			item.LastSyncStatus = types.StringValue(*m.LastSyncStatus)
		} else {
			item.LastSyncStatus = types.StringNull()
		}
		items[i] = item
	}

	config.TerraformMirrors = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
