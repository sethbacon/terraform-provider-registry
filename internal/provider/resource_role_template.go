package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ resource.Resource = &RoleTemplateResource{}
var _ resource.ResourceWithImportState = &RoleTemplateResource{}

type RoleTemplateResource struct {
	client *client.Client
}

type RoleTemplateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Scopes      types.List   `tfsdk:"scopes"`
	IsSystem    types.Bool   `tfsdk:"is_system"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewRoleTemplateResource() resource.Resource {
	return &RoleTemplateResource{}
}

func (r *RoleTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_template"
}

func (r *RoleTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry RBAC role template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the role template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Machine-readable name (e.g., 'publisher').",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the role.",
				Optional:    true,
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "List of permission scopes assigned to this role.",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_system": schema.BoolAttribute{
				Description: "Whether this is a built-in system role (cannot be deleted).",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the role template was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the role template was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *RoleTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateRoleTemplateRequest{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Scopes:      scopes,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}

	rt, err := r.client.CreateRoleTemplate(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleTemplateToModel(rt))...)
}

func (r *RoleTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rt, err := r.client.GetRoleTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Role Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleTemplateToModel(rt))...)
}

func (r *RoleTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateRoleTemplateRequest{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Scopes:      scopes,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}

	rt, err := r.client.UpdateRoleTemplate(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Role Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleTemplateToModel(rt))...)
}

func (r *RoleTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRoleTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Role Template", err.Error())
	}
}

func (r *RoleTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rt, err := r.client.GetRoleTemplate(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Role Template", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleTemplateToModel(rt))...)
}

func roleTemplateToModel(rt *client.RoleTemplate) RoleTemplateResourceModel {
	scopeValues := make([]types.String, len(rt.Scopes))
	for i, s := range rt.Scopes {
		scopeValues[i] = types.StringValue(s)
	}
	scopeList, _ := types.ListValueFrom(context.Background(), types.StringType, scopeValues)

	model := RoleTemplateResourceModel{
		ID:          types.StringValue(rt.ID),
		Name:        types.StringValue(rt.Name),
		DisplayName: types.StringValue(rt.DisplayName),
		Scopes:      scopeList,
		IsSystem:    types.BoolValue(rt.IsSystem),
		CreatedAt:   types.StringValue(normalizeTimestamp(rt.CreatedAt)),
		UpdatedAt:   types.StringValue(normalizeTimestamp(rt.UpdatedAt)),
	}
	if rt.Description != nil {
		model.Description = types.StringValue(*rt.Description)
	} else {
		model.Description = types.StringNull()
	}
	return model
}
