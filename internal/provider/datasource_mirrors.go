package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &MirrorsDataSource{}

type MirrorsDataSource struct {
	client *client.Client
}

type MirrorsDataSourceModel struct {
	Mirrors []MirrorDSItem `tfsdk:"mirrors"`
}

type MirrorDSItem struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	UpstreamRegistryURL types.String `tfsdk:"upstream_registry_url"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	SyncIntervalHours   types.Int64  `tfsdk:"sync_interval_hours"`
	LastSyncAt          types.String `tfsdk:"last_sync_at"`
	LastSyncStatus      types.String `tfsdk:"last_sync_status"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func NewMirrorsDataSource() datasource.DataSource {
	return &MirrorsDataSource{}
}

func (d *MirrorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mirrors"
}

func (d *MirrorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all provider mirror configurations.",
		Attributes: map[string]schema.Attribute{
			"mirrors": schema.ListNestedAttribute{
				Description: "List of mirror configurations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true, Description: "UUID."},
						"name":                  schema.StringAttribute{Computed: true, Description: "Mirror name."},
						"description":           schema.StringAttribute{Computed: true, Description: "Description."},
						"upstream_registry_url": schema.StringAttribute{Computed: true, Description: "Upstream registry URL."},
						"organization_id":       schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"enabled":               schema.BoolAttribute{Computed: true, Description: "Whether syncing is enabled."},
						"sync_interval_hours":   schema.Int64Attribute{Computed: true, Description: "Sync interval in hours."},
						"last_sync_at":          schema.StringAttribute{Computed: true, Description: "Last sync timestamp."},
						"last_sync_status":      schema.StringAttribute{Computed: true, Description: "Last sync status."},
						"created_at":            schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at":            schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *MirrorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MirrorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MirrorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mirrors, err := d.client.ListMirrors(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Mirrors", err.Error())
		return
	}

	items := make([]MirrorDSItem, len(mirrors))
	for i, m := range mirrors {
		item := MirrorDSItem{
			ID:                  types.StringValue(m.ID),
			Name:                types.StringValue(m.Name),
			UpstreamRegistryURL: types.StringValue(m.UpstreamRegistryURL),
			Enabled:             types.BoolValue(m.Enabled),
			SyncIntervalHours:   types.Int64Value(int64(m.SyncIntervalHours)),
			CreatedAt:           types.StringValue(m.CreatedAt),
			UpdatedAt:           types.StringValue(m.UpdatedAt),
		}
		if m.Description != nil {
			item.Description = types.StringValue(*m.Description)
		} else {
			item.Description = types.StringNull()
		}
		if m.OrganizationID != nil {
			item.OrganizationID = types.StringValue(*m.OrganizationID)
		} else {
			item.OrganizationID = types.StringNull()
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

	config.Mirrors = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
