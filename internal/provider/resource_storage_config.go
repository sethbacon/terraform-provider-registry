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

var _ resource.Resource = &StorageConfigResource{}
var _ resource.ResourceWithImportState = &StorageConfigResource{}

type StorageConfigResource struct {
	client *client.Client
}

type StorageConfigResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Backend   types.String `tfsdk:"backend"`
	Config    types.Map    `tfsdk:"config"`
	Active    types.Bool   `tfsdk:"active"`
	Activate  types.Bool   `tfsdk:"activate"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewStorageConfigResource() resource.Resource {
	return &StorageConfigResource{}
}

func (r *StorageConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_config"
}

func (r *StorageConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a registry storage backend configuration (local, S3, Azure Blob, GCS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the storage configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backend": schema.StringAttribute{
				Description: "Storage backend type: 'local', 's3', 'azure', or 'gcs'.",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Backend-specific configuration key-value pairs (e.g., bucket name, credentials). All values are stored encrypted.",
				Required:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"activate": schema.BoolAttribute{
				Description: "Set to true to make this the active storage backend after creation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"active": schema.BoolAttribute{
				Description: "Whether this is the currently active storage backend.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the config was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the config was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *StorageConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StorageConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configMap map[string]string
	resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &configMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateStorageConfigRequest{
		Backend:  plan.Backend.ValueString(),
		Config:   configMap,
		Activate: plan.Activate.ValueBool(),
	}

	sc, err := r.client.CreateStorageConfig(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Storage Config", err.Error())
		return
	}

	model := storageConfigToModel(ctx, sc)
	model.Activate = plan.Activate
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *StorageConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StorageConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sc, err := r.client.GetStorageConfig(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Storage Config", err.Error())
		return
	}

	model := storageConfigToModel(ctx, sc)
	model.Activate = state.Activate
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *StorageConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StorageConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configMap map[string]string
	resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &configMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sc, err := r.client.UpdateStorageConfig(ctx, plan.ID.ValueString(), client.UpdateStorageConfigRequest{
		Backend: plan.Backend.ValueString(),
		Config:  configMap,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Storage Config", err.Error())
		return
	}

	if plan.Activate.ValueBool() && !sc.Active {
		if err := r.client.ActivateStorageConfig(ctx, sc.ID); err != nil {
			resp.Diagnostics.AddError("Error Activating Storage Config", err.Error())
			return
		}
		sc.Active = true
	}

	model := storageConfigToModel(ctx, sc)
	model.Activate = plan.Activate
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *StorageConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StorageConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteStorageConfig(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Storage Config", err.Error())
	}
}

func (r *StorageConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sc, err := r.client.GetStorageConfig(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Storage Config", err.Error())
		return
	}
	model := storageConfigToModel(ctx, sc)
	model.Activate = types.BoolValue(false)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func storageConfigToModel(ctx context.Context, sc *client.StorageConfig) StorageConfigResourceModel {
	configMap, _ := types.MapValueFrom(ctx, types.StringType, sc.Config)
	return StorageConfigResourceModel{
		ID:        types.StringValue(sc.ID),
		Backend:   types.StringValue(sc.Backend),
		Config:    configMap,
		Active:    types.BoolValue(sc.Active),
		CreatedAt: types.StringValue(sc.CreatedAt),
		UpdatedAt: types.StringValue(sc.UpdatedAt),
	}
}
