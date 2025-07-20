package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"profit-trend-display/internal/calculator"
	"profit-trend-display/internal/chart"
	"profit-trend-display/internal/database"
	"profit-trend-display/internal/models"
)

const (
	defaultDSN  = "root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true"
	defaultDays = 30
)

func main() {
	// Command line flags
	var (
		dsn        = flag.String("dsn", defaultDSN, "Database connection string")
		days       = flag.Int("days", defaultDays, "Number of days to analyze (default: 30)")
		width      = flag.Int("width", 60, "Chart width (default: 60)")
		height     = flag.Int("height", 15, "Chart height (default: 15)")
		showGrid   = flag.Bool("grid", true, "Show grid lines (default: true)")
		showStats  = flag.Bool("stats", true, "Show statistics (default: true)")
		summaryOnly = flag.Bool("summary", false, "Show only summary (default: false)")
		help       = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Handle additional positional arguments
	args := flag.Args()
	if len(args) > 0 {
		if parsedDays, err := strconv.Atoi(args[0]); err == nil && parsedDays > 0 {
			*days = parsedDays
		}
	}

	fmt.Printf("=== 粗利推移表示プログラム ===\n")
	fmt.Printf("分析期間: 過去%d日間\n", *days)
	fmt.Printf("接続先: %s\n", *dsn)
	fmt.Println()

	// Initialize database repository
	repo, err := database.NewProfitRepository(*dsn)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer repo.Close()

	// Initialize calculator
	calc := calculator.NewProfitCalculator()

	// Calculate date range
	startDate, endDate := calc.GetDateRange(*days)
	fmt.Printf("対象期間: %s から %s まで\n\n", 
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Fetch profit data
	fmt.Println("データを取得中...")
	profitData, err := repo.GetProfitTrendsForPeriod(startDate, endDate)
	if err != nil {
		log.Fatalf("データ取得エラー: %v", err)
	}

	if len(profitData) == 0 {
		fmt.Println("指定された期間にデータが見つかりませんでした。")
		return
	}

	fmt.Printf("取得データ数: %d件\n", len(profitData))

	// Group by company and warehouse
	fmt.Println("データを分析中...")
	groupedData := calc.GroupByCompanyWarehouse(profitData)

	// Fill missing dates for complete trend visualization
	for key, data := range groupedData {
		groupedData[key] = calc.FillMissingDates(data, startDate, endDate)
	}

	// Create profit trends with statistics
	trends := calc.CreateProfitTrends(groupedData)

	if len(trends) == 0 {
		fmt.Println("分析可能なトレンドデータがありませんでした。")
		return
	}

	// Configure chart renderer
	chartConfig := models.ChartConfig{
		Width:     *width,
		Height:    *height,
		ShowGrid:  *showGrid,
		ShowStats: *showStats,
	}

	chartRenderer := chart.NewTextChart(chartConfig)

	// Display results
	fmt.Printf("\n=== 分析結果 ===\n")
	fmt.Printf("対象組織数: %d\n\n", len(trends))

	if *summaryOnly {
		// Show only summary
		fmt.Print(chartRenderer.RenderSummary(trends))
	} else {
		// Show individual trends
		for i, trend := range trends {
			fmt.Printf("(%d/%d) ", i+1, len(trends))
			fmt.Print(chartRenderer.RenderProfitTrend(trend))
			
			// Add separator between charts
			if i < len(trends)-1 {
				fmt.Println(strings.Repeat("-", 80))
				fmt.Println()
			}
		}

		// Show summary at the end
		fmt.Println(strings.Repeat("=", 80))
		fmt.Print(chartRenderer.RenderSummary(trends))
	}

	fmt.Println("分析完了!")
}

func showHelp() {
	fmt.Println("粗利推移表示プログラム")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  profit-trend-display [オプション] [日数]")
	fmt.Println()
	fmt.Println("オプション:")
	fmt.Println("  -dsn string       データベース接続文字列 (default: root:mypass@tcp(mysql.local:3306)/sample_mysql?parseTime=true)")
	fmt.Println("  -days int         分析対象日数 (default: 30)")
	fmt.Println("  -width int        グラフ幅 (default: 60)")
	fmt.Println("  -height int       グラフ高さ (default: 15)")
	fmt.Println("  -grid             グリッド線を表示 (default: true)")
	fmt.Println("  -stats            統計情報を表示 (default: true)")
	fmt.Println("  -summary          サマリーのみ表示 (default: false)")
	fmt.Println("  -help             このヘルプを表示")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  profit-trend-display                    # デフォルト設定で実行")
	fmt.Println("  profit-trend-display -days 7            # 過去7日間を分析")
	fmt.Println("  profit-trend-display -summary           # サマリーのみ表示")
	fmt.Println("  profit-trend-display -width 80 -height 20  # グラフサイズ変更")
	fmt.Println()
	fmt.Println("機能:")
	fmt.Println("  - 売上データと原価データから粗利を計算")
	fmt.Println("  - 会社別・倉庫別にグループ化して表示")
	fmt.Println("  - テキストベースのグラフで推移を視覚化")
	fmt.Println("  - 統計情報（最大・最小・平均・合計）を表示")
	fmt.Println("  - 欠損日のデータは0として補完")
}