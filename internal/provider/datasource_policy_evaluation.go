package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

var _ datasource.DataSource = &PolicyEvaluationDataSource{}

type PolicyEvaluationDataSource struct {
	client *client.Client
}

type PolicyEvaluationDataSourceModel struct {
	InputJSON  types.String `tfsdk:"input_json"`
	Query      types.String `tfsdk:"query"`
	Allowed    types.Bool   `tfsdk:"allowed"`
	Reason     types.String `tfsdk:"reason"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func NewPolicyEvaluationDataSource() datasource.DataSource {
	return &PolicyEvaluationDataSource{}
}

func (d *PolicyEvaluationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_evaluation"
}

func (d *PolicyEvaluationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Evaluates input against the rego policy bundle at plan time. Useful for policy-gating resource creation.",
		Attributes: map[string]schema.Attribute{
			"input_json": schema.StringAttribute{
				Description: "JSON-encoded input document to evaluate against the policy bundle.",
				Required:    true,
			},
			"query": schema.StringAttribute{
				Description: "Optional rego query path (e.g. 'data.authz.allow'). Defaults to the bundle entry point if omitted.",
				Optional:    true,
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Whether the policy evaluation returned an allow decision.",
				Computed:    true,
			},
			"reason": schema.StringAttribute{
				Description: "Human-readable reason returned by the policy (if provided).",
				Computed:    true,
			},
			"result_json": schema.StringAttribute{
				Description: "Full policy result serialised as JSON.",
				Computed:    true,
			},
		},
	}
}

func (d *PolicyEvaluationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PolicyEvaluationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PolicyEvaluationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(state.InputJSON.ValueString()), &input); err != nil {
		resp.Diagnostics.AddError("Invalid input_json", "input_json must be a valid JSON object: "+err.Error())
		return
	}

	evalReq := client.PolicyEvaluateRequest{Input: input}
	if !state.Query.IsNull() && !state.Query.IsUnknown() {
		evalReq.Query = state.Query.ValueString()
	}

	result, err := d.client.EvaluatePolicy(ctx, evalReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Evaluating Policy", err.Error())
		return
	}

	state.Allowed = types.BoolValue(result.Allowed)
	if result.Reason != nil {
		state.Reason = types.StringValue(*result.Reason)
	} else {
		state.Reason = types.StringValue("")
	}
	if state.Query.IsNull() || state.Query.IsUnknown() {
		state.Query = types.StringValue("")
	}

	resultJSON := "{}"
	if result.Result != nil {
		if b, err := json.Marshal(result.Result); err == nil {
			resultJSON = string(b)
		}
	}
	state.ResultJSON = types.StringValue(resultJSON)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
