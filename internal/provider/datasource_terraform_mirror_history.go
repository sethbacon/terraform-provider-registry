package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &TerraformMirrorHistoryDataSource{}

type TerraformMirrorHistoryDataSource struct {
	client *client.Client
}

type TerraformMirrorHistoryEntryModel struct {
	ID          types.String `tfsdk:"id"`
	StartedAt   types.String `tfsdk:"started_at"`
	CompletedAt types.String `tfsdk:"completed_at"`
	Status      types.String `tfsdk:"status"`
	Message     types.String `tfsdk:"message"`
	VersionsNew types.Int64  `tfsdk:"versions_new"`
}

type TerraformMirrorHistoryDataSourceModel struct {
	MirrorID types.String                       `tfsdk:"mirror_id"`
	History  []TerraformMirrorHistoryEntryModel `tfsdk:"history"`
}

func NewTerraformMirrorHistoryDataSource() datasource.DataSource {
	return &TerraformMirrorHistoryDataSource{}
}

func (d *TerraformMirrorHistoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_mirror_history"
}

func (d *TerraformMirrorHistoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the sync history entries for a Terraform/OpenTofu mirror configuration.",
		Attributes: map[string]schema.Attribute{
			"mirror_id": schema.StringAttribute{
				Description: "UUID of the Terraform mirror configuration.",
				Required:    true,
			},
			"history": schema.ListNestedAttribute{
				Description: "Sync history entries, newest first.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "UUID of the history record.",
							Computed:    true,
						},
						"started_at": schema.StringAttribute{
							Description: "ISO 8601 timestamp when the sync started.",
							Computed:    true,
						},
						"completed_at": schema.StringAttribute{
							Description: "ISO 8601 timestamp when the sync completed (empty if still running).",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Sync status: 'pending', 'running', 'succeeded', or 'failed'.",
							Computed:    true,
						},
						"message": schema.StringAttribute{
							Description: "Optional status message or error detail.",
							Computed:    true,
						},
						"versions_new": schema.Int64Attribute{
							Description: "Number of new versions discovered in this sync.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TerraformMirrorHistoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerraformMirrorHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TerraformMirrorHistoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := d.client.ListTerraformMirrorHistory(ctx, state.MirrorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Terraform Mirror History", err.Error())
		return
	}

	items := make([]TerraformMirrorHistoryEntryModel, 0, len(entries))
	for _, e := range entries {
		item := TerraformMirrorHistoryEntryModel{
			ID:          types.StringValue(e.ID),
			StartedAt:   types.StringValue(normalizeTimestamp(e.StartedAt)),
			Status:      types.StringValue(e.Status),
			VersionsNew: types.Int64Value(int64(e.VersionsNew)),
		}
		if e.CompletedAt != nil {
			item.CompletedAt = types.StringValue(normalizeTimestamp(*e.CompletedAt))
		} else {
			item.CompletedAt = types.StringValue("")
		}
		if e.Message != nil {
			item.Message = types.StringValue(*e.Message)
		} else {
			item.Message = types.StringValue("")
		}
		items = append(items, item)
	}

	state.History = items
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
