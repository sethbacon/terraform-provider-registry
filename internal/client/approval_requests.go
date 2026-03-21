package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateApprovalRequest(ctx context.Context, req CreateApprovalRequestRequest) (*ApprovalRequest, error) {
	var ar ApprovalRequest
	if err := c.Post(ctx, "/api/v1/admin/approvals", req, &ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func (c *Client) GetApprovalRequest(ctx context.Context, id string) (*ApprovalRequest, error) {
	var ar ApprovalRequest
	if err := c.Get(ctx, "/api/v1/admin/approvals/"+id, &ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func (c *Client) DeleteApprovalRequest(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/admin/approvals/"+id)
}

func (c *Client) ListApprovalRequests(ctx context.Context) ([]ApprovalRequest, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/admin/approvals", "approval_requests")
	if err != nil {
		return nil, err
	}

	requests := make([]ApprovalRequest, 0, len(items))
	for _, raw := range items {
		var a ApprovalRequest
		if err := json.Unmarshal(raw, &a); err != nil {
			return nil, fmt.Errorf("unmarshaling approval request: %w", err)
		}
		requests = append(requests, a)
	}
	return requests, nil
}
