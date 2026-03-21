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

var _ resource.Resource = &ProviderRecordResource{}
var _ resource.ResourceWithImportState = &ProviderRecordResource{}

type ProviderRecordResource struct {
	client *client.Client
}

type ProviderRecordResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Namespace      types.String `tfsdk:"namespace"`
	Type           types.String `tfsdk:"type"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	CreatedBy      types.String `tfsdk:"created_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewProviderRecordResource() resource.Resource {
	return &ProviderRecordResource{}
}

func (r *ProviderRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_record"
}

func (r *ProviderRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry provider record. Provider binaries are uploaded separately.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the provider record.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization that owns this provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Namespace for the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Provider type (e.g., 'aws', 'azurerm', 'google').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the provider.",
				Optional:    true,
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Optional source repository URL.",
				Optional:    true,
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "UUID of the user who created this provider record.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the provider record was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the provider record was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ProviderRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProviderRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProviderRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateProviderRecordRequest{
		OrganizationID: plan.OrganizationID.ValueString(),
		Namespace:      plan.Namespace.ValueString(),
		Type:           plan.Type.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		v := plan.Source.ValueString()
		createReq.Source = &v
	}

	p, err := r.client.CreateProviderRecord(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Provider Record", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerRecordToModel(p))...)
}

func (r *ProviderRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProviderRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.GetProviderRecordByID(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Provider Record", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerRecordToModel(p))...)
}

func (r *ProviderRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProviderRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProviderRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateProviderRecordRequest{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		v := plan.Source.ValueString()
		updateReq.Source = &v
	}

	p, err := r.client.UpdateProviderRecord(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Provider Record", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerRecordToModel(p))...)
}

func (r *ProviderRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProviderRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProviderRecord(ctx, state.Namespace.ValueString(), state.Type.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Provider Record", err.Error())
	}
}

func (r *ProviderRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	p, err := r.client.GetProviderRecordByID(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Provider Record", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, providerRecordToModel(p))...)
}

func providerRecordToModel(p *client.ProviderRecord) ProviderRecordResourceModel {
	model := ProviderRecordResourceModel{
		ID:             types.StringValue(p.ID),
		OrganizationID: types.StringValue(p.OrganizationID),
		Namespace:      types.StringValue(p.Namespace),
		Type:           types.StringValue(p.Type),
		CreatedAt:      types.StringValue(p.CreatedAt),
		UpdatedAt:      types.StringValue(p.UpdatedAt),
	}
	if p.Description != nil {
		model.Description = types.StringValue(*p.Description)
	} else {
		model.Description = types.StringNull()
	}
	if p.Source != nil {
		model.Source = types.StringValue(*p.Source)
	} else {
		model.Source = types.StringNull()
	}
	if p.CreatedBy != nil {
		model.CreatedBy = types.StringValue(*p.CreatedBy)
	} else {
		model.CreatedBy = types.StringNull()
	}
	return model
}
