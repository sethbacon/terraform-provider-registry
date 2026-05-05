package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &ScanDataSource{}

type ScanDataSource struct {
	client *client.Client
}

type ScanDataSourceModel struct {
	ScanDataModel
}

func NewScanDataSource() datasource.DataSource {
	return &ScanDataSource{}
}

func (d *ScanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scan"
}

func (d *ScanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single security scan by UUID.",
		Attributes:  scanResultSchema(),
	}
}

func (d *ScanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ScanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetScan(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Scan", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, ScanDataSourceModel{scanToModel(s)})...)
}

// scanResultSchema returns the shared schema for scan result attributes.
func scanResultSchema() map[string]schema.Attribute {
	findingAttrs := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Description: "UUID of the finding.", Computed: true},
			"rule_id":     schema.StringAttribute{Description: "Rule identifier.", Computed: true},
			"severity":    schema.StringAttribute{Description: "Finding severity.", Computed: true},
			"title":       schema.StringAttribute{Description: "Finding title.", Computed: true},
			"description": schema.StringAttribute{Description: "Finding description.", Computed: true},
			"resource":    schema.StringAttribute{Description: "Affected resource name.", Computed: true},
			"file_path":   schema.StringAttribute{Description: "Source file path.", Computed: true},
			"line_number": schema.Int64Attribute{Description: "Line number in the source file.", Computed: true},
		},
	}
	return map[string]schema.Attribute{
		"id":            schema.StringAttribute{Description: "UUID of the scan (required for registry_scan; computed for registry_module_scan).", Optional: true, Computed: true},
		"status":        schema.StringAttribute{Description: "Scan status.", Computed: true},
		"scanner":       schema.StringAttribute{Description: "Scanner tool name.", Computed: true},
		"passed":        schema.BoolAttribute{Description: "Whether the scan passed.", Computed: true},
		"execution_log": schema.StringAttribute{Description: "Raw scanner output.", Computed: true},
		"started_at":    schema.StringAttribute{Description: "ISO 8601 start timestamp.", Computed: true},
		"completed_at":  schema.StringAttribute{Description: "ISO 8601 completion timestamp.", Computed: true},
		"created_at":    schema.StringAttribute{Description: "ISO 8601 creation timestamp.", Computed: true},
		"findings": schema.ListNestedAttribute{
			Description:  "Security findings from the scan.",
			Computed:     true,
			NestedObject: findingAttrs,
		},
	}
}

// moduleScanSchema extends scanResultSchema with module-addressing attributes.
func moduleScanSchema() map[string]schema.Attribute {
	attrs := scanResultSchema()
	attrs["namespace"] = schema.StringAttribute{Description: "Module namespace.", Required: true}
	attrs["name"] = schema.StringAttribute{Description: "Module name.", Required: true}
	attrs["system"] = schema.StringAttribute{Description: "Module system.", Required: true}
	attrs["version"] = schema.StringAttribute{Description: "Module version.", Required: true}
	// id is computed only for module scan (not required)
	delete(attrs, "id")
	attrs["id"] = schema.StringAttribute{Description: "UUID of the scan.", Computed: true}
	return attrs
}

// ModuleScanDataSource fetches the most recent scan for a module version.
type ModuleScanDataSource struct {
	client *client.Client
}

var _ datasource.DataSource = &ModuleScanDataSource{}

type ModuleScanDataSourceModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	System    types.String `tfsdk:"system"`
	Version   types.String `tfsdk:"version"`
	ScanDataModel
}

func NewModuleScanDataSource() datasource.DataSource {
	return &ModuleScanDataSource{}
}

func (d *ModuleScanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_scan"
}

func (d *ModuleScanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the most recent security scan result for a specific module version.",
		Attributes:  moduleScanSchema(),
	}
}

func (d *ModuleScanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModuleScanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ModuleScanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetModuleScan(ctx,
		state.Namespace.ValueString(), state.Name.ValueString(),
		state.System.ValueString(), state.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Module Scan", err.Error())
		return
	}

	scanModel := scanToModel(s)
	result := ModuleScanDataSourceModel{
		Namespace:     state.Namespace,
		Name:          state.Name,
		System:        state.System,
		Version:       state.Version,
		ScanDataModel: scanModel,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}
