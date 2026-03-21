package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &AuditLogsDataSource{}

type AuditLogsDataSource struct {
	client *client.Client
}

type AuditLogsDataSourceModel struct {
	Action       types.String   `tfsdk:"action"`
	ResourceType types.String   `tfsdk:"resource_type"`
	Limit        types.Int64    `tfsdk:"limit"`
	Offset       types.Int64    `tfsdk:"offset"`
	Total        types.Int64    `tfsdk:"total"`
	AuditLogs    []AuditLogItem `tfsdk:"audit_logs"`
}

type AuditLogItem struct {
	ID             types.String `tfsdk:"id"`
	UserID         types.String `tfsdk:"user_id"`
	UserEmail      types.String `tfsdk:"user_email"`
	UserName       types.String `tfsdk:"user_name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Action         types.String `tfsdk:"action"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ResourceID     types.String `tfsdk:"resource_id"`
	Metadata       types.String `tfsdk:"metadata"`
	IPAddress      types.String `tfsdk:"ip_address"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func NewAuditLogsDataSource() datasource.DataSource {
	return &AuditLogsDataSource{}
}

func (d *AuditLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_logs"
}

func (d *AuditLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads audit log entries with optional filtering.",
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "Filter by action type (e.g., 'create', 'update', 'delete').",
				Optional:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Filter by resource type (e.g., 'module', 'user').",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of entries to return.",
				Optional:    true,
			},
			"offset": schema.Int64Attribute{
				Description: "Number of entries to skip.",
				Optional:    true,
			},
			"total": schema.Int64Attribute{
				Description: "Total number of matching audit log entries.",
				Computed:    true,
			},
			"audit_logs": schema.ListNestedAttribute{
				Description: "List of audit log entries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "UUID."},
						"user_id":         schema.StringAttribute{Computed: true, Description: "User UUID."},
						"user_email":      schema.StringAttribute{Computed: true, Description: "User email."},
						"user_name":       schema.StringAttribute{Computed: true, Description: "User name."},
						"organization_id": schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"action":          schema.StringAttribute{Computed: true, Description: "Action type."},
						"resource_type":   schema.StringAttribute{Computed: true, Description: "Resource type."},
						"resource_id":     schema.StringAttribute{Computed: true, Description: "Resource UUID."},
						"metadata":        schema.StringAttribute{Computed: true, Description: "JSON-encoded metadata."},
						"ip_address":      schema.StringAttribute{Computed: true, Description: "IP address."},
						"created_at":      schema.StringAttribute{Computed: true, Description: "Timestamp."},
					},
				},
			},
		},
	}
}

func (d *AuditLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuditLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AuditLogsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logs, total, err := d.client.ListAuditLogs(ctx,
		config.Action.ValueString(),
		config.ResourceType.ValueString(),
		int(config.Limit.ValueInt64()),
		int(config.Offset.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Audit Logs", err.Error())
		return
	}

	items := make([]AuditLogItem, len(logs))
	for i, l := range logs {
		metadataJSON := "{}"
		if l.Metadata != nil {
			if b, err := json.Marshal(l.Metadata); err == nil {
				metadataJSON = string(b)
			}
		}

		item := AuditLogItem{
			ID:        types.StringValue(l.ID),
			Action:    types.StringValue(l.Action),
			Metadata:  types.StringValue(metadataJSON),
			CreatedAt: types.StringValue(l.CreatedAt),
		}
		strOrNull := func(s *string) types.String {
			if s != nil {
				return types.StringValue(*s)
			}
			return types.StringNull()
		}
		item.UserID = strOrNull(l.UserID)
		item.UserEmail = strOrNull(l.UserEmail)
		item.UserName = strOrNull(l.UserName)
		item.OrganizationID = strOrNull(l.OrganizationID)
		item.ResourceType = strOrNull(l.ResourceType)
		item.ResourceID = strOrNull(l.ResourceID)
		item.IPAddress = strOrNull(l.IPAddress)
		items[i] = item
	}

	config.Total = types.Int64Value(int64(total))
	config.AuditLogs = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
