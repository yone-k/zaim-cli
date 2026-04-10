package bulk

import (
	"fmt"
	"strconv"
	"time"
)

func ValidatePaymentRows(rows []Row) []RowError {
	return validateRows(rows, []fieldRule{
		numericField("amount"),
		dateField("date"),
		numericField("category_id"),
		numericField("genre_id"),
	})
}

func ValidateIncomeRows(rows []Row) []RowError {
	return validateRows(rows, []fieldRule{
		numericField("amount"),
		dateField("date"),
		numericField("category_id"),
	})
}

func ValidateTransferRows(rows []Row) []RowError {
	return validateRows(rows, []fieldRule{
		numericField("amount"),
		dateField("date"),
		numericField("from_account_id"),
		numericField("to_account_id"),
	})
}

func ValidateUpdateRows(rows []Row) []RowError {
	return validateRows(rows, []fieldRule{
		numericField("id"),
		modeField("mode"),
	})
}

type fieldRule struct {
	name     string
	validate func(value string) string
}

func numericField(name string) fieldRule {
	return fieldRule{
		name: name,
		validate: func(value string) string {
			if _, err := strconv.Atoi(value); err != nil {
				return fmt.Sprintf("%s must be numeric", name)
			}
			return ""
		},
	}
}

func dateField(name string) fieldRule {
	return fieldRule{
		name: name,
		validate: func(value string) string {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return fmt.Sprintf("%s must be in YYYY-MM-DD format", name)
			}
			return ""
		},
	}
}

func modeField(name string) fieldRule {
	return fieldRule{
		name: name,
		validate: func(value string) string {
			switch value {
			case "payment", "income", "transfer":
				return ""
			default:
				return fmt.Sprintf("%s must be one of payment, income, transfer", name)
			}
		},
	}
}

func validateRows(rows []Row, rules []fieldRule) []RowError {
	var errors []RowError
	for _, row := range rows {
		for _, rule := range rules {
			value, ok := row.Fields[rule.name]
			if !ok || value == "" {
				errors = append(errors, RowError{
					Row:   row.Index,
					Error: fmt.Sprintf("%s is required", rule.name),
				})
				continue
			}
			if message := rule.validate(value); message != "" {
				errors = append(errors, RowError{
					Row:   row.Index,
					Error: message,
				})
			}
		}
	}
	return errors
}
