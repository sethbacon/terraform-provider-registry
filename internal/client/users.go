package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.Post(ctx, "/api/v1/users", req, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.Get(ctx, "/api/v1/users/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.Put(ctx, "/api/v1/users/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/users/"+id)
}

func (c *Client) ListUsers(ctx context.Context, search string) ([]User, error) {
	path := "/api/v1/users"
	if search != "" {
		path += BuildQuery(map[string]string{"q": search})
	}

	items, err := FetchAllPages(ctx, c, path, "users")
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(items))
	for _, raw := range items {
		var u User
		if err := json.Unmarshal(raw, &u); err != nil {
			return nil, fmt.Errorf("unmarshaling user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}
