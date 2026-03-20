package client

import "context"

func (c *Client) CreateModuleSCMLink(ctx context.Context, moduleID string, req CreateModuleSCMLinkRequest) (*ModuleSCMLink, error) {
	var resp struct {
		Link ModuleSCMLink `json:"link"`
	}
	if err := c.Post(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Link, nil
}

func (c *Client) GetModuleSCMLink(ctx context.Context, moduleID string) (*ModuleSCMLink, error) {
	var resp struct {
		Link ModuleSCMLink `json:"link"`
	}
	if err := c.Get(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", &resp); err != nil {
		return nil, err
	}
	return &resp.Link, nil
}

func (c *Client) UpdateModuleSCMLink(ctx context.Context, moduleID string, req UpdateModuleSCMLinkRequest) (*ModuleSCMLink, error) {
	var resp struct {
		Link ModuleSCMLink `json:"link"`
	}
	if err := c.Put(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Link, nil
}

func (c *Client) DeleteModuleSCMLink(ctx context.Context, moduleID string) error {
	return c.Delete(ctx, "/api/v1/admin/modules/"+moduleID+"/scm")
}
