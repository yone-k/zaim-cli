package bulk

import (
	"context"

	"github.com/yone-k/zaim-cli/pkg/zaim"
)

type MoneyClient interface {
	CreatePayment(ctx context.Context, req *zaim.CreatePaymentRequest) error
	CreateIncome(ctx context.Context, req *zaim.CreateIncomeRequest) error
	CreateTransfer(ctx context.Context, req *zaim.CreateTransferRequest) error
	UpdateMoney(ctx context.Context, id int, mode string, req *zaim.UpdateMoneyRequest) error
}

func ExecuteCreatePayments(ctx context.Context, client MoneyClient, rows []Row, dryRun bool) BulkResult {
	result := BulkResult{Total: len(rows)}
	for _, row := range rows {
		if validationErrors := ValidatePaymentRows([]Row{row}); len(validationErrors) > 0 {
			result.Errors = append(result.Errors, validationErrors...)
			result.Failed++
			continue
		}

		req, err := ToCreatePaymentRequest(row)
		if err != nil {
			result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
			result.Failed++
			continue
		}

		if !dryRun {
			if err := client.CreatePayment(ctx, req); err != nil {
				result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
				result.Failed++
				continue
			}
		}
		result.Succeeded++
	}
	return result
}

func ExecuteCreateIncomes(ctx context.Context, client MoneyClient, rows []Row, dryRun bool) BulkResult {
	result := BulkResult{Total: len(rows)}
	for _, row := range rows {
		if validationErrors := ValidateIncomeRows([]Row{row}); len(validationErrors) > 0 {
			result.Errors = append(result.Errors, validationErrors...)
			result.Failed++
			continue
		}

		req, err := ToCreateIncomeRequest(row)
		if err != nil {
			result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
			result.Failed++
			continue
		}

		if !dryRun {
			if err := client.CreateIncome(ctx, req); err != nil {
				result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
				result.Failed++
				continue
			}
		}
		result.Succeeded++
	}
	return result
}

func ExecuteCreateTransfers(ctx context.Context, client MoneyClient, rows []Row, dryRun bool) BulkResult {
	result := BulkResult{Total: len(rows)}
	for _, row := range rows {
		if validationErrors := ValidateTransferRows([]Row{row}); len(validationErrors) > 0 {
			result.Errors = append(result.Errors, validationErrors...)
			result.Failed++
			continue
		}

		req, err := ToCreateTransferRequest(row)
		if err != nil {
			result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
			result.Failed++
			continue
		}

		if !dryRun {
			if err := client.CreateTransfer(ctx, req); err != nil {
				result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
				result.Failed++
				continue
			}
		}
		result.Succeeded++
	}
	return result
}

func ExecuteUpdateMoney(ctx context.Context, client MoneyClient, rows []Row, dryRun bool) BulkResult {
	result := BulkResult{Total: len(rows)}
	for _, row := range rows {
		if validationErrors := ValidateUpdateRows([]Row{row}); len(validationErrors) > 0 {
			result.Errors = append(result.Errors, validationErrors...)
			result.Failed++
			continue
		}

		id, mode, req, err := ToUpdateMoneyRequest(row)
		if err != nil {
			result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
			result.Failed++
			continue
		}

		if !dryRun {
			if err := client.UpdateMoney(ctx, id, mode, req); err != nil {
				result.Errors = append(result.Errors, RowError{Row: row.Index, Error: err.Error()})
				result.Failed++
				continue
			}
		}
		result.Succeeded++
	}
	return result
}
