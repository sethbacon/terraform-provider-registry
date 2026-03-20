package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Rules       types.String `tfsdk:"rules"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
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
			"rules": schema.StringAttribute{
				Description: "JSON-encoded policy rules.",
				Required:    true,
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
		Name:  plan.Name.ValueString(),
		Rules: plan.Rules.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
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
		Name:  plan.Name.ValueString(),
		Rules: plan.Rules.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
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
		ID:        types.StringValue(p.ID),
		Name:      types.StringValue(p.Name),
		Rules:     types.StringValue(p.Rules),
		CreatedAt: types.StringValue(p.CreatedAt),
		UpdatedAt: types.StringValue(p.UpdatedAt),
	}
	if p.Description != nil {
		model.Description = types.StringValue(*p.Description)
	} else {
		model.Description = types.StringNull()
	}
	return model
}
