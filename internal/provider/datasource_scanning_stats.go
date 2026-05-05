package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &ScanningStatsDataSource{}

type ScanningStatsDataSource struct {
	client *client.Client
}

type ScanningStatsDataSourceModel struct {
	TotalScans    types.Int64 `tfsdk:"total_scans"`
	PassedScans   types.Int64 `tfsdk:"passed_scans"`
	FailedScans   types.Int64 `tfsdk:"failed_scans"`
	PendingScans  types.Int64 `tfsdk:"pending_scans"`
	TotalFindings types.Int64 `tfsdk:"total_findings"`
	CriticalCount types.Int64 `tfsdk:"critical_count"`
	HighCount     types.Int64 `tfsdk:"high_count"`
	MediumCount   types.Int64 `tfsdk:"medium_count"`
	LowCount      types.Int64 `tfsdk:"low_count"`
}

func NewScanningStatsDataSource() datasource.DataSource {
	return &ScanningStatsDataSource{}
}

func (d *ScanningStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scanning_stats"
}

func (d *ScanningStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads aggregate security scan statistics.",
		Attributes: map[string]schema.Attribute{
			"total_scans":    schema.Int64Attribute{Description: "Total number of scans.", Computed: true},
			"passed_scans":   schema.Int64Attribute{Description: "Number of scans that passed.", Computed: true},
			"failed_scans":   schema.Int64Attribute{Description: "Number of scans that failed.", Computed: true},
			"pending_scans":  schema.Int64Attribute{Description: "Number of scans pending or running.", Computed: true},
			"total_findings": schema.Int64Attribute{Description: "Total findings across all scans.", Computed: true},
			"critical_count": schema.Int64Attribute{Description: "Critical severity findings.", Computed: true},
			"high_count":     schema.Int64Attribute{Description: "High severity findings.", Computed: true},
			"medium_count":   schema.Int64Attribute{Description: "Medium severity findings.", Computed: true},
			"low_count":      schema.Int64Attribute{Description: "Low severity findings.", Computed: true},
		},
	}
}

func (d *ScanningStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScanningStatsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	stats, err := d.client.GetScanningStats(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Scanning Stats", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, ScanningStatsDataSourceModel{
		TotalScans:    types.Int64Value(int64(stats.TotalScans)),
		PassedScans:   types.Int64Value(int64(stats.PassedScans)),
		FailedScans:   types.Int64Value(int64(stats.FailedScans)),
		PendingScans:  types.Int64Value(int64(stats.PendingScans)),
		TotalFindings: types.Int64Value(int64(stats.TotalFindings)),
		CriticalCount: types.Int64Value(int64(stats.CriticalCount)),
		HighCount:     types.Int64Value(int64(stats.HighCount)),
		MediumCount:   types.Int64Value(int64(stats.MediumCount)),
		LowCount:      types.Int64Value(int64(stats.LowCount)),
	})...)
}
