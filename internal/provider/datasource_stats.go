package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &StatsDataSource{}

type StatsDataSource struct {
	client *client.Client
}

type StatsDataSourceModel struct {
	TotalModules   types.Int64 `tfsdk:"total_modules"`
	TotalProviders types.Int64 `tfsdk:"total_providers"`
	TotalUsers     types.Int64 `tfsdk:"total_users"`
	TotalOrgs      types.Int64 `tfsdk:"total_organizations"`
	TotalMirrors   types.Int64 `tfsdk:"total_mirrors"`
	TotalAPIKeys   types.Int64 `tfsdk:"total_api_keys"`
}

func NewStatsDataSource() datasource.DataSource {
	return &StatsDataSource{}
}

func (d *StatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stats"
}

func (d *StatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads dashboard statistics from the registry.",
		Attributes: map[string]schema.Attribute{
			"total_modules":       schema.Int64Attribute{Computed: true, Description: "Total number of modules."},
			"total_providers":     schema.Int64Attribute{Computed: true, Description: "Total number of providers."},
			"total_users":         schema.Int64Attribute{Computed: true, Description: "Total number of users."},
			"total_organizations": schema.Int64Attribute{Computed: true, Description: "Total number of organizations."},
			"total_mirrors":       schema.Int64Attribute{Computed: true, Description: "Total number of mirrors."},
			"total_api_keys":      schema.Int64Attribute{Computed: true, Description: "Total number of API keys."},
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

	resp.Diagnostics.Append(resp.State.Set(ctx, StatsDataSourceModel{
		TotalModules:   types.Int64Value(int64(stats.TotalModules)),
		TotalProviders: types.Int64Value(int64(stats.TotalProviders)),
		TotalUsers:     types.Int64Value(int64(stats.TotalUsers)),
		TotalOrgs:      types.Int64Value(int64(stats.TotalOrgs)),
		TotalMirrors:   types.Int64Value(int64(stats.TotalMirrors)),
		TotalAPIKeys:   types.Int64Value(int64(stats.TotalAPIKeys)),
	})...)
}
