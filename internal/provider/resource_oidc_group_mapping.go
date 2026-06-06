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

var _ resource.Resource = &OIDCGroupMappingResource{}

type OIDCGroupMappingResource struct {
	client *client.Client
}

type OIDCGroupMappingResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Group        types.String `tfsdk:"group"`
	Organization types.String `tfsdk:"organization"`
	Role         types.String `tfsdk:"role"`
}

func NewOIDCGroupMappingResource() resource.Resource {
	return &OIDCGroupMappingResource{}
}

func (r *OIDCGroupMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_group_mapping"
}

func (r *OIDCGroupMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single OIDC group → organization role mapping. The backend stores mappings as a full-replace list; this resource reads the current list, adds or removes its entry, and writes the list back.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier: {group}:{organization}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				Description: "OIDC groups claim value to match.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Organization name or ID this mapping applies to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "Role name to assign members of this group.",
				Required:    true,
			},
		},
	}
}

func (r *OIDCGroupMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OIDCGroupMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OIDCGroupMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetOIDCConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OIDC Config", err.Error())
		return
	}

	entry := client.OIDCGroupMapping{
		Group:        plan.Group.ValueString(),
		Organization: plan.Organization.ValueString(),
		Role:         plan.Role.ValueString(),
	}

	mappings := upsertMapping(cfg.GroupMappings, entry)
	input := client.OIDCGroupMappingInput{
		GroupClaimName: cfg.GroupClaimName,
		GroupMappings:  mappings,
		DefaultRole:    cfg.DefaultRole,
	}
	if _, err := r.client.SetOIDCGroupMappings(ctx, input); err != nil {
		resp.Diagnostics.AddError("Error Setting OIDC Group Mappings", err.Error())
		return
	}

	plan.ID = types.StringValue(oidcMappingID(plan.Group.ValueString(), plan.Organization.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OIDCGroupMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OIDCGroupMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mappings, err := r.client.GetOIDCGroupMappings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OIDC Group Mappings", err.Error())
		return
	}

	for _, m := range mappings {
		if m.Group == state.Group.ValueString() && m.Organization == state.Organization.ValueString() {
			state.Role = types.StringValue(m.Role)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *OIDCGroupMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OIDCGroupMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetOIDCConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OIDC Config", err.Error())
		return
	}

	entry := client.OIDCGroupMapping{
		Group:        plan.Group.ValueString(),
		Organization: plan.Organization.ValueString(),
		Role:         plan.Role.ValueString(),
	}

	mappings := upsertMapping(cfg.GroupMappings, entry)
	input := client.OIDCGroupMappingInput{
		GroupClaimName: cfg.GroupClaimName,
		GroupMappings:  mappings,
		DefaultRole:    cfg.DefaultRole,
	}
	if _, err := r.client.SetOIDCGroupMappings(ctx, input); err != nil {
		resp.Diagnostics.AddError("Error Setting OIDC Group Mappings", err.Error())
		return
	}

	plan.ID = types.StringValue(oidcMappingID(plan.Group.ValueString(), plan.Organization.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OIDCGroupMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OIDCGroupMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetOIDCConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OIDC Config", err.Error())
		return
	}

	mappings := removeMapping(cfg.GroupMappings, state.Group.ValueString(), state.Organization.ValueString())
	input := client.OIDCGroupMappingInput{
		GroupClaimName: cfg.GroupClaimName,
		GroupMappings:  mappings,
		DefaultRole:    cfg.DefaultRole,
	}
	if _, err := r.client.SetOIDCGroupMappings(ctx, input); err != nil {
		resp.Diagnostics.AddError("Error Setting OIDC Group Mappings", err.Error())
	}
}

func oidcMappingID(group, org string) string {
	return fmt.Sprintf("%s:%s", group, org)
}

func upsertMapping(current []client.OIDCGroupMapping, entry client.OIDCGroupMapping) []client.OIDCGroupMapping {
	for i, m := range current {
		if m.Group == entry.Group && m.Organization == entry.Organization {
			current[i] = entry
			return current
		}
	}
	return append(current, entry)
}

func removeMapping(current []client.OIDCGroupMapping, group, org string) []client.OIDCGroupMapping {
	result := make([]client.OIDCGroupMapping, 0, len(current))
	for _, m := range current {
		if m.Group != group || m.Organization != org {
			result = append(result, m)
		}
	}
	return result
}
