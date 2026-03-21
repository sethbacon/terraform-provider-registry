package client

import "context"

func (c *Client) CreateModuleSCMLink(ctx context.Context, moduleID string, req CreateModuleSCMLinkRequest) (*ModuleSCMLink, error) {
	var created struct {
		LinkID string `json:"link_id"`
	}
	if err := c.Post(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", req, &created); err != nil {
		return nil, err
	}
	// Backend returns only link_id on create; fetch the full record
	return c.GetModuleSCMLink(ctx, moduleID)
}

func (c *Client) GetModuleSCMLink(ctx context.Context, moduleID string) (*ModuleSCMLink, error) {
	var link ModuleSCMLink
	if err := c.Get(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (c *Client) UpdateModuleSCMLink(ctx context.Context, moduleID string, req UpdateModuleSCMLinkRequest) (*ModuleSCMLink, error) {
	// Backend returns only {"message": "..."} on update; fetch the updated record after
	if err := c.Put(ctx, "/api/v1/admin/modules/"+moduleID+"/scm", req, nil); err != nil {
		return nil, err
	}
	return c.GetModuleSCMLink(ctx, moduleID)
}

func (c *Client) DeleteModuleSCMLink(ctx context.Context, moduleID string) error {
	return c.Delete(ctx, "/api/v1/admin/modules/"+moduleID+"/scm")
}
