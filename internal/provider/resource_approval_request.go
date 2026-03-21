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

var _ resource.Resource = &ApprovalRequestResource{}
var _ resource.ResourceWithImportState = &ApprovalRequestResource{}

type ApprovalRequestResource struct {
	client *client.Client
}

type ApprovalRequestResourceModel struct {
	ID                types.String `tfsdk:"id"`
	MirrorID          types.String `tfsdk:"mirror_id"`
	ProviderNamespace types.String `tfsdk:"provider_namespace"`
	ProviderName      types.String `tfsdk:"provider_name"`
	Justification     types.String `tfsdk:"justification"`
	ReviewStatus      types.String `tfsdk:"review_status"`
	ReviewerID        types.String `tfsdk:"reviewer_id"`
	ReviewNote        types.String `tfsdk:"review_note"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func NewApprovalRequestResource() resource.Resource {
	return &ApprovalRequestResource{}
}

func (r *ApprovalRequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_approval_request"
}

func (r *ApprovalRequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a mirror approval request. The review (approve/reject) is performed by an admin separately.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the approval request.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mirror_id": schema.StringAttribute{
				Description: "UUID of the mirror this approval request is for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_namespace": schema.StringAttribute{
				Description: "Provider namespace to request access for (e.g., 'hashicorp').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_name": schema.StringAttribute{
				Description: "Specific provider name within the namespace. Omit to request access for the entire namespace.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"justification": schema.StringAttribute{
				Description: "Justification / reason for the request.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"review_status": schema.StringAttribute{
				Description: "Current review status: pending, approved, or rejected.",
				Computed:    true,
			},
			"reviewer_id": schema.StringAttribute{
				Description: "UUID of the reviewing user.",
				Computed:    true,
			},
			"review_note": schema.StringAttribute{
				Description: "Note from the reviewer.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the request was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the request was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ApprovalRequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApprovalRequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApprovalRequestResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateApprovalRequestRequest{
		MirrorConfigID:    plan.MirrorID.ValueString(),
		ProviderNamespace: plan.ProviderNamespace.ValueString(),
	}
	if !plan.ProviderName.IsNull() && !plan.ProviderName.IsUnknown() {
		v := plan.ProviderName.ValueString()
		createReq.ProviderName = &v
	}
	if !plan.Justification.IsNull() && !plan.Justification.IsUnknown() {
		createReq.Reason = plan.Justification.ValueString()
	}

	ar, err := r.client.CreateApprovalRequest(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Approval Request", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, approvalRequestToModel(ar))...)
}

func (r *ApprovalRequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApprovalRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ar, err := r.client.GetApprovalRequest(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Approval Request", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, approvalRequestToModel(ar))...)
}

func (r *ApprovalRequestResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Approval requests are immutable — any attribute change forces replace
}

func (r *ApprovalRequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApprovalRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApprovalRequest(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Approval Request", err.Error())
	}
}

func (r *ApprovalRequestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ar, err := r.client.GetApprovalRequest(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Approval Request", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, approvalRequestToModel(ar))...)
}

func approvalRequestToModel(a *client.ApprovalRequest) ApprovalRequestResourceModel {
	model := ApprovalRequestResourceModel{
		ID:                types.StringValue(a.ID),
		MirrorID:          types.StringValue(a.MirrorConfigID),
		ProviderNamespace: types.StringValue(a.ProviderNamespace),
		Justification:     types.StringValue(a.Reason),
		ReviewStatus:      types.StringValue(a.Status),
		CreatedAt:         types.StringValue(normalizeTimestamp(a.CreatedAt)),
		UpdatedAt:         types.StringValue(normalizeTimestamp(a.UpdatedAt)),
	}
	if a.ProviderName != nil {
		model.ProviderName = types.StringValue(*a.ProviderName)
	} else {
		model.ProviderName = types.StringNull()
	}
	if a.ReviewedBy != nil {
		model.ReviewerID = types.StringValue(*a.ReviewedBy)
	} else {
		model.ReviewerID = types.StringNull()
	}
	if a.ReviewNotes != nil {
		model.ReviewNote = types.StringValue(*a.ReviewNotes)
	} else {
		model.ReviewNote = types.StringNull()
	}
	return model
}
