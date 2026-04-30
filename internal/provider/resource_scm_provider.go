package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	BaseURL        types.String `tfsdk:"base_url"`
	TenantID       types.String `tfsdk:"tenant_id"`
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	WebhookSecret  types.String `tfsdk:"webhook_secret"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
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
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization this SCM integration is scoped to. Omit for a global integration.",
				Optional:    true,
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
				Description: "Base URL for self-hosted SCM instances (e.g., 'https://github.mycompany.com'). Required for Bitbucket Data Center.",
				Optional:    true,
				Computed:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Tenant ID for Azure DevOps integrations.",
				Optional:    true,
				Computed:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "OAuth application client ID. Required for OAuth-based providers (github, gitlab, azure, bitbucket cloud). Not returned after creation.",
				Optional:    true,
				Computed:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "OAuth application client secret. Required for OAuth-based providers. Not returned after creation.",
				Optional:    true,
				Sensitive:   true,
			},
			"webhook_secret": schema.StringAttribute{
				Description: "Shared secret used to validate incoming webhook payloads. Not returned after creation.",
				Optional:    true,
				Sensitive:   true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the SCM integration is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
		Name:         plan.Name.ValueString(),
		ProviderType: plan.Type.ValueString(),
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := plan.OrganizationID.ValueString()
		createReq.OrganizationID = &v
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		createReq.BaseURL = &v
	}
	if !plan.TenantID.IsNull() && !plan.TenantID.IsUnknown() {
		v := plan.TenantID.ValueString()
		createReq.TenantID = &v
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() {
		createReq.ClientID = plan.ClientID.ValueString()
	}
	if !plan.ClientSecret.IsNull() && !plan.ClientSecret.IsUnknown() {
		createReq.ClientSecret = plan.ClientSecret.ValueString()
	}
	if !plan.WebhookSecret.IsNull() && !plan.WebhookSecret.IsUnknown() {
		createReq.WebhookSecret = plan.WebhookSecret.ValueString()
	}

	scm, err := r.client.CreateSCMProvider(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating SCM Provider", err.Error())
		return
	}

	model := scmProviderToModel(scm)
	// Preserve write-only credentials from plan — backend never returns the
	// secret values after creation.
	model.ClientSecret = plan.ClientSecret
	model.WebhookSecret = plan.WebhookSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
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

	model := scmProviderToModel(scm)
	// Preserve write-only secrets from state — not returned by API.
	model.ClientSecret = state.ClientSecret
	model.WebhookSecret = state.WebhookSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *SCMProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SCMProviderResourceModel
	var state SCMProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	updateReq := client.UpdateSCMProviderRequest{
		Name: &name,
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		updateReq.BaseURL = &v
	}
	if !plan.TenantID.IsNull() && !plan.TenantID.IsUnknown() {
		v := plan.TenantID.ValueString()
		updateReq.TenantID = &v
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() {
		v := plan.ClientID.ValueString()
		updateReq.ClientID = &v
	}
	if !plan.ClientSecret.IsNull() && !plan.ClientSecret.IsUnknown() {
		v := plan.ClientSecret.ValueString()
		updateReq.ClientSecret = &v
	}
	if !plan.WebhookSecret.IsNull() && !plan.WebhookSecret.IsUnknown() {
		v := plan.WebhookSecret.ValueString()
		updateReq.WebhookSecret = &v
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		updateReq.IsActive = &v
	}

	scm, err := r.client.UpdateSCMProvider(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating SCM Provider", err.Error())
		return
	}

	model := scmProviderToModel(scm)
	// Preserve write-only secrets from plan — not returned by API.
	model.ClientSecret = plan.ClientSecret
	model.WebhookSecret = plan.WebhookSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
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
	model := scmProviderToModel(scm)
	// Secrets are never returned by the API and cannot be recovered on import.
	model.ClientSecret = types.StringNull()
	model.WebhookSecret = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func scmProviderToModel(s *client.SCMProvider) SCMProviderResourceModel {
	model := SCMProviderResourceModel{
		ID:        types.StringValue(s.ID),
		Name:      types.StringValue(s.Name),
		Type:      types.StringValue(s.ProviderType),
		IsActive:  types.BoolValue(s.IsActive),
		CreatedAt: types.StringValue(normalizeTimestamp(s.CreatedAt)),
		UpdatedAt: types.StringValue(normalizeTimestamp(s.UpdatedAt)),
	}
	if s.OrganizationID != nil {
		model.OrganizationID = types.StringValue(*s.OrganizationID)
	} else {
		model.OrganizationID = types.StringNull()
	}
	if s.BaseURL != nil {
		model.BaseURL = types.StringValue(*s.BaseURL)
	} else {
		model.BaseURL = types.StringNull()
	}
	if s.TenantID != nil {
		model.TenantID = types.StringValue(*s.TenantID)
	} else {
		model.TenantID = types.StringNull()
	}
	if s.ClientID != "" {
		model.ClientID = types.StringValue(s.ClientID)
	} else {
		model.ClientID = types.StringNull()
	}
	return model
}
