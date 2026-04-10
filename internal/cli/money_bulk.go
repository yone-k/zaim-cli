package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yone-k/zaim-cli/internal/bulk"
	"github.com/yone-k/zaim-cli/internal/formatter"
)

var (
	moneyBulkCreatePaymentFile   string
	moneyBulkCreatePaymentDryRun bool

	moneyBulkCreateIncomeFile   string
	moneyBulkCreateIncomeDryRun bool

	moneyBulkCreateTransferFile   string
	moneyBulkCreateTransferDryRun bool

	moneyBulkUpdateFile   string
	moneyBulkUpdateDryRun bool
)

var moneyBulkCreatePaymentCmd = &cobra.Command{
	Use:   "bulk-create-payment",
	Short: "支出を一括記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runMoneyBulkCommand(cmd, moneyBulkCreatePaymentFile, moneyBulkCreatePaymentDryRun, bulk.ExecuteCreatePayments)
	},
}

var moneyBulkCreateIncomeCmd = &cobra.Command{
	Use:   "bulk-create-income",
	Short: "収入を一括記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runMoneyBulkCommand(cmd, moneyBulkCreateIncomeFile, moneyBulkCreateIncomeDryRun, bulk.ExecuteCreateIncomes)
	},
}

var moneyBulkCreateTransferCmd = &cobra.Command{
	Use:   "bulk-create-transfer",
	Short: "振替を一括記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runMoneyBulkCommand(cmd, moneyBulkCreateTransferFile, moneyBulkCreateTransferDryRun, bulk.ExecuteCreateTransfers)
	},
}

var moneyBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "レコードを一括更新",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runMoneyBulkCommand(cmd, moneyBulkUpdateFile, moneyBulkUpdateDryRun, bulk.ExecuteUpdateMoney)
	},
}

func init() {
	moneyCmd.AddCommand(
		moneyBulkCreatePaymentCmd,
		moneyBulkCreateIncomeCmd,
		moneyBulkCreateTransferCmd,
		moneyBulkUpdateCmd,
	)

	moneyBulkCreatePaymentCmd.Flags().StringVar(&moneyBulkCreatePaymentFile, "file", "", "一括処理するファイル")
	moneyBulkCreatePaymentCmd.Flags().BoolVar(&moneyBulkCreatePaymentDryRun, "dry-run", false, "APIを呼び出さずに検証のみ行う")
	markRequired(moneyBulkCreatePaymentCmd, "file")

	moneyBulkCreateIncomeCmd.Flags().StringVar(&moneyBulkCreateIncomeFile, "file", "", "一括処理するファイル")
	moneyBulkCreateIncomeCmd.Flags().BoolVar(&moneyBulkCreateIncomeDryRun, "dry-run", false, "APIを呼び出さずに検証のみ行う")
	markRequired(moneyBulkCreateIncomeCmd, "file")

	moneyBulkCreateTransferCmd.Flags().StringVar(&moneyBulkCreateTransferFile, "file", "", "一括処理するファイル")
	moneyBulkCreateTransferCmd.Flags().BoolVar(&moneyBulkCreateTransferDryRun, "dry-run", false, "APIを呼び出さずに検証のみ行う")
	markRequired(moneyBulkCreateTransferCmd, "file")

	moneyBulkUpdateCmd.Flags().StringVar(&moneyBulkUpdateFile, "file", "", "一括処理するファイル")
	moneyBulkUpdateCmd.Flags().BoolVar(&moneyBulkUpdateDryRun, "dry-run", false, "APIを呼び出さずに検証のみ行う")
	markRequired(moneyBulkUpdateCmd, "file")
}

type moneyBulkExecutor func(ctx context.Context, client bulk.MoneyClient, rows []bulk.Row, dryRun bool) bulk.BulkResult

func runMoneyBulkCommand(cmd *cobra.Command, file string, dryRun bool, execute moneyBulkExecutor) error {
	rows, err := bulk.ParseFile(file)
	if err != nil {
		return err
	}

	result := execute(cmd.Context(), Client, rows, dryRun)
	if OutputFormat == formatter.FormatJSON {
		return formatter.OutputJSON(cmd.OutOrStdout(), result)
	}

	return outputMoneyBulkResult(cmd, result)
}

func outputMoneyBulkResult(cmd *cobra.Command, result bulk.BulkResult) error {
	if _, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"結果: %d件成功, %d件失敗 (全%d件)\n",
		result.Succeeded,
		result.Failed,
		result.Total,
	); err != nil {
		return err
	}

	for _, rowErr := range result.Errors {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  行%d: %s\n", rowErr.Row, rowErr.Error); err != nil {
			return err
		}
	}

	return nil
}
