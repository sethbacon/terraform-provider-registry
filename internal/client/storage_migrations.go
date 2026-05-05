package client

import (
	"context"
	"fmt"
)

// CreateStorageMigration starts a new migration via POST /api/v1/admin/storage/migrations.
func (c *Client) CreateStorageMigration(ctx context.Context, req CreateStorageMigrationRequest) (*StorageMigration, error) {
	var m StorageMigration
	if err := c.Post(ctx, "/api/v1/admin/storage/migrations", req, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// GetStorageMigration fetches migration status via GET /api/v1/admin/storage/migrations/{id}.
func (c *Client) GetStorageMigration(ctx context.Context, id string) (*StorageMigration, error) {
	var m StorageMigration
	if err := c.Get(ctx, "/api/v1/admin/storage/migrations/"+id, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// CancelStorageMigration cancels an in-flight migration via POST /api/v1/admin/storage/migrations/{id}/cancel.
func (c *Client) CancelStorageMigration(ctx context.Context, id string) (*StorageMigration, error) {
	var m StorageMigration
	if err := c.Post(ctx, fmt.Sprintf("/api/v1/admin/storage/migrations/%s/cancel", id), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
