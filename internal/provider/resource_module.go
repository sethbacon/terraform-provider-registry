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

var _ resource.Resource = &ModuleResource{}
var _ resource.ResourceWithImportState = &ModuleResource{}

type ModuleResource struct {
	client *client.Client
}

type ModuleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	System         types.String `tfsdk:"system"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	CreatedBy      types.String `tfsdk:"created_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewModuleResource() resource.Resource {
	return &ModuleResource{}
}

func (r *ModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

func (r *ModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry module record. Module versions are uploaded separately via the registry API or SCM integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the module.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization that owns this module. Set by the server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Namespace (organization name) for the module.",
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
				Description: "Provider system (e.g., 'aws', 'azurerm', 'google').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the module.",
				Optional:    true,
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Optional source repository URL.",
				Optional:    true,
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "UUID of the user who created this module.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the module was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the module was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateModuleRequest{
		Namespace: plan.Namespace.ValueString(),
		Name:      plan.Name.ValueString(),
		System:    plan.System.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		v := plan.Source.ValueString()
		createReq.Source = &v
	}

	mod, err := r.client.CreateModule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Module", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleToModel(mod))...)
}

func (r *ModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleResourceModel
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

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleToModel(mod))...)
}

func (r *ModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateModuleRequest{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		v := plan.Source.ValueString()
		updateReq.Source = &v
	}

	mod, err := r.client.UpdateModule(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Module", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, moduleToModel(mod))...)
}

func (r *ModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteModule(ctx, state.Namespace.ValueString(), state.Name.ValueString(), state.System.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Module", err.Error())
	}
}

func (r *ModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mod, err := r.client.GetModuleByID(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Module", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, moduleToModel(mod))...)
}

func moduleToModel(m *client.Module) ModuleResourceModel {
	model := ModuleResourceModel{
		ID:             types.StringValue(m.ID),
		OrganizationID: types.StringValue(m.OrganizationID),
		Namespace:      types.StringValue(m.Namespace),
		Name:           types.StringValue(m.Name),
		System:         types.StringValue(m.System),
		CreatedAt:      types.StringValue(normalizeTimestamp(m.CreatedAt)),
		UpdatedAt:      types.StringValue(normalizeTimestamp(m.UpdatedAt)),
	}
	if m.Description != nil {
		model.Description = types.StringValue(*m.Description)
	} else {
		model.Description = types.StringNull()
	}
	if m.Source != nil {
		model.Source = types.StringValue(*m.Source)
	} else {
		model.Source = types.StringNull()
	}
	if m.CreatedBy != nil {
		model.CreatedBy = types.StringValue(*m.CreatedBy)
	} else {
		model.CreatedBy = types.StringNull()
	}
	return model
}
