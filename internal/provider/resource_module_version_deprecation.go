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

var _ resource.Resource = &ModuleVersionDeprecationResource{}

type ModuleVersionDeprecationResource struct {
	client *client.Client
}

type ModuleVersionDeprecationResourceModel struct {
	Namespace         types.String `tfsdk:"namespace"`
	Name              types.String `tfsdk:"name"`
	System            types.String `tfsdk:"system"`
	Version           types.String `tfsdk:"version"`
	Message           types.String `tfsdk:"message"`
	ReplacementSource types.String `tfsdk:"replacement_source"`
	DeprecatedAt      types.String `tfsdk:"deprecated_at"`
}

func NewModuleVersionDeprecationResource() resource.Resource {
	return &ModuleVersionDeprecationResource{}
}

func (r *ModuleVersionDeprecationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_version_deprecation"
}

func (r *ModuleVersionDeprecationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Marks a specific module version as deprecated. Destroying this resource removes the deprecation.",
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
			"version": schema.StringAttribute{
				Description: "Semantic version string of the module version to deprecate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				Description: "Deprecation message shown to users.",
				Required:    true,
			},
			"replacement_source": schema.StringAttribute{
				Description: "Optional replacement source address (used by Terraform CLI ≥1.10 for upgrade guidance).",
				Optional:    true,
				Computed:    true,
			},
			"deprecated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the version was deprecated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ModuleVersionDeprecationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleVersionDeprecationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateModuleRequest{
		Message: plan.Message.ValueString(),
	}
	if !plan.ReplacementSource.IsNull() && !plan.ReplacementSource.IsUnknown() {
		v := plan.ReplacementSource.ValueString()
		dreq.ReplacementSource = &v
	}

	mv, err := r.client.DeprecateModuleVersion(ctx,
		plan.Namespace.ValueString(), plan.Name.ValueString(),
		plan.System.ValueString(), plan.Version.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deprecating Module Version", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleVersionDeprecationToModel(plan, mv))...)
}

func (r *ModuleVersionDeprecationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mv, err := r.client.GetModuleVersion(ctx,
		state.Namespace.ValueString(), state.Name.ValueString(),
		state.System.ValueString(), state.Version.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Module Version", err.Error())
		return
	}

	if !mv.Deprecated {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleVersionDeprecationToModel(state, mv))...)
}

func (r *ModuleVersionDeprecationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateModuleRequest{
		Message: plan.Message.ValueString(),
	}
	if !plan.ReplacementSource.IsNull() && !plan.ReplacementSource.IsUnknown() {
		v := plan.ReplacementSource.ValueString()
		dreq.ReplacementSource = &v
	}

	mv, err := r.client.DeprecateModuleVersion(ctx,
		plan.Namespace.ValueString(), plan.Name.ValueString(),
		plan.System.ValueString(), plan.Version.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Module Version Deprecation", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleVersionDeprecationToModel(plan, mv))...)
}

func (r *ModuleVersionDeprecationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModuleVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UndeprecateModuleVersion(ctx,
		state.Namespace.ValueString(), state.Name.ValueString(),
		state.System.ValueString(), state.Version.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Removing Module Version Deprecation", err.Error())
	}
}

func moduleVersionDeprecationToModel(ref ModuleVersionDeprecationResourceModel, mv *client.ModuleVersion) ModuleVersionDeprecationResourceModel {
	m := ModuleVersionDeprecationResourceModel{
		Namespace: ref.Namespace,
		Name:      ref.Name,
		System:    ref.System,
		Version:   types.StringValue(mv.Version),
		Message:   ref.Message,
	}
	if mv.DeprecationMessage != nil {
		m.Message = types.StringValue(*mv.DeprecationMessage)
	}
	if mv.ReplacementSource != nil {
		m.ReplacementSource = types.StringValue(*mv.ReplacementSource)
	} else {
		m.ReplacementSource = types.StringNull()
	}
	if mv.DeprecatedAt != nil {
		m.DeprecatedAt = types.StringValue(normalizeTimestamp(*mv.DeprecatedAt))
	} else {
		m.DeprecatedAt = types.StringValue("")
	}
	return m
}
