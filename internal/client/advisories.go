package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListAdvisories fetches advisories from the admin endpoint.
// If activeOnly is true it uses the public /active endpoint.
func (c *Client) ListAdvisories(ctx context.Context, activeOnly bool, severities []string) ([]Advisory, error) {
	var path string
	if activeOnly {
		path = "/api/v1/advisories/active"
	} else {
		path = "/api/v1/admin/advisories"
	}

	params := map[string]string{}
	if len(severities) > 0 {
		for _, s := range severities {
			params["severity"] = s // last one wins; backend may accept repeated params
		}
	}
	if len(params) > 0 {
		path += BuildQuery(params)
	}

	items, err := FetchAllPages(ctx, c, path, "advisories")
	if err != nil {
		// Fall back: try as a flat array
		var flat []Advisory
		if err2 := c.Get(ctx, path, &flat); err2 == nil {
			return flat, nil
		}
		return nil, err
	}
	advisories := make([]Advisory, 0, len(items))
	for _, raw := range items {
		var a Advisory
		if err := json.Unmarshal(raw, &a); err != nil {
			return nil, fmt.Errorf("unmarshaling advisory: %w", err)
		}
		advisories = append(advisories, a)
	}
	return advisories, nil
}
