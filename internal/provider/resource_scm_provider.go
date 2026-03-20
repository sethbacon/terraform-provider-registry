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

var _ resource.Resource = &SCMProviderResource{}
var _ resource.ResourceWithImportState = &SCMProviderResource{}

type SCMProviderResource struct {
	client *client.Client
}

type SCMProviderResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	BaseURL     types.String `tfsdk:"base_url"`
	OAuthStatus types.String `tfsdk:"oauth_status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewSCMProviderResource() resource.Resource {
	return &SCMProviderResource{}
}

func (r *SCMProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scm_provider"
}

func (r *SCMProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a source control integration. OAuth token setup is performed separately via the registry UI or API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the SCM provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name for this SCM integration.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "SCM provider type: 'github', 'gitlab', 'azure', or 'bitbucket'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"base_url": schema.StringAttribute{
				Description: "Base URL for self-hosted SCM instances (e.g., 'https://github.mycompany.com').",
				Optional:    true,
				Computed:    true,
			},
			"oauth_status": schema.StringAttribute{
				Description: "Current OAuth token status.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the SCM provider was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the SCM provider was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *SCMProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SCMProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SCMProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateSCMProviderRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		createReq.BaseURL = &v
	}

	scm, err := r.client.CreateSCMProvider(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating SCM Provider", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, scmProviderToModel(scm))...)
}

func (r *SCMProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SCMProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scm, err := r.client.GetSCMProvider(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading SCM Provider", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, scmProviderToModel(scm))...)
}

func (r *SCMProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SCMProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateSCMProviderRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		updateReq.BaseURL = &v
	}

	scm, err := r.client.UpdateSCMProvider(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating SCM Provider", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, scmProviderToModel(scm))...)
}

func (r *SCMProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SCMProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSCMProvider(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting SCM Provider", err.Error())
	}
}

func (r *SCMProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scm, err := r.client.GetSCMProvider(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing SCM Provider", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, scmProviderToModel(scm))...)
}

func scmProviderToModel(s *client.SCMProvider) SCMProviderResourceModel {
	model := SCMProviderResourceModel{
		ID:        types.StringValue(s.ID),
		Name:      types.StringValue(s.Name),
		Type:      types.StringValue(s.Type),
		CreatedAt: types.StringValue(s.CreatedAt),
		UpdatedAt: types.StringValue(s.UpdatedAt),
	}
	if s.BaseURL != nil {
		model.BaseURL = types.StringValue(*s.BaseURL)
	} else {
		model.BaseURL = types.StringNull()
	}
	if s.OAuthStatus != nil {
		model.OAuthStatus = types.StringValue(*s.OAuthStatus)
	} else {
		model.OAuthStatus = types.StringNull()
	}
	return model
}
