package bulk

import (
	"strings"
	"testing"
)

func TestValidatePaymentRows(t *testing.T) {
	tests := []struct {
		name string
		rows []Row
		want []expectedRowError
	}{
		{
			name: "valid row has no errors",
			rows: []Row{
				row(1, map[string]string{
					"amount":      "1200",
					"date":        "2026-04-10",
					"category_id": "101",
					"genre_id":    "201",
				}),
			},
		},
		{
			name: "missing required fields",
			rows: []Row{
				row(2, map[string]string{}),
			},
			want: []expectedRowError{
				{row: 2, contains: []string{"amount"}},
				{row: 2, contains: []string{"date"}},
				{row: 2, contains: []string{"category_id"}},
				{row: 2, contains: []string{"genre_id"}},
			},
		},
		{
			name: "numeric and date fields are invalid",
			rows: []Row{
				row(3, map[string]string{
					"amount":      "not-a-number",
					"date":        "04/10/2026",
					"category_id": "food",
					"genre_id":    "lunch",
				}),
			},
			want: []expectedRowError{
				{row: 3, contains: []string{"amount", "numeric"}},
				{row: 3, contains: []string{"date", "YYYY-MM-DD"}},
				{row: 3, contains: []string{"category_id", "numeric"}},
				{row: 3, contains: []string{"genre_id", "numeric"}},
			},
		},
		{
			name: "multiple rows return all errors",
			rows: []Row{
				row(4, map[string]string{
					"amount":      "",
					"date":        "2026-04-10",
					"category_id": "101",
					"genre_id":    "201",
				}),
				row(5, map[string]string{
					"amount":      "1200",
					"date":        "2026/04/10",
					"category_id": "101",
					"genre_id":    "bad",
				}),
			},
			want: []expectedRowError{
				{row: 4, contains: []string{"amount"}},
				{row: 5, contains: []string{"date", "YYYY-MM-DD"}},
				{row: 5, contains: []string{"genre_id", "numeric"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePaymentRows(tt.rows)
			assertRowErrors(t, got, tt.want)
		})
	}
}

func TestValidateIncomeRows(t *testing.T) {
	tests := []struct {
		name string
		rows []Row
		want []expectedRowError
	}{
		{
			name: "valid row has no errors",
			rows: []Row{
				row(1, map[string]string{
					"amount":      "300000",
					"date":        "2026-04-10",
					"category_id": "301",
				}),
			},
		},
		{
			name: "missing required fields",
			rows: []Row{
				row(2, map[string]string{}),
			},
			want: []expectedRowError{
				{row: 2, contains: []string{"amount"}},
				{row: 2, contains: []string{"date"}},
				{row: 2, contains: []string{"category_id"}},
			},
		},
		{
			name: "numeric and date fields are invalid",
			rows: []Row{
				row(3, map[string]string{
					"amount":      "salary",
					"date":        "20260410",
					"category_id": "bonus",
				}),
			},
			want: []expectedRowError{
				{row: 3, contains: []string{"amount", "numeric"}},
				{row: 3, contains: []string{"date", "YYYY-MM-DD"}},
				{row: 3, contains: []string{"category_id", "numeric"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateIncomeRows(tt.rows)
			assertRowErrors(t, got, tt.want)
		})
	}
}

func TestValidateTransferRows(t *testing.T) {
	tests := []struct {
		name string
		rows []Row
		want []expectedRowError
	}{
		{
			name: "valid row has no errors",
			rows: []Row{
				row(1, map[string]string{
					"amount":          "50000",
					"date":            "2026-04-10",
					"from_account_id": "401",
					"to_account_id":   "402",
				}),
			},
		},
		{
			name: "missing required fields",
			rows: []Row{
				row(2, map[string]string{}),
			},
			want: []expectedRowError{
				{row: 2, contains: []string{"amount"}},
				{row: 2, contains: []string{"date"}},
				{row: 2, contains: []string{"from_account_id"}},
				{row: 2, contains: []string{"to_account_id"}},
			},
		},
		{
			name: "numeric and date fields are invalid",
			rows: []Row{
				row(3, map[string]string{
					"amount":          "many",
					"date":            "10-04-2026",
					"from_account_id": "wallet",
					"to_account_id":   "bank",
				}),
			},
			want: []expectedRowError{
				{row: 3, contains: []string{"amount", "numeric"}},
				{row: 3, contains: []string{"date", "YYYY-MM-DD"}},
				{row: 3, contains: []string{"from_account_id", "numeric"}},
				{row: 3, contains: []string{"to_account_id", "numeric"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTransferRows(tt.rows)
			assertRowErrors(t, got, tt.want)
		})
	}
}

func TestValidateUpdateRows(t *testing.T) {
	tests := []struct {
		name string
		rows []Row
		want []expectedRowError
	}{
		{
			name: "valid payment row has no errors",
			rows: []Row{
				row(1, map[string]string{
					"id":   "901",
					"mode": "payment",
				}),
			},
		},
		{
			name: "valid income row has no errors",
			rows: []Row{
				row(2, map[string]string{
					"id":   "902",
					"mode": "income",
				}),
			},
		},
		{
			name: "valid transfer row has no errors",
			rows: []Row{
				row(3, map[string]string{
					"id":   "903",
					"mode": "transfer",
				}),
			},
		},
		{
			name: "missing required fields",
			rows: []Row{
				row(4, map[string]string{}),
			},
			want: []expectedRowError{
				{row: 4, contains: []string{"id"}},
				{row: 4, contains: []string{"mode"}},
			},
		},
		{
			name: "id is invalid and mode is unsupported",
			rows: []Row{
				row(5, map[string]string{
					"id":   "abc",
					"mode": "refund",
				}),
			},
			want: []expectedRowError{
				{row: 5, contains: []string{"id", "numeric"}},
				{row: 5, contains: []string{"mode", "payment", "income", "transfer"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUpdateRows(tt.rows)
			assertRowErrors(t, got, tt.want)
		})
	}
}

type expectedRowError struct {
	row      int
	contains []string
}

func row(index int, fields map[string]string) Row {
	return Row{
		Index:  index,
		Fields: fields,
	}
}

func assertRowErrors(t *testing.T, got []RowError, want []expectedRowError) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(errors) = %d, want %d; errors = %#v", len(got), len(want), got)
	}

	for i := range want {
		if got[i].Row != want[i].row {
			t.Errorf("errors[%d].Row = %d, want %d", i, got[i].Row, want[i].row)
		}
		for _, part := range want[i].contains {
			if !strings.Contains(got[i].Error, part) {
				t.Errorf("errors[%d].Error = %q, want it to contain %q", i, got[i].Error, part)
			}
		}
	}
}
