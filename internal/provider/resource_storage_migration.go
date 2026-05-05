package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

const (
	migrationPollInterval  = 15 * time.Second
	migrationDefaultTimeout = 60 * time.Minute
)

var _ resource.Resource = &StorageMigrationResource{}

type StorageMigrationResource struct {
	client *client.Client
}

type StorageMigrationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	SourceConfigID  types.String `tfsdk:"source_config_id"`
	TargetConfigID  types.String `tfsdk:"target_config_id"`
	TimeoutMinutes  types.Int64  `tfsdk:"timeout_minutes"`
	Status          types.String `tfsdk:"status"`
	ObjectsTotal    types.Int64  `tfsdk:"objects_total"`
	ObjectsMigrated types.Int64  `tfsdk:"objects_migrated"`
	ObjectsFailed   types.Int64  `tfsdk:"objects_failed"`
	StartedAt       types.String `tfsdk:"started_at"`
	CompletedAt     types.String `tfsdk:"completed_at"`
	Error           types.String `tfsdk:"error"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func NewStorageMigrationResource() resource.Resource {
	return &StorageMigrationResource{}
}

func (r *StorageMigrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_migration"
}

func (r *StorageMigrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a one-shot storage backend migration. Create starts the job and polls until it reaches a terminal state. Update is blocked (any input change forces replacement). Delete attempts cancellation if the job is still running.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the migration job.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_config_id": schema.StringAttribute{
				Description: "UUID of the storage config to migrate from.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_config_id": schema.StringAttribute{
				Description: "UUID of the storage config to migrate to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout_minutes": schema.Int64Attribute{
				Description: "Maximum time in minutes to wait for the migration to complete. Defaults to 60.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"status": schema.StringAttribute{
				Description: "Migration status: 'pending', 'running', 'succeeded', or 'failed'.",
				Computed:    true,
			},
			"objects_total": schema.Int64Attribute{
				Description: "Total number of objects to migrate.",
				Computed:    true,
			},
			"objects_migrated": schema.Int64Attribute{
				Description: "Number of objects successfully migrated.",
				Computed:    true,
			},
			"objects_failed": schema.Int64Attribute{
				Description: "Number of objects that failed to migrate.",
				Computed:    true,
			},
			"started_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the migration started.",
				Computed:    true,
			},
			"completed_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the migration completed.",
				Computed:    true,
			},
			"error": schema.StringAttribute{
				Description: "Error message if the migration failed.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the migration job was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "ISO 8601 timestamp when the migration job was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *StorageMigrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageMigrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StorageMigrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.CreateStorageMigration(ctx, client.CreateStorageMigrationRequest{
		SourceConfigID: plan.SourceConfigID.ValueString(),
		TargetConfigID: plan.TargetConfigID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Starting Storage Migration", err.Error())
		return
	}

	timeout := migrationDefaultTimeout
	if !plan.TimeoutMinutes.IsNull() && !plan.TimeoutMinutes.IsUnknown() {
		timeout = time.Duration(plan.TimeoutMinutes.ValueInt64()) * time.Minute
	}

	m, err = r.pollMigration(ctx, m.ID, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting For Migration", err.Error())
		return
	}

	model := migrationToModel(m)
	model.TimeoutMinutes = plan.TimeoutMinutes
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *StorageMigrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StorageMigrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetStorageMigration(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Storage Migration", err.Error())
		return
	}

	model := migrationToModel(m)
	model.TimeoutMinutes = state.TimeoutMinutes
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *StorageMigrationResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable fields have RequiresReplace, so Update is never called.
}

func (r *StorageMigrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StorageMigrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status := state.Status.ValueString()
	if status == "pending" || status == "running" {
		if _, err := r.client.CancelStorageMigration(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error Cancelling Storage Migration", err.Error())
		}
	}
}

// pollMigration polls until the migration reaches a terminal state or the timeout is exceeded.
func (r *StorageMigrationResource) pollMigration(ctx context.Context, id string, timeout time.Duration) (*client.StorageMigration, error) {
	deadline := time.Now().Add(timeout)
	for {
		m, err := r.client.GetStorageMigration(ctx, id)
		if err != nil {
			return nil, err
		}

		switch m.Status {
		case "succeeded":
			return m, nil
		case "failed":
			msg := "migration failed"
			if m.Error != nil {
				msg = *m.Error
			}
			return m, fmt.Errorf("%s", msg)
		}

		if time.Now().After(deadline) {
			return m, fmt.Errorf("timed out after %s waiting for migration %s (status: %s)", timeout, id, m.Status)
		}

		select {
		case <-ctx.Done():
			return m, ctx.Err()
		case <-time.After(migrationPollInterval):
		}
	}
}

func migrationToModel(m *client.StorageMigration) StorageMigrationResourceModel {
	model := StorageMigrationResourceModel{
		ID:              types.StringValue(m.ID),
		SourceConfigID:  types.StringValue(m.SourceConfigID),
		TargetConfigID:  types.StringValue(m.TargetConfigID),
		Status:          types.StringValue(m.Status),
		ObjectsTotal:    types.Int64Value(int64(m.ObjectsTotal)),
		ObjectsMigrated: types.Int64Value(int64(m.ObjectsMigrated)),
		ObjectsFailed:   types.Int64Value(int64(m.ObjectsFailed)),
		CreatedAt:       types.StringValue(normalizeTimestamp(m.CreatedAt)),
		UpdatedAt:       types.StringValue(normalizeTimestamp(m.UpdatedAt)),
	}
	if m.StartedAt != nil {
		model.StartedAt = types.StringValue(normalizeTimestamp(*m.StartedAt))
	} else {
		model.StartedAt = types.StringValue("")
	}
	if m.CompletedAt != nil {
		model.CompletedAt = types.StringValue(normalizeTimestamp(*m.CompletedAt))
	} else {
		model.CompletedAt = types.StringValue("")
	}
	if m.Error != nil {
		model.Error = types.StringValue(*m.Error)
	} else {
		model.Error = types.StringValue("")
	}
	return model
}
