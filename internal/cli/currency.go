package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yone-k/zaim-cli/internal/formatter"
)

var (
	currencyCmd = &cobra.Command{
		Use:   "currency",
		Short: "通貨を管理",
	}

	currencyListCmd = &cobra.Command{
		Use:   "list",
		Short: "通貨一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			currencies, err := Client.ListCurrencies(cmd.Context())
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(cmd.OutOrStdout(), currencies)
			case "table":
				header := []string{"コード", "名前", "単位"}
				rows := make([][]string, 0, len(currencies))
				for _, currency := range currencies {
					rows = append(rows, []string{
						currency.CurrencyCode,
						currency.Name,
						currency.Unit,
					})
				}
				formatter.RenderTable(cmd.OutOrStdout(), header, rows)
				return nil
			default:
				return fmt.Errorf("invalid output format %q", OutputFormat)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(currencyCmd)
	currencyCmd.AddCommand(currencyListCmd)
}
