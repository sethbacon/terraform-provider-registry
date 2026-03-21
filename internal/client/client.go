// Package client provides an HTTP client for the Terraform Registry Backend API.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
	defaultPageSize   = 100

	initialBackoff = 100 * time.Millisecond
	maxBackoff     = 30 * time.Second
	backoffFactor  = 2.0
)

// Client is the API client for the Terraform Registry Backend.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	maxRetries int
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			_ = transport
		}
		c.httpClient.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retries for failed requests.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithInsecure disables TLS certificate verification.
func WithInsecure(insecure bool) Option {
	return func(c *Client) {
		if insecure {
			transport := c.httpClient.Transport.(*http.Transport)
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		}
	}
}

// NewClient creates a new API client.
// endpoint is the base URL of the registry (e.g., "https://registry.example.com").
// token is a JWT or API key used as the Bearer token.
func NewClient(endpoint, token string, opts ...Option) (*Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("registry endpoint must not be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("registry token must not be empty")
	}

	// Normalize: strip trailing slash, ensure scheme
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return nil, fmt.Errorf("registry endpoint must start with http:// or https://")
	}

	c := &Client{
		baseURL: endpoint,
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		},
		maxRetries: defaultMaxRetries,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Do executes an HTTP request with retry logic for 429 and 5xx responses.
// It sets the Authorization header automatically.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	reqURL := c.baseURL + path

	var attempt int
	backoff := initialBackoff

	for {
		var bodyReader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshaling request body: %w", err)
			}
			bodyReader = bytes.NewReader(b)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		tflog.Debug(ctx, "registry API request", map[string]interface{}{
			"method": method,
			"url":    reqURL,
		})

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			_ = resp.Body.Close()

			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("request failed after %d retries: HTTP %d", attempt, resp.StatusCode)
			}

			// Respect Retry-After header if present
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, parseErr := strconv.Atoi(ra); parseErr == nil {
					backoff = time.Duration(secs) * time.Second
				}
			}

			tflog.Debug(ctx, "retrying request", map[string]interface{}{
				"attempt": attempt + 1,
				"backoff": backoff.String(),
				"status":  resp.StatusCode,
			})

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * backoffFactor)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			attempt++
			continue
		}

		return resp, nil
	}
}

// Get performs a GET request and decodes the JSON response body into result.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return parseResponse(resp, result)
}

// Post performs a POST request with body and decodes the JSON response into result.
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return parseResponse(resp, result)
}

// Put performs a PUT request with body and decodes the JSON response into result.
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.Do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return parseResponse(resp, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil // treat 404 on delete as success
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return parseResponseError(resp)
}

// parseResponse decodes a successful response or returns an APIError.
func parseResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result == nil || resp.StatusCode == http.StatusNoContent {
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		return nil
	}
	return parseResponseError(resp)
}

// parseResponseError reads an error response body and returns an APIError.
func parseResponseError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	apiErr := &APIError{StatusCode: resp.StatusCode}

	var raw struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}
	if err := json.Unmarshal(b, &raw); err == nil && raw.Error != "" {
		apiErr.Message = raw.Error
		apiErr.Fields = raw.Fields
	} else {
		apiErr.Message = string(b)
	}
	return apiErr
}

// BuildQuery constructs a query string from a map of parameters, omitting empty values.
func BuildQuery(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}
