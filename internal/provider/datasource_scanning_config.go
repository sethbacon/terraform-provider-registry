package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &ScanningConfigDataSource{}

type ScanningConfigDataSource struct {
	client *client.Client
}

type ScanningConfigDataSourceModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	BinaryPath      types.String `tfsdk:"binary_path"`
	DetectedVersion types.String `tfsdk:"detected_version"`
	EnabledTools    types.List   `tfsdk:"enabled_tools"`
}

func NewScanningConfigDataSource() datasource.DataSource {
	return &ScanningConfigDataSource{}
}

func (d *ScanningConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scanning_config"
}

func (d *ScanningConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the backend security scanner configuration.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether security scanning is enabled.",
				Computed:    true,
			},
			"binary_path": schema.StringAttribute{
				Description: "Path to the scanner binary.",
				Computed:    true,
			},
			"detected_version": schema.StringAttribute{
				Description: "Detected scanner version string.",
				Computed:    true,
			},
			"enabled_tools": schema.ListAttribute{
				Description: "List of enabled scanning tools.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ScanningConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScanningConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	cfg, err := d.client.GetScanningConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Scanning Config", err.Error())
		return
	}

	tools, diags := types.ListValueFrom(ctx, types.StringType, cfg.EnabledTools)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := ScanningConfigDataSourceModel{
		Enabled:      types.BoolValue(cfg.Enabled),
		EnabledTools: tools,
	}
	if cfg.BinaryPath != nil {
		model.BinaryPath = types.StringValue(*cfg.BinaryPath)
	} else {
		model.BinaryPath = types.StringValue("")
	}
	if cfg.DetectedVersion != nil {
		model.DetectedVersion = types.StringValue(*cfg.DetectedVersion)
	} else {
		model.DetectedVersion = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
