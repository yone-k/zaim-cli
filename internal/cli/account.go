package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/formatter"
)

var (
	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "アカウントを管理",
	}

	accountListCmd = &cobra.Command{
		Use:   "list",
		Short: "ユーザーアカウント一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			accounts, err := Client.ListUserAccounts(cmd.Context())
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(os.Stdout, accounts)
			case "table":
				header := []string{"ID", "名前", "種別", "ソート順", "有効"}
				rows := make([][]string, 0, len(accounts))
				for _, account := range accounts {
					rows = append(rows, []string{
						strconv.Itoa(account.ID),
						account.Name,
						account.Mode,
						strconv.Itoa(account.Sort),
						activeLabel(account.Active),
					})
				}
				formatter.RenderTable(os.Stdout, header, rows)
				return nil
			default:
				return fmt.Errorf("invalid output format %q", OutputFormat)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.AddCommand(accountListCmd)
}
