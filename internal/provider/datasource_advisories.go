package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &AdvisoriesDataSource{}

type AdvisoriesDataSource struct {
	client *client.Client
}

type AdvisoryItemModel struct {
	ID          types.String `tfsdk:"id"`
	ExternalID  types.String `tfsdk:"external_id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Severity    types.String `tfsdk:"severity"`
	URL         types.String `tfsdk:"url"`
	PublishedAt types.String `tfsdk:"published_at"`
	ResolvedAt  types.String `tfsdk:"resolved_at"`
	Active      types.Bool   `tfsdk:"active"`
}

type AdvisoriesDataSourceModel struct {
	ActiveOnly types.Bool          `tfsdk:"active_only"`
	Severities types.List          `tfsdk:"severities"`
	Advisories []AdvisoryItemModel `tfsdk:"advisories"`
}

func NewAdvisoriesDataSource() datasource.DataSource {
	return &AdvisoriesDataSource{}
}

func (d *AdvisoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_advisories"
}

func (d *AdvisoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists CVE security advisories from the registry. Use active_only = true to restrict to currently active advisories.",
		Attributes: map[string]schema.Attribute{
			"active_only": schema.BoolAttribute{
				Description: "If true, only returns currently active advisories. Defaults to false (returns all).",
				Optional:    true,
			},
			"severities": schema.ListAttribute{
				Description: "Optional list of severity values to filter by (e.g. ['high', 'critical']).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"advisories": schema.ListNestedAttribute{
				Description: "List of matching advisories.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Description: "Internal UUID.", Computed: true},
						"external_id":  schema.StringAttribute{Description: "CVE identifier (e.g. CVE-2024-12345).", Computed: true},
						"title":        schema.StringAttribute{Description: "Advisory title.", Computed: true},
						"description":  schema.StringAttribute{Description: "Advisory description.", Computed: true},
						"severity":     schema.StringAttribute{Description: "Severity level.", Computed: true},
						"url":          schema.StringAttribute{Description: "Reference URL.", Computed: true},
						"published_at": schema.StringAttribute{Description: "ISO 8601 publish timestamp.", Computed: true},
						"resolved_at":  schema.StringAttribute{Description: "ISO 8601 resolution timestamp (empty if unresolved).", Computed: true},
						"active":       schema.BoolAttribute{Description: "Whether the advisory is currently active.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *AdvisoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdvisoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AdvisoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	activeOnly := !state.ActiveOnly.IsNull() && !state.ActiveOnly.IsUnknown() && state.ActiveOnly.ValueBool()

	var severities []string
	if !state.Severities.IsNull() && !state.Severities.IsUnknown() {
		resp.Diagnostics.Append(state.Severities.ElementsAs(ctx, &severities, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	advisories, err := d.client.ListAdvisories(ctx, activeOnly, severities)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Advisories", err.Error())
		return
	}

	items := make([]AdvisoryItemModel, 0, len(advisories))
	for _, a := range advisories {
		item := AdvisoryItemModel{
			ID:         types.StringValue(a.ID),
			ExternalID: types.StringValue(a.ExternalID),
			Title:      types.StringValue(a.Title),
			Severity:   types.StringValue(a.Severity),
			Active:     types.BoolValue(a.Active),
		}
		optStr := func(s *string) types.String {
			if s != nil {
				return types.StringValue(*s)
			}
			return types.StringValue("")
		}
		item.Description = optStr(a.Description)
		item.URL = optStr(a.URL)
		if a.PublishedAt != nil {
			item.PublishedAt = types.StringValue(normalizeTimestamp(*a.PublishedAt))
		} else {
			item.PublishedAt = types.StringValue("")
		}
		if a.ResolvedAt != nil {
			item.ResolvedAt = types.StringValue(normalizeTimestamp(*a.ResolvedAt))
		} else {
			item.ResolvedAt = types.StringValue("")
		}
		items = append(items, item)
	}

	state.Advisories = items
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
