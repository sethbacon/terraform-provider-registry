package provider

import (
	"context"
	"fmt"

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
	OrganizationID    types.String `tfsdk:"organization_id"`
	ProviderNamespace types.String `tfsdk:"provider_namespace"`
	ProviderName      types.String `tfsdk:"provider_name"`
	Justification     types.String `tfsdk:"justification"`
	ReviewStatus      types.String `tfsdk:"review_status"`
	RequestedBy       types.String `tfsdk:"requested_by"`
	RequestedByName   types.String `tfsdk:"requested_by_name"`
	ReviewerID        types.String `tfsdk:"reviewer_id"`
	ReviewerName      types.String `tfsdk:"reviewer_name"`
	ReviewedAt        types.String `tfsdk:"reviewed_at"`
	ReviewNote        types.String `tfsdk:"review_note"`
	ExpiresAt         types.String `tfsdk:"expires_at"`
	AutoApproved      types.Bool   `tfsdk:"auto_approved"`
	MirrorName        types.String `tfsdk:"mirror_name"`
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
		// As with registry_policy, the backend only stores these requests and
		// their review status; nothing in mirror sync or pull-through reads
		// them. The description says so rather than implying a gate.
		Description: "Creates a mirror approval request for a provider namespace (or a single provider) on a mirror. The review (approve/reject) is performed by an admin separately. " +
			"The registry only records requests and their review status: mirror sync and pull-through do not check them, so a pending, approved or rejected request has no effect on what the mirror fetches or serves. " +
			"To control what a mirror fetches, use the `namespace_filter`, `provider_filter`, `version_filter` and `platform_filter` attributes of `registry_mirror`. " +
			"To review versions before Terraform clients can install them, use the registry's version approvals (the mirror configuration's own approval setting, which `registry_mirror` does not manage yet). " +
			"Approval requests cannot be deleted: the registry keeps every request as a review record and has no API to delete or withdraw one. " +
			"Destroying this resource, or replacing it because an argument changed, only removes it from Terraform state (with a warning); " +
			"the request stays in the registry with its current review status, and a pending request can still be approved or rejected by an admin.",
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
			"organization_id": schema.StringAttribute{
				Description: "UUID of the organization this request was scoped to (null for global).",
				Computed:    true,
			},
			"review_status": schema.StringAttribute{
				Description: "Current review status: pending, approved, or rejected.",
				Computed:    true,
			},
			"requested_by": schema.StringAttribute{
				Description: "UUID of the user who created the request.",
				Computed:    true,
			},
			"requested_by_name": schema.StringAttribute{
				Description: "Display name of the requesting user.",
				Computed:    true,
			},
			"reviewer_id": schema.StringAttribute{
				Description: "UUID of the reviewing user.",
				Computed:    true,
			},
			"reviewer_name": schema.StringAttribute{
				Description: "Display name of the reviewing user.",
				Computed:    true,
			},
			"reviewed_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the request was reviewed.",
				Computed:    true,
			},
			"review_note": schema.StringAttribute{
				Description: "Note from the reviewer.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when an approved request expires; null if it does not expire.",
				Computed:    true,
			},
			"auto_approved": schema.BoolAttribute{
				Description: "Whether the request was auto-approved. The registry does not currently auto-approve requests, so this is false.",
				Computed:    true,
			},
			"mirror_name": schema.StringAttribute{
				Description: "Name of the mirror this approval request applies to.",
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

// Delete forgets the approval request; it never calls the registry. Approval
// requests are review records, and the backend serves list, get, create,
// PUT .../review and POST .../token for them but no DELETE or withdraw. This
// used to send DELETE /api/v1/admin/approvals/{id}, which matched no route;
// Client.Delete treats a 404 on delete as success, so destroy reported success
// while the request stayed pending in the registry. Mapping destroy onto
// PUT .../review (a rejection) would be wrong too: that is an admin's review
// decision, not the requester's cleanup.
//
// Returning without an error lets the framework drop the resource from state;
// the warning makes it visible in the run that the request lives on.
func (r *ApprovalRequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApprovalRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail := fmt.Sprintf("Approval request %s was removed from Terraform state only. "+
		"The registry has no API to delete or withdraw an approval request, so it remains in the registry", state.ID.ValueString())
	if status := state.ReviewStatus.ValueString(); status != "" {
		detail += fmt.Sprintf(" with its current review status (last read as %q)", status)
	}
	detail += ". A pending request can still be approved or rejected by an admin."
	resp.Diagnostics.AddWarning("Approval Request Remains in the Registry", detail)
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
		AutoApproved:      types.BoolValue(a.AutoApproved),
		CreatedAt:         types.StringValue(normalizeTimestamp(a.CreatedAt)),
		UpdatedAt:         types.StringValue(normalizeTimestamp(a.UpdatedAt)),
	}
	if a.ProviderName != nil {
		model.ProviderName = types.StringValue(*a.ProviderName)
	} else {
		model.ProviderName = types.StringNull()
	}
	if a.OrganizationID != nil {
		model.OrganizationID = types.StringValue(*a.OrganizationID)
	} else {
		model.OrganizationID = types.StringNull()
	}
	if a.RequestedBy != nil {
		model.RequestedBy = types.StringValue(*a.RequestedBy)
	} else {
		model.RequestedBy = types.StringNull()
	}
	if a.RequestedByName != "" {
		model.RequestedByName = types.StringValue(a.RequestedByName)
	} else {
		model.RequestedByName = types.StringNull()
	}
	if a.ReviewedBy != nil {
		model.ReviewerID = types.StringValue(*a.ReviewedBy)
	} else {
		model.ReviewerID = types.StringNull()
	}
	if a.ReviewedByName != "" {
		model.ReviewerName = types.StringValue(a.ReviewedByName)
	} else {
		model.ReviewerName = types.StringNull()
	}
	if a.ReviewedAt != nil {
		model.ReviewedAt = types.StringValue(normalizeTimestamp(*a.ReviewedAt))
	} else {
		model.ReviewedAt = types.StringNull()
	}
	if a.ReviewNotes != nil {
		model.ReviewNote = types.StringValue(*a.ReviewNotes)
	} else {
		model.ReviewNote = types.StringNull()
	}
	if a.ExpiresAt != nil {
		model.ExpiresAt = types.StringValue(normalizeTimestamp(*a.ExpiresAt))
	} else {
		model.ExpiresAt = types.StringNull()
	}
	if a.MirrorName != "" {
		model.MirrorName = types.StringValue(a.MirrorName)
	} else {
		model.MirrorName = types.StringNull()
	}
	return model
}
