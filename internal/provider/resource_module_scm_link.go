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

var _ resource.Resource = &ModuleSCMLinkResource{}
var _ resource.ResourceWithImportState = &ModuleSCMLinkResource{}

type ModuleSCMLinkResource struct {
	client *client.Client
}

type ModuleSCMLinkResourceModel struct {
	ModuleID      types.String `tfsdk:"module_id"`
	SCMProviderID types.String `tfsdk:"scm_provider_id"`
	Owner         types.String `tfsdk:"owner"`
	Repo          types.String `tfsdk:"repo"`
	Branch        types.String `tfsdk:"branch"`
	TagPattern    types.String `tfsdk:"tag_pattern"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewModuleSCMLinkResource() resource.Resource {
	return &ModuleSCMLinkResource{}
}

func (r *ModuleSCMLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_scm_link"
}

func (r *ModuleSCMLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Links a module to an SCM repository for automatic version publishing via webhooks.",
		Attributes: map[string]schema.Attribute{
			"module_id": schema.StringAttribute{
				Description: "UUID of the module to link.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scm_provider_id": schema.StringAttribute{
				Description: "UUID of the SCM provider integration.",
				Required:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Repository owner (organization or user).",
				Required:    true,
			},
			"repo": schema.StringAttribute{
				Description: "Repository name.",
				Required:    true,
			},
			"branch": schema.StringAttribute{
				Description: "Default branch to watch.",
				Required:    true,
			},
			"tag_pattern": schema.StringAttribute{
				Description: "Optional glob pattern for version tags (e.g., 'v*').",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the link was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the link was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ModuleSCMLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleSCMLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleSCMLinkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateModuleSCMLinkRequest{
		SCMProviderID:   plan.SCMProviderID.ValueString(),
		RepositoryOwner: plan.Owner.ValueString(),
		RepositoryName:  plan.Repo.ValueString(),
		DefaultBranch:   plan.Branch.ValueString(),
	}
	if !plan.TagPattern.IsNull() && !plan.TagPattern.IsUnknown() {
		createReq.TagPattern = plan.TagPattern.ValueString()
	}

	link, err := r.client.CreateModuleSCMLink(ctx, plan.ModuleID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Module SCM Link", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleSCMLinkToModel(link))...)
}

func (r *ModuleSCMLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleSCMLinkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	link, err := r.client.GetModuleSCMLink(ctx, state.ModuleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Module SCM Link", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleSCMLinkToModel(link))...)
}

func (r *ModuleSCMLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleSCMLinkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateModuleSCMLinkRequest{
		RepositoryOwner: plan.Owner.ValueString(),
		RepositoryName:  plan.Repo.ValueString(),
		DefaultBranch:   plan.Branch.ValueString(),
	}
	if !plan.TagPattern.IsNull() && !plan.TagPattern.IsUnknown() {
		updateReq.TagPattern = plan.TagPattern.ValueString()
	}

	link, err := r.client.UpdateModuleSCMLink(ctx, plan.ModuleID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Module SCM Link", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleSCMLinkToModel(link))...)
}

func (r *ModuleSCMLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModuleSCMLinkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteModuleSCMLink(ctx, state.ModuleID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Module SCM Link", err.Error())
	}
}

func (r *ModuleSCMLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	link, err := r.client.GetModuleSCMLink(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Module SCM Link", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, moduleSCMLinkToModel(link))...)
}

func moduleSCMLinkToModel(l *client.ModuleSCMLink) ModuleSCMLinkResourceModel {
	model := ModuleSCMLinkResourceModel{
		ModuleID:      types.StringValue(l.ModuleID),
		SCMProviderID: types.StringValue(l.SCMProviderID),
		Owner:         types.StringValue(l.RepositoryOwner),
		Repo:          types.StringValue(l.RepositoryName),
		Branch:        types.StringValue(l.DefaultBranch),
		CreatedAt:     types.StringValue(l.CreatedAt),
		UpdatedAt:     types.StringValue(l.UpdatedAt),
	}
	if l.TagPattern != "" {
		model.TagPattern = types.StringValue(l.TagPattern)
	} else {
		model.TagPattern = types.StringNull()
	}
	return model
}
