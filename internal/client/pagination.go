package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Pagination holds metadata from paginated list responses.
type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// FetchAllPages fetches all pages from a paginated endpoint and returns
// the raw JSON items as []json.RawMessage. The caller is responsible for
// unmarshaling individual items.
//
// The endpoint should be the path without query params (e.g., "/api/v1/users").
// itemsKey is the JSON key holding the array (e.g., "users").
func FetchAllPages(ctx context.Context, c *Client, path, itemsKey string) ([]json.RawMessage, error) {
	var all []json.RawMessage
	page := 1
	perPage := defaultPageSize

	for {
		reqPath := fmt.Sprintf("%s?page=%d&per_page=%d", path, page, perPage)

		resp, err := c.Do(ctx, http.MethodGet, reqPath, nil)
		if err != nil {
			return nil, err
		}

		var raw map[string]json.RawMessage
		if err := decodeAndClose(resp, &raw); err != nil {
			return nil, err
		}

		items, ok := raw[itemsKey]
		if !ok {
			break
		}

		var batch []json.RawMessage
		if err := json.Unmarshal(items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshaling %q: %w", itemsKey, err)
		}
		all = append(all, batch...)

		// Check pagination
		paginationRaw, hasPagination := raw["pagination"]
		if !hasPagination || len(batch) == 0 {
			break
		}

		var pg Pagination
		if err := json.Unmarshal(paginationRaw, &pg); err != nil {
			break
		}

		if page*perPage >= pg.Total {
			break
		}
		page++
	}

	return all, nil
}

// decodeAndClose decodes the response body into v and closes the body.
func decodeAndClose(resp *http.Response, v interface{}) error {
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return parseResponseError(resp)
}
