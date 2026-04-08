package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/formatter"
)

var (
	genreDefaultsMode string

	genreCmd = &cobra.Command{
		Use:   "genre",
		Short: "ジャンルを管理",
	}

	genreListCmd = &cobra.Command{
		Use:   "list",
		Short: "ユーザージャンル一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			genres, err := Client.ListUserGenres(cmd.Context())
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(os.Stdout, genres)
			case "table":
				header := []string{"ID", "名前", "カテゴリID", "ソート順", "有効"}
				rows := make([][]string, 0, len(genres))
				for _, genre := range genres {
					rows = append(rows, []string{
						strconv.Itoa(genre.ID),
						genre.Name,
						strconv.Itoa(genre.CategoryID),
						strconv.Itoa(genre.Sort),
						activeLabel(genre.Active),
					})
				}
				formatter.RenderTable(os.Stdout, header, rows)
				return nil
			default:
				return fmt.Errorf("invalid output format %q", OutputFormat)
			}
		},
	}

	genreDefaultsCmd = &cobra.Command{
		Use:   "defaults",
		Short: "デフォルトジャンル一覧",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateMode(genreDefaultsMode); err != nil {
				return err
			}

			genres, err := Client.ListDefaultGenres(cmd.Context(), genreDefaultsMode)
			if err != nil {
				return err
			}

			switch OutputFormat {
			case "json":
				return formatter.OutputJSON(os.Stdout, genres)
			case "table":
				header := []string{"ID", "名前", "カテゴリID", "ソート順", "有効"}
				rows := make([][]string, 0, len(genres))
				for _, genre := range genres {
					rows = append(rows, []string{
						strconv.Itoa(genre.ID),
						genre.Name,
						strconv.Itoa(genre.CategoryID),
						strconv.Itoa(genre.Sort),
						activeLabel(genre.Active),
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
	genreDefaultsCmd.Flags().StringVar(&genreDefaultsMode, "mode", "", "Genre mode (payment or income)")
	_ = genreDefaultsCmd.MarkFlagRequired("mode")

	rootCmd.AddCommand(genreCmd)
	genreCmd.AddCommand(genreListCmd, genreDefaultsCmd)
}
