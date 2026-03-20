package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*Organization, error) {
	var resp struct {
		Organization Organization `json:"organization"`
	}
	if err := c.Post(ctx, "/api/v1/organizations", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Organization, nil
}

func (c *Client) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	var resp struct {
		Organization Organization `json:"organization"`
	}
	if err := c.Get(ctx, "/api/v1/organizations/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Organization, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest) (*Organization, error) {
	var resp struct {
		Organization Organization `json:"organization"`
	}
	if err := c.Put(ctx, "/api/v1/organizations/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Organization, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/organizations/"+id)
}

func (c *Client) ListOrganizations(ctx context.Context, search string) ([]Organization, error) {
	path := "/api/v1/organizations"
	if search != "" {
		path += BuildQuery(map[string]string{"q": search})
	}

	items, err := FetchAllPages(ctx, c, path, "organizations")
	if err != nil {
		return nil, err
	}

	orgs := make([]Organization, 0, len(items))
	for _, raw := range items {
		var o Organization
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("unmarshaling organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	return orgs, nil
}

func (c *Client) AddOrganizationMember(ctx context.Context, orgID string, req AddMemberRequest) (*OrganizationMember, error) {
	var resp struct {
		Member OrganizationMember `json:"member"`
	}
	if err := c.Post(ctx, "/api/v1/organizations/"+orgID+"/members", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Member, nil
}

func (c *Client) GetOrganizationMember(ctx context.Context, orgID, userID string) (*OrganizationMember, error) {
	members, err := c.ListOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		if m.UserID == userID {
			return &m, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("member %s not found in org %s", userID, orgID)}
}

func (c *Client) UpdateOrganizationMember(ctx context.Context, orgID, userID string, req UpdateMemberRequest) (*OrganizationMember, error) {
	var resp struct {
		Member OrganizationMember `json:"member"`
	}
	if err := c.Put(ctx, "/api/v1/organizations/"+orgID+"/members/"+userID, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Member, nil
}

func (c *Client) RemoveOrganizationMember(ctx context.Context, orgID, userID string) error {
	return c.Delete(ctx, "/api/v1/organizations/"+orgID+"/members/"+userID)
}

func (c *Client) ListOrganizationMembers(ctx context.Context, orgID string) ([]OrganizationMember, error) {
	items, err := FetchAllPages(ctx, c, "/api/v1/organizations/"+orgID+"/members", "members")
	if err != nil {
		return nil, err
	}

	members := make([]OrganizationMember, 0, len(items))
	for _, raw := range items {
		var m OrganizationMember
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("unmarshaling member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}
