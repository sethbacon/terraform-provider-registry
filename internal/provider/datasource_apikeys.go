package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &APIKeysDataSource{}

type APIKeysDataSource struct {
	client *client.Client
}

type APIKeysDataSourceModel struct {
	UserID  types.String   `tfsdk:"user_id"`
	APIKeys []APIKeyDSItem `tfsdk:"api_keys"`
}

type APIKeyDSItem struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	KeyPrefix      types.String `tfsdk:"key_prefix"`
	Scopes         types.List   `tfsdk:"scopes"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{}
}

func (d *APIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *APIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists API keys, optionally filtered by user.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Description: "Filter by user UUID.",
				Optional:    true,
			},
			"api_keys": schema.ListNestedAttribute{
				Description: "List of API keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "UUID."},
						"organization_id": schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"name":            schema.StringAttribute{Computed: true, Description: "Key name."},
						"description":     schema.StringAttribute{Computed: true, Description: "Description."},
						"key_prefix":      schema.StringAttribute{Computed: true, Description: "Key prefix for identification."},
						"scopes": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Permission scopes.",
						},
						"expires_at":   schema.StringAttribute{Computed: true, Description: "Expiration timestamp."},
						"last_used_at": schema.StringAttribute{Computed: true, Description: "Last use timestamp."},
						"created_at":   schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
					},
				},
			},
		},
	}
}

func (d *APIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *APIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config APIKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := d.client.ListAPIKeys(ctx, config.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing API Keys", err.Error())
		return
	}

	items := make([]APIKeyDSItem, len(keys))
	for i, k := range keys {
		scopeValues := make([]types.String, len(k.Scopes))
		for j, s := range k.Scopes {
			scopeValues[j] = types.StringValue(s)
		}
		scopeList, _ := types.ListValueFrom(ctx, types.StringType, scopeValues)

		item := APIKeyDSItem{
			ID:             types.StringValue(k.ID),
			OrganizationID: types.StringValue(k.OrganizationID),
			Name:           types.StringValue(k.Name),
			KeyPrefix:      types.StringValue(k.KeyPrefix),
			Scopes:         scopeList,
			CreatedAt:      types.StringValue(k.CreatedAt),
		}
		if k.Description != nil {
			item.Description = types.StringValue(*k.Description)
		} else {
			item.Description = types.StringNull()
		}
		if k.ExpiresAt != nil {
			item.ExpiresAt = types.StringValue(*k.ExpiresAt)
		} else {
			item.ExpiresAt = types.StringNull()
		}
		if k.LastUsedAt != nil {
			item.LastUsedAt = types.StringValue(*k.LastUsedAt)
		} else {
			item.LastUsedAt = types.StringNull()
		}
		items[i] = item
	}

	config.APIKeys = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
