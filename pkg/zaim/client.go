package zaim

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://api.zaim.net"

type Client struct {
	oauthConfig OAuthConfig
	baseURL     string
	httpClient  *http.Client
}

func New(config OAuthConfig) *Client {
	return &Client{
		oauthConfig: config,
		baseURL:     defaultBaseURL,
		httpClient:  http.DefaultClient,
	}
}

func (c *Client) do(ctx context.Context, method string, path string, params map[string]string) (*http.Response, error) {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	endpoint, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("build request url: %w", err)
	}

	signatureParams, err := newOAuthParams(c.oauthConfig.ConsumerKey)
	if err != nil {
		return nil, fmt.Errorf("build oauth params: %w", err)
	}
	signatureParams["oauth_token"] = c.oauthConfig.AccessToken

	var body io.Reader

	switch method {
	case http.MethodGet:
		query := endpoint.Query()
		query.Set("mapping", "1")
		for key, value := range params {
			query.Set(key, value)
		}
		endpoint.RawQuery = query.Encode()

		for key, values := range query {
			if len(values) == 0 {
				signatureParams[key] = ""
				continue
			}
			signatureParams[key] = values[0]
		}
	case http.MethodPost, http.MethodPut:
		form := url.Values{}
		for key, value := range params {
			form.Set(key, value)
			signatureParams[key] = value
		}
		body = strings.NewReader(form.Encode())
	default:
		for key, value := range params {
			signatureParams[key] = value
		}
	}

	signatureURL := endpoint.Scheme + "://" + endpoint.Host + endpoint.Path
	signatureParams["oauth_signature"] = GenerateSignature(
		method,
		signatureURL,
		signatureParams,
		c.oauthConfig.ConsumerSecret,
		c.oauthConfig.AccessTokenSecret,
	)

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Authorization", buildAuthorizationHeader(signatureParams))
	req.Header.Set("User-Agent", "Zaim-CLI/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("read error response: %w", readErr)
		}

		return nil, fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return resp, nil
}
