package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ resource.Resource = &ModuleReanalyzeResource{}

type ModuleReanalyzeResource struct {
	client *client.Client
}

type ModuleReanalyzeResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	System    types.String `tfsdk:"system"`
	Version   types.String `tfsdk:"version"`
	Triggers  types.Map    `tfsdk:"triggers"`
}

func NewModuleReanalyzeResource() resource.Resource {
	return &ModuleReanalyzeResource{}
}

func (r *ModuleReanalyzeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_reanalyze"
}

func (r *ModuleReanalyzeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers re-analysis of a module version (terraform-docs, scanning, SCM verification). Recreates whenever triggers change; otherwise no-op on subsequent applies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier: {namespace}/{name}/{system}/{version}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Module namespace.",
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
				Description: "Module system (provider name).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "Module version to reanalyze.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				Description: "Arbitrary key-value map. Changing any value forces the reanalyze action to run again.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ModuleReanalyzeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleReanalyzeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleReanalyzeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ReanalyzeModuleVersion(ctx,
		plan.Namespace.ValueString(), plan.Name.ValueString(),
		plan.System.ValueString(), plan.Version.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Triggering Module Reanalyze", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s",
		plan.Namespace.ValueString(), plan.Name.ValueString(),
		plan.System.ValueString(), plan.Version.ValueString()))

	if plan.Triggers.IsNull() || plan.Triggers.IsUnknown() {
		plan.Triggers = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ModuleReanalyzeResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No-op: this resource has no server-side persistent state to read back.
}

func (r *ModuleReanalyzeResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes have RequiresReplace; Update is never called.
}

func (r *ModuleReanalyzeResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to undo — analysis is idempotent.
}
