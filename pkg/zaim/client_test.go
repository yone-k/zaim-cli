package zaim

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNew_shouldReturnNonNilClient_whenOAuthConfigIsProvided(t *testing.T) {
	t.Parallel()

	// Given: a minimal OAuth configuration.
	config := OAuthConfig{
		ConsumerKey:       "consumer-key",
		ConsumerSecret:    "consumer-secret",
		AccessToken:       "access-token",
		AccessTokenSecret: "access-token-secret",
	}

	// When: a new SDK client is created.
	client := New(config)

	// Then: a non-nil client is returned.
	if client == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDoGet_shouldAddOAuthHeaderUserAgentAndMappingQuery_whenSendingGetRequest(t *testing.T) {
	t.Parallel()

	// Given: a test server that validates the outbound GET request.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "OAuth ") {
			t.Fatalf("Authorization = %q, want prefix %q", got, "OAuth ")
		}
		if got := r.Header.Get("User-Agent"); got != "Zaim-CLI/1.0" {
			t.Fatalf("User-Agent = %q, want %q", got, "Zaim-CLI/1.0")
		}
		if got := r.URL.Query().Get("mapping"); got != "1" {
			t.Fatalf("mapping = %q, want %q", got, "1")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}
	params := map[string]string{
		"mode": "payment",
	}

	// When: the GET request is executed.
	resp, err := client.do(context.Background(), http.MethodGet, "/money", params)

	// Then: the request succeeds without client-side error.
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
}

func TestDoPost_shouldSendFormEncodedBody_whenSendingPostRequest(t *testing.T) {
	t.Parallel()

	// Given: a test server that validates the outbound POST request.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %q, want %q", got, "application/x-www-form-urlencoded")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll(r.Body) error = %v", err)
		}

		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery(body) error = %v", err)
		}
		if got := values.Get("amount"); got != "1200" {
			t.Fatalf("amount = %q, want %q", got, "1200")
		}
		if got := values.Get("comment"); got != "lunch expense" {
			t.Fatalf("comment = %q, want %q", got, "lunch expense")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}
	params := map[string]string{
		"amount":  "1200",
		"comment": "lunch expense",
	}

	// When: the POST request is executed.
	resp, err := client.do(context.Background(), http.MethodPost, "/money/payment", params)

	// Then: the request succeeds without client-side error.
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
}

func TestDoErrorHandling_shouldReturnError_whenResponseStatusIsBadRequestOrHigher(t *testing.T) {
	t.Parallel()

	// Given: a test server that returns a client error response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	// When: the request receives a 4xx response.
	resp, err := client.do(context.Background(), http.MethodGet, "/money", nil)

	// Then: an error is returned.
	if err == nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatal("do() error = nil, want non-nil")
	}
	if resp != nil {
		resp.Body.Close()
	}
}
