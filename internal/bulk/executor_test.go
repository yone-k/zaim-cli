package bulk

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yone-k/zaim-cli/pkg/zaim"
)

type mockClient struct {
	createPaymentFn  func(ctx context.Context, req *zaim.CreatePaymentRequest) error
	createIncomeFn   func(ctx context.Context, req *zaim.CreateIncomeRequest) error
	createTransferFn func(ctx context.Context, req *zaim.CreateTransferRequest) error
	updateMoneyFn    func(ctx context.Context, id int, mode string, req *zaim.UpdateMoneyRequest) error
}

func (m *mockClient) CreatePayment(ctx context.Context, req *zaim.CreatePaymentRequest) error {
	if m.createPaymentFn == nil {
		return nil
	}
	return m.createPaymentFn(ctx, req)
}

func (m *mockClient) CreateIncome(ctx context.Context, req *zaim.CreateIncomeRequest) error {
	if m.createIncomeFn == nil {
		return nil
	}
	return m.createIncomeFn(ctx, req)
}

func (m *mockClient) CreateTransfer(ctx context.Context, req *zaim.CreateTransferRequest) error {
	if m.createTransferFn == nil {
		return nil
	}
	return m.createTransferFn(ctx, req)
}

func (m *mockClient) UpdateMoney(ctx context.Context, id int, mode string, req *zaim.UpdateMoneyRequest) error {
	if m.updateMoneyFn == nil {
		return nil
	}
	return m.updateMoneyFn(ctx, id, mode, req)
}

func TestExecuteCreatePaymentsAllSuccess(t *testing.T) {
	ctx := context.Background()
	rows := paymentRows()
	calls := 0
	client := &mockClient{
		createPaymentFn: func(ctx context.Context, req *zaim.CreatePaymentRequest) error {
			calls++
			return nil
		},
	}

	got := ExecuteCreatePayments(ctx, client, rows, false)

	want := BulkResult{Total: 2, Succeeded: 2, Failed: 0}
	assertBulkResult(t, got, want)
	if calls != 2 {
		t.Fatalf("CreatePayment calls = %d, want 2", calls)
	}
}

func TestExecuteCreatePaymentsPartialSuccess(t *testing.T) {
	ctx := context.Background()
	rows := paymentRows()
	calls := 0
	client := &mockClient{
		createPaymentFn: func(ctx context.Context, req *zaim.CreatePaymentRequest) error {
			calls++
			if calls == 2 {
				return errors.New("api unavailable")
			}
			return nil
		},
	}

	got := ExecuteCreatePayments(ctx, client, rows, false)

	want := BulkResult{
		Total:     2,
		Succeeded: 1,
		Failed:    1,
		Errors: []RowError{
			{Row: 2, Error: "api unavailable"},
		},
	}
	assertBulkResult(t, got, want)
	if calls != 2 {
		t.Fatalf("CreatePayment calls = %d, want 2", calls)
	}
}

func TestExecuteCreatePaymentsDryRun(t *testing.T) {
	ctx := context.Background()
	rows := paymentRows()
	client := &mockClient{
		createPaymentFn: func(ctx context.Context, req *zaim.CreatePaymentRequest) error {
			t.Fatal("CreatePayment was called during dry run")
			return nil
		},
	}

	got := ExecuteCreatePayments(ctx, client, rows, true)

	want := BulkResult{Total: 2, Succeeded: 2, Failed: 0}
	assertBulkResult(t, got, want)
}

func TestExecuteUpdateMoneySuccess(t *testing.T) {
	ctx := context.Background()
	rows := []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"id":      "901",
				"mode":    "payment",
				"amount":  "2500",
				"date":    "2026-04-10",
				"comment": "dinner",
			},
		},
	}
	calls := 0
	client := &mockClient{
		updateMoneyFn: func(ctx context.Context, id int, mode string, req *zaim.UpdateMoneyRequest) error {
			calls++
			if id != 901 {
				t.Errorf("id = %d, want 901", id)
			}
			if mode != "payment" {
				t.Errorf("mode = %q, want %q", mode, "payment")
			}
			return nil
		},
	}

	got := ExecuteUpdateMoney(ctx, client, rows, false)

	want := BulkResult{Total: 1, Succeeded: 1, Failed: 0}
	assertBulkResult(t, got, want)
	if calls != 1 {
		t.Fatalf("UpdateMoney calls = %d, want 1", calls)
	}
}

func TestExecuteUpdateMoneyValidationError(t *testing.T) {
	ctx := context.Background()
	rows := []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"id":   "901",
				"mode": "refund",
			},
		},
	}
	client := &mockClient{
		updateMoneyFn: func(ctx context.Context, id int, mode string, req *zaim.UpdateMoneyRequest) error {
			t.Fatal("UpdateMoney was called for an invalid row")
			return nil
		},
	}

	got := ExecuteUpdateMoney(ctx, client, rows, false)

	want := BulkResult{
		Total:     1,
		Succeeded: 0,
		Failed:    1,
		Errors: []RowError{
			{Row: 1, Error: "mode"},
		},
	}
	assertBulkResult(t, got, want)
}

func paymentRows() []Row {
	return []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"amount":          "1200",
				"date":            "2026-04-10",
				"category_id":     "101",
				"genre_id":        "201",
				"from_account_id": "301",
				"comment":         "lunch",
			},
		},
		{
			Index: 2,
			Fields: map[string]string{
				"amount":          "3400",
				"date":            "2026-04-11",
				"category_id":     "102",
				"genre_id":        "202",
				"from_account_id": "302",
				"comment":         "book",
			},
		},
	}
}

func assertBulkResult(t *testing.T, got BulkResult, want BulkResult) {
	t.Helper()

	if got.Total != want.Total {
		t.Errorf("Total = %d, want %d", got.Total, want.Total)
	}
	if got.Succeeded != want.Succeeded {
		t.Errorf("Succeeded = %d, want %d", got.Succeeded, want.Succeeded)
	}
	if got.Failed != want.Failed {
		t.Errorf("Failed = %d, want %d", got.Failed, want.Failed)
	}
	if len(got.Errors) != len(want.Errors) {
		t.Fatalf("len(Errors) = %d, want %d; errors = %#v", len(got.Errors), len(want.Errors), got.Errors)
	}
	for i := range want.Errors {
		if got.Errors[i].Row != want.Errors[i].Row {
			t.Errorf("Errors[%d].Row = %d, want %d", i, got.Errors[i].Row, want.Errors[i].Row)
		}
		if !strings.Contains(got.Errors[i].Error, want.Errors[i].Error) {
			t.Errorf("Errors[%d].Error = %q, want it to contain %q", i, got.Errors[i].Error, want.Errors[i].Error)
		}
	}
}
