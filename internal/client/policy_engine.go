package client

import "context"

// GetPolicyEngineConfig fetches the rego bundle config via GET /api/v1/admin/policy/config.
func (c *Client) GetPolicyEngineConfig(ctx context.Context) (*PolicyEngineConfig, error) {
	var cfg PolicyEngineConfig
	if err := c.Get(ctx, "/api/v1/admin/policy/config", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// EvaluatePolicy evaluates input against the rego bundle via POST /api/v1/admin/policy/evaluate.
func (c *Client) EvaluatePolicy(ctx context.Context, req PolicyEvaluateRequest) (*PolicyEvaluateResult, error) {
	var result PolicyEvaluateResult
	if err := c.Post(ctx, "/api/v1/admin/policy/evaluate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
