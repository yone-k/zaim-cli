package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUserCategories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/category" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/category")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"categories":[{"id":1,"name":"Food","mode":"payment","sort":1,"active":1,"created":"2024-01-01","modified":"2024-01-02"}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	categories, err := client.ListUserCategories(context.Background())
	if err != nil {
		t.Fatalf("ListUserCategories() error = %v", err)
	}
	if len(categories) != 1 {
		t.Fatalf("len(categories) = %d, want %d", len(categories), 1)
	}
	if categories[0].ID != 1 {
		t.Fatalf("categories[0].ID = %d, want %d", categories[0].ID, 1)
	}
	if categories[0].Name != "Food" {
		t.Fatalf("categories[0].Name = %q, want %q", categories[0].Name, "Food")
	}
}

func TestListUserCategories_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/category" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/category")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	categories, err := client.ListUserCategories(context.Background())
	if err == nil {
		t.Fatal("ListUserCategories() error = nil, want non-nil")
	}
	if categories != nil {
		t.Fatalf("categories = %#v, want nil", categories)
	}
}

func TestListDefaultCategories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if got := r.URL.RequestURI(); got != "/v2/category?mapping=1&mode=payment" {
			t.Fatalf("request URI = %s, want %s", got, "/v2/category?mapping=1&mode=payment")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"categories":[{"id":2,"name":"Utilities","mode":"payment","sort":2,"active":1,"created":"2024-02-01","modified":"2024-02-02"}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	categories, err := client.ListDefaultCategories(context.Background(), "payment")
	if err != nil {
		t.Fatalf("ListDefaultCategories() error = %v", err)
	}
	if len(categories) != 1 {
		t.Fatalf("len(categories) = %d, want %d", len(categories), 1)
	}
	if categories[0].ID != 2 {
		t.Fatalf("categories[0].ID = %d, want %d", categories[0].ID, 2)
	}
	if categories[0].Name != "Utilities" {
		t.Fatalf("categories[0].Name = %q, want %q", categories[0].Name, "Utilities")
	}
}

func TestListDefaultCategories_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if got := r.URL.RequestURI(); got != "/v2/category?mapping=1&mode=payment" {
			t.Fatalf("request URI = %s, want %s", got, "/v2/category?mapping=1&mode=payment")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	categories, err := client.ListDefaultCategories(context.Background(), "payment")
	if err == nil {
		t.Fatal("ListDefaultCategories() error = nil, want non-nil")
	}
	if categories != nil {
		t.Fatalf("categories = %#v, want nil", categories)
	}
}
