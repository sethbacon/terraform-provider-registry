package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
	"github.com/terraform-registry/terraform-provider-registry/internal/client/spec"
)

var _ datasource.DataSource = &UsersDataSource{}

type UsersDataSource struct {
	client *client.Client
}

type UsersDataSourceModel struct {
	Search types.String `tfsdk:"search"`
	Users  []UserModel  `tfsdk:"users"`
}

type UserModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all registry users, with optional search filtering.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Description: "Optional search string to filter users by name or email.",
				Optional:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "UUID of the user."},
						"email":      schema.StringAttribute{Computed: true, Description: "User email."},
						"name":       schema.StringAttribute{Computed: true, Description: "User display name."},
						"created_at": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
						"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.ListUsers(ctx, config.Search.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Users", err.Error())
		return
	}

	config.Users = make([]UserModel, len(users))
	for i, u := range users {
		config.Users[i] = specUserToModel(&u)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func specUserToModel(u *spec.User) UserModel {
	deref := func(s *string) types.String {
		if s == nil {
			return types.StringValue("")
		}
		return types.StringValue(*s)
	}
	return UserModel{
		ID:        deref(u.Id),
		Email:     deref(u.Email),
		Name:      deref(u.Name),
		CreatedAt: deref(u.CreatedAt),
		UpdatedAt: deref(u.UpdatedAt),
	}
}
