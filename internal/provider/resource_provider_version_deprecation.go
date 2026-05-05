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

var _ resource.Resource = &ProviderVersionDeprecationResource{}

type ProviderVersionDeprecationResource struct {
	client *client.Client
}

type ProviderVersionDeprecationResourceModel struct {
	Namespace    types.String `tfsdk:"namespace"`
	Type         types.String `tfsdk:"type"`
	Version      types.String `tfsdk:"version"`
	Message      types.String `tfsdk:"message"`
	DeprecatedAt types.String `tfsdk:"deprecated_at"`
}

func NewProviderVersionDeprecationResource() resource.Resource {
	return &ProviderVersionDeprecationResource{}
}

func (r *ProviderVersionDeprecationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_version_deprecation"
}

func (r *ProviderVersionDeprecationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Marks a specific provider version as deprecated. Destroying this resource removes the deprecation.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Namespace (organization name) owning the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Provider type name (e.g. 'aws', 'azurerm').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "Semantic version string of the provider version to deprecate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				Description: "Deprecation message shown to users.",
				Required:    true,
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

func (r *ProviderVersionDeprecationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProviderVersionDeprecationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProviderVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateProviderVersionRequest{
		Message: plan.Message.ValueString(),
	}

	pv, err := r.client.DeprecateProviderVersion(ctx,
		plan.Namespace.ValueString(), plan.Type.ValueString(), plan.Version.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deprecating Provider Version", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerVersionDeprecationToModel(plan, pv))...)
}

func (r *ProviderVersionDeprecationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProviderVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pv, err := r.client.GetProviderVersion(ctx,
		state.Namespace.ValueString(), state.Type.ValueString(), state.Version.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Provider Version", err.Error())
		return
	}

	if !pv.Deprecated {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerVersionDeprecationToModel(state, pv))...)
}

func (r *ProviderVersionDeprecationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProviderVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dreq := client.DeprecateProviderVersionRequest{
		Message: plan.Message.ValueString(),
	}

	pv, err := r.client.DeprecateProviderVersion(ctx,
		plan.Namespace.ValueString(), plan.Type.ValueString(), plan.Version.ValueString(), dreq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Provider Version Deprecation", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerVersionDeprecationToModel(plan, pv))...)
}

func (r *ProviderVersionDeprecationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProviderVersionDeprecationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UndeprecateProviderVersion(ctx,
		state.Namespace.ValueString(), state.Type.ValueString(), state.Version.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Removing Provider Version Deprecation", err.Error())
	}
}

func providerVersionDeprecationToModel(ref ProviderVersionDeprecationResourceModel, pv *client.ProviderVersion) ProviderVersionDeprecationResourceModel {
	m := ProviderVersionDeprecationResourceModel{
		Namespace: ref.Namespace,
		Type:      ref.Type,
		Version:   types.StringValue(pv.Version),
		Message:   ref.Message,
	}
	if pv.DeprecationMessage != nil {
		m.Message = types.StringValue(*pv.DeprecationMessage)
	}
	if pv.DeprecatedAt != nil {
		m.DeprecatedAt = types.StringValue(normalizeTimestamp(*pv.DeprecatedAt))
	} else {
		m.DeprecatedAt = types.StringValue("")
	}
	return m
}
