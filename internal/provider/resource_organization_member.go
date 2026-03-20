package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ resource.Resource = &OrganizationMemberResource{}
var _ resource.ResourceWithImportState = &OrganizationMemberResource{}

type OrganizationMemberResource struct {
	client *client.Client
}

type OrganizationMemberResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	OrganizationID          types.String `tfsdk:"organization_id"`
	UserID                  types.String `tfsdk:"user_id"`
	RoleTemplateID          types.String `tfsdk:"role_template_id"`
	RoleTemplateName        types.String `tfsdk:"role_template_name"`
	RoleTemplateDisplayName types.String `tfsdk:"role_template_display_name"`
	UserName                types.String `tfsdk:"user_name"`
	UserEmail               types.String `tfsdk:"user_email"`
	CreatedAt               types.String `tfsdk:"created_at"`
}

func NewOrganizationMemberResource() resource.Resource {
	return &OrganizationMemberResource{}
}

func (r *OrganizationMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *OrganizationMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's membership in a registry organization. Import using '<organization_id>/<user_id>'.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID in the format '<organization_id>/<user_id>'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of the user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_template_id": schema.StringAttribute{
				Description: "UUID of the role template to assign to this member.",
				Optional:    true,
				Computed:    true,
			},
			"role_template_name": schema.StringAttribute{
				Description: "Name of the assigned role template.",
				Computed:    true,
			},
			"role_template_display_name": schema.StringAttribute{
				Description: "Display name of the assigned role template.",
				Computed:    true,
			},
			"user_name": schema.StringAttribute{
				Description: "Name of the member user.",
				Computed:    true,
			},
			"user_email": schema.StringAttribute{
				Description: "Email of the member user.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the membership was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OrganizationMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addReq := client.AddMemberRequest{
		UserID: plan.UserID.ValueString(),
	}
	if !plan.RoleTemplateID.IsNull() && !plan.RoleTemplateID.IsUnknown() {
		v := plan.RoleTemplateID.ValueString()
		addReq.RoleTemplateID = &v
	}

	member, err := r.client.AddOrganizationMember(ctx, plan.OrganizationID.ValueString(), addReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Adding Organization Member", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, memberToModel(member))...)
}

func (r *OrganizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetOrganizationMember(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Organization Member", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, memberToModel(member))...)
}

func (r *OrganizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateMemberRequest{}
	if !plan.RoleTemplateID.IsNull() && !plan.RoleTemplateID.IsUnknown() {
		v := plan.RoleTemplateID.ValueString()
		updateReq.RoleTemplateID = &v
	}

	member, err := r.client.UpdateOrganizationMember(ctx, plan.OrganizationID.ValueString(), plan.UserID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Organization Member", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, memberToModel(member))...)
}

func (r *OrganizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RemoveOrganizationMember(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Removing Organization Member", err.Error())
	}
}

func (r *OrganizationMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected '<organization_id>/<user_id>', got: %s", req.ID))
		return
	}

	member, err := r.client.GetOrganizationMember(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Organization Member", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, memberToModel(member))...)
}

func memberToModel(m *client.OrganizationMember) OrganizationMemberResourceModel {
	model := OrganizationMemberResourceModel{
		ID:             types.StringValue(m.OrganizationID + "/" + m.UserID),
		OrganizationID: types.StringValue(m.OrganizationID),
		UserID:         types.StringValue(m.UserID),
		UserName:       types.StringValue(m.UserName),
		UserEmail:      types.StringValue(m.UserEmail),
		CreatedAt:      types.StringValue(m.CreatedAt),
	}
	if m.RoleTemplateID != nil {
		model.RoleTemplateID = types.StringValue(*m.RoleTemplateID)
	} else {
		model.RoleTemplateID = types.StringNull()
	}
	if m.RoleTemplateName != nil {
		model.RoleTemplateName = types.StringValue(*m.RoleTemplateName)
	} else {
		model.RoleTemplateName = types.StringNull()
	}
	if m.RoleTemplateDisplayName != nil {
		model.RoleTemplateDisplayName = types.StringValue(*m.RoleTemplateDisplayName)
	} else {
		model.RoleTemplateDisplayName = types.StringNull()
	}
	return model
}
