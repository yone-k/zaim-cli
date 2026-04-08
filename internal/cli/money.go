package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/formatter"
	"github.com/yone/zaim-cli/pkg/zaim"
)

var (
	moneyCmd = &cobra.Command{
		Use:   "money",
		Short: "家計簿レコードを管理",
	}

	moneyListMode       string
	moneyListStartDate  string
	moneyListEndDate    string
	moneyListCategoryID int
	moneyListLimit      int
	moneyListPage       int

	moneyCreatePaymentAmount        int
	moneyCreatePaymentDate          string
	moneyCreatePaymentCategoryID    int
	moneyCreatePaymentGenreID       int
	moneyCreatePaymentFromAccountID int
	moneyCreatePaymentPlace         string
	moneyCreatePaymentComment       string
	moneyCreatePaymentName          string

	moneyCreateIncomeAmount      int
	moneyCreateIncomeDate        string
	moneyCreateIncomeCategoryID  int
	moneyCreateIncomeToAccountID int
	moneyCreateIncomePlace       string
	moneyCreateIncomeComment     string

	moneyCreateTransferAmount        int
	moneyCreateTransferDate          string
	moneyCreateTransferFromAccountID int
	moneyCreateTransferToAccountID   int
	moneyCreateTransferComment       string

	moneyUpdateAmount        int
	moneyUpdateDate          string
	moneyUpdateCategoryID    int
	moneyUpdateGenreID       int
	moneyUpdateFromAccountID int
	moneyUpdateToAccountID   int
	moneyUpdatePlace         string
	moneyUpdateComment       string
	moneyUpdateName          string
)

var moneyListCmd = &cobra.Command{
	Use:   "list",
	Short: "家計簿レコード一覧",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		opts := &zaim.ListMoneyOptions{
			Mode:       moneyListMode,
			StartDate:  moneyListStartDate,
			EndDate:    moneyListEndDate,
			CategoryID: moneyListCategoryID,
			Limit:      moneyListLimit,
			Page:       moneyListPage,
		}

		records, err := Client.ListMoney(ctx, opts)
		if err != nil {
			return err
		}

		if moneyListLimit > 0 && len(records) > moneyListLimit {
			records = records[:moneyListLimit]
		}

		if OutputFormat == formatter.FormatJSON {
			return formatter.OutputJSON(cmd.OutOrStdout(), records)
		}

		rows := make([][]string, 0, len(records))
		for _, record := range records {
			rows = append(rows, []string{
				strconv.Itoa(record.ID),
				record.Date,
				record.Mode,
				strconv.Itoa(record.CategoryID),
				strconv.Itoa(record.GenreID),
				strconv.Itoa(record.Amount),
				record.Name,
				record.Place,
				record.Comment,
			})
		}

		formatter.RenderTable(
			cmd.OutOrStdout(),
			[]string{"ID", "日付", "種別", "カテゴリID", "ジャンルID", "金額", "名前", "場所", "コメント"},
			rows,
		)

		return nil
	},
}

var moneyCreatePaymentCmd = &cobra.Command{
	Use:   "create-payment",
	Short: "支出を記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		req := &zaim.CreatePaymentRequest{
			Amount:        moneyCreatePaymentAmount,
			Date:          moneyCreatePaymentDate,
			CategoryID:    moneyCreatePaymentCategoryID,
			GenreID:       moneyCreatePaymentGenreID,
			FromAccountID: moneyCreatePaymentFromAccountID,
			Place:         moneyCreatePaymentPlace,
			Comment:       moneyCreatePaymentComment,
			Name:          moneyCreatePaymentName,
		}

		if err := Client.CreatePayment(cmd.Context(), req); err != nil {
			return err
		}

		return outputMoneyCommandSuccess(cmd, "支出を記録しました")
	},
}

var moneyCreateIncomeCmd = &cobra.Command{
	Use:   "create-income",
	Short: "収入を記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		req := &zaim.CreateIncomeRequest{
			Amount:      moneyCreateIncomeAmount,
			Date:        moneyCreateIncomeDate,
			CategoryID:  moneyCreateIncomeCategoryID,
			ToAccountID: moneyCreateIncomeToAccountID,
			Place:       moneyCreateIncomePlace,
			Comment:     moneyCreateIncomeComment,
		}

		if err := Client.CreateIncome(cmd.Context(), req); err != nil {
			return err
		}

		return outputMoneyCommandSuccess(cmd, "収入を記録しました")
	},
}

var moneyCreateTransferCmd = &cobra.Command{
	Use:   "create-transfer",
	Short: "振替を記録",
	RunE: func(cmd *cobra.Command, _ []string) error {
		req := &zaim.CreateTransferRequest{
			Amount:        moneyCreateTransferAmount,
			Date:          moneyCreateTransferDate,
			FromAccountID: moneyCreateTransferFromAccountID,
			ToAccountID:   moneyCreateTransferToAccountID,
			Comment:       moneyCreateTransferComment,
		}

		if err := Client.CreateTransfer(cmd.Context(), req); err != nil {
			return err
		}

		return outputMoneyCommandSuccess(cmd, "振替を記録しました")
	},
}

var moneyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "レコードを更新",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[0], err)
		}

		mode := args[1]
		flags := cmd.Flags()
		req := &zaim.UpdateMoneyRequest{}

		if flags.Changed("amount") {
			req.Amount = &moneyUpdateAmount
		}
		if flags.Changed("date") {
			req.Date = &moneyUpdateDate
		}
		if flags.Changed("category-id") {
			req.CategoryID = &moneyUpdateCategoryID
		}
		if flags.Changed("genre-id") {
			req.GenreID = &moneyUpdateGenreID
		}
		if flags.Changed("from-account-id") {
			req.FromAccountID = &moneyUpdateFromAccountID
		}
		if flags.Changed("to-account-id") {
			req.ToAccountID = &moneyUpdateToAccountID
		}
		if flags.Changed("place") {
			req.Place = &moneyUpdatePlace
		}
		if flags.Changed("comment") {
			req.Comment = &moneyUpdateComment
		}
		if flags.Changed("name") {
			req.Name = &moneyUpdateName
		}

		if err := Client.UpdateMoney(cmd.Context(), id, mode, req); err != nil {
			return err
		}

		return outputMoneyCommandSuccess(cmd, "レコードを更新しました")
	},
}

var moneyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "レコードを削除",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[0], err)
		}

		mode := args[1]

		_, _ = fmt.Fprint(cmd.OutOrStdout(), "本当に削除しますか? (y/N): ")

		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			return err
		}
		if answer != "y" && answer != "Y" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "削除を中止しました")
			return nil
		}

		if err := Client.DeleteMoney(cmd.Context(), id, mode); err != nil {
			return err
		}

		return outputMoneyCommandSuccess(cmd, "レコードを削除しました")
	},
}

func init() {
	rootCmd.AddCommand(moneyCmd)
	moneyCmd.AddCommand(
		moneyListCmd,
		moneyCreatePaymentCmd,
		moneyCreateIncomeCmd,
		moneyCreateTransferCmd,
		moneyUpdateCmd,
		moneyDeleteCmd,
	)

	moneyListCmd.Flags().StringVar(&moneyListMode, "mode", "", "絞り込む種別")
	moneyListCmd.Flags().StringVar(&moneyListStartDate, "start-date", "", "開始日")
	moneyListCmd.Flags().StringVar(&moneyListEndDate, "end-date", "", "終了日")
	moneyListCmd.Flags().IntVar(&moneyListCategoryID, "category-id", 0, "カテゴリID")
	moneyListCmd.Flags().IntVar(&moneyListLimit, "limit", 0, "取得件数")
	moneyListCmd.Flags().IntVar(&moneyListPage, "page", 0, "ページ番号")

	moneyCreatePaymentCmd.Flags().IntVar(&moneyCreatePaymentAmount, "amount", 0, "金額")
	moneyCreatePaymentCmd.Flags().StringVar(&moneyCreatePaymentDate, "date", "", "日付")
	moneyCreatePaymentCmd.Flags().IntVar(&moneyCreatePaymentCategoryID, "category-id", 0, "カテゴリID")
	moneyCreatePaymentCmd.Flags().IntVar(&moneyCreatePaymentGenreID, "genre-id", 0, "ジャンルID")
	moneyCreatePaymentCmd.Flags().IntVar(&moneyCreatePaymentFromAccountID, "from-account-id", 0, "支払い元口座ID")
	moneyCreatePaymentCmd.Flags().StringVar(&moneyCreatePaymentPlace, "place", "", "場所")
	moneyCreatePaymentCmd.Flags().StringVar(&moneyCreatePaymentComment, "comment", "", "コメント")
	moneyCreatePaymentCmd.Flags().StringVar(&moneyCreatePaymentName, "name", "", "名前")
	markRequired(moneyCreatePaymentCmd, "amount", "date", "category-id", "genre-id")

	moneyCreateIncomeCmd.Flags().IntVar(&moneyCreateIncomeAmount, "amount", 0, "金額")
	moneyCreateIncomeCmd.Flags().StringVar(&moneyCreateIncomeDate, "date", "", "日付")
	moneyCreateIncomeCmd.Flags().IntVar(&moneyCreateIncomeCategoryID, "category-id", 0, "カテゴリID")
	moneyCreateIncomeCmd.Flags().IntVar(&moneyCreateIncomeToAccountID, "to-account-id", 0, "入金先口座ID")
	moneyCreateIncomeCmd.Flags().StringVar(&moneyCreateIncomePlace, "place", "", "場所")
	moneyCreateIncomeCmd.Flags().StringVar(&moneyCreateIncomeComment, "comment", "", "コメント")
	markRequired(moneyCreateIncomeCmd, "amount", "date", "category-id")

	moneyCreateTransferCmd.Flags().IntVar(&moneyCreateTransferAmount, "amount", 0, "金額")
	moneyCreateTransferCmd.Flags().StringVar(&moneyCreateTransferDate, "date", "", "日付")
	moneyCreateTransferCmd.Flags().IntVar(&moneyCreateTransferFromAccountID, "from-account-id", 0, "振替元口座ID")
	moneyCreateTransferCmd.Flags().IntVar(&moneyCreateTransferToAccountID, "to-account-id", 0, "振替先口座ID")
	moneyCreateTransferCmd.Flags().StringVar(&moneyCreateTransferComment, "comment", "", "コメント")
	markRequired(moneyCreateTransferCmd, "amount", "date", "from-account-id", "to-account-id")

	moneyUpdateCmd.Flags().IntVar(&moneyUpdateAmount, "amount", 0, "金額")
	moneyUpdateCmd.Flags().StringVar(&moneyUpdateDate, "date", "", "日付")
	moneyUpdateCmd.Flags().IntVar(&moneyUpdateCategoryID, "category-id", 0, "カテゴリID")
	moneyUpdateCmd.Flags().IntVar(&moneyUpdateGenreID, "genre-id", 0, "ジャンルID")
	moneyUpdateCmd.Flags().IntVar(&moneyUpdateFromAccountID, "from-account-id", 0, "支払い元口座ID")
	moneyUpdateCmd.Flags().IntVar(&moneyUpdateToAccountID, "to-account-id", 0, "入金先口座ID")
	moneyUpdateCmd.Flags().StringVar(&moneyUpdatePlace, "place", "", "場所")
	moneyUpdateCmd.Flags().StringVar(&moneyUpdateComment, "comment", "", "コメント")
	moneyUpdateCmd.Flags().StringVar(&moneyUpdateName, "name", "", "名前")
}

func markRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
}

func outputMoneyCommandSuccess(cmd *cobra.Command, message string) error {
	if OutputFormat == formatter.FormatJSON {
		return formatter.OutputJSON(cmd.OutOrStdout(), map[string]string{
			"status":  "success",
			"message": message,
		})
	}

	_, err := fmt.Fprintln(cmd.OutOrStdout(), message)
	return err
}
