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

var _ resource.Resource = &TerraformMirrorResource{}
var _ resource.ResourceWithImportState = &TerraformMirrorResource{}

type TerraformMirrorResource struct {
	client *client.Client
}

type TerraformMirrorResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Tool              types.String `tfsdk:"tool"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	UpstreamURL       types.String `tfsdk:"upstream_url"`
	PlatformFilter    types.List   `tfsdk:"platform_filter"`
	VersionFilter     types.String `tfsdk:"version_filter"`
	GPGVerify         types.Bool   `tfsdk:"gpg_verify"`
	StableOnly        types.Bool   `tfsdk:"stable_only"`
	SyncIntervalHours types.Int64  `tfsdk:"sync_interval_hours"`
	LastSyncAt        types.String `tfsdk:"last_sync_at"`
	LastSyncStatus    types.String `tfsdk:"last_sync_status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewTerraformMirrorResource() resource.Resource {
	return &TerraformMirrorResource{}
}

func (r *TerraformMirrorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_mirror"
}

func (r *TerraformMirrorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Terraform/OpenTofu binary mirror configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the Terraform mirror.",
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
			"tool": schema.StringAttribute{
				Description: "Tool to mirror: 'terraform', 'opentofu', or 'custom'.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether periodic syncing is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"upstream_url": schema.StringAttribute{
				Description: "URL of the upstream release source.",
				Required:    true,
			},
			"platform_filter": schema.ListAttribute{
				Description: "Platform allowlist in 'os/arch' format.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"version_filter": schema.StringAttribute{
				Description: "Version expression to filter which versions to mirror.",
				Optional:    true,
				Computed:    true,
			},
			"gpg_verify": schema.BoolAttribute{
				Description: "Whether to verify GPG signatures on downloaded binaries.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"stable_only": schema.BoolAttribute{
				Description: "Only mirror stable releases (no alpha/beta/rc).",
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
				Description: "Status of last sync.",
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

func (r *TerraformMirrorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TerraformMirrorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TerraformMirrorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateTerraformMirrorRequest{
		Name:              plan.Name.ValueString(),
		Tool:              plan.Tool.ValueString(),
		Enabled:           plan.Enabled.ValueBool(),
		UpstreamURL:       plan.UpstreamURL.ValueString(),
		GPGVerify:         plan.GPGVerify.ValueBool(),
		StableOnly:        plan.StableOnly.ValueBool(),
		SyncIntervalHours: int(plan.SyncIntervalHours.ValueInt64()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.VersionFilter.IsNull() && !plan.VersionFilter.IsUnknown() {
		v := plan.VersionFilter.ValueString()
		createReq.VersionFilter = &v
	}
	if !plan.PlatformFilter.IsNull() && !plan.PlatformFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.PlatformFilter.ElementsAs(ctx, &createReq.PlatformFilter, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.CreateTerraformMirror(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Terraform Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfMirrorToModel(ctx, mirror))...)
}

func (r *TerraformMirrorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TerraformMirrorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.GetTerraformMirror(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Terraform Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfMirrorToModel(ctx, mirror))...)
}

func (r *TerraformMirrorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TerraformMirrorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateTerraformMirrorRequest{
		Name:              plan.Name.ValueString(),
		Tool:              plan.Tool.ValueString(),
		Enabled:           plan.Enabled.ValueBool(),
		UpstreamURL:       plan.UpstreamURL.ValueString(),
		GPGVerify:         plan.GPGVerify.ValueBool(),
		StableOnly:        plan.StableOnly.ValueBool(),
		SyncIntervalHours: int(plan.SyncIntervalHours.ValueInt64()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.VersionFilter.IsNull() && !plan.VersionFilter.IsUnknown() {
		v := plan.VersionFilter.ValueString()
		updateReq.VersionFilter = &v
	}
	if !plan.PlatformFilter.IsNull() && !plan.PlatformFilter.IsUnknown() {
		resp.Diagnostics.Append(plan.PlatformFilter.ElementsAs(ctx, &updateReq.PlatformFilter, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mirror, err := r.client.UpdateTerraformMirror(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Terraform Mirror", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfMirrorToModel(ctx, mirror))...)
}

func (r *TerraformMirrorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TerraformMirrorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTerraformMirror(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Terraform Mirror", err.Error())
	}
}

func (r *TerraformMirrorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mirror, err := r.client.GetTerraformMirror(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Terraform Mirror", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfMirrorToModel(ctx, mirror))...)
}

func tfMirrorToModel(ctx context.Context, m *client.TerraformMirror) TerraformMirrorResourceModel {
	plf := m.PlatformFilter
	if plf == nil {
		plf = []string{}
	}
	platList, _ := types.ListValueFrom(ctx, types.StringType, plf)

	model := TerraformMirrorResourceModel{
		ID:                types.StringValue(m.ID),
		Name:              types.StringValue(m.Name),
		Tool:              types.StringValue(m.Tool),
		Enabled:           types.BoolValue(m.Enabled),
		UpstreamURL:       types.StringValue(m.UpstreamURL),
		GPGVerify:         types.BoolValue(m.GPGVerify),
		StableOnly:        types.BoolValue(m.StableOnly),
		SyncIntervalHours: types.Int64Value(int64(m.SyncIntervalHours)),
		PlatformFilter:    platList,
		CreatedAt:         types.StringValue(normalizeTimestamp(m.CreatedAt)),
		UpdatedAt:         types.StringValue(normalizeTimestamp(m.UpdatedAt)),
	}
	if m.Description != nil {
		model.Description = types.StringValue(*m.Description)
	} else {
		model.Description = types.StringNull()
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
