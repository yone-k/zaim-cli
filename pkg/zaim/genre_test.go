package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUserGenres(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/genre" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/genre")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"genres":[{"id":10,"name":"Groceries","category_id":1,"mode":"payment","sort":1,"active":1,"created":"2024-03-01","modified":"2024-03-02"}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	genres, err := client.ListUserGenres(context.Background())
	if err != nil {
		t.Fatalf("ListUserGenres() error = %v", err)
	}
	if len(genres) != 1 {
		t.Fatalf("len(genres) = %d, want %d", len(genres), 1)
	}
	if genres[0].ID != 10 {
		t.Fatalf("genres[0].ID = %d, want %d", genres[0].ID, 10)
	}
	if genres[0].Name != "Groceries" {
		t.Fatalf("genres[0].Name = %q, want %q", genres[0].Name, "Groceries")
	}
}

func TestListUserGenres_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/genre" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/genre")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	genres, err := client.ListUserGenres(context.Background())
	if err == nil {
		t.Fatal("ListUserGenres() error = nil, want non-nil")
	}
	if genres != nil {
		t.Fatalf("genres = %#v, want nil", genres)
	}
}

func TestListDefaultGenres(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if got := r.URL.RequestURI(); got != "/v2/genre?mapping=1&mode=payment" {
			t.Fatalf("request URI = %s, want %s", got, "/v2/genre?mapping=1&mode=payment")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"genres":[{"id":11,"name":"Transport","category_id":2,"mode":"payment","sort":2,"active":1,"created":"2024-04-01","modified":"2024-04-02"}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	genres, err := client.ListDefaultGenres(context.Background(), "payment")
	if err != nil {
		t.Fatalf("ListDefaultGenres() error = %v", err)
	}
	if len(genres) != 1 {
		t.Fatalf("len(genres) = %d, want %d", len(genres), 1)
	}
	if genres[0].ID != 11 {
		t.Fatalf("genres[0].ID = %d, want %d", genres[0].ID, 11)
	}
	if genres[0].Name != "Transport" {
		t.Fatalf("genres[0].Name = %q, want %q", genres[0].Name, "Transport")
	}
}

func TestListDefaultGenres_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if got := r.URL.RequestURI(); got != "/v2/genre?mapping=1&mode=payment" {
			t.Fatalf("request URI = %s, want %s", got, "/v2/genre?mapping=1&mode=payment")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	genres, err := client.ListDefaultGenres(context.Background(), "payment")
	if err == nil {
		t.Fatal("ListDefaultGenres() error = nil, want non-nil")
	}
	if genres != nil {
		t.Fatalf("genres = %#v, want nil", genres)
	}
}
