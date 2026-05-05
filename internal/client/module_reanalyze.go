package client

import (
	"context"
	"fmt"
)

// ReanalyzeModuleVersion triggers re-analysis of a module version via
// POST /api/v1/modules/{ns}/{name}/{sys}/versions/{version}/reanalyze.
func (c *Client) ReanalyzeModuleVersion(ctx context.Context, namespace, name, system, version string) error {
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s/versions/%s/reanalyze", namespace, name, system, version)
	resp, err := c.Do(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return parseResponseError(resp)
}
