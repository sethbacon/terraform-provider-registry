package client

import "context"

// BackendVersion holds the response from GET /version.
type BackendVersion struct {
	Version         string   `json:"version"`
	BuildDate       string   `json:"build_date"`
	APIVersion      string   `json:"api_version"`
	DefaultLanguage string   `json:"default_language"`
	Protocols       []string `json:"protocols"`
	Capabilities    []string `json:"capabilities"`
}

// GetVersion calls GET /version and returns the backend version info.
func (c *Client) GetVersion(ctx context.Context) (*BackendVersion, error) {
	var v BackendVersion
	if err := c.Get(ctx, "/version", &v); err != nil {
		return nil, err
	}
	return &v, nil
}
