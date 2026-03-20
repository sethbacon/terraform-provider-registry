package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) ListAuditLogs(ctx context.Context, action, resourceType string, limit, offset int) ([]AuditLog, int, error) {
	params := map[string]string{}
	if action != "" {
		params["action"] = action
	}
	if resourceType != "" {
		params["resource_type"] = resourceType
	}
	if limit > 0 {
		params["per_page"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["page"] = fmt.Sprintf("%d", offset/limit+1)
	}

	path := "/api/v1/admin/audit-logs"
	if len(params) > 0 {
		path += BuildQuery(params)
	}

	var raw struct {
		AuditLogs  []json.RawMessage `json:"audit_logs"`
		Pagination Pagination        `json:"pagination"`
	}
	if err := c.Get(ctx, path, &raw); err != nil {
		return nil, 0, err
	}

	logs := make([]AuditLog, 0, len(raw.AuditLogs))
	for _, r := range raw.AuditLogs {
		var l AuditLog
		if err := json.Unmarshal(r, &l); err != nil {
			return nil, 0, fmt.Errorf("unmarshaling audit log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, raw.Pagination.Total, nil
}

func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	var resp struct {
		Stats Stats `json:"stats"`
	}
	if err := c.Get(ctx, "/api/v1/admin/stats/dashboard", &resp); err != nil {
		return nil, err
	}
	return &resp.Stats, nil
}
