package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListTerraformMirrorVersions returns all versions for a given mirror config.
func (c *Client) ListTerraformMirrorVersions(ctx context.Context, mirrorID string) ([]TerraformMirrorVersion, error) {
	path := fmt.Sprintf("/api/v1/admin/terraform-mirrors/%s/versions", mirrorID)
	items, err := FetchAllPages(ctx, c, path, "versions")
	if err != nil {
		return nil, err
	}
	versions := make([]TerraformMirrorVersion, 0, len(items))
	for _, raw := range items {
		var v TerraformMirrorVersion
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("unmarshaling mirror version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// GetTerraformMirrorVersion returns a single version (including platforms).
func (c *Client) GetTerraformMirrorVersion(ctx context.Context, mirrorID, version string) (*TerraformMirrorVersion, error) {
	var v TerraformMirrorVersion
	path := fmt.Sprintf("/api/v1/admin/terraform-mirrors/%s/versions/%s", mirrorID, version)
	if err := c.Get(ctx, path, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ListTerraformMirrorHistory returns the sync history for a mirror.
func (c *Client) ListTerraformMirrorHistory(ctx context.Context, mirrorID string) ([]TerraformMirrorHistoryEntry, error) {
	path := fmt.Sprintf("/api/v1/admin/terraform-mirrors/%s/history", mirrorID)
	items, err := FetchAllPages(ctx, c, path, "history")
	if err != nil {
		return nil, err
	}
	entries := make([]TerraformMirrorHistoryEntry, 0, len(items))
	for _, raw := range items {
		var e TerraformMirrorHistoryEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("unmarshaling history entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
