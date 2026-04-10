package bulk

import (
	"fmt"
	"strconv"

	"github.com/yone-k/zaim-cli/pkg/zaim"
)

func ToCreatePaymentRequest(row Row) (*zaim.CreatePaymentRequest, error) {
	amount, err := parseIntField(row, "amount")
	if err != nil {
		return nil, err
	}
	categoryID, err := parseIntField(row, "category_id")
	if err != nil {
		return nil, err
	}
	genreID, err := parseIntField(row, "genre_id")
	if err != nil {
		return nil, err
	}
	fromAccountID, err := parseIntField(row, "from_account_id")
	if err != nil {
		return nil, err
	}

	return &zaim.CreatePaymentRequest{
		Amount:        amount,
		Date:          row.Fields["date"],
		CategoryID:    categoryID,
		GenreID:       genreID,
		FromAccountID: fromAccountID,
		Place:         row.Fields["place"],
		Comment:       row.Fields["comment"],
		Name:          row.Fields["name"],
	}, nil
}

func ToCreateIncomeRequest(row Row) (*zaim.CreateIncomeRequest, error) {
	amount, err := parseIntField(row, "amount")
	if err != nil {
		return nil, err
	}
	categoryID, err := parseIntField(row, "category_id")
	if err != nil {
		return nil, err
	}
	toAccountID, err := parseIntField(row, "to_account_id")
	if err != nil {
		return nil, err
	}

	return &zaim.CreateIncomeRequest{
		Amount:      amount,
		Date:        row.Fields["date"],
		CategoryID:  categoryID,
		ToAccountID: toAccountID,
		Place:       row.Fields["place"],
		Comment:     row.Fields["comment"],
	}, nil
}

func ToCreateTransferRequest(row Row) (*zaim.CreateTransferRequest, error) {
	amount, err := parseIntField(row, "amount")
	if err != nil {
		return nil, err
	}
	fromAccountID, err := parseIntField(row, "from_account_id")
	if err != nil {
		return nil, err
	}
	toAccountID, err := parseIntField(row, "to_account_id")
	if err != nil {
		return nil, err
	}

	return &zaim.CreateTransferRequest{
		Amount:        amount,
		Date:          row.Fields["date"],
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Comment:       row.Fields["comment"],
	}, nil
}

func ToUpdateMoneyRequest(row Row) (int, string, *zaim.UpdateMoneyRequest, error) {
	id, err := strconv.Atoi(row.Fields["id"])
	if err != nil {
		return 0, "", nil, fmt.Errorf("parse id: %w", err)
	}

	amount, err := parseOptionalIntField(row, "amount")
	if err != nil {
		return 0, "", nil, err
	}
	categoryID, err := parseOptionalIntField(row, "category_id")
	if err != nil {
		return 0, "", nil, err
	}
	genreID, err := parseOptionalIntField(row, "genre_id")
	if err != nil {
		return 0, "", nil, err
	}
	fromAccountID, err := parseOptionalIntField(row, "from_account_id")
	if err != nil {
		return 0, "", nil, err
	}
	toAccountID, err := parseOptionalIntField(row, "to_account_id")
	if err != nil {
		return 0, "", nil, err
	}

	return id, row.Fields["mode"], &zaim.UpdateMoneyRequest{
		Amount:        amount,
		Date:          optionalStringField(row, "date"),
		CategoryID:    categoryID,
		GenreID:       genreID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Place:         optionalStringField(row, "place"),
		Comment:       optionalStringField(row, "comment"),
		Name:          optionalStringField(row, "name"),
	}, nil
}

func parseIntField(row Row, name string) (int, error) {
	value := row.Fields[name]
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}

func parseOptionalIntField(row Row, name string) (*int, error) {
	value := row.Fields[name]
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	return &parsed, nil
}

func optionalStringField(row Row, name string) *string {
	value := row.Fields[name]
	if value == "" {
		return nil
	}
	return &value
}
