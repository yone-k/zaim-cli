package zaim_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yone-k/zaim-cli/pkg/zaim"
)

func testOAuthConfig() zaim.OAuthConfig {
	return zaim.OAuthConfig{ConsumerKey: "consumer-key", ConsumerSecret: "consumer-secret", AccessToken: "access-token", AccessTokenSecret: "access-secret"}
}

func TestRequestPreservesRawJSONAndSignedParameters(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			body := ` { "money": { "id": 42, "amount": 10.25, "new_field": "kept" } } `
			var calls atomic.Int32
			params := map[string]string{"amount": "10.25", "comment": "メモ & coffee", "zero": "0", "empty": ""}
			if method == http.MethodDelete {
				params = nil
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != method || r.URL.Path != "/v2/home/money" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if !strings.Contains(r.Header.Get("Authorization"), `oauth_token="access-token"`) {
					t.Error("missing signed OAuth header")
				}
				if r.Header.Get("User-Agent") != "Zaim-CLI/1.0" {
					t.Error("SDK User-Agent changed")
				}
				switch method {
				case http.MethodGet:
					want := make(map[string]string, len(params)+1)
					for key, value := range params {
						want[key] = value
					}
					want["mapping"] = "1"
					if !reflect.DeepEqual(flatValues(r.URL.Query()), want) {
						t.Errorf("query = %v, want %v", r.URL.Query(), want)
					}
				case http.MethodPost, http.MethodPut:
					if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
						t.Error("missing form content type")
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(flatValues(r.PostForm), params) {
						t.Errorf("form = %v, want %v", r.PostForm, params)
					}
				case http.MethodDelete:
					data, _ := io.ReadAll(r.Body)
					if len(data) != 0 {
						t.Errorf("DELETE body = %q", data)
					}
				}
				_, _ = io.WriteString(w, body)
			}))
			defer server.Close()
			client := zaim.NewWithOptions(testOAuthConfig(), zaim.ClientOptions{BaseURL: server.URL, HTTPClient: server.Client()})
			got, err := client.Request(context.Background(), method, "/v2/home/money", params)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != body {
				t.Errorf("JSON = %s, want original %s", got, body)
			}
			if calls.Load() != 1 {
				t.Errorf("calls = %d, want 1", calls.Load())
			}
		})
	}
}

func flatValues(values map[string][]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}
	return result
}

func TestRequestReturnsHTTPErrorWithResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"message":"denied"}`)
	}))
	defer server.Close()
	client := zaim.NewWithOptions(testOAuthConfig(), zaim.ClientOptions{BaseURL: server.URL})
	_, err := client.Request(context.Background(), http.MethodGet, "/v2/home/money", nil)
	var httpErr *zaim.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error = %v, want HTTPError", err)
	}
	if httpErr.StatusCode != 401 || httpErr.Body != `{"message":"denied"}` {
		t.Errorf("HTTP error = %#v", httpErr)
	}
	if err.Error() != `request failed: status=401 body={"message":"denied"}` {
		t.Errorf("legacy error text changed: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (body *trackedBody) Close() error { body.closed = true; return nil }

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestRequestClosesResponseOnEveryOutcome(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		reader    io.Reader
		wantError bool
	}{
		{"success", 200, strings.NewReader(`{"money":[]}`), false},
		{"invalid JSON", 200, strings.NewReader(`{`), true},
		{"read failure", 200, failingReader{}, true},
		{"HTTP failure", 400, strings.NewReader(`{"error":"bad"}`), true},
		{"HTTP body read failure", 400, failingReader{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: tc.reader}
			httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			client := zaim.NewWithOptions(testOAuthConfig(), zaim.ClientOptions{HTTPClient: httpClient})
			_, err := client.Request(context.Background(), http.MethodGet, "/v2/home/money", nil)
			if (err != nil) != tc.wantError {
				t.Errorf("error = %v, wantError = %v", err, tc.wantError)
			}
			if !body.closed {
				t.Error("response body was not closed")
			}
			if tc.name == "invalid JSON" {
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Errorf("error = %v, want JSON syntax error", err)
				}
			}
			if strings.Contains(tc.name, "read failure") && !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Errorf("read error not preserved: %v", err)
			}
		})
	}
}

func TestRequestPropagatesCancellation(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	defer server.Close()
	client := zaim.NewWithOptions(testOAuthConfig(), zaim.ClientOptions{BaseURL: server.URL})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := client.Request(ctx, http.MethodGet, "/v2/home/money", nil); result <- err }()
	select {
	case <-started:
	case err := <-result:
		t.Fatalf("request exited before reaching server: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("request did not reach server")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not stop request")
	}
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("server request was not canceled")
	}
}

func TestRequestRejectsUnsupportedMethodsAndNonRelativePaths(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{http.MethodPatch, "/v2/home/money"},
		{http.MethodGet, "https://example.test/v2/home/money"},
		{http.MethodGet, "//example.test/v2/home/money"},
		{http.MethodGet, "v2/home/money"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			calls := 0
			client := zaim.NewWithOptions(testOAuthConfig(), zaim.ClientOptions{HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("unexpected HTTP request")
			})}})
			if _, err := client.Request(context.Background(), tc.method, tc.path, nil); err == nil {
				t.Error("expected validation error")
			}
			if calls != 0 {
				t.Errorf("invalid request reached transport %d times", calls)
			}
		})
	}
}
