package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &MTLSConfigDataSource{}

type MTLSConfigDataSource struct {
	client *client.Client
}

type MTLSConfigDataSourceModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	ClientCACN types.String `tfsdk:"client_ca_cn"`
	ServerCert types.String `tfsdk:"server_cert_subject"`
}

func NewMTLSConfigDataSource() datasource.DataSource {
	return &MTLSConfigDataSource{}
}

func (d *MTLSConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mtls_config"
}

func (d *MTLSConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the backend mTLS server certificate and CA configuration (read-only, sourced from backend startup config).",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether mTLS is enabled on the backend.",
				Computed:    true,
			},
			"client_ca_cn": schema.StringAttribute{
				Description: "Common name of the client CA certificate.",
				Computed:    true,
			},
			"server_cert_subject": schema.StringAttribute{
				Description: "Subject of the server TLS certificate.",
				Computed:    true,
			},
		},
	}
}

func (d *MTLSConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MTLSConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	cfg, err := d.client.GetMTLSConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading mTLS Config", err.Error())
		return
	}

	model := MTLSConfigDataSourceModel{
		Enabled: types.BoolValue(cfg.Enabled),
	}
	if cfg.ClientCACN != nil {
		model.ClientCACN = types.StringValue(*cfg.ClientCACN)
	} else {
		model.ClientCACN = types.StringValue("")
	}
	if cfg.ServerCert != nil {
		model.ServerCert = types.StringValue(*cfg.ServerCert)
	} else {
		model.ServerCert = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
