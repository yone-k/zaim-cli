package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yone-k/zaim-cli/internal/formatter"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "ユーザー情報を表示",
	RunE: func(cmd *cobra.Command, _ []string) error {
		user, err := Client.VerifyAuth(cmd.Context())
		if err != nil {
			return err
		}

		switch OutputFormat {
		case "json":
			return formatter.OutputJSON(cmd.OutOrStdout(), user)
		case "table":
			header := []string{"ID", "ログイン名", "名前", "通貨", "入力回数", "継続日数"}
			rows := [][]string{
				{
					fmt.Sprintf("%d", user.ID),
					user.Login,
					user.Name,
					user.CurrencyCode,
					fmt.Sprintf("%d", user.InputCount),
					fmt.Sprintf("%d", user.RepeatCount),
				},
			}
			formatter.RenderTable(cmd.OutOrStdout(), header, rows)
			return nil
		default:
			return fmt.Errorf("invalid output format %q", OutputFormat)
		}
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
}
