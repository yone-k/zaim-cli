package bulk

import (
	"reflect"
	"testing"

	"github.com/yone-k/zaim-cli/pkg/zaim"
)

func TestToCreatePaymentRequest(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":          "1200",
			"date":            "2026-04-10",
			"category_id":     "101",
			"genre_id":        "201",
			"from_account_id": "301",
			"place":           "Tokyo",
			"comment":         "lunch",
			"name":            "team lunch",
		},
	}

	got, err := ToCreatePaymentRequest(row)
	if err != nil {
		t.Fatalf("ToCreatePaymentRequest() error = %v", err)
	}

	want := &zaim.CreatePaymentRequest{
		Amount:        1200,
		Date:          "2026-04-10",
		CategoryID:    101,
		GenreID:       201,
		FromAccountID: 301,
		Place:         "Tokyo",
		Comment:       "lunch",
		Name:          "team lunch",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToCreatePaymentRequest() = %#v, want %#v", got, want)
	}
}

func TestToCreateIncomeRequest(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":        "300000",
			"date":          "2026-04-10",
			"category_id":   "401",
			"to_account_id": "501",
			"place":         "Office",
			"comment":       "salary",
		},
	}

	got, err := ToCreateIncomeRequest(row)
	if err != nil {
		t.Fatalf("ToCreateIncomeRequest() error = %v", err)
	}

	want := &zaim.CreateIncomeRequest{
		Amount:      300000,
		Date:        "2026-04-10",
		CategoryID:  401,
		ToAccountID: 501,
		Place:       "Office",
		Comment:     "salary",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToCreateIncomeRequest() = %#v, want %#v", got, want)
	}
}

func TestToCreateTransferRequest(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":          "50000",
			"date":            "2026-04-10",
			"from_account_id": "601",
			"to_account_id":   "602",
			"comment":         "savings",
		},
	}

	got, err := ToCreateTransferRequest(row)
	if err != nil {
		t.Fatalf("ToCreateTransferRequest() error = %v", err)
	}

	want := &zaim.CreateTransferRequest{
		Amount:        50000,
		Date:          "2026-04-10",
		FromAccountID: 601,
		ToAccountID:   602,
		Comment:       "savings",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToCreateTransferRequest() = %#v, want %#v", got, want)
	}
}

func TestToUpdateMoneyRequest(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"id":              "901",
			"mode":            "payment",
			"amount":          "2500",
			"date":            "2026-04-10",
			"category_id":     "101",
			"genre_id":        "201",
			"from_account_id": "301",
			"to_account_id":   "302",
			"place":           "Tokyo",
			"comment":         "dinner",
			"name":            "team dinner",
		},
	}

	gotID, gotMode, gotReq, err := ToUpdateMoneyRequest(row)
	if err != nil {
		t.Fatalf("ToUpdateMoneyRequest() error = %v", err)
	}

	wantID := 901
	wantMode := "payment"
	wantReq := &zaim.UpdateMoneyRequest{
		Amount:        intPtr(2500),
		Date:          stringPtr("2026-04-10"),
		CategoryID:    intPtr(101),
		GenreID:       intPtr(201),
		FromAccountID: intPtr(301),
		ToAccountID:   intPtr(302),
		Place:         stringPtr("Tokyo"),
		Comment:       stringPtr("dinner"),
		Name:          stringPtr("team dinner"),
	}
	if gotID != wantID {
		t.Errorf("id = %d, want %d", gotID, wantID)
	}
	if gotMode != wantMode {
		t.Errorf("mode = %q, want %q", gotMode, wantMode)
	}
	if !reflect.DeepEqual(gotReq, wantReq) {
		t.Errorf("request = %#v, want %#v", gotReq, wantReq)
	}
}

func TestToUpdateMoneyRequestEmptyFieldsAreNil(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"id":              "902",
			"mode":            "income",
			"amount":          "",
			"date":            "",
			"category_id":     "",
			"genre_id":        "",
			"from_account_id": "",
			"to_account_id":   "",
			"place":           "",
			"comment":         "",
			"name":            "",
		},
	}

	gotID, gotMode, gotReq, err := ToUpdateMoneyRequest(row)
	if err != nil {
		t.Fatalf("ToUpdateMoneyRequest() error = %v", err)
	}

	if gotID != 902 {
		t.Errorf("id = %d, want 902", gotID)
	}
	if gotMode != "income" {
		t.Errorf("mode = %q, want %q", gotMode, "income")
	}
	wantReq := &zaim.UpdateMoneyRequest{}
	if !reflect.DeepEqual(gotReq, wantReq) {
		t.Errorf("request = %#v, want %#v", gotReq, wantReq)
	}
}

func TestToCreatePaymentRequestNumericParseError(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":          "bad",
			"category_id":     "101",
			"genre_id":        "201",
			"from_account_id": "301",
		},
	}

	if _, err := ToCreatePaymentRequest(row); err == nil {
		t.Fatal("ToCreatePaymentRequest() error = nil, want error")
	}
}

func TestToCreateIncomeRequestNumericParseError(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":        "300000",
			"category_id":   "bad",
			"to_account_id": "501",
		},
	}

	if _, err := ToCreateIncomeRequest(row); err == nil {
		t.Fatal("ToCreateIncomeRequest() error = nil, want error")
	}
}

func TestToCreateTransferRequestNumericParseError(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"amount":          "50000",
			"from_account_id": "bad",
			"to_account_id":   "602",
		},
	}

	if _, err := ToCreateTransferRequest(row); err == nil {
		t.Fatal("ToCreateTransferRequest() error = nil, want error")
	}
}

func TestToUpdateMoneyRequestNumericParseError(t *testing.T) {
	row := Row{
		Index: 1,
		Fields: map[string]string{
			"id":     "901",
			"mode":   "payment",
			"amount": "bad",
		},
	}

	if _, _, _, err := ToUpdateMoneyRequest(row); err == nil {
		t.Fatal("ToUpdateMoneyRequest() error = nil, want error")
	}
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
