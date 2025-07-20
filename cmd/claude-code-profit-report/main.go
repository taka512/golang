package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/taka512/golang/cmd/claude-code-profit-report/config"
	"github.com/taka512/golang/cmd/claude-code-profit-report/infrastructure/database"
	"github.com/taka512/golang/cmd/claude-code-profit-report/presentation/cli"
	"github.com/taka512/golang/cmd/claude-code-profit-report/presentation/slack"
)

var (
	companyID   uint
	warehouseID uint
	startDate   string
	endDate     string
	outputSlack bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "claude-code-profit-report",
		Short: "売上・コスト・粗利のトレンドを表示するコマンド",
		Long:  `指定期間の売上・コスト・粗利のトレンドを表示します。出力先は標準出力またはSlackです。`,
		RunE:  runCommand,
	}

	rootCmd.Flags().UintVarP(&companyID, "company", "c", 0, "会社ID (オプション: 未指定時は全社)")
	rootCmd.Flags().UintVarP(&warehouseID, "warehouse", "w", 0, "倉庫ID (オプション: 未指定時は全倉庫)")
	rootCmd.Flags().StringVarP(&startDate, "start", "s", "", "開始日 (YYYY-MM-DD) (必須)")
	rootCmd.Flags().StringVarP(&endDate, "end", "e", "", "終了日 (YYYY-MM-DD) (必須)")
	rootCmd.Flags().BoolVar(&outputSlack, "slack", false, "Slackに出力する")

	rootCmd.MarkFlagRequired("start")
	rootCmd.MarkFlagRequired("end")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	if start.After(end) {
		return fmt.Errorf("start date must be before or equal to end date")
	}

	dbConfig := database.NewDBConfig()
	db, err := database.NewDB(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	container := config.NewContainer(db)

	report, err := container.ProfitReportUseCase.GenerateProfitReport(ctx, companyID, warehouseID, start, end)
	if err != nil {
		return fmt.Errorf("failed to generate profit report: %w", err)
	}

	formatter := cli.NewTextFormatter()
	output := formatter.FormatProfitReport(report)
	fmt.Print(output)

	if outputSlack {
		webhookURL := os.Getenv("SLACK_HOOK")
		if webhookURL == "" {
			return fmt.Errorf("SLACK_HOOK environment variable is not set")
		}

		slackClient := slack.NewClient(webhookURL)
		if err := slackClient.SendProfitReport(report); err != nil {
			return fmt.Errorf("failed to send to slack: %w", err)
		}
		fmt.Println("\nSlackに送信しました。")
	}

	return nil
}