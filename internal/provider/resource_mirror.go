package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ resource.Resource = &MirrorResource{}
var _ resource.ResourceWithImportState = &MirrorResource{}

type MirrorResource struct {
	client *client.Client
}

type MirrorResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	UpstreamRegistryURL types.String `tfsdk:"upstream_registry_url"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	NamespaceFilter     types.List   `tfsdk:"namespace_filter"`
	ProviderFilter      types.List   `tfsdk:"provider_filter"`
	VersionFilter       types.String `tfsdk:"version_filter"`
	PlatformFilter      types.List   `tfsdk:"platform_filter"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	SyncIntervalHours   types.Int64  `tfsdk:"sync_interval_hours"`
	LastSyncAt          types.String `tfsdk:"last_sync_at"`
	LastSyncStatus      types.String `tfsdk:"last_sync_status"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func NewMirrorResource() resource.Resource {
	return &MirrorResource{}
}

func (r *MirrorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mirror"
}

func (r *MirrorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a provider mirror configuration for mirroring providers from an upstream registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the mirror.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the mirror configuration.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description.",
				Optional:    true,
				Computed:    true,
			},
			"upstream_registry_url": schema.StringAttribute{
				Description: "URL of the upstream registry to mirror from.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization to publish mirrored providers under.",
				Optional:    true,
				Computed:    true,
			},
			"namespace_filter": schema.ListAttribute{
				Description: "Namespace allowlist. Empty list means all namespaces.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"provider_filter": schema.ListAttribute{
				Description: "Provider type allowlist. Empty list means all providers.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"version_filter": schema.StringAttribute{
				Description: "Version expression (e.g., '>=3.0.0', 'latest:5').",
				Optional:    true,
				Computed:    true,
			},
			"platform_filter": schema.ListAttribute{
				Description: "Platform allowlist in 'os/arch' format (e.g., ['linux/amd64']).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether periodic syncing is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"sync_interval_hours": schema.Int64Attribute{
				Description: "How often to sync in hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(24),
			},
			"last_sync_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp of last sync.",
				Computed:    true,
			},
			"last_sync_status": schema.StringAttribute{
				Description: "Status of last sync: success, failed, or in_progress.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the mirror was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the mirror was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *MirrorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	r.client = c
}

func (r *MirrorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MirrorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateMirrorRequest{
		Name:                plan.Name.ValueString(),
		UpstreamRegistryURL: plan.UpstreamRegistryURL.ValueString(),
		Enabled:             plan.Enabled.ValueBool(),
		SyncIntervalHours:   int(plan.SyncIntervalHours.ValueInt64()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := plan.OrganizationID.ValueString()
		createReq.OrganizationID = &v
	}
	if !plan.VersionFilter.IsNull() && !plan.VersionFilter.IsUnknown() {
		v := plan.VersionFilter.ValueString()
		createReq.VersionFilter = &v
	}
	if !plan.NamespaceFilter.IsNull() && !plan.NamespaceFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.NamespaceFilter.ElementsAs(ctx, &createReq.NamespaceFilter, false)...)
	}
	if !plan.ProviderFilter.IsNull() && !plan.ProviderFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.ProviderFilter.ElementsAs(ctx, &createReq.ProviderFilter, false)...)
	}
	if !plan.PlatformFilter.IsNull() && !plan.PlatformFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.PlatformFilter.ElementsAs(ctx, &createReq.PlatformFilter, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.CreateMirror(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, mirrorToModel(ctx, mirror))...)
}

func (r *MirrorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MirrorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.GetMirror(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, mirrorToModel(ctx, mirror))...)
}

func (r *MirrorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MirrorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateMirrorRequest{
		Name:                plan.Name.ValueString(),
		UpstreamRegistryURL: plan.UpstreamRegistryURL.ValueString(),
		Enabled:             plan.Enabled.ValueBool(),
		SyncIntervalHours:   int(plan.SyncIntervalHours.ValueInt64()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := plan.OrganizationID.ValueString()
		updateReq.OrganizationID = &v
	}
	if !plan.VersionFilter.IsNull() && !plan.VersionFilter.IsUnknown() {
		v := plan.VersionFilter.ValueString()
		updateReq.VersionFilter = &v
	}
	if !plan.NamespaceFilter.IsNull() && !plan.NamespaceFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.NamespaceFilter.ElementsAs(ctx, &updateReq.NamespaceFilter, false)...)
	}
	if !plan.ProviderFilter.IsNull() && !plan.ProviderFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.ProviderFilter.ElementsAs(ctx, &updateReq.ProviderFilter, false)...)
	}
	if !plan.PlatformFilter.IsNull() && !plan.PlatformFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.PlatformFilter.ElementsAs(ctx, &updateReq.PlatformFilter, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.UpdateMirror(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, mirrorToModel(ctx, mirror))...)
}

func (r *MirrorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MirrorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMirror(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Mirror", err.Error())
	}
}

func (r *MirrorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mirror, err := r.client.GetMirror(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Mirror", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, mirrorToModel(ctx, mirror))...)
}

func mirrorToModel(ctx context.Context, m *client.Mirror) MirrorResourceModel {
	nsf := m.NamespaceFilter
	if nsf == nil {
		nsf = []string{}
	}
	prf := m.ProviderFilter
	if prf == nil {
		prf = []string{}
	}
	plf := m.PlatformFilter
	if plf == nil {
		plf = []string{}
	}
	nsList, _ := types.ListValueFrom(ctx, types.StringType, nsf)
	provList, _ := types.ListValueFrom(ctx, types.StringType, prf)
	platList, _ := types.ListValueFrom(ctx, types.StringType, plf)

	model := MirrorResourceModel{
		ID:                  types.StringValue(m.ID),
		Name:                types.StringValue(m.Name),
		UpstreamRegistryURL: types.StringValue(m.UpstreamRegistryURL),
		Enabled:             types.BoolValue(m.Enabled),
		SyncIntervalHours:   types.Int64Value(int64(m.SyncIntervalHours)),
		NamespaceFilter:     nsList,
		ProviderFilter:      provList,
		PlatformFilter:      platList,
		CreatedAt:           types.StringValue(normalizeTimestamp(m.CreatedAt)),
		UpdatedAt:           types.StringValue(normalizeTimestamp(m.UpdatedAt)),
	}
	if m.Description != nil {
		model.Description = types.StringValue(*m.Description)
	} else {
		model.Description = types.StringNull()
	}
	if m.OrganizationID != nil {
		model.OrganizationID = types.StringValue(*m.OrganizationID)
	} else {
		model.OrganizationID = types.StringNull()
	}
	if m.VersionFilter != nil {
		model.VersionFilter = types.StringValue(*m.VersionFilter)
	} else {
		model.VersionFilter = types.StringNull()
	}
	if m.LastSyncAt != nil {
		model.LastSyncAt = types.StringValue(*m.LastSyncAt)
	} else {
		model.LastSyncAt = types.StringNull()
	}
	if m.LastSyncStatus != nil {
		model.LastSyncStatus = types.StringValue(*m.LastSyncStatus)
	} else {
		model.LastSyncStatus = types.StringNull()
	}
	return model
}
