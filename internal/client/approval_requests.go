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

// There is deliberately no DeleteApprovalRequest. The backend serves no
// DELETE (or withdraw) route for approval requests: they are review records,
// reachable only through list, get, create, PUT .../review and POST .../token.
// A DELETE to /api/v1/admin/approvals/{id} matches no route and gets a 404,
// which Client.Delete reports as success, so a method here would make destroy
// look like it worked while the request stayed pending.
// registry_approval_request's Delete removes the resource from state instead.

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
