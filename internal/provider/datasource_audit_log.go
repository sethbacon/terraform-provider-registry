package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &AuditLogDataSource{}

type AuditLogDataSource struct {
	client *client.Client
}

type AuditLogDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	UserID         types.String `tfsdk:"user_id"`
	UserEmail      types.String `tfsdk:"user_email"`
	UserName       types.String `tfsdk:"user_name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Action         types.String `tfsdk:"action"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ResourceID     types.String `tfsdk:"resource_id"`
	MetadataJSON   types.String `tfsdk:"metadata_json"`
	IPAddress      types.String `tfsdk:"ip_address"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func NewAuditLogDataSource() datasource.DataSource {
	return &AuditLogDataSource{}
}

func (d *AuditLogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_log"
}

func (d *AuditLogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single audit log entry by UUID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the audit log entry to fetch.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of the user who performed the action.",
				Computed:    true,
			},
			"user_email": schema.StringAttribute{
				Description: "Email of the user who performed the action.",
				Computed:    true,
			},
			"user_name": schema.StringAttribute{
				Description: "Display name of the user.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization context.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "Action performed (e.g. 'create', 'update', 'delete').",
				Computed:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Type of resource affected.",
				Computed:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "UUID of the affected resource.",
				Computed:    true,
			},
			"metadata_json": schema.StringAttribute{
				Description: "Additional metadata serialised as JSON.",
				Computed:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "IP address of the requester.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp of the event.",
				Computed:    true,
			},
		},
	}
}

func (d *AuditLogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuditLogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AuditLogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := d.client.GetAuditLog(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Audit Log", err.Error())
		return
	}

	state.Action = types.StringValue(l.Action)
	state.CreatedAt = types.StringValue(normalizeTimestamp(l.CreatedAt))

	setOptStr := func(dest *types.String, src *string) {
		if src != nil {
			*dest = types.StringValue(*src)
		} else {
			*dest = types.StringValue("")
		}
	}
	setOptStr(&state.UserID, l.UserID)
	setOptStr(&state.UserEmail, l.UserEmail)
	setOptStr(&state.UserName, l.UserName)
	setOptStr(&state.OrganizationID, l.OrganizationID)
	setOptStr(&state.ResourceType, l.ResourceType)
	setOptStr(&state.ResourceID, l.ResourceID)
	setOptStr(&state.IPAddress, l.IPAddress)

	metaJSON := "{}"
	if l.Metadata != nil {
		if b, err := json.Marshal(l.Metadata); err == nil {
			metaJSON = string(b)
		}
	}
	state.MetadataJSON = types.StringValue(metaJSON)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
