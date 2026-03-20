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

var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

type OrganizationResource struct {
	client *client.Client
}

type OrganizationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

func (r *OrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry organization (namespace).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the organization.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "URL-safe namespace name (e.g., 'my-org'). Used in module/provider paths.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the organization was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the organization was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *OrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.CreateOrganization(ctx, client.CreateOrganizationRequest{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Organization", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(org))...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Organization", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(org))...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrganization(ctx, plan.ID.ValueString(), client.UpdateOrganizationRequest{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Organization", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(org))...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOrganization(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Organization", err.Error())
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	org, err := r.client.GetOrganization(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Organization", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, orgToModel(org))...)
}

func orgToModel(o *client.Organization) OrganizationResourceModel {
	return OrganizationResourceModel{
		ID:          types.StringValue(o.ID),
		Name:        types.StringValue(o.Name),
		DisplayName: types.StringValue(o.DisplayName),
		CreatedAt:   types.StringValue(o.CreatedAt),
		UpdatedAt:   types.StringValue(o.UpdatedAt),
	}
}
