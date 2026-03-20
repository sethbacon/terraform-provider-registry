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

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	client *client.Client
}

type UserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
	OIDCSub   types.String `tfsdk:"oidc_sub"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry user account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Display name of the user.",
				Required:    true,
			},
			"oidc_sub": schema.StringAttribute{
				Description: "OIDC subject identifier. Set to link this user to an external identity provider subject.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the user was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the user was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateUserRequest{
		Email: plan.Email.ValueString(),
		Name:  plan.Name.ValueString(),
	}
	if !plan.OIDCSub.IsNull() && !plan.OIDCSub.IsUnknown() {
		v := plan.OIDCSub.ValueString()
		createReq.OIDCSub = &v
	}

	user, err := r.client.CreateUser(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating User", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(user))...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading User", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(user))...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateUserRequest{
		Email: plan.Email.ValueString(),
		Name:  plan.Name.ValueString(),
	}
	if !plan.OIDCSub.IsNull() && !plan.OIDCSub.IsUnknown() {
		v := plan.OIDCSub.ValueString()
		updateReq.OIDCSub = &v
	}

	user, err := r.client.UpdateUser(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating User", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(user))...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting User", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	user, err := r.client.GetUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing User", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(user))...)
}

func userToModel(u *client.User) UserResourceModel {
	m := UserResourceModel{
		ID:        types.StringValue(u.ID),
		Email:     types.StringValue(u.Email),
		Name:      types.StringValue(u.Name),
		CreatedAt: types.StringValue(u.CreatedAt),
		UpdatedAt: types.StringValue(u.UpdatedAt),
	}
	if u.OIDCSub != nil {
		m.OIDCSub = types.StringValue(*u.OIDCSub)
	} else {
		m.OIDCSub = types.StringNull()
	}
	return m
}
