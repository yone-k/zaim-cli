package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyAuth_Success(t *testing.T) {
	t.Parallel()

	// Given: a verify-auth endpoint that validates the outbound request.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/user/verify" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/user/verify")
		}
		if got := r.URL.Query().Get("mapping"); got != "1" {
			t.Fatalf("mapping = %q, want %q", got, "1")
		}
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "OAuth ") {
			t.Fatalf("Authorization = %q, want prefix %q", got, "OAuth ")
		}
		if got := r.Header.Get("User-Agent"); got != "Zaim-CLI/1.0" {
			t.Fatalf("User-Agent = %q, want %q", got, "Zaim-CLI/1.0")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"me":{"id":1,"login":"test","name":"Test User","profile_image_url":"https://example.com/avatar.png","input_count":10,"repeat_count":2,"day":"2026-04-08"}}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	// When: the authenticated user is requested.
	user, err := client.VerifyAuth(context.Background())

	// Then: the me payload is parsed into User.
	if err != nil {
		t.Fatalf("VerifyAuth() error = %v", err)
	}
	if user == nil {
		t.Fatal("VerifyAuth() user = nil, want non-nil")
	}

	want := &User{
		ID:              1,
		Login:           "test",
		Name:            "Test User",
		ProfileImageURL: "https://example.com/avatar.png",
		InputCount:      10,
		RepeatCount:     2,
		Day:             "2026-04-08",
	}

	if *user != *want {
		t.Fatalf("VerifyAuth() user = %+v, want %+v", *user, *want)
	}
}

func TestVerifyAuth_Error(t *testing.T) {
	t.Parallel()

	// Given: a verify-auth endpoint that returns a server error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/user/verify" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/user/verify")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	// When: the authenticated user request receives a server error.
	user, err := client.VerifyAuth(context.Background())

	// Then: an error is returned.
	if err == nil {
		t.Fatal("VerifyAuth() error = nil, want non-nil")
	}
	if user != nil {
		t.Fatalf("VerifyAuth() user = %+v, want nil", *user)
	}
}
