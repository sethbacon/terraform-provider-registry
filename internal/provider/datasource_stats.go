package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &StatsDataSource{}

type StatsDataSource struct {
	client *client.Client
}

// StatsDataSourceModel mirrors backend admin.DashboardStats. The shape was
// reworked alongside the v1 backend release: the previous flat counters
// (total_modules, total_users, ...) were replaced with per-resource
// sub-objects. See data source docs for migration notes.
type StatsDataSourceModel struct {
	Users           types.Int64  `tfsdk:"users"`
	Organizations   types.Int64  `tfsdk:"organizations"`
	SCMProviders    types.Int64  `tfsdk:"scm_providers"`
	Downloads       types.Int64  `tfsdk:"downloads"`
	Modules         types.Object `tfsdk:"modules"`
	Providers       types.Object `tfsdk:"providers"`
	ProviderMirrors types.Object `tfsdk:"provider_mirrors"`
	BinaryMirrors   types.Object `tfsdk:"binary_mirrors"`
	RecentSyncs     types.List   `tfsdk:"recent_syncs"`
}

func NewStatsDataSource() datasource.DataSource {
	return &StatsDataSource{}
}

func (d *StatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stats"
}

func moduleSystemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system": types.StringType,
		"count":  types.Int64Type,
	}
}

func binaryToolAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tool":  types.StringType,
		"count": types.Int64Type,
	}
}

func moduleStatsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"total":     types.Int64Type,
		"versions":  types.Int64Type,
		"downloads": types.Int64Type,
		"by_system": types.ListType{ElemType: types.ObjectType{AttrTypes: moduleSystemAttrTypes()}},
	}
}

func providerStatsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"total":             types.Int64Type,
		"total_versions":    types.Int64Type,
		"manual":            types.Int64Type,
		"manual_versions":   types.Int64Type,
		"mirrored":          types.Int64Type,
		"mirrored_versions": types.Int64Type,
		"downloads":         types.Int64Type,
	}
}

func providerMirrorStatsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"total":   types.Int64Type,
		"healthy": types.Int64Type,
		"failed":  types.Int64Type,
	}
}

func binaryMirrorStatsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"total":     types.Int64Type,
		"healthy":   types.Int64Type,
		"failed":    types.Int64Type,
		"syncing":   types.Int64Type,
		"downloads": types.Int64Type,
		"platforms": types.Int64Type,
		"by_tool":   types.ListType{ElemType: types.ObjectType{AttrTypes: binaryToolAttrTypes()}},
	}
}

func recentSyncAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mirror_name":      types.StringType,
		"mirror_type":      types.StringType,
		"status":           types.StringType,
		"triggered_by":     types.StringType,
		"started_at":       types.StringType,
		"completed_at":     types.StringType,
		"versions_synced":  types.Int64Type,
		"platforms_synced": types.Int64Type,
	}
}

func (d *StatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads dashboard statistics from the registry. The shape mirrors the backend admin.DashboardStats response — the flat `total_*` counters from earlier provider versions are no longer present; use the nested `modules`, `providers`, `provider_mirrors`, and `binary_mirrors` blocks instead.",
		Attributes: map[string]schema.Attribute{
			"users":         schema.Int64Attribute{Computed: true, Description: "Total number of users."},
			"organizations": schema.Int64Attribute{Computed: true, Description: "Total number of organizations."},
			"scm_providers": schema.Int64Attribute{Computed: true, Description: "Total number of SCM provider integrations."},
			"downloads":     schema.Int64Attribute{Computed: true, Description: "Aggregate download count across all artifacts."},
			"modules": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Module counters.",
				Attributes: map[string]schema.Attribute{
					"total":     schema.Int64Attribute{Computed: true, Description: "Total module records."},
					"versions":  schema.Int64Attribute{Computed: true, Description: "Total module versions across all modules."},
					"downloads": schema.Int64Attribute{Computed: true, Description: "Aggregate module download count."},
					"by_system": schema.ListNestedAttribute{
						Computed:    true,
						Description: "Per-system breakdown.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"system": schema.StringAttribute{Computed: true},
								"count":  schema.Int64Attribute{Computed: true},
							},
						},
					},
				},
			},
			"providers": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Provider counters, split between manually published and mirror-synced records.",
				Attributes: map[string]schema.Attribute{
					"total":             schema.Int64Attribute{Computed: true},
					"total_versions":    schema.Int64Attribute{Computed: true},
					"manual":            schema.Int64Attribute{Computed: true, Description: "Providers published directly to this registry."},
					"manual_versions":   schema.Int64Attribute{Computed: true},
					"mirrored":          schema.Int64Attribute{Computed: true, Description: "Providers synced from upstream registries."},
					"mirrored_versions": schema.Int64Attribute{Computed: true},
					"downloads":         schema.Int64Attribute{Computed: true},
				},
			},
			"provider_mirrors": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Provider-mirror configuration health.",
				Attributes: map[string]schema.Attribute{
					"total":   schema.Int64Attribute{Computed: true},
					"healthy": schema.Int64Attribute{Computed: true, Description: "Last sync succeeded or never run but enabled."},
					"failed":  schema.Int64Attribute{Computed: true, Description: "Last sync failed."},
				},
			},
			"binary_mirrors": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Terraform/OpenTofu binary mirror configuration health.",
				Attributes: map[string]schema.Attribute{
					"total":     schema.Int64Attribute{Computed: true},
					"healthy":   schema.Int64Attribute{Computed: true},
					"failed":    schema.Int64Attribute{Computed: true},
					"syncing":   schema.Int64Attribute{Computed: true, Description: "Sync currently running."},
					"downloads": schema.Int64Attribute{Computed: true},
					"platforms": schema.Int64Attribute{Computed: true, Description: "Total synced platform binaries."},
					"by_tool": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"tool":  schema.StringAttribute{Computed: true},
								"count": schema.Int64Attribute{Computed: true},
							},
						},
					},
				},
			},
			"recent_syncs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Recent mirror sync runs (provider + binary), most recent first.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mirror_name":      schema.StringAttribute{Computed: true},
						"mirror_type":      schema.StringAttribute{Computed: true, Description: `"binary" or "provider".`},
						"status":           schema.StringAttribute{Computed: true},
						"triggered_by":     schema.StringAttribute{Computed: true},
						"started_at":       schema.StringAttribute{Computed: true},
						"completed_at":     schema.StringAttribute{Computed: true},
						"versions_synced":  schema.Int64Attribute{Computed: true},
						"platforms_synced": schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *StatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	stats, err := d.client.GetStats(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Stats", err.Error())
		return
	}

	model := StatsDataSourceModel{
		Users:         types.Int64Value(int64(stats.Users)),
		Organizations: types.Int64Value(int64(stats.Organizations)),
		SCMProviders:  types.Int64Value(int64(stats.SCMProviders)),
		Downloads:     types.Int64Value(stats.Downloads),
	}

	model.Modules = buildModulesObject(ctx, &stats.Modules, &resp.Diagnostics)
	model.Providers = buildProvidersObject(ctx, &stats.Providers, &resp.Diagnostics)
	model.ProviderMirrors = buildProviderMirrorsObject(ctx, &stats.ProviderMirrors, &resp.Diagnostics)
	model.BinaryMirrors = buildBinaryMirrorsObject(ctx, &stats.BinaryMirrors, &resp.Diagnostics)
	model.RecentSyncs = buildRecentSyncsList(ctx, stats.RecentSyncs, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func buildModulesObject(ctx context.Context, m *client.ModuleStats, diags *diag.Diagnostics) types.Object {
	bySystemList := types.ListNull(types.ObjectType{AttrTypes: moduleSystemAttrTypes()})
	if len(m.BySystem) > 0 {
		elems := make([]attr.Value, 0, len(m.BySystem))
		for _, e := range m.BySystem {
			obj, d := types.ObjectValue(moduleSystemAttrTypes(), map[string]attr.Value{
				"system": types.StringValue(e.System),
				"count":  types.Int64Value(int64(e.Count)),
			})
			diags.Append(d...)
			elems = append(elems, obj)
		}
		l, d := types.ListValue(types.ObjectType{AttrTypes: moduleSystemAttrTypes()}, elems)
		diags.Append(d...)
		bySystemList = l
	}
	obj, d := types.ObjectValue(moduleStatsAttrTypes(), map[string]attr.Value{
		"total":     types.Int64Value(int64(m.Total)),
		"versions":  types.Int64Value(int64(m.Versions)),
		"downloads": types.Int64Value(m.Downloads),
		"by_system": bySystemList,
	})
	diags.Append(d...)
	return obj
}

func buildProvidersObject(ctx context.Context, p *client.ProviderStats, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValue(providerStatsAttrTypes(), map[string]attr.Value{
		"total":             types.Int64Value(int64(p.Total)),
		"total_versions":    types.Int64Value(int64(p.TotalVersions)),
		"manual":            types.Int64Value(int64(p.Manual)),
		"manual_versions":   types.Int64Value(int64(p.ManualVersions)),
		"mirrored":          types.Int64Value(int64(p.Mirrored)),
		"mirrored_versions": types.Int64Value(int64(p.MirroredVersions)),
		"downloads":         types.Int64Value(p.Downloads),
	})
	diags.Append(d...)
	return obj
}

func buildProviderMirrorsObject(ctx context.Context, p *client.ProviderMirrorStats, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValue(providerMirrorStatsAttrTypes(), map[string]attr.Value{
		"total":   types.Int64Value(int64(p.Total)),
		"healthy": types.Int64Value(int64(p.Healthy)),
		"failed":  types.Int64Value(int64(p.Failed)),
	})
	diags.Append(d...)
	return obj
}

func buildBinaryMirrorsObject(ctx context.Context, b *client.BinaryMirrorStats, diags *diag.Diagnostics) types.Object {
	byToolList := types.ListNull(types.ObjectType{AttrTypes: binaryToolAttrTypes()})
	if len(b.ByTool) > 0 {
		elems := make([]attr.Value, 0, len(b.ByTool))
		for _, e := range b.ByTool {
			obj, d := types.ObjectValue(binaryToolAttrTypes(), map[string]attr.Value{
				"tool":  types.StringValue(e.Tool),
				"count": types.Int64Value(int64(e.Count)),
			})
			diags.Append(d...)
			elems = append(elems, obj)
		}
		l, d := types.ListValue(types.ObjectType{AttrTypes: binaryToolAttrTypes()}, elems)
		diags.Append(d...)
		byToolList = l
	}
	obj, d := types.ObjectValue(binaryMirrorStatsAttrTypes(), map[string]attr.Value{
		"total":     types.Int64Value(int64(b.Total)),
		"healthy":   types.Int64Value(int64(b.Healthy)),
		"failed":    types.Int64Value(int64(b.Failed)),
		"syncing":   types.Int64Value(int64(b.Syncing)),
		"downloads": types.Int64Value(b.Downloads),
		"platforms": types.Int64Value(int64(b.Platforms)),
		"by_tool":   byToolList,
	})
	diags.Append(d...)
	return obj
}

func buildRecentSyncsList(ctx context.Context, entries []client.RecentSyncEntry, diags *diag.Diagnostics) types.List {
	if len(entries) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: recentSyncAttrTypes()})
	}
	elems := make([]attr.Value, 0, len(entries))
	for _, e := range entries {
		obj, d := types.ObjectValue(recentSyncAttrTypes(), map[string]attr.Value{
			"mirror_name":      types.StringValue(e.MirrorName),
			"mirror_type":      types.StringValue(e.MirrorType),
			"status":           types.StringValue(e.Status),
			"triggered_by":     types.StringValue(e.TriggeredBy),
			"started_at":       types.StringValue(e.StartedAt),
			"completed_at":     types.StringValue(e.CompletedAt),
			"versions_synced":  types.Int64Value(int64(e.VersionsSynced)),
			"platforms_synced": types.Int64Value(int64(e.PlatformsSynced)),
		})
		diags.Append(d...)
		elems = append(elems, obj)
	}
	l, d := types.ListValue(types.ObjectType{AttrTypes: recentSyncAttrTypes()}, elems)
	diags.Append(d...)
	return l
}
