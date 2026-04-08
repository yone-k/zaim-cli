package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUserAccounts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/account" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/account")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accounts":[{"id":100,"name":"Wallet","modified":"2024-05-02","sort":1,"active":1,"local_id":0,"website_id":0,"parent_account_id":0}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	accounts, err := client.ListUserAccounts(context.Background())
	if err != nil {
		t.Fatalf("ListUserAccounts() error = %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("len(accounts) = %d, want %d", len(accounts), 1)
	}
	if accounts[0].ID != 100 {
		t.Fatalf("accounts[0].ID = %d, want %d", accounts[0].ID, 100)
	}
	if accounts[0].Name != "Wallet" {
		t.Fatalf("accounts[0].Name = %q, want %q", accounts[0].Name, "Wallet")
	}
}

func TestListUserAccounts_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/account" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/account")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	accounts, err := client.ListUserAccounts(context.Background())
	if err == nil {
		t.Fatal("ListUserAccounts() error = nil, want non-nil")
	}
	if accounts != nil {
		t.Fatalf("accounts = %#v, want nil", accounts)
	}
}
