package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateModule(ctx context.Context, req CreateModuleRequest) (*Module, error) {
	var mod Module
	if err := c.Post(ctx, "/api/v1/admin/modules/create", req, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func (c *Client) GetModule(ctx context.Context, namespace, name, system string) (*Module, error) {
	var mod Module
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s", namespace, name, system)
	if err := c.Get(ctx, path, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func (c *Client) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	var mod Module
	if err := c.Get(ctx, "/api/v1/admin/modules/"+id, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func (c *Client) UpdateModule(ctx context.Context, id string, req UpdateModuleRequest) (*Module, error) {
	var mod Module
	if err := c.Put(ctx, "/api/v1/admin/modules/"+id, req, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func (c *Client) DeleteModule(ctx context.Context, namespace, name, system string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/v1/modules/%s/%s/%s", namespace, name, system))
}

func (c *Client) ListModules(ctx context.Context, namespace, search string) ([]Module, error) {
	path := "/api/v1/modules/search"
	params := map[string]string{}
	if namespace != "" {
		params["namespace"] = namespace
	}
	if search != "" {
		params["q"] = search
	}
	if len(params) > 0 {
		path += BuildQuery(params)
	}

	items, err := FetchAllPages(ctx, c, path, "modules")
	if err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(items))
	for _, raw := range items {
		var m Module
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("unmarshaling module: %w", err)
		}
		modules = append(modules, m)
	}
	return modules, nil
}
