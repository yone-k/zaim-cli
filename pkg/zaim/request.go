package zaim

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClientOptions configures the API endpoint and HTTP transport.
type ClientOptions struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewWithOptions constructs a client with explicit endpoint and transport options.
func NewWithOptions(config OAuthConfig, options ClientOptions) *Client {
	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{oauthConfig: config, baseURL: baseURL, httpClient: httpClient}
}

// HTTPError preserves the status and body returned by the API.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("request failed: status=%d body=%s", err.StatusCode, strings.TrimSpace(err.Body))
}

// Request returns the original JSON response from a signed API request.
func (c *Client) Request(ctx context.Context, method, path string, params map[string]string) (json.RawMessage, error) {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
	parsed, err := url.ParseRequestURI(path)
	if err != nil || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || parsed.IsAbs() {
		return nil, fmt.Errorf("API path must be relative: %q", path)
	}
	response, err := c.do(ctx, method, path, params)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if !json.Valid(body) {
		var value any
		return nil, fmt.Errorf("decode response: %w", json.Unmarshal(body, &value))
	}
	return json.RawMessage(body), nil
}
