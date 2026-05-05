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

var _ resource.Resource = &ModuleDeprecationResource{}

type ModuleDeprecationResource struct {
	client *client.Client
}

type ModuleDeprecationResourceModel struct {
	Namespace         types.String `tfsdk:"namespace"`
	Name              types.String `tfsdk:"name"`
	System            types.String `tfsdk:"system"`
	Message           types.String `tfsdk:"message"`
	SuccessorModuleID types.String `tfsdk:"successor_module_id"`
	DeprecatedAt      types.String `tfsdk:"deprecated_at"`
}

func NewModuleDeprecationResource() resource.Resource {
	return &ModuleDeprecationResource{}
}

func (r *ModuleDeprecationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_deprecation"
}

func (r *ModuleDeprecationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Marks a module as deprecated. Destroying this resource removes the deprecation.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Namespace (organization name) owning the module.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Module name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"system": schema.StringAttribute{
				Description: "Module system (provider name, e.g. 'aws').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				Description: "Deprecation message shown to users.",
				Required:    true,
			},
			"successor_module_id": schema.StringAttribute{
				Description: "Optional UUID of the successor module (used by Terraform CLI ≥1.10 for replacement guidance).",
				Optional:    true,
				Computed:    true,
			},
			"deprecated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the module was deprecated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ModuleDeprecationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleDeprecationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateModuleRequest{
		Message: plan.Message.ValueString(),
	}
	if !plan.SuccessorModuleID.IsNull() && !plan.SuccessorModuleID.IsUnknown() {
		v := plan.SuccessorModuleID.ValueString()
		dreq.SuccessorModuleID = &v
	}

	mod, err := r.client.DeprecateModule(ctx, plan.Namespace.ValueString(), plan.Name.ValueString(), plan.System.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deprecating Module", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleDeprecationToModel(plan, mod))...)
}

func (r *ModuleDeprecationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mod, err := r.client.GetModule(ctx, state.Namespace.ValueString(), state.Name.ValueString(), state.System.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Module", err.Error())
		return
	}

	if !mod.Deprecated {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleDeprecationToModel(state, mod))...)
}

func (r *ModuleDeprecationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateModuleRequest{
		Message: plan.Message.ValueString(),
	}
	if !plan.SuccessorModuleID.IsNull() && !plan.SuccessorModuleID.IsUnknown() {
		v := plan.SuccessorModuleID.ValueString()
		dreq.SuccessorModuleID = &v
	}

	mod, err := r.client.DeprecateModule(ctx, plan.Namespace.ValueString(), plan.Name.ValueString(), plan.System.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Module Deprecation", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleDeprecationToModel(plan, mod))...)
}

func (r *ModuleDeprecationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModuleDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UndeprecateModule(ctx, state.Namespace.ValueString(), state.Name.ValueString(), state.System.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Removing Module Deprecation", err.Error())
	}
}

func moduleDeprecationToModel(ref ModuleDeprecationResourceModel, mod *client.Module) ModuleDeprecationResourceModel {
	m := ModuleDeprecationResourceModel{
		Namespace: types.StringValue(mod.Namespace),
		Name:      types.StringValue(mod.Name),
		System:    types.StringValue(mod.System),
		Message:   ref.Message,
	}
	if mod.DeprecationMessage != nil {
		m.Message = types.StringValue(*mod.DeprecationMessage)
	}
	if mod.SuccessorModuleID != nil {
		m.SuccessorModuleID = types.StringValue(*mod.SuccessorModuleID)
	} else {
		m.SuccessorModuleID = types.StringNull()
	}
	if mod.DeprecatedAt != nil {
		m.DeprecatedAt = types.StringValue(normalizeTimestamp(*mod.DeprecatedAt))
	} else {
		m.DeprecatedAt = types.StringValue("")
	}
	return m
}
