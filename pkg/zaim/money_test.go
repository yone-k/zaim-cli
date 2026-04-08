package zaim

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestListMoney_shouldSendQueryParamsAndDecodeResponse_whenOptionsProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v2/home/money" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money")
		}

		query := r.URL.Query()
		assertQueryValue(t, query, "mapping", "1")
		assertQueryValue(t, query, "mode", "payment")
		assertQueryValue(t, query, "start_date", "2026-04-01")
		assertQueryValue(t, query, "end_date", "2026-04-30")
		assertQueryValue(t, query, "category_id", "10")
		assertQueryValue(t, query, "genre_id", "20")
		assertQueryValue(t, query, "account_id", "30")
		assertQueryValue(t, query, "limit", "100")
		assertQueryValue(t, query, "page", "2")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"money":[{"id":1,"mode":"payment","date":"2026-04-01","category_id":10,"genre_id":20,"amount":1200,"comment":"lunch","name":"","receipt_id":0,"place":""}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	opts := &ListMoneyOptions{
		Mode:       "payment",
		StartDate:  "2026-04-01",
		EndDate:    "2026-04-30",
		CategoryID: 10,
		GenreID:    20,
		AccountID:  30,
		Limit:      100,
		Page:       2,
	}

	got, err := client.ListMoney(context.Background(), opts)

	if err != nil {
		t.Fatalf("ListMoney() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(ListMoney()) = %d, want %d", len(got), 1)
	}
	if got[0].ID != 1 {
		t.Fatalf("money[0].ID = %d, want %d", got[0].ID, 1)
	}
	if got[0].Mode != "payment" {
		t.Fatalf("money[0].Mode = %q, want %q", got[0].Mode, "payment")
	}
	if got[0].Amount != 1200 {
		t.Fatalf("money[0].Amount = %d, want %d", got[0].Amount, 1200)
	}
}

func TestCreatePayment_shouldPostFormEncodedBody_whenRequiredAndOptionalFieldsProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v2/home/money/payment" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money/payment")
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %q, want %q", got, "application/x-www-form-urlencoded")
		}

		values := readFormBody(t, r)
		assertQueryValue(t, values, "amount", "1200")
		assertQueryValue(t, values, "date", "2026-04-01")
		assertQueryValue(t, values, "category_id", "10")
		assertQueryValue(t, values, "genre_id", "20")
		assertQueryValue(t, values, "from_account_id", "30")
		assertQueryValue(t, values, "place", "Cafe")
		assertQueryValue(t, values, "comment", "team lunch")
		assertQueryValue(t, values, "name", "Lunch")
		assertQueryValue(t, values, "mapping", "1")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	req := &CreatePaymentRequest{
		Amount:        1200,
		Date:          "2026-04-01",
		CategoryID:    10,
		GenreID:       20,
		FromAccountID: 30,
		Place:         "Cafe",
		Comment:       "team lunch",
		Name:          "Lunch",
	}

	if err := client.CreatePayment(context.Background(), req); err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}
}

func TestCreateIncome_shouldPostFormEncodedBody_whenOptionalFieldsProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v2/home/money/income" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money/income")
		}

		values := readFormBody(t, r)
		assertQueryValue(t, values, "amount", "500000")
		assertQueryValue(t, values, "date", "2026-04-25")
		assertQueryValue(t, values, "category_id", "99")
		assertQueryValue(t, values, "to_account_id", "77")
		assertQueryValue(t, values, "place", "Company")
		assertQueryValue(t, values, "comment", "salary")
		assertQueryValue(t, values, "mapping", "1")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	req := &CreateIncomeRequest{
		Amount:      500000,
		Date:        "2026-04-25",
		CategoryID:  99,
		ToAccountID: 77,
		Place:       "Company",
		Comment:     "salary",
	}

	if err := client.CreateIncome(context.Background(), req); err != nil {
		t.Fatalf("CreateIncome() error = %v", err)
	}
}

func TestCreateTransfer_shouldPostFormEncodedBody_whenRequiredFieldsProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v2/home/money/transfer" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money/transfer")
		}

		values := readFormBody(t, r)
		assertQueryValue(t, values, "amount", "30000")
		assertQueryValue(t, values, "date", "2026-04-10")
		assertQueryValue(t, values, "from_account_id", "10")
		assertQueryValue(t, values, "to_account_id", "20")
		assertQueryValue(t, values, "comment", "move savings")
		assertQueryValue(t, values, "mapping", "1")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	req := &CreateTransferRequest{
		Amount:        30000,
		Date:          "2026-04-10",
		FromAccountID: 10,
		ToAccountID:   20,
		Comment:       "move savings",
	}

	if err := client.CreateTransfer(context.Background(), req); err != nil {
		t.Fatalf("CreateTransfer() error = %v", err)
	}
}

func TestUpdateMoney_shouldPutOnlyChangedFields_whenOptionalPointersAreSet(t *testing.T) {
	t.Parallel()

	amount := 1500
	comment := "updated comment"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPut)
		}
		if r.URL.Path != "/v2/home/money/payment/42" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money/payment/42")
		}

		values := readFormBody(t, r)
		assertQueryValue(t, values, "amount", "1500")
		assertQueryValue(t, values, "comment", "updated comment")

		if got := values.Get("date"); got != "" {
			t.Fatalf("date = %q, want empty", got)
		}
		if got := values.Get("category_id"); got != "" {
			t.Fatalf("category_id = %q, want empty", got)
		}
		if got := values.Get("mapping"); got != "" {
			t.Fatalf("mapping = %q, want empty", got)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	req := &UpdateMoneyRequest{
		Amount:  &amount,
		Comment: &comment,
	}

	if err := client.UpdateMoney(context.Background(), 42, "payment", req); err != nil {
		t.Fatalf("UpdateMoney() error = %v", err)
	}
}

func TestUpdateMoney_shouldReturnError_whenModeIsInvalid(t *testing.T) {
	t.Parallel()

	client := &Client{}

	err := client.UpdateMoney(context.Background(), 42, "invalid", nil)
	if err == nil {
		t.Fatal("UpdateMoney() error = nil, want error")
	}

	want := "invalid mode: invalid (must be payment, income, or transfer)"
	if err.Error() != want {
		t.Fatalf("UpdateMoney() error = %q, want %q", err.Error(), want)
	}
}

func TestDeleteMoney_shouldDeleteResource_whenIDAndModeProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodDelete)
		}
		if r.URL.Path != "/v2/home/money/income/99" {
			t.Fatalf("request path = %s, want %s", r.URL.Path, "/v2/home/money/income/99")
		}
		if r.ContentLength > 0 {
			t.Fatalf("ContentLength = %d, want 0", r.ContentLength)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client()}

	if err := client.DeleteMoney(context.Background(), 99, "income"); err != nil {
		t.Fatalf("DeleteMoney() error = %v", err)
	}
}

func TestDeleteMoney_shouldReturnError_whenModeIsInvalid(t *testing.T) {
	t.Parallel()

	client := &Client{}

	err := client.DeleteMoney(context.Background(), 99, "invalid")
	if err == nil {
		t.Fatal("DeleteMoney() error = nil, want error")
	}

	want := "invalid mode: invalid (must be payment, income, or transfer)"
	if err.Error() != want {
		t.Fatalf("DeleteMoney() error = %q, want %q", err.Error(), want)
	}
}

func readFormBody(t *testing.T, r *http.Request) url.Values {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("ReadAll(r.Body) error = %v", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("ParseQuery(body) error = %v", err)
	}

	return values
}

func assertQueryValue(t *testing.T, values url.Values, key string, want string) {
	t.Helper()

	if got := values.Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
