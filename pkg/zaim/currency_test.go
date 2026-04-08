package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCurrencies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/currency" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/currency")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"currencies":[{"currency_code":"JPY","name":"Japanese Yen","unit":"円","point":0}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	currencies, err := client.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies() error = %v", err)
	}
	if len(currencies) != 1 {
		t.Fatalf("len(currencies) = %d, want %d", len(currencies), 1)
	}
	if currencies[0].CurrencyCode != "JPY" {
		t.Fatalf("currencies[0].CurrencyCode = %q, want %q", currencies[0].CurrencyCode, "JPY")
	}
	if currencies[0].Name != "Japanese Yen" {
		t.Fatalf("currencies[0].Name = %q, want %q", currencies[0].Name, "Japanese Yen")
	}
}

func TestListCurrencies_shouldReturnError_whenServerReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/currency" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/currency")
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL}

	currencies, err := client.ListCurrencies(context.Background())
	if err == nil {
		t.Fatal("ListCurrencies() error = nil, want non-nil")
	}
	if currencies != nil {
		t.Fatalf("currencies = %#v, want nil", currencies)
	}
}
