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

var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}

type APIKeyResource struct {
	client *client.Client
}

type APIKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Scopes         types.List   `tfsdk:"scopes"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	KeyPrefix      types.String `tfsdk:"key_prefix"`
	Key            types.String `tfsdk:"key"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry API key. The raw key value is only available at creation time and stored in state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization this key belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Friendly name for the API key.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description.",
				Optional:    true,
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "List of permission scopes (e.g., ['modules:read', 'modules:write']).",
				Required:    true,
				ElementType: types.StringType,
			},
			"expires_at": schema.StringAttribute{
				Description: "ISO 8601 expiration timestamp. Omit for no expiration.",
				Optional:    true,
				Computed:    true,
			},
			"key_prefix": schema.StringAttribute{
				Description: "First few characters of the key, for identification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The raw API key value. Only populated at creation time — store this securely.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp of last key use.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the key was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateAPIKeyRequest{
		OrganizationID: plan.OrganizationID.ValueString(),
		Name:           plan.Name.ValueString(),
		Scopes:         scopes,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueString()
		createReq.ExpiresAt = &v
	}

	created, err := r.client.CreateAPIKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating API Key", err.Error())
		return
	}

	model := apikeyToModel(&client.APIKey{
		ID:             created.ID,
		OrganizationID: plan.OrganizationID.ValueString(),
		Name:           created.Name,
		Description:    created.Description,
		KeyPrefix:      created.KeyPrefix,
		Scopes:         created.Scopes,
		ExpiresAt:      created.ExpiresAt,
		CreatedAt:      created.CreatedAt,
	})
	model.Key = types.StringValue(created.RawKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading API Key", err.Error())
		return
	}

	model := apikeyToModel(key)
	// Preserve the raw key from state — it's never returned after creation
	model.Key = state.Key
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan APIKeyResourceModel
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateAPIKeyRequest{
		Name:   plan.Name.ValueString(),
		Scopes: scopes,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueString()
		updateReq.ExpiresAt = &v
	}

	key, err := r.client.UpdateAPIKey(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating API Key", err.Error())
		return
	}

	model := apikeyToModel(key)
	model.Key = state.Key
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAPIKey(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting API Key", err.Error())
	}
}

func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	key, err := r.client.GetAPIKey(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing API Key", err.Error())
		return
	}

	model := apikeyToModel(key)
	// Raw key is not recoverable on import
	model.Key = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func apikeyToModel(k *client.APIKey) APIKeyResourceModel {
	scopeValues := make([]types.String, len(k.Scopes))
	for i, s := range k.Scopes {
		scopeValues[i] = types.StringValue(s)
	}
	scopeList, _ := types.ListValueFrom(context.Background(), types.StringType, scopeValues)

	model := APIKeyResourceModel{
		ID:             types.StringValue(k.ID),
		OrganizationID: types.StringValue(k.OrganizationID),
		Name:           types.StringValue(k.Name),
		KeyPrefix:      types.StringValue(k.KeyPrefix),
		Scopes:         scopeList,
		CreatedAt:      types.StringValue(k.CreatedAt),
	}
	if k.Description != nil {
		model.Description = types.StringValue(*k.Description)
	} else {
		model.Description = types.StringNull()
	}
	if k.ExpiresAt != nil {
		model.ExpiresAt = types.StringValue(*k.ExpiresAt)
	} else {
		model.ExpiresAt = types.StringNull()
	}
	if k.LastUsedAt != nil {
		model.LastUsedAt = types.StringValue(*k.LastUsedAt)
	} else {
		model.LastUsedAt = types.StringNull()
	}
	return model
}
