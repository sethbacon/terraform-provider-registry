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

var _ resource.Resource = &PolicyResource{}
var _ resource.ResourceWithImportState = &PolicyResource{}

type PolicyResource struct {
	client *client.Client
}

type PolicyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	PolicyType       types.String `tfsdk:"policy_type"`
	UpstreamRegistry types.String `tfsdk:"upstream_registry"`
	NamespacePattern types.String `tfsdk:"namespace_pattern"`
	ProviderPattern  types.String `tfsdk:"provider_pattern"`
	Priority         types.Int64  `tfsdk:"priority"`
	IsActive         types.Bool   `tfsdk:"is_active"`
	RequiresApproval types.Bool   `tfsdk:"requires_approval"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mirror approval policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description.",
				Optional:    true,
				Computed:    true,
			},
			"policy_type": schema.StringAttribute{
				Description: "Policy type: 'allow' or 'deny'.",
				Required:    true,
			},
			"upstream_registry": schema.StringAttribute{
				Description: "Upstream registry URL to match.",
				Optional:    true,
				Computed:    true,
			},
			"namespace_pattern": schema.StringAttribute{
				Description: "Namespace pattern to match (supports wildcards).",
				Optional:    true,
				Computed:    true,
			},
			"provider_pattern": schema.StringAttribute{
				Description: "Provider pattern to match (supports wildcards).",
				Optional:    true,
				Computed:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "Policy evaluation priority (lower numbers evaluated first).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the policy is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"requires_approval": schema.BoolAttribute{
				Description: "Whether matching mirrors require manual approval.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the policy was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the policy was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreatePolicyRequest{
		Name:             plan.Name.ValueString(),
		PolicyType:       plan.PolicyType.ValueString(),
		Priority:         int(plan.Priority.ValueInt64()),
		IsActive:         plan.IsActive.ValueBool(),
		RequiresApproval: plan.RequiresApproval.ValueBool(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.UpstreamRegistry.IsNull() && !plan.UpstreamRegistry.IsUnknown() {
		v := plan.UpstreamRegistry.ValueString()
		createReq.UpstreamRegistry = &v
	}
	if !plan.NamespacePattern.IsNull() && !plan.NamespacePattern.IsUnknown() {
		v := plan.NamespacePattern.ValueString()
		createReq.NamespacePattern = &v
	}
	if !plan.ProviderPattern.IsNull() && !plan.ProviderPattern.IsUnknown() {
		v := plan.ProviderPattern.ValueString()
		createReq.ProviderPattern = &v
	}

	p, err := r.client.CreatePolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyToModel(p))...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.GetPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyToModel(p))...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdatePolicyRequest{
		Name:             plan.Name.ValueString(),
		PolicyType:       plan.PolicyType.ValueString(),
		Priority:         int(plan.Priority.ValueInt64()),
		IsActive:         plan.IsActive.ValueBool(),
		RequiresApproval: plan.RequiresApproval.ValueBool(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.UpstreamRegistry.IsNull() && !plan.UpstreamRegistry.IsUnknown() {
		v := plan.UpstreamRegistry.ValueString()
		updateReq.UpstreamRegistry = &v
	}
	if !plan.NamespacePattern.IsNull() && !plan.NamespacePattern.IsUnknown() {
		v := plan.NamespacePattern.ValueString()
		updateReq.NamespacePattern = &v
	}
	if !plan.ProviderPattern.IsNull() && !plan.ProviderPattern.IsUnknown() {
		v := plan.ProviderPattern.ValueString()
		updateReq.ProviderPattern = &v
	}

	p, err := r.client.UpdatePolicy(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyToModel(p))...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePolicy(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Policy", err.Error())
	}
}

func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	p, err := r.client.GetPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, policyToModel(p))...)
}

func policyToModel(p *client.Policy) PolicyResourceModel {
	model := PolicyResourceModel{
		ID:               types.StringValue(p.ID),
		Name:             types.StringValue(p.Name),
		PolicyType:       types.StringValue(p.PolicyType),
		Priority:         types.Int64Value(int64(p.Priority)),
		IsActive:         types.BoolValue(p.IsActive),
		RequiresApproval: types.BoolValue(p.RequiresApproval),
		CreatedAt:        types.StringValue(normalizeTimestamp(p.CreatedAt)),
		UpdatedAt:        types.StringValue(normalizeTimestamp(p.UpdatedAt)),
	}
	if p.Description != nil {
		model.Description = types.StringValue(*p.Description)
	} else {
		model.Description = types.StringNull()
	}
	if p.UpstreamRegistry != nil {
		model.UpstreamRegistry = types.StringValue(*p.UpstreamRegistry)
	} else {
		model.UpstreamRegistry = types.StringNull()
	}
	if p.NamespacePattern != nil {
		model.NamespacePattern = types.StringValue(*p.NamespacePattern)
	} else {
		model.NamespacePattern = types.StringNull()
	}
	if p.ProviderPattern != nil {
		model.ProviderPattern = types.StringValue(*p.ProviderPattern)
	} else {
		model.ProviderPattern = types.StringNull()
	}
	return model
}
