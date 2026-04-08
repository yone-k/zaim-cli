package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/formatter"
)

var (
	categoryDefaultsMode string

	categoryCmd = &cobra.Command{
		Use:   "category",
		Short: "カテゴリを管理",
	}

	categoryListCmd = &cobra.Command{
		Use:   "list",
		Short: "ユーザーカテゴリ一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			categories, err := Client.ListUserCategories(cmd.Context())
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(cmd.OutOrStdout(), categories)
			case "table":
				header := []string{"ID", "名前", "種別", "ソート順", "有効"}
				rows := make([][]string, 0, len(categories))
				for _, category := range categories {
					rows = append(rows, []string{
						strconv.Itoa(category.ID),
						category.Name,
						category.Mode,
						strconv.Itoa(category.Sort),
						activeLabel(category.Active),
					})
				}
				formatter.RenderTable(cmd.OutOrStdout(), header, rows)
				return nil
			default:
				return fmt.Errorf("invalid output format %q", OutputFormat)
			}
		},
	}

	categoryDefaultsCmd = &cobra.Command{
		Use:   "defaults",
		Short: "デフォルトカテゴリ一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateMode(categoryDefaultsMode); err != nil {
				return err
			}

			categories, err := Client.ListDefaultCategories(cmd.Context(), categoryDefaultsMode)
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(cmd.OutOrStdout(), categories)
			case "table":
				header := []string{"ID", "名前", "種別", "ソート順", "有効"}
				rows := make([][]string, 0, len(categories))
				for _, category := range categories {
					rows = append(rows, []string{
						strconv.Itoa(category.ID),
						category.Name,
						category.Mode,
						strconv.Itoa(category.Sort),
						activeLabel(category.Active),
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
	categoryDefaultsCmd.Flags().StringVar(&categoryDefaultsMode, "mode", "", "Category mode (payment or income)")
	_ = categoryDefaultsCmd.MarkFlagRequired("mode")

	rootCmd.AddCommand(categoryCmd)
	categoryCmd.AddCommand(categoryListCmd, categoryDefaultsCmd)
}

func activeLabel(active int) string {
	if active == 1 {
		return "有効"
	}

	return "無効"
}

func validateMode(mode string) error {
	if mode == "payment" || mode == "income" {
		return nil
	}

	return fmt.Errorf("invalid mode %q: must be payment or income", mode)
}
